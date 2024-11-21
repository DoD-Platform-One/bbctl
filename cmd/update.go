package cmd

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/output"
)

// NewUpdateCmd returns a command for the `bbctl update` command
func NewUpdateCmd(factory bbUtil.Factory) *cobra.Command {
	var (
		updateUse     = `update`
		updateShort   = i18n.T(`Manage updates to bbctl the Big Bang product`)
		updateLong    = templates.LongDesc(i18n.T(`Manage updates to bbctl and the Big Bang product`))
		updateExample = templates.Examples(i18n.T(`
			# No update functionality has been implemented yet`))
	)

	cmd := &cobra.Command{
		Use:     updateUse,
		Short:   updateShort,
		Long:    updateLong,
		Example: updateExample,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := factory.GetOutputClient(cmd)
			if err != nil {
				return err
			}
			return client.Output(&output.BasicOutput{
				Vals: map[string]any{
					"msg": "No update functionality has been implemented yet",
				},
			})
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	return cmd
}
