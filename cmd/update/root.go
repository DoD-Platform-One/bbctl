package update

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/output"
)

var (
	updateRootUse = `update`

	updateRootShort = i18n.T(`Manage updates to the Big Bang product`)

	updateRootLong = templates.LongDesc(i18n.T(`Manage updates to the Big Bang product`))

	updateRootExample = templates.Examples(i18n.T(`
	    # Update functionality is implemented in sub-commands. See the specific subcommand help for more information.`))
)

// NewUpdateCmd - Returns a minimal parent command for the `bbctl update` commands
func NewUpdateCmd(factory bbUtil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     updateRootUse,
		Short:   updateRootShort,
		Long:    updateRootLong,
		Example: updateRootExample,
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
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(NewUpdateCheckCmd(factory))

	return cmd
}
