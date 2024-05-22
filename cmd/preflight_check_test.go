package cmd

import (
	"fmt"
	"testing"

	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/config/schemas"
	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	coreV1 "k8s.io/api/core/v1"
	storageV1 "k8s.io/api/storage/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"
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

func TestCheckMetricsServer(t *testing.T) {
	arl := metaV1.APIResourceList{
		GroupVersion: "metrics.k8s.io/v1beta1",
		APIResources: []metaV1.APIResource{
			{
				Name: "PodMetrics",
			},
		},
	}

	var tests = []struct {
		desc      string
		expected  string
		resources []*metaV1.APIResourceList
	}{
		{
			"metrics server not available",
			"Check Failed - Metrics API not available.",
			[]*metaV1.APIResourceList{},
		},
		{
			"metrics server available",
			"Check Passed - Metrics API available.",
			[]*metaV1.APIResourceList{&arl},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			factory := bbTestUtil.GetFakeFactory()
			factory.SetResources(test.resources)

			streams, _, buf, _ := genericIOOptions.NewTestIOStreams()
			checkMetricsServer(nil, factory, streams, nil)
			assert.Contains(t, buf.String(), test.expected)
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
		desc     string
		expected string
		objects  []runtime.Object
	}{
		{
			"no storage class",
			"Check Failed - Default storage class not found.",
			[]runtime.Object{},
		},
		{
			"default storage class",
			"Check Passed - Default storage class foo found.",
			[]runtime.Object{fooSC},
		},
		{
			"no default storage class",
			"Check Failed - Default storage class not found.",
			[]runtime.Object{barSC},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			factory := bbTestUtil.GetFakeFactory()
			factory.SetObjects(test.objects)
			streams, _, buf, _ := genericIOOptions.NewTestIOStreams()
			checkDefaultStorageClass(nil, factory, streams, nil)
			assert.Contains(t, buf.String(), test.expected)
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
		desc     string
		expected string
		objects  []runtime.Object
	}{
		{
			"no helm controller pod",
			"Check Failed - flux helm-controller pod not found in flux-system namespace.",
			[]runtime.Object{},
		},
		{
			"no kustomize controller pod",
			"Check Failed - flux kustomize-controller pod not found in flux-system namespace.",
			[]runtime.Object{hcPodFailed, scPodFailed, ncPodFailed},
		},
		{
			"failed kustomize controller pod",
			"Check Failed - flux kustomize-controller pod not in running state.",
			[]runtime.Object{kcPodFailed, hcPodRunning, ncPodRunning, scPodRunning},
		},
		{
			"flux controller running",
			"Check Passed - flux kustomize-controller pod running.",
			[]runtime.Object{kcPodRunning, hcPodRunning, ncPodRunning, scPodRunning},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			factory := bbTestUtil.GetFakeFactory()
			factory.SetObjects(test.objects)
			streams, _, buf, _ := genericIOOptions.NewTestIOStreams()
			checkFluxController(nil, factory, streams, nil)
			assert.Contains(t, buf.String(), test.expected)
		})
	}

}

func TestCheckSystemParameters(t *testing.T) {
	var tests = []struct {
		desc          string
		expected      string
		commandResult map[string]string
	}{
		{
			"check failed for max_map_count (ECK)",
			"Check Failed - vm.max_map_count needs to be at least 262144 for ECK to work.",
			map[string]string{"cat /proc/sys/vm/max_map_count": "262100"},
		},
		{
			"check passed for max_map_count (ECK)",
			"Check Passed - vm.max_map_count 262184 is suitable for ECK to work.",
			map[string]string{"cat /proc/sys/vm/max_map_count": "262184"},
		},
		{
			"check failed for max_map_count (Sonarqube)",
			"Check Failed - vm.max_map_count needs to be at least 524288 for Sonarqube to work.",
			map[string]string{"cat /proc/sys/vm/max_map_count": "524280"},
		},
		{
			"check passed for max_map_count (Sonarqube)",
			"Check Passed - vm.max_map_count 524288 is suitable for Sonarqube to work.",
			map[string]string{"cat /proc/sys/vm/max_map_count": "524288"},
		},
		{
			"check failed for file-max (Sonarqube)",
			"Check Failed - fs.file-max needs to be at least 131072 for Sonarqube to work.",
			map[string]string{"cat /proc/sys/fs/file-max": "131070"},
		},
		{
			"check passed for file-max (Sonarqube)",
			"Check Passed - fs.file-max 131074 is suitable for Sonarqube to work.",
			map[string]string{"cat /proc/sys/fs/file-max": "131074"},
		},
		{
			"check failed for ulimit -n (Sonarqube)",
			"Check Failed - ulimit -n needs to be at least 131072 for Sonarqube to work.",
			map[string]string{"ulimit -n": "131070"},
		},
		{
			"check passed for ulimit -n (Sonarqube)",
			"Check Passed - ulimit -n unlimited is suitable for Sonarqube to work.",
			map[string]string{"ulimit -n": "unlimited"},
		},
		{
			"check failed for ulimit -n (Sonarqube)",
			"Check Undetermined - ulimit -n needs to be at least 131072 for Sonarqube to work. Current value unknown",
			map[string]string{"ulimit -n": "unknown"},
		},
	}

	pfcPod := pod("pfc", "preflight-check", coreV1.PodRunning)
	pfcPod.ObjectMeta.Labels["job-name"] = "preflightcheck"

	command := &cobra.Command{}
	factory := bbTestUtil.GetFakeFactory()
	configClient, err := factory.GetConfigClient(command)
	assert.Nil(t, err)
	viperInstance := factory.GetViper()
	assert.Nil(t, configClient.SetAndBindFlag("big-bang-repo", "/tmp", "Location on the filesystem where the bigbang product repo is checked out"))
	assert.Nil(t, configClient.SetAndBindFlag("registryserver", "registry.foo", "Image registry server url"))
	assert.Nil(t, configClient.SetAndBindFlag("registryusername", "user", "Image registry username"))
	assert.Nil(t, configClient.SetAndBindFlag("registrypassword", "pass", "Image registry password"))
	assert.Nil(t, viperInstance.BindPFlags(command.Flags()))

	config := configClient.GetConfig()
	config.PreflightCheckConfiguration.RegistryServer = "registry.foo"
	config.PreflightCheckConfiguration.RegistryUsername = "user"
	config.PreflightCheckConfiguration.RegistryPassword = "pass"

	factory.SetObjects([]runtime.Object{pfcPod})
	streams, _, buf, _ := genericIOOptions.NewTestIOStreams()
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			bbTestUtil.GetFakeCommandExecutor().CommandResult = test.commandResult
			checkSystemParameters(command, factory, streams, config)
			assert.Contains(t, buf.String(), test.expected)
			buf.Reset()
		})
	}

}

func TestPreflightCheck(t *testing.T) {
	passFunc := func(cmd *cobra.Command, factory bbUtil.Factory, streams genericIOOptions.IOStreams, config *schemas.GlobalConfiguration) preflightCheckStatus {
		return passed
	}

	failFunc := func(cmd *cobra.Command, factory bbUtil.Factory, streams genericIOOptions.IOStreams, config *schemas.GlobalConfiguration) preflightCheckStatus {
		return failed
	}

	unknFunc := func(cmd *cobra.Command, factory bbUtil.Factory, streams genericIOOptions.IOStreams, config *schemas.GlobalConfiguration) preflightCheckStatus {
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
	factory.SetObjects([]runtime.Object{})
	streams, _, buf, _ := genericIOOptions.NewTestIOStreams()
	factory.GetViper().Set("big-bang-repo", "/tmp")

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			command := &cobra.Command{}
			err := bbPreflightCheck(nil, factory, streams, command, []preflightCheck{test.check})
			assert.NoError(t, err)
			output := buf.String()
			assert.Contains(t, output, test.expected[0])
			assert.Contains(t, output, test.expected[1])
			buf.Reset()
		})
	}

}

func TestPreflightCheckCmd(t *testing.T) {
	passFunc := func(cmd *cobra.Command, factory bbUtil.Factory, streams genericIOOptions.IOStreams, config *schemas.GlobalConfiguration) preflightCheckStatus {
		return passed
	}

	failFunc := func(cmd *cobra.Command, factory bbUtil.Factory, streams genericIOOptions.IOStreams, config *schemas.GlobalConfiguration) preflightCheckStatus {
		return failed
	}

	unknFunc := func(cmd *cobra.Command, factory bbUtil.Factory, streams genericIOOptions.IOStreams, config *schemas.GlobalConfiguration) preflightCheckStatus {
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
	factory.SetObjects([]runtime.Object{})
	streams, _, buf, _ := genericIOOptions.NewTestIOStreams()
	factory.GetViper().Set("big-bang-repo", "/tmp")
	cmd := NewPreflightCheckCmd(factory, streams)
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
