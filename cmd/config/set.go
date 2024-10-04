package config

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util"
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/output"
)

// Set releases config deployed by Big Bang
var (
	setUse   = i18n.T("set [key] [value]")
	setShort = i18n.T("Set a configuration value")
	setLong  = i18n.T("Example usage: bbctl config set KEY VALUE")
)

// Function that returns the set command
func NewSetCmd(factory bbUtil.Factory) (*cobra.Command, error) {
	var setCmd = &cobra.Command{
		Use:   setUse,
		Short: setShort,
		Long:  setLong,
		Args:  cobra.ExactArgs(2), // Ensure exactly 2 arguments: key and value
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			key := args[0]
			value := args[1]
			outputClient, err := factory.GetOutputClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get output client: %w", err)
			}

			filename, _ := cmd.Flags().GetString("output")
			err = setConfigValue(factory, key, value, filename)
			if err != nil {
				return fmt.Errorf("failed to set config value: %w", err)
			}
			return outputClient.Output(&output.BasicOutput{
				Vals: map[string]interface{}{
					"message": "Configuration updated",
					"changes": map[string]string{
						key: value,
					},
				},
			})
		},
	}

	configClient, err := factory.GetConfigClient(setCmd)
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

	// Return the setCmd command to be used elsewhere
	return setCmd, nil
}

// SetConfigValue updates the key-value pair in the config.yaml file
func setConfigValue(factory util.Factory, key string, value string, filename string) error {
	viperInstance, err := factory.GetViper()
	if err != nil {
		return fmt.Errorf("failed to get viper: %w", err)
	}

	viperInstance.Set(key, value)

	if filename != "" {
		return viperInstance.WriteConfigAs(filename)
	}
	return viperInstance.WriteConfig()
}
