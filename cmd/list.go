package cmd

import (
	"fmt"

	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"

	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/cli/output"
	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"
	cmdUtil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	listUse = `list`

	listShort = i18n.T(`List all the helm releases deployed by BigBang.`)

	listLong = templates.LongDesc(i18n.T(`List all the helm releases deployed by BigBang.`))

	listExample = templates.Examples(i18n.T(`
		# Get a list of helm releases in bigbang namespace 
		# (equivalent of helm -n bigbang ls)
		bbctl list`))
)

// NewReleasesCmd - new releases command
func NewReleasesCmd(factory bbUtil.Factory, streams genericIOOptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     listUse,
		Short:   listShort,
		Long:    listLong,
		Example: listExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdUtil.CheckErr(listHelmReleases(factory, streams))
		},
	}

	return cmd
}

// query the cluster using helm module to get information on bigbang release
func listHelmReleases(factory bbUtil.Factory, streams genericIOOptions.IOStreams) error {
	client, err := factory.GetHelmClient(BigBangNamespace)
	if err != nil {
		return err
	}

	// use helm list to get detailed information on charts deployed by bigbang
	releases, err := client.GetList()
	if err != nil {
		return fmt.Errorf("error getting helm releases in namespace %s: %s",
			BigBangNamespace, err.Error())
	}

	table := uitable.New()
	table.AddRow("NAME", "NAMESPACE", "REVISION", "STATUS", "CHART", "APPVERSION")
	for _, r := range releases {
		chart := fmt.Sprintf("%s-%s", r.Chart.Metadata.Name, r.Chart.Metadata.Version)
		table.AddRow(r.Name, r.Namespace, r.Version, r.Info.Status, chart, r.Chart.Metadata.AppVersion)
	}

	return output.EncodeTable(streams.Out, table)
}
