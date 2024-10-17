package gatekeeper

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

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
	}
}

func crdList() *unstructured.UnstructuredList {
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

	return crdList
}

func constraintList() *unstructured.UnstructuredList {
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

	return constraintList
}

func TestFetchGatekeeperCrds(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	factory.SetObjects([]runtime.Object{crdList()})
	factory.SetGVRToListKind(gvrToListKind())
	client, _ := factory.GetK8sDynamicClient(nil)
	crds, _ := FetchGatekeeperCrds(client)

	assert.Equal(t, "foos.constraints.gatekeeper.sh", crds.Items[0].GetName())
}

func TestFetchGatekeeperConstraints(t *testing.T) {
	var tests = []struct {
		desc     string
		arg      string
		expected []string
		objects  []runtime.Object
	}{
		{
			"no constraints exist",
			"foos.constraints.gatekeeper.sh",
			[]string{},
			[]runtime.Object{crdList()},
		},
		{
			"constraints exist",
			"foos.constraints.gatekeeper.sh",
			[]string{"foo-1", "foo-2"},
			[]runtime.Object{crdList(), constraintList()},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			factory := bbTestUtil.GetFakeFactory()
			factory.SetObjects(test.objects)
			factory.SetGVRToListKind(gvrToListKind())
			client, _ := factory.GetK8sDynamicClient(nil)
			constraints, _ := FetchGatekeeperConstraints(client, test.arg)
			for i, constraint := range constraints.Items {
				assert.Equal(t, test.expected[i], constraint.GetName())
			}
		})
	}
}

func TestFetchGatekeeperConstraintsError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.SetObjects([]runtime.Object{crdList()})
	factory.SetGVRToListKind(gvrToListKind())
	client := bbTestUtil.GetBadClient()

	// Act
	result, err := FetchGatekeeperConstraints(client, "nop.constraints.gatekeeper.sh")

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error getting gatekeeper constraint")
	assert.Nil(t, result)
}

func TestFetchGatekeeperCrdsError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.SetObjects([]runtime.Object{constraintList()})
	factory.SetGVRToListKind(gvrToListKind())
	client := bbTestUtil.GetBadClient()
	client.FailCrd = true

	// Act
	result, err := FetchGatekeeperCrds(client)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error getting gatekeeper crds")
	assert.Nil(t, result)
}
