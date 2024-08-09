package update

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
)

var (
	updateRootUse = `update`

	updateRootShort = i18n.T(`Manage updates to the Big Bang product`)

	updateRootLong = templates.LongDesc(i18n.T(`Manage updates to the Big Bang product`))

	updateRootExample = templates.Examples(i18n.T(`
	    # Update functionality is implemented in sub-commands. See the specific subcommand help for more information.`))
)

// NewUpdateCmd - Returns a minimal parent command for the `bbctl update` commands
func NewUpdateCmd(factory bbUtil.Factory) (*cobra.Command, error) {
	streams, err := factory.GetIOStream()
	if err != nil {
		return nil, fmt.Errorf("Unable to create IO streams: %v", err)
	}
	cmd := &cobra.Command{
		Use:     updateRootUse,
		Short:   updateRootShort,
		Long:    updateRootLong,
		Example: updateRootExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := streams.Out.Write([]byte(fmt.Sprintln("Please provide a subcommand for update (see help)")))
			if err != nil {
				return fmt.Errorf("Unable to write to output stream: %v", err)
			}
			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(NewUpdateCheckCmd(factory))

	return cmd, nil
}
