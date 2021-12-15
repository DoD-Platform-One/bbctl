package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	k8sclient "k8s.io/client-go/kubernetes"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	bbutil "repo1.dso.mil/platform-one/big-bang/apps/product-tools/bbctl/util"

	helmv2beta1 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev1beta1 "github.com/fluxcd/kustomize-controller/api/v1beta1"
	sourcev1beta1 "github.com/fluxcd/source-controller/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	helm "repo1.dso.mil/platform-one/big-bang/apps/product-tools/bbctl/util/helm"
)

var (
	statusUse = `status`

	statusShort = i18n.T(`Show status of BigBang deployment.`)

	statusLong = templates.LongDesc(i18n.T(`Show status of BigBang deployment.`))

	statusExample = templates.Examples(i18n.T(`
		# Get overall status
		bbctl status`))
)

// NewStatusCmd - new status command
func NewStatusCmd(factory bbutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {

	cmd := &cobra.Command{
		Use:     statusUse,
		Short:   statusShort,
		Long:    statusLong,
		Example: statusExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(bbStatus(factory, streams))
		},
	}

	return cmd
}

// Pod
type podData struct {
	namespace string
	name      string
	status    string
}

// StatefulSet
type stsData struct {
	namespace string
	name      string
	replicas  int32
	ready     int32
	status    string
}

// deployments
type dpmtData struct {
	namespace string
	name      string
	replicas  int32
	ready     int32
	status    string
}

// Daemonsets
type dmstData struct {
	namespace string
	name      string
	desired   int32
	available int32
	status    string
}

// Flux HelmRelease Data
type fluxHRData struct {
	namespace string
	name      string
	status    string
}

// Flux GitRepository Data
type fluxGRData struct {
	namespace string
	name      string
	status    string
}

// Flux Kustomizations Data
type fluxKZData struct {
	namespace string
	name      string
	status    string
}

func bbStatus(factory bbutil.Factory, streams genericclioptions.IOStreams) error {

	// get client-go client
	clientset, err := factory.GetK8sClientset()
	if err != nil {
		return err
	}

	// get runtime controller client
	// register the GitOps Toolkit schema definitions
	scheme := runtime.NewScheme()
	_ = sourcev1beta1.AddToScheme(scheme)
	_ = helmv2beta1.AddToScheme(scheme)
	_ = kustomizev1beta1.AddToScheme(scheme)

	fluxclient, err := factory.GetRuntimeClient(scheme)
	if err != nil {
		return err
	}

	// get helm client
	helmclient, err := factory.GetHelmClient(BigBangNamespace)
	if err != nil {
		return err
	}

	// get BigBang helm release status
	fmt.Println(getBigBangStatus(helmclient))

	// get k8s pod status
	fmt.Println(getPodStatus(clientset))

	// get k8s statefulset status
	fmt.Println(getStsStatus(clientset))

	// get k8s deployment status
	fmt.Println(getDpmtStatus(clientset))

	// get k8s daemonset status
	fmt.Println(getDmstStatus(clientset))

	// get flux helmrelease status
	fmt.Println(getFluxHelmReleases(fluxclient))

	// get flux gitrepository status
	fmt.Println(getFluxGitRepositories(fluxclient))

	// get flux kustomization status
	fmt.Println(getFluxKustomizations(fluxclient))

	return nil
}

func getBigBangStatus(helmclient helm.Client) string {

	var sb strings.Builder

	release, err := helmclient.GetRelease(BigBangHelmReleaseName)
	if err != nil {
		sb.WriteString("No BigBang release was found.\n")
		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("Found %s release version %s status: %s\n", release.Chart.Metadata.Name, release.Chart.Metadata.Version, release.Info.Status))
	
	return sb.String()
}

func getFluxKustomizations(fc client.Client) string {

	var sb strings.Builder

	// set a deadline for the Kubernetes API operations
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	kzl := &kustomizev1beta1.KustomizationList{}

	listErr := fc.List(ctx, kzl, &client.ListOptions{})
	if listErr != nil {
		sb.WriteString(listErr.Error())
		return sb.String()
	}

	// declare empty slice of fluxKZData
	var fkzs = []fluxKZData{}

	for _, fkzObj := range kzl.Items {
		// initalize fluxKZData
		var fkzd fluxKZData
		fkzd.namespace = fkzObj.ObjectMeta.Namespace
		fkzd.name = fkzObj.ObjectMeta.Name

		for _, cndtn := range fkzObj.Status.Conditions {
			if cndtn.Type == "Ready" && cndtn.Status != "True" {
				fkzd.status = cndtn.Message
				// add to list of not ready flux kustomizations
				fkzs = append(fkzs, fkzd)
			}
		}
	}

	if len(kzl.Items) == 0 {
		sb.WriteString("No Flux kustomizations were found.\n")
	} else if len(fkzs) == 0 {
		sb.WriteString("All Flux kustomizations are ready.\n")
	} else {
		sb.WriteString(fmt.Sprintf("There are %d Flux kustomizations that are not ready:\n", len(fkzs)))
		for _, fkzd := range fkzs {
			sb.WriteString(fmt.Sprintf("namespace: %s, name: %s, status: %s\n", fkzd.namespace, fkzd.name, fkzd.status))
		}
	}

	return sb.String()
}

func getFluxGitRepositories(fc client.Client) string {

	var sb strings.Builder

	// set a deadline for the Kubernetes API operations
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	grl := &sourcev1beta1.GitRepositoryList{}

	listErr := fc.List(ctx, grl, &client.ListOptions{})
	if listErr != nil {
		sb.WriteString(listErr.Error())
		return sb.String()
	}

	// declare empty slice of fluxGRData
	var fgrs = []fluxGRData{}

	for _, fgrObj := range grl.Items {
		// initalize fluxGRData
		var fgrd fluxGRData
		fgrd.namespace = fgrObj.ObjectMeta.Namespace
		fgrd.name = fgrObj.ObjectMeta.Name

		for _, cndtn := range fgrObj.Status.Conditions {
			if cndtn.Type == "Ready" && cndtn.Status != "True" {
				fgrd.status = cndtn.Message
				// add to list of not ready flux gitrepositories
				fgrs = append(fgrs, fgrd)
			}
		}
	}

	if len(grl.Items) == 0 {
		sb.WriteString("No Flux gitrepositories were found.\n")
	} else if len(fgrs) == 0 {
		sb.WriteString("All Flux gitrepositories are ready.\n")
	} else {
		sb.WriteString(fmt.Sprintf("There are %d Flux gitrepositories that are not ready:\n", len(fgrs)))
		for _, fgrd := range fgrs {
			sb.WriteString(fmt.Sprintf("namespace: %s, name: %s, status: %s\n", fgrd.namespace, fgrd.name, fgrd.status))
		}
	}

	return sb.String()
}

func getFluxHelmReleases(fc client.Client) string {

	var sb strings.Builder

	// set a deadline for the Kubernetes API operations
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	hrl := &helmv2beta1.HelmReleaseList{}

	listErr := fc.List(ctx, hrl, &client.ListOptions{})
	if listErr != nil {
		sb.WriteString(listErr.Error())
		return sb.String()
	}

	// declare empty slice of fluxHRData
	var fhrs = []fluxHRData{}

	for _, fhrObj := range hrl.Items {
		// initalize fluxHRData
		var fhrd fluxHRData
		fhrd.namespace = fhrObj.ObjectMeta.Namespace
		fhrd.name = fhrObj.ObjectMeta.Name

		for _, cndtn := range fhrObj.Status.Conditions {
			if cndtn.Type == "Ready" && cndtn.Status != "True" {
				fhrd.status = cndtn.Message
				// add to list of not ready flux helmreleases
				fhrs = append(fhrs, fhrd)
			}
		}
	}

	if len(hrl.Items) == 0 {
		sb.WriteString("No Flux helmreleases were found.\n")
	} else if len(fhrs) == 0 {
		sb.WriteString("All Flux helmreleases are reconciled.\n")
	} else {
		sb.WriteString(fmt.Sprintf("There are %d Flux helmreleases that are not reconciled:\n", len(fhrs)))
		for _, fhrd := range fhrs {
			sb.WriteString(fmt.Sprintf("namespace: %s, name: %s, status: %s\n", fhrd.namespace, fhrd.name, fhrd.status))
		}
	}

	return sb.String()
}

func getDmstStatus(clientset k8sclient.Interface) string {

	var sb strings.Builder

	dmstObj, err := clientset.AppsV1().DaemonSets("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		sb.WriteString(err.Error())
		return sb.String()
	}

	// declare empty slice of DmstData
	var dmsts = []dmstData{}

	// iterate daemonsets
	for _, dmstObj := range dmstObj.Items {
		// initialize dmstData
		var dmst dmstData
		dmst.namespace = dmstObj.ObjectMeta.Namespace
		dmst.name = dmstObj.ObjectMeta.Name
		dmst.desired = dmstObj.Status.DesiredNumberScheduled
		dmst.available = dmstObj.Status.NumberAvailable

		if dmst.available < dmst.desired {
			dmst.status = "Not Available " + strconv.FormatInt(int64(dmst.available), 10) + "/" + strconv.FormatInt(int64(dmst.desired), 10)
			// add to list of not ready daemonsets
			dmsts = append(dmsts, dmst)
		}
	}

	if len(dmstObj.Items) == 0 {
		sb.WriteString("No Daemonsets were found.\n")
	} else if len(dmsts) == 0 {
		sb.WriteString("All Daemonsets are available.\n")
	} else {
		sb.WriteString(fmt.Sprintf("There are %d DaemonSets that are not available:\n", len(dmsts)))
		for _, dmst := range dmsts {
			sb.WriteString(fmt.Sprintf("namespace: %s, name: %s, status: %s\n", dmst.namespace, dmst.name, dmst.status))
		}
	}

	return sb.String()
}

func getDpmtStatus(clientset k8sclient.Interface) string {

	var sb strings.Builder

	dpmtObj, err := clientset.AppsV1().Deployments("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		sb.WriteString(err.Error())
		return sb.String()
	}

	// declare empty slice of DpmtData
	var dpmts = []dpmtData{}

	// iterate deployments to determine if requested replicas equal ready replicas
	for _, dpmtObj := range dpmtObj.Items {
		// initialize dpmtData
		var dpmt dpmtData
		dpmt.namespace = dpmtObj.ObjectMeta.Namespace
		dpmt.name = dpmtObj.ObjectMeta.Name
		dpmt.replicas = dpmtObj.Status.Replicas
		dpmt.ready = dpmtObj.Status.ReadyReplicas

		if dpmt.ready < dpmt.replicas {
			dpmt.status = "Not Ready " + strconv.FormatInt(int64(dpmt.ready), 10) + "/" + strconv.FormatInt(int64(dpmt.replicas), 10)
			// add to list of not ready Deployments
			dpmts = append(dpmts, dpmt)
		}
	}

	if len(dpmtObj.Items) == 0 {
		sb.WriteString("No Deployments were found.\n")
	} else if len(dpmts) == 0 {
		sb.WriteString("All Deployments are ready.\n")
	} else {
		sb.WriteString(fmt.Sprintf("There are %d k8s Deployments that are not ready:\n", len(dpmts)))
		for _, dpmt := range dpmts {
			sb.WriteString(fmt.Sprintf("namespace: %s, name: %s, status: %s\n", dpmt.namespace, dpmt.name, dpmt.status))
		}
	}

	return sb.String()
}

func getStsStatus(clientset k8sclient.Interface) string {

	var sb strings.Builder

	stsObj, err := clientset.AppsV1().StatefulSets("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		sb.WriteString(err.Error())
		return sb.String()
	}

	// declare empty slice of StsData
	var stss = []stsData{}

	// iterate statefulsets to determine if requested replicas equal ready replicas
	for _, stsObj := range stsObj.Items {
		// initialize podData
		var sts stsData
		sts.namespace = stsObj.ObjectMeta.Namespace
		sts.name = stsObj.ObjectMeta.Name
		sts.replicas = stsObj.Status.Replicas
		sts.ready = stsObj.Status.ReadyReplicas

		if sts.ready < sts.replicas {
			sts.status = "Not Ready " + strconv.FormatInt(int64(sts.ready), 10) + "/" + strconv.FormatInt(int64(sts.replicas), 10)
			// add to list of not ready sts
			stss = append(stss, sts)
		}
	}

	if len(stsObj.Items) == 0 {
		sb.WriteString("No StatefulSets were found.\n")
	} else if len(stss) == 0 {
		sb.WriteString("All StatefulSets are ready.\n")
	} else {
		sb.WriteString(fmt.Sprintf("There are %d StatefulSets that are not ready:\n", len(stss)))
		for _, sts := range stss {
			sb.WriteString(fmt.Sprintf("namespace: %s, name: %s, status: %s\n", sts.namespace, sts.name, sts.status))
		}
	}

	return sb.String()
}

func getPodStatus(clientset k8sclient.Interface) string {

	var sb strings.Builder

	podsObj, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		sb.WriteString(err.Error())
		return sb.String()
	}

	// declare empty slice of podData
	var pods = []podData{}

	// iterate bad pods to extract status
	for _, podObj := range podsObj.Items {
		// initialize podData
		var pod podData
		pod.namespace = podObj.Namespace
		pod.name = podObj.Name

		podready := true

		// add bad pods to slice of podData
		// there are 5 possible phases: Pending, Running, Succeeded, Failed, Unknown
		switch podObj.Status.Phase {
		case "Running":
			// check if all containers are ready
			for _, cs := range podObj.Status.ContainerStatuses {
				if !cs.Ready {
					podready = false
					if cs.State.Waiting != nil {
						if pod.status != "CrashLoopBackOff" {
							pod.status = cs.State.Waiting.Reason
						}
					}
				}
			}

			if !podready {
				if pod.status == "" {
					pod.status = "error"
				}
				// add to list of bad pods
				pods = append(pods, pod)
			}

		case "Succeeded":
			// do nothing
		default:
			// evaluate status of init containers
			for _, ics := range podObj.Status.InitContainerStatuses {
				if !ics.Ready {
					podready = false
					if ics.State.Waiting != nil {
						if pod.status != "init:CrashLoopBackOff" {
							pod.status = "init:" + ics.State.Waiting.Reason
						}
					}
				}
			}

			if !podready {
				if pod.status == "" {
					pod.status = "error"
				}
				// add to list of bad pods
				pods = append(pods, pod)
			}

		}
	}

	if len(pods) == 0 {
		sb.WriteString("All pods are ready.\n")
	} else {
		sb.WriteString(fmt.Sprintf("There are %d pods that are not ready:\n", len(pods)))
		for _, pod := range pods {
			sb.WriteString(fmt.Sprintf("namespace: %s, name: %s, status: %s\n", pod.namespace, pod.name, pod.status))
		}
	}

	return sb.String()
}
