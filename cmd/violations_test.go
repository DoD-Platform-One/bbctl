package cmd

import (
	"fmt"
	"strings"
	"testing"
	"time"

	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeSchema "k8s.io/apimachinery/pkg/runtime/schema"
	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"
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

func violationsCmd(factory bbUtil.Factory, streams genericIOOptions.IOStreams, ns string, args []string) *cobra.Command {
	cmd := NewViolationsCmd(factory, streams)
	cmd.PersistentFlags().StringP("namespace", "n", "", "namespace")
	cmdArgs := []string{}
	if ns != "" {
		cmdArgs = []string{"--namespace", ns}
	}
	cmdArgs = append(cmdArgs, args...)
	cmd.SetArgs(cmdArgs)
	return cmd
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
			"No violations found in audit",
			"Resource: r1, Kind: k1, Namespace: ns1",
			[]runtime.Object{crdList, constraintList},
		},
		{
			"no violations in any namespace",
			"",
			"",
			"Resource: r1, Kind: k1, Namespace: ns1",
			[]runtime.Object{},
		},

		{
			"violations in given namespace",
			"ns1",
			"Resource: r1, Kind: k1, Namespace: ns1",
			"Resource: r2, Kind: k2, Namespace: ns2",
			[]runtime.Object{crdList, constraintList},
		},
		{
			"violations in any namespace",
			"",
			"Resource: r2, Kind: k2, Namespace: ns2",
			"No violations found in audit",
			[]runtime.Object{crdList, constraintList},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			factory := bbTestUtil.GetFakeFactory()
			factory.SetObjects(test.objects)
			factory.SetGVRToListKind(gvrToListKindForGatekeeper())
			streams, _, buf, _ := genericIOOptions.NewTestIOStreams()
			factory.GetViper().Set("big-bang-repo", "test")
			cmd := violationsCmd(factory, streams, test.namespace, []string{"--audit"})
			err := cmd.Execute()
			if !strings.Contains(buf.String(), test.expected) {
				t.Errorf("unexpected output: %s", buf.String())
			}
			if strings.Contains(buf.String(), test.unexpected) {
				t.Errorf("unexpected output: %s", buf.String())
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
			"No events found for deny violations",
			"Resource: foo, Kind: k1, Namespace: ns1",
			[]runtime.Object{crdList},
		},
		{
			"no violations in given namespace",
			"ns0",
			"No events found for deny violations",
			"Resource: foo, Kind: k1, Namespace: ns1",
			[]runtime.Object{crdList, evt1, evt2},
		},
		{
			"violations in given namespace",
			"ns1",
			"Resource: foo, Kind: k1, Namespace: ns1",
			"Resource: bar, Kind: k2, Namespace: ns2",
			[]runtime.Object{crdList, evt1, evt2},
		},
		{
			"violations in any namespace",
			"",
			"Resource: bar, Kind: k2, Namespace: ns2",
			"No violation events found",
			[]runtime.Object{crdList, evt1, evt2},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			factory := bbTestUtil.GetFakeFactory()
			factory.SetObjects(test.objects)
			factory.SetGVRToListKind(gvrToListKindForGatekeeper())
			factory.GetViper().Set("big-bang-repo", "test")
			streams, _, buf, _ := genericIOOptions.NewTestIOStreams()
			cmd := violationsCmd(factory, streams, test.namespace, nil)
			err := cmd.Execute()
			if !strings.Contains(buf.String(), test.expected) {
				t.Errorf("unexpected output: %s", buf.String())
			}
			if strings.Contains(buf.String(), test.unexpected) {
				t.Errorf("unexpected output: %s", buf.String())
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
			"No events found for policy violations",
			"Resource: foo, Kind: k1, Namespace: ns1",
			[]runtime.Object{crdList},
		},
		{
			"no violations in given namespace",
			"ns0",
			"No events found for policy violations",
			"Resource: foo, Kind: k1, Namespace: ns1",
			[]runtime.Object{crdList, evt1, evt2},
		},
		{
			"violations in given namespace",
			"ns1",
			"Resource: foo, Kind: k1, Namespace: ns1",
			"Resource: bar, Kind: k2, Namespace: ns2",
			[]runtime.Object{crdList, evt1, evt2},
		},
		{
			"violations in any namespace",
			"",
			"Resource: bar, Kind: k2, Namespace: ns2",
			"No events found for policy violations",
			[]runtime.Object{crdList, evt1, evt2},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			factory := bbTestUtil.GetFakeFactory()
			factory.SetObjects(test.objects)
			factory.SetGVRToListKind(gvrToListKindForKyverno())
			factory.GetViper().Set("big-bang-repo", "test")
			streams, _, buf, _ := genericIOOptions.NewTestIOStreams()
			cmd := violationsCmd(factory, streams, test.namespace, []string{"--audit"})
			err := cmd.Execute()
			if !strings.Contains(buf.String(), test.expected) {
				t.Errorf("unexpected output: %s", buf.String())
			}
			if strings.Contains(buf.String(), test.unexpected) {
				t.Errorf("unexpected output: %s", buf.String())
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
			"No events found for policy violations",
			"Resource: NA, Kind: po, Namespace: ns1",
			[]runtime.Object{crdList},
		},
		{
			"no violations in given namespace",
			"ns0",
			"No events found for policy violations",
			"Resource: NA, Kind: po, Namespace: ns1",
			[]runtime.Object{crdList, evt1, evt2},
		},
		{
			"violations in given namespace",
			"ns1",
			"Resource: NA, Kind: po, Namespace: ns1",
			"Resource: NA, Kind: po, Namespace: ns2",
			[]runtime.Object{crdList, evt1, evt2},
		},
		{
			"violations in any namespace",
			"",
			"Resource: NA, Kind: cp, Namespace: ns2",
			"No events found for policy violations",
			[]runtime.Object{crdList, evt1, evt2},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			factory := bbTestUtil.GetFakeFactory()
			factory.SetObjects(test.objects)
			factory.SetGVRToListKind(gvrToListKindForKyverno())
			factory.GetViper().Set("big-bang-repo", "test")
			streams, _, buf, _ := genericIOOptions.NewTestIOStreams()
			cmd := violationsCmd(factory, streams, test.namespace, nil)
			err := cmd.Execute()
			if !strings.Contains(buf.String(), test.expected) {
				t.Errorf("unexpected output: %s", buf.String())
			}
			if strings.Contains(buf.String(), test.unexpected) {
				t.Errorf("unexpected output: %s", buf.String())
			}
			assert.Nil(t, err)
		})
	}
}

func TestGetViolations(t *testing.T) {
	tests := []struct {
		desc                       string
		expected                   string
		out                        string
		errout                     string
		errorGettingConfigClient   bool
		errorCheckingForGatekeeper bool
		errorCheckingForKyverno    bool
	}{
		{
			"no errors",
			"",
			"",
			"",
			false,
			false,
			false,
		},
		{
			"error getting config client",
			"failed to get config client",
			"",
			"",
			true,
			false,
			false,
		},
		{
			"error checking for gatekeeper",
			"failed to get k8s dynamic client",
			"",
			"",
			false,
			true,
			false,
		},
		{
			"error checking for kyverno",
			"error getting kyverno crds: error in list crds",
			"",
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
			factory.GetViper().Set("big-bang-repo", "test")
			streams, in, out, errout := genericIOOptions.NewTestIOStreams()
			cmd := violationsCmd(factory, streams, "", nil)
			factory.SetFail.GetConfigClient = test.errorGettingConfigClient
			gvrs := map[runtimeSchema.GroupVersionResource]string{}
			for gvr, listKind := range gvrToListKindForGatekeeper() {
				gvrs[gvr] = listKind
			}
			for gvr, listKind := range gvrToListKindForKyverno() {
				gvrs[gvr] = listKind
			}
			factory.SetGVRToListKind(gvrs)
			if test.errorCheckingForGatekeeper {
				factory.SetFail.GetK8sDynamicClient = true
			}
			if test.errorCheckingForKyverno {
				runCount := 0 // hack to only fail on the second call
				modFunc := func(client *dynamicFake.FakeDynamicClient) {
					client.Fake.PrependReactor("list", "customresourcedefinitions", func(action k8sTesting.Action) (bool, runtime.Object, error) {
						if runCount == 0 {
							runCount++
							return false, nil, nil
						}
						return true, nil, fmt.Errorf("error in list crds")
					})
				}
				factory.SetFail.GetK8sDynamicClientPrepFuncs = append(factory.SetFail.GetK8sDynamicClientPrepFuncs, &modFunc)
			}

			// Act
			err := getViolations(cmd, factory, streams)

			// Assert
			assert.Empty(t, in.String())
			assert.Equal(t, test.errout, errout.String())
			assert.Equal(t, test.out, out.String())
			if test.errorGettingConfigClient || test.errorCheckingForGatekeeper || test.errorCheckingForKyverno {
				assert.NotNil(t, err)
				assert.Equal(t, test.expected, err.Error())
			} else {
				assert.Nil(t, err)
			}
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
			"error getting dynamic client",
			"failed to get k8s dynamic client",
			true,
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
			streams, in, out, errout := genericIOOptions.NewTestIOStreams()
			cmd := violationsCmd(factory, streams, "", nil)
			factory.SetFail.GetK8sDynamicClient = test.errorOnDynamicClient
			if test.errorOnListCRDs {
				modFunc := func(client *dynamicFake.FakeDynamicClient) {
					client.Fake.PrependReactor("list", "customresourcedefinitions", func(action k8sTesting.Action) (bool, runtime.Object, error) {
						return true, nil, fmt.Errorf("error in list crds")
					})
				}
				factory.SetFail.GetK8sDynamicClientPrepFuncs = append(factory.SetFail.GetK8sDynamicClientPrepFuncs, &modFunc)
			}
			factory.SetGVRToListKind(gvrToListKindForKyverno())

			// Act
			exists, err := kyvernoExists(cmd, factory, streams)

			// Assert
			assert.Empty(t, in.String())
			assert.Empty(t, errout.String())
			assert.False(t, exists)
			if test.errorOnDynamicClient || test.errorOnListCRDs {
				assert.NotNil(t, err)
				assert.Equal(t, test.expected, err.Error())
				assert.Empty(t, out.String())
			} else {
				assert.Nil(t, err)
				assert.Equal(t, test.expected, out.String())
			}
		})
	}
}

func TestListKyvernoViolations(t *testing.T) {
	tests := []struct {
		desc              string
		expected          string
		errorGetClientset bool
		errorOnListEvents bool
		noViolations      bool
		auditViolations   bool
	}{
		{
			"no violations",
			"No events found for policy violations\n\n",
			false,
			false,
			true,
			false,
		},
		{
			"admission violations",
			"Resource: NA, Kind: k1, Namespace: ns1\nPolicy: foo\nFailedAdmission\n\n",
			false,
			false,
			false,
			false,
		},
		{
			"audit violations",
			"Resource: bar, Kind: k2, Namespace: ns1\nFailedAudit\n\n",
			false,
			false,
			false,
			true,
		},
		{
			"error getting clientset",
			"failed to get k8s clientset",
			true,
			false,
			false,
			false,
		},
		{
			"error listing events",
			"error in list events",
			false,
			true,
			false,
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			// Arrange
			factory := bbTestUtil.GetFakeFactory()
			streams, in, out, errout := genericIOOptions.NewTestIOStreams()
			cmd := violationsCmd(factory, streams, "", nil)
			factory.SetFail.GetK8sClientset = test.errorGetClientset
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

			// Act
			err := listKyvernoViolations(cmd, factory, streams, "ns1", test.auditViolations)

			// Assert
			assert.Empty(t, in.String())
			assert.Empty(t, errout.String())
			if test.errorGetClientset || test.errorOnListEvents {
				assert.NotNil(t, err)
				assert.Equal(t, test.expected, err.Error())
				assert.Empty(t, out.String())
			} else {
				assert.Nil(t, err)
				assert.Contains(t, out.String(), test.expected)
			}
		})
	}
}

func TestGatekeeperExists(t *testing.T) {
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
			"error getting dynamic client",
			"failed to get k8s dynamic client",
			true,
			false,
		},
		{
			"error listing crds",
			"error getting gatekeeper crds: error in list crds",
			false,
			true,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			// Arrange
			factory := bbTestUtil.GetFakeFactory()
			streams, in, out, errout := genericIOOptions.NewTestIOStreams()
			cmd := violationsCmd(factory, streams, "", nil)
			factory.SetFail.GetK8sDynamicClient = test.errorOnDynamicClient
			if test.errorOnListCRDs {
				modFunc := func(client *dynamicFake.FakeDynamicClient) {
					client.Fake.PrependReactor("list", "customresourcedefinitions", func(action k8sTesting.Action) (bool, runtime.Object, error) {
						return true, nil, fmt.Errorf("error in list crds")
					})
				}
				factory.SetFail.GetK8sDynamicClientPrepFuncs = append(factory.SetFail.GetK8sDynamicClientPrepFuncs, &modFunc)
			}
			factory.SetGVRToListKind(gvrToListKindForGatekeeper())

			// Act
			exists, err := gatekeeperExists(cmd, factory, streams)

			// Assert
			assert.Empty(t, in.String())
			assert.Empty(t, errout.String())
			assert.False(t, exists)
			if test.errorOnDynamicClient || test.errorOnListCRDs {
				assert.NotNil(t, err)
				assert.Equal(t, test.expected, err.Error())
				assert.Empty(t, out.String())
			} else {
				assert.Nil(t, err)
				assert.Equal(t, test.expected, out.String())
			}
		})
	}
}

// listGkViolations tested in previous tests

func TestListGkDenyViolations(t *testing.T) {
	tests := []struct {
		desc              string
		expected          string
		errorGetClientset bool
		errorOnListEvents bool
		noViolations      bool
	}{
		{
			"no violations",
			"No events found for deny violations\n\n",
			false,
			false,
			true,
		},
		{
			"deny violations",
			"Resource: foo, Kind: k1, Namespace: ns1\nConstraint: \nabc\n\n",
			false,
			false,
			false,
		},
		{
			"error getting clientset",
			"failed to get k8s clientset",
			true,
			false,
			false,
		},
		{
			"error listing events",
			"error in list events",
			false,
			true,
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			// Arrange
			factory := bbTestUtil.GetFakeFactory()
			streams, in, out, errout := genericIOOptions.NewTestIOStreams()
			cmd := violationsCmd(factory, streams, "", nil)
			factory.SetFail.GetK8sClientset = test.errorGetClientset
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

			// Act
			err := listGkDenyViolations(cmd, factory, streams, "ns1")

			// Assert
			assert.Empty(t, in.String())
			assert.Empty(t, errout.String())
			if test.errorGetClientset || test.errorOnListEvents {
				assert.NotNil(t, err)
				assert.Equal(t, test.expected, err.Error())
				assert.Empty(t, out.String())
			} else {
				assert.Nil(t, err)
				assert.Contains(t, out.String(), test.expected)
			}
		})
	}
}

func TestListGkAuditViolations(t *testing.T) {
	tests := []struct {
		desc                   string
		expected               string
		errorGetDynamicClient  bool
		errorOnListCrds        bool
		noViolations           bool
		errorOnListConstraints bool
	}{
		{
			"no violations",
			"No violations found in audit\n\n\n",
			false,
			false,
			true,
			false,
		},
		{
			"audit violations",
			"Resource: foo, Kind: k1, Namespace: ns1\nFailedAdmission\n\n",
			false,
			false,
			false,
			false,
		},
		{
			"error getting dynamic clientset",
			"failed to get k8s dynamic client",
			true,
			false,
			false,
			false,
		},
		{
			"error listing crds",
			"error getting gatekeeper crds: error in list events",
			false,
			true,
			false,
			false,
		},
		{
			"error listing constraints",
			"error getting gatekeeper constraint: error in list constraints",
			false,
			false,
			false,
			true,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			// Arrange
			factory := bbTestUtil.GetFakeFactory()
			streams, in, out, errout := genericIOOptions.NewTestIOStreams()
			cmd := violationsCmd(factory, streams, "", nil)
			factory.SetFail.GetK8sDynamicClient = test.errorGetDynamicClient
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

			// Act
			err := listGkAuditViolations(cmd, factory, streams, "ns1")

			// Assert
			assert.Empty(t, in.String())
			assert.Empty(t, errout.String())
			if test.errorGetDynamicClient || test.errorOnListCrds || test.errorOnListConstraints {
				assert.NotNil(t, err)
				assert.Equal(t, test.expected, err.Error())
				assert.Empty(t, out.String())
			} else {
				assert.Nil(t, err)
				assert.Contains(t, out.String(), test.expected)
			}
		})
	}
}

func TestGetGkConstraintViolations(t *testing.T) {
	tests := []struct {
		desc                   string
		expected               string
		errorNestedFieldNoCopy bool
		errorOnNestedSlice     bool
	}{
		{
			"no violations",
			"No violations found in audit\n\n\n",
			false,
			false,
		},
		{
			"violations",
			"Resource: foo, Kind: k1, Namespace: ns1\nFailedAdmission\n\n",
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
			} else {
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
			}
		})
	}
}

// processGkViolations tested in previous tests
// printViolations tested in previous tests
