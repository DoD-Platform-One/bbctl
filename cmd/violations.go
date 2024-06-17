package cmd

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/spf13/cobra"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"

	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/gatekeeper"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/kyverno"

	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
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

// NewViolationsCmd - new violations command
func NewViolationsCmd(factory bbUtil.Factory, streams genericIOOptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     violationsUse,
		Short:   violationsShort,
		Long:    violationsLong,
		Example: violationsExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return getViolations(cmd, factory, streams)
		},
	}

	loggingClient := factory.GetLoggingClient()
	configClient, err := factory.GetConfigClient(cmd)
	loggingClient.HandleError("Error getting config client: %v", err)

	loggingClient.HandleError(
		"Error binding flags: %v",
		configClient.SetAndBindFlag(
			"audit",
			false,
			"list violations in audit mode",
		),
	)

	return cmd
}

func getViolations(cmd *cobra.Command, factory bbUtil.Factory, streams genericIOOptions.IOStreams) error {
	logger := factory.GetLoggingClient()
	configClient, err := factory.GetConfigClient(cmd)
	if err != nil {
		return err
	}
	config := configClient.GetConfig()

	namespace := config.UtilK8sConfiguration.Namespace
	audit := config.ViolationsConfiguration.Audit

	gkFound, err := gatekeeperExists(cmd, factory, streams)
	if err != nil {
		return err
	}

	if gkFound {
		logger.Debug("Gatekeeper exists in cluster. Checking for Gatekeeper violations.")
		return listGkViolations(cmd, factory, streams, namespace, audit)
	}

	kyvernoFound, err := kyvernoExists(cmd, factory, streams)
	if err != nil {
		return err
	}

	if kyvernoFound {
		logger.Debug("Kyverno exists in cluster. Checking for Kyverno violations.")
		return listKyvernoViolations(cmd, factory, streams, namespace, audit)
	}

	return nil
}

func kyvernoExists(cmd *cobra.Command, factory bbUtil.Factory, _ genericIOOptions.IOStreams) (bool, error) {
	client, err := factory.GetK8sDynamicClient(cmd)
	if err != nil {
		return false, err
	}

	kyvernoCrds, err := kyverno.FetchKyvernoCrds(client)
	if err != nil {
		return false, err
	}

	return len(kyvernoCrds.Items) != 0, nil
}

func listKyvernoViolations(cmd *cobra.Command, factory bbUtil.Factory, streams genericIOOptions.IOStreams, namespace string, listAuditViolations bool) error {
	client, err := factory.GetK8sClientset(cmd)
	if err != nil {
		return err
	}

	events, err := client.CoreV1().Events("").List(context.TODO(), metaV1.ListOptions{
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

		// Kyverno doesn't correctly report the InvolvedObject attributes in the Event
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

func gatekeeperExists(cmd *cobra.Command, factory bbUtil.Factory, _ genericIOOptions.IOStreams) (bool, error) {
	client, err := factory.GetK8sDynamicClient(cmd)
	if err != nil {
		return false, err
	}

	gkCrds, err := gatekeeper.FetchGatekeeperCrds(client)
	if err != nil {
		return false, err
	}

	return len(gkCrds.Items) != 0, nil
}

func listGkViolations(cmd *cobra.Command, factory bbUtil.Factory, streams genericIOOptions.IOStreams, namespace string, audit bool) error {
	if audit {
		return listGkAuditViolations(cmd, factory, streams, namespace)
	}

	return listGkDenyViolations(cmd, factory, streams, namespace)
}

func listGkDenyViolations(cmd *cobra.Command, factory bbUtil.Factory, streams genericIOOptions.IOStreams, namespace string) error {
	client, err := factory.GetK8sClientset(cmd)
	if err != nil {
		return err
	}

	events, err := client.CoreV1().Events("").List(context.TODO(), metaV1.ListOptions{
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
func listGkAuditViolations(cmd *cobra.Command, factory bbUtil.Factory, streams genericIOOptions.IOStreams, namespace string) error {
	client, err := factory.GetK8sDynamicClient(cmd)
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

func processGkViolations(namespace string, violations *[]policyViolation, crdName string, streams genericIOOptions.IOStreams) {
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
