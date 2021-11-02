package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	helm "repo1.dso.mil/platform-one/big-bang/apps/product-tools/bbctl/util/helm"
	bbk8sutil "repo1.dso.mil/platform-one/big-bang/apps/product-tools/bbctl/util/k8s"
)

var (
	versionUse     = `version`
	versionShort   = `Print BigBang Deployment and BigBang CLI version.`
	versionLong    = `Print BigBang Deployment and BigBang CLI version.`
	versionExample = `
		# Print version  
		bbctl version`
)

// versionCmd represents the version subcommand
var versionCmd = &cobra.Command{
	Use:     versionUse,
	Short:   versionShort,
	Long:    versionLong,
	Example: versionExample,
	Run: func(cmd *cobra.Command, args []string) {
		flags := cmd.Flags()
		cmdutil.CheckErr(bbVersion(flags, bbk8sutil.GetIOStream()))
	},
}

func init() {
	bbctlCmd.AddCommand(versionCmd)
	versionCmd.Flags().BoolP("client", "c", false, "Print bbctl version only")
}

// query the cluster using helm module to get information on bigbang release
func bbVersion(flags *pflag.FlagSet, streams genericclioptions.IOStreams) error {

	clientVersionOnly, _ := flags.GetBool("client")
	fmt.Fprintf(streams.Out, "bigbang cli version %s\n", BigBangCliVersion)

	if clientVersionOnly {
		return nil
	}

	config, err := bbk8sutil.BuildKubeConfigFromFlags(flags)
	if err != nil {
		return err
	}

	opt := &helm.Options{
		Namespace:        BigBangNamespace,
		RepositoryCache:  "/tmp/.helmcache",
		RepositoryConfig: "/tmp/.helmrepo",
		Debug:            true,
		Linting:          true,
		RestConfig:       config,
	}

	helmClient, err := helm.New(opt)
	if err != nil {
		return err
	}

	release, err := helmClient.GetRelease(BigBangHelmReleaseName)
	if err != nil {
		return fmt.Errorf("error getting helm information for release %s in namespace %s: %s",
			BigBangHelmReleaseName, BigBangNamespace, err.Error())
	}

	fmt.Fprintf(streams.Out, "%s release version %s\n", release.Chart.Metadata.Name, release.Chart.Metadata.Version)

	// use helm list to get detailed information on charts deployed by bigbang
	// releases, _ := helmClient.GetList()

	return nil
}
