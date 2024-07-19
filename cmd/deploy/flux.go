package deploy

import (
	"fmt"
	"path"
	"slices"

	"github.com/spf13/cobra"
	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"
	cmdUtil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
)

var (
	fluxUse     = `flux`
	fluxShort   = i18n.T(`Deploy flux to your kubernetes cluster`)
	fluxLong    = templates.LongDesc(i18n.T(`Deploy flux to your kubernetes cluster in a way specifically designed to support the deployment of Big Bang`))
	fluxExample = templates.Examples(i18n.T(`# Deploy flux to your cluster
		bbctl deploy flux`))
)

// NewDeployFluxCmd - parent for deploy commands
func NewDeployFluxCmd(factory bbUtil.Factory, streams genericIOOptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     fluxUse,
		Short:   fluxShort,
		Long:    fluxLong,
		Example: fluxExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdUtil.CheckErr(deployFluxToCluster(factory, cmd, streams, args))
		},
	}

	return cmd
}

func deployFluxToCluster(factory bbUtil.Factory, command *cobra.Command, streams genericIOOptions.IOStreams, args []string) error {
	loggingClient := factory.GetLoggingClient()
	configClient, err := factory.GetConfigClient(command)
	if err != nil {
		return err
	}
	loggingClient.HandleError("error getting config client: %v", err)
	config := configClient.GetConfig()
	credentialHelper := factory.GetCredentialHelper()
	username, err := credentialHelper("username", "registry1.dso.mil")
	if err != nil {
		return fmt.Errorf("unable to get username: %w", err)
	}
	password, err := credentialHelper("password", "registry1.dso.mil")
	if err != nil {
		return fmt.Errorf("unable to get password: %w", err)
	}
	installFluxPath := path.Join(config.BigBangRepo,
		"scripts",
		"install_flux.sh",
	)
	fluxArgs := slices.Clone(args)
	fluxArgs = append(fluxArgs,
		"-u",
		username,
		"-p",
		password,
	)
	cmd := factory.GetCommandWrapper(installFluxPath, fluxArgs...)
	cmd.SetStdout(streams.Out)
	cmd.SetStderr(streams.ErrOut)
	err = cmd.Run()
	return err
}
