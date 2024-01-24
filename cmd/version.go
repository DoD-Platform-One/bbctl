package cmd

import (
	"fmt"

	bbutil "repo1.dso.mil/big-bang/product/packages/bbctl/util"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	versionUse = `version`

	versionShort = i18n.T(`Print BigBang Deployment and BigBang CLI version.`)

	versionLong = templates.LongDesc(i18n.T(`Print BigBang Deployment and BigBang CLI version.`))

	versionExample = templates.Examples(i18n.T(`
		# Print version
		bbctl version
		
		# Print client version only
		bbctl -c`))
)

// NewVersionCmd - new version command
func NewVersionCmd(factory bbutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     versionUse,
		Short:   versionShort,
		Long:    versionLong,
		Example: versionExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(bbVersion(factory, streams, cmd.Flags()))
		},
	}

	cmd.Flags().BoolP("client", "c", false, "Print bbctl version only")
	return cmd
}

// query the cluster using helm module to get information on bigbang release
func bbVersion(factory bbutil.Factory, streams genericclioptions.IOStreams, flags *pflag.FlagSet) error {

	clientVersionOnly, _ := flags.GetBool("client")
	fmt.Fprintf(streams.Out, "bigbang cli version %s\n", BigBangCliVersion)

	if clientVersionOnly {
		return nil
	}

	client, err := factory.GetHelmClient(BigBangNamespace)
	if err != nil {
		return err
	}

	release, err := client.GetRelease(BigBangHelmReleaseName)
	if err != nil {
		return fmt.Errorf("error getting helm information for release %s in namespace %s: %s",
			BigBangHelmReleaseName, BigBangNamespace, err.Error())
	}

	fmt.Fprintf(streams.Out, "%s release version %s\n", release.Chart.Metadata.Name, release.Chart.Metadata.Version)

	return nil
}
