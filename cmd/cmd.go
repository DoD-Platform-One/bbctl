package cmd

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	bbutil "repo1.dso.mil/platform-one/big-bang/apps/product-tools/bbctl/util"
	bbk8sutil "repo1.dso.mil/platform-one/big-bang/apps/product-tools/bbctl/util/k8s"
)

var (
	cmdUse = `bbctl`

	cmdShort = i18n.T(`BigBang command-line tool.`)

	cmdLong = templates.LongDesc(i18n.T(
		`BigBang command-line tool allows you to run commands against Kubernetes clusters 
		to simplify development, deployment, auditing, and troubleshooting of BigBang.`))

	cmdExample = templates.Examples(i18n.T(`
		# Get help
		bbctl help`))
)

func NewRootCmd(factory bbutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {

	cmd := &cobra.Command{
		Use:     cmdUse,
		Short:   cmdShort,
		Long:    cmdLong,
		Example: cmdExample,
	}

	cmd.AddCommand(NewVersionCmd(factory, bbk8sutil.GetIOStream()))
	cmd.AddCommand(NewGetReleasesCmd(factory, bbk8sutil.GetIOStream()))
	cmd.AddCommand(NewGetValuesCmd(factory, bbk8sutil.GetIOStream()))

	return cmd
}
