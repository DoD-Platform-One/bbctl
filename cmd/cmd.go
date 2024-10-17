package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	config "repo1.dso.mil/big-bang/product/packages/bbctl/cmd/config"
	deploy "repo1.dso.mil/big-bang/product/packages/bbctl/cmd/deploy"
	k3d "repo1.dso.mil/big-bang/product/packages/bbctl/cmd/k3d"
	update "repo1.dso.mil/big-bang/product/packages/bbctl/cmd/update"
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
)

// NewRootCmd - create a new Cobra root command
func NewRootCmd(factory bbUtil.Factory) (*cobra.Command, error) {
	var (
		cmdUse   = `bbctl`
		cmdShort = i18n.T(`Big Bang Control command-line tool.`)
		cmdLong  = templates.LongDesc(i18n.T(
			`Big Bang Control command-line tool allows you to run commands against Kubernetes clusters 
			to simplify development, deployment, auditing, and troubleshooting of Big Bang.`))
		cmdExample = templates.Examples(i18n.T(`
			# List all available commands
			bbctl help`))
	)

	cmd := &cobra.Command{
		Use:     cmdUse,
		Short:   cmdShort,
		Long:    cmdLong,
		Example: cmdExample,
	}

	cmd.CompletionOptions.DisableDefaultCmd = false
	cmd.CompletionOptions.DisableNoDescFlag = true
	cmd.CompletionOptions.DisableDescriptions = false

	completionCmd, completionCmdError := NewCompletionCmd(factory)
	if completionCmdError != nil {
		return nil, fmt.Errorf("error retrieving Completion Command: %w", completionCmdError)
	}
	cmd.AddCommand(completionCmd)
	configCmd, configCmdError := config.NewConfigCmd(factory)
	if configCmdError != nil {
		return nil, fmt.Errorf("error retrieving Config Command: %w", configCmdError)
	}
	cmd.AddCommand(configCmd)
	versionCmd, versionCmdError := NewVersionCmd(factory)
	if versionCmdError != nil {
		return nil, fmt.Errorf("error retrieving Version Command: %w", versionCmdError)
	}
	cmd.AddCommand(versionCmd)
	cmd.AddCommand(NewReleasesCmd(factory))
	cmd.AddCommand(NewStatusCmd(factory))
	cmd.AddCommand(NewValuesCmd(factory))
	violationsCmd, violationsCmdError := NewViolationsCmd(factory)
	if violationsCmdError != nil {
		return nil, fmt.Errorf("error retrieving Violations Command: %w", violationsCmdError)
	}
	cmd.AddCommand(violationsCmd)
	policiesCmd, policiesCmdError := NewPoliciesCmd(factory)
	if policiesCmdError != nil {
		return nil, fmt.Errorf("error retrieving Policies Command: %w", policiesCmdError)
	}
	cmd.AddCommand(policiesCmd)
	preflightCheckCmd, preflightCheckCmdError := NewPreflightCheckCmd(factory)
	if preflightCheckCmdError != nil {
		return nil, fmt.Errorf("error retrieving PreflightCheck Command: %w", preflightCheckCmdError)
	}
	cmd.AddCommand(preflightCheckCmd)
	k3dCmd, K3dCmdError := k3d.NewK3dCmd(factory)
	if K3dCmdError != nil {
		return nil, fmt.Errorf("error retrieving k3d Command: %w", K3dCmdError)
	}
	cmd.AddCommand(k3dCmd)
	deployCmd, deployCmdError := deploy.NewDeployCmd(factory)
	if deployCmdError != nil {
		return nil, fmt.Errorf("error retrieving Deploy Command: %w", deployCmdError)
	}
	cmd.AddCommand(deployCmd)
	cmd.AddCommand(update.NewUpdateCmd(factory))

	return cmd, nil
}
