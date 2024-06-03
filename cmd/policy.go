package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"
	cmdUtil "k8s.io/kubectl/pkg/cmd/util"

	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/gatekeeper"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/kyverno"

	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	policyUse = `policy CONSTRAINT_NAME`

	policyShort = i18n.T(`Describe policy implemented by Gatekeeper or Kyverno.`)

	policyLong = templates.LongDesc(i18n.T(`
		Describe policy implemented by Gatekeeper or Kyverno.
		Use either --gatekeeper or --kyverno flag to select the policy provider.
	`))

	policyExample = templates.Examples(i18n.T(`
		# Describe gatekeeper policy
		bbctl policy --gatekeeper CONSTRAINT_NAME
	
	    # Get a list of active gatekeeper policies
		bbctl policy --gatekeeper
		
		# Describe kyverno policy
		bbctl policy --kyverno CONSTRAINT_NAME
	
	    # Get a list of active kyverno policies
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

// NewPoliciesCmd - new policies command
func NewPoliciesCmd(factory bbUtil.Factory, streams genericIOOptions.IOStreams) *cobra.Command {
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
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 1 {
				cmdUtil.CheckErr(listPoliciesByName(cmd, factory, streams, args[0]))
			} else {
				cmdUtil.CheckErr(listAllPolicies(cmd, factory, streams))
			}
		},
	}

	loggingClient := factory.GetLoggingClient()
	configClient, err := factory.GetConfigClient(cmd)
	loggingClient.HandleError("Unable to get config client: %v", err)

	loggingClient.HandleError(
		"Unable to add flags to command: %v",
		configClient.SetAndBindFlag(
			"gatekeeper",
			false,
			"Print gatekeeper policy",
		),
	)
	loggingClient.HandleError(
		"Unable to add flags to command: %v",
		configClient.SetAndBindFlag(
			"kyverno",
			false,
			"Print kyverno policy",
		),
	)

	return cmd
}

// find policies with given prefix for command completion
func matchingPolicyNames(cmd *cobra.Command, factory bbUtil.Factory, hint string) ([]string, cobra.ShellCompDirective) {
	loggingClient := factory.GetLoggingClient()
	configClient, err := factory.GetConfigClient(cmd)
	loggingClient.HandleError("Unable to get config client: %v", err)
	config := configClient.GetConfig()

	if config.PolicyConfiguration.Gatekeeper && !config.PolicyConfiguration.Kyverno {
		return matchingGatekeeperPolicyNames(cmd, factory, hint)
	}

	if !config.PolicyConfiguration.Gatekeeper && config.PolicyConfiguration.Kyverno {
		return matchingKyvernoPolicyNames(cmd, factory, hint)
	}

	return nil, cobra.ShellCompDirectiveDefault
}

// find policies with given prefix for command completion
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

// query the cluster using dynamic client to get information on the following:
// gatekeeper constraint crds
// kyverno cluster policy crds
func listPoliciesByName(cmd *cobra.Command, factory bbUtil.Factory, streams genericIOOptions.IOStreams, name string) error {
	loggingClient := factory.GetLoggingClient()
	configClient, err := factory.GetConfigClient(cmd)
	loggingClient.HandleError("Unable to get config client: %v", err)
	config := configClient.GetConfig()

	if config.PolicyConfiguration.Gatekeeper && !config.PolicyConfiguration.Kyverno {
		return listGatekeeperPoliciesByName(cmd, factory, streams, name)
	}

	if !config.PolicyConfiguration.Gatekeeper && config.PolicyConfiguration.Kyverno {
		return listKyvernoPoliciesByName(cmd, factory, streams, name)
	}

	return fmt.Errorf("either --gatekeeper or --kyverno must be specified, but not both")
}

// query the cluster using dynamic client to get information on gatekeeper constraint crds
func listGatekeeperPoliciesByName(cmd *cobra.Command, factory bbUtil.Factory, streams genericIOOptions.IOStreams, name string) error {
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

// query the cluster using dynamic client to get information on kyverno policies
func listKyvernoPoliciesByName(cmd *cobra.Command, factory bbUtil.Factory, streams genericIOOptions.IOStreams, name string) error {
	client, err := factory.GetK8sDynamicClient(cmd)
	if err != nil {
		return nil
	}

	kyvernoCrds, err := kyverno.FetchKyvernoCrds(client)
	if err != nil {
		return nil
	}

	for _, crd := range kyvernoCrds.Items {
		crdName, _, _ := unstructured.NestedString(crd.Object, "metadata", "name")
		policies, err := kyverno.FetchKyvernoPolicies(client, crdName)
		if err != nil {
			return err
		}
		for _, c := range policies.Items {
			policyName, _, _ := unstructured.NestedString(c.Object, "metadata", "name")
			if strings.Compare(policyName, name) == 0 {
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

// query the cluster using dynamic client to get information on gatekeeper constraint crds
// and kyverno cluster policy crds
func listAllPolicies(cmd *cobra.Command, factory bbUtil.Factory, streams genericIOOptions.IOStreams) error {
	loggingClient := factory.GetLoggingClient()
	configClient, err := factory.GetConfigClient(cmd)
	loggingClient.HandleError("Unable to get config client: %v", err)
	config := configClient.GetConfig()

	if config.PolicyConfiguration.Gatekeeper && !config.PolicyConfiguration.Kyverno {
		return listAllGatekeeperPolicies(cmd, factory, streams)
	}

	if !config.PolicyConfiguration.Gatekeeper && config.PolicyConfiguration.Kyverno {
		return listAllKyvernoPolicies(cmd, factory, streams)
	}

	return fmt.Errorf("either --gatekeeper or --kyverno must be specified")
}

// query the cluster using dynamic client to get information on gatekeeper constraint crds
func listAllGatekeeperPolicies(cmd *cobra.Command, factory bbUtil.Factory, streams genericIOOptions.IOStreams) error {
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

func listAllKyvernoPolicies(cmd *cobra.Command, factory bbUtil.Factory, streams genericIOOptions.IOStreams) error {
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

func printPolicyDescriptor(d *policyDescriptor, w io.Writer) {
	if d.namespace != "" {
		fmt.Fprintf(w, "\nKind: %s, Name: %s, Namespace: %s, EnforcementAction: %s\n", d.kind, d.name, d.namespace, d.action)
	} else {
		fmt.Fprintf(w, "\nKind: %s, Name: %s, EnforcementAction: %s\n", d.kind, d.name, d.action)
	}
	fmt.Fprintf(w, "\n%s\n\n\n", d.desc)
}
