package cmd

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	bbtestutil "repo1.dso.mil/platform-one/big-bang/apps/product-tools/bbctl/util/test"

	helmv2beta1 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev1beta1 "github.com/fluxcd/kustomize-controller/api/v1beta1"
	sourcev1beta1 "github.com/fluxcd/source-controller/api/v1beta1"
)

func TestGetStatus(t *testing.T) {

	factory := bbtestutil.GetFakeFactory(nil, nil, nil)

	streams, _, buf, _ := genericclioptions.NewTestIOStreams()

	cmd := NewStatusCmd(factory, streams)
	cmd.Run(cmd, []string{})

	response := strings.Split(buf.String(), "\n")

	// fuctionality is tested separately.
	// only checking for not nil to get code coverage for cobra cmd
	assert.NotNil(t, response)
}

func TestGetBigBangStatus(t *testing.T) {

	// prepare mock data for helm
	chartBB := &chart.Chart{
		Metadata: &chart.Metadata{
			Name:    "bigbang",
			Version: "1.0.2",
		},
	}

	releaseFixture := []*release.Release{
		{
			Name:      "bigbang",
			Version:   1,
			Namespace: "bigbang",
			Info: &release.Info{
				Status: release.StatusDeployed,
			},
			Chart: chartBB,
		},
	}

	// prepare the helm client with no data
	factory := bbtestutil.GetFakeFactory(nil, nil, nil)
	helmclient, _ := factory.GetHelmClient(BigBangNamespace)
	var response = getBigBangStatus(helmclient)
	assert.Contains(t, response, "No BigBang release was found")

	// prepare the helm client with bigbang release
	factory = bbtestutil.GetFakeFactory(releaseFixture, nil, nil)
	helmclient, _ = factory.GetHelmClient(BigBangNamespace)
	response = getBigBangStatus(helmclient)
	assert.Contains(t, response, "Found bigbang release version")
}

func TestGetFluxKustomizations(t *testing.T) {

	// prepare the client
	// get runtime controller client
	// register the GitOps Toolkit schema definitions
	scheme := runtime.NewScheme()
	_ = kustomizev1beta1.AddToScheme(scheme)

	factory := bbtestutil.GetFakeFactory(nil, nil, nil)
	fluxClient, _ := factory.GetRuntimeClient(scheme)

	// prepare mock data for flux kustomization
	readyK := kustomizev1beta1.Kustomization{
		ObjectMeta: metaV1.ObjectMeta{
			Name: "readyK",
		},
		Status: kustomizev1beta1.KustomizationStatus{
			Conditions: []metaV1.Condition{
				{
					Type:   "Ready",
					Status: "True",
				},
			},
		},
	}

	notReadyK := kustomizev1beta1.Kustomization{
		ObjectMeta: metaV1.ObjectMeta{
			Name: "notReadyK",
		},
		Status: kustomizev1beta1.KustomizationStatus{
			Conditions: []metaV1.Condition{
				{
					Type:   "Ready",
					Status: "False",
				},
			},
		},
	}

	// test with no flux gitrepositories
	var response = getFluxKustomizations(fluxClient)
	assert.Contains(t, response, "No Flux kustomizations were found")

	// test with reconciled flux gitrepositories
	fluxClient.Create(context.TODO(), &readyK)
	response = getFluxKustomizations(fluxClient)
	assert.Contains(t, response, "All Flux kustomizations are ready")

	// test with unreconciled flux gitrepositories
	fluxClient.Create(context.TODO(), &notReadyK)
	response = getFluxKustomizations(fluxClient)
	assert.Contains(t, response, "There are 1 Flux kustomizations that are not ready")
}

func TestGetFluxGitRepositories(t *testing.T) {

	// prepare the client
	// get runtime controller client
	// register the GitOps Toolkit schema definitions
	scheme := runtime.NewScheme()
	_ = sourcev1beta1.AddToScheme(scheme)

	factory := bbtestutil.GetFakeFactory(nil, nil, nil)
	fluxClient, _ := factory.GetRuntimeClient(scheme)

	// prepare mock data for flux gitrepository
	readyGR := sourcev1beta1.GitRepository{
		ObjectMeta: metaV1.ObjectMeta{
			Name: "readyGR",
		},
		Status: sourcev1beta1.GitRepositoryStatus{
			Conditions: []metaV1.Condition{
				{
					Type:   "Ready",
					Status: "True",
				},
			},
		},
	}

	notReadyGR := sourcev1beta1.GitRepository{
		ObjectMeta: metaV1.ObjectMeta{
			Name: "notReadyGR",
		},
		Status: sourcev1beta1.GitRepositoryStatus{
			Conditions: []metaV1.Condition{
				{
					Type:   "Ready",
					Status: "False",
				},
			},
		},
	}

	// test with no flux gitrepositories
	var response = getFluxGitRepositories(fluxClient)
	assert.Contains(t, response, "No Flux gitrepositories were found")

	// test with reconciled flux gitrepositories
	fluxClient.Create(context.TODO(), &readyGR)
	response = getFluxGitRepositories(fluxClient)
	assert.Contains(t, response, "All Flux gitrepositories are ready")

	// test with unreconciled flux gitrepositories
	fluxClient.Create(context.TODO(), &notReadyGR)
	response = getFluxGitRepositories(fluxClient)
	assert.Contains(t, response, "There are 1 Flux gitrepositories that are not ready")
}

func TestGetFluxHelmReleases(t *testing.T) {

	// prepare the client
	// get runtime controller client
	// register the GitOps Toolkit schema definitions
	scheme := runtime.NewScheme()
	_ = helmv2beta1.AddToScheme(scheme)

	factory := bbtestutil.GetFakeFactory(nil, nil, nil)
	fluxClient, _ := factory.GetRuntimeClient(scheme)

	// prepare mock data for flux helmrelease
	reconciledHR := helmv2beta1.HelmRelease{
		ObjectMeta: metaV1.ObjectMeta{
			Name: "reconciledHR",
		},
		Status: helmv2beta1.HelmReleaseStatus{
			Conditions: []metaV1.Condition{
				{
					Type:   "Ready",
					Status: "True",
				},
			},
		},
	}

	unreconciledHR := helmv2beta1.HelmRelease{
		ObjectMeta: metaV1.ObjectMeta{
			Name: "unreconciledHR",
		},
		Status: helmv2beta1.HelmReleaseStatus{
			Conditions: []metaV1.Condition{
				{
					Type:   "Ready",
					Status: "False",
				},
			},
		},
	}

	// test with no flux helmreleases
	var response = getFluxHelmReleases(fluxClient)
	assert.Contains(t, response, "No Flux helmreleases were found")

	// test with reconciled flux helmrelease
	fluxClient.Create(context.TODO(), &reconciledHR)
	response = getFluxHelmReleases(fluxClient)
	assert.Contains(t, response, "All Flux helmreleases are reconciled")

	// test with unreconciled flux helmrelease
	fluxClient.Create(context.TODO(), &unreconciledHR)
	response = getFluxHelmReleases(fluxClient)
	assert.Contains(t, response, "There are 1 Flux helmreleases that are not reconciled")
}

func TestGetDmstStatus(t *testing.T) {

	// prepare the client
	factory := bbtestutil.GetFakeFactory(nil, nil, nil)
	clientSet, _ := factory.GetClientSet()

	// prepare mock data for k8s clientset
	availDmst := &appsV1.DaemonSet{
		ObjectMeta: metaV1.ObjectMeta{
			Name: "availDmst",
		},
		Status: appsV1.DaemonSetStatus{
			DesiredNumberScheduled: 1,
			NumberAvailable:        1,
		},
	}

	notAvailDmst := &appsV1.DaemonSet{
		ObjectMeta: metaV1.ObjectMeta{
			Name: "notAvailDmst",
		},
		Status: appsV1.DaemonSetStatus{
			DesiredNumberScheduled: 1,
			NumberAvailable:        0,
		},
	}

	// test with no daemonset data
	var response = getDmstStatus(clientSet)
	assert.Contains(t, response, "No Daemonsets were found")

	// test with available daemonset
	_, err1 := clientSet.AppsV1().DaemonSets("test-ns").Create(context.TODO(), availDmst, metaV1.CreateOptions{})
	if err1 != nil {
		t.Errorf("error injecting daemonset add: %v", err1)
	}
	response = getDmstStatus(clientSet)
	assert.Contains(t, response, "All Daemonsets are available")

	// test with not available daemonset
	_, err1 = clientSet.AppsV1().DaemonSets("test-ns").Create(context.TODO(), notAvailDmst, metaV1.CreateOptions{})
	if err1 != nil {
		t.Errorf("error injecting daemonset add: %v", err1)
	}
	response = getDmstStatus(clientSet)
	assert.Contains(t, response, "There are 1 DaemonSets that are not available")
}

func TestGetDpmtStatus(t *testing.T) {

	// prepare the client
	factory := bbtestutil.GetFakeFactory(nil, nil, nil)
	clientSet, _ := factory.GetClientSet()

	// prepare mock data for k8s clientset
	readyDpmt := &appsV1.Deployment{
		ObjectMeta: metaV1.ObjectMeta{
			Name: "readyDpmt",
		},
		Status: appsV1.DeploymentStatus{
			Replicas:      1,
			ReadyReplicas: 1,
		},
	}

	notReadyDpmt := &appsV1.Deployment{
		ObjectMeta: metaV1.ObjectMeta{
			Name: "notReadyDpmt",
		},
		Status: appsV1.DeploymentStatus{
			Replicas:      1,
			ReadyReplicas: 0,
		},
	}

	// test with no deployment data
	var response = getDpmtStatus(clientSet)
	assert.Contains(t, response, "No Deployments were found")

	// test with ready deployment
	_, err1 := clientSet.AppsV1().Deployments("test-ns").Create(context.TODO(), readyDpmt, metaV1.CreateOptions{})
	if err1 != nil {
		t.Errorf("error injecting deployment add: %v", err1)
	}
	response = getDpmtStatus(clientSet)
	assert.Contains(t, response, "All Deployments are ready")

	// test with not ready deployment
	_, err2 := clientSet.AppsV1().Deployments("test-ns").Create(context.TODO(), notReadyDpmt, metaV1.CreateOptions{})
	if err2 != nil {
		t.Errorf("error injecting Deployment add: %v", err1)
	}
	response = getDpmtStatus(clientSet)
	assert.Contains(t, response, "There are 1 k8s Deployments that are not ready")
}

func TestGetStsStatus(t *testing.T) {

	// prepare the client
	factory := bbtestutil.GetFakeFactory(nil, nil, nil)
	clientSet, _ := factory.GetClientSet()

	// prepare mock data for k8s clientset
	readySts := &appsV1.StatefulSet{
		ObjectMeta: metaV1.ObjectMeta{
			Name: "readySts",
		},
		Status: appsV1.StatefulSetStatus{
			Replicas:      1,
			ReadyReplicas: 1,
		},
	}

	notReadySts := &appsV1.StatefulSet{
		ObjectMeta: metaV1.ObjectMeta{
			Name: "notReadySts",
		},
		Status: appsV1.StatefulSetStatus{
			Replicas:      1,
			ReadyReplicas: 0,
		},
	}

	// test with no statefulset data
	var response = getStsStatus(clientSet)
	assert.Contains(t, response, "No StatefulSets were found")

	// test with ready statefulset
	_, err1 := clientSet.AppsV1().StatefulSets("test-ns").Create(context.TODO(), readySts, metaV1.CreateOptions{})
	if err1 != nil {
		t.Errorf("error injecting statefulset add: %v", err1)
	}
	response = getStsStatus(clientSet)
	assert.Contains(t, response, "All StatefulSets are ready")

	// test with not ready statufulset
	_, err2 := clientSet.AppsV1().StatefulSets("test-ns").Create(context.TODO(), notReadySts, metaV1.CreateOptions{})
	if err2 != nil {
		t.Errorf("error injecting statefulset add: %v", err1)
	}
	response = getStsStatus(clientSet)
	assert.Contains(t, response, "There are 1 StatefulSets that are not ready")
}

func TestGetPodStatus(t *testing.T) {

	// prepare the client
	factory := bbtestutil.GetFakeFactory(nil, nil, nil)
	clientSet, _ := factory.GetClientSet()

	// prepare mock data for k8s clientset
	readyPod := &coreV1.Pod{
		ObjectMeta: metaV1.ObjectMeta{
			Name: "ready-pod",
		},
		Status: coreV1.PodStatus{
			Phase: "Running",
			ContainerStatuses: []coreV1.ContainerStatus{
				{
					State: coreV1.ContainerState{
						Waiting: &coreV1.ContainerStateWaiting{},
					},
					Ready: true,
				},
			},
		},
	}

	runningBadPod := &coreV1.Pod{
		ObjectMeta: metaV1.ObjectMeta{
			Name: "running-bad-pod",
		},
		Status: coreV1.PodStatus{
			Phase: "Running",
			ContainerStatuses: []coreV1.ContainerStatus{
				{
					State: coreV1.ContainerState{
						Waiting: &coreV1.ContainerStateWaiting{
							Reason: "CrashLoopBackoff",
						},
					},
					Ready: false,
				},
			},
		},
	}

	pendingBadPod := &coreV1.Pod{
		ObjectMeta: metaV1.ObjectMeta{
			Name: "pending-bad-pod",
		},
		Status: coreV1.PodStatus{
			Phase: "Pending",
			InitContainerStatuses: []coreV1.ContainerStatus{
				{
					State: coreV1.ContainerState{
						Waiting: &coreV1.ContainerStateWaiting{
							Reason: "CrashLoopBackoff",
						},
					},
					Ready: false,
				},
			},
		},
	}

	// add first test for all healthy pods
	_, err1 := clientSet.CoreV1().Pods("test-ns").Create(context.TODO(), readyPod, metaV1.CreateOptions{})
	if err1 != nil {
		t.Errorf("error injecting pod add: %v", err1)
	}
	var response = getPodStatus(clientSet)
	assert.Contains(t, response, "All pods are ready")

	// add test with unhealthy pods
	_, err2 := clientSet.CoreV1().Pods("test-ns").Create(context.TODO(), runningBadPod, metaV1.CreateOptions{})
	if err2 != nil {
		t.Errorf("error injecting pod add: %v", err2)
	}
	_, err3 := clientSet.CoreV1().Pods("test-ns").Create(context.TODO(), pendingBadPod, metaV1.CreateOptions{})
	if err3 != nil {
		t.Errorf("error injecting pod add: %v", err3)
	}
	response = getPodStatus(clientSet)
	assert.Contains(t, response, "2 pods that are not ready")
}
