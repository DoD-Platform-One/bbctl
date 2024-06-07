package kyverno

import (
	"testing"

	"github.com/stretchr/testify/assert"

	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// gvrToListKind returns a list of schema.GroupVersionResource resources
// that map to a resource type to be used as test values for util tests
func gvrToListKind() map[schema.GroupVersionResource]string {
	return map[schema.GroupVersionResource]string{
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

func crdList() *unstructured.UnstructuredList {
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
		"spec": map[string]any{
			"group": "kyverno.io",
		},
	})

	crdList := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1",
			"kind":       "customresourcedefinitionList",
		},
		Items: []unstructured.Unstructured{*crd1, *crd2},
	}

	return crdList
}

func policyList() *unstructured.UnstructuredList {
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

	return policyList
}

func TestFetchKyvernoCrds(t *testing.T) {

	var tests = []struct {
		desc     string
		expected []string
		objects  []runtime.Object
	}{
		{
			"no crd exist",
			[]string{},
			[]runtime.Object{},
		},
		{
			"crds exist",
			[]string{"foos.policies.kyverno.io", "bars.policies.kyverno.io"},
			[]runtime.Object{crdList()},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			factory := bbTestUtil.GetFakeFactory()
			factory.SetObjects(test.objects)
			factory.SetGVRToListKind(gvrToListKind())
			client, _ := factory.GetK8sDynamicClient(nil)
			crds, _ := FetchKyvernoCrds(client)
			assert.Len(t, crds.Items, len(test.expected))
			for _, crd := range crds.Items {
				assert.Contains(t, test.expected, crd.GetName())
			}
		})
	}
}

func TestFetchKyvernoPolicies(t *testing.T) {
	var tests = []struct {
		desc     string
		arg      string
		expected []string
		objects  []runtime.Object
	}{
		{
			"no policies exist",
			"foos.policies.kyverno.io",
			[]string{},
			[]runtime.Object{crdList()},
		},
		{
			"policies exist",
			"foos.constraints.gatekeeper.sh",
			[]string{"foo-1"},
			[]runtime.Object{crdList(), policyList()},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			factory := bbTestUtil.GetFakeFactory()
			factory.SetObjects(test.objects)
			factory.SetGVRToListKind(gvrToListKind())
			client, _ := factory.GetK8sDynamicClient(nil)
			policies, _ := FetchKyvernoPolicies(client, test.arg)
			assert.Len(t, policies.Items, len(test.expected))
			for i, policy := range policies.Items {
				assert.Equal(t, test.expected[i], policy.GetName())
			}
		})
	}
}
