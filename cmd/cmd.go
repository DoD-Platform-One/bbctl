package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	deploy "repo1.dso.mil/big-bang/product/packages/bbctl/cmd/deploy"
	k3d "repo1.dso.mil/big-bang/product/packages/bbctl/cmd/k3d"
	update "repo1.dso.mil/big-bang/product/packages/bbctl/cmd/update"
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
)

var (
	cmdUse = `bbctl`

	cmdShort = i18n.T(`Big Bang Control command-line tool.`)

	cmdLong = templates.LongDesc(i18n.T(
		`Big Bang Control command-line tool allows you to run commands against Kubernetes clusters 
		to simplify development, deployment, auditing, and troubleshooting of Big Bang.`))

	cmdExample = templates.Examples(i18n.T(`
		# List all available commands
		bbctl help`))
)

// NewRootCmd - create a new Cobra root command
func NewRootCmd(factory bbUtil.Factory, streams genericIOOptions.IOStreams) (*cobra.Command, error) {

	cmd := &cobra.Command{
		Use:     cmdUse,
		Short:   cmdShort,
		Long:    cmdLong,
		Example: cmdExample,
	}

	cmd.CompletionOptions.DisableDefaultCmd = false
	cmd.CompletionOptions.DisableNoDescFlag = true
	cmd.CompletionOptions.DisableDescriptions = false

	cmd.AddCommand(NewCompletionCmd(factory, streams))
	cmd.AddCommand(NewConfigCmd(factory, streams))
	cmd.AddCommand(NewVersionCmd(factory, streams))
	cmd.AddCommand(NewReleasesCmd(factory, streams))
	cmd.AddCommand(NewValuesCmd(factory, streams))
	cmd.AddCommand(NewStatusCmd(factory, streams))
	cmd.AddCommand(NewViolationsCmd(factory, streams))
	policiesCmd, policiesCmdError := NewPoliciesCmd(factory, streams)
	if policiesCmdError != nil {
		return nil, fmt.Errorf("Error retrieving Policies Command: %v", policiesCmdError)
	}
	cmd.AddCommand(policiesCmd)

	preflightCheckCmd, preflightCheckCmdError := NewPreflightCheckCmd(factory, streams)
	if preflightCheckCmdError != nil {
		return nil, fmt.Errorf("Error retrieving PreflightCheck Command: %w", preflightCheckCmdError)
	}
	cmd.AddCommand(preflightCheckCmd)

	k3dCmd, K3dCmdError := k3d.NewK3dCmd(factory, streams)
	if K3dCmdError != nil {
		return nil, fmt.Errorf("Error retrieving k3d Command: %v", K3dCmdError)
	}
	cmd.AddCommand(k3dCmd)
	cmd.AddCommand(deploy.NewDeployCmd(factory, streams))
	cmd.AddCommand(update.NewUpdateCmd(factory, streams))

	return cmd, nil
}
