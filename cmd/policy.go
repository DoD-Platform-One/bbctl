package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/gatekeeper"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/kyverno"
	outputSchema "repo1.dso.mil/big-bang/product/packages/bbctl/util/output/schemas"

	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

type policyDescriptor struct {
	name      string // policy name
	namespace string // policy namespace (kyverno policy)
	kind      string // policy kind
	desc      string // policy description
	action    string // enforcement action
}

// NewPoliciesCmd - Creates a new Cobra command which implements the `bbctl policy` functionality
func NewPoliciesCmd(factory bbUtil.Factory) (*cobra.Command, error) {
	var (
		policyUse   = `policy --PROVIDER CONSTRAINT_NAME`
		policyShort = i18n.T(`Describe the configured policies implemented by Gatekeeper or Kyverno.`)
		policyLong  = templates.LongDesc(i18n.T(`
			Describe the configured policies implemented by Gatekeeper or Kyverno.

			Supported values for the required policy provider flag are --gatekeeper and --kyverno.

			The optional constraint name argument can be provided to limit results to policies with the same name.
		`))
		policyExample = templates.Examples(i18n.T(`
			# Describe a secific gatekeeper policy named "restrictedtainttoleration"
			bbctl policy --gatekeeper restrictedtainttoleration
		
			# Get a list of all active gatekeeper policies
			bbctl policy --gatekeeper
			
			# Describe a specific kyverno policy named "restrict-seccomp"
			bbctl policy --kyverno restrict-seccomp
		
			# Get a list of all active kyverno policies
			bbctl policy --kyverno
		`))
	)

	cmd := &cobra.Command{
		Use:     policyUse,
		Short:   policyShort,
		Long:    policyLong,
		Example: policyExample,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, hint string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return matchingPolicyNames(cmd, factory, hint)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var output outputSchema.PolicyListOutput
			var err error

			if len(args) == 1 {
				output, err = listPoliciesByName(cmd, factory, args[0])
			} else {
				output, err = listAllPolicies(cmd, factory)
			}

			if err != nil {
				return err
			}

			outputClient, outClientErr := factory.GetOutputClient(cmd)
			if outClientErr != nil {
				return fmt.Errorf("error getting output client: %w", outClientErr)
			}

			outputErr := outputClient.Output(&output)
			if outputErr != nil {
				return fmt.Errorf("error outputting policies: %w", outputErr)
			}

			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	configClient, clientError := factory.GetConfigClient(cmd)
	if clientError != nil {
		return nil, fmt.Errorf("unable to get config client: %w", clientError)
	}

	gatekeeperError := configClient.SetAndBindFlag("gatekeeper", "", false, "Print gatekeeper policy")
	if gatekeeperError != nil {
		return nil, fmt.Errorf("unable to add flags to command: %w", gatekeeperError)
	}

	kyvernoError := configClient.SetAndBindFlag("kyverno", "", false, "Print kyverno policy")
	if kyvernoError != nil {
		return nil, fmt.Errorf("unable to add flags to command: %w", kyvernoError)
	}

	return cmd, nil
}

// Internal helper function to find policy CRDs matching the given the prefix hint for command completion
func matchingPolicyNames(cmd *cobra.Command, factory bbUtil.Factory, hint string) ([]string, cobra.ShellCompDirective) {
	configClient, err := factory.GetConfigClient(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveDefault
	}
	config, configErr := configClient.GetConfig()
	if configErr != nil {
		return nil, cobra.ShellCompDirectiveDefault
	}

	if config.PolicyConfiguration.Gatekeeper && !config.PolicyConfiguration.Kyverno {
		return matchingGatekeeperPolicyNames(cmd, factory, hint)
	}

	if !config.PolicyConfiguration.Gatekeeper && config.PolicyConfiguration.Kyverno {
		return matchingKyvernoPolicyNames(cmd, factory, hint)
	}
	return nil, cobra.ShellCompDirectiveDefault
}

// Internal helper function to find Gatekeeper policy CRDs matching the given the prefix hint for command completion
func matchingGatekeeperPolicyNames(cmd *cobra.Command, factory bbUtil.Factory, hint string) ([]string, cobra.ShellCompDirective) {
	client, err := factory.GetK8sDynamicClient(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveDefault
	}

	gkCrds, err := gatekeeper.FetchGatekeeperCrds(client)
	if err != nil {
		return nil, cobra.ShellCompDirectiveDefault
	}

	var matches = make([]string, 0)

	for _, crd := range gkCrds.Items {
		crdName, _, _ := unstructured.NestedString(crd.Object, "metadata", "name")
		shortName := strings.Split(crdName, ".")[0]

		if strings.HasPrefix(shortName, hint) {
			matches = append(matches, shortName)
		}
	}

	return matches, cobra.ShellCompDirectiveNoFileComp
}

// Internal helper function to find Kyverno policy CRDs matching the given the prefix hint for command completion
func matchingKyvernoPolicyNames(cmd *cobra.Command, factory bbUtil.Factory, hint string) ([]string, cobra.ShellCompDirective) {
	client, err := factory.GetK8sDynamicClient(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveDefault
	}

	kyvernoCrds, err := kyverno.FetchKyvernoCrds(client)
	if err != nil {
		return nil, cobra.ShellCompDirectiveDefault
	}

	var matches = make([]string, 0)

	for _, crd := range kyvernoCrds.Items {
		crdName, _, _ := unstructured.NestedString(crd.Object, "metadata", "name")
		policies, err := kyverno.FetchKyvernoPolicies(client, crdName)
		if err != nil {
			loggingClient, loggingErr := factory.GetLoggingClient()
			if loggingErr != nil {
				return nil, cobra.ShellCompDirectiveError
			}
			loggingClient.Warn("Error getting kyverno policies: %s", err.Error())
			return nil, cobra.ShellCompDirectiveDefault
		}
		for _, c := range policies.Items {
			name, _, _ := unstructured.NestedString(c.Object, "metadata", "name")
			if strings.HasPrefix(name, hint) {
				matches = append(matches, name)
			}
		}
	}

	return matches, cobra.ShellCompDirectiveNoFileComp
}

// Internal helper function to query the cluster for resources matching the given the prefix hint on the following:
//   - Gatekeeper constraint CRDs
//   - Kyverno cluster policy CRDs
func listPoliciesByName(cmd *cobra.Command, factory bbUtil.Factory, name string) (outputSchema.PolicyListOutput, error) {
	var policyListOutput outputSchema.PolicyListOutput

	configClient, err := factory.GetConfigClient(cmd)
	if err != nil {
		return policyListOutput, fmt.Errorf("unable to get config client: %w", err)
	}
	config, configErr := configClient.GetConfig()
	if configErr != nil {
		return policyListOutput, fmt.Errorf("error getting config: %w", configErr)
	}

	if config.PolicyConfiguration.Gatekeeper && !config.PolicyConfiguration.Kyverno {
		return listGatekeeperPoliciesByName(cmd, factory, name)
	}

	if !config.PolicyConfiguration.Gatekeeper && config.PolicyConfiguration.Kyverno {
		return listKyvernoPoliciesByName(cmd, factory, name)
	}

	return policyListOutput, errors.New("either --gatekeeper or --kyverno must be specified, but not both")
}

// Internal helper function to query the cluster for Gatekeeper constraint CRDs matching the given prefix
func listGatekeeperPoliciesByName(cmd *cobra.Command, factory bbUtil.Factory, name string) (outputSchema.PolicyListOutput, error) {
	policyOutput := outputSchema.PolicyListOutput{}
	client, err := factory.GetK8sDynamicClient(cmd)
	if err != nil {
		return policyOutput, err
	}

	crdName := name + ".constraints.gatekeeper.sh"

	constraints, err := gatekeeper.FetchGatekeeperConstraints(client, crdName)
	if err != nil {
		return policyOutput, err
	}

	crdPolicyOutput := outputSchema.CrdPolicyOutput{
		CrdName: crdName,
	}

	if len(constraints.Items) == 0 {
		policyOutput.Messages = append(policyOutput.Messages, "No constraints found")
	}

	for _, c := range constraints.Items {
		d, err := getGatekeeperPolicyDescriptor(&c)
		if err != nil {
			return policyOutput, err
		}
		policy := outputSchema.PolicyOutput{
			Name:        d.name,
			Namespace:   d.namespace,
			Kind:        d.kind,
			Description: d.desc,
			Action:      d.action,
		}
		crdPolicyOutput.Policies = append(crdPolicyOutput.Policies, policy)
	}
	policyOutput.CrdPolicies = append(policyOutput.CrdPolicies, crdPolicyOutput)

	return policyOutput, nil
}

// Internal helper function to query the cluster for Kyverno cluster policy CRDs matching the given name
func listKyvernoPoliciesByName(cmd *cobra.Command, factory bbUtil.Factory, name string) (outputSchema.PolicyListOutput, error) {
	policyOutput := outputSchema.PolicyListOutput{}
	client, err := factory.GetK8sDynamicClient(cmd)
	if err != nil {
		return policyOutput, err
	}

	kyvernoCrds, err := kyverno.FetchKyvernoCrds(client)
	if err != nil {
		return policyOutput, err
	}

	for _, crd := range kyvernoCrds.Items {
		crdName, _, _ := unstructured.NestedString(crd.Object, "metadata", "name")
		policies, err := kyverno.FetchKyvernoPolicies(client, crdName)
		if err != nil {
			loggingClient, loggingErr := factory.GetLoggingClient()
			if loggingErr != nil {
				return policyOutput, fmt.Errorf("error getting logging client: %w for error: %w", loggingErr, err)
			}
			loggingClient.Warn("Error getting kyverno policies: %s", err.Error())
			return policyOutput, err
		}
		crdPolicyOutput := outputSchema.CrdPolicyOutput{
			CrdName: crdName,
		}
		for _, c := range policies.Items {
			policyName, _, _ := unstructured.NestedString(c.Object, "metadata", "name")
			if policyName == name {
				d, err := getKyvernoPolicyDescriptor(&c)
				if err != nil {
					return policyOutput, err
				}
				policy := outputSchema.PolicyOutput{
					Name:        d.name,
					Namespace:   d.namespace,
					Kind:        d.kind,
					Description: d.desc,
					Action:      d.action,
				}
				crdPolicyOutput.Policies = append(crdPolicyOutput.Policies, policy)
				policyOutput.CrdPolicies = append(policyOutput.CrdPolicies, crdPolicyOutput)
				return policyOutput, nil
			}
		}
	}

	policyOutput.Messages = append(policyOutput.Messages, "No Matching Policy Found")

	return policyOutput, nil
}

// Internal helper function to query the cluster using the dynamic client to get information on the following:
//   - All Gatekeeper constraint CRDs
//   - All Kyverno policy CRDs
func listAllPolicies(cmd *cobra.Command, factory bbUtil.Factory) (outputSchema.PolicyListOutput, error) {
	var policyListOutput outputSchema.PolicyListOutput

	configClient, err := factory.GetConfigClient(cmd)
	if err != nil {
		return policyListOutput, fmt.Errorf("unable to get config client: %w", err)
	}
	config, configErr := configClient.GetConfig()
	if configErr != nil {
		return policyListOutput, fmt.Errorf("error getting config: %w", configErr)
	}

	if config.PolicyConfiguration.Gatekeeper && !config.PolicyConfiguration.Kyverno {
		return listAllGatekeeperPolicies(cmd, factory)
	}

	if !config.PolicyConfiguration.Gatekeeper && config.PolicyConfiguration.Kyverno {
		return listAllKyvernoPolicies(cmd, factory)
	}

	return policyListOutput, errors.New("either --gatekeeper or --kyverno must be specified")
}

// Internal helper function to query the cluster for Gatekeeper constraint CRDs
func listAllGatekeeperPolicies(cmd *cobra.Command, factory bbUtil.Factory) (outputSchema.PolicyListOutput, error) {
	policyOutput := outputSchema.PolicyListOutput{}

	client, err := factory.GetK8sDynamicClient(cmd)
	if err != nil {
		return policyOutput, err
	}

	gkCrds, err := gatekeeper.FetchGatekeeperCrds(client)
	if err != nil {
		return policyOutput, err
	}

	if len(gkCrds.Items) == 0 {
		policyOutput.Messages = append(policyOutput.Messages, "No Gatekeeper Policies Found")
		return policyOutput, nil
	}

	policyOutput.Messages = append(policyOutput.Messages, "Gatekeeper Policies")

	for _, crd := range gkCrds.Items {
		crdName, _, _ := unstructured.NestedString(crd.Object, "metadata", "name")
		constraints, err := gatekeeper.FetchGatekeeperConstraints(client, crdName)
		if err != nil {
			return policyOutput, err
		}
		crdPolicyOutput := outputSchema.CrdPolicyOutput{
			CrdName: crdName,
		}
		if len(constraints.Items) == 0 {
			crdPolicyOutput.Message = "No constraints found"
		}
		for _, c := range constraints.Items {
			d, err := getGatekeeperPolicyDescriptor(&c)
			if err != nil {
				return policyOutput, err
			}
			policy := outputSchema.PolicyOutput{
				Name:        d.name,
				Namespace:   d.namespace,
				Kind:        d.kind,
				Description: d.desc,
				Action:      d.action,
			}
			crdPolicyOutput.Policies = append(crdPolicyOutput.Policies, policy)
		}
		policyOutput.CrdPolicies = append(policyOutput.CrdPolicies, crdPolicyOutput)
	}

	return policyOutput, nil
}

// Internal helper function to query the cluster for Kyverno policy CRDs
func listAllKyvernoPolicies(cmd *cobra.Command, factory bbUtil.Factory) (outputSchema.PolicyListOutput, error) {
	policyOutput := outputSchema.PolicyListOutput{}

	client, err := factory.GetK8sDynamicClient(cmd)
	if err != nil {
		return policyOutput, err
	}

	kyvernoCrds, err := kyverno.FetchKyvernoCrds(client)
	if err != nil {
		return policyOutput, err
	}

	if len(kyvernoCrds.Items) == 0 {
		policyOutput.Messages = append(policyOutput.Messages, "No Kyverno Policies Found")
		return policyOutput, nil
	}

	policyOutput.Messages = append(policyOutput.Messages, "Kyverno Policies")

	for _, crd := range kyvernoCrds.Items {
		crdName, _, _ := unstructured.NestedString(crd.Object, "metadata", "name")
		policies, err := kyverno.FetchKyvernoPolicies(client, crdName)
		if err != nil {
			loggingClient, loggingErr := factory.GetLoggingClient()
			if loggingErr != nil {
				return policyOutput, fmt.Errorf("error getting logging client: %w for error: %w", loggingErr, err)
			}
			loggingClient.Warn("Error getting kyverno policies: %s", err.Error())
			return policyOutput, err
		}
		crdPolicyOutput := outputSchema.CrdPolicyOutput{
			CrdName: crdName,
		}
		if len(policies.Items) == 0 {
			crdPolicyOutput.Message = "No policies found"
		}
		for _, c := range policies.Items {
			d, err := getKyvernoPolicyDescriptor(&c)
			if err != nil {
				return policyOutput, err
			}
			policy := outputSchema.PolicyOutput{
				Name:        d.name,
				Namespace:   d.namespace,
				Kind:        d.kind,
				Description: d.desc,
				Action:      d.action,
			}
			crdPolicyOutput.Policies = append(crdPolicyOutput.Policies, policy)
		}
		policyOutput.CrdPolicies = append(policyOutput.CrdPolicies, crdPolicyOutput)
	}

	return policyOutput, nil
}

// Internal helper function to query the cluster for Gatekeeper policy descriptors
func getGatekeeperPolicyDescriptor(resource *unstructured.Unstructured) (*policyDescriptor, error) {
	kind, _, err := unstructured.NestedString(resource.Object, "kind")
	if err != nil {
		return nil, err
	}

	name, _, err := unstructured.NestedString(resource.Object, "metadata", "name")
	if err != nil {
		return nil, err
	}

	desc, _, err := unstructured.NestedString(resource.Object, "metadata", "annotations", "constraints.gatekeeper/description")
	if err != nil {
		return nil, err
	}

	action, _, err := unstructured.NestedString(resource.Object, "spec", "enforcementAction")
	if err != nil {
		return nil, err
	}

	descriptor := &policyDescriptor{
		kind:   kind,
		name:   name,
		desc:   desc,
		action: action,
	}

	return descriptor, nil
}

// Internal helper function to query the cluster for Kyverno policy descriptors
func getKyvernoPolicyDescriptor(resource *unstructured.Unstructured) (*policyDescriptor, error) {
	kind, _, err := unstructured.NestedString(resource.Object, "kind")
	if err != nil {
		return nil, err
	}

	name, _, err := unstructured.NestedString(resource.Object, "metadata", "name")
	if err != nil {
		return nil, err
	}

	namespace, _, err := unstructured.NestedString(resource.Object, "metadata", "namespace")
	if err != nil {
		return nil, err
	}

	desc, _, err := unstructured.NestedString(resource.Object, "metadata", "annotations", "policies.kyverno.io/description")
	if err != nil {
		return nil, err
	}

	action, _, err := unstructured.NestedString(resource.Object, "spec", "validationFailureAction")
	if err != nil {
		return nil, err
	}

	descriptor := &policyDescriptor{
		kind:      kind,
		name:      name,
		namespace: namespace,
		desc:      desc,
		action:    action,
	}

	return descriptor, nil
}
