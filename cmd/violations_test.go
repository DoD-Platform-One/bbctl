package cmd

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	bbutil "repo1.dso.mil/platform-one/big-bang/apps/product-tools/bbctl/util"
	bbtestutil "repo1.dso.mil/platform-one/big-bang/apps/product-tools/bbctl/util/test"
)

func event(rName string, rKind string, ns string, reason string, msg string, time time.Time) *v1.Event {

	annotations := make(map[string]string)
	annotations["resource_name"] = rName
	annotations["resource_kind"] = rKind
	annotations["resource_namespace"] = ns

	evt := &v1.Event{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:              rName,
			Annotations:       annotations,
			CreationTimestamp: meta_v1.Time{Time: time},
		},
		Reason:  reason,
		Message: msg,
	}

	return evt
}

func violationsCmd(factory bbutil.Factory, streams genericclioptions.IOStreams, ns string, args []string) *cobra.Command {
	cmd := NewViolationsCmd(factory, streams)
	cmd.Flags().StringP("namespace", "n", "", "namespace")
	cmdArgs := []string{}
	if ns != "" {
		cmdArgs = []string{"--namespace", ns}
	}
	cmdArgs = append(cmdArgs, args...)
	cmd.SetArgs(cmdArgs)
	return cmd
}

func gvrToListKind() map[schema.GroupVersionResource]string {
	return map[schema.GroupVersionResource]string{
		{
			Group:    "apiextensions.k8s.io",
			Version:  "v1",
			Resource: "customresourcedefinitions",
		}: "customresourcedefinitionsList",
		{
			Group:    "constraints.gatekeeper.sh",
			Version:  "v1beta1",
			Resource: "foos",
		}: "fooList",
		{
			Group:    "constraints.gatekeeper.sh",
			Version:  "v1beta1",
			Resource: "foo",
		}: "fooList",
		{
			Group:    "constraints.gatekeeper.sh",
			Version:  "v1beta1",
			Resource: "bar",
		}: "fooList",
		{
			Group:    "constraints.gatekeeper.sh",
			Version:  "v1beta1",
			Resource: "nop",
		}: "fooList",
		{
			Group:    "constraints.gatekeeper.sh",
			Version:  "v1beta1",
			Resource: "fudges",
		}: "fudgeList",
	}
}

func TestAuditViolations(t *testing.T) {

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
			"name": "foo",
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
			"kind":       "fooList",
		},
		Items: []unstructured.Unstructured{*constraint},
	}

	var tests = []struct {
		desc       string
		namespace  string
		expected   string
		unexpected string
		objs       []runtime.Object
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
			"No violations found in audit",
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
			factory := bbtestutil.GetFakeFactory(nil, test.objs, gvrToListKind())
			streams, _, buf, _ := genericclioptions.NewTestIOStreams()
			cmd := violationsCmd(factory, streams, test.namespace, []string{"--audit"})
			cmd.Execute()
			if !strings.Contains(buf.String(), test.expected) {
				t.Errorf("unexpected output: %s", buf.String())
			}
			if strings.Contains(buf.String(), test.unexpected) {
				t.Errorf("unexpected output: %s", buf.String())
			}
		})
	}
}

func TestDenyViolations(t *testing.T) {

	layout := "Mon Jan 2 15:04:05 -0700 MST 2006"
	ts, _ := time.Parse(layout, "Wed Dec 1 13:01:05 -0700 MST 2021")

	evt1 := event("foo", "k1", "ns1", "FailedAdmission", "abc", ts)
	evt2 := event("bar", "k2", "ns2", "FailedAdmission", "xyz", ts)

	var tests = []struct {
		desc       string
		namespace  string
		expected   string
		unexpected string
		objs       []runtime.Object
	}{
		{
			"no violations in any namespace",
			"",
			"No violation events found",
			"Resource: foo, Kind: k1, Namespace: ns1",
			[]runtime.Object{},
		},
		{
			"no violations in given namespace",
			"ns0",
			"No violation events found",
			"Resource: foo, Kind: k1, Namespace: ns1",
			[]runtime.Object{evt1, evt2},
		},
		{
			"violations in given namespace",
			"ns1",
			"Resource: foo, Kind: k1, Namespace: ns1",
			"Resource: bar, Kind: k2, Namespace: ns2",
			[]runtime.Object{evt1, evt2},
		},
		{
			"violations in any namespace",
			"",
			"Resource: bar, Kind: k2, Namespace: ns2",
			"No violation events found",
			[]runtime.Object{evt1, evt2},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			factory := bbtestutil.GetFakeFactory(nil, test.objs, nil)
			streams, _, buf, _ := genericclioptions.NewTestIOStreams()
			cmd := violationsCmd(factory, streams, test.namespace, nil)
			cmd.Execute()
			if !strings.Contains(buf.String(), test.expected) {
				t.Errorf("unexpected output: %s", buf.String())
			}
			if strings.Contains(buf.String(), test.unexpected) {
				t.Errorf("unexpected output: %s", buf.String())
			}
		})
	}

}

func policiesCmd(factory bbutil.Factory, streams genericclioptions.IOStreams, args []string) *cobra.Command {
	cmd := NewPoliciesCmd(factory, streams)
	cmd.SetArgs(args)
	return cmd
}

func TestPolicies(t *testing.T) {

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

	constraint1 := &unstructured.Unstructured{}
	constraint1.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "constraints.gatekeeper.sh/v1beta1",
		"kind":       "foo",
		"metadata": map[string]interface{}{
			"name": "foo",
			"labels": map[string]interface{}{
				"app.kubernetes.io/name": "gatekeeper",
			},
			"annotations": map[string]interface{}{
				"constraints.gatekeeper/description": "invalid config",
			},
		},
		"spec": map[string]interface{}{
			"enforcementAction": "deny",
		},
	})

	constraint2 := &unstructured.Unstructured{}
	constraint2.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "constraints.gatekeeper.sh/v1beta1",
		"kind":       "foo",
		"metadata": map[string]interface{}{
			"name": "bar",
			"labels": map[string]interface{}{
				"app.kubernetes.io/name": "gatekeeper",
			},
			"annotations": map[string]interface{}{
				"constraints.gatekeeper/description": "invalid config",
			},
		},
		"spec": map[string]interface{}{
			"enforcementAction": "dryrun",
		},
	})

	constraintList := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"apiVersion": "constraints.gatekeeper.sh/v1beta1",
			"kind":       "fooList",
		},
		Items: []unstructured.Unstructured{*constraint1, *constraint2},
	}

	var tests = []struct {
		desc     string
		args     []string
		expected []string
		objs     []runtime.Object
	}{
		{
			"list all policies",
			[]string{},
			[]string{"foos.constraints.gatekeeper.sh", "deny", "invalid config"},
			[]runtime.Object{crdList, constraintList},
		},
		{
			"list policy with given name",
			[]string{"foos.constraints.gatekeeper.sh"},
			[]string{"foos.constraints.gatekeeper.sh", "foo", "bar", "deny", "dry", "invalid config"},
			[]runtime.Object{crdList, constraintList},
		},
		{
			"list non existent policy",
			[]string{"nop"},
			[]string{"No constraints found"},
			[]runtime.Object{},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			factory := bbtestutil.GetFakeFactory(nil, test.objs, gvrToListKind())
			streams, _, buf, _ := genericclioptions.NewTestIOStreams()
			cmd := policiesCmd(factory, streams, test.args)
			cmd.Execute()
			for _, exp := range test.expected {
				if !strings.Contains(buf.String(), exp) {
					t.Errorf("unexpected output: %s", buf.String())
				}
			}
		})
	}
}

func TestPoliciesCompletion(t *testing.T) {
	crd1 := &unstructured.Unstructured{}
	crd1.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "apiextensions.k8s.io/v1",
		"kind":       "customresourcedefinition",
		"metadata": map[string]interface{}{
			"name": "foos.constraints.gatekeeper.sh",
			"labels": map[string]interface{}{
				"app.kubernetes.io/name": "gatekeeper",
			},
		},
	})

	crd2 := &unstructured.Unstructured{}
	crd2.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "apiextensions.k8s.io/v1",
		"kind":       "customresourcedefinition",
		"metadata": map[string]interface{}{
			"name": "fudges.constraints.gatekeeper.sh",
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
		Items: []unstructured.Unstructured{*crd1, *crd2},
	}

	var tests = []struct {
		desc     string
		hint     string
		expected []string
		objs     []runtime.Object
	}{
		{
			"match all policies",
			"",
			[]string{"foos", "fudges"},
			[]runtime.Object{crdList},
		},
		{
			"match policies with given prefix",
			"f",
			[]string{"foos", "fudges"},
			[]runtime.Object{crdList},
		},
		{
			"match policy with given prefix",
			"fud",
			[]string{"fudges"},
			[]runtime.Object{crdList},
		},
		{
			"match no policy",
			"z",
			[]string{},
			[]runtime.Object{},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			factory := bbtestutil.GetFakeFactory(nil, test.objs, gvrToListKind())
			streams, _, _, _ := genericclioptions.NewTestIOStreams()
			cmd := NewPoliciesCmd(factory, streams)
			suggestions, _ := cmd.ValidArgsFunction(cmd, []string{}, test.hint)
			if !reflect.DeepEqual(test.expected, suggestions) {
				t.Fatalf("expected: %v, got: %v", test.expected, suggestions)
			}
		})
	}

}
