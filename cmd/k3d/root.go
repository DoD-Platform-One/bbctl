package k3d

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/output"
)

var (
	k3dUse = `k3d`

	k3dShort = i18n.T(`Manage k3d cluster`)

	k3dLong = templates.LongDesc(
		i18n.T(`Manage a minimal k3d cluster for Big Bang development or testing.
		This command mirrors some of the functionality of the k3d-dev.sh script in the Big Bang product repo.
	`),
	)

	k3dExample = templates.Examples(i18n.T(`
	    # k3d functionality is implemented in sub-commands. See the specific subcommand help for more information.`))
)

// NewK3dCmd - Returns a minimal parent command for the default k3d commands
func NewK3dCmd(factory bbUtil.Factory) (*cobra.Command, error) {
	_, err := factory.GetIOStream()
	if err != nil {
		return nil, fmt.Errorf("unable to get IOStreams: %w", err)
	}
	cmd := &cobra.Command{
		Use:     k3dUse,
		Short:   k3dShort,
		Long:    k3dLong,
		Example: k3dExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			outputClient, err := factory.GetOutputClient(cmd)
			if err != nil {
				return fmt.Errorf("Unable to  create output client: %w", err)
			}
			outputMap := map[string]interface{}{
				"Message": "Please provide a subcommand for k3d (see help)",
			}

			err = outputClient.Output(&output.BasicOutput{Vals: outputMap})
			return err
		},
	}

	cmd.AddCommand(NewCreateClusterCmd(factory))
	cmd.AddCommand(NewDestroyClusterCmd(factory))
	cmd.AddCommand(NewShellProfileCmd(factory))
	sshCmd, sshCmdError := NewSSHCmd(factory)
	if sshCmdError != nil {
		return nil, fmt.Errorf("error retrieving ssh command: %w", sshCmdError)
	}
	cmd.AddCommand(sshCmd)
	hostsCmd, hostsCmdError := NewHostsCmd(factory)
	if hostsCmdError != nil {
		return nil, fmt.Errorf("error retrieving hosts command: %w", hostsCmdError)
	}
	cmd.AddCommand(hostsCmd)

	return cmd, nil
}
