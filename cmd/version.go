package cmd

import (
	"fmt"

	"repo1.dso.mil/big-bang/product/packages/bbctl/static"
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"

	"github.com/spf13/cobra"
	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"
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
		bbctl version --client`))
)

// NewVersionCmd - new version command
func NewVersionCmd(factory bbUtil.Factory, streams genericIOOptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     versionUse,
		Short:   versionShort,
		Long:    versionLong,
		Example: versionExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return bbVersion(cmd, factory, streams)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
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
	constants, err := static.GetDefaultConstants()
	if err != nil {
		return err
	}
	fmt.Fprintf(streams.Out, "bigbang cli version %s\n", constants.BigBangCliVersion)

	// Config client error handling is done in the public NewVersionCmd function above
	configClient, _ := factory.GetConfigClient(cmd)
	config := configClient.GetConfig()

	if config.VersionConfiguration.Client {
		return nil
	}

	client, err := factory.GetHelmClient(cmd, constants.BigBangNamespace)
	if err != nil {
		return err
	}

	release, err := client.GetRelease(constants.BigBangHelmReleaseName)
	if err != nil {
		return fmt.Errorf("error getting helm information for release %s in namespace %s: %s",
			constants.BigBangHelmReleaseName, constants.BigBangNamespace, err.Error())
	}

	fmt.Fprintf(streams.Out, "%s release version %s\n", release.Chart.Metadata.Name, release.Chart.Metadata.Version)

	return nil
}
