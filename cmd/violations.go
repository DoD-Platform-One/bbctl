package cmd

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	bbutil "repo1.dso.mil/platform-one/big-bang/apps/product-tools/bbctl/util"
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

	gkCrds, err := fetchGatekeeperCrds(client)
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

	kyvernoCrds, err := fetchKyvernoCrds(client)
	if err != nil {
		return nil, cobra.ShellCompDirectiveDefault
	}

	var matches []string = make([]string, 0)

	for _, crd := range kyvernoCrds.Items {
		crdName, _, _ := unstructured.NestedString(crd.Object, "metadata", "name")
		policies, err := fetchKyvernoPolicies(client, crdName)
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

	constraints, err := fetchGatekeeperConstraints(client, crdName)
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

	kyvernoCrds, err := fetchKyvernoCrds(client)
	if err != nil {
		return nil
	}

	for _, crd := range kyvernoCrds.Items {
		crdName, _, _ := unstructured.NestedString(crd.Object, "metadata", "name")
		policies, err := fetchKyvernoPolicies(client, crdName)
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

	gkCrds, err := fetchGatekeeperCrds(client)
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
		constraints, err := fetchGatekeeperConstraints(client, crdName)
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

	kyvernoCrds, err := fetchKyvernoCrds(client)
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
		policies, err := fetchKyvernoPolicies(client, crdName)
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

func fetchGatekeeperCrds(client dynamic.Interface) (*unstructured.UnstructuredList, error) {

	var customResource = schema.GroupVersionResource{Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"}

	opts := metav1.ListOptions{LabelSelector: "app.kubernetes.io/name=gatekeeper"}

	gkResources, err := client.Resource(customResource).List(context.TODO(), opts)
	if err != nil {
		return nil, fmt.Errorf("error getting gatekeeper crds: %s", err.Error())
	}

	return gkResources, nil
}

func fetchGatekeeperConstraints(client dynamic.Interface, name string) (*unstructured.UnstructuredList, error) {

	resourceName := strings.Split(name, ".")[0]

	var constraintResource = schema.GroupVersionResource{Group: "constraints.gatekeeper.sh", Version: "v1beta1", Resource: resourceName}

	resources, err := client.Resource(constraintResource).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting gatekeeper constraint: %s", err.Error())
	}

	return resources, nil
}

func fetchKyvernoCrds(client dynamic.Interface) (*unstructured.UnstructuredList, error) {

	var customResource = schema.GroupVersionResource{Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"}

	opts := metav1.ListOptions{LabelSelector: "app.kubernetes.io/name=kyverno"}

	kyvernoResources, err := client.Resource(customResource).List(context.TODO(), opts)
	if err != nil {
		return nil, fmt.Errorf("error getting kyverno crds: %s", err.Error())
	}

	items := make([]unstructured.Unstructured, 0)
	for _, crd := range kyvernoResources.Items {
		crdName, _, _ := unstructured.NestedString(crd.Object, "metadata", "name")
		if strings.HasSuffix(crdName, "policies.kyverno.io") {
			items = append(items, crd)
		}
	}

	kyvernoResources.Items = items

	return kyvernoResources, nil
}

func fetchKyvernoPolicies(client dynamic.Interface, name string) (*unstructured.UnstructuredList, error) {

	resourceName := strings.Split(name, ".")[0]

	var policyResource = schema.GroupVersionResource{Group: "kyverno.io", Version: "v1", Resource: resourceName}

	resources, err := client.Resource(policyResource).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting kyverno policies: %s", err.Error())
	}

	return resources, nil
}

func getConstraintViolations(resource *unstructured.Unstructured) (*[]constraintViolation, error) {

	var violationTimestamp string = ""
	ts, ok, err := unstructured.NestedFieldNoCopy(resource.Object, "status", "auditTimestamp")
	if err != nil {
		return nil, err
	}
	if ok {
		timestamp, _ := ts.(string)
		violationTimestamp = timestamp
	}

	statusViolations, _, err := unstructured.NestedSlice(resource.Object, "status", "violations")
	if err != nil {
		return nil, err
	}

	violations := make([]constraintViolation, len(statusViolations))

	for i, v := range statusViolations {
		details, _ := v.(map[string]interface{})
		violations[i] = constraintViolation{
			name:      fmt.Sprintf("%s", details["name"]),
			namespace: fmt.Sprintf("%s", details["namespace"]),
			kind:      fmt.Sprintf("%s", details["kind"]),
			action:    fmt.Sprintf("%s", details["enforcementAction"]),
			message:   fmt.Sprintf("%s", details["message"]),
			timestamp: violationTimestamp,
		}
	}

	return &violations, nil
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

var (
	violationsUse = `violations`

	violationsShort = i18n.T(`List policy violations resulting in request denial.`)

	violationsLong = templates.LongDesc(i18n.T(`
		List policy violations reported by admission webhook that result in request denial.
	`))

	violationsExample = templates.Examples(i18n.T(`
		# Get a list of policy violations resulting in request denial across all namespaces
		bbctl violations 
		
		# Get a list of policy violations resulting in request denial in the given namespace
		bbctl violations -n NAMESPACE
		
		# Get a list of policy violations reported by audit process (dryrun mode) across all namespaces
		bbctl violations --audit	
		
		# Get a list of policy violations reported by audit process (dryrun mode) in the given namespace
		bbctl violations --audit --namespace NAMESPACE	
	`))
)

type constraintViolation struct {
	name       string // resource name
	kind       string // resource kind
	namespace  string // resource namespace
	constraint string // constraint name
	message    string // policy violation message
	action     string // enforcement action
	timestamp  string // utc time
}

// NewViolationsCmd - new violations commmand
func NewViolationsCmd(factory bbutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {

	cmd := &cobra.Command{
		Use:     violationsUse,
		Short:   violationsShort,
		Long:    violationsLong,
		Example: violationsExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(getViolations(factory, streams, cmd.Flags()))
		},
	}

	cmd.Flags().BoolP("audit", "d", false, "list violations in dry-run mode")

	return cmd
}

func getViolations(factory bbutil.Factory, streams genericclioptions.IOStreams, flags *pflag.FlagSet) error {

	namespace, _ := flags.GetString("namespace")

	audit, _ := flags.GetBool("audit")

	if audit {
		return listAuditViolations(factory, streams, namespace)
	}

	return listDenyViolations(factory, streams, namespace)
}

func listDenyViolations(factory bbutil.Factory, streams genericclioptions.IOStreams, namespace string) error {

	client, err := factory.GetK8sClientset()
	if err != nil {
		return err
	}

	events, err := client.CoreV1().Events("").List(context.TODO(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("%s=%s", "reason", "FailedAdmission"),
	})
	if err != nil {
		return err
	}

	violationsFound := false
	for _, event := range events.Items {

		if namespace != "" && event.Annotations["resource_namespace"] != namespace {
			continue
		}

		violationsFound = true

		violation := &constraintViolation{
			name:       event.Annotations["resource_name"],
			kind:       event.Annotations["resource_kind"],
			namespace:  event.Annotations["resource_namespace"],
			constraint: fmt.Sprintf("%s:%s", strings.ToLower(event.Annotations["constraint_kind"]), event.Annotations["constraint_name"]),
			message:    event.Message,
			timestamp:  event.CreationTimestamp.UTC().Format(time.RFC3339),
		}

		printViolation(violation, streams.Out)
	}

	if !violationsFound {
		fmt.Fprintf(streams.Out, "No events found for deny violations\n\n")
		fmt.Fprintf(streams.Out, "Do you have the following values defined for the gatekeeper chart?\n\n")
		fmt.Fprintf(streams.Out, "gatekeeper:\n")
		fmt.Fprintf(streams.Out, "  emitAdmissionEvents: true\n")
		fmt.Fprintf(streams.Out, "  logDenies: true\n\n")
		fmt.Fprintf(streams.Out, "Note that violations in dryrun and warn mode are not effected by these settings.\n")
		fmt.Fprintf(streams.Out, "To list dryrun violations, use --audit flag.\n")
	}

	return nil
}

// query the cluster using dynamic client to get audit violation information from gatekeeper constraint crds
func listAuditViolations(factory bbutil.Factory, streams genericclioptions.IOStreams, namespace string) error {

	client, err := factory.GetK8sDynamicClient()
	if err != nil {
		return err
	}

	gkCrds, err := fetchGatekeeperCrds(client)
	if err != nil {
		return err
	}

	if len(gkCrds.Items) == 0 {
		fmt.Fprintf(streams.Out, "No violations found in audit\n\n\n")
		return nil
	}

	for _, crd := range gkCrds.Items {
		crdName, _, _ := unstructured.NestedString(crd.Object, "metadata", "name")
		constraints, err := fetchGatekeeperConstraints(client, crdName)
		if err != nil {
			return err
		}
		for _, c := range constraints.Items {
			// get violations
			violations, _ := getConstraintViolations(&c)
			// process violations
			processViolations(namespace, violations, crdName, streams)
		}
	}

	return nil
}

func processViolations(namespace string, violations *[]constraintViolation, crdName string, streams genericclioptions.IOStreams) {
	if len(*violations) != 0 {
		fmt.Fprintf(streams.Out, "%s\n\n", crdName)
		violationsFound := false
		for _, v := range *violations {
			if namespace != "" && v.namespace != namespace {
				continue
			}
			violationsFound = true
			printViolation(&v, streams.Out)
		}
		if !violationsFound {
			fmt.Fprintf(streams.Out, "No violations found in audit\n\n\n")
		}
	}
}

func printViolation(v *constraintViolation, w io.Writer) {
	fmt.Fprintf(w, "Time: %s, Resource: %s, Kind: %s, Namespace: %s\n", v.timestamp, v.name, v.kind, v.namespace)
	if v.constraint != "" {
		fmt.Fprintf(w, "\nConstraint: %s\n", v.constraint)
	}
	fmt.Fprintf(w, "\n%s\n\n\n", v.message)
}
