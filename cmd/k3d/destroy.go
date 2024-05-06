package k3d

import (
	"path"

	"github.com/spf13/cobra"
	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"
	cmdUtil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
)

var (
	destroyUse = `destroy`

	destroyShort = i18n.T(`destroy a k3d cluster`)

	destroyLong = templates.LongDesc(i18n.T(`destroy a previously created AWS k3d cluster.
	This is a wrapper around the k3d-dev.sh script. It must be checked out at --big-bang-repo location.
	Any command line arguments after -- are passed to k3d-dev.sh (including --help).`))

	destroyExample = templates.Examples(i18n.T(`
	    # destroy your k3d cluster previously built with 'bbctl k3d create'
		bbctl k3d destroy
		
		# To get the full help message from k3d-dev.sh
		bbctl k3d destroy -- --help`))
)

// NewDestroyClusterCmd - command to destroy a k3d cluster
func NewDestroyClusterCmd(factory bbUtil.Factory, streams genericIOOptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     destroyUse,
		Short:   destroyShort,
		Long:    destroyLong,
		Example: destroyExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdUtil.CheckErr(destroyCluster(factory, cmd, streams, args))
		},
	}

	return cmd
}

// destroyCluster - pass through options to the k3d-dev script to destroy a cluster
func destroyCluster(factory bbUtil.Factory, cobraCmd *cobra.Command, streams genericIOOptions.IOStreams, args []string) error {
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
	args = append([]string{"-d"}, args...)
	cmd := factory.GetCommandWrapper(command, args...)
	cmd.SetStderr(streams.ErrOut)
	cmd.SetStdout(streams.Out)
	cmd.SetStdin(streams.In)
	err = cmd.Run()
	return err
}
