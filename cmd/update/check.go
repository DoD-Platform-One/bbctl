package update

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
)

var (
	checkUse = `check`

	checkShort = i18n.T(`Checks for Big Bang updates`)

	checkLong = templates.LongDesc(i18n.T(`Checks for Big Bang product updates that can be applied to the cluster`))

	checkExample = templates.Examples(i18n.T(`
	    # Check for all updates
		bbctl update check
	`))
)

// NewUpdateCheckCmd - returns a new Cobra command which implements the `bbctl update check` functionality
// TODO: Implement in bbctl #196
func NewUpdateCheckCmd(factory bbUtil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     checkUse,
		Short:   checkShort,
		Long:    checkLong,
		Example: checkExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("Not Implemented")
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	return cmd
}
