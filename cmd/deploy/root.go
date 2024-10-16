package deploy

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
)

// NewDeployCmd - parent for deploy commands
func NewDeployCmd(factory bbUtil.Factory) (*cobra.Command, error) {
	var (
		deployUse   = `deploy`
		deployShort = i18n.T(`Deploy Big Bang components and preqrequisites`)
		deployLong  = templates.LongDesc(i18n.T(`Deploy Big Bang components and prerequisites.
	
		Note: Before deploying Big Bang, you must first deploy flux into the cluster. See "bbctl deploy flux --help" for more information.
		`))
		deployExample = templates.Examples(i18n.T(``))
	)

	cmd := &cobra.Command{
		Use:     deployUse,
		Short:   deployShort,
		Long:    deployLong,
		Example: deployExample,
		RunE: func(cmd *cobra.Command, _ []string) error {
			subCommands := cmd.Commands()

			var validCommands string
			for i, subCmd := range subCommands {
				validCommands += subCmd.Use
				if i != len(subCommands)-1 {
					validCommands += ", "
				}
			}
			loggingClient, err := factory.GetLoggingClient()
			if err != nil {
				return fmt.Errorf("unable to get logging client: %w", err)
			}
			loggingClient.Error(fmt.Sprintf("error: must specify one of: %s\n\n", validCommands))

			err = cmd.Help()
			if err != nil {
				return fmt.Errorf("unable to write to output stream: %w", err)
			}
			return nil
		},
	}

	cmd.AddCommand(NewDeployFluxCmd(factory))
	bigBangCmd, bigBangCmdError := NewDeployBigBangCmd(factory)
	if bigBangCmdError != nil {
		return nil, fmt.Errorf("error retrieving BigBang Command: %w", bigBangCmdError)
	}
	cmd.AddCommand(bigBangCmd)

	return cmd, nil
}
