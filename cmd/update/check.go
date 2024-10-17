package update

import (
	"errors"

	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
)

// NewUpdateCheckCmd - returns a new Cobra command which implements the `bbctl update check` functionality
func NewUpdateCheckCmd(_ bbUtil.Factory) *cobra.Command {
	var (
		checkUse     = `check`
		checkShort   = i18n.T(`Checks for Big Bang updates`)
		checkLong    = templates.LongDesc(i18n.T(`Checks for Big Bang product updates that can be applied to the cluster`))
		checkExample = templates.Examples(i18n.T(`
			# Check for all updates
			bbctl update check
		`))
	)

	cmd := &cobra.Command{
		Use:     checkUse,
		Short:   checkShort,
		Long:    checkLong,
		Example: checkExample,
		RunE: func(_ *cobra.Command, _ []string) error {
			return errors.New("not implemented")
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	return cmd
}
