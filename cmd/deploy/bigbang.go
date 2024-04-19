package deploy

import (
	"fmt"
	"path"
	"slices"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"
	cmdUtil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
)

var (
	bigBangUse = `bigbang`

	bigBangShort = i18n.T(`deploy big bang components to your cluster`)

	bigBangLong = templates.LongDesc(i18n.T(`Deploy big bang components to your cluster`))

	bigBangExample = templates.Examples(i18n.T(`
	    # Deploy big bang product to your cluster
		bbctl deploy big bang
		`))

	bbUseK3d bool

	bbAddonList []string
)

// NewDeployBigBangCmd - deploy big bang to your cluster
func NewDeployBigBangCmd(factory bbUtil.Factory, streams genericIOOptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     bigBangUse,
		Short:   bigBangShort,
		Long:    bigBangLong,
		Example: bigBangExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdUtil.CheckErr(deployBigBangToCluster(factory, streams, args))
		},
	}

	cmd.Flags().BoolVar(&bbUseK3d,
		"k3d",
		false,
		"Include some boilerplate suitable for deploying into k3d")
	cmd.Flags().StringSliceVar(&bbAddonList,
		"addon",
		nil,
		"Enable this bigbang addon in the deployment")

	return cmd
}

func getChartRelativePath(factory bbUtil.Factory, pathCmp ...string) string {
	repoPath := viper.GetString("big-bang-repo")
	if repoPath == "" {
		factory.GetLoggingClient().Error("Big bang repository location not defined (\"big-bang-repo\")")
	}
	return path.Join(slices.Insert(pathCmp, 0, repoPath)...)
}

func insertHelmOptForExampleConfig(factory bbUtil.Factory, helmOpts []string, chartName string) []string {
	return slices.Insert(helmOpts,
		0,
		"-f",
		getChartRelativePath(factory,
			"docs",
			"assets",
			"configs",
			"example",
			chartName,
		),
	)
}

func insertHelmOptForRelativeChart(factory bbUtil.Factory, helmOpts []string, chartName string) []string {
	return slices.Insert(helmOpts,
		0,
		"-f",
		getChartRelativePath(factory,
			"chart",
			chartName,
		),
	)
}

func deployBigBangToCluster(factory bbUtil.Factory, streams genericIOOptions.IOStreams, args []string) error {
	credentialHelper := factory.GetCredentialHelper()
	username := credentialHelper("username", "registry1.dso.mil")
	password := credentialHelper("password", "registry1.dso.mil")

	chartPath := getChartRelativePath(factory, "chart")
	helmOpts := slices.Clone(args)
	if bbUseK3d {
		helmOpts = insertHelmOptForExampleConfig(factory, helmOpts, "policy-overrides-k3d.yaml")
		helmOpts = insertHelmOptForRelativeChart(factory, helmOpts, "ingress-certs.yaml")
	}
	for _, x := range bbAddonList {
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
	err := cmd.Run()
	return err
}
