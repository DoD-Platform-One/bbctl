package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
	bbConfig "repo1.dso.mil/big-bang/product/packages/bbctl/util/config"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/config/schemas"
	output "repo1.dso.mil/big-bang/product/packages/bbctl/util/output"
	outputSchemas "repo1.dso.mil/big-bang/product/packages/bbctl/util/output/schemas"
	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"
	apiWrappers "repo1.dso.mil/big-bang/product/packages/bbctl/util/test/apiwrappers"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	coreV1 "k8s.io/api/core/v1"
	storageV1 "k8s.io/api/storage/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	fake "k8s.io/client-go/kubernetes/fake"
	fakeTypedBatchV1 "k8s.io/client-go/kubernetes/typed/batch/v1/fake"
	fakeTypedCoreV1 "k8s.io/client-go/kubernetes/typed/core/v1/fake"
	fakeTyped "k8s.io/client-go/kubernetes/typed/storage/v1/fake"
	k8sTesting "k8s.io/client-go/testing"
)

func pod(app string, ns string, phase coreV1.PodPhase) *coreV1.Pod {
	pod := &coreV1.Pod{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      fmt.Sprintf("%s-fequ", app),
			Namespace: ns,
			Labels: map[string]string{
				"app": app,
			},
		},
		Status: coreV1.PodStatus{
			Phase: phase,
		},
	}

	return pod
}

func ns(name string, phase coreV1.NamespacePhase) *coreV1.Namespace {
	ns := &coreV1.Namespace{
		ObjectMeta: metaV1.ObjectMeta{
			Name: name,
		},
		Status: coreV1.NamespaceStatus{
			Phase: phase,
		},
	}

	return ns
}

func checkOutput(t *testing.T, expected []string, actual []string) {
	for _, value := range expected {
		if !assert.Contains(t, actual, value) {
			t.Errorf("\n\nexpected:\n%s\n\nactual:\n%s\n\n", value, strings.Join(actual, "\n"))
		}
	}
}

func TestCheckMetricsServer(t *testing.T) {
	arl := metaV1.APIResourceList{
		GroupVersion: "metrics.k8s.io/v1beta1",
		APIResources: []metaV1.APIResource{
			{
				Name: "PodMetrics",
			},
		},
	}
	badArl := metaV1.APIResourceList{
		GroupVersion: "this/is/wrong",
		APIResources: []metaV1.APIResource{
			{
				Name: "this/is/wrong/too",
			},
		},
	}

	var tests = []struct {
		desc             string
		expected         []string
		status           preflightCheckStatus
		resources        []*metaV1.APIResourceList
		failGetClient    bool
		failServerGroups bool
	}{
		{
			"Metrics Unavailable",
			[]string{
				"Checking metrics server...",
				"Check Failed - Metrics API not available",
			},
			failed,
			[]*metaV1.APIResourceList{},
			false,
			false,
		},
		{
			"metrics server available",
			[]string{
				"Checking metrics server...",
				"Check Passed - Metrics API available",
			},
			passed,
			[]*metaV1.APIResourceList{&arl},
			false,
			false,
		},
		{
			"Get K8s Client Failure",
			[]string{
				"Checking metrics server...",
				"Failed to get k8s clientset: testing error",
			},
			unknown,
			[]*metaV1.APIResourceList{},
			true,
			false,
		},
		{
			"Get Server Groups Failure",
			[]string{
				"Checking metrics server...",
				"Failed to get server groups: unexpected GroupVersion string: this/is/wrong",
			},
			unknown,
			[]*metaV1.APIResourceList{&badArl},
			false,
			true,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			factory := bbTestUtil.GetFakeFactory()
			factory.SetResources(test.resources)
			factory.SetFail.GetK8sClientset = test.failGetClient

			msgs, status := checkMetricsServer(nil, factory, nil)
			checkOutput(t, msgs, test.expected)
			assert.Equal(t, status, test.status)
		})
	}
}

func TestCheckDefaultStorageClass(t *testing.T) {
	barSC := &storageV1.StorageClass{
		ObjectMeta: metaV1.ObjectMeta{
			Name: "bar",
		},
	}

	fooSC := &storageV1.StorageClass{
		ObjectMeta: metaV1.ObjectMeta{
			Name: "foo",
			Annotations: map[string]string{
				"storageclass.kubernetes.io/is-default-class": "true",
			},
		},
	}

	var tests = []struct {
		desc                 string
		expected             []string
		status               preflightCheckStatus
		objects              []runtime.Object
		failGetClientset     bool
		failListStorageClass bool
	}{
		{
			"No Storage Class",
			[]string{
				"Checking default storage class...",
				"Check Failed - Default storage class not found",
			},
			failed,
			[]runtime.Object{},
			false,
			false,
		},
		{
			"Default Storage Class",
			[]string{
				"Checking default storage class...",
				"Check Passed - Default storage class foo found",
			},
			passed,
			[]runtime.Object{fooSC},
			false,
			false,
		},
		{
			"no default storage class",
			[]string{
				"Checking default storage class...",
				"Check Failed - Default storage class not found",
			},
			failed,
			[]runtime.Object{barSC},
			false,
			false,
		},
		{
			"Failed Getting ClientSet",
			[]string{
				"Checking default storage class...",
				"Failed to get k8s clientset: testing error",
			},
			unknown,
			[]runtime.Object{},
			true,
			false,
		},
		{
			"Failed GettingStorage Class",
			[]string{
				"Checking default storage class...",
				"Failed to get storage classes: testing error",
			},
			unknown,
			[]runtime.Object{},
			false,
			true,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			// Arrange
			factory := bbTestUtil.GetFakeFactory()
			factory.ResetIOStream()
			factory.SetObjects(test.objects)
			factory.SetFail.GetK8sClientset = test.failGetClientset

			if test.failListStorageClass {
				failFunc := func(action k8sTesting.Action) (bool, runtime.Object, error) {
					return true, nil, fmt.Errorf("testing error")
				}
				modFunc := func(clientset *fake.Clientset) {
					clientset.StorageV1().StorageClasses().(*fakeTyped.FakeStorageClasses).Fake.PrependReactor("list", "storageclasses", failFunc)
				}
				factory.SetFail.GetK8sClientsetPrepFuncs = append(factory.SetFail.GetK8sClientsetPrepFuncs, &modFunc)
			}

			// Act
			msgs, status := checkDefaultStorageClass(nil, factory, nil)

			// Assert
			checkOutput(t, test.expected, msgs)
			assert.Equal(t, test.status, status)
		})
	}
}

func TestCheckFluxController(t *testing.T) {
	hcPodRunning := pod("helm-controller", "flux-system", coreV1.PodRunning)
	hcPodFailed := pod("helm-controller", "flux-system", coreV1.PodFailed)
	kcPodRunning := pod("kustomize-controller", "flux-system", coreV1.PodRunning)
	kcPodFailed := pod("kustomize-controller", "flux-system", coreV1.PodFailed)
	scPodRunning := pod("source-controller", "flux-system", coreV1.PodRunning)
	scPodFailed := pod("source-controller", "flux-system", coreV1.PodFailed)
	ncPodRunning := pod("notification-controller", "flux-system", coreV1.PodRunning)
	ncPodFailed := pod("notification-controller", "flux-system", coreV1.PodFailed)

	var tests = []struct {
		desc             string
		expected         []string
		status           preflightCheckStatus
		objects          []runtime.Object
		failGetClientset bool
		failListPods     bool
	}{
		{
			"No Helm Controller",
			[]string{
				"Checking flux installation...",
				"Check Failed - flux helm-controller pod not found in flux-system namespace",
			},
			failed,
			[]runtime.Object{},
			false,
			false,
		},
		{
			"No Kustomize Controller",
			[]string{
				"Checking flux installation...",
				"Check Failed - flux kustomize-controller pod not found in flux-system namespace",
			},
			failed,
			[]runtime.Object{hcPodFailed, scPodFailed, ncPodFailed},
			false,
			false,
		},
		{
			"Failing Kustomize Controller",
			[]string{
				"Checking flux installation...",
				"Check Failed - flux kustomize-controller pod not in running state",
			},
			failed,
			[]runtime.Object{kcPodFailed, hcPodRunning, ncPodRunning, scPodRunning},
			false,
			false,
		},
		{
			"Flux Controller Running",
			[]string{
				"Checking flux installation...",
				"Check Passed - flux kustomize-controller pod running",
			},
			passed,
			[]runtime.Object{kcPodRunning, hcPodRunning, ncPodRunning, scPodRunning},
			false,
			false,
		},
		{
			"Failed Getting ClientSet",
			[]string{
				"Checking flux installation...",
				"Failed to get k8s clientset: testing error",
			},
			unknown,
			[]runtime.Object{},
			true,
			false,
		},
		{
			"Failed Getting Pods",
			[]string{
				"Checking flux installation...",
				"Failed to get pods: testing error",
			},
			unknown,
			[]runtime.Object{},
			false,
			true,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			// Arrange
			factory := bbTestUtil.GetFakeFactory()
			factory.SetObjects(test.objects)
			factory.SetFail.GetK8sClientset = test.failGetClientset

			if test.failListPods {
				failFunc := func(action k8sTesting.Action) (bool, runtime.Object, error) {
					return true, nil, fmt.Errorf("testing error")
				}
				modFunc := func(clientset *fake.Clientset) {
					clientset.CoreV1().Pods("flux-system").(*fakeTypedCoreV1.FakePods).Fake.PrependReactor("list", "pods", failFunc)
				}
				factory.SetFail.GetK8sClientsetPrepFuncs = append(factory.SetFail.GetK8sClientsetPrepFuncs, &modFunc)
			}

			// Act
			msgs, status := checkFluxController(nil, factory, nil)

			// Assert
			checkOutput(t, test.expected, msgs)
			assert.Equal(t, test.status, status)
		})
	}
}

func TestCheckSystemParameters(t *testing.T) {
	defaultPassingParams := map[string]string{
		"cat /proc/sys/vm/max_map_count": "524288",
		"cat /proc/sys/fs/file-max":      "131072",
		"ulimit -n":                      "131072",
		"ulimit -u":                      "8192",
	}
	fullExpectedPassingOutput := []string{
		"Creating namespace for command execution...",
		"Namespace preflight-check already exists... It will be recreated",
		"Creating registry secret for command execution...",
		"Creating job for command execution...",
		"Waiting for job preflightcheck to be ready...",
		"Checking system parameters...",
		"Checking vm.max_map_count",
		"Checking fs.file-max",
		"Checking ulimit -n",
		"Checking ulimit -u",
		"Deleting namespace for command execution...",
	}
	var tests = []struct {
		desc                   string
		expected               []string
		paramOverrides         map[string]string
		status                 preflightCheckStatus
		failGetClientset       bool
		failGetCommandExecutor bool
		failDeleteNamespace    bool
	}{
		{
			"check failed for max_map_count (ECK)",
			[]string{"Check Failed - vm.max_map_count needs to be at least 262144 for ECK to work. Current value 262100\n"},
			map[string]string{"cat /proc/sys/vm/max_map_count": "262100"},
			failed,
			false,
			false,
			false,
		},
		{
			"check passed for max_map_count (ECK)",
			[]string{"Check Passed - vm.max_map_count 262144 is suitable for ECK to work.\n"},
			map[string]string{"cat /proc/sys/vm/max_map_count": "262144"},
			failed, // Sonarqube is higher than ECK, so this should fail
			false,
			false,
			false,
		},
		{
			"check failed for max_map_count (Sonarqube)",
			[]string{"Check Failed - vm.max_map_count needs to be at least 524288 for Sonarqube to work. Current value 524280\n"},
			map[string]string{"cat /proc/sys/vm/max_map_count": "524280"},
			failed,
			false,
			false,
			false,
		},
		{
			"check passed for max_map_count (Sonarqube)",
			append(fullExpectedPassingOutput, "Check Passed - vm.max_map_count 524288 is suitable for Sonarqube to work.\n"),
			map[string]string{"cat /proc/sys/vm/max_map_count": "524288"},
			passed,
			false,
			false,
			false,
		},
		{
			"check failed for file-max (Sonarqube)",
			[]string{"Check Failed - fs.file-max needs to be at least 131072 for Sonarqube to work. Current value 131070\n"},
			map[string]string{"cat /proc/sys/fs/file-max": "131070"},
			failed,
			false,
			false,
			false,
		},
		{
			"check passed for file-max (Sonarqube)",
			append(fullExpectedPassingOutput, "Check Passed - fs.file-max 131074 is suitable for Sonarqube to work.\n"),
			map[string]string{"cat /proc/sys/fs/file-max": "131074"},
			passed,
			false,
			false,
			false,
		},
		{
			"check failed for ulimit -n (Sonarqube)",
			[]string{"Check Failed - ulimit -n needs to be at least 131072 for Sonarqube to work. Current value 131070\n"},
			map[string]string{"ulimit -n": "131070"},
			failed,
			false,
			false,
			false,
		},
		{
			"check passed for ulimit -n (Sonarqube) unlimited",
			append(fullExpectedPassingOutput, "Check Passed - ulimit -n unlimited is suitable for Sonarqube to work.\n"),
			map[string]string{"ulimit -n": "unlimited"},
			passed,
			false,
			false,
			false,
		},
		{
			"check failed for ulimit -n (Sonarqube) unknown",
			[]string{"Check Undetermined - ulimit -n needs to be at least 131072 for Sonarqube to work. Current value unknown\n"},
			map[string]string{"ulimit -n": "unknown"},
			failed,
			false,
			false,
			false,
		},
		{
			"check failed for ulimit -u (Sonarqube)",
			[]string{"Check Failed - ulimit -u needs to be at least 8192 for Sonarqube to work. Current value 8190\n"},
			map[string]string{"ulimit -u": "8190"},
			failed,
			false,
			false,
			false,
		},
		{
			"check passed for ulimit -u (Sonarqube)",
			append(fullExpectedPassingOutput, "Check Passed - ulimit -u 8192 is suitable for Sonarqube to work.\n"),
			map[string]string{"ulimit -u": "8192"},
			passed,
			false,
			false,
			false,
		},
		{
			"failed to get clientset",
			[]string{"Failed to create resources for command execution: testing error"},
			map[string]string{},
			unknown,
			true,
			false,
			false,
		},
		{
			"failed to get command executor",
			[]string{"Failed to get command executor: testing error"},
			map[string]string{},
			unknown,
			false,
			true,
			false,
		},
		{
			"failed to delete namespace",
			[]string{
				"Creating namespace for command execution...",
				"Creating registry secret for command execution...",
				"Creating job for command execution...",
				"Waiting for job preflightcheck to be ready...",
				"Checking system parameters...",
				"Checking vm.max_map_count",
				"vm.max_map_count = 524288",
				"Check Passed - vm.max_map_count 524288 is suitable for ECK to work.\n",
				"vm.max_map_count = 524288",
				"Check Passed - vm.max_map_count 524288 is suitable for Sonarqube to work.\n",
				"Checking fs.file-max",
				"fs.file-max = 131072",
				"Check Passed - fs.file-max 131072 is suitable for Sonarqube to work.\n",
				"Checking ulimit -n",
				"ulimit -n = 131072",
				"Check Passed - ulimit -n 131072 is suitable for Sonarqube to work.\n",
				"Checking ulimit -u",
				"ulimit -u = 8192",
				"Check Passed - ulimit -u 8192 is suitable for Sonarqube to work.\n",
				"Deleting namespace for command execution...",
				"Error occurred when deleting system parameter check resources: namespaces \"preflight-check\" not found",
			},
			map[string]string{},
			unknown,
			false,
			false,
			true,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {

			// Arrange
			pfcNs := ns("preflight-check", coreV1.NamespaceActive)
			pfcPod := pod("pfc", pfcNs.Name, coreV1.PodRunning)
			pfcPod.ObjectMeta.Labels["job-name"] = "preflightcheck"

			command := &cobra.Command{}
			factory := bbTestUtil.GetFakeFactory()
			factory.ResetIOStream()
			configClient, err := factory.GetConfigClient(command)
			assert.Nil(t, err)
			viperInstance, _ := factory.GetViper()
			assert.Nil(t, configClient.SetAndBindFlag("big-bang-repo", "", "/tmp", "Location on the filesystem where the bigbang product repo is checked out"))
			assert.Nil(t, configClient.SetAndBindFlag("registryserver", "", "registry.foo", "Image registry server url"))
			assert.Nil(t, configClient.SetAndBindFlag("registryusername", "", "user", "Image registry username"))
			assert.Nil(t, configClient.SetAndBindFlag("registrypassword", "", "pass", "Image registry password"))
			assert.Nil(t, viperInstance.BindPFlags(command.Flags()))

			config, configErr := configClient.GetConfig()
			assert.NoError(t, configErr)
			config.OutputConfiguration.Format = "text"
			config.PreflightCheckConfiguration.RegistryServer = "registry.foo"
			config.PreflightCheckConfiguration.RegistryUsername = "user"
			config.PreflightCheckConfiguration.RegistryPassword = "pass"
			executor, err := factory.GetFakeCommandExecutor()
			assert.Nil(t, err)
			if test.failDeleteNamespace {
				factory.SetObjects([]runtime.Object{pfcPod})
			} else {
				factory.SetObjects([]runtime.Object{pfcPod, pfcNs})
			}
			factory.SetFail.GetK8sClientset = test.failGetClientset
			factory.SetFail.GetCommandExecutor = test.failGetCommandExecutor
			modifiedParams := make(map[string]string)

		CHECK_DEFAULTS:
			for dk, dv := range defaultPassingParams {
				for ok, ov := range test.paramOverrides {
					if dk == ok {
						modifiedParams[dk] = ov
						continue CHECK_DEFAULTS
					}
				}
				modifiedParams[dk] = dv
			}
			executor.CommandResult = modifiedParams

			// Act
			msgs, status := checkSystemParameters(command, factory, config)

			// Assert
			checkOutput(t, test.expected, msgs)
			assert.Equal(t, test.status, status)
		})
	}
}

// all of checkSystemParameter is tested in the previous test

func TestSupportedMetricsAPIVersionAvailable(t *testing.T) {
	tests := []struct {
		desc     string
		expected bool
		names    []string
		versions []string
	}{
		{
			desc:     "no api groups",
			expected: false,
			names:    []string{},
			versions: []string{},
		},
		{
			desc:     "no metrics api group",
			expected: false,
			names:    []string{"foo", "bar"},
			versions: []string{"v1", "v2"},
		},
		{
			desc:     "no metrics api versions",
			expected: false,
			names:    []string{"metrics.k8s.io"},
			versions: []string{},
		},
		{
			desc:     "metrics api version available",
			expected: true,
			names:    []string{"metrics.k8s.io"},
			versions: []string{"v1beta1"},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			// Arrange
			apiGroupList := &metaV1.APIGroupList{
				Groups: []metaV1.APIGroup{},
			}
			for _, name := range test.names {
				groupVersions := []metaV1.GroupVersionForDiscovery{}
				for _, version := range test.versions {
					groupVersions = append(groupVersions, metaV1.GroupVersionForDiscovery{Version: version})
				}
				apiGroupList.Groups = append(apiGroupList.Groups, metaV1.APIGroup{
					Name:     name,
					Versions: groupVersions,
				})
			}

			// Act
			available := supportedMetricsAPIVersionAvailable(apiGroupList)

			// Assert
			assert.Equal(t, test.expected, available)
		})
	}
}

func TestCreateResourcesForCommandExecution(t *testing.T) {
	podToFind := &coreV1.Pod{
		ObjectMeta: metaV1.ObjectMeta{
			Name: "preflightcheck",
			Labels: map[string]string{
				"job-name": "preflightcheck",
			},
		},
		Status: coreV1.PodStatus{
			Phase: coreV1.PodRunning,
		},
	}
	tests := []struct {
		desc                 string
		expectedOut          []string
		failGetClientset     bool
		failCreateNamespace  bool
		failCreateSecret     bool
		failGetPod           bool
		failTimeoutPod       bool
		failDeleteNamespace  bool
		failTimeoutNamespace bool
		failCreatePod        bool
	}{
		{
			"success",
			[]string{
				"Creating namespace for command execution...",
				"Creating registry secret for command execution...",
				"Creating job for command execution...",
				"Waiting for job preflightcheck to be ready...",
			},
			false,
			false,
			false,
			false,
			false,
			false,
			false,
			false,
		},
		{
			"failed to get clientset",
			[]string{
				"testing error",
			},
			true,
			false,
			false,
			false,
			false,
			false,
			false,
			false,
		},
		{
			"failed to create namespace",
			[]string{
				"Creating namespace for command execution...",
				"failed to create namespace",
			},
			false,
			true,
			false,
			false,
			false,
			false,
			false,
			false,
		},
		{
			"failed to create secret",
			[]string{
				"Creating namespace for command execution...",
				"Creating registry secret for command execution...",
				"***Invalid registry credentials provided. Ensure the registry server, username, and password values are all set!***",
			},
			false,
			false,
			true,
			false,
			false,
			false,
			false,
			false,
		},
		{
			"failed to get pod",
			[]string{
				"Creating namespace for command execution...",
				"Creating registry secret for command execution...",
				"Creating job for command execution...",
				"Waiting for job preflightcheck to be ready...",
				"Failed to fetch preflightcheck pod status: failed to get pod",
			},
			false,
			false,
			false,
			true,
			false,
			false,
			false,
			false,
		},
		{
			"failed to timeout pod",
			[]string{
				"Creating namespace for command execution...",
				"Creating registry secret for command execution...",
				"Creating job for command execution...",
				"Waiting for job preflightcheck to be ready...",
				"Timeout waiting for command execution job to be ready",
			},
			false,
			false,
			false,
			false,
			true,
			false,
			false,
			false,
		},
		{
			"failed to delete namespace",
			[]string{
				"Creating namespace for command execution...",
				"Namespace preflight-check already exists... It will be recreated",
				"namespaces \"preflight-check\" not found",
			},
			false,
			false,
			false,
			false,
			false,
			true,
			false,
			false,
		},
		{
			"failed to timeout namespace",
			[]string{
				"Creating namespace for command execution...",
				"Namespace preflight-check already exists... It will be recreated",
				"namespaces \"preflight-check\" already exists",
			},
			false,
			false,
			false,
			false,
			false,
			false,
			true,
			false,
		},
		{
			"failed to create pod",
			[]string{
				"Creating namespace for command execution...",
				"Creating registry secret for command execution...",
				"Creating job for command execution...",
				"Failed to create preflightcheck job: failed to create pod",
			},
			false,
			false,
			false,
			false,
			false,
			false,
			false,
			true,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			// Arrange
			startTime := time.Now()
			command := &cobra.Command{}
			factory := bbTestUtil.GetFakeFactory()
			factory.ResetIOStream()

			factory.SetFail.GetK8sClientset = test.failGetClientset
			config := &schemas.GlobalConfiguration{}
			config.PreflightCheckConfiguration.RetryCount = 1
			config.PreflightCheckConfiguration.RetryDelay = 1

			if test.failCreateNamespace {
				failFunc := func(action k8sTesting.Action) (bool, runtime.Object, error) {
					return true, nil, fmt.Errorf("failed to create namespace")
				}
				modFunc := func(clientset *fake.Clientset) {
					clientset.CoreV1().Namespaces().(*fakeTypedCoreV1.FakeNamespaces).Fake.PrependReactor("create", "namespaces", failFunc)
				}
				factory.SetFail.GetK8sClientsetPrepFuncs = append(factory.SetFail.GetK8sClientsetPrepFuncs, &modFunc)
			}
			if test.failGetPod {
				modFunc := func(clientset *fake.Clientset) {
					clientset.CoreV1().Pods("preflight-check").(*fakeTypedCoreV1.FakePods).Fake.PrependReactor("list", "pods", func(action k8sTesting.Action) (bool, runtime.Object, error) {
						return true, nil, fmt.Errorf("failed to get pod")
					})
				}
				factory.SetFail.GetK8sClientsetPrepFuncs = append(factory.SetFail.GetK8sClientsetPrepFuncs, &modFunc)
			} else if !test.failTimeoutPod {
				modFunc := func(clientset *fake.Clientset) {
					clientset.CoreV1().Pods("preflight-check").(*fakeTypedCoreV1.FakePods).Fake.PrependReactor("list", "pods", func(action k8sTesting.Action) (bool, runtime.Object, error) {
						return true, &coreV1.PodList{
							Items: []coreV1.Pod{
								*podToFind,
							},
						}, nil
					})
				}
				factory.SetFail.GetK8sClientsetPrepFuncs = append(factory.SetFail.GetK8sClientsetPrepFuncs, &modFunc)
			}
			if !test.failCreateSecret {
				config.PreflightCheckConfiguration.RegistryServer = "registry.foo"
				config.PreflightCheckConfiguration.RegistryUsername = "user"
				config.PreflightCheckConfiguration.RegistryPassword = "pass"
			}
			if test.failDeleteNamespace {
				modFunc := func(clientset *fake.Clientset) {
					clientset.CoreV1().Namespaces().(*fakeTypedCoreV1.FakeNamespaces).Fake.PrependReactor("delete", "namespaces", func(action k8sTesting.Action) (bool, runtime.Object, error) {
						return true, nil, fmt.Errorf("namespaces \"preflight-check\" not found")
					})
				}
				factory.SetFail.GetK8sClientsetPrepFuncs = append(factory.SetFail.GetK8sClientsetPrepFuncs, &modFunc)
			}
			if test.failTimeoutNamespace || test.failDeleteNamespace {
				modFunc := func(clientset *fake.Clientset) {
					clientset.CoreV1().Namespaces().(*fakeTypedCoreV1.FakeNamespaces).Fake.PrependReactor("create", "namespaces", func(action k8sTesting.Action) (bool, runtime.Object, error) {
						return true, nil, k8sErrors.NewAlreadyExists(schema.GroupResource{Group: "", Resource: "namespaces"}, "preflight-check")
					})
				}
				factory.SetFail.GetK8sClientsetPrepFuncs = append(factory.SetFail.GetK8sClientsetPrepFuncs, &modFunc)
			}
			if test.failTimeoutNamespace {
				modFunc := func(clientset *fake.Clientset) {
					clientset.CoreV1().Namespaces().(*fakeTypedCoreV1.FakeNamespaces).Fake.PrependReactor("delete", "namespaces", func(action k8sTesting.Action) (bool, runtime.Object, error) {
						return true, &coreV1.Namespace{}, nil
					})
				}
				factory.SetFail.GetK8sClientsetPrepFuncs = append(factory.SetFail.GetK8sClientsetPrepFuncs, &modFunc)
			}
			if test.failCreatePod {
				modFunc := func(clientset *fake.Clientset) {
					clientset.BatchV1().Jobs("preflight-check").(*fakeTypedBatchV1.FakeJobs).Fake.PrependReactor("create", "jobs", func(action k8sTesting.Action) (bool, runtime.Object, error) {
						return true, nil, fmt.Errorf("failed to create pod")
					})
				}
				factory.SetFail.GetK8sClientsetPrepFuncs = append(factory.SetFail.GetK8sClientsetPrepFuncs, &modFunc)
			}

			// Act
			pod, msgs, err := createResourcesForCommandExecution(command, factory, config)

			// Assert
			checkOutput(t, test.expectedOut, msgs)
			if !(test.failGetClientset || test.failCreateNamespace || test.failCreateSecret || test.failGetPod || test.failTimeoutPod || test.failDeleteNamespace || test.failTimeoutNamespace || test.failCreatePod) {
				assert.NoError(t, err)
				assert.NotNil(t, pod)
				assert.Equal(t, podToFind, pod)
			} else {
				// do more testing on these
				assert.Error(t, err)
				assert.Nil(t, pod)
			}
			if test.failTimeoutNamespace || test.failTimeoutPod {
				assert.True(t, time.Since(startTime) > time.Duration(config.PreflightCheckConfiguration.RetryCount)*time.Second)
			}
		})
	}
}

// all of createNamespaceForCommandExecution is tested in the previous test
// all of createRegistrySecretForCommandExecution is tested in the previous test
// all of createJobForCommandExecution is tested in the previous test

func TestDeleteResourcesForCommandExecution(t *testing.T) {
	tests := []struct {
		desc             string
		failGetClientset bool
	}{
		{
			"success",
			false,
		},
		{
			"failed to get clientset",
			true,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			// Arrange
			command := &cobra.Command{}
			factory := bbTestUtil.GetFakeFactory()
			factory.SetFail.GetK8sClientset = test.failGetClientset
			if !test.failGetClientset {
				factory.SetObjects([]runtime.Object{ns("preflight-check", coreV1.NamespaceActive)})
			}

			// Act
			err := deleteResourcesForCommandExecution(command, factory)

			// Assert
			if !test.failGetClientset {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, "testing error", err.Error())
			}
		})
	}
}

func makeOutputSummary(checks []preflightCheck, format output.OutputFormat) string {
	steps := []outputSchemas.CheckStepOutput{}

	for _, value := range checks {
		message := "System Error Occured - Execute command again to retry"
		if value.status == passed {
			message = value.successMessage
		} else if value.status == failed {
			message = value.failureMessage
		}
		steps = append(steps, outputSchemas.CheckStepOutput{
			Name:   value.desc,
			Output: []string{message},
			Status: string(value.status),
		})
	}
	summary := &outputSchemas.PreflightCheckOutput{
		Name:  "Preflight Check Summary",
		Steps: steps,
	}

	result := ""
	switch format {
	case output.TEXT:
		byteSummary, _ := summary.MarshalHumanReadable()
		result = string(byteSummary) + "\n"
	case output.JSON:
		byteSummary, _ := summary.MarshalJson()
		result = string(byteSummary)
	case output.YAML:
		byteSummary, _ := summary.MarshalYaml()
		result = string(byteSummary)
	}
	return result
}

func fakeCheckFn(cmd *cobra.Command, factory bbUtil.Factory, config *schemas.GlobalConfiguration) ([]string, preflightCheckStatus) {
	return []string{}, passed
}

func getFakeChecks(status1 preflightCheckStatus, status2 preflightCheckStatus, status3 preflightCheckStatus) []preflightCheck {
	return []preflightCheck{
		{
			desc:           "Check 1",
			function:       fakeCheckFn,
			status:         status1,
			failureMessage: "Check 1 Failed",
			successMessage: "Check 1 Passed",
		},
		{
			desc:           "Check 2",
			function:       fakeCheckFn,
			status:         status2,
			failureMessage: "Check 2 Failed",
			successMessage: "Check 2 Passed",
		},
		{
			desc:           "Check 3",
			function:       fakeCheckFn,
			status:         status3,
			failureMessage: "Check 3 Failed",
			successMessage: "Check 3 Passed",
		},
	}
}

func TestPrintPreflightCheckSummary(t *testing.T) {
	allPassed := getFakeChecks(passed, passed, passed)
	allFailed := getFakeChecks(failed, failed, failed)
	allUnknown := getFakeChecks(unknown, unknown, unknown)
	allMixed := getFakeChecks(passed, failed, unknown)

	tests := []struct {
		desc              string
		expected          string
		checks            []preflightCheck
		failWritingOutput bool
	}{
		{
			"all passed",
			makeOutputSummary(allPassed, output.TEXT),
			allPassed,
			false,
		},
		{
			"all failed",
			makeOutputSummary(allFailed, output.TEXT),
			allFailed,
			false,
		},
		{
			"all unknown",
			makeOutputSummary(allUnknown, output.TEXT),
			allUnknown,
			false,
		},
		{
			"mixed",
			makeOutputSummary(allMixed, output.TEXT),
			allMixed,
			false,
		},
		{
			"failed to write output",
			"failed to create preflight check output: unable to write human-readable output: FakeWriter intentionally errored",
			[]preflightCheck{
				{
					desc:           "Check 1",
					function:       fakeCheckFn,
					failureMessage: "Check 1 Failed",
					successMessage: "Check 1 Passed",
				},
			},
			true,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			// Arrange
			factory := bbTestUtil.GetFakeFactory()
			factory.ResetIOStream()
			v, _ := factory.GetViper()
			v.Set("big-bang-repo", "/tmp")
			v.Set("output-config.format", "text")

			streams, _ := factory.GetIOStream()
			in := streams.In.(*bytes.Buffer)
			out := streams.Out.(*bytes.Buffer)
			errOut := streams.ErrOut.(*bytes.Buffer)

			if test.failWritingOutput {
				streams.Out = apiWrappers.CreateFakeWriterFromStream(t, test.failWritingOutput, streams.Out)
			}

			// Act
			err := printPreflightCheckSummary(nil, factory, test.checks)

			// Assert
			assert.Empty(t, in.String())
			assert.Empty(t, errOut.String())
			if !test.failWritingOutput {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, out.String())
			} else {
				assert.Error(t, err)
				assert.Equal(t, test.expected, err.Error())
				assert.Empty(t, out.String())
			}
		})
	}
}

func TestPreflightCheck(t *testing.T) {
	passFunc := func(cmd *cobra.Command, factory bbUtil.Factory, config *schemas.GlobalConfiguration) ([]string, preflightCheckStatus) {
		return []string{}, passed
	}

	failFunc := func(cmd *cobra.Command, factory bbUtil.Factory, config *schemas.GlobalConfiguration) ([]string, preflightCheckStatus) {
		return []string{}, failed
	}

	unknFunc := func(cmd *cobra.Command, factory bbUtil.Factory, config *schemas.GlobalConfiguration) ([]string, preflightCheckStatus) {
		return []string{}, unknown
	}

	var tests = []struct {
		desc     string
		expected []string
		check    preflightCheck
	}{
		{
			desc: "Check Failure",
			expected: []string{
				"Status: Failed",
				"Foo Service Down",
			},
			check: preflightCheck{
				desc:           "Foo Service Check",
				function:       failFunc,
				failureMessage: "Foo Service Down",
				successMessage: "Foo Service Up",
			},
		},
		{
			desc: "Check Success",
			expected: []string{
				"Status: Passed",
				"Foo Service Up",
			},
			check: preflightCheck{
				desc:           "Foo Service Check",
				function:       passFunc,
				failureMessage: "Foo Service Down",
				successMessage: "Foo Service Up",
			},
		},
		{
			desc: "Check Error",
			expected: []string{
				"Status: Unknown",
				"System Error",
			},
			check: preflightCheck{
				desc:           "Foo Service Check",
				function:       unknFunc,
				failureMessage: "Foo Service Down",
				successMessage: "Foo Service Up",
			},
		},
	}

	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	factory.SetObjects([]runtime.Object{})
	streams, _ := factory.GetIOStream()
	buf := streams.Out.(*bytes.Buffer)
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "/tmp")
	v.Set("output-config.format", "text")

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			command := &cobra.Command{}
			err := bbPreflightCheck(nil, factory, command, []preflightCheck{test.check})
			assert.NoError(t, err)
			output := buf.String()
			assert.Contains(t, output, test.expected[0])
			assert.Contains(t, output, test.expected[1])
			buf.Reset()
		})
	}
}

func TestPreflightCheckCmd(t *testing.T) {
	passFunc := func(cmd *cobra.Command, factory bbUtil.Factory, config *schemas.GlobalConfiguration) ([]string, preflightCheckStatus) {
		return []string{}, passed
	}

	failFunc := func(cmd *cobra.Command, factory bbUtil.Factory, config *schemas.GlobalConfiguration) ([]string, preflightCheckStatus) {
		return []string{}, failed
	}

	unknFunc := func(cmd *cobra.Command, factory bbUtil.Factory, config *schemas.GlobalConfiguration) ([]string, preflightCheckStatus) {
		return []string{}, unknown
	}

	preflightChecks = []preflightCheck{
		{
			desc:           "Foo Service Check",
			function:       failFunc,
			failureMessage: "Foo Service Down",
			successMessage: "Foo Service Up",
		},
		{
			desc:           "Bar Service Check",
			function:       passFunc,
			failureMessage: "Bar Service Down",
			successMessage: "Bar Service Up",
		},
		{
			desc:           "Hello Service Check",
			function:       unknFunc,
			failureMessage: "Hello Service Down",
			successMessage: "Hello Service Up",
		},
	}

	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	factory.SetObjects([]runtime.Object{})
	streams, _ := factory.GetIOStream()
	buf := streams.Out.(*bytes.Buffer)
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "/tmp")
	v.Set("output-config.format", "text")
	cmd, cmdError := NewPreflightCheckCmd(factory)
	assert.NoError(t, cmdError)
	err := cmd.Execute()
	assert.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Status: Failed")
	assert.Contains(t, output, "Foo Service Down")
	assert.Contains(t, output, "Status: Passed")
	assert.Contains(t, output, "Bar Service Up")
	assert.Contains(t, output, "Status: Unknown")
	assert.Contains(t, output, "System Error Occured - Execute command again to retry")
}

func TestGetPreflightCheckCmdConfigClientError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.SetObjects([]runtime.Object{})
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "/tmp")
	v.Set("output-config.format", "text")
	factory.SetFail.GetConfigClient = 1
	// Act
	cmd, cmdError := NewPreflightCheckCmd(factory)
	// Assert
	assert.Nil(t, cmd)
	assert.Error(t, cmdError)
	if !assert.Contains(t, cmdError.Error(), "unable to get config client:") {
		t.Errorf("unexpected output: %s", cmdError.Error())
	}
}

func TestPreflightCheckFailToGetConfig(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	loggingClient, _ := factory.GetLoggingClient()
	cmd, _ := NewPreflightCheckCmd(factory)
	viper, _ := factory.GetViper()
	expected := ""
	getConfigFunc := func(client *bbConfig.ConfigClient) (*schemas.GlobalConfiguration, error) {
		return &schemas.GlobalConfiguration{
			BigBangRepo: expected,
		}, fmt.Errorf("Dummy Error")
	}
	client, _ := bbConfig.NewClient(getConfigFunc, nil, &loggingClient, cmd, viper)
	factory.SetConfigClient(client)

	// Act
	err := cmd.Execute()

	// Assert
	assert.Error(t, err)
	if !assert.Contains(t, err.Error(), "error getting config:") {
		t.Errorf("unexpected output: %s", err.Error())
	}
}

func TestBBPreflightCheckConfigClientError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.SetObjects([]runtime.Object{})
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "/tmp")
	v.Set("output-config.format", "text")
	cmd, _ := NewPreflightCheckCmd(factory)
	// Act
	factory.SetFail.GetConfigClient = 1
	err := cmd.Execute()
	// Assert
	assert.NotNil(t, cmd)
	assert.Error(t, err)
	if !assert.Contains(t, err.Error(), "unable to get config client:") {
		t.Errorf("unexpected output: %s", err.Error())
	}
}

func TestOutputFormatting(t *testing.T) {
	checks := getFakeChecks(passed, failed, unknown)
	var tests = []struct {
		desc     string
		format   output.OutputFormat
		expected string
	}{
		{
			desc:     "TEXT Format",
			format:   output.TEXT,
			expected: makeOutputSummary(checks, output.TEXT),
		},
		{
			desc:     "JSON Format",
			format:   output.JSON,
			expected: makeOutputSummary(checks, output.JSON),
		},
		{
			desc:     "YAML Format",
			format:   output.YAML,
			expected: makeOutputSummary(checks, output.YAML),
		},
	}

	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	factory.SetObjects([]runtime.Object{})
	streams, _ := factory.GetIOStream()
	buf := streams.Out.(*bytes.Buffer)
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "/tmp")

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			v.Set("output-config.format", string(test.format))
			err := printPreflightCheckSummary(nil, factory, checks)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, buf.String())
			buf.Reset()
		})
	}
}
