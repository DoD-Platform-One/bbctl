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

	policyShort = i18n.T(`Describe gatekeeper policy.`)

	policyLong = templates.LongDesc(i18n.T(`Describe gatekeeper policy.`))

	policyExample = templates.Examples(i18n.T(`
		# Describe policy
		bbctl policy CONSTRAINT_NAME
	
	    # Get a list of active policies
		bbctl policy
	`))
)

type policyDescriptor struct {
	name   string // constraint name
	kind   string // constraint kind
	desc   string // constraint description
	action string // enforcement action
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
			return matchingPolicyNames(factory, hint)
		},
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 1 {
				cmdutil.CheckErr(listPoliciesByName(factory, streams, args[0]))
			} else {
				cmdutil.CheckErr(listAllPolicies(factory, streams))
			}
		},
	}

	return cmd
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

// query the cluster using dynamic client to get information on gatekeeper constraint crds
func listPoliciesByName(factory bbutil.Factory, streams genericclioptions.IOStreams, name string) error {

	client, err := factory.GetK8sDynamicClient()
	if err != nil {
		return err
	}

	crdName := fmt.Sprintf("%s.constraints.gatekeeper.sh", name)

	constraints, err := fetchConstraints(client, crdName)
	if err != nil {
		return err
	}

	fmt.Fprintf(streams.Out, "%s\n", crdName)

	if len(constraints.Items) == 0 {
		fmt.Fprint(streams.Out, "\nNo constraints found\n")
	}

	for _, c := range constraints.Items {
		d, err := getPolicyDescriptor(&c)
		if err != nil {
			return err
		}
		printDescriptor(d, streams.Out)
	}

	return nil
}

// query the cluster using dynamic client to get information on gatekeeper constraint crds
func listAllPolicies(factory bbutil.Factory, streams genericclioptions.IOStreams) error {

	client, err := factory.GetK8sDynamicClient()
	if err != nil {
		return err
	}

	gkCrds, err := fetchGatekeeperCrds(client)
	if err != nil {
		return err
	}

	for _, crd := range gkCrds.Items {
		crdName, _, _ := unstructured.NestedString(crd.Object, "metadata", "name")
		constraints, err := fetchConstraints(client, crdName)
		if err != nil {
			return err
		}
		fmt.Fprintf(streams.Out, "%s\n", crdName)
		if len(constraints.Items) == 0 {
			fmt.Fprint(streams.Out, "\nNo constraints found\n\n")
		}
		for _, c := range constraints.Items {
			d, err := getPolicyDescriptor(&c)
			if err != nil {
				return err
			}
			printDescriptor(d, streams.Out)
		}
	}

	return nil
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
		constraints, err := fetchConstraints(client, crdName)
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

func fetchGatekeeperCrds(client dynamic.Interface) (*unstructured.UnstructuredList, error) {

	var customResource = schema.GroupVersionResource{Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"}

	opts := metav1.ListOptions{LabelSelector: "app.kubernetes.io/name=gatekeeper"}

	gkResources, err := client.Resource(customResource).List(context.TODO(), opts)
	if err != nil {
		return nil, fmt.Errorf("error getting gatekeeper crds: %s", err.Error())
	}

	return gkResources, nil
}

func fetchConstraints(client dynamic.Interface, name string) (*unstructured.UnstructuredList, error) {

	resourceName := strings.Split(name, ".")[0]

	var constraintResource = schema.GroupVersionResource{Group: "constraints.gatekeeper.sh", Version: "v1beta1", Resource: resourceName}

	resources, err := client.Resource(constraintResource).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting gatekeeper constraint: %s", err.Error())
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

func getPolicyDescriptor(resource *unstructured.Unstructured) (*policyDescriptor, error) {

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

// find policies with given prefix for command completion
func matchingPolicyNames(factory bbutil.Factory, hint string) ([]string, cobra.ShellCompDirective) {

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

func printViolation(v *constraintViolation, w io.Writer) {
	fmt.Fprintf(w, "Time: %s, Resource: %s, Kind: %s, Namespace: %s\n", v.timestamp, v.name, v.kind, v.namespace)
	if v.constraint != "" {
		fmt.Fprintf(w, "\nConstraint: %s\n", v.constraint)
	}
	fmt.Fprintf(w, "\n%s\n\n\n", v.message)
}

func printDescriptor(d *policyDescriptor, w io.Writer) {
	fmt.Fprintf(w, "\nKind: %s, Name: %s, EnforcementAction: %s\n", d.kind, d.name, d.action)
	fmt.Fprintf(w, "\n%s\n\n\n", d.desc)
}
