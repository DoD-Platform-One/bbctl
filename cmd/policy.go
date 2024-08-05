package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/gatekeeper"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/kyverno"

	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	policyUse = `policy --PROVIDER CONSTRAINT_NAME`

	policyShort = i18n.T(`Describe the configured policies implemented by Gatekeeper or Kyverno.`)

	policyLong = templates.LongDesc(i18n.T(`
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

type policyDescriptor struct {
	name      string // policy name
	namespace string // policy namespace (kyverno policy)
	kind      string // policy kind
	desc      string // policy description
	action    string // enforcement action
}

// NewPoliciesCmd - Creates a new Cobra command which implements the `bbctl policy` functionality
func NewPoliciesCmd(factory bbUtil.Factory) (*cobra.Command, error) {
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
			if len(args) == 1 {
				return listPoliciesByName(cmd, factory, args[0])
			} else {
				return listAllPolicies(cmd, factory)
			}
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	configClient, clientError := factory.GetConfigClient(cmd)
	if clientError != nil {
		return nil, fmt.Errorf("Unable to get config client: %v", clientError)
	}

	gatekeeperError := configClient.SetAndBindFlag("gatekeeper", false, "Print gatekeeper policy")
	if gatekeeperError != nil {
		return nil, fmt.Errorf("Unable to add flags to command: %v", gatekeeperError)
	}

	kyvernoError := configClient.SetAndBindFlag("kyverno", false, "Print kyverno policy")
	if kyvernoError != nil {
		return nil, fmt.Errorf("Unable to add flags to command: %v", kyvernoError)
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

	var matches []string = make([]string, 0)

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

	var matches []string = make([]string, 0)

	for _, crd := range kyvernoCrds.Items {
		crdName, _, _ := unstructured.NestedString(crd.Object, "metadata", "name")
		policies, err := kyverno.FetchKyvernoPolicies(client, crdName)
		if err != nil {
			factory.GetLoggingClient().Warn("Error getting kyverno policies: %s", err.Error())
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
func listPoliciesByName(cmd *cobra.Command, factory bbUtil.Factory, name string) error {
	configClient, err := factory.GetConfigClient(cmd)
	if err != nil {
		return fmt.Errorf("Unable to get config client: %v", err)
	}
	config, configErr := configClient.GetConfig()
	if configErr != nil {
		return fmt.Errorf("error getting config: %w", configErr)
	}

	if config.PolicyConfiguration.Gatekeeper && !config.PolicyConfiguration.Kyverno {
		return listGatekeeperPoliciesByName(cmd, factory, name)
	}

	if !config.PolicyConfiguration.Gatekeeper && config.PolicyConfiguration.Kyverno {
		return listKyvernoPoliciesByName(cmd, factory, name)
	}

	return fmt.Errorf("either --gatekeeper or --kyverno must be specified, but not both")
}

// Internal helper function to query the cluster for Gatekeeper constraint CRDs matching the given prefix
func listGatekeeperPoliciesByName(cmd *cobra.Command, factory bbUtil.Factory, name string) error {
	streams := factory.GetIOStream()
	client, err := factory.GetK8sDynamicClient(cmd)
	if err != nil {
		return err
	}

	crdName := fmt.Sprintf("%s.constraints.gatekeeper.sh", name)

	constraints, err := gatekeeper.FetchGatekeeperConstraints(client, crdName)
	if err != nil {
		return err
	}

	fmt.Fprintf(streams.Out, "%s\n", crdName)

	if len(constraints.Items) == 0 {
		fmt.Fprint(streams.Out, "\nNo constraints found\n")
	}

	for _, c := range constraints.Items {
		d, err := getGatekeeperPolicyDescriptor(&c)
		if err != nil {
			return err
		}
		printPolicyDescriptor(d, streams.Out)
	}

	return nil
}

// Internal helper function to query the cluster for Kyverno cluster policy CRDs matching the given name
func listKyvernoPoliciesByName(cmd *cobra.Command, factory bbUtil.Factory, name string) error {
	streams := factory.GetIOStream()
	client, err := factory.GetK8sDynamicClient(cmd)
	if err != nil {
		return err
	}

	kyvernoCrds, err := kyverno.FetchKyvernoCrds(client)
	if err != nil {
		return err
	}

	for _, crd := range kyvernoCrds.Items {
		crdName, _, _ := unstructured.NestedString(crd.Object, "metadata", "name")
		policies, err := kyverno.FetchKyvernoPolicies(client, crdName)
		if err != nil {
			factory.GetLoggingClient().Warn("Error getting kyverno policies: %s", err.Error())
			return err
		}
		for _, c := range policies.Items {
			policyName, _, _ := unstructured.NestedString(c.Object, "metadata", "name")
			if policyName == name {
				d, err := getKyvernoPolicyDescriptor(&c)
				if err != nil {
					return err
				}
				printPolicyDescriptor(d, streams.Out)
				return nil
			}
		}
	}

	fmt.Fprint(streams.Out, "No Matching Policy Found\n")

	return nil
}

// Internal helper function to query the cluster using the dynamic client to get information on the following:
//   - All Gatekeeper constraint CRDs
//   - All Kyverno policy CRDs
func listAllPolicies(cmd *cobra.Command, factory bbUtil.Factory) error {
	configClient, err := factory.GetConfigClient(cmd)
	if err != nil {
		return fmt.Errorf("Unable to get config client: %v", err)
	}
	config, configErr := configClient.GetConfig()
	if configErr != nil {
		return fmt.Errorf("error getting config: %w", configErr)
	}

	if config.PolicyConfiguration.Gatekeeper && !config.PolicyConfiguration.Kyverno {
		return listAllGatekeeperPolicies(cmd, factory)
	}

	if !config.PolicyConfiguration.Gatekeeper && config.PolicyConfiguration.Kyverno {
		return listAllKyvernoPolicies(cmd, factory)
	}

	return fmt.Errorf("either --gatekeeper or --kyverno must be specified")
}

// Internal helper function to query the cluster for Gatekeeper constraint CRDs
func listAllGatekeeperPolicies(cmd *cobra.Command, factory bbUtil.Factory) error {
	streams := factory.GetIOStream()
	client, err := factory.GetK8sDynamicClient(cmd)
	if err != nil {
		return err
	}

	gkCrds, err := gatekeeper.FetchGatekeeperCrds(client)
	if err != nil {
		return err
	}

	if len(gkCrds.Items) != 0 {
		fmt.Fprint(streams.Out, "\nGatekeeper Policies:\n\n")
	} else {
		fmt.Fprint(streams.Out, "No Gatekeeper Policies Found\n")
	}

	for _, crd := range gkCrds.Items {
		crdName, _, _ := unstructured.NestedString(crd.Object, "metadata", "name")
		constraints, err := gatekeeper.FetchGatekeeperConstraints(client, crdName)
		if err != nil {
			return err
		}
		fmt.Fprintf(streams.Out, "%s\n", crdName)
		if len(constraints.Items) == 0 {
			fmt.Fprint(streams.Out, "\nNo constraints found\n\n\n")
		}
		for _, c := range constraints.Items {
			d, err := getGatekeeperPolicyDescriptor(&c)
			if err != nil {
				return err
			}
			printPolicyDescriptor(d, streams.Out)
		}
	}

	return nil
}

// Internal helper function to query the cluster for Kyverno policy CRDs
func listAllKyvernoPolicies(cmd *cobra.Command, factory bbUtil.Factory) error {
	streams := factory.GetIOStream()
	client, err := factory.GetK8sDynamicClient(cmd)
	if err != nil {
		return err
	}

	kyvernoCrds, err := kyverno.FetchKyvernoCrds(client)
	if err != nil {
		return err
	}

	if len(kyvernoCrds.Items) != 0 {
		fmt.Fprint(streams.Out, "\nKyverno Policies\n")
	} else {
		fmt.Fprint(streams.Out, "No Kyverno Policies Found\n")
	}

	for _, crd := range kyvernoCrds.Items {
		crdName, _, _ := unstructured.NestedString(crd.Object, "metadata", "name")
		policies, err := kyverno.FetchKyvernoPolicies(client, crdName)
		if err != nil {
			factory.GetLoggingClient().Warn("Error getting kyverno policies: %s", err.Error())
			return err
		}
		fmt.Fprintf(streams.Out, "\n%s\n\n", crdName)
		if len(policies.Items) == 0 {
			fmt.Fprint(streams.Out, "No policies found\n\n\n")
		}
		for _, c := range policies.Items {
			d, err := getKyvernoPolicyDescriptor(&c)
			if err != nil {
				return err
			}
			printPolicyDescriptor(d, streams.Out)
		}
	}

	return nil
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

// Internal helper function to print policy information to the console
func printPolicyDescriptor(d *policyDescriptor, w io.Writer) {
	if d.namespace != "" {
		fmt.Fprintf(w, "\nKind: %s, Name: %s, Namespace: %s, EnforcementAction: %s\n", d.kind, d.name, d.namespace, d.action)
	} else {
		fmt.Fprintf(w, "\nKind: %s, Name: %s, EnforcementAction: %s\n", d.kind, d.name, d.action)
	}
	fmt.Fprintf(w, "\n%s\n\n\n", d.desc)
}
