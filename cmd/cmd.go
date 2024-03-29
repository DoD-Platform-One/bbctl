package cmd

import (
	bbutil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
	bbk8sutil "repo1.dso.mil/big-bang/product/packages/bbctl/util/k8s"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
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

// NewRootCmd - create a new Cobra root command
func NewRootCmd(factory bbutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {

	cmd := &cobra.Command{
		Use:     cmdUse,
		Short:   cmdShort,
		Long:    cmdLong,
		Example: cmdExample,
	}

	cmd.CompletionOptions.DisableDefaultCmd = false
	cmd.CompletionOptions.DisableNoDescFlag = true
	cmd.CompletionOptions.DisableDescriptions = false

	cmd.AddCommand(NewCompletionCmd(factory, bbk8sutil.GetIOStream()))
	cmd.AddCommand(NewVersionCmd(factory, bbk8sutil.GetIOStream()))
	cmd.AddCommand(NewReleasesCmd(factory, bbk8sutil.GetIOStream()))
	cmd.AddCommand(NewValuesCmd(factory, bbk8sutil.GetIOStream()))
	cmd.AddCommand(NewStatusCmd(factory, bbk8sutil.GetIOStream()))
	cmd.AddCommand(NewViolationsCmd(factory, bbk8sutil.GetIOStream()))
	cmd.AddCommand(NewPoliciesCmd(factory, bbk8sutil.GetIOStream()))
	cmd.AddCommand(NewPreflightCheckCmd(factory, bbk8sutil.GetIOStream()))

	return cmd
}
