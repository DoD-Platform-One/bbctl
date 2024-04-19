package k3d

import (
	"path"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"
	cmdUtil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
)

var (
	createUse = `create`

	createShort = i18n.T(`Create a k3d cluster`)

	createLong = templates.LongDesc(i18n.T(`Create a minimal k3d cluster in AWS for development or testing.
	This is a wrapper around the k3d-dev.sh script. It must be checked out at --big-bang-repo location.
	Any command line arguments after -- are passed to k3d-dev.sh (including --help).`))

	createExample = templates.Examples(i18n.T(`
	    # Create a default k3d cluster in AWS
		bbctl k3d create

		# Get the full help message from k3d-dev.sh
		bbctl k3d create -- --help
		
		# Create a k3d cluster in AWS on a BIG M5 with a private IP and metalLB
		bbctl k3d create -- -b -p -m`))
)

// NewCreateClusterCmd - command to create a k3d cluster
func NewCreateClusterCmd(factory bbUtil.Factory, streams genericIOOptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     createUse,
		Short:   createShort,
		Long:    createLong,
		Example: createExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdUtil.CheckErr(createCluster(factory, streams, args))
		},
	}

	return cmd
}

// createCluster - pass through options to the k3d-dev script to create a cluster
func createCluster(factory bbUtil.Factory, streams genericIOOptions.IOStreams, args []string) error {
	repoPath := viper.GetString("big-bang-repo")
	if repoPath == "" {
		factory.GetLoggingClient().Error("Big bang repository location not defined (\"big-bang-repo\")")
	}
	command := path.Join(repoPath,
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
	err := cmd.Run()
	return err
}
