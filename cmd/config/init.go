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
	commonInterfaces "repo1.dso.mil/big-bang/product/packages/bbctl/util/common_interfaces"
)

var (
	initUse = `init`

	initShort = i18n.T(`Initializes bbctl configuration information.`)

	initLong = templates.LongDesc(i18n.T(`Initializes the bbctl configurations through prompts and sets the information to a configuration file.`))
)

// NewConfigInitCmd - create a new Cobra config init command
func NewConfigInitCmd(factory bbUtil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   initUse,
		Short:                 initShort,
		Long:                  initLong,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return initBBConfig(factory)
		},
	}
	return cmd
}

func initBBConfig(factory bbUtil.Factory) (err error) {
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

	fmt.Println("Please enter values for the following configurations.")
	for _, c := range configKeys {
		var input string
		// These don't use the standard output client because they are interactive
		fmt.Fprintln(streams.Out, strings.Replace(c.key, "-", " ", -1))
		fmt.Fprintln(streams.Out, c.info)
		if c.optional {
			fmt.Fprintln(streams.Out, "Press enter to skip")
		}
		fmt.Fprint(streams.Out, "$ ")
		fmt.Fscanln(streams.In, &input)
		if c.optional && input != "" || !c.optional {
			config[c.key] = input
		}
	}
	fmt.Println(config)

	return writeConfigFile(&config, yaml.Marshal, os.UserHomeDir, func(name string) (commonInterfaces.FileLike, error) { return os.Create(name) })
}

func writeConfigFile(
	config *map[string]interface{},
	marshallFunc func(interface{}) ([]byte, error),
	homeDirFunc func() (string, error),
	createFunc func(string) (commonInterfaces.FileLike, error),
) (err error) {
	configYaml, err := marshallFunc(&config)
	if err != nil {
		return err
	}
	homedir, err := homeDirFunc()
	if err != nil {
		return err
	}
	configFile, err := createFunc(path.Join(homedir, ".bbctl", "config.yaml"))
	if err != nil {
		return err
	}
	defer func() {
		if newErr := configFile.Close(); newErr != nil {
			if err == nil {
				err = fmt.Errorf("(sole deferred error: %w)", newErr)
			} else {
				err = fmt.Errorf("%w (additional deferred error: %v)", err, newErr)
			}
		}
	}()

	_, err = io.WriteString(configFile, string(configYaml))
	return err
}
