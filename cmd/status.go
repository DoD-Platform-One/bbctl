package cmd

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"repo1.dso.mil/big-bang/product/packages/bbctl/static"
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"

	"github.com/spf13/cobra"
	k8sCoreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sClient "k8s.io/client-go/kubernetes"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"

	helmV2Beta1 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizeV1Beta1 "github.com/fluxcd/kustomize-controller/api/v1beta1"
	sourceV1Beta1 "github.com/fluxcd/source-controller/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	helm "repo1.dso.mil/big-bang/product/packages/bbctl/util/helm"
	outputSchema "repo1.dso.mil/big-bang/product/packages/bbctl/util/output/schemas"
)

var (
	statusUse = `status`

	statusShort = i18n.T(`Show the deployment status of the Big Bang deployment and its subcomponents.`)

	statusLong = templates.LongDesc(i18n.T(`Show the deployment status of Big Bang deployment and its subcomponents.
		This command queries the cluster and returns the deplyoment status of all Big Bang-controlled resources.
	`))

	statusExample = templates.Examples(i18n.T(`
		# Get the overall Big Bang status
		bbctl status`))
)

const (
	statusString = "namespace: %s, name: %s, status: %s\n"
	commandHelp  = "Command Help:\n"
)

// NewStatusCmd - new status command
func NewStatusCmd(factory bbUtil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     statusUse,
		Short:   statusShort,
		Long:    statusLong,
		Example: statusExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return bbStatus(cmd, factory)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	return cmd
}

// podData is the Pod data returned from the Kubernetes cluster
type podData struct {
	namespace string
	name      string
	status    string
}

// statefulSetData is the StatefulSet data returned the Kubernetes cluster
type statefulSetData struct {
	namespace string
	name      string
	replicas  int32
	ready     int32
	status    string
}

// deploymentData is the Deployment data returned the Kubernetes cluster
type deploymentData struct {
	namespace string
	name      string
	replicas  int32
	ready     int32
	status    string
}

// daemonSetData is the DaemonSet data returned the Kubernetes cluster
type daemonSetData struct {
	namespace string
	name      string
	desired   int32
	available int32
	status    string
}

// fluxHRData is the Flux HelmRelease data returned the Kubernetes cluster
type fluxHRData struct {
	namespace string
	name      string
	status    string
}

// fluxGRData is the Flux GitRepository data returned the Kubernetes cluster
type fluxGRData struct {
	namespace string
	name      string
	status    string
}

// fluxKZData is the Flux Kustomization data returned the Kubernetes cluster
type fluxKZData struct {
	namespace string
	name      string
	status    string
}

// bbStatus queries the Kubernetes cluster and gets the Status of the various bigbang-controlled components
func bbStatus(cmd *cobra.Command, factory bbUtil.Factory) error {
	output := &outputSchema.StatusOutput{Name: "Status Summary"}

	// get client-go client
	clientset, err := factory.GetK8sClientset(cmd)
	if err != nil {
		return fmt.Errorf("Failed to get k8s clientset: %w", err)
	}

	//get output client
	outputClient, err := factory.GetOutputClient(cmd)
	if err != nil {
		return fmt.Errorf("Failed to get output client: %w", err)
	}

	// get runtime controller client
	// register the GitOps Toolkit schema definitions
	scheme := runtime.NewScheme()
	_ = sourceV1Beta1.AddToScheme(scheme)
	_ = helmV2Beta1.AddToScheme(scheme)
	_ = kustomizeV1Beta1.AddToScheme(scheme)

	fluxClient, err := factory.GetRuntimeClient(scheme)
	if err != nil {
		return fmt.Errorf("Failed to get runtime client: %w", err)
	}

	// get constants
	constants, err := static.GetDefaultConstants()
	if err != nil {
		return err
	}

	// get helm client
	helmClient, err := factory.GetHelmClient(cmd, constants.BigBangNamespace)
	if err != nil {
		return err
	}

	// get Big Bang helm release status
	bigBangStatus := &outputSchema.CheckStatusOutput{
		Name:   "BigBang Status",
		Output: getBigBangStatus(helmClient),
	}
	output.Statuses = append(output.Statuses, *bigBangStatus)

	// get k8s pod status
	podStatus := &outputSchema.CheckStatusOutput{
		Name:   "Pod Status",
		Output: getPodStatus(clientset),
	}
	output.Statuses = append(output.Statuses, *podStatus)

	// get k8s statefulset status
	statefulSetStatus := &outputSchema.CheckStatusOutput{
		Name:   "StatefulSet Status",
		Output: getStatefulSetStatus(clientset),
	}
	output.Statuses = append(output.Statuses, *statefulSetStatus)

	// get k8s deployment status
	deploymentStatus := &outputSchema.CheckStatusOutput{
		Name:   "k8s Deployment Status",
		Output: getDeploymentStatus(clientset),
	}
	output.Statuses = append(output.Statuses, *deploymentStatus)

	// get k8s daemonset status
	daemonsetStatus := &outputSchema.CheckStatusOutput{
		Name:   "k8s DaemonSet Status",
		Output: getDaemonSetsStatus(clientset),
	}
	output.Statuses = append(output.Statuses, *daemonsetStatus)

	// get flux helm release status
	fluxHelmStatus := &outputSchema.CheckStatusOutput{
		Name:   "Flux Helm Release Status",
		Output: getFluxHelmReleases(fluxClient),
	}
	output.Statuses = append(output.Statuses, *fluxHelmStatus)

	// get flux git repository status
	fluxGitStatus := &outputSchema.CheckStatusOutput{
		Name:   "Flux Git Repository Status",
		Output: getFluxGitRepositories(fluxClient),
	}
	output.Statuses = append(output.Statuses, *fluxGitStatus)

	// get flux kustomization status
	fluxKustomizationStatus := &outputSchema.CheckStatusOutput{
		Name:   "Flux Kustomization Status",
		Output: getFluxKustomizations(fluxClient),
	}
	output.Statuses = append(output.Statuses, *fluxKustomizationStatus)

	outputErr := outputClient.Output(output)
	if outputErr != nil {
		return fmt.Errorf("failed to create status output: %w", outputErr)
	}
	return nil
}

// getBigBangStatus gets the Status of the "bigbang" HelmRelease itself
func getBigBangStatus(helmClient helm.Client) []string {
	// get constants
	constants, err := static.GetDefaultConstants()
	if err != nil {
		return []string{err.Error()}
	}

	release, err := helmClient.GetRelease(constants.BigBangHelmReleaseName)
	if err != nil {
		return []string{"No Big Bang release was found.\n"}
	}
	return []string{fmt.Sprintf("Found %s release version %s status: %s\n", release.Chart.Metadata.Name, release.Chart.Metadata.Version, release.Info.Status)}
}

// getFluxKustomizations queries the cluster for Flux Kustomizations and returns a string with the Status of
// the Kustomizations. If the kustomization is not ready, the status is reported as "Not Ready" and remediation
// steps are provided.
func getFluxKustomizations(fc client.Client) []string {
	outputMessages := []string{}

	// set a deadline for the Kubernetes API operations
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	kustomizationsList := &kustomizeV1Beta1.KustomizationList{}

	listErr := fc.List(ctx, kustomizationsList, &client.ListOptions{})
	if listErr != nil {
		return []string{listErr.Error()}
	}

	// declare empty slice of fluxKZData
	var fluxKZs = []fluxKZData{}

	for _, fkzObj := range kustomizationsList.Items {
		// initialize fluxKZDataObj
		var fluxKZDataObj fluxKZData
		fluxKZDataObj.namespace = fkzObj.ObjectMeta.Namespace
		fluxKZDataObj.name = fkzObj.ObjectMeta.Name

		for _, condition := range fkzObj.Status.Conditions {
			if condition.Type == "Ready" && condition.Status != "True" {
				fluxKZDataObj.status = condition.Message
				// add to list of not ready flux kustomizations
				fluxKZs = append(fluxKZs, fluxKZDataObj)
			}
		}
	}

	if len(kustomizationsList.Items) == 0 {
		outputMessages = append(outputMessages, "No Flux kustomizations were found.\n")
	} else if len(fluxKZs) == 0 {
		outputMessages = append(outputMessages, "All Flux kustomizations are ready.\n")
	} else {
		outputMessages = append(outputMessages, fmt.Sprintf("There are %d Flux kustomizations that are not ready:\n", len(fluxKZs)))
		for _, fluxKZDataObj := range fluxKZs {
			outputMessages = append(outputMessages, fmt.Sprintf(statusString, fluxKZDataObj.namespace, fluxKZDataObj.name, fluxKZDataObj.status))
			outputMessages = append(outputMessages, commandHelp)
			outputMessages = append(outputMessages, fmt.Sprintf("  flux reconcile kustomization %s -n %s --with-source\n", fluxKZDataObj.name, fluxKZDataObj.namespace))
		}
	}

	return outputMessages
}

// getFluxGitRepositories queries the cluster for Flux GitRepository resources and returns a string with the Status of
// the GitRepositories. If the GitRepository is not ready, the status is reported as "Not Ready" and remediation
// steps are provided.
func getFluxGitRepositories(fluxClient client.Client) []string {
	outputMessages := []string{}

	// set a deadline for the Kubernetes API operations
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	fluxGRList := &sourceV1Beta1.GitRepositoryList{}

	listErr := fluxClient.List(ctx, fluxGRList, &client.ListOptions{})
	if listErr != nil {
		return []string{listErr.Error()}
	}

	// declare empty slice of fluxGRData
	var fluxGRs = []fluxGRData{}

	for _, fluxGR := range fluxGRList.Items {
		// initialize fluxGRDataObj
		var fluxGRDataObj fluxGRData
		fluxGRDataObj.namespace = fluxGR.ObjectMeta.Namespace
		fluxGRDataObj.name = fluxGR.ObjectMeta.Name

		for _, condition := range fluxGR.Status.Conditions {
			if condition.Type == "Ready" && condition.Status != "True" {
				fluxGRDataObj.status = condition.Message
				// add to list of not ready flux git repositories
				fluxGRs = append(fluxGRs, fluxGRDataObj)
			}
		}
	}

	if len(fluxGRList.Items) == 0 {
		outputMessages = append(outputMessages, "No Flux git repositories were found.\n")
	} else if len(fluxGRs) == 0 {
		outputMessages = append(outputMessages, "All Flux git repositories are ready.\n")
	} else {
		outputMessages = append(outputMessages, fmt.Sprintf("There are %d Flux git repositories that are not ready:\n", len(fluxGRs)))
		for _, fluxGRDataObj := range fluxGRs {
			outputMessages = append(outputMessages, fmt.Sprintf(statusString, fluxGRDataObj.namespace, fluxGRDataObj.name, fluxGRDataObj.status))
			outputMessages = append(outputMessages, commandHelp)
			outputMessages = append(outputMessages, fmt.Sprintf("  kubectl describe git repository %s -n %s\n", fluxGRDataObj.name, fluxGRDataObj.namespace))
			outputMessages = append(outputMessages, fmt.Sprintf("  flux reconcile source git %s -n %s\n", fluxGRDataObj.name, fluxGRDataObj.namespace))
		}
	}

	return outputMessages
}

// getFluxHelmReleases queries the Kubernetes cluster for Flux HelmRelease resources and returns a string with the Status of
// the HelmReleases. If the HelmRelease is not ready, the status is reported as "Not Ready" and remediation
// steps are provided.
func getFluxHelmReleases(fluxClient client.Client) []string {
	outputMessages := []string{}

	// set a deadline for the Kubernetes API operations
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	helmReleaseList := &helmV2Beta1.HelmReleaseList{}

	listErr := fluxClient.List(ctx, helmReleaseList, &client.ListOptions{})
	if listErr != nil {
		return []string{listErr.Error()}
	}

	// declare empty slice of fluxHRData
	var fluxHRs = []fluxHRData{}

	for _, fluxHRObj := range helmReleaseList.Items {
		// initialize fluxHRDataObj
		var fluxHRDataObj fluxHRData
		fluxHRDataObj.namespace = fluxHRObj.ObjectMeta.Namespace
		fluxHRDataObj.name = fluxHRObj.ObjectMeta.Name

		for _, condition := range fluxHRObj.Status.Conditions {
			if condition.Type == "Ready" && condition.Status != "True" {
				fluxHRDataObj.status = condition.Message
				// add to list of not ready flux helm releases
				fluxHRs = append(fluxHRs, fluxHRDataObj)
			}
		}
	}

	if len(helmReleaseList.Items) == 0 {
		outputMessages = append(outputMessages, "No Flux helm releases were found.\n")
	} else if len(fluxHRs) == 0 {
		outputMessages = append(outputMessages, "All Flux helm releases are reconciled.\n")
	} else {
		outputMessages = append(outputMessages, fmt.Sprintf("There are %d Flux helm releases that are not reconciled:\n", len(fluxHRs)))
		for _, fluxHRDataObj := range fluxHRs {
			outputMessages = append(outputMessages, fmt.Sprintf(statusString, fluxHRDataObj.namespace, fluxHRDataObj.name, fluxHRDataObj.status))
			outputMessages = append(outputMessages, commandHelp)
			outputMessages = append(outputMessages, fmt.Sprintf("  flux suspend helm release %s -n %s\n", fluxHRDataObj.name, fluxHRDataObj.namespace))
			outputMessages = append(outputMessages, fmt.Sprintf("  flux resume helm release %s -n %s\n", fluxHRDataObj.name, fluxHRDataObj.namespace))
			outputMessages = append(outputMessages, fmt.Sprintf("  flux reconcile helm release %s -n %s --with-source\n", fluxHRDataObj.name, fluxHRDataObj.namespace))
		}
	}

	return outputMessages
}

// getDaemonSetsStatus queries the Kubernetes cluster for DaemonSet resources and returns a string with the Status of
// the DaemonSets. If the DaemonSets are not available, the status is reported as "Not Available" and manual debugging
// steps are provided.
func getDaemonSetsStatus(clientset k8sClient.Interface) []string {
	outputMessages := []string{}

	daemonSetList, err := clientset.AppsV1().DaemonSets("").List(context.Background(), metaV1.ListOptions{})
	if err != nil {
		return []string{err.Error()}
	}

	// declare empty slice of DmstData
	var daemonSetDataList = []daemonSetData{}

	// iterate daemonsets
	for _, daemonSetObj := range daemonSetList.Items {
		// initialize daemonSetData
		var daemonSetDataObj daemonSetData
		daemonSetDataObj.namespace = daemonSetObj.ObjectMeta.Namespace
		daemonSetDataObj.name = daemonSetObj.ObjectMeta.Name
		daemonSetDataObj.desired = daemonSetObj.Status.DesiredNumberScheduled
		daemonSetDataObj.available = daemonSetObj.Status.NumberAvailable

		if daemonSetDataObj.available < daemonSetDataObj.desired {
			daemonSetDataObj.status = "Not Available " + strconv.FormatInt(int64(daemonSetDataObj.available), 10) + "/" + strconv.FormatInt(int64(daemonSetDataObj.desired), 10)
			// add to list of not ready daemonsets
			daemonSetDataList = append(daemonSetDataList, daemonSetDataObj)
		}
	}

	if len(daemonSetList.Items) == 0 {
		outputMessages = append(outputMessages, "No Daemonsets were found.\n")
	} else if len(daemonSetDataList) == 0 {
		outputMessages = append(outputMessages, "All Daemonsets are available.\n")
	} else {
		outputMessages = append(outputMessages, fmt.Sprintf("There are %d DaemonSets that are not available:\n", len(daemonSetDataList)))
		for _, daemonSetDataObj := range daemonSetDataList {
			outputMessages = append(outputMessages, fmt.Sprintf(statusString, daemonSetDataObj.namespace, daemonSetDataObj.name, daemonSetDataObj.status))
			outputMessages = append(outputMessages, commandHelp)
			outputMessages = append(outputMessages, fmt.Sprintf("  kubectl describe daemonset %s -n %s\n", daemonSetDataObj.name, daemonSetDataObj.namespace))
			outputMessages = append(outputMessages, fmt.Sprintf("  use kubectl to view logs of any daemonset pods in namespace %s\n", daemonSetDataObj.namespace))
		}
	}

	return outputMessages
}

// getDeploymentStatus queries the Kubernetes cluster for Deployment resources and returns a string with the Status of
// the Deployments. If the Deployments are not available, the status is reported as "Not Available" and manual debugging
// steps are provided.
func getDeploymentStatus(clientset k8sClient.Interface) []string {
	outputMessages := []string{}

	deploymentList, err := clientset.AppsV1().Deployments("").List(context.Background(), metaV1.ListOptions{})
	if err != nil {
		return []string{err.Error()}
	}

	// declare empty slice of DpmtData
	var deploymentDataList = []deploymentData{}

	// iterate deployments to determine if requested replicas equal ready replicas
	for _, deploymentObject := range deploymentList.Items {
		// initialize deploymentData
		var deploymentDataObj deploymentData
		deploymentDataObj.namespace = deploymentObject.ObjectMeta.Namespace
		deploymentDataObj.name = deploymentObject.ObjectMeta.Name
		deploymentDataObj.replicas = deploymentObject.Status.Replicas
		deploymentDataObj.ready = deploymentObject.Status.ReadyReplicas

		if deploymentDataObj.ready < deploymentDataObj.replicas {
			deploymentDataObj.status = "Not Ready " + strconv.FormatInt(int64(deploymentDataObj.ready), 10) + "/" + strconv.FormatInt(int64(deploymentDataObj.replicas), 10)
			// add to list of not ready Deployments
			deploymentDataList = append(deploymentDataList, deploymentDataObj)
		}
	}

	if len(deploymentList.Items) == 0 {
		outputMessages = append(outputMessages, "No Deployments were found.\n")
	} else if len(deploymentDataList) == 0 {
		outputMessages = append(outputMessages, "All Deployments are ready.\n")
	} else {
		outputMessages = append(outputMessages, fmt.Sprintf("There are %d k8s Deployments that are not ready:\n", len(deploymentDataList)))

		for _, deploymentDataObj := range deploymentDataList {
			outputMessages = append(outputMessages, fmt.Sprintf(statusString, deploymentDataObj.namespace, deploymentDataObj.name, deploymentDataObj.status))
			outputMessages = append(outputMessages, commandHelp)
			outputMessages = append(outputMessages, fmt.Sprintf("  Use kubectl to check the logs of the related pods in namespace %s\n", deploymentDataObj.namespace))
		}
	}

	return outputMessages
}

// getStatefulSetStatus queries the Kubernetes cluster for StatefulSet resources and returns a string with the Status of
// the StatefulSets. If the StatefulSets are not available, the status is reported as "Not Available" and manual debugging
// steps are provided.
func getStatefulSetStatus(clientset k8sClient.Interface) []string {
	outputMessages := []string{}

	statefulSetList, err := clientset.AppsV1().StatefulSets("").List(context.Background(), metaV1.ListOptions{})
	if err != nil {
		return []string{err.Error()}
	}

	// declare empty slice of StsData
	var statefulSetDataList = []statefulSetData{}

	// iterate statefulsets to determine if requested replicas equal ready replicas
	for _, statefulSetObj := range statefulSetList.Items {
		// initialize podData
		var statefulSetDataObj statefulSetData
		statefulSetDataObj.namespace = statefulSetObj.ObjectMeta.Namespace
		statefulSetDataObj.name = statefulSetObj.ObjectMeta.Name
		statefulSetDataObj.replicas = statefulSetObj.Status.Replicas
		statefulSetDataObj.ready = statefulSetObj.Status.ReadyReplicas

		if statefulSetDataObj.ready < statefulSetDataObj.replicas {
			statefulSetDataObj.status = "Not Ready " + strconv.FormatInt(int64(statefulSetDataObj.ready), 10) + "/" + strconv.FormatInt(int64(statefulSetDataObj.replicas), 10)
			// add to list of not ready sts
			statefulSetDataList = append(statefulSetDataList, statefulSetDataObj)
		}
	}

	if len(statefulSetList.Items) == 0 {
		outputMessages = append(outputMessages, "No StatefulSets were found.\n")
	} else if len(statefulSetDataList) == 0 {
		outputMessages = append(outputMessages, "All StatefulSets are ready.\n")
	} else {
		outputMessages = append(outputMessages, fmt.Sprintf("There are %d StatefulSets that are not ready:\n", len(statefulSetDataList)))
		for _, statefulSetDataObj := range statefulSetDataList {
			outputMessages = append(outputMessages, fmt.Sprintf(statusString, statefulSetDataObj.namespace, statefulSetDataObj.name, statefulSetDataObj.status))
			outputMessages = append(outputMessages, commandHelp)
			outputMessages = append(outputMessages, fmt.Sprintf("  Use kubectl to check the logs of the related pods in namespace %s\n", statefulSetDataObj.namespace))
		}
	}

	return outputMessages
}

// getPodStatus queries the Kubernetes cluster for Pod resources and returns a string with the Status of
// the Pods. If the Pods are not available, the status is reported as "Not Available" and manual debugging
// steps are provided.
func getPodStatus(clientset k8sClient.Interface) []string {
	outputMessages := []string{}

	podsList, err := clientset.CoreV1().Pods("").List(context.Background(), metaV1.ListOptions{})
	if err != nil {
		return []string{err.Error()}
	}

	// declare empty slice of podData
	var podDataList = []podData{}

	// iterate bad pods to extract status
	for _, podObj := range podsList.Items {
		// initialize podData
		var podDataObj podData
		podDataObj.namespace = podObj.Namespace
		podDataObj.name = podObj.Name

		podReady := true

		// add bad pods to slice of podData
		// there are 5 possible phases: Pending, Running, Succeeded, Failed, Unknown
		switch podObj.Status.Phase {
		case "Running":
			// check if all containers are ready
			getContainerStatus(podObj.Status.ContainerStatuses, &podDataObj, &podReady, false)
			// process pod status
			processPodStatus(&podDataObj, &podDataList, podReady)

		case "Succeeded":
			// do nothing

		default:
			// evaluate status of init containers
			getContainerStatus(podObj.Status.InitContainerStatuses, &podDataObj, &podReady, true)
			// process pod status
			processPodStatus(&podDataObj, &podDataList, podReady)
		}
	}

	if len(podDataList) == 0 {
		outputMessages = append(outputMessages, "All pods are ready.\n")
	} else {
		outputMessages = append(outputMessages, fmt.Sprintf("There are %d pods that are not ready:\n", len(podDataList)))
		for _, pod := range podDataList {
			outputMessages = append(outputMessages, fmt.Sprintf(statusString, pod.namespace, pod.name, pod.status))
			outputMessages = append(outputMessages, commandHelp)
			outputMessages = append(outputMessages, fmt.Sprintf("  kubectl logs %s -n %s\n", pod.name, pod.namespace))
		}
	}

	return outputMessages
}

// processPodStatus processes the status of a pod and adds it to the list of pods that are ready
func processPodStatus(pod *podData, pods *[]podData, podReady bool) {
	if !podReady {
		if pod.status == "" {
			pod.status = "error"
		}
		// add to list of bad pods
		*pods = append(*pods, *pod)
	}
}

// getContainerStatus processes the status of a pod's containers and adds it to the list of pods that are ready
func getContainerStatus(containerStatuses []k8sCoreV1.ContainerStatus, pod *podData, podReady *bool, isInit bool) {
	var shortStatus string
	var longStatus string

	if isInit {
		longStatus = "init:CrashLoopBackOff"
		shortStatus = "init:"
	} else {
		longStatus = "CrashLoopBackOff"
		shortStatus = ""
	}

	for _, cs := range containerStatuses {
		if !cs.Ready {
			*podReady = false
			if cs.State.Waiting != nil {
				if pod.status != longStatus {
					pod.status = shortStatus + cs.State.Waiting.Reason
				}
			}
		}
	}
}
