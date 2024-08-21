package deploy

import (
	"fmt"
	"path"
	"slices"

	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/config/schemas"
)

var (
	bigBangUse = `bigbang`

	bigBangShort = i18n.T(`Deploy Big Bang components to your cluster`)

	bigBangLong = templates.LongDesc(i18n.T(`Deploy Big Bang and optional Big Bang addons to your cluster.
		This command invokes the helm command, so arguments after -- are passed to the underlying helm command.

		Note: deployment of Big Bang requires Flux to have been deployed to your cluster. See "bbctl deploy flux" for more information.
	`))

	bigBangExample = templates.Examples(i18n.T(`
	    # Deploy Big Bang to your cluster
		bbctl deploy bigbang

		# Deploy Big Bang with additional configurations for a k3d development cluster
		bbctl deploy bigbang --k3d

		# Deploy Big Bang with a helm overrides file. All arguments after -- are passed to the underlying helm command
		bbctl deploy bigbang -- -f ../path/to/overrides/values.yaml
		`))
)

// NewDeployBigBangCmd - deploy Big Bang to your cluster
func NewDeployBigBangCmd(factory bbUtil.Factory) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:     bigBangUse,
		Short:   bigBangShort,
		Long:    bigBangLong,
		Example: bigBangExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return deployBigBangToCluster(cmd, factory, args)
		},
	}

	configClient, clientError := factory.GetConfigClient(cmd)
	if clientError != nil {
		return nil, fmt.Errorf("Unable to get config client: %w", clientError)
	}

	k3dFlagError := configClient.SetAndBindFlag(
		"k3d",
		"",
		false,
		"Include some boilerplate suitable for deploying into k3d",
	)
	if k3dFlagError != nil {
		return nil, fmt.Errorf("Error setting k3d flag: %w", k3dFlagError)
	}

	addOnFlagError := configClient.SetAndBindFlag(
		"addon",
		"",
		[]string(nil),
		"Enable this Big Bang addon in the deployment",
	)
	if addOnFlagError != nil {
		return nil, fmt.Errorf("error setting addon flag: %w", addOnFlagError)
	}

	return cmd, nil
}

func getChartRelativePath(configClient *schemas.GlobalConfiguration, pathCmp ...string) string {
	repoPath := configClient.BigBangRepo
	return path.Join(slices.Insert(pathCmp, 0, repoPath)...)
}

func insertHelmOptForExampleConfig(config *schemas.GlobalConfiguration, helmOpts []string, chartName string) []string {
	return slices.Insert(helmOpts,
		0,
		"-f",
		getChartRelativePath(
			config,
			"docs",
			"assets",
			"configs",
			"example",
			chartName,
		),
	)
}

func insertHelmOptForRelativeChart(config *schemas.GlobalConfiguration, helmOpts []string, chartName string) []string {
	return slices.Insert(helmOpts,
		0,
		"-f",
		getChartRelativePath(
			config,
			"chart",
			chartName,
		),
	)
}

func deployBigBangToCluster(command *cobra.Command, factory bbUtil.Factory, args []string) error {
	loggingClient, err := factory.GetLoggingClient()
	if err != nil {
		return err
	}
	configClient, err := factory.GetConfigClient(command)
	if err != nil {
		return err
	}
	config, configErr := configClient.GetConfig()
	if configErr != nil {
		return fmt.Errorf("error getting config: %w", configErr)
	}
	credentialHelper, err := factory.GetCredentialHelper()
	if err != nil {
		return fmt.Errorf("unable to get credential helper: %w", err)
	}
	username, err := credentialHelper("username", "registry1.dso.mil")
	if err != nil {
		return fmt.Errorf("unable to get username: %w", err)
	}
	password, err := credentialHelper("password", "registry1.dso.mil")
	if err != nil {
		return fmt.Errorf("unable to get password: %w", err)
	}

	chartPath := getChartRelativePath(config, "chart")
	helmOpts := slices.Clone(args)
	loggingClient.Info(fmt.Sprintf("preparing to deploy Big Bang to cluster, k3d=%v", config.DeployBigBangConfiguration.K3d))
	if config.DeployBigBangConfiguration.K3d {
		loggingClient.Info("Using k3d configuration")
		helmOpts = insertHelmOptForExampleConfig(config, helmOpts, "policy-overrides-k3d.yaml")
		helmOpts = insertHelmOptForRelativeChart(config, helmOpts, "ingress-certs.yaml")
	}
	for _, x := range config.DeployBigBangConfiguration.Addon {
		helmOpts = slices.Insert(helmOpts,
			0,
			"--set",
			fmt.Sprintf("addons.%s.enabled=true", x),
		)
	}
	helmOpts = slices.Insert(helmOpts,
		0,
		"upgrade",
		"-i",
		"bigbang",
		chartPath,
		"-n",
		"bigbang",
		"--create-namespace",
		"--set",
		fmt.Sprintf("registryCredentials.username=%v", username),
		"--set",
		fmt.Sprintf("registryCredentials.password=%v", password),
	)

	streams, err := factory.GetIOStream()
	if err != nil {
		return fmt.Errorf("Unable to create IO streams: %w", err)
	}
	cmd, err := factory.GetCommandWrapper("helm", helmOpts...)
	if err != nil {
		return fmt.Errorf("Unable to get command wrapper: %w", err)
	}
	cmd.SetStdout(streams.Out)
	cmd.SetStderr(streams.ErrOut)
	err = cmd.Run()
	return err
}
