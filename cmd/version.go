package cmd

import (
	"fmt"

	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"

	"github.com/spf13/cobra"
	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"
	cmdUtil "k8s.io/kubectl/pkg/cmd/util"
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
		bbctl --client`))
)

// NewVersionCmd - new version command
func NewVersionCmd(factory bbUtil.Factory, streams genericIOOptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     versionUse,
		Short:   versionShort,
		Long:    versionLong,
		Example: versionExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdUtil.CheckErr(bbVersion(cmd, factory, streams))
		},
	}

	loggingClient := factory.GetLoggingClient()
	configClient, err := factory.GetConfigClient(cmd)
	loggingClient.HandleError("error getting config client: %v", err)

	loggingClient.HandleError(
		"error setting and binding flags: %v",
		configClient.SetAndBindFlag(
			"client",
			false,
			"Print bbctl version only",
		),
	)

	return cmd
}

// query the cluster using helm module to get information on bigbang release
func bbVersion(cmd *cobra.Command, factory bbUtil.Factory, streams genericIOOptions.IOStreams) error {
	fmt.Fprintf(streams.Out, "bigbang cli version %s\n", BigBangCliVersion)
	configClient, err := factory.GetConfigClient(cmd)
	if err != nil {
		return err
	}
	config := configClient.GetConfig()

	if config.VersionConfiguration.Client {
		return nil
	}

	client, err := factory.GetHelmClient(cmd, BigBangNamespace)
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
