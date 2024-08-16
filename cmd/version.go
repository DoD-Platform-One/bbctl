package cmd

import (
	"fmt"

	"repo1.dso.mil/big-bang/product/packages/bbctl/static"
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/output"

	"github.com/spf13/cobra"
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
func NewVersionCmd(factory bbUtil.Factory) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:     versionUse,
		Short:   versionShort,
		Long:    versionLong,
		Example: versionExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return bbVersion(cmd, factory)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	configClient, clientError := factory.GetConfigClient(cmd)
	if clientError != nil {
		return nil, fmt.Errorf("unable to get config client: %w", clientError)
	}

	flagError := configClient.SetAndBindFlag(
		"client",
		false,
		"Print the bbctl client version only",
	)
	if flagError != nil {
		return nil, fmt.Errorf("error setting and binding flags: %w", flagError)
	}

	return cmd, nil
}

// bbVersion queries the cluster using helm module to get information on Big Bang release
func bbVersion(cmd *cobra.Command, factory bbUtil.Factory) error {
	constants, err := static.GetDefaultConstants()
	if err != nil {
		return err
	}

	// Config client error handling is done in the public NewVersionCmd function above
	configClient, _ := factory.GetConfigClient(cmd)
	config, configErr := configClient.GetConfig()
	if configErr != nil {
		return fmt.Errorf("error getting config: %w", configErr)
	}

	outputClient, outClientErr := factory.GetOutputClient(cmd)
	if outClientErr != nil {
		return fmt.Errorf("error getting output client: %w", outClientErr)
	}

	if config.VersionConfiguration.Client {
		bbctlKey := "bbctl client version"
		outputMap := map[string]interface{}{
			bbctlKey: constants.BigBangCliVersion,
		}
		outputErr := outputClient.Output(
			&output.BasicOutput{
				Vals: outputMap,
			})
		if outputErr != nil {
			return fmt.Errorf("error marshaling %s: %w", bbctlKey, outputErr)
		}
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

	bbctlKey := "bbctl client version"
	bbKey := fmt.Sprintf("%s release version", release.Chart.Metadata.Name)
	outputMap := map[string]interface{}{
		bbctlKey: constants.BigBangCliVersion,
		bbKey:    release.Chart.Metadata.Version,
	}
	outputErr := outputClient.Output(
		&output.BasicOutput{
			Vals: outputMap,
		})
	if outputErr != nil {
		return fmt.Errorf("error marshaling %s and %s: %w", bbctlKey, bbKey, outputErr)
	}

	return nil
}
