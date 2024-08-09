package cmd

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"

	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"repo1.dso.mil/big-bang/product/packages/bbctl/static"
	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"

	helmV2Beta1 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizeV1Beta1 "github.com/fluxcd/kustomize-controller/api/v1beta1"
	sourceV1Beta1 "github.com/fluxcd/source-controller/api/v1beta1"
)

func TestGetStatusUsage(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()

	cmd := NewStatusCmd(factory)

	assert.Equal(t, cmd.Use, "status")
	assert.Contains(t, cmd.Example, "bbctl status")
}

func TestGetStatus(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()

	streams, _ := factory.GetIOStream()
	buf := streams.Out.(*bytes.Buffer)

	cmd := NewStatusCmd(factory)
	result := cmd.RunE(cmd, []string{})

	output := strings.Split(buf.String(), "\n")

	// functionality is tested separately.
	// only checking for no error to get code coverage for cobra cmd
	assert.Nil(t, result)
	assert.NotNil(t, output)
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
	factory := bbTestUtil.GetFakeFactory()
	constants, err := static.GetDefaultConstants()
	assert.NoError(t, err)
	helmClient, err := factory.GetHelmClient(nil, constants.BigBangNamespace)
	assert.NoError(t, err)
	var response = getBigBangStatus(helmClient)
	assert.Contains(t, response, "No Big Bang release was found")

	// prepare the helm client with big bang release
	factory = bbTestUtil.GetFakeFactory()
	factory.SetHelmReleases(releaseFixture)
	helmClient, _ = factory.GetHelmClient(nil, constants.BigBangNamespace)
	response = getBigBangStatus(helmClient)
	assert.Contains(t, response, "Found bigbang release version")
}

func TestGetFluxKustomizations(t *testing.T) {
	// prepare the client
	// get runtime controller client
	// register the GitOps Toolkit schema definitions
	scheme := runtime.NewScheme()
	_ = kustomizeV1Beta1.AddToScheme(scheme)

	factory := bbTestUtil.GetFakeFactory()
	fluxClient, _ := factory.GetRuntimeClient(scheme)

	// prepare mock data for flux kustomization
	readyK := kustomizeV1Beta1.Kustomization{
		ObjectMeta: metaV1.ObjectMeta{
			Name: "readyK",
		},
		Status: kustomizeV1Beta1.KustomizationStatus{
			Conditions: []metaV1.Condition{
				{
					Type:   "Ready",
					Status: "True",
				},
			},
		},
	}

	notReadyK := kustomizeV1Beta1.Kustomization{
		ObjectMeta: metaV1.ObjectMeta{
			Name: "notReadyK",
		},
		Status: kustomizeV1Beta1.KustomizationStatus{
			Conditions: []metaV1.Condition{
				{
					Type:   "Ready",
					Status: "False",
				},
			},
		},
	}

	// test with no flux git repositories
	var response = getFluxKustomizations(fluxClient)
	assert.Contains(t, response, "No Flux kustomizations were found")

	// test with reconciled flux git repositories
	err := fluxClient.Create(context.TODO(), &readyK)
	assert.NoError(t, err)
	response = getFluxKustomizations(fluxClient)
	assert.Contains(t, response, "All Flux kustomizations are ready")

	// test with unreconciled flux git repositories
	err = fluxClient.Create(context.TODO(), &notReadyK)
	assert.NoError(t, err)
	response = getFluxKustomizations(fluxClient)
	assert.Contains(t, response, "There are 1 Flux kustomizations that are not ready")
}

func TestGetFluxGitRepositories(t *testing.T) {
	// prepare the client
	// get runtime controller client
	// register the GitOps Toolkit schema definitions
	scheme := runtime.NewScheme()
	_ = sourceV1Beta1.AddToScheme(scheme)

	factory := bbTestUtil.GetFakeFactory()
	fluxClient, _ := factory.GetRuntimeClient(scheme)

	// prepare mock data for flux git repository
	readyGR := sourceV1Beta1.GitRepository{
		ObjectMeta: metaV1.ObjectMeta{
			Name: "readyGR",
		},
		Status: sourceV1Beta1.GitRepositoryStatus{
			Conditions: []metaV1.Condition{
				{
					Type:   "Ready",
					Status: "True",
				},
			},
		},
	}

	notReadyGR := sourceV1Beta1.GitRepository{
		ObjectMeta: metaV1.ObjectMeta{
			Name: "notReadyGR",
		},
		Status: sourceV1Beta1.GitRepositoryStatus{
			Conditions: []metaV1.Condition{
				{
					Type:   "Ready",
					Status: "False",
				},
			},
		},
	}

	// test with no flux git repositories
	var response = getFluxGitRepositories(fluxClient)
	assert.Contains(t, response, "No Flux git repositories were found")

	// test with reconciled flux git repositories
	err := fluxClient.Create(context.TODO(), &readyGR)
	assert.NoError(t, err)
	response = getFluxGitRepositories(fluxClient)
	assert.Contains(t, response, "All Flux git repositories are ready")

	// test with unreconciled flux git repositories
	err = fluxClient.Create(context.TODO(), &notReadyGR)
	assert.NoError(t, err)
	response = getFluxGitRepositories(fluxClient)
	assert.Contains(t, response, "There are 1 Flux git repositories that are not ready")
}

func TestGetFluxHelmReleases(t *testing.T) {
	// prepare the client
	// get runtime controller client
	// register the GitOps Toolkit schema definitions
	scheme := runtime.NewScheme()
	_ = helmV2Beta1.AddToScheme(scheme)

	factory := bbTestUtil.GetFakeFactory()
	fluxClient, _ := factory.GetRuntimeClient(scheme)

	// prepare mock data for flux helm release
	reconciledHR := helmV2Beta1.HelmRelease{
		ObjectMeta: metaV1.ObjectMeta{
			Name: "reconciledHR",
		},
		Status: helmV2Beta1.HelmReleaseStatus{
			Conditions: []metaV1.Condition{
				{
					Type:   "Ready",
					Status: "True",
				},
			},
		},
	}

	unreconciledHR := helmV2Beta1.HelmRelease{
		ObjectMeta: metaV1.ObjectMeta{
			Name: "unreconciledHR",
		},
		Status: helmV2Beta1.HelmReleaseStatus{
			Conditions: []metaV1.Condition{
				{
					Type:   "Ready",
					Status: "False",
				},
			},
		},
	}

	// test with no flux helm releases
	var response = getFluxHelmReleases(fluxClient)
	assert.Contains(t, response, "No Flux helm releases were found")

	// test with reconciled flux helm release
	err := fluxClient.Create(context.TODO(), &reconciledHR)
	assert.NoError(t, err)
	response = getFluxHelmReleases(fluxClient)
	assert.Contains(t, response, "All Flux helm releases are reconciled")

	// test with unreconciled flux helm release
	err = fluxClient.Create(context.TODO(), &unreconciledHR)
	assert.NoError(t, err)
	response = getFluxHelmReleases(fluxClient)
	assert.Contains(t, response, "There are 1 Flux helm releases that are not reconciled")
}

func TestGetDaemonSetStatus(t *testing.T) {
	// prepare the client
	factory := bbTestUtil.GetFakeFactory()
	clientSet, _ := factory.GetClientSet()

	// prepare mock data for k8s clientset
	availDaemonSet := &appsV1.DaemonSet{
		ObjectMeta: metaV1.ObjectMeta{
			Name: "availDaemonSet",
		},
		Status: appsV1.DaemonSetStatus{
			DesiredNumberScheduled: 1,
			NumberAvailable:        1,
		},
	}

	notAvailDaemonSet := &appsV1.DaemonSet{
		ObjectMeta: metaV1.ObjectMeta{
			Name: "notAvailDaemonSet",
		},
		Status: appsV1.DaemonSetStatus{
			DesiredNumberScheduled: 1,
			NumberAvailable:        0,
		},
	}

	// test with no daemonset data
	var response = getDaemonSetsStatus(clientSet)
	assert.Contains(t, response, "No Daemonsets were found")

	// test with available daemonset
	_, err1 := clientSet.AppsV1().DaemonSets("test-ns").Create(context.TODO(), availDaemonSet, metaV1.CreateOptions{})
	if err1 != nil {
		t.Errorf("error injecting daemonset add: %v", err1)
	}
	response = getDaemonSetsStatus(clientSet)
	assert.Contains(t, response, "All Daemonsets are available")

	// test with not available daemonset
	_, err1 = clientSet.AppsV1().DaemonSets("test-ns").Create(context.TODO(), notAvailDaemonSet, metaV1.CreateOptions{})
	if err1 != nil {
		t.Errorf("error injecting daemonset add: %v", err1)
	}
	response = getDaemonSetsStatus(clientSet)
	assert.Contains(t, response, "There are 1 DaemonSets that are not available")
}

func TestGetDeploymentStatus(t *testing.T) {
	// prepare the client
	factory := bbTestUtil.GetFakeFactory()
	clientSet, _ := factory.GetClientSet()

	// prepare mock data for k8s clientset
	readyD := &appsV1.Deployment{
		ObjectMeta: metaV1.ObjectMeta{
			Name: "readyD",
		},
		Status: appsV1.DeploymentStatus{
			Replicas:      1,
			ReadyReplicas: 1,
		},
	}

	notReadyD := &appsV1.Deployment{
		ObjectMeta: metaV1.ObjectMeta{
			Name: "notReadyD",
		},
		Status: appsV1.DeploymentStatus{
			Replicas:      1,
			ReadyReplicas: 0,
		},
	}

	// test with no deployment data
	var response = getDeploymentStatus(clientSet)
	assert.Contains(t, response, "No Deployments were found")

	// test with ready deployment
	_, err1 := clientSet.AppsV1().Deployments("test-ns").Create(context.TODO(), readyD, metaV1.CreateOptions{})
	if err1 != nil {
		t.Errorf("error injecting deployment add: %v", err1)
	}
	response = getDeploymentStatus(clientSet)
	assert.Contains(t, response, "All Deployments are ready")

	// test with not ready deployment
	_, err2 := clientSet.AppsV1().Deployments("test-ns").Create(context.TODO(), notReadyD, metaV1.CreateOptions{})
	if err2 != nil {
		t.Errorf("error injecting Deployment add: %v", err1)
	}
	response = getDeploymentStatus(clientSet)
	assert.Contains(t, response, "There are 1 k8s Deployments that are not ready")
}

func TestGetStatefulSetStatus(t *testing.T) {
	// prepare the client
	factory := bbTestUtil.GetFakeFactory()
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
	var response = getStatefulSetStatus(clientSet)
	assert.Contains(t, response, "No StatefulSets were found")

	// test with ready statefulset
	_, err1 := clientSet.AppsV1().StatefulSets("test-ns").Create(context.TODO(), readySts, metaV1.CreateOptions{})
	if err1 != nil {
		t.Errorf("error injecting statefulset add: %v", err1)
	}
	response = getStatefulSetStatus(clientSet)
	assert.Contains(t, response, "All StatefulSets are ready")

	// test with not ready statefulset
	_, err2 := clientSet.AppsV1().StatefulSets("test-ns").Create(context.TODO(), notReadySts, metaV1.CreateOptions{})
	if err2 != nil {
		t.Errorf("error injecting statefulset add: %v", err1)
	}
	response = getStatefulSetStatus(clientSet)
	assert.Contains(t, response, "There are 1 StatefulSets that are not ready")
}

func TestGetPodStatus(t *testing.T) {
	// prepare the client
	factory := bbTestUtil.GetFakeFactory()
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

func TestProcessPodStatus(t *testing.T) {
	// Arrange
	readyPod := podData{
		namespace: "bigbang",
		name:      "readyPod",
		status:    "Ready",
	}
	notReadyPod := podData{
		namespace: "bigbang",
		name:      "notReadyPod",
		status:    "NotReady",
	}
	errorPod := podData{
		namespace: "bigbang",
		name:      "errorPod",
		status:    "",
	}

	var podData = []podData{}

	// Act
	processPodStatus(&readyPod, &podData, true)
	processPodStatus(&notReadyPod, &podData, false)
	processPodStatus(&errorPod, &podData, false)

	// Assert
	assert.NotContains(t, podData, readyPod)

	assert.Contains(t, podData, notReadyPod)

	assert.Equal(t, errorPod.status, "error")
	assert.Contains(t, podData, errorPod)
}

func TestGetContainerStatus(t *testing.T) {
	// Arrange
	var tests = []struct {
		desc              string
		pod               podData
		podReady          bool
		isInit            bool
		containerStatuses []coreV1.ContainerStatus
	}{
		{
			desc: "ReadyContainer",
			pod: podData{
				namespace: "bigbang",
				name:      "readyPod",
				status:    "Ready",
			},
			podReady: true,
			isInit:   false,
			containerStatuses: []coreV1.ContainerStatus{
				{
					Ready: true,
				},
				{
					Ready: true,
				},
			},
		},
		{
			desc: "UnreadyContainerWithNoInfo",
			pod: podData{
				namespace: "bigbang",
				name:      "noInfoPod",
				status:    "NotReady",
			},
			podReady: false,
			isInit:   false,
			containerStatuses: []coreV1.ContainerStatus{
				{
					Ready: false,
				},
			},
		},
		{
			desc: "MultipleUnreadyContainers",
			pod: podData{
				namespace: "bigbang",
				name:      "multipleErrorsPod",
				status:    "ImagePullBackOff",
			},
			podReady: false,
			isInit:   true,
			containerStatuses: []coreV1.ContainerStatus{
				{
					State: coreV1.ContainerState{
						Waiting: &coreV1.ContainerStateWaiting{
							Reason: "ImagePullBackOff",
						},
					},
					Ready: false,
				},
				{
					State: coreV1.ContainerState{
						Waiting: &coreV1.ContainerStateWaiting{
							Reason: "CrashLoopBackOff",
						},
					},
					Ready: false,
				},
			},
		},
		{
			desc: "MixedReadinessContainers",
			pod: podData{
				namespace: "bigbang",
				name:      "mixedResutsPod",
				status:    "CrashLoopBackOff",
			},
			podReady: false,
			isInit:   true,
			containerStatuses: []coreV1.ContainerStatus{
				{
					Ready: true,
				},
				{
					State: coreV1.ContainerState{
						Waiting: &coreV1.ContainerStateWaiting{
							Reason: "CrashLoopBackOff",
						},
					},
					Ready: false,
				},
			},
		},
	}

	var podReadyPointer bool
	prefix := ""
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			if test.isInit {
				prefix = "init:"
			} else {
				prefix = ""
			}

			// Act
			podReadyPointer = true
			getContainerStatus(test.containerStatuses, &test.pod, &podReadyPointer, test.isInit)

			// Assert
			assert.Equal(t, podReadyPointer, test.podReady)
			for i := len(test.containerStatuses) - 1; i >= 0; i-- {
				if test.containerStatuses[i].State.Waiting != nil {
					// Multiple containers overwrite pod status so only the last status gets captured
					assert.Equal(t, test.pod.status, prefix+test.containerStatuses[i].State.Waiting.Reason)
					break
				}
			}
		})
	}
}
