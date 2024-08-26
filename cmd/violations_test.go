package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
	bbConfig "repo1.dso.mil/big-bang/product/packages/bbctl/util/config"
	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeSchema "k8s.io/apimachinery/pkg/runtime/schema"
	dynamicFake "k8s.io/client-go/dynamic/fake"
	fake "k8s.io/client-go/kubernetes/fake"
	typedFake "k8s.io/client-go/kubernetes/typed/core/v1/fake"
	k8sTesting "k8s.io/client-go/testing"
)

func eventGK(rName string, rKind string, ns string, reason string, msg string, time time.Time) *v1.Event {
	annotations := make(map[string]string)
	annotations["resource_name"] = rName
	annotations["resource_kind"] = rKind
	annotations["resource_namespace"] = ns

	evt := &v1.Event{
		ObjectMeta: metaV1.ObjectMeta{
			Name:              rName,
			Annotations:       annotations,
			CreationTimestamp: metaV1.Time{Time: time},
		},
		Reason:  reason,
		Message: msg,
	}

	return evt
}

func eventKyverno(rName string, rKind string, ns string, component string, msg string, time time.Time) *v1.Event {
	evt := &v1.Event{
		ObjectMeta: metaV1.ObjectMeta{
			Name:              rName,
			Namespace:         ns,
			CreationTimestamp: metaV1.Time{Time: time},
		},
		InvolvedObject: v1.ObjectReference{
			Name: rName,
			Kind: rKind,
		},
		Source: v1.EventSource{
			Component: component,
		},
		Message: msg,
	}

	return evt
}

func violationsCmd(factory bbUtil.Factory, ns string, args []string) *cobra.Command {
	cmd, _ := NewViolationsCmd(factory)
	cmd.PersistentFlags().StringP("namespace", "n", "", "namespace")
	cmdArgs := []string{}
	if ns != "" {
		cmdArgs = []string{"--namespace", ns}
	}
	cmdArgs = append(cmdArgs, args...)
	cmd.SetArgs(cmdArgs)
	return cmd
}

func TestGetViolationsWithConfigError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.SetHelmReleases(nil)
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "test")
	factory.SetFail.GetConfigClient = 1

	// Act
	factory.SetFail.GetConfigClient = 1
	cmd, err := NewViolationsCmd(factory)

	// Assert
	assert.Nil(t, cmd)
	assert.Error(t, err)
	if !assert.Contains(t, err.Error(), "unable to get config client:") {
		t.Errorf("unexpected output: %s", err.Error())
	}
}

func TestGetViolationsWithK8sClientsetError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.SetHelmReleases(nil)
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "test")
	factory.SetFail.GetK8sClientset = true
	cmd, _ := NewViolationsCmd(factory)
	// Act
	err := cmd.Execute()
	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "testing error")
}

func TestViolationsFailToGetConfig(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.SetHelmReleases(nil)
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "test")
	v.Set("output-config.format", "yaml")
	factory.SetFail.GetConfigClient = 1

	// Act
	cmd, err := NewViolationsCmd(factory)

	// Assert
	assert.Nil(t, cmd)
	assert.Error(t, err)
	if !assert.Contains(t, err.Error(), "unable to get config client") {
		t.Errorf("unexpected output: %s", err.Error())
	}

}

func TestViolationsCmdHelperError(t *testing.T) {
	testCases := []struct {
		name                    string
		shouldFail              bool
		errorOnK8sDynamicClient bool
		errorOnK8sClientset     bool
		errorOnLoggingClient    bool
		errorOnConfigClient     bool
		errorOnIOStream         bool
		expectedErrorMessage    string
	}{
		{
			name:                    "should not error",
			shouldFail:              false,
			errorOnK8sDynamicClient: false,
			errorOnK8sClientset:     false,
			errorOnLoggingClient:    false,
			errorOnConfigClient:     false,
			errorOnIOStream:         false,
			expectedErrorMessage:    "",
		},
		{
			name:                    "error on k8s dynamic client",
			shouldFail:              true,
			errorOnK8sDynamicClient: true,
			errorOnK8sClientset:     false,
			errorOnLoggingClient:    false,
			errorOnConfigClient:     false,
			errorOnIOStream:         false,
			expectedErrorMessage:    "failed to get K8sDynamicClient",
		},
		{
			name:                    "error on k8s clientset",
			shouldFail:              true,
			errorOnK8sDynamicClient: false,
			errorOnK8sClientset:     true,
			errorOnLoggingClient:    false,
			errorOnConfigClient:     false,
			errorOnIOStream:         false,
			expectedErrorMessage:    "testing error",
		},
		{
			name:                    "error on logging client",
			shouldFail:              true,
			errorOnK8sDynamicClient: false,
			errorOnK8sClientset:     false,
			errorOnLoggingClient:    true,
			errorOnConfigClient:     false,
			errorOnIOStream:         false,
			expectedErrorMessage:    "failed to get logging client",
		},
		{
			name:                    "error on config client",
			shouldFail:              true,
			errorOnK8sDynamicClient: false,
			errorOnK8sClientset:     false,
			errorOnLoggingClient:    false,
			errorOnConfigClient:     true,
			errorOnIOStream:         false,
			expectedErrorMessage:    "failed to get config client",
		},
		{
			name:                    "error on io stream",
			shouldFail:              true,
			errorOnK8sDynamicClient: false,
			errorOnK8sClientset:     false,
			errorOnLoggingClient:    false,
			errorOnConfigClient:     false,
			errorOnIOStream:         true,
			expectedErrorMessage:    "failed to get streams",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			factory := bbTestUtil.GetFakeFactory()
			factory.SetHelmReleases(nil)
			v, _ := factory.GetViper()
			v.Set("big-bang-repo", "test")
			cmd, _ := NewViolationsCmd(factory)
			if tc.errorOnK8sDynamicClient {
				factory.SetFail.GetK8sDynamicClient = true
			}
			if tc.errorOnK8sClientset {
				factory.SetFail.GetK8sClientset = true
			}
			if tc.errorOnLoggingClient {
				factory.SetFail.GetLoggingClient = true
			}
			if tc.errorOnConfigClient {
				factory.SetFail.GetConfigClient = 1
			}
			if tc.errorOnIOStream {
				factory.SetFail.GetIOStreams = 1
			}
			// Act
			result, err := newViolationsCmdHelper(cmd, factory)
			// Assert
			if tc.shouldFail {
				assert.Nil(t, result)
				assert.NotNil(t, cmd)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrorMessage)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func gvrToListKindForGatekeeper() map[runtimeSchema.GroupVersionResource]string {
	return map[runtimeSchema.GroupVersionResource]string{
		{
			Group:    "apiextensions.k8s.io",
			Version:  "v1",
			Resource: "customresourcedefinitions",
		}: "customresourcedefinitionsList",
		{
			Group:    "constraints.gatekeeper.sh",
			Version:  "v1beta1",
			Resource: "foos",
		}: "gkPolicyList",
	}
}

func gvrToListKindForKyverno() map[runtimeSchema.GroupVersionResource]string {
	return map[runtimeSchema.GroupVersionResource]string{
		{
			Group:    "apiextensions.k8s.io",
			Version:  "v1",
			Resource: "customresourcedefinitions",
		}: "customresourcedefinitionsList",
		{
			Group:    "kyverno.io",
			Version:  "v1",
			Resource: "foos",
		}: "kyvernoPolicyList",
		{
			Group:    "kyverno.io",
			Version:  "v1",
			Resource: "foo",
		}: "kyvernoPolicyList",
	}
}

func TestGatekeeperAuditViolations(t *testing.T) {
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "apiextensions.k8s.io/v1",
		"kind":       "customresourcedefinition",
		"metadata": map[string]interface{}{
			"name": "foos.constraints.gatekeeper.sh",
			"labels": map[string]interface{}{
				"app.kubernetes.io/name": "gatekeeper",
			},
		},
	})

	crdList := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1",
			"kind":       "customresourcedefinitionList",
		},
		Items: []unstructured.Unstructured{*crd},
	}

	constraint := &unstructured.Unstructured{}
	constraint.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "constraints.gatekeeper.sh/v1beta1",
		"kind":       "foo",
		"metadata": map[string]interface{}{
			"name": "foo-1",
			"labels": map[string]interface{}{
				"app.kubernetes.io/name": "gatekeeper",
			},
		},
		"status": map[string]interface{}{
			"auditTimestamp": "2021-11-27T23:55:33Z",
			"violations": []interface{}{
				map[string]interface{}{"kind": "k1", "name": "r1", "namespace": "ns1", "message": "invalid config"},
				map[string]interface{}{"kind": "k2", "name": "r2", "namespace": "ns2", "message": "invalid config"},
			},
		},
	})

	constraintList := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"apiVersion": "constraints.gatekeeper.sh/v1beta1",
			"kind":       "gkPolicyList",
		},
		Items: []unstructured.Unstructured{*constraint},
	}

	violation1 := `- name: r1
  kind: k1
  namespace: ns1
  policy: ""
  constraint: '%!s(<nil>)'
  message: invalid config
  action: '%!s(<nil>)'
  timestamp: "2021-11-27T23:55:33Z"`
	violation2 := `- name: r2
  kind: k2
  namespace: ns2
  policy: ""
  constraint: '%!s(<nil>)'
  message: invalid config
  action: '%!s(<nil>)'
  timestamp: "2021-11-27T23:55:33Z"`

	var tests = []struct {
		desc       string
		namespace  string
		expected   string
		unexpected string
		objects    []runtime.Object
	}{
		{
			"no violations in given namespace",
			"ns0",
			"name: No Violations\nviolations: []",
			violation1,
			[]runtime.Object{crdList, constraintList},
		},
		{
			"no violations in any namespace",
			"",
			"name: No Violations\nviolations: []",
			violation1,
			[]runtime.Object{},
		},

		{
			"violations in given namespace 1",
			"ns1",
			violation1,
			violation2,
			[]runtime.Object{crdList, constraintList},
		},
		{
			"violations in given namespace 2",
			"ns2",
			violation2,
			violation1,
			[]runtime.Object{crdList, constraintList},
		},
		{
			"violations in any namespace",
			"",
			violation2,
			"name: No Violations\nviolations: []",
			[]runtime.Object{crdList, constraintList},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			factory := bbTestUtil.GetFakeFactory()
			factory.ResetIOStream()
			factory.SetObjects(test.objects)
			factory.SetGVRToListKind(gvrToListKindForGatekeeper())
			streams, _ := factory.GetIOStream()
			buf := streams.Out.(*bytes.Buffer)
			v, _ := factory.GetViper()
			v.Set("big-bang-repo", "test")
			v.Set("output-config.format", "yaml")
			cmd := violationsCmd(factory, test.namespace, []string{"--audit"})
			err := cmd.Execute()
			if !strings.Contains(buf.String(), test.expected) {
				t.Errorf("unexpected output:\nOutput:\n%s\nExpected:\n%s", buf.String(), test.expected)
			}
			if strings.Contains(buf.String(), test.unexpected) && test.unexpected != "" {
				t.Errorf("unexpected output:\nOutput:\n%s\nUnexpected:\n%s", buf.String(), test.unexpected)
			}
			assert.Nil(t, err)
		})
	}
}

func TestGatekeeperDenyViolations(t *testing.T) {
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "apiextensions.k8s.io/v1",
		"kind":       "customresourcedefinition",
		"metadata": map[string]interface{}{
			"name": "foos.constraints.gatekeeper.sh",
			"labels": map[string]interface{}{
				"app.kubernetes.io/name": "gatekeeper",
			},
		},
	})

	crdList := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1",
			"kind":       "customresourcedefinitionList",
		},
		Items: []unstructured.Unstructured{*crd},
	}

	layout := "Mon Jan 2 15:04:05 -0700 MST 2006"
	ts, _ := time.Parse(layout, "Wed Dec 1 13:01:05 -0700 MST 2021")

	evt1 := eventGK("foo", "k1", "ns1", "FailedAdmission", "abc", ts)
	evt2 := eventGK("bar", "k2", "ns2", "FailedAdmission", "xyz", ts)
	violation1 := `- name: foo
  kind: k1
  namespace: ns1
  policy: ""
  constraint: ':'
  message: abc
  action: ""
  timestamp: "2021-12-01T20:01:05Z"`
	violation2 := `- name: bar
  kind: k2
  namespace: ns2
  policy: ""
  constraint: ':'
  message: xyz
  action: ""
  timestamp: "2021-12-01T20:01:05Z"`

	var tests = []struct {
		desc       string
		namespace  string
		expected   string
		unexpected string
		objects    []runtime.Object
	}{
		{
			"no violations in any namespace",
			"",
			"name: No Violations\nviolations: []",
			violation1,
			[]runtime.Object{crdList},
		},
		{
			"no violations in given namespace",
			"ns0",
			"name: No Violations\nviolations: []",
			violation1,
			[]runtime.Object{crdList, evt1, evt2},
		},
		{
			"violations in given namespace 1",
			"ns1",
			violation1,
			violation2,
			[]runtime.Object{crdList, evt1, evt2},
		},
		{
			"violations in given namespace 2",
			"ns2",
			violation2,
			violation1,
			[]runtime.Object{crdList, evt1, evt2},
		},
		{
			"violations in any namespace",
			"",
			violation2,
			"name: No Violations\nviolations: []",
			[]runtime.Object{crdList, evt1, evt2},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			factory := bbTestUtil.GetFakeFactory()
			factory.ResetIOStream()
			factory.SetObjects(test.objects)
			factory.SetGVRToListKind(gvrToListKindForGatekeeper())
			v, _ := factory.GetViper()
			v.Set("big-bang-repo", "test")
			v.Set("output-config.format", "yaml")
			streams, _ := factory.GetIOStream()
			buf := streams.Out.(*bytes.Buffer)
			cmd := violationsCmd(factory, test.namespace, nil)
			err := cmd.Execute()
			if !strings.Contains(buf.String(), test.expected) {
				t.Errorf("unexpected output:\nOutput:\n%s\nExpected:\n%s", buf.String(), test.expected)
			}
			if strings.Contains(buf.String(), test.unexpected) && test.unexpected != "" {
				t.Errorf("unexpected output:\nOutput:\n%s\nUnexpected:\n%s", buf.String(), test.unexpected)
			}
			assert.Nil(t, err)
		})
	}
}

func TestKyvernoAuditViolations(t *testing.T) {
	crd1 := &unstructured.Unstructured{}
	crd1.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "apiextensions.k8s.io/v1",
		"kind":       "customresourcedefinition",
		"metadata": map[string]interface{}{
			"name": "foos.policies.kyverno.io",
			"labels": map[string]interface{}{
				"app.kubernetes.io/name": "kyverno",
			},
		},
		"spec": map[string]any{
			"group": "kyverno.io",
		},
	})

	crdList := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1",
			"kind":       "customresourcedefinitionList",
		},
		Items: []unstructured.Unstructured{*crd1},
	}

	layout := "Mon Jan 2 15:04:05 -0700 MST 2006"
	ts, _ := time.Parse(layout, "Wed Dec 1 13:01:05 -0700 MST 2021")

	evt1 := eventKyverno("foo", "k1", "ns1", "policy-controller", "FailedAdmission", ts)
	evt2 := eventKyverno("bar", "k2", "ns2", "policy-controller", "FailedAdmission", ts)

	violation1 := `- name: foo
  kind: k1
  namespace: ns1
  policy: ""
  constraint: ""
  message: FailedAdmission
  action: ""
  timestamp: "2021-12-01T20:01:05Z"`
	violation2 := `- name: bar
  kind: k2
  namespace: ns2
  policy: ""
  constraint: ""
  message: FailedAdmission
  action: ""
  timestamp: "2021-12-01T20:01:05Z"`

	var tests = []struct {
		desc       string
		namespace  string
		expected   string
		unexpected string
		objects    []runtime.Object
	}{
		{
			"no violations in any namespace",
			"",
			"name: No Violations\nviolations: []",
			violation1,
			[]runtime.Object{crdList},
		},
		{
			"no violations in given namespace",
			"ns0",
			"name: No Violations\nviolations: []",
			violation1,
			[]runtime.Object{crdList, evt1, evt2},
		},
		{
			"violations in given namespace 1",
			"ns1",
			violation1,
			violation2,
			[]runtime.Object{crdList, evt1, evt2},
		},
		{
			"violations in given namespace 2",
			"ns2",
			violation2,
			violation1,
			[]runtime.Object{crdList, evt1, evt2},
		},
		{
			"violations in any namespace",
			"",
			violation2,
			"name: No Violations\nviolations: []",
			[]runtime.Object{crdList, evt1, evt2},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			factory := bbTestUtil.GetFakeFactory()
			factory.ResetIOStream()
			factory.SetObjects(test.objects)
			factory.SetGVRToListKind(gvrToListKindForKyverno())
			v, _ := factory.GetViper()
			v.Set("big-bang-repo", "test")
			v.Set("output-config.format", "yaml")
			streams, _ := factory.GetIOStream()
			buf := streams.Out.(*bytes.Buffer)
			cmd := violationsCmd(factory, test.namespace, []string{"--audit"})
			err := cmd.Execute()
			if !strings.Contains(buf.String(), test.expected) {
				t.Errorf("unexpected output:\nOutput:\n%s\nExpected:\n%s", buf.String(), test.expected)
			}
			if strings.Contains(buf.String(), test.unexpected) && test.unexpected != "" {
				t.Errorf("unexpected output:\nOutput:\n%s\nUnexpected:\n%s", buf.String(), test.unexpected)
			}
			assert.Nil(t, err)
		})
	}
}

func TestKyvernoEnforceViolations(t *testing.T) {
	crd1 := &unstructured.Unstructured{}
	crd1.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "apiextensions.k8s.io/v1",
		"kind":       "customresourcedefinition",
		"metadata": map[string]interface{}{
			"name": "foos.policies.kyverno.io",
			"labels": map[string]interface{}{
				"app.kubernetes.io/name": "kyverno",
			},
		},
		"spec": map[string]any{
			"group": "kyverno.io",
		},
	})

	crdList := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1",
			"kind":       "customresourcedefinitionList",
		},
		Items: []unstructured.Unstructured{*crd1},
	}

	layout := "Mon Jan 2 15:04:05 -0700 MST 2006"
	ts, _ := time.Parse(layout, "Wed Dec 1 13:01:05 -0700 MST 2021")

	evt1 := eventKyverno("foo", "po", "ns1", "admission-controller", "FailedAdmission", ts)
	evt2 := eventKyverno("bar", "cp", "ns2", "admission-controller", "FailedAdmission", ts)

	violation1 := `- name: NA
  kind: po
  namespace: ns1
  policy: foo
  constraint: ""
  message: FailedAdmission
  action: ""
  timestamp: "2021-12-01T20:01:05Z"`
	violation2 := `- name: NA
  kind: cp
  namespace: ns2
  policy: bar
  constraint: ""
  message: FailedAdmission
  action: ""
  timestamp: "2021-12-01T20:01:05Z"`

	var tests = []struct {
		desc       string
		namespace  string
		expected   string
		unexpected string
		objects    []runtime.Object
	}{
		{
			"no violations in any namespace",
			"",
			"name: No Violations\nviolations: []",
			violation1,
			[]runtime.Object{crdList},
		},
		{
			"no violations in given namespace",
			"ns0",
			"name: No Violations\nviolations: []",
			violation1,
			[]runtime.Object{crdList, evt1, evt2},
		},
		{
			"violations in given namespace 1",
			"ns1",
			violation1,
			violation2,
			[]runtime.Object{crdList, evt1, evt2},
		},
		{
			"violations in given namespace 2",
			"ns2",
			violation2,
			violation1,
			[]runtime.Object{crdList, evt1, evt2},
		},
		{
			"violations in any namespace",
			"",
			violation2,
			"name: No Violations\nviolations: []",
			[]runtime.Object{crdList, evt1, evt2},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			factory := bbTestUtil.GetFakeFactory()
			factory.ResetIOStream()
			factory.SetObjects(test.objects)
			factory.SetGVRToListKind(gvrToListKindForKyverno())
			v, _ := factory.GetViper()
			v.Set("big-bang-repo", "test")
			v.Set("output-config.format", "yaml")
			streams, _ := factory.GetIOStream()
			buf := streams.Out.(*bytes.Buffer)
			cmd := violationsCmd(factory, test.namespace, nil)
			err := cmd.Execute()
			if !strings.Contains(buf.String(), test.expected) {
				t.Errorf("unexpected output:\nOutput:\n%s\nExpected:\n%s", buf.String(), test.expected)
			}
			if strings.Contains(buf.String(), test.unexpected) && test.unexpected != "" {
				t.Errorf("unexpected output:\nOutput:\n%s\nUnexpected:\n%s", buf.String(), test.unexpected)
			}
			assert.Nil(t, err)
		})
	}
}

func TestGetViolations(t *testing.T) {
	tests := []struct {
		desc                          string
		expected                      string
		out                           string
		errorCheckingGatekeeperExists bool
		errorCheckingKyvernoExists    bool
		errorCheckingViolations       bool
	}{
		{
			"no errors",
			"",
			"name: No Violations\nviolations: []\n",
			false,
			false,
			false,
		},
		{
			"error checking gatekeeper exists",
			"errors occurred while listing violations: [error getting gatekeeper crds: error in list crds]",
			"",
			true,
			false,
			false,
		},
		{
			"error checking kyverno exists",
			"errors occurred while listing violations: [error getting kyverno crds: error in list crds]",
			"",
			false,
			true,
			false,
		},

		{
			"errors listing both kyverno and gatekeeper",
			"errors occurred while listing violations: [error listing gatekeeper violations: error in list events error listing kyverno violations: error in list events]",
			"",
			false,
			false,
			true,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			// Arrange
			factory := bbTestUtil.GetFakeFactory()
			factory.ResetIOStream()
			v, _ := factory.GetViper()
			v.Set("big-bang-repo", "test")
			v.Set("output-config.format", "yaml")
			streams, _ := factory.GetIOStream()
			in := streams.In.(*bytes.Buffer)
			out := streams.Out.(*bytes.Buffer)
			cmd := violationsCmd(factory, "", nil)

			gvrs := map[runtimeSchema.GroupVersionResource]string{}
			for gvr, listKind := range gvrToListKindForGatekeeper() {
				gvrs[gvr] = listKind
			}
			for gvr, listKind := range gvrToListKindForKyverno() {
				gvrs[gvr] = listKind
			}
			factory.SetGVRToListKind(gvrs)

			// Fail to list CRDs only on the first call (checking for gatekeeper CRDs), succeed on subquent calls
			if test.errorCheckingGatekeeperExists {
				var runCount int
				modFunc := func(client *dynamicFake.FakeDynamicClient) {
					client.Fake.PrependReactor("list", "customresourcedefinitions", func(action k8sTesting.Action) (bool, runtime.Object, error) {
						runCount++
						if runCount == 1 {
							return true, nil, fmt.Errorf("error in list crds")
						}
						return false, nil, nil
					})
				}
				factory.SetFail.GetK8sDynamicClientPrepFuncs = append(factory.SetFail.GetK8sDynamicClientPrepFuncs, &modFunc)
			}

			// Fail to list CRDs only on the second call (checking for Kyverno CRDs), succeed on first and subquent calls
			if test.errorCheckingKyvernoExists {
				var runCount int
				modFunc := func(client *dynamicFake.FakeDynamicClient) {
					client.Fake.PrependReactor("list", "customresourcedefinitions", func(action k8sTesting.Action) (bool, runtime.Object, error) {
						runCount++
						if runCount == 2 {
							return true, nil, fmt.Errorf("error in list crds")
						}
						return false, nil, nil
					})
				}
				factory.SetFail.GetK8sDynamicClientPrepFuncs = append(factory.SetFail.GetK8sDynamicClientPrepFuncs, &modFunc)
			}

			// Fail to list CRDs on the third and beyond call (after the xExists calls have been made, when violations are checked)
			if test.errorCheckingViolations {
				runCount := 0
				modFunc := func(client *dynamicFake.FakeDynamicClient) {
					client.Fake.PrependReactor("list", "customresourcedefinitions", func(action k8sTesting.Action) (bool, runtime.Object, error) {
						runCount++

						// Error on the third and beyond calls
						if runCount >= 3 {
							return true, nil, fmt.Errorf("error in list crds")
						}

						// Return a valid list of CRDs on the first call
						crds := &unstructured.UnstructuredList{
							Object: map[string]interface{}{
								"kind":       "CustomResourceDefinitionList",
								"apiVersion": "apiextensions.k8s.io/v1",
							},
							Items: []unstructured.Unstructured{
								{
									Object: map[string]interface{}{
										"kind":       "CustomResourceDefinition",
										"apiVersion": "apiextensions.k8s.io/v1",
										"metadata": map[string]interface{}{
											"name": "example1.kyverno.io",
										},
										"spec": map[string]interface{}{
											"group": "kyverno.io",
										},
									},
								},
								{
									Object: map[string]interface{}{
										"kind":       "CustomResourceDefinition",
										"apiVersion": "apiextensions.k8s.io/v1",
										"metadata": map[string]interface{}{
											"name": "example2.gatekeeper.sh",
											"labels": map[string]interface{}{
												"app.kubernetes.io/name": "gatekeeper",
											},
										},
									},
								},
							},
						}
						return true, crds, nil
					})
				}
				factory.SetFail.GetK8sDynamicClientPrepFuncs = append(factory.SetFail.GetK8sDynamicClientPrepFuncs, &modFunc)

				clientSetModFunc := func(client *fake.Clientset) {
					client.CoreV1().Events("").(*typedFake.FakeEvents).Fake.PrependReactor("list", "events", func(action k8sTesting.Action) (bool, runtime.Object, error) {
						return true, nil, fmt.Errorf("error in list events")
					})
				}
				factory.SetFail.GetK8sClientsetPrepFuncs = append(factory.SetFail.GetK8sClientsetPrepFuncs, &clientSetModFunc)

			}

			tv, err := newViolationsCmdHelper(cmd, factory)
			assert.Nil(t, err)

			// if test.errorCheckingForKyverno or test.errorCheckingForGatekeeper is true,
			// we'll expect the fake kubernetes client to return an error when listing violations
			err = tv.getViolations()

			// Assert
			assert.Empty(t, in.String())
			if err != nil {
				assert.Equal(t, test.expected, err.Error())
			}
			assert.Equal(t, test.out, out.String())
		})
	}
}

func TestKyvernoExists(t *testing.T) {
	tests := []struct {
		desc                 string
		expected             string
		errorOnDynamicClient bool
		errorOnListCRDs      bool
	}{
		{
			"no errors",
			"",
			false,
			false,
		},
		{
			"error listing crds",
			"error getting kyverno crds: error in list crds",
			false,
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
			cmd := violationsCmd(factory, "", nil)
			factory.SetFail.GetK8sDynamicClient = test.errorOnDynamicClient
			factory.SetGVRToListKind(gvrToListKindForKyverno())
			v, _ := factory.GetViper()
			v.Set("big-bang-repo", "test")
			v.Set("output-config.format", "yaml")

			if test.errorOnListCRDs {
				modFunc := func(client *dynamicFake.FakeDynamicClient) {
					client.Fake.PrependReactor("list", "customresourcedefinitions", func(action k8sTesting.Action) (bool, runtime.Object, error) {
						return true, nil, fmt.Errorf("error in list crds")
					})
				}
				factory.SetFail.GetK8sDynamicClientPrepFuncs = append(factory.SetFail.GetK8sDynamicClientPrepFuncs, &modFunc)
			}

			tv, err := newViolationsCmdHelper(cmd, factory)
			assert.Nil(t, err)
			assert.NotNil(t, tv)

			// Act
			exists, err := tv.kyvernoExists()

			// Assert
			if test.errorOnListCRDs {
				assert.NotNil(t, err)
				assert.Equal(t, test.expected, err.Error())
				assert.False(t, exists)
				return
			}

			assert.Empty(t, in.String())
			assert.False(t, exists)
			assert.Nil(t, err)
			assert.Equal(t, test.expected, out.String())
		})
	}
}

func TestListKyvernoViolations(t *testing.T) {
	violation1 := `- name: NA
  kind: k1
  namespace: ns1
  policy: foo
  constraint: ""
  message: FailedAdmission
  action: ""`
	violation2 := `- name: bar
  kind: k2
  namespace: ns1
  policy: ""
  constraint: ""
  message: FailedAudit
  action: ""`

	tests := []struct {
		desc              string
		expected          string
		errorOnListEvents bool
		noViolations      bool
		auditViolations   bool
	}{
		{
			"no violations",
			"",
			false,
			true,
			false,
		},
		{
			"admission violations",
			violation1,
			false,
			false,
			false,
		},
		{
			"audit violations",
			violation2,
			false,
			false,
			true,
		},
		{
			"error listing events",
			"error in list events",
			true,
			false,
			false,
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
			v, _ := factory.GetViper()
			v.Set("big-bang-repo", "test")
			v.Set("output-config.format", "yaml")
			cmd := violationsCmd(factory, "", nil)

			if test.errorOnListEvents {
				modFunc := func(client *fake.Clientset) {
					client.CoreV1().Events("").(*typedFake.FakeEvents).Fake.PrependReactor("list", "events", func(action k8sTesting.Action) (bool, runtime.Object, error) {
						return true, nil, fmt.Errorf("error in list events")
					})
				}
				factory.SetFail.GetK8sClientsetPrepFuncs = append(factory.SetFail.GetK8sClientsetPrepFuncs, &modFunc)
			}
			factory.SetGVRToListKind(gvrToListKindForKyverno())

			if !test.noViolations {
				eventList := &v1.EventList{
					Items: []v1.Event{
						*eventKyverno("foo", "k1", "ns1", "admission-controller", "FailedAdmission", time.Now()),
						*eventKyverno("bar", "k2", "ns1", "policy-controller", "FailedAudit", time.Now()),
					},
				}
				factory.SetObjects([]runtime.Object{eventList})
			}

			tv, err := newViolationsCmdHelper(cmd, factory)
			assert.Nil(t, err)

			// Act
			err = tv.listKyvernoViolations("ns1", test.auditViolations)
			printErr := tv.printViolation()

			// Assert
			assert.Nil(t, printErr)
			assert.Empty(t, in.String())
			if test.errorOnListEvents {
				assert.NotNil(t, err)
				assert.Equal(t, test.expected, err.Error())
				assert.Equal(t, out.String(), "name: No Violations\nviolations: []\n")
				return
			}

			assert.Nil(t, err)
			assert.Contains(t, out.String(), test.expected)
		})
	}
}

func TestGatekeeperExists(t *testing.T) {
	tests := []struct {
		desc            string
		expected        string
		errorOnListCRDs bool
	}{
		{
			"no errors",
			"",
			false,
		},
		{
			"error listing crds",
			"error getting gatekeeper crds: error in list crds",
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
			v, _ := factory.GetViper()
			v.Set("big-bang-repo", "test")
			v.Set("output-config.format", "yaml")
			cmd := violationsCmd(factory, "", nil)

			if test.errorOnListCRDs {
				modFunc := func(client *dynamicFake.FakeDynamicClient) {
					client.Fake.PrependReactor("list", "customresourcedefinitions", func(action k8sTesting.Action) (bool, runtime.Object, error) {
						return true, nil, fmt.Errorf("error in list crds")
					})
				}
				factory.SetFail.GetK8sDynamicClientPrepFuncs = append(factory.SetFail.GetK8sDynamicClientPrepFuncs, &modFunc)
			}
			factory.SetGVRToListKind(gvrToListKindForGatekeeper())

			tv, err := newViolationsCmdHelper(cmd, factory)
			assert.Nil(t, err)

			// Act
			exists, err := tv.gatekeeperExists()

			// Assert
			assert.Empty(t, in.String())
			assert.False(t, exists)
			if test.errorOnListCRDs {
				assert.NotNil(t, err)
				assert.Equal(t, test.expected, err.Error())
				assert.Empty(t, out.String())
				return
			}

			assert.Nil(t, err)
			assert.Equal(t, test.expected, out.String())
		})
	}
}

// listGkViolations tested in previous tests

func TestListGkDenyViolations(t *testing.T) {
	violation1 := `- name: foo
  kind: k1
  namespace: ns1
  policy: ""
  constraint: ':'
  message: abc
  action: ""`
	tests := []struct {
		desc              string
		expected          string
		errorOnListEvents bool
		noViolations      bool
	}{
		{
			"no violations",
			"name: No Violations\nviolations: []\n",
			false,
			true,
		},
		{
			"deny violations",
			violation1,
			false,
			false,
		},
		{
			"error listing events",
			"error in list events",
			true,
			false,
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
			v, _ := factory.GetViper()
			v.Set("big-bang-repo", "test")
			v.Set("output-config.format", "yaml")
			cmd := violationsCmd(factory, "", nil)

			if test.errorOnListEvents {
				modFunc := func(client *fake.Clientset) {
					client.CoreV1().Events("").(*typedFake.FakeEvents).Fake.PrependReactor("list", "events", func(action k8sTesting.Action) (bool, runtime.Object, error) {
						return true, nil, fmt.Errorf("error in list events")
					})
				}
				factory.SetFail.GetK8sClientsetPrepFuncs = append(factory.SetFail.GetK8sClientsetPrepFuncs, &modFunc)
			}
			if !test.noViolations {
				eventList := &v1.EventList{
					Items: []v1.Event{
						*eventGK("foo", "k1", "ns1", "FailedAdmission", "abc", time.Now()),
					},
				}
				factory.SetObjects([]runtime.Object{eventList})
			}

			tv, err := newViolationsCmdHelper(cmd, factory)
			assert.Nil(t, err)

			// Act
			err = tv.listGkDenyViolations("ns1")
			printErr := tv.printViolation()

			// Assert
			assert.Nil(t, printErr)
			assert.Empty(t, in.String())
			if test.errorOnListEvents {
				assert.NotNil(t, err)
				assert.Equal(t, test.expected, err.Error())
				assert.Equal(t, out.String(), "name: No Violations\nviolations: []\n")
				return
			}
			assert.Nil(t, err)
			assert.Contains(t, out.String(), test.expected)
		})
	}
}

func TestListGkAuditViolations(t *testing.T) {
	violation1 := `- name: foo
  kind: k1
  namespace: ns1
  policy: ""
  constraint: '%!s(<nil>)'
  message: FailedAdmission
  action: deny`
	tests := []struct {
		desc                   string
		expected               string
		errorOnListCrds        bool
		noViolations           bool
		errorOnListConstraints bool
	}{
		{
			"no violations",
			"name: No Violations\nviolations: []\n",
			false,
			true,
			false,
		},
		{
			"audit violations",
			violation1,
			false,
			false,
			false,
		},
		{
			"error listing crds",
			"error getting gatekeeper crds: error in list events",
			true,
			false,
			false,
		},
		{
			"error listing constraints",
			"error getting gatekeeper constraint: error in list constraints",
			false,
			false,
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
			v, _ := factory.GetViper()
			v.Set("big-bang-repo", "test")
			v.Set("output-config.format", "yaml")
			cmd := violationsCmd(factory, "", nil)
			factory.SetGVRToListKind(gvrToListKindForGatekeeper())

			if test.errorOnListCrds {
				modFunc := func(client *dynamicFake.FakeDynamicClient) {
					client.Fake.PrependReactor("list", "customresourcedefinitions", func(action k8sTesting.Action) (bool, runtime.Object, error) {
						return true, nil, fmt.Errorf("error in list events")
					})
				}
				factory.SetFail.GetK8sDynamicClientPrepFuncs = append(factory.SetFail.GetK8sDynamicClientPrepFuncs, &modFunc)
			} else if !test.noViolations {
				var objects []runtime.Object
				crdList := unstructured.UnstructuredList{
					Object: map[string]interface{}{
						"apiVersion": "apiextensions.k8s.io/v1",
						"kind":       "customresourcedefinitionList",
					},
					Items: []unstructured.Unstructured{
						{
							Object: map[string]interface{}{
								"apiVersion": "apiextensions.k8s.io/v1",
								"kind":       "customresourcedefinition",
								"metadata": map[string]interface{}{
									"name": "foos.constraints.gatekeeper.sh",
									"labels": map[string]interface{}{
										"app.kubernetes.io/name": "gatekeeper",
									},
								},
							},
						},
					},
				}
				objects = append(objects, &crdList)
				if !test.errorOnListConstraints {
					constraints := unstructured.UnstructuredList{
						Object: map[string]interface{}{
							"apiVersion": "constraints.gatekeeper.sh/v1beta1",
							"kind":       "gkPolicyList",
						},
						Items: []unstructured.Unstructured{
							{
								Object: map[string]interface{}{
									"apiVersion": "constraints.gatekeeper.sh/v1beta1",
									"kind":       "foo",
									"metadata": map[string]interface{}{
										"name": "foo-1",
										"labels": map[string]interface{}{
											"app.kubernetes.io/name": "gatekeeper",
										},
									},
									"status": map[string]interface{}{
										"auditTimestamp": "2021-11-27T23:55:33Z",
										"violations": []interface{}{
											map[string]interface{}{"kind": "k1", "name": "foo", "namespace": "ns1", "message": "FailedAdmission", "enforcementAction": "deny"},
										},
									},
								},
							},
						},
					}
					objects = append(objects, &constraints)
				} else {
					modFunc := func(client *dynamicFake.FakeDynamicClient) {
						client.Fake.PrependReactor("list", "foos", func(action k8sTesting.Action) (bool, runtime.Object, error) {
							return true, nil, fmt.Errorf("error in list constraints")
						})
					}
					factory.SetFail.GetK8sDynamicClientPrepFuncs = append(factory.SetFail.GetK8sDynamicClientPrepFuncs, &modFunc)
				}
				factory.SetObjects(objects)
			}

			tv, err := newViolationsCmdHelper(cmd, factory)
			assert.Nil(t, err)

			// Act
			err = tv.listGkAuditViolations("ns1")
			printErr := tv.printViolation()

			// Assert
			assert.Nil(t, printErr)
			assert.Empty(t, in.String())
			if test.errorOnListCrds || test.errorOnListConstraints {
				assert.NotNil(t, err)
				assert.Equal(t, test.expected, err.Error())
				assert.Equal(t, out.String(), "name: No Violations\nviolations: []\n")
				return
			}
			assert.Nil(t, err)
			assert.Contains(t, out.String(), test.expected)
		})
	}
}

func TestGetGkConstraintViolations(t *testing.T) {
	violation1 := `- name: foo
  kind: k1
  namespace: ns1
  policy: ""
  constraint: ""
  message: FailedAdmission
  action: deny`
	tests := []struct {
		desc                   string
		expected               string
		errorNestedFieldNoCopy bool
		errorOnNestedSlice     bool
	}{
		{
			"no violations",
			"name: No Violations\nviolations: []\n",
			false,
			false,
		},
		{
			"violations",
			violation1,
			false,
			false,
		},
		{
			"error getting nested field",
			".status.auditTimestamp accessor error: 4 is of the type int, expected map[string]interface{}",
			true,
			false,
		},
		{
			"error on nested slice",
			".status.violations accessor error: 4 is of the type int, expected []interface{}",
			false,
			true,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			// Arrange
			constraint := unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "constraints.gatekeeper.sh/v1beta1",
					"kind":       "foo",
					"metadata": map[string]interface{}{
						"name": "foo-1",
						"labels": map[string]interface{}{
							"app.kubernetes.io/name": "gatekeeper",
						},
					},
					"status": map[string]interface{}{
						"auditTimestamp": "2021-11-27T23:55:33Z",
						"violations": []interface{}{
							map[string]interface{}{"kind": "k1", "name": "foo", "namespace": "ns1", "message": "FailedAdmission", "enforcementAction": "deny"},
						},
					},
				},
			}
			if test.errorNestedFieldNoCopy {
				constraint.Object["status"] = 4
			}
			if test.errorOnNestedSlice {
				constraint.Object["status"].(map[string]interface{})["violations"] = 4
			}

			// Act
			violations, err := getGkConstraintViolations(&constraint)

			// Assert
			if test.errorNestedFieldNoCopy || test.errorOnNestedSlice {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), test.expected)
				assert.Nil(t, violations)
				return
			}
			assert.Nil(t, err)
			assert.NotNil(t, violations)
			assert.Equal(t, 1, len(*violations))
			violation := (*violations)[0]
			assert.Equal(t, "foo", violation.name)
			assert.Equal(t, "ns1", violation.namespace)
			assert.Equal(t, "k1", violation.kind)
			assert.Equal(t, "deny", violation.action)
			assert.Equal(t, "FailedAdmission", violation.message)
			assert.Equal(t, "2021-11-27T23:55:33Z", violation.timestamp)
		})
	}
}

func TestNewViolationsCmdHelper(t *testing.T) {
	tests := []struct {
		desc                    string
		expected                string
		errorOnK8sDynamicClient bool
		errorOnK8sClientSet     bool
		errorOnConfigClient     int
	}{
		{
			desc:                    "no errors",
			expected:                "",
			errorOnK8sDynamicClient: false,
			errorOnK8sClientSet:     false,
			errorOnConfigClient:     0,
		},
		{
			desc:                    "error getting k8s dynamic client",
			expected:                "failed to get K8sDynamicClient client",
			errorOnK8sDynamicClient: true,
			errorOnK8sClientSet:     false,
			errorOnConfigClient:     0,
		},
		{
			desc:                    "error getting k8s clientset",
			expected:                "testing error",
			errorOnK8sDynamicClient: false,
			errorOnK8sClientSet:     true,
			errorOnConfigClient:     0,
		},
		{
			desc:                    "error getting config client",
			expected:                "failed to get config client",
			errorOnK8sDynamicClient: false,
			errorOnK8sClientSet:     false,
			errorOnConfigClient:     1,
		},
	}

	for _, test := range tests {

		t.Run(test.desc, func(t *testing.T) {
			factory := bbTestUtil.GetFakeFactory()
			factory.ResetIOStream()
			streams, _ := factory.GetIOStream()
			out := streams.Out.(*bytes.Buffer)
			v, _ := factory.GetViper()
			v.Set("big-bang-repo", "test")
			v.Set("output-config.format", "yaml")
			cmd := violationsCmd(factory, "", nil)
			factory.SetFail.GetK8sDynamicClient = test.errorOnK8sDynamicClient
			factory.SetFail.GetK8sClientset = test.errorOnK8sClientSet
			factory.SetFail.GetConfigClient = test.errorOnConfigClient
			factory.SetGVRToListKind(gvrToListKindForKyverno())

			tv, err := newViolationsCmdHelper(cmd, factory)

			// This is the only test case where we don't expect an error
			if test.desc == "no errors" {
				assert.NotNil(t, tv)
				assert.Nil(t, err)
				return
			}

			// We should return an empty pointer if we fail any setup instructions
			assert.Nil(t, tv)
			assert.Empty(t, out.String())

			// We expect an error
			assert.NotNil(t, err)

			// Assert that the error is as expected for the given failed client
			assert.Equal(t, test.expected, err.Error())

		})
	}

}

// processGkViolations tested in previous tests
// printViolations tested in previous tests

func TestViolationsOutputClientError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.SetFail.GetIOStreams = 1
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "test")
	v.Set("output-config.format", "")

	// Act
	cmd, _ := NewViolationsCmd(factory)
	cmdHelper, err := newViolationsCmdHelper(cmd, factory)

	// Assert
	assert.Nil(t, cmdHelper)
	expectedError := "error getting output client: failed to get streams"
	if err == nil || err.Error() != expectedError {
		t.Errorf("Expected error: %s, got: %v", expectedError, err)
	}
}

func TestViolationsErrorBindingFlags(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "test")

	expectedError := fmt.Errorf("failed to set and bind flag")
	setAndBindFlagFunc := func(client *bbConfig.ConfigClient, name string, shorthand string, value interface{}, description string) error {
		if name == "audit" {
			return expectedError
		}
		return nil
	}

	logClient, _ := factory.GetLoggingClient()
	configClient, err := bbConfig.NewClient(nil, setAndBindFlagFunc, &logClient, nil, v)
	assert.Nil(t, err)
	factory.SetConfigClient(configClient)

	// Act
	cmd, err := NewViolationsCmd(factory)

	// Assert
	assert.Nil(t, cmd)
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Sprintf("error setting and binding flags: %s", expectedError.Error()), err.Error())
}

func TestViolationsPrintViolationsErr(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "test")
	v.Set("output-config.format", "")
	cmd := violationsCmd(factory, "", nil)
	eventList := &v1.EventList{
		Items: []v1.Event{
			*eventKyverno("foo", "k1", "ns1", "admission-controller", "FailedAdmission", time.Now()),
			*eventKyverno("bar", "k2", "ns1", "policy-controller", "FailedAudit", time.Now()),
		},
	}
	factory.SetObjects([]runtime.Object{eventList})

	tv, err := newViolationsCmdHelper(cmd, factory)
	assert.Nil(t, err)

	// Act
	err = tv.listKyvernoViolations("ns1", true)
	printErr := tv.printViolation()

	// Assert
	assert.Nil(t, err)
	assert.Nil(t, printErr)
}

func TestNoViolationsPrintViolationsErr(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "test")
	v.Set("output-config.format", "yaml")
	cmd := violationsCmd(factory, "", nil)

	tv, err := newViolationsCmdHelper(cmd, factory)
	assert.Nil(t, err)

	// Act
	printErr := tv.printViolation()

	// Assert
	assert.Nil(t, printErr)
}
