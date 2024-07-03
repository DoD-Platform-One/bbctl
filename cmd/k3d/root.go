package k3d

import (
	"fmt"

	"github.com/spf13/cobra"
	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
)

var (
	k3dUse = `k3d`

	k3dShort = i18n.T(`Manage k3d cluster`)

	k3dLong = templates.LongDesc(i18n.T(`Manage a minimal k3d cluster for bigbang development or testing`))

	k3dExample = templates.Examples(i18n.T(`
	    # K3D functionality is implemented in sub-commands. See the specific subcommand help for more information.`))
)

// NewK3dCmd - Returns a minimal parent command for the default k3d commands
func NewK3dCmd(factory bbUtil.Factory, streams genericIOOptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     k3dUse,
		Short:   k3dShort,
		Long:    k3dLong,
		Example: k3dExample,
		Run: func(cmd *cobra.Command, args []string) {
			_, err := streams.Out.Write([]byte(fmt.Sprintln("Please provide a subcommand for k3d (see help)")))
			factory.GetLoggingClient().HandleError("Unable to write to output stream", err)
		},
	}

	cmd.AddCommand(NewCreateClusterCmd(factory, streams))
	cmd.AddCommand(NewDestroyClusterCmd(factory, streams))
	cmd.AddCommand(NewShellProfileCmd(factory, streams))
	cmd.AddCommand(NewSSHCmd(factory, streams))
	cmd.AddCommand(NewHostsCmd(factory, streams))

	return cmd
}
