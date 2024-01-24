package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"

	bbutil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
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
func NewPoliciesCmd(factory bbutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {

	cmd := &cobra.Command{
		Use:     policyUse,
		Short:   policyShort,
		Long:    policyLong,
		Example: policyExample,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, hint string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return matchingPolicyNames(factory, hint, cmd.Flags())
		},
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 1 {
				cmdutil.CheckErr(listPoliciesByName(factory, streams, args[0], cmd.Flags()))
			} else {
				cmdutil.CheckErr(listAllPolicies(factory, streams, cmd.Flags()))
			}
		},
	}

	cmd.Flags().IntP("gatekeeper", "g", 0, "Print gatekeeper policy")
	cmd.Flags().Lookup("gatekeeper").NoOptDefVal = "1"

	cmd.Flags().IntP("kyverno", "k", 0, "Print kyverno policy")
	cmd.Flags().Lookup("kyverno").NoOptDefVal = "1"

	return cmd
}

// find policies with given prefix for command completion
func matchingPolicyNames(factory bbutil.Factory, hint string, flags *pflag.FlagSet) ([]string, cobra.ShellCompDirective) {

	// either --kyverno or --gatekeeper must be specified.
	// No option default value is 1 for either of these flags.

	kyverno, _ := flags.GetInt("kyverno")
	gatekeeper, _ := flags.GetInt("gatekeeper")

	if gatekeeper == 1 && kyverno == 0 {
		return matchingGatekeeperPolicyNames(factory, hint)
	}

	if kyverno == 1 && gatekeeper == 0 {
		return matchingKyvernoPolicyNames(factory, hint)
	}

	return nil, cobra.ShellCompDirectiveDefault
}

// find policies with given prefix for command completion
func matchingGatekeeperPolicyNames(factory bbutil.Factory, hint string) ([]string, cobra.ShellCompDirective) {

	client, err := factory.GetK8sDynamicClient()
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

func matchingKyvernoPolicyNames(factory bbutil.Factory, hint string) ([]string, cobra.ShellCompDirective) {
	client, err := factory.GetK8sDynamicClient()
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
// kyverno clusterpolicy crds
func listPoliciesByName(factory bbutil.Factory, streams genericclioptions.IOStreams, name string, flags *pflag.FlagSet) error {

	// either --kyverno or --gatekeeper must be specified.
	// No option default value is 1 for either of these flags.

	kyverno, _ := flags.GetInt("kyverno")
	gatekeeper, _ := flags.GetInt("gatekeeper")

	if gatekeeper == 1 && kyverno == 0 {
		return listGatekeeperPoliciesByName(factory, streams, name)
	}

	if kyverno == 1 && gatekeeper == 0 {
		return listKyvernoPoliciesByName(factory, streams, name)
	}

	return fmt.Errorf("either --gatekeeper or --kyverno must be specified")
}

// query the cluster using dynamic client to get information on gatekeeper constraint crds
func listGatekeeperPoliciesByName(factory bbutil.Factory, streams genericclioptions.IOStreams, name string) error {

	client, err := factory.GetK8sDynamicClient()
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
func listKyvernoPoliciesByName(factory bbutil.Factory, streams genericclioptions.IOStreams, name string) error {

	client, err := factory.GetK8sDynamicClient()
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
func listAllPolicies(factory bbutil.Factory, streams genericclioptions.IOStreams, flags *pflag.FlagSet) error {

	kyverno, _ := flags.GetInt("kyverno")
	gatekeeper, _ := flags.GetInt("gatekeeper")

	if gatekeeper == 1 && kyverno == 0 {
		return listAllGatekeeperPolicies(factory, streams)
	}

	if kyverno == 1 && gatekeeper == 0 {
		return listAllKyvernoPolicies(factory, streams)
	}

	return fmt.Errorf("either --gatekeeper or --kyverno must be specified")
}

// query the cluster using dynamic client to get information on gatekeeper constraint crds
func listAllGatekeeperPolicies(factory bbutil.Factory, streams genericclioptions.IOStreams) error {

	client, err := factory.GetK8sDynamicClient()
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
			fmt.Fprint(streams.Out, "\nNo constraints found\n\n")
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

func listAllKyvernoPolicies(factory bbutil.Factory, streams genericclioptions.IOStreams) error {

	client, err := factory.GetK8sDynamicClient()
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
			fmt.Fprint(streams.Out, "No policies found\n\n")
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
