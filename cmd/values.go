package cmd

import (
	"fmt"
	"strings"

	"repo1.dso.mil/big-bang/product/packages/bbctl/static"
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"

	"github.com/spf13/cobra"
	"helm.sh/helm/v3/cmd/helm/require"
	"helm.sh/helm/v3/pkg/cli/output"
	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"
	cmdUtil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	valuesUse = `values RELEASE_NAME`

	valuesShort = i18n.T(`Get all the values for a given release deployed by BigBang.`)

	valuesLong = templates.LongDesc(i18n.T(`Get all the values for a given release deployed by BigBang.`))

	valuesExample = templates.Examples(i18n.T(`
		# Get values for a helm releases in bigbang namespace 
		# (equivalent of helm -n bigbang get values <RELEASE_NAME>)
		bbctl values RELEASE_NAME`))
)

// NewValuesCmd - new values command
func NewValuesCmd(factory bbUtil.Factory, streams genericIOOptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     valuesUse,
		Short:   valuesShort,
		Long:    valuesLong,
		Example: valuesExample,
		Args:    require.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, hint string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return matchingReleaseNames(cmd, factory, hint)
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmdUtil.CheckErr(getHelmValues(cmd, factory, streams, args[0]))
		},
	}

	return cmd
}

// query the cluster using helm module to get information on big bang release values
func getHelmValues(cmd *cobra.Command, factory bbUtil.Factory, streams genericIOOptions.IOStreams, name string) error {
	constants, err := static.GetConstants()
	if err != nil {
		return err
	}
	client, err := factory.GetHelmClient(cmd, constants.BigBangNamespace)
	if err != nil {
		return err
	}

	// use helm get values to get release values
	releases, err := client.GetValues(name)
	if err != nil {
		return fmt.Errorf("error getting helm release values in namespace %s: %s",
			constants.BigBangNamespace, err.Error())
	}

	return output.EncodeYAML(streams.Out, releases)
}

// find helm releases with given prefix for command completion
func matchingReleaseNames(cmd *cobra.Command, factory bbUtil.Factory, hint string) ([]string, cobra.ShellCompDirective) {
	constants, err := static.GetConstants()
	if err != nil {
		return nil, cobra.ShellCompDirectiveDefault
	}
	client, err := factory.GetHelmClient(cmd, constants.BigBangNamespace)
	if err != nil {
		return nil, cobra.ShellCompDirectiveDefault
	}

	// use helm list to get detailed information on charts deployed by bigbang
	releases, err := client.GetList()
	if err != nil {
		return nil, cobra.ShellCompDirectiveDefault
	}

	var matches []string = make([]string, 0)

	for _, r := range releases {
		if strings.HasPrefix(r.Name, hint) {
			matches = append(matches, r.Name)
		}
	}

	return matches, cobra.ShellCompDirectiveNoFileComp
}
