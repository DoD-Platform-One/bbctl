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
	cmdutil "k8s.io/kubectl/pkg/cmd/util"

	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	bbutil "repo1.dso.mil/platform-one/big-bang/apps/product-tools/bbctl/util"
	"repo1.dso.mil/platform-one/big-bang/apps/product-tools/bbctl/util/gatekeeper"
	"repo1.dso.mil/platform-one/big-bang/apps/product-tools/bbctl/util/kyverno"
)

var (
	violationsUse = `violations`

	violationsShort = i18n.T(`List policy violations.`)

	violationsLong = templates.LongDesc(i18n.T(`
		List policy violations reported by Gatekeeper or Kyverno Policy Engine.

		Note: In case of kyverno, violations are reported using the default namespace for kyverno policy resource
		of kind ClusterPolicy irrespective of the namespace of the resource that failed the policy. Any violations
		that occur because of namespace specific policy i.e. kind Policy is reported using the namespace the resource
		is associated with. If it is desired to see the violations because of ClusterPolicy objects, use the command
		as follows:

		bbctl violations -n default
	`))

	violationsExample = templates.Examples(i18n.T(`
		# Get a list of policy violations resulting in request denial across all namespaces
		bbctl violations 
		
		# Get a list of policy violations resulting in request denial in the given namespace.
		bbctl violations -n NAMESPACE		
		
		# Get a list of policy violations reported by audit process across all namespaces
		bbctl violations --audit	
		
		# Get a list of policy violations reported by audit process in the given namespace
		bbctl violations --audit --namespace NAMESPACE	
	`))

	admissionControllerKyvernoEventSource = "admission-controller"

	policyControllerKyvernoEventSource = "policy-controller"
)

type policyViolation struct {
	name       string // resource name
	kind       string // resource kind
	namespace  string // resource namespace
	policy     string // kyverno policy name
	constraint string // gatekeeper constraint name
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

	cmd.Flags().BoolP("audit", "d", false, "list violations in audit mode")

	return cmd
}

func getViolations(factory bbutil.Factory, streams genericclioptions.IOStreams, flags *pflag.FlagSet) error {

	namespace, _ := flags.GetString("namespace")

	audit, _ := flags.GetBool("audit")

	gkFound, err := gatekeeperExists(factory, streams)
	if err != nil {
		return err
	}

	if gkFound {
		err = listGkViolations(factory, streams, namespace, audit)
		if err != nil {
			return err
		}
	}

	kyvernoFound, err := kyvernoExists(factory, streams)
	if err != nil {
		return err
	}

	if kyvernoFound {
		return listKyvernoViolations(factory, streams, namespace, audit)
	}

	return nil
}

func kyvernoExists(factory bbutil.Factory, streams genericclioptions.IOStreams) (bool, error) {

	client, err := factory.GetK8sDynamicClient()
	if err != nil {
		return false, err
	}

	kyvernoCrds, err := kyverno.FetchKyvernoCrds(client)
	if err != nil {
		return false, err
	}

	return len(kyvernoCrds.Items) != 0, nil

}

func listKyvernoViolations(factory bbutil.Factory, streams genericclioptions.IOStreams, namespace string, listAuditViolations bool) error {

	client, err := factory.GetK8sClientset()
	if err != nil {
		return err
	}

	events, err := client.CoreV1().Events("").List(context.TODO(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("%s=%s", "reason", "PolicyViolation"),
	})
	if err != nil {
		return err
	}

	violationsFound := false
	for _, event := range events.Items {

		if namespace != "" && event.GetObjectMeta().GetNamespace() != namespace {
			continue
		}

		// Kyverno doesn't currectly report the InvoledObject attributes in the Event
		// object that is generated as a result of policy violation.
		// InvolvedObject is of kind Policy or ClusterPolicy when admission control denies request
		// InvolvedObject is actual kind when event is generated during background scan
		// InvolvedObject is actual kind when event is generated
		// during policy evaluation in case of validationFailureAction: Audit
		// Bug: https://github.com/kyverno/kyverno/issues/4234

		auditEvent := event.Source.Component == policyControllerKyvernoEventSource

		admissionEvent := event.Source.Component == admissionControllerKyvernoEventSource

		if listAuditViolations && admissionEvent {
			continue
		}

		if !listAuditViolations && auditEvent {
			continue
		}

		policy := ""
		name := event.InvolvedObject.Name
		if admissionEvent {
			policy = event.InvolvedObject.Name
			name = "NA"
		}

		violation := &policyViolation{
			name:      name,
			kind:      event.InvolvedObject.Kind,
			namespace: event.GetNamespace(),
			policy:    policy,
			message:   event.Message,
			timestamp: event.CreationTimestamp.UTC().Format(time.RFC3339),
		}

		violationsFound = true

		printViolation(violation, streams.Out)
	}

	if !violationsFound {
		fmt.Fprintf(streams.Out, "No events found for policy violations\n\n")
	}

	return nil
}

func gatekeeperExists(factory bbutil.Factory, streams genericclioptions.IOStreams) (bool, error) {

	client, err := factory.GetK8sDynamicClient()
	if err != nil {
		return false, err
	}

	gkCrds, err := gatekeeper.FetchGatekeeperCrds(client)
	if err != nil {
		return false, err
	}

	return len(gkCrds.Items) != 0, nil

}

func listGkViolations(factory bbutil.Factory, streams genericclioptions.IOStreams, namespace string, audit bool) error {

	if audit {
		return listGkAuditViolations(factory, streams, namespace)
	}

	return listGkDenyViolations(factory, streams, namespace)
}

func listGkDenyViolations(factory bbutil.Factory, streams genericclioptions.IOStreams, namespace string) error {

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

		violation := &policyViolation{
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
func listGkAuditViolations(factory bbutil.Factory, streams genericclioptions.IOStreams, namespace string) error {

	client, err := factory.GetK8sDynamicClient()
	if err != nil {
		return err
	}

	gkCrds, err := gatekeeper.FetchGatekeeperCrds(client)
	if err != nil {
		return err
	}

	if len(gkCrds.Items) == 0 {
		fmt.Fprintf(streams.Out, "No violations found in audit\n\n\n")
		return nil
	}

	for _, crd := range gkCrds.Items {
		crdName, _, _ := unstructured.NestedString(crd.Object, "metadata", "name")
		constraints, err := gatekeeper.FetchGatekeeperConstraints(client, crdName)
		if err != nil {
			return err
		}
		for _, c := range constraints.Items {
			// get violations
			violations, _ := getGkConstraintViolations(&c)
			// process violations
			processGkViolations(namespace, violations, crdName, streams)
		}
	}

	return nil
}

func getGkConstraintViolations(resource *unstructured.Unstructured) (*[]policyViolation, error) {

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

	violations := make([]policyViolation, len(statusViolations))

	for i, v := range statusViolations {
		details, _ := v.(map[string]interface{})
		violations[i] = policyViolation{
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

func processGkViolations(namespace string, violations *[]policyViolation, crdName string, streams genericclioptions.IOStreams) {
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

func printViolation(v *policyViolation, w io.Writer) {
	fmt.Fprintf(w, "Time: %s, Resource: %s, Kind: %s, Namespace: %s\n", v.timestamp, v.name, v.kind, v.namespace)
	if v.constraint != "" {
		fmt.Fprintf(w, "Constraint: %s\n", v.policy)
	}
	if v.policy != "" {
		fmt.Fprintf(w, "Policy: %s\n", v.policy)
	}
	fmt.Fprintf(w, "%s\n\n", v.message)
}
