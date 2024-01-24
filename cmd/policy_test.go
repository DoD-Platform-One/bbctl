package cmd

import (
	"reflect"
	"strings"
	"testing"

	bbutil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
	bbtestutil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func policiesCmd(factory bbutil.Factory, streams genericclioptions.IOStreams, args []string) *cobra.Command {
	cmd := NewPoliciesCmd(factory, streams)
	cmd.SetArgs(args)
	return cmd
}

func gvrToListKindForPolicies() map[schema.GroupVersionResource]string {
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
		}: "gkPolicyList",
		{
			Group:    "constraints.gatekeeper.sh",
			Version:  "v1beta1",
			Resource: "foo",
		}: "gkPolicyList",
		{
			Group:    "constraints.gatekeeper.sh",
			Version:  "v1beta1",
			Resource: "nop",
		}: "gkPolicyList",
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
		{
			Group:    "kyverno.io",
			Version:  "v1",
			Resource: "bars",
		}: "kyvernoPolicyList",
		{
			Group:    "kyverno.io",
			Version:  "v1",
			Resource: "bar",
		}: "kyvernoPolicyList",
		{
			Group:    "kyverno.io",
			Version:  "v1",
			Resource: "nop",
		}: "kyvernoPolicyList",
	}
}

func TestGatekeeperPolicies(t *testing.T) {

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
			"name": "foo-1",
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
			"name": "foo-2",
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
			"kind":       "gkPolicyList",
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
			[]string{"--gatekeeper"},
			[]string{"foos.constraints.gatekeeper.sh", "deny", "invalid config"},
			[]runtime.Object{crdList, constraintList},
		},
		{
			"list policy with given name",
			[]string{"--gatekeeper", "foos.constraints.gatekeeper.sh"},
			[]string{"foos.constraints.gatekeeper.sh", "foo-1", "foo-2", "deny", "dry", "invalid config"},
			[]runtime.Object{crdList, constraintList},
		},
		{
			"list non existent policy",
			[]string{"--gatekeeper", "nop"},
			[]string{"No constraints found"},
			[]runtime.Object{},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			factory := bbtestutil.GetFakeFactory(nil, test.objs, gvrToListKindForPolicies(), nil)
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

func TestKyvernoPolicies(t *testing.T) {

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
	})

	crd2 := &unstructured.Unstructured{}
	crd2.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "apiextensions.k8s.io/v1",
		"kind":       "customresourcedefinition",
		"metadata": map[string]interface{}{
			"name": "bars.policies.kyverno.io",
			"labels": map[string]interface{}{
				"app.kubernetes.io/name": "kyverno",
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

	policy1 := &unstructured.Unstructured{}
	policy1.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "kyverno.io/v1",
		"kind":       "foo",
		"metadata": map[string]interface{}{
			"name": "foo-1",
			"labels": map[string]interface{}{
				"app.kubernetes.io/name": "kyverno",
			},
			"annotations": map[string]interface{}{
				"policies.kyverno.io/description": "invalid config",
			},
		},
		"spec": map[string]interface{}{
			"validationFailureAction": "enforce",
		},
	})

	policy2 := &unstructured.Unstructured{}
	policy2.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "kyverno.io/v1",
		"kind":       "bar",
		"metadata": map[string]interface{}{
			"name":      "bar-1",
			"namespace": "demo",
			"labels": map[string]interface{}{
				"app.kubernetes.io/name": "kyverno",
			},
			"annotations": map[string]interface{}{
				"policies.kyverno.io/description": "invalid config",
			},
		},
		"spec": map[string]interface{}{
			"validationFailureAction": "audit",
		},
	})

	policyList := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"apiVersion": "kyverno.io/v1",
			"kind":       "kyvernoPolicyList",
		},
		Items: []unstructured.Unstructured{*policy1, *policy2},
	}

	var tests = []struct {
		desc     string
		args     []string
		expected []string
		objs     []runtime.Object
	}{
		{
			"list all policies",
			[]string{"--kyverno"},
			[]string{"foos.policies.kyverno.io", "enforce", "invalid config"},
			[]runtime.Object{crdList, policyList},
		},
		{
			"list policy with given name",
			[]string{"--kyverno", "bar-1"},
			[]string{"bar", "bar-1", "demo", "audit", "invalid config"},
			[]runtime.Object{crdList, policyList},
		},
		{
			"list non existent policy",
			[]string{"--kyverno", "nop"},
			[]string{"No Matching Policy Found"},
			[]runtime.Object{},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			factory := bbtestutil.GetFakeFactory(nil, test.objs, gvrToListKindForPolicies(), nil)
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

func TestGatekeeperPoliciesCompletion(t *testing.T) {
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
			factory := bbtestutil.GetFakeFactory(nil, test.objs, gvrToListKindForPolicies(), nil)
			streams, _, _, _ := genericclioptions.NewTestIOStreams()
			cmd := NewPoliciesCmd(factory, streams)
			cmd.Flags().Lookup("gatekeeper").Value.Set("1")
			suggestions, _ := cmd.ValidArgsFunction(cmd, []string{}, test.hint)
			if !reflect.DeepEqual(test.expected, suggestions) {
				t.Fatalf("expected: %v, got: %v", test.expected, suggestions)
			}
		})
	}
}

func TestKyvernoPoliciesCompletion(t *testing.T) {

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
	})

	crdList := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1",
			"kind":       "customresourcedefinitionList",
		},
		Items: []unstructured.Unstructured{*crd1},
	}

	policy1 := &unstructured.Unstructured{}
	policy1.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "kyverno.io/v1",
		"kind":       "foo",
		"metadata": map[string]interface{}{
			"name": "fu-bar",
			"labels": map[string]interface{}{
				"app.kubernetes.io/name": "kyverno",
			},
			"annotations": map[string]interface{}{
				"policies.kyverno.io/description": "invalid config",
			},
		},
		"spec": map[string]interface{}{
			"validationFailureAction": "enforce",
		},
	})

	policy2 := &unstructured.Unstructured{}
	policy2.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "kyverno.io/v1",
		"kind":       "foo",
		"metadata": map[string]interface{}{
			"name":      "fudge-bar",
			"namespace": "demo",
			"labels": map[string]interface{}{
				"app.kubernetes.io/name": "kyverno",
			},
			"annotations": map[string]interface{}{
				"policies.kyverno.io/description": "invalid config",
			},
		},
		"spec": map[string]interface{}{
			"validationFailureAction": "audit",
		},
	})

	policyList := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"apiVersion": "kyverno.io/v1",
			"kind":       "kyvernoPolicyList",
		},
		Items: []unstructured.Unstructured{*policy1, *policy2},
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
			[]string{"fu-bar", "fudge-bar"},
			[]runtime.Object{crdList, policyList},
		},
		{
			"match policies with given prefix",
			"fu",
			[]string{"fu-bar", "fudge-bar"},
			[]runtime.Object{crdList, policyList},
		},
		{
			"match policy with given prefix",
			"fud",
			[]string{"fudge-bar"},
			[]runtime.Object{crdList, policyList},
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
			factory := bbtestutil.GetFakeFactory(nil, test.objs, gvrToListKindForPolicies(), nil)
			streams, _, _, _ := genericclioptions.NewTestIOStreams()
			cmd := NewPoliciesCmd(factory, streams)
			cmd.Flags().Lookup("kyverno").Value.Set("1")
			suggestions, _ := cmd.ValidArgsFunction(cmd, []string{}, test.hint)
			if !reflect.DeepEqual(test.expected, suggestions) {
				t.Fatalf("expected: %v, got: %v", test.expected, suggestions)
			}
		})
	}
}
