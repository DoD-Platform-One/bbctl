package cmd

import (
	"fmt"

	"repo1.dso.mil/big-bang/product/packages/bbctl/static"
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
	outputSchema "repo1.dso.mil/big-bang/product/packages/bbctl/util/output/schemas"

	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

// NewReleasesCmd creates a new command for listing new releases.
//
// Returns a cobra.Command configured to list releases.
func NewReleasesCmd(factory bbUtil.Factory) *cobra.Command {
	var (
		listUse   = `list`
		listShort = i18n.T(`List all the helm releases deployed by Big Bang.`)
		listLong  = templates.LongDesc(i18n.T(`List all the helm releases deployed by Big Bang.
	
		This command queries the cluster and displays information about all helm releases in the bigbang namespace.
	
		It displays information including Name, Namespace, Revision, Status, Chart, and Appversion.
		`))
		listExample = templates.Examples(i18n.T(`
			# Get a list of helm releases in bigbang namespace 
			# (equivalent of helm -n bigbang ls)
			bbctl list`))
	)

	cmd := &cobra.Command{
		Use:     listUse,
		Short:   listShort,
		Long:    listLong,
		Example: listExample,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return listHelmReleases(cmd, factory, static.DefaultClient)
		},
	}

	return cmd
}

// listHelmReleases queries the cluster and retrieves information about helm releases in the bigbang namespace
//
// Returns an error if the release information could not be found
func listHelmReleases(cmd *cobra.Command, factory bbUtil.Factory, constantClient static.ConstantsClient) error {
	constants, err := constantClient.GetConstants()
	if err != nil {
		return err
	}

	client, err := factory.GetHelmClient(cmd, constants.BigBangNamespace)
	if err != nil {
		return err
	}

	// use helm list to get detailed information on charts deployed by bigbang
	releases, err := client.GetList()
	if err != nil {
		return fmt.Errorf("error getting helm releases in namespace %s: %s",
			constants.BigBangNamespace, err.Error())
	}

	outputClient, outClientErr := factory.GetOutputClient(cmd)
	if outClientErr != nil {
		return fmt.Errorf("error getting output client: %w", outClientErr)
	}

	var tableOutput outputSchema.HelmReleaseTableOutput
	for _, r := range releases {
		releaseOutput := outputSchema.HelmReleaseOutput{
			Name:       r.Name,
			Namespace:  r.Namespace,
			Revision:   r.Version,
			Status:     r.Info.Status.String(),
			Chart:      fmt.Sprintf("%s-%s", r.Chart.Metadata.Name, r.Chart.Metadata.Version),
			AppVersion: r.Chart.Metadata.AppVersion,
		}

		tableOutput.Releases = append(tableOutput.Releases, releaseOutput)
	}

	outputErr := outputClient.Output(&tableOutput)
	if outputErr != nil {
		return fmt.Errorf("error marshaling Helm release output: %w", outputErr)
	}
	return nil
}
