package gatekeeper

import (
	"context"
	"fmt"
	"testing"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
)

func gvrToListKind(t *testing.T) map[schema.GroupVersionResource]string {
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

func crdList(t *testing.T) *unstructured.UnstructuredList {
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

func constraintList(t *testing.T) *unstructured.UnstructuredList {
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

type badClient struct{}

func (b *badClient) Resource(resource schema.GroupVersionResource) dynamic.NamespaceableResourceInterface {
	return badResource{}
}

type badResource struct{}

func (b badResource) Namespace(name string) dynamic.ResourceInterface {
	return b
}

func (badResource) Apply(ctx context.Context, name string, obj *unstructured.Unstructured, options metaV1.ApplyOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, fmt.Errorf("Intentional error for testing, Apply() not implemented")
}

func (badResource) ApplyStatus(ctx context.Context, name string, obj *unstructured.Unstructured, options metaV1.ApplyOptions) (*unstructured.Unstructured, error) {
	return nil, fmt.Errorf("Intentional error for testing, ApplyStatus() not implemented")
}

func (badResource) Create(ctx context.Context, obj *unstructured.Unstructured, options metaV1.CreateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, fmt.Errorf("Intentional error for testing, Create() not implemented")
}

func (badResource) Delete(ctx context.Context, name string, options metaV1.DeleteOptions, subresources ...string) error {
	return fmt.Errorf("Intentional error for testing, Delete() not implemented")
}

func (badResource) DeleteCollection(ctx context.Context, options metaV1.DeleteOptions, listOptions metaV1.ListOptions) error {
	return fmt.Errorf("Intentional error for testing, DeleteCollection() not implemented")
}

func (badResource) Get(ctx context.Context, name string, options metaV1.GetOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, fmt.Errorf("Intentional error for testing, Get() not implemented")
}

func (badResource) List(ctx context.Context, opts metaV1.ListOptions) (*unstructured.UnstructuredList, error) {
	return nil, fmt.Errorf("Intentional error for testing, List() not implemented")
}

func (badResource) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, options metaV1.PatchOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, fmt.Errorf("Intentional error for testing, Patch() not implemented")
}

func (badResource) Update(ctx context.Context, obj *unstructured.Unstructured, options metaV1.UpdateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, fmt.Errorf("Intentional error for testing, Update() not implemented")
}

func (badResource) UpdateStatus(ctx context.Context, obj *unstructured.Unstructured, options metaV1.UpdateOptions) (*unstructured.Unstructured, error) {
	return nil, fmt.Errorf("Intentional error for testing, UpdateStatus() not implemented")
}

func (badResource) Watch(ctx context.Context, opts metaV1.ListOptions) (watch.Interface, error) {
	return nil, fmt.Errorf("Intentional error for testing, Watch() not implemented")
}
