package cmd

import (
	"fmt"
	"strings"

	"repo1.dso.mil/big-bang/product/packages/bbctl/static"
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
	helm "repo1.dso.mil/big-bang/product/packages/bbctl/util/helm"
	oc "repo1.dso.mil/big-bang/product/packages/bbctl/util/output"

	"github.com/spf13/cobra"
	"helm.sh/helm/v3/cmd/helm/require"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

// valuesCmdHelper is a structure for storing shared clients, values, and methods used in the values command
type valuesCmdHelper struct {
	// Clients
	constantsClient static.ConstantsClient
	helmClient      helm.Client
	outputClient    oc.Client

	// Values
	constants static.Constants
}

// newValuesCmdHelper returns a valuesCmdHelper with the default constants
func newValuesCmdHelper(cmd *cobra.Command, factory bbUtil.Factory, constantsClient static.ConstantsClient) (*valuesCmdHelper, error) {
	constants, err := constantsClient.GetConstants()
	if err != nil {
		return nil, err
	}

	helmClient, err := factory.GetHelmClient(cmd, constants.BigBangNamespace)
	if err != nil {
		return nil, err
	}

	outputClient, outClientErr := factory.GetOutputClient(cmd)
	if outClientErr != nil {
		return nil, fmt.Errorf("error getting output client: %w", outClientErr)
	}

	return &valuesCmdHelper{
		constantsClient: constantsClient,
		helmClient:      helmClient,
		outputClient:    outputClient,
		constants:       constants,
	}, nil
}

// NewValuesCmd returns a new values command
func NewValuesCmd(factory bbUtil.Factory) *cobra.Command {
	var (
		valuesUse   = `values RELEASE_NAME`
		valuesShort = i18n.T(`Get all the values for a given release deployed by Big Bang.`)
		valuesLong  = templates.LongDesc(i18n.T(`Get all the values for a given release deployed by Big Bang.
			Running this comamnd is the equivalent of running "helm -n bigbang get values RELEASE_NAME".
	
			This command only looks for releases in the namespace in which the Big Bang umbrella chart is deployed.
		`))
		valuesExample = templates.Examples(i18n.T(`
			# Get values for a helm release deployed by Big Bang
			bbctl values RELEASE_NAME`))
	)

	cmd := &cobra.Command{
		Use:     valuesUse,
		Short:   valuesShort,
		Long:    valuesLong,
		Example: valuesExample,
		Args:    require.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, hint string) ([]string, cobra.ShellCompDirective) {
			// If we fail to get the values helper client, we should return an error
			// as the command will not work
			v, err := newValuesCmdHelper(cmd, factory, static.DefaultClient)
			if err != nil {
				return nil, cobra.ShellCompDirectiveError
			}
			// We shouldn't try and attempt to continue completing as values only takes a single argument
			// But we also don't want completion to begin suggesting file names
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return v.matchingReleaseNames(hint)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := newValuesCmdHelper(cmd, factory, static.DefaultClient)
			if err != nil {
				return err
			}
			return v.getHelmValues(args[0])
		},
	}

	return cmd
}

// getHelmValues queries the cluster using the helm module to get information on big bang release values
func (v *valuesCmdHelper) getHelmValues(name string) error {
	// use helm get values to get release values
	releases, err := v.helmClient.GetValues(name)
	if err != nil {
		return fmt.Errorf("error getting helm release values in namespace %s: %s",
			v.constants.BigBangNamespace, err.Error())
	}
	return v.outputClient.Output(&oc.BasicOutput{Vals: releases})
}

// matchingReleaseNames searches the helm releases with given prefix for command completion
func (v *valuesCmdHelper) matchingReleaseNames(hint string) ([]string, cobra.ShellCompDirective) {
	// use helm list to get detailed information on charts deployed by bigbang
	releases, err := v.helmClient.GetList()
	if err != nil {
		return nil, cobra.ShellCompDirectiveDefault
	}

	var matches = make([]string, 0)

	for _, r := range releases {
		if strings.HasPrefix(r.Name, hint) {
			matches = append(matches, r.Name)
		}
	}

	return matches, cobra.ShellCompDirectiveNoFileComp
}
