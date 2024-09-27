package config

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/output"
)

var (
	configUse = `config`

	configShort = i18n.T(`Manages bbctl configuration information.`)

	configLong = templates.LongDesc(i18n.T(`Manage the current bbctl configurations in the bbctl configuration file.
		This command mirrors some of the functionality of the config-dev.sh script in the Big Bang product repo.
	`))

	configExample = templates.Examples(i18n.T(`
	    # config functionality is implemented in sub-commands. See the specific subcommand help for more information.`))
)

// NewConfigCmd - Returns a minimal parent command for the default config commands
func NewConfigCmd(factory bbUtil.Factory) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:     configUse,
		Short:   configShort,
		Long:    configLong,
		Example: configExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := factory.GetOutputClient(cmd)
			if err != nil {
				return err
			}
			return client.Output(&output.BasicOutput{
				Vals: map[string]interface{}{
					"msg": "Please provide a subcommand for config (see help)",
				},
			})
		},
	}

	cmd.AddCommand(NewConfigViewCmd(factory))
	initCmd, initCmdError := NewConfigInitCmd(factory)
	if initCmdError != nil {
		return nil, fmt.Errorf("error retrieving init command: %w", initCmdError)
	}
	cmd.AddCommand(initCmd)

	return cmd, nil
}
