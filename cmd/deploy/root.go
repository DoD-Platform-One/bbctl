package deploy

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
)

var (
	deployUse = `deploy`

	deployShort = i18n.T(`Deploy Big Bang components and preqrequisites`)

	deployLong = templates.LongDesc(i18n.T(`Deploy Big Bang components and prerequisites.

	Note: Before deploying Big Bang, you must first deploy flux into the cluster. See "bbctl deploy flux --help" for more information.
	`))

	deployExample = templates.Examples(i18n.T(``))
)

// NewDeployCmd - parent for deploy commands
func NewDeployCmd(factory bbUtil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     deployUse,
		Short:   deployShort,
		Long:    deployLong,
		Example: deployExample,
		Run: func(cmd *cobra.Command, args []string) {
			subCommands := cmd.Commands()

			var validCommands string
			for i, subCmd := range subCommands {
				validCommands += fmt.Sprintf("%s", subCmd.Use)
				if i != len(subCommands)-1 {
					validCommands += ", "
				}
			}
			_, err := factory.GetIOStream().Out.Write([]byte(fmt.Sprintf("error: must specify one of: %s\n\n", validCommands)))
			factory.GetLoggingClient().HandleError("Unable to write to output stream", err)

			err = cmd.Help()
			factory.GetLoggingClient().HandleError("Unable to write to output stream", err)
		},
	}

	cmd.AddCommand(NewDeployFluxCmd(factory))
	cmd.AddCommand(NewDeployBigBangCmd(factory))

	return cmd
}
