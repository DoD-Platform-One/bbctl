package cmd

import (
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
