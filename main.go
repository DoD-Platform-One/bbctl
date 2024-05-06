package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	pFlag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	genericCliOptions "k8s.io/cli-runtime/pkg/genericclioptions"
	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"
	"repo1.dso.mil/big-bang/product/packages/bbctl/cmd"
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
	bbK8sUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/k8s"
)

func main() {
	flags := pFlag.NewFlagSet("bbctl", pFlag.ExitOnError)
	factory := bbUtil.NewFactory()
	streams := bbK8sUtil.GetIOStream()
	injectableMain(factory, flags, streams)
}

func injectableMain(factory bbUtil.Factory, flags *pFlag.FlagSet, streams genericIOOptions.IOStreams) {
	flags.Bool("bbctl-log-add-source",
		false,
		"Add source to log output")
	flags.String("bbctl-log-file",
		"",
		"Log file for bbctl. Only used if bbctl-log-output is set to file")
	flags.String("bbctl-log-format",
		"",
		"Log format for bbctl. Options are json, text")
	flags.String("bbctl-log-level",
		"",
		"Log level for bbctl. Options are debug, info, warn, error")
	flags.String("bbctl-log-output",
		"",
		"Log output for bbctl. Options are stdout, stderr, file")
	flags.String("big-bang-credential-helper",
		"",
		"Location of a program that bbctl can use as a credential helper")
	flags.String("big-bang-repo",
		"",
		"Location on the filesystem where the bigbang product repo is checked out")

	// setup the init logger
	initSlogHandlerOptions := slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelWarn,
	}
	// logs to stderr
	initLogger := slog.New(slog.NewJSONHandler(streams.ErrOut, &initSlogHandlerOptions))
	viperInstance := factory.GetViper()

	cobra.OnInitialize(func() {
		// automatically read in environment variables that match supported flags
		// e.g. kubeconfig is a recognized flag so the corresponding env variable is KUBECONFIG
		viperInstance.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
		viperInstance.AutomaticEnv()

		homeDirname, err := os.UserHomeDir()
		if err != nil {
			initLogger.Error("Error getting user home directory: %v", err)
			panic(err)
		}
		viperInstance.SetConfigName("config")
		viperInstance.SetConfigType("yaml")
		viperInstance.AddConfigPath(path.Join(homeDirname,
			".bbctl"))
		viperInstance.AddConfigPath("/etc/bbctl")
		viperInstance.AddConfigPath(".")
		// Support XDG_CONFIG_HOME standard, default to $HOME/.config/bbctl
		xdgConfigHome, exists := os.LookupEnv("XDG_CONFIG_HOME")
		if !exists {
			xdgConfigHome = filepath.Join(homeDirname, ".config")
		}
		viperInstance.AddConfigPath(filepath.Join(xdgConfigHome, "bbctl"))

		if err := viperInstance.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				// Config file not found; ignore error if desired
				initLogger.Warn("Config file not found (~/.bbctl/config, /etc/bbctl/config, or ./config).")
			} else {
				// Config file was found but another error was produced
				initLogger.Error("Error reading config file: %v", err)
				panic(err)
			}
		}

		err = viperInstance.BindPFlags(flags)
		if err != nil {
			initLogger.Error("Error binding flags to viper: %v", err)
			panic(err)
		}
		configClient, err := factory.GetConfigClient(nil)
		if err != nil {
			initLogger.Error("Error getting config client: %v", err)
			panic(err)
		}
		config := configClient.GetConfig()
		logger := setupSlog(initLogger,
			streams,
			config.LogAddSource,
			config.LogFile,
			config.LogFormat,
			config.LogLevel,
			config.LogOutput,
		)
		logger.Debug("Logger setup complete")
		allSettings, err := json.Marshal(viperInstance.AllSettings())
		if err != nil {
			logger.Error("Error marshalling all settings: %v", err)
			panic(err)
		}
		logger.Debug(fmt.Sprintf("Command line settings: %v", string(allSettings)))
	})

	bbctlCmd := cmd.NewRootCmd(factory, streams)

	flags.AddFlagSet(bbctlCmd.PersistentFlags())
	pFlag.CommandLine = flags

	// This set of flags is the one used for the kubectl configuration such as:
	// namespace, kube-config, insecure, and so on
	kubeConfigFlags := genericCliOptions.NewConfigFlags(false)
	kubeConfigFlags.AddFlags(flags)

	// It is a set of flags related to a specific resource such as: label selector
	kubeResourceBuilderFlags := genericCliOptions.NewResourceBuilderFlags()
	kubeResourceBuilderFlags.AddFlags(flags)

	// Bind the flags to viper
	factory.GetLoggingClient().HandleError("error binding flags to viper: %v", viperInstance.BindPFlags(flags))

	// echo the flags
	factory.GetLoggingClient().Debug(fmt.Sprintf("Global Flags: %v", flags.Args()))

	cobra.CheckErr(bbctlCmd.Execute())
}

// setupSlog - setup the slog logger
func setupSlog(
	initLogger *slog.Logger,
	streams genericIOOptions.IOStreams,
	addSource bool,
	logFileString string,
	logFormatString string,
	logLevelString string,
	logOutputString string,
) *slog.Logger {
	// log level
	var logLevel slog.Level
	switch ll := logLevelString; ll {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	case "":
		logLevel = slog.LevelWarn
		initLogger.Warn("No log level defined, defaulting to warn")
	default:
		initLogger.Error(fmt.Sprintf("Invalid log level: %v", ll))
		panic("Invalid log level")
	}

	// handler options
	slogHandlerOptions := slog.HandlerOptions{
		AddSource: addSource,
		Level:     logLevel,
	}

	// log output
	var writer io.Writer
	switch lo := logOutputString; lo {
	case "file":
		if logFileString == "" {
			initLogger.Error("Log file not defined")
			panic("Log file not defined")
		}
		file, err := os.OpenFile(logFileString, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			initLogger.Error(fmt.Sprintf("Error opening log file: %v", err))
			panic(err)
		}
		defer file.Close()
		writer = file
	case "stdout":
		writer = streams.Out
	case "stderr":
		writer = streams.ErrOut
	case "":
		writer = streams.ErrOut
		initLogger.Warn("No log output defined, defaulting to stderr")
	default:
		initLogger.Error(fmt.Sprintf("Invalid log output: %v", logOutputString))
		panic("Invalid log output")
	}

	// logger
	var logger *slog.Logger
	switch lf := logFormatString; lf {
	case "json":
		logger = slog.New(slog.NewJSONHandler(writer, &slogHandlerOptions))
	case "text":
		logger = slog.New(slog.NewTextHandler(writer, &slogHandlerOptions))
	case "":
		logger = slog.New(slog.NewTextHandler(writer, &slogHandlerOptions))
		initLogger.Warn("No log format defined, defaulting to text")
	default:
		initLogger.Error(fmt.Sprintf("Invalid log format: %v", logFormatString))
		panic("Invalid log format")
	}

	slog.SetDefault(logger)
	return logger
}
