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
func NewDeployCmd(factory bbUtil.Factory) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:     deployUse,
		Short:   deployShort,
		Long:    deployLong,
		Example: deployExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			subCommands := cmd.Commands()

			var validCommands string
			for i, subCmd := range subCommands {
				validCommands += fmt.Sprintf("%s", subCmd.Use)
				if i != len(subCommands)-1 {
					validCommands += ", "
				}
			}
			_, err := factory.GetIOStream().Out.Write([]byte(fmt.Sprintf("error: must specify one of: %s\n\n", validCommands)))
			if err != nil {
				return fmt.Errorf("Unable to write to output stream: %w", err)
			}

			err = cmd.Help()
			if err != nil {
				return fmt.Errorf("Unable to write to output stream: %w", err)
			}
			return nil
		},
	}

	cmd.AddCommand(NewDeployFluxCmd(factory))
	bigBangCmd, bigBangCmdError := NewDeployBigBangCmd(factory)
	if bigBangCmdError != nil {
		return nil, fmt.Errorf("Error retrieving BigBang Command: %w", bigBangCmdError)
	}
	cmd.AddCommand(bigBangCmd)

	return cmd, nil
}
