package cmd

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"

	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/config"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/gatekeeper"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/kyverno"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/log"
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

// violationsCmdHelper is a container to store the various shared clients and fields
// that are used by the various components of the violations command
//
// This prevents us from having to pass the same clients and fields to each component
// or create new instasnces of clients within each function
type violationsCmdHelper struct {
	// Clients
	k8sClient    dynamic.Interface
	k8sClientSet kubernetes.Interface
	configClient *config.ConfigClient
	logger       log.Client
	streams      genericIOOptions.IOStreams
}

func newViolationsCmdHelper(cmd *cobra.Command, factory bbUtil.Factory) (*violationsCmdHelper, error) {
	k8sClient, err := factory.GetK8sDynamicClient(cmd)
	if err != nil {
		return nil, err
	}

	k8sClientSet, err := factory.GetK8sClientset(cmd)
	if err != nil {
		return nil, err
	}

	loggingClient := factory.GetLoggingClient()

	configClient, err := factory.GetConfigClient(cmd)
	if err != nil {
		return nil, err
	}

	streams := factory.GetIOStream()

	return &violationsCmdHelper{
		k8sClient:    k8sClient,
		k8sClientSet: k8sClientSet,
		logger:       loggingClient,
		configClient: configClient,
		streams:      *streams,
	}, nil
}

// NewViolationsCmd - new violations command
func NewViolationsCmd(factory bbUtil.Factory) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:     violationsUse,
		Short:   violationsShort,
		Long:    violationsLong,
		Example: violationsExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := newViolationsCmdHelper(cmd, factory)
			if err != nil {
				return fmt.Errorf("Error getting violations helper client: %v", err)
			}

			return v.getViolations()
		},
	}

	configClient, clientError := factory.GetConfigClient(cmd)
	if clientError != nil {
		return nil, fmt.Errorf("Unable to get config client: %w", clientError)
	}

	flagError := configClient.SetAndBindFlag(
		"audit",
		false,
		"list violations in audit mode",
	)
	if flagError != nil {
		return nil, fmt.Errorf("Error setting and binding flags: %w", flagError)
	}

	return cmd, nil
}

// getViolations detects if the cluster has gatekeeper or kyverno installed and
// prints out the violations in the cluster.
func (v *violationsCmdHelper) getViolations() error {
	config := v.configClient.GetConfig()

	namespace := config.UtilK8sConfiguration.Namespace
	audit := config.ViolationsConfiguration.Audit

	// Aggregate fetching errors to prevent short-circuiting
	var errs []error

	gkFound, err := v.gatekeeperExists()
	if err != nil {
		errs = append(errs, err)
	}

	kyvernoFound, err := v.kyvernoExists()
	if err != nil {
		errs = append(errs, err)
	}

	if gkFound {
		v.logger.Debug("Gatekeeper exists in cluster. Checking for Gatekeeper violations.")
		if err := v.listGkViolations(namespace, audit); err != nil {
			errs = append(errs, fmt.Errorf("error listing gatekeeper violations: %v", err))
		}
	}

	if kyvernoFound {
		v.logger.Debug("Kyverno exists in cluster. Checking for Kyverno violations.")
		if err := v.listKyvernoViolations(namespace, audit); err != nil {
			errs = append(errs, fmt.Errorf("error listing kyverno violations: %v", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("Errors occurred while listing violations: %v", errs)
	}

	return nil
}

// kyvernoExists checks if kyverno is installed in the cluster by checking if the kyverno crds are present
func (v *violationsCmdHelper) kyvernoExists() (bool, error) {
	kyvernoCrds, err := kyverno.FetchKyvernoCrds(v.k8sClient)
	if err != nil {
		return false, err
	}

	return len(kyvernoCrds.Items) != 0, nil
}

// listKyvernoViolations prints the violations in the cluster using the kyverno crds
func (v *violationsCmdHelper) listKyvernoViolations(namespace string, listAuditViolations bool) error {
	events, err := v.k8sClientSet.CoreV1().Events("").List(context.TODO(), metaV1.ListOptions{
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

		printViolation(violation, v.streams.Out)
	}

	if !violationsFound {
		fmt.Fprintf(v.streams.Out, "No events found for policy violations\n\n")
	}

	return nil
}

// gatekeeperExists	queries the cluster for gatekeeper CRDs and returns true if the CRDs are present
func (v *violationsCmdHelper) gatekeeperExists() (bool, error) {
	gkCrds, err := gatekeeper.FetchGatekeeperCrds(v.k8sClient)
	if err != nil {
		return false, err
	}

	return len(gkCrds.Items) != 0, nil
}

// listGkViolations prints the violations in the cluster using the gatekeeper crds. If audit is true, it prints the audit violations,
// otherwise it prints the deny violations by default.
func (v *violationsCmdHelper) listGkViolations(namespace string, audit bool) error {
	if audit {
		return v.listGkAuditViolations(namespace)
	}

	return v.listGkDenyViolations(namespace)
}

// listGkDenyViolations prints the deny violations in the cluster using the gatekeeper crds
func (v *violationsCmdHelper) listGkDenyViolations(namespace string) error {
	events, err := v.k8sClientSet.CoreV1().Events("").List(context.TODO(), metaV1.ListOptions{
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

		printViolation(violation, v.streams.Out)
	}

	if !violationsFound {
		fmt.Fprintf(v.streams.Out, "No events found for deny violations\n\n")
		fmt.Fprintf(v.streams.Out, "Do you have the following values defined for the gatekeeper chart?\n\n")
		fmt.Fprintf(v.streams.Out, "gatekeeper:\n")
		fmt.Fprintf(v.streams.Out, "  emitAdmissionEvents: true\n")
		fmt.Fprintf(v.streams.Out, "  logDenies: true\n\n")
		fmt.Fprintf(v.streams.Out, "Note that violations in dryrun and warn mode are not effected by these settings.\n")
		fmt.Fprintf(v.streams.Out, "To list dryrun violations, use --audit flag.\n")
	}

	return nil
}

// listGkAuditViolations prints the audit violations in the cluster using the gatekeeper crds
func (v *violationsCmdHelper) listGkAuditViolations(namespace string) error {
	gkCrds, err := gatekeeper.FetchGatekeeperCrds(v.k8sClient)
	if err != nil {
		return err
	}

	if len(gkCrds.Items) == 0 {
		fmt.Fprintf(v.streams.Out, "No violations found in audit\n\n\n")
		return nil
	}

	for _, crd := range gkCrds.Items {
		crdName, _, _ := unstructured.NestedString(crd.Object, "metadata", "name")
		constraints, err := gatekeeper.FetchGatekeeperConstraints(v.k8sClient, crdName)
		if err != nil {
			return err
		}
		for _, c := range constraints.Items {
			// get violations
			violations, _ := getGkConstraintViolations(&c)
			// process violations
			v.processGkViolations(namespace, violations, crdName)
		}
	}

	return nil
}

// getGkConstraintViolations queries the cluster using dynamic client to get audit violation information from gatekeeper constraint crds
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

// processGkViolations filters the violations based on the namespace and prints them out
func (v *violationsCmdHelper) processGkViolations(namespace string, violations *[]policyViolation, crdName string) {
	if len(*violations) != 0 {
		fmt.Fprintf(v.streams.Out, "%s\n\n", crdName)
		violationsFound := false
		for _, violation := range *violations {
			if namespace != "" && violation.namespace != namespace {
				continue
			}
			violationsFound = true
			printViolation(&violation, v.streams.Out)
		}
		if !violationsFound {
			fmt.Fprintf(v.streams.Out, "No violations found in audit\n\n\n")
		}
	}
}

// printViolation prints the violation information to the defined io.Writer
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
