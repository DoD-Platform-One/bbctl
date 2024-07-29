package k3d

import (
	"path"

	"github.com/spf13/cobra"
	cmdUtil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
)

var (
	createUse = `create`

	createShort = i18n.T(`Creates a k3d cluster`)

	createLong = templates.LongDesc(i18n.T(`Creates a minimal k3d cluster in AWS for development or testing.
	This is a wrapper around the k3d-dev.sh script. It must be checked out at --big-bang-repo location.
	Any command line arguments following -- are passed to k3d-dev.sh (including --help).`))

	createExample = templates.Examples(i18n.T(`
	    # Create a default k3d cluster in AWS
		bbctl k3d create

		# Get the full help message from k3d-dev.sh
		bbctl k3d create -- --help
		
		# Create a k3d cluster in AWS on a BIG M5 with a private IP and metalLB installed
		bbctl k3d create -- -b -p -m`))
)

// NewCreateClusterCmd - Returns a command to create the k3d cluster using createCluster
func NewCreateClusterCmd(factory bbUtil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     createUse,
		Short:   createShort,
		Long:    createLong,
		Example: createExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdUtil.CheckErr(createCluster(factory, cmd, args))
		},
	}

	return cmd
}

// createCluster - Passes through the global configurations, the path to the script, and command line arguments to the k3d-dev script to create the k3d dev cluster
func createCluster(factory bbUtil.Factory, cobraCmd *cobra.Command, args []string) error {
	streams := factory.GetIOStream()
	configClient, err := factory.GetConfigClient(cobraCmd)
	if err != nil {
		return err
	}
	config := configClient.GetConfig()
	command := path.Join(config.BigBangRepo,
		"docs",
		"assets",
		"scripts",
		"developer",
		"k3d-dev.sh",
	)
	cmd := factory.GetCommandWrapper(command, args...)
	cmd.SetStdout(streams.Out)
	cmd.SetStderr(streams.ErrOut)
	cmd.SetStdin(streams.In)
	err = cmd.Run()
	return err
}
