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
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/config/schemas"
)

var (
	bigBangUse = `bigbang`

	bigBangShort = i18n.T(`deploy big bang components to your cluster`)

	bigBangLong = templates.LongDesc(i18n.T(`Deploy big bang components to your cluster`))

	bigBangExample = templates.Examples(i18n.T(`
	    # Deploy big bang product to your cluster
		bbctl deploy big bang
		`))
)

// NewDeployBigBangCmd - deploy big bang to your cluster
func NewDeployBigBangCmd(factory bbUtil.Factory, streams genericIOOptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     bigBangUse,
		Short:   bigBangShort,
		Long:    bigBangLong,
		Example: bigBangExample,
		Run: func(command *cobra.Command, args []string) {
			cmdUtil.CheckErr(deployBigBangToCluster(command, factory, streams, args))
		},
	}

	loggingClient := factory.GetLoggingClient()
	configClient, err := factory.GetConfigClient(cmd)
	loggingClient.HandleError("error getting config client", err)
	loggingClient.HandleError(
		"error setting k3dflag",
		configClient.SetAndBindFlag(
			"k3d",
			false,
			"Include some boilerplate suitable for deploying into k3d",
		),
	)
	loggingClient.HandleError(
		"error setting addon flag",
		configClient.SetAndBindFlag(
			"addon",
			[]string(nil),
			"Enable this bigbang addon in the deployment",
		),
	)

	return cmd
}

func getChartRelativePath(factory bbUtil.Factory, configClient *schemas.GlobalConfiguration, pathCmp ...string) string {
	repoPath := configClient.BigBangRepo
	return path.Join(slices.Insert(pathCmp, 0, repoPath)...)
}

func insertHelmOptForExampleConfig(factory bbUtil.Factory, config *schemas.GlobalConfiguration, helmOpts []string, chartName string) []string {
	return slices.Insert(helmOpts,
		0,
		"-f",
		getChartRelativePath(
			factory,
			config,
			"docs",
			"assets",
			"configs",
			"example",
			chartName,
		),
	)
}

func insertHelmOptForRelativeChart(factory bbUtil.Factory, config *schemas.GlobalConfiguration, helmOpts []string, chartName string) []string {
	return slices.Insert(helmOpts,
		0,
		"-f",
		getChartRelativePath(factory,
			config,
			"chart",
			chartName,
		),
	)
}

func deployBigBangToCluster(command *cobra.Command, factory bbUtil.Factory, streams genericIOOptions.IOStreams, args []string) error {
	loggingClient := factory.GetLoggingClient()
	configClient, err := factory.GetConfigClient(command)
	if err != nil {
		return err
	}
	config := configClient.GetConfig(factory.GetViper())
	credentialHelper := factory.GetCredentialHelper()
	username := credentialHelper("username", "registry1.dso.mil")
	password := credentialHelper("password", "registry1.dso.mil")

	chartPath := getChartRelativePath(factory, config, "chart")
	helmOpts := slices.Clone(args)
	loggingClient.Info(fmt.Sprintf("preparing to deploy big bang to cluster, k3d=%v", config.DeployBigBangConfiguration.K3d))
	if config.DeployBigBangConfiguration.K3d {
		loggingClient.Info("Using k3d configuration")
		helmOpts = insertHelmOptForExampleConfig(factory, config, helmOpts, "policy-overrides-k3d.yaml")
		helmOpts = insertHelmOptForRelativeChart(factory, config, helmOpts, "ingress-certs.yaml")
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

	cmd := factory.GetCommandWrapper("helm", helmOpts...)
	cmd.SetStdout(streams.Out)
	cmd.SetStderr(streams.ErrOut)
	err = cmd.Run()
	return err
}
