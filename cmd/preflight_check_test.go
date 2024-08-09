package cmd

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
	bbConfig "repo1.dso.mil/big-bang/product/packages/bbctl/util/config"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/config/schemas"
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
		expected         string
		resources        []*metaV1.APIResourceList
		failGetClient    bool
		failServerGroups bool
	}{
		{
			"metrics server not available",
			"Check Failed - Metrics API not available.",
			[]*metaV1.APIResourceList{},
			false,
			false,
		},
		{
			"metrics server available",
			"Check Passed - Metrics API available.",
			[]*metaV1.APIResourceList{&arl},
			false,
			false,
		},
		{
			"metrics server not available - get client failed",
			"failed to get k8s clientset",
			[]*metaV1.APIResourceList{},
			true,
			false,
		},
		{
			"metrics server not available - server groups failed",
			"unexpected GroupVersion string: this/is/wrong",
			[]*metaV1.APIResourceList{&badArl},
			false,
			true,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			factory := bbTestUtil.GetFakeFactory()
			factory.ResetIOStream()
			factory.SetResources(test.resources)
			factory.SetFail.GetK8sClientset = test.failGetClient

			streams, _ := factory.GetIOStream()
			in := streams.In.(*bytes.Buffer)
			out := streams.Out.(*bytes.Buffer)
			errOut := streams.ErrOut.(*bytes.Buffer)

			checkMetricsServer(nil, factory, nil)
			assert.Empty(t, in.String())
			if !(test.failGetClient || test.failServerGroups) {
				assert.Contains(t, out.String(), test.expected)
				assert.Empty(t, errOut.String())
			} else {
				assert.Contains(t, errOut.String(), test.expected)
				assert.Equal(t, "Checking metrics server...\n", out.String())
			}
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
		expected             string
		objects              []runtime.Object
		failGetClientset     bool
		failListStorageClass bool
	}{
		{
			"no storage class",
			"Check Failed - Default storage class not found.",
			[]runtime.Object{},
			false,
			false,
		},
		{
			"default storage class",
			"Check Passed - Default storage class foo found.",
			[]runtime.Object{fooSC},
			false,
			false,
		},
		{
			"no default storage class",
			"Check Failed - Default storage class not found.",
			[]runtime.Object{barSC},
			false,
			false,
		},
		{
			"failed to get clientset",
			"failed to get k8s clientset",
			[]runtime.Object{},
			true,
			false,
		},
		{
			"failed to list storage class",
			"failed to list storage class",
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

			streams, _ := factory.GetIOStream()
			in := streams.In.(*bytes.Buffer)
			out := streams.Out.(*bytes.Buffer)
			errOut := streams.ErrOut.(*bytes.Buffer)

			if test.failListStorageClass {
				failFunc := func(action k8sTesting.Action) (bool, runtime.Object, error) {
					return true, nil, fmt.Errorf("failed to list storage class")
				}
				modFunc := func(clientset *fake.Clientset) {
					clientset.StorageV1().StorageClasses().(*fakeTyped.FakeStorageClasses).Fake.PrependReactor("list", "storageclasses", failFunc)
				}
				factory.SetFail.GetK8sClientsetPrepFuncs = append(factory.SetFail.GetK8sClientsetPrepFuncs, &modFunc)
			}

			// Act
			checkDefaultStorageClass(nil, factory, nil)

			// Assert
			assert.Empty(t, in.String())
			if !(test.failGetClientset || test.failListStorageClass) {
				assert.Contains(t, out.String(), test.expected)
				assert.Empty(t, errOut.String())
			} else {
				assert.Equal(t, "Checking default storage class...\n", out.String())
				assert.Contains(t, errOut.String(), test.expected)
			}
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
		expected         string
		objects          []runtime.Object
		failGetClientset bool
		failListPods     bool
	}{
		{
			"no helm controller pod",
			"Check Failed - flux helm-controller pod not found in flux-system namespace.",
			[]runtime.Object{},
			false,
			false,
		},
		{
			"no kustomize controller pod",
			"Check Failed - flux kustomize-controller pod not found in flux-system namespace.",
			[]runtime.Object{hcPodFailed, scPodFailed, ncPodFailed},
			false,
			false,
		},
		{
			"failed kustomize controller pod",
			"Check Failed - flux kustomize-controller pod not in running state.",
			[]runtime.Object{kcPodFailed, hcPodRunning, ncPodRunning, scPodRunning},
			false,
			false,
		},
		{
			"flux controller running",
			"Check Passed - flux kustomize-controller pod running.",
			[]runtime.Object{kcPodRunning, hcPodRunning, ncPodRunning, scPodRunning},
			false,
			false,
		},
		{
			"failed to get clientset",
			"failed to get k8s clientset",
			[]runtime.Object{},
			true,
			false,
		},
		{
			"failed to list pods",
			"failed to list pods",
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

			streams, _ := factory.GetIOStream()
			in := streams.In.(*bytes.Buffer)
			out := streams.Out.(*bytes.Buffer)
			errOut := streams.ErrOut.(*bytes.Buffer)

			if test.failListPods {
				failFunc := func(action k8sTesting.Action) (bool, runtime.Object, error) {
					return true, nil, fmt.Errorf("failed to list pods")
				}
				modFunc := func(clientset *fake.Clientset) {
					clientset.CoreV1().Pods("flux-system").(*fakeTypedCoreV1.FakePods).Fake.PrependReactor("list", "pods", failFunc)
				}
				factory.SetFail.GetK8sClientsetPrepFuncs = append(factory.SetFail.GetK8sClientsetPrepFuncs, &modFunc)
			}

			// Act
			checkFluxController(nil, factory, nil)

			// Assert
			assert.Empty(t, in.String())
			if !(test.failGetClientset || test.failListPods) {
				assert.Contains(t, out.String(), test.expected)
				assert.Empty(t, errOut.String())
			} else {
				assert.Equal(t, "Checking flux installation...\n", out.String())
				assert.Contains(t, errOut.String(), test.expected)
			}
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
		expected               string
		paramOverrides         map[string]string
		checkFailed            bool
		failGetClientset       bool
		failGetCommandExecutor bool
		failDeleteNamespace    bool
		extraExpected          []string
	}{
		{
			"check failed for max_map_count (ECK)",
			"Check Failed - vm.max_map_count needs to be at least 262144 for ECK to work.",
			map[string]string{"cat /proc/sys/vm/max_map_count": "262100"},
			true,
			false,
			false,
			false,
			[]string{},
		},
		{
			"check passed for max_map_count (ECK)",
			"Check Passed - vm.max_map_count 262144 is suitable for ECK to work.",
			map[string]string{"cat /proc/sys/vm/max_map_count": "262144"},
			true, // Sonarqube is higher than ECK, so this should fail
			false,
			false,
			false,
			[]string{},
		},
		{
			"check failed for max_map_count (Sonarqube)",
			"Check Failed - vm.max_map_count needs to be at least 524288 for Sonarqube to work.",
			map[string]string{"cat /proc/sys/vm/max_map_count": "524280"},
			true,
			false,
			false,
			false,
			[]string{},
		},
		{
			"check passed for max_map_count (Sonarqube)",
			"Check Passed - vm.max_map_count 524288 is suitable for Sonarqube to work.",
			map[string]string{"cat /proc/sys/vm/max_map_count": "524288"},
			false,
			false,
			false,
			false,
			[]string{},
		},
		{
			"check failed for file-max (Sonarqube)",
			"Check Failed - fs.file-max needs to be at least 131072 for Sonarqube to work.",
			map[string]string{"cat /proc/sys/fs/file-max": "131070"},
			true,
			false,
			false,
			false,
			[]string{},
		},
		{
			"check passed for file-max (Sonarqube)",
			"Check Passed - fs.file-max 131074 is suitable for Sonarqube to work.",
			map[string]string{"cat /proc/sys/fs/file-max": "131074"},
			false,
			false,
			false,
			false,
			[]string{},
		},
		{
			"check failed for ulimit -n (Sonarqube)",
			"Check Failed - ulimit -n needs to be at least 131072 for Sonarqube to work.",
			map[string]string{"ulimit -n": "131070"},
			true,
			false,
			false,
			false,
			[]string{},
		},
		{
			"check passed for ulimit -n (Sonarqube) unlimited",
			"Check Passed - ulimit -n unlimited is suitable for Sonarqube to work.",
			map[string]string{"ulimit -n": "unlimited"},
			false,
			false,
			false,
			false,
			[]string{},
		},
		{
			"check failed for ulimit -n (Sonarqube) unknown",
			"Check Undetermined - ulimit -n needs to be at least 131072 for Sonarqube to work. Current value unknown",
			map[string]string{"ulimit -n": "unknown"},
			true,
			false,
			false,
			false,
			[]string{},
		},
		{
			"check failed for ulimit -u (Sonarqube)",
			"Check Failed - ulimit -u needs to be at least 8192 for Sonarqube to work.",
			map[string]string{"ulimit -u": "8190"},
			true,
			false,
			false,
			false,
			[]string{},
		},
		{
			"check passed for ulimit -u (Sonarqube)",
			"Check Passed - ulimit -u 8192 is suitable for Sonarqube to work.",
			map[string]string{"ulimit -u": "8192"},
			false,
			false,
			false,
			false,
			[]string{},
		},
		{
			"failed to get clientset",
			"failed to get k8s clientset",
			map[string]string{},
			false,
			true,
			false,
			false,
			[]string{},
		},
		{
			"failed to get command executor",
			"failed to get command executor",
			map[string]string{},
			false,
			false,
			true,
			false,
			fullExpectedPassingOutput,
		},
		{
			"failed to delete namespace",
			"namespaces \"preflight-check\" not found",
			map[string]string{},
			false,
			false,
			false,
			true,
			[]string{
				"Creating namespace for command execution...",
				"Creating registry secret for command execution...",
				"Creating job for command execution...",
				"Waiting for job preflightcheck to be ready...",
				"Checking system parameters...",
				"Checking vm.max_map_count",
				"vm.max_map_count = 524288",
				"Check Passed - vm.max_map_count 524288 is suitable for ECK to work.",
				"vm.max_map_count = 524288",
				"Check Passed - vm.max_map_count 524288 is suitable for Sonarqube to work.",
				"Checking fs.file-max",
				"fs.file-max = 131072",
				"Check Passed - fs.file-max 131072 is suitable for Sonarqube to work.",
				"Checking ulimit -n",
				"ulimit -n = 131072",
				"Check Passed - ulimit -n 131072 is suitable for Sonarqube to work.",
				"Checking ulimit -u",
				"ulimit -u = 8192",
				"Check Passed - ulimit -u 8192 is suitable for Sonarqube to work.",
				"Deleting namespace for command execution...",
			},
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
			assert.Nil(t, configClient.SetAndBindFlag("big-bang-repo", "/tmp", "Location on the filesystem where the bigbang product repo is checked out"))
			assert.Nil(t, configClient.SetAndBindFlag("registryserver", "registry.foo", "Image registry server url"))
			assert.Nil(t, configClient.SetAndBindFlag("registryusername", "user", "Image registry username"))
			assert.Nil(t, configClient.SetAndBindFlag("registrypassword", "pass", "Image registry password"))
			assert.Nil(t, viperInstance.BindPFlags(command.Flags()))

			config, configErr := configClient.GetConfig()
			assert.NoError(t, configErr)
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

			streams, _ := factory.GetIOStream()
			in := streams.In.(*bytes.Buffer)
			out := streams.Out.(*bytes.Buffer)
			errOut := streams.ErrOut.(*bytes.Buffer)

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
			status := checkSystemParameters(command, factory, config)

			// Assert
			assert.Empty(t, in.String())
			if !(test.failGetClientset || test.failGetCommandExecutor || test.failDeleteNamespace) {
				assert.Contains(t, out.String(), test.expected)
				assert.Empty(t, errOut.String())
				if test.checkFailed {
					assert.Equal(t, failed, status)
				} else {
					assert.Equal(t, passed, status)
				}
			} else {
				for _, line := range test.extraExpected {
					assert.Contains(t, out.String(), line)
				}
				assert.Contains(t, errOut.String(), test.expected)
				assert.Equal(t, unknown, status)
			}
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
		expectedOut          string
		expectedErrOut       string
		failGetClientset     bool
		failCreateNamespace  bool
		failCreateSecret     bool
		failGetPod           bool
		failTimeoutPod       bool
		failDeleteNamespace  bool
		failTimeoutNamespace bool
		failCreatePod        bool
		errorString          string
	}{
		{
			"success",
			"Creating namespace for command execution...\nCreating registry secret for command execution...\nCreating job for command execution...\nWaiting for job preflightcheck to be ready...\n",
			"",
			false,
			false,
			false,
			false,
			false,
			false,
			false,
			false,
			"",
		},
		{
			"failed to get clientset",
			"",
			"",
			true,
			false,
			false,
			false,
			false,
			false,
			false,
			false,
			"failed to get k8s clientset",
		},
		{
			"failed to create namespace",
			"Creating namespace for command execution...\n",
			"",
			false,
			true,
			false,
			false,
			false,
			false,
			false,
			false,
			"failed to create namespace",
		},
		{
			"failed to create secret",
			"Creating namespace for command execution...\nCreating registry secret for command execution...\n",
			"",
			false,
			false,
			true,
			false,
			false,
			false,
			false,
			false,
			"\n***Invalid registry credentials provided. Ensure the registry server, username, and password values are all set!***",
		},
		{
			"failed to get pod",
			"Creating namespace for command execution...\nCreating registry secret for command execution...\nCreating job for command execution...\nWaiting for job preflightcheck to be ready...\n",
			"",
			false,
			false,
			false,
			true,
			false,
			false,
			false,
			false,
			"failed to get pod",
		},
		{
			"failed to timeout pod",
			"Creating namespace for command execution...\nCreating registry secret for command execution...\nCreating job for command execution...\nWaiting for job preflightcheck to be ready...\n",
			"",
			false,
			false,
			false,
			false,
			true,
			false,
			false,
			false,
			"timeout waiting for command execution job to be ready",
		},
		{
			"failed to delete namespace",
			"Creating namespace for command execution...\nNamespace preflight-check already exists... It will be recreated\n",
			"",
			false,
			false,
			false,
			false,
			false,
			true,
			false,
			false,
			"namespaces \"preflight-check\" not found",
		},
		{
			"failed to timeout namespace",
			"Creating namespace for command execution...\nNamespace preflight-check already exists... It will be recreated\n",
			"",
			false,
			false,
			false,
			false,
			false,
			false,
			true,
			false,
			"namespaces \"preflight-check\" already exists",
		},
		{
			"failed to create pod",
			"Creating namespace for command execution...\nCreating registry secret for command execution...\nCreating job for command execution...\n",
			"",
			false,
			false,
			false,
			false,
			false,
			false,
			false,
			true,
			"failed to create pod",
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			// Arrange
			startTime := time.Now()
			command := &cobra.Command{}
			factory := bbTestUtil.GetFakeFactory()
			factory.ResetIOStream()
			streams, _ := factory.GetIOStream()
			in := streams.In.(*bytes.Buffer)
			out := streams.Out.(*bytes.Buffer)
			errOut := streams.ErrOut.(*bytes.Buffer)

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
			pod, err := createResourcesForCommandExecution(command, factory, config)

			// Assert
			assert.Empty(t, in.String())
			assert.Equal(t, test.expectedOut, out.String())
			assert.Equal(t, test.expectedErrOut, errOut.String())
			if !(test.failGetClientset || test.failCreateNamespace || test.failCreateSecret || test.failGetPod || test.failTimeoutPod || test.failDeleteNamespace || test.failTimeoutNamespace || test.failCreatePod) {
				assert.NoError(t, err)
				assert.NotNil(t, pod)
				assert.Equal(t, podToFind, pod)
			} else {
				// do more testing on these
				assert.Error(t, err)
				assert.Equal(t, test.errorString, err.Error())
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
		expectedOut      string
		expectedErrOut   string
		failGetClientset bool
	}{
		{
			"success",
			"Deleting namespace for command execution...\n",
			"",
			false,
		},
		{
			"failed to get clientset",
			"",
			"",
			true,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			// Arrange
			command := &cobra.Command{}
			factory := bbTestUtil.GetFakeFactory()
			factory.ResetIOStream()
			streams, _ := factory.GetIOStream()
			in := streams.In.(*bytes.Buffer)
			out := streams.Out.(*bytes.Buffer)
			errOut := streams.ErrOut.(*bytes.Buffer)

			factory.SetFail.GetK8sClientset = test.failGetClientset
			if !test.failGetClientset {
				factory.SetObjects([]runtime.Object{ns("preflight-check", coreV1.NamespaceActive)})
			}

			// Act
			err := deleteResourcesForCommandExecution(command, factory)

			// Assert
			assert.Empty(t, in.String())
			assert.Equal(t, test.expectedOut, out.String())
			assert.Equal(t, test.expectedErrOut, errOut.String())
			if !test.failGetClientset {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, "failed to get k8s clientset", err.Error())
			}
		})
	}
}

// all of execCommandInPod is tested in check system parameters test

func TestPrintPreflightCheckSummary(t *testing.T) {
	tests := []struct {
		desc              string
		expected          string
		checks            []preflightCheck
		failWritingOutput bool
	}{
		{
			"all passed",
			"\n\nPreflight Check Summary\n\nCheck 1 ...\nCheck 1 Failed\n\nCheck 2 ...\nCheck 2 Failed\n\nCheck 3 ...\nCheck 3 Failed\n\n",
			[]preflightCheck{
				{
					desc: "Check 1",
					function: func(cmd *cobra.Command, factory bbUtil.Factory, config *schemas.GlobalConfiguration) preflightCheckStatus {
						return passed
					},
					failureMessage: "Check 1 Failed",
					successMessage: "Check 1 Passed",
				},
				{
					desc: "Check 2",
					function: func(cmd *cobra.Command, factory bbUtil.Factory, config *schemas.GlobalConfiguration) preflightCheckStatus {
						return passed
					},
					failureMessage: "Check 2 Failed",
					successMessage: "Check 2 Passed",
				},
				{
					desc: "Check 3",
					function: func(cmd *cobra.Command, factory bbUtil.Factory, config *schemas.GlobalConfiguration) preflightCheckStatus {
						return passed
					},
					failureMessage: "Check 3 Failed",
					successMessage: "Check 3 Passed",
				},
			},
			false,
		},
		{
			"all failed",
			"\n\nPreflight Check Summary\n\nCheck 1 ...\nCheck 1 Failed\n\nCheck 2 ...\nCheck 2 Failed\n\nCheck 3 ...\nCheck 3 Failed\n\n",
			[]preflightCheck{
				{
					desc: "Check 1",
					function: func(cmd *cobra.Command, factory bbUtil.Factory, config *schemas.GlobalConfiguration) preflightCheckStatus {
						return failed
					},
					failureMessage: "Check 1 Failed",
					successMessage: "Check 1 Passed",
				},
				{
					desc: "Check 2",
					function: func(cmd *cobra.Command, factory bbUtil.Factory, config *schemas.GlobalConfiguration) preflightCheckStatus {
						return failed
					},
					failureMessage: "Check 2 Failed",
					successMessage: "Check 2 Passed",
				},
				{
					desc: "Check 3",
					function: func(cmd *cobra.Command, factory bbUtil.Factory, config *schemas.GlobalConfiguration) preflightCheckStatus {
						return failed
					},
					failureMessage: "Check 3 Failed",
					successMessage: "Check 3 Passed",
				},
			},
			false,
		},
		{
			"all unknown",
			"\n\nPreflight Check Summary\n\nCheck 1 ...\nCheck 1 Failed\n\nCheck 2 ...\nCheck 2 Failed\n\nCheck 3 ...\nCheck 3 Failed\n\n",
			[]preflightCheck{
				{
					desc: "Check 1",
					function: func(cmd *cobra.Command, factory bbUtil.Factory, config *schemas.GlobalConfiguration) preflightCheckStatus {
						return unknown
					},
					failureMessage: "Check 1 Failed",
					successMessage: "Check 1 Passed",
				},
				{
					desc: "Check 2",
					function: func(cmd *cobra.Command, factory bbUtil.Factory, config *schemas.GlobalConfiguration) preflightCheckStatus {
						return unknown
					},
					failureMessage: "Check 2 Failed",
					successMessage: "Check 2 Passed",
				},
				{
					desc: "Check 3",
					function: func(cmd *cobra.Command, factory bbUtil.Factory, config *schemas.GlobalConfiguration) preflightCheckStatus {
						return unknown
					},
					failureMessage: "Check 3 Failed",
					successMessage: "Check 3 Passed",
				},
			},
			false,
		},
		{
			"mixed",
			"\n\nPreflight Check Summary\n\nCheck 1 ...\nCheck 1 Failed\n\nCheck 2 ...\nCheck 2 Failed\n\nCheck 3 ...\nCheck 3 Failed\n\n",
			[]preflightCheck{
				{
					desc: "Check 1",
					function: func(cmd *cobra.Command, factory bbUtil.Factory, config *schemas.GlobalConfiguration) preflightCheckStatus {
						return passed
					},
					failureMessage: "Check 1 Failed",
					successMessage: "Check 1 Passed",
				},
				{
					desc: "Check 2",
					function: func(cmd *cobra.Command, factory bbUtil.Factory, config *schemas.GlobalConfiguration) preflightCheckStatus {
						return failed
					},
					failureMessage: "Check 2 Failed",
					successMessage: "Check 2 Passed",
				},
				{
					desc: "Check 3",
					function: func(cmd *cobra.Command, factory bbUtil.Factory, config *schemas.GlobalConfiguration) preflightCheckStatus {
						return unknown
					},
					failureMessage: "Check 3 Failed",
					successMessage: "Check 3 Passed",
				},
			},
			false,
		},
		{
			"failed to write output",
			"FakeWriter intentionally errored\nFakeWriter intentionally errored",
			[]preflightCheck{
				{
					desc: "Check 1",
					function: func(cmd *cobra.Command, factory bbUtil.Factory, config *schemas.GlobalConfiguration) preflightCheckStatus {
						return passed
					},
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
			streams, _ := factory.GetIOStream()
			in := streams.In.(*bytes.Buffer)
			out := streams.Out.(*bytes.Buffer)
			errOut := streams.ErrOut.(*bytes.Buffer)

			if test.failWritingOutput {
				streams.Out = apiWrappers.CreateFakeWriterFromStream(t, test.failWritingOutput, streams.Out)
			}

			// Act
			err := printPreflightCheckSummary(factory, test.checks)

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
	passFunc := func(cmd *cobra.Command, factory bbUtil.Factory, config *schemas.GlobalConfiguration) preflightCheckStatus {
		return passed
	}

	failFunc := func(cmd *cobra.Command, factory bbUtil.Factory, config *schemas.GlobalConfiguration) preflightCheckStatus {
		return failed
	}

	unknFunc := func(cmd *cobra.Command, factory bbUtil.Factory, config *schemas.GlobalConfiguration) preflightCheckStatus {
		return unknown
	}

	var tests = []struct {
		desc     string
		expected []string
		check    preflightCheck
	}{
		{
			desc: "Check Failure",
			expected: []string{
				"Foo Service Check Failed",
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
				"Foo Service Check Passed",
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
				"Foo Service Check Unknown",
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
	passFunc := func(cmd *cobra.Command, factory bbUtil.Factory, config *schemas.GlobalConfiguration) preflightCheckStatus {
		return passed
	}

	failFunc := func(cmd *cobra.Command, factory bbUtil.Factory, config *schemas.GlobalConfiguration) preflightCheckStatus {
		return failed
	}

	unknFunc := func(cmd *cobra.Command, factory bbUtil.Factory, config *schemas.GlobalConfiguration) preflightCheckStatus {
		return unknown
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
	cmd, cmdError := NewPreflightCheckCmd(factory)
	assert.NoError(t, cmdError)
	err := cmd.Execute()
	assert.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Foo Service Check Failed")
	assert.Contains(t, output, "Foo Service Down")
	assert.Contains(t, output, "Bar Service Check Passed")
	assert.Contains(t, output, "Bar Service Up")
	assert.Contains(t, output, "Hello Service Check Unknown")
	assert.Contains(t, output, "System Error - Execute command again to run Hello Service Check")
}

func TestGetPreflightCheckCmdConfigClientError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.SetObjects([]runtime.Object{})
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "/tmp")
	factory.SetFail.GetConfigClient = true
	// Act
	cmd, cmdError := NewPreflightCheckCmd(factory)
	// Assert
	assert.Nil(t, cmd)
	assert.Error(t, cmdError)
	if !assert.Contains(t, cmdError.Error(), "Unable to get config client:") {
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
	cmd, _ := NewPreflightCheckCmd(factory)
	// Act
	factory.SetFail.GetConfigClient = true
	err := cmd.Execute()
	// Assert
	assert.NotNil(t, cmd)
	assert.Error(t, err)
	if !assert.Contains(t, err.Error(), "Unable to get config client:") {
		t.Errorf("unexpected output: %s", err.Error())
	}
}
