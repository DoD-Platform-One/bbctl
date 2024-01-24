package cmd

import (
	"fmt"
	"testing"

	bbutil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
	bbtestutil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func pod(app string, ns string, phase corev1.PodPhase) *corev1.Pod {

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-fequ", app),
			Namespace: ns,
			Labels: map[string]string{
				"app": app,
			},
		},
		Status: corev1.PodStatus{
			Phase: phase,
		},
	}

	return pod
}

func TestCheckMetricsServer(t *testing.T) {

	arl := metav1.APIResourceList{
		GroupVersion: "metrics.k8s.io/v1beta1",
		APIResources: []metav1.APIResource{
			{
				Name: "PodMetrics",
			},
		},
	}

	var tests = []struct {
		desc      string
		expected  string
		resources []*metav1.APIResourceList
	}{
		{
			"metrics server not available",
			"Check Failed - Metrics API not available.",
			[]*metav1.APIResourceList{},
		},
		{
			"metrics server available",
			"Check Passed - Metrics API available.",
			[]*metav1.APIResourceList{&arl},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			factory := bbtestutil.GetFakeFactory(nil, nil, nil, test.resources)

			streams, _, buf, _ := genericclioptions.NewTestIOStreams()
			checkMetricsServer(factory, streams, nil)
			assert.Contains(t, buf.String(), test.expected)
		})
	}

}

func TestCheckDefaultStorageClass(t *testing.T) {

	barSC := &storagev1.StorageClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "bar",
		},
	}

	fooSC := &storagev1.StorageClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
			Annotations: map[string]string{
				"storageclass.kubernetes.io/is-default-class": "true",
			},
		},
	}

	var tests = []struct {
		desc     string
		expected string
		objs     []runtime.Object
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
			factory := bbtestutil.GetFakeFactory(nil, test.objs, nil, nil)
			streams, _, buf, _ := genericclioptions.NewTestIOStreams()
			checkDefaultStorageClass(factory, streams, nil)
			assert.Contains(t, buf.String(), test.expected)
		})
	}

}

func TestCheckFluxController(t *testing.T) {

	hcPodRunning := pod("helm-controller", "flux-system", corev1.PodRunning)
	hcPodFailed := pod("helm-controller", "flux-system", corev1.PodFailed)
	kcPodRunning := pod("kustomize-controller", "flux-system", corev1.PodRunning)
	kcPodFailed := pod("kustomize-controller", "flux-system", corev1.PodFailed)
	scPodRunning := pod("source-controller", "flux-system", corev1.PodRunning)
	scPodFailed := pod("source-controller", "flux-system", corev1.PodFailed)
	ncPodRunning := pod("notification-controller", "flux-system", corev1.PodRunning)
	ncPodFailed := pod("notification-controller", "flux-system", corev1.PodFailed)

	var tests = []struct {
		desc     string
		expected string
		objs     []runtime.Object
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
			factory := bbtestutil.GetFakeFactory(nil, test.objs, nil, nil)
			streams, _, buf, _ := genericclioptions.NewTestIOStreams()
			checkFluxController(factory, streams, nil)
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

	var flags *pflag.FlagSet = &pflag.FlagSet{}
	flags.String("registryserver", "registry.foo", "Image registry server url")
	flags.String("registryusername", "user", "Image registry username")
	flags.String("registrypassword", "pass", "Image registry password")

	pfcPod := pod("pfc", "preflight-check", corev1.PodRunning)
	pfcPod.ObjectMeta.Labels["job-name"] = "preflightcheck"

	factory := bbtestutil.GetFakeFactory(nil, []runtime.Object{pfcPod}, nil, nil)
	streams, _, buf, _ := genericclioptions.NewTestIOStreams()
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			bbtestutil.GetFakeCommandExecutor().CommandResult = test.commandResult
			checkSystemParameters(factory, streams, flags)
			assert.Contains(t, buf.String(), test.expected)
			buf.Reset()
		})
	}

}

func TestPreflightCheckGetParameter(t *testing.T) {

	type testInitFunc func() *pflag.FlagSet

	var tests = []struct {
		desc     string
		input    string
		expected string
		initFunc testInitFunc
	}{
		{
			desc:     "Check parameter",
			input:    "registryserver",
			expected: "registry.foo",
			initFunc: func() *pflag.FlagSet {
				var flags *pflag.FlagSet = &pflag.FlagSet{}
				flags.String("registryserver", "registry.foo", "Image registry server url")
				viper.Set("registryserver", "registry.io")
				return flags
			},
		},
		{
			desc:     "Check env variable",
			input:    "registryserver",
			expected: "registry.io",
			initFunc: func() *pflag.FlagSet {
				var flags *pflag.FlagSet = &pflag.FlagSet{}
				viper.Set("registryserver", "registry.io")
				return flags
			},
		},
		{
			desc:     "Check missing value",
			input:    "registryserver",
			expected: "",
			initFunc: func() *pflag.FlagSet {
				var flags *pflag.FlagSet = &pflag.FlagSet{}
				viper.Set("registryserver", "")
				return flags
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			value := getParameter(test.initFunc(), test.input)
			assert.Equal(t, test.expected, value)
		})
	}

}

func TestPreflightCheck(t *testing.T) {

	passFunc := func(factory bbutil.Factory, streams genericclioptions.IOStreams, flags *pflag.FlagSet) preflightCheckStatus {
		return passed
	}

	failFunc := func(factory bbutil.Factory, streams genericclioptions.IOStreams, flags *pflag.FlagSet) preflightCheckStatus {
		return failed
	}

	unknFunc := func(factory bbutil.Factory, streams genericclioptions.IOStreams, flags *pflag.FlagSet) preflightCheckStatus {
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

	factory := bbtestutil.GetFakeFactory(nil, []runtime.Object{}, nil, nil)
	streams, _, buf, _ := genericclioptions.NewTestIOStreams()

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			bbPreflightCheck(factory, streams, nil, []preflightCheck{test.check})
			output := buf.String()
			assert.Contains(t, output, test.expected[0])
			assert.Contains(t, output, test.expected[1])
			buf.Reset()
		})
	}

}

func TestPreflightCheckCmd(t *testing.T) {

	passFunc := func(factory bbutil.Factory, streams genericclioptions.IOStreams, flags *pflag.FlagSet) preflightCheckStatus {
		return passed
	}

	failFunc := func(factory bbutil.Factory, streams genericclioptions.IOStreams, flags *pflag.FlagSet) preflightCheckStatus {
		return failed
	}

	unknFunc := func(factory bbutil.Factory, streams genericclioptions.IOStreams, flags *pflag.FlagSet) preflightCheckStatus {
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

	factory := bbtestutil.GetFakeFactory(nil, []runtime.Object{}, nil, nil)
	streams, _, buf, _ := genericclioptions.NewTestIOStreams()
	cmd := NewPreflightCheckCmd(factory, streams)
	cmd.Execute()
	output := buf.String()
	assert.Contains(t, output, "Foo Service Check Failed")
	assert.Contains(t, output, "Foo Service Down")
	assert.Contains(t, output, "Bar Service Check Passed")
	assert.Contains(t, output, "Bar Service Up")
	assert.Contains(t, output, "Hello Service Check Unknown")
	assert.Contains(t, output, "System Error - Execute command again to run Hello Service Check")
}
