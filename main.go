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
	bbUtilPool "repo1.dso.mil/big-bang/product/packages/bbctl/util/pool"
)

// main - main exectable function for bbctl
func main() {
	flags := pFlag.NewFlagSet("bbctl", pFlag.ExitOnError)
	pooledFactory := bbUtilPool.NewPooledFactory()
	factory := bbUtil.NewFactory(pooledFactory)
	pooledFactory.SetUnderlyingFactory(factory)
	injectableMain(pooledFactory, flags)
}

// injectableMain - helper function for main
func injectableMain(factory bbUtil.Factory, flags *pFlag.FlagSet) {
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
		"Location on the filesystem where the Big Bang product repo is checked out")

	// setup the init logger
	streams, err := factory.GetIOStream()
	if err != nil {
		fmt.Printf("Error getting IO streams: %v", err.Error())
		os.Exit(1)
	}
	initSlogHandlerOptions := slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelWarn,
	}
	// logs to stderr
	initLogger := slog.New(slog.NewJSONHandler(streams.ErrOut, &initSlogHandlerOptions))
	viperInstance, err := factory.GetViper()
	if err != nil {
		initLogger.Error(fmt.Sprintf("Error getting viper: %v", err.Error()))
		os.Exit(1)
	}

	cobra.OnInitialize(func() {
		// automatically read in environment variables that match supported flags
		// e.g. kubeconfig is a recognized flag so the corresponding env variable is KUBECONFIG
		viperInstance.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
		viperInstance.AutomaticEnv()

		homeDirname, err := os.UserHomeDir()
		if err != nil {
			initLogger.Error(fmt.Sprintf("Error getting user home directory: %v", err.Error()))
			os.Exit(1)
		}
		viperInstance.SetConfigName("config")
		viperInstance.SetConfigType("yaml")
		// Order the config paths so that the search is from most specific to least specific
		viperInstance.AddConfigPath(".")
		viperInstance.AddConfigPath(path.Join(".", ".bbctl"))
		viperInstance.AddConfigPath(path.Join(homeDirname,
			".bbctl"))
		// Support XDG_CONFIG_HOME standard, default to $HOME/.config/bbctl
		xdgConfigHome, exists := os.LookupEnv("XDG_CONFIG_HOME")
		if !exists {
			xdgConfigHome = filepath.Join(homeDirname, ".config")
		}
		viperInstance.AddConfigPath(filepath.Join(xdgConfigHome, "bbctl"))
		viperInstance.AddConfigPath("/etc/bbctl")

		if err := viperInstance.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				// Config file not found; ignore error if desired
				initLogger.Warn("Config file not found (~/.bbctl/config, /etc/bbctl/config, or ./config).")
			} else {
				// Config file was found but another error was produced
				initLogger.Error(fmt.Sprintf("Error reading config file: %v", err.Error()))
				os.Exit(1)
			}
		}

		err = viperInstance.BindPFlags(flags)
		if err != nil {
			initLogger.Error(fmt.Sprintf("Error binding flags to viper: %v", err.Error()))
			os.Exit(1)
		}
		configClient, err := factory.GetConfigClient(nil)
		if err != nil {
			initLogger.Error(fmt.Sprintf("Error getting config client: %v", err.Error()))
			os.Exit(1)
		}
		config, configErr := configClient.GetConfig()
		if configErr != nil {
			initLogger.Error(fmt.Sprintf("Error getting config: %v", configErr.Error()))
			os.Exit(1)
		}
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
			logger.Error(fmt.Sprintf("Error marshalling all settings: %v", err.Error()))
			os.Exit(1)
		}
		logger.Debug(fmt.Sprintf("Command line settings: %v", string(allSettings)))
	})

	bbctlCmd, rootCmdError := cmd.NewRootCmd(factory)
	if rootCmdError != nil {
		initLogger.Error(fmt.Sprintf("Error retrieving root command: %v", rootCmdError.Error()))
		os.Exit(1)
	}

	flags.AddFlagSet(bbctlCmd.PersistentFlags())
	pFlag.CommandLine = flags

	// This set of flags is the one used for the kubectl configuration such as:
	// namespace, kube-config, insecure, and so on
	kubeConfigFlags := genericCliOptions.NewConfigFlags(false)
	kubeConfigFlags.AddFlags(flags)

	// It is a set of flags related to a specific resource such as: label selector
	kubeResourceBuilderFlags := genericCliOptions.NewResourceBuilderFlags()
	kubeResourceBuilderFlags.AddFlags(flags)

	logger, err := factory.GetLoggingClient()
	if err != nil {
		initLogger.Error(fmt.Sprintf("Error getting logging client: %v", err.Error()))
		os.Exit(1)
	}

	// Bind the flags to viper
	err = viperInstance.BindPFlags(flags)
	if err != nil {
		initLogger.Error(fmt.Sprintf("Error binding flags to viper: %v", err.Error()))
		os.Exit(1)
	}

	// echo the flags
	logger.Debug(fmt.Sprintf("Global Flags: %v", flags.Args()))
	err = bbctlCmd.Execute()
	if err != nil {
		initLogger.Error(fmt.Sprintf("Error executing command: %v", err.Error()))
		os.Exit(1)
	}
}

// setupSlog - setup the slog logger
func setupSlog(
	initLogger *slog.Logger,
	streams *genericIOOptions.IOStreams,
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
		os.Exit(1)
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
			os.Exit(1)
		}
		file, err := os.OpenFile(logFileString, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			initLogger.Error(fmt.Sprintf("Error opening log file: %v", err.Error()))
			os.Exit(1)
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
		os.Exit(1)
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
		os.Exit(1)
	}

	slog.SetDefault(logger)
	return logger
}
