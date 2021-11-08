package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"helm.sh/helm/v3/cmd/helm/require"
	"helm.sh/helm/v3/pkg/cli/output"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	bbutil "repo1.dso.mil/platform-one/big-bang/apps/product-tools/bbctl/util"
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

func NewGetValuesCmd(factory bbutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {

	cmd := &cobra.Command{
		Use:     valuesUse,
		Short:   valuesShort,
		Long:    valuesLong,
		Example: valuesExample,
		Args:    require.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(getHelmValues(factory, streams, cmd.Flags(), args[0]))
		},
	}

	cmd.Flags().BoolP("all", "a", false, "dump all (computed) values")

	return cmd
}

// query the cluster using helm module to get information on bigbang release values
func getHelmValues(factory bbutil.Factory, streams genericclioptions.IOStreams, flags *pflag.FlagSet, name string) error {

	client, err := factory.GetHelmClient(BigBangNamespace)
	if err != nil {
		return err
	}

	allValues, _ := flags.GetBool("all")

	// use helm get values to get release values
	releases, err := client.GetValues(name, allValues)
	if err != nil {
		return fmt.Errorf("error getting helm release values in namespace %s: %s",
			BigBangNamespace, err.Error())
	}

	return output.EncodeYAML(streams.Out, releases)
}
