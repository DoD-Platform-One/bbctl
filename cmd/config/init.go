package config

import (
	"fmt"
	"path"
	"strings"

	"io"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
	commonInterfaces "repo1.dso.mil/big-bang/product/packages/bbctl/util/commoninterfaces"
)

// NewConfigInitCmd - create a new Cobra config init command
func NewConfigInitCmd(factory bbUtil.Factory) (*cobra.Command, error) {
	var (
		initUse   = `init`
		initShort = i18n.T(`Initializes bbctl configuration information.`)
		initLong  = templates.LongDesc(i18n.T(`Initializes the bbctl configurations through prompts and sets the information to a configuration file.`))
	)

	cmd := &cobra.Command{
		Use:                   initUse,
		Short:                 initShort,
		Long:                  initLong,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return initBBConfig(factory, cmd)
		},
	}

	configClient, err := factory.GetConfigClient(cmd)
	if err != nil {
		return nil, fmt.Errorf("unable to get config client: %w", err)
	}

	err = configClient.SetAndBindFlag(
		"output",
		"o",
		"$HOME/.bbctl/",
		"Specify the output file where all configurations will be stored",
	)
	if err != nil {
		return nil, fmt.Errorf("error setting and binding output flag: %w", err)
	}
	err = configClient.SetAndBindFlag(
		"bbctl-log-level",
		"",
		"info",
		"Log level for bbctl. Options are debug, info, warn, error",
	)
	if err != nil {
		return nil, fmt.Errorf("error setting and binding bbctl-log-level flag: %w", err)
	}
	err = configClient.SetAndBindFlag(
		"bbctl-log-add-source",
		"",
		false,
		"Add source to log output",
	)
	if err != nil {
		return nil, fmt.Errorf("error setting and binding bbctl-log-add-source flag: %w", err)
	}
	err = configClient.SetAndBindFlag(
		"bbctl-log-format",
		"",
		"json",
		"Log format for bbctl. Options are json, text",
	)
	if err != nil {
		return nil, fmt.Errorf("error setting and binding bbctl-log-format flag: %w", err)
	}
	err = configClient.SetAndBindFlag(
		"big-bang-repo",
		"",
		"",
		"Location on the filesystem where the Big Bang product repo is checked out",
	)
	if err != nil {
		return nil, fmt.Errorf("error setting and binding big-bang-repo flag: %w", err)
	}
	err = configClient.SetAndBindFlag(
		"bbctl-log-output",
		"",
		"stdout",
		"Log output for bbctl. Options are stdout, stderr, file",
	)
	if err != nil {
		return nil, fmt.Errorf("error setting and binding bbctl-log-output flag: %w", err)
	}
	err = configClient.SetAndBindFlag(
		"big-bang-credential-helper",
		"",
		"",
		"Location of a program that bbctl can use as a credential helper",
	)
	if err != nil {
		return nil, fmt.Errorf("error setting and binding big-bang-credential-helper flag: %w", err)
	}

	return cmd, nil
}

func initBBConfig(factory bbUtil.Factory, command *cobra.Command) error {
	streams, err := factory.GetIOStream()
	if err != nil {
		return fmt.Errorf("error getting IO streams: %w", err)
	}
	config := make(map[string]interface{})

	configKeys := []struct {
		key      string
		info     string
		optional bool
	}{{
		key:      "bbctl-log-format",
		info:     "Log format for bbctl. Options are json, text",
		optional: false,
	},
		{
			key:      "bbctl-log-level",
			info:     "Log level for bbctl. Options are debug, info, warn, error",
			optional: false,
		},
		{
			key:      "bbctl-log-output",
			info:     "Log output for bbctl. Options are stdout, stderr, file",
			optional: false,
		},
		{
			key:      "big-bang-credential-helper",
			info:     "Location of a program that bbctl can use as a credential helper",
			optional: true,
		},
		{
			key:      "big-bang-repo",
			info:     "Location on the filesystem where the Big Bang product repo is checked out",
			optional: false,
		}}

	fmt.Println("Please enter values for the following configurations.") //nolint:forbidigo
	for _, c := range configKeys {
		var input string
		flag, _ := command.Flags().GetString(c.key)
		// These don't use the standard output client because they are interactive
		if flag != "" {
			config[c.key] = flag
		} else {
			fmt.Fprintln(streams.Out, strings.ReplaceAll(c.key, "-", " "))
			fmt.Fprintln(streams.Out, c.info)
			if c.optional {
				fmt.Fprintln(streams.Out, "Press enter to skip")
			}
			fmt.Fprint(streams.Out, "$ ")
			fmt.Fscanln(streams.In, &input) //nolint:errcheck
			if c.optional && input != "" || !c.optional {
				config[c.key] = input
			}
		}
	}

	output, _ := command.Flags().GetString("output")
	if output == "" {
		var input string
		fmt.Println("Please enter the output path for the config.yaml file.") //nolint:forbidigo
		fmt.Fprintln(streams.Out, "Press enter to skip")
		fmt.Fprint(streams.Out, "$ ")
		fmt.Fscanln(streams.In, &input) //nolint:errcheck
		if input != "" {
			output = input
		} else {
			homedir, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			output = path.Join(homedir, ".bbctl")
		}
	}
	return writeConfigFile(&config, yaml.Marshal, output, func(name string) (commonInterfaces.FileLike, error) { return os.Create(name) })
}

func writeConfigFile(
	config *map[string]interface{},
	marshallFunc func(interface{}) ([]byte, error),
	outputDir string,
	createFunc func(string) (commonInterfaces.FileLike, error),
) (err error) {
	configYaml, err := marshallFunc(&config)
	if err != nil {
		return err
	}
	configFile, err := createFunc(path.Join(outputDir, "config.yaml"))
	if err != nil {
		return err
	}
	defer func() {
		if newErr := configFile.Close(); newErr != nil {
			if err == nil {
				err = fmt.Errorf("(sole deferred error: %w)", newErr)
			} else {
				err = fmt.Errorf("%w (additional deferred error: %w)", err, newErr)
			}
		}
	}()

	_, err = io.Writer.Write(configFile, configYaml)
	return err
}
