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

	versionShort = i18n.T(`Print the current bbctl client version and the version of the Big Bang currently deployed.`)

	versionLong = templates.LongDesc(i18n.T(`Print the version of the bbctl client and the version of Big Bang currently deployed.
	 The Big Bang deployment version is pulled from the cluster currently referenced by your KUBECONFIG setting if no cluster parameters are provided.
	 Using the --client flag will only return the bbctl client version.`))

	versionExample = templates.Examples(i18n.T(`
		# Print version
		bbctl version
		
		# Print the bbctl client version only
		bbctl version --client`))
)

// NewVersionCmd - Creates a new Cobra command which implements the `bbctl version` functionality
func NewVersionCmd(factory bbUtil.Factory, streams genericIOOptions.IOStreams) (*cobra.Command, error) {
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

	configClient, clientError := factory.GetConfigClient(cmd)
	if clientError != nil {
		return nil, fmt.Errorf("Unable to get config client: %w", clientError)
	}

	flagError := configClient.SetAndBindFlag(
		"client",
		false,
		"Print the bbctl client version only",
	)
	if flagError != nil {
		return nil, fmt.Errorf("Error setting and binding flags: %w", flagError)
	}

	return cmd, nil
}

// bbVersion queries the cluster using helm module to get information on Big Bang release
func bbVersion(cmd *cobra.Command, factory bbUtil.Factory, streams genericIOOptions.IOStreams) error {
	constants, err := static.GetDefaultConstants()
	if err != nil {
		return err
	}
	fmt.Fprintf(streams.Out, "bbctl client version %s\n", constants.BigBangCliVersion)

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
