package test

import (
	"context"
	"fmt"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
)

type badClient struct {
	FailCrd        bool
	FailDescriptor bool
	Gatekeeper     bool
	DescriptorType string
}

func (b *badClient) Resource(resource schema.GroupVersionResource) dynamic.NamespaceableResourceInterface {
	mockResource := badResource{}
	if resource.Resource == "customresourcedefinitions" && !b.FailCrd {
		mockResource.MockCrds = true
		return mockResource
	}
	if b.FailDescriptor {
		mockResource.DescriptorType = b.DescriptorType
		if b.Gatekeeper {
			mockResource.MockConstraints = true
			return mockResource
		}
		mockResource.MockPolicies = true
		return mockResource
	}
	return mockResource
}

func GetBadClient() *badClient {
	client := &badClient{}
	return client
}

type badResource struct {
	MockCrds        bool
	MockConstraints bool
	MockPolicies    bool
	DescriptorType  string
}

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

func (b badResource) List(ctx context.Context, opts metaV1.ListOptions) (*unstructured.UnstructuredList, error) {
	if b.MockCrds {
		crdList := b.MockCrd()
		b.MockCrds = false
		return crdList, nil
	}
	if b.MockConstraints {
		var constraintList *unstructured.UnstructuredList
		switch {
		case b.DescriptorType == "kind":
			constraintList = b.MockConstraintKind()
		case b.DescriptorType == "name":
			constraintList = b.MockConstraintName()
		case b.DescriptorType == "desc":
			constraintList = b.MockConstraintDesc()
		case b.DescriptorType == "action":
			constraintList = b.MockConstraintAction()
		}
		return constraintList, nil
	}
	if b.MockPolicies {
		var policyList *unstructured.UnstructuredList
		switch {
		case b.DescriptorType == "kind":
			policyList = b.MockPolicyKind()
		case b.DescriptorType == "name":
			policyList = b.MockPolicyName()
		case b.DescriptorType == "namespace":
			policyList = b.MockPolicyNamespace()
		case b.DescriptorType == "desc":
			policyList = b.MockPolicyDesc()
		case b.DescriptorType == "action":
			policyList = b.MockPolicyAction()
		}
		return policyList, nil
	}
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

func (badResource) MockCrd() *unstructured.UnstructuredList {
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
		Items: []unstructured.Unstructured{*crd1, *crd2},
	}
	return crdList
}

func (badResource) MockConstraintKind() *unstructured.UnstructuredList {
	constraint1 := &unstructured.Unstructured{}
	constraint1.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "constraints.gatekeeper.sh/v1beta1",
		"kind":       1,
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

	constraintList := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"apiVersion": "constraints.gatekeeper.sh/v1beta1",
			"kind":       "gkPolicyList",
		},
		Items: []unstructured.Unstructured{*constraint1},
	}
	return constraintList
}

func (badResource) MockConstraintName() *unstructured.UnstructuredList {
	constraint1 := &unstructured.Unstructured{}
	constraint1.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "constraints.gatekeeper.sh/v1beta1",
		"kind":       "foo",
		"metadata": map[string]interface{}{
			"name": 1,
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

	constraintList := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"apiVersion": "constraints.gatekeeper.sh/v1beta1",
			"kind":       "gkPolicyList",
		},
		Items: []unstructured.Unstructured{*constraint1},
	}
	return constraintList
}

func (badResource) MockConstraintDesc() *unstructured.UnstructuredList {
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
				"constraints.gatekeeper/description": 1,
			},
		},
		"spec": map[string]interface{}{
			"enforcementAction": "deny",
		},
	})

	constraintList := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"apiVersion": "constraints.gatekeeper.sh/v1beta1",
			"kind":       "gkPolicyList",
		},
		Items: []unstructured.Unstructured{*constraint1},
	}
	return constraintList
}

func (badResource) MockConstraintAction() *unstructured.UnstructuredList {
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
			"enforcementAction": 1,
		},
	})

	constraintList := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"apiVersion": "constraints.gatekeeper.sh/v1beta1",
			"kind":       "gkPolicyList",
		},
		Items: []unstructured.Unstructured{*constraint1},
	}
	return constraintList
}

func (badResource) MockPolicyKind() *unstructured.UnstructuredList {
	policy1 := &unstructured.Unstructured{}
	policy1.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "kyverno.io/v1",
		"kind":       1,
		"metadata": map[string]interface{}{
			"name":      "foo-1",
			"namespace": "demo",
			"labels": map[string]interface{}{
				"app.kubernetes.io/name": "kyverno",
			},
			"annotations": map[string]interface{}{
				"policies.kyverno.io/description": "invalid config",
			},
		},
		"spec": map[string]interface{}{
			"validationFailureAction": "enforce",
			"group":                   "kyverno.io",
		},
	})

	policyList := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"apiVersion": "kyverno.io/v1",
			"kind":       "kyvernoPolicyList",
		},
		Items: []unstructured.Unstructured{*policy1},
	}

	return policyList
}

func (badResource) MockPolicyName() *unstructured.UnstructuredList {
	policy1 := &unstructured.Unstructured{}
	policy1.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "kyverno.io/v1",
		"kind":       "foo",
		"metadata": map[string]interface{}{
			"name":      1,
			"namespace": "demo",
			"labels": map[string]interface{}{
				"app.kubernetes.io/name": "kyverno",
			},
			"annotations": map[string]interface{}{
				"policies.kyverno.io/description": "invalid config",
			},
		},
		"spec": map[string]interface{}{
			"validationFailureAction": "enforce",
			"group":                   "kyverno.io",
		},
	})

	policyList := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"apiVersion": "kyverno.io/v1",
			"kind":       "kyvernoPolicyList",
		},
		Items: []unstructured.Unstructured{*policy1},
	}

	return policyList
}

func (badResource) MockPolicyNamespace() *unstructured.UnstructuredList {
	policy1 := &unstructured.Unstructured{}
	policy1.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "kyverno.io/v1",
		"kind":       "foo",
		"metadata": map[string]interface{}{
			"name":      "foo-1",
			"namespace": 1,
			"labels": map[string]interface{}{
				"app.kubernetes.io/name": "kyverno",
			},
			"annotations": map[string]interface{}{
				"policies.kyverno.io/description": "invalid config",
			},
		},
		"spec": map[string]interface{}{
			"validationFailureAction": "enforce",
			"group":                   "kyverno.io",
		},
	})

	policyList := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"apiVersion": "kyverno.io/v1",
			"kind":       "kyvernoPolicyList",
		},
		Items: []unstructured.Unstructured{*policy1},
	}

	return policyList
}

func (badResource) MockPolicyDesc() *unstructured.UnstructuredList {
	policy1 := &unstructured.Unstructured{}
	policy1.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "kyverno.io/v1",
		"kind":       "foo",
		"metadata": map[string]interface{}{
			"name":      "foo-1",
			"namespace": "demo",
			"labels": map[string]interface{}{
				"app.kubernetes.io/name": "kyverno",
			},
			"annotations": map[string]interface{}{
				"policies.kyverno.io/description": 1,
			},
		},
		"spec": map[string]interface{}{
			"validationFailureAction": "enforce",
			"group":                   "kyverno.io",
		},
	})

	policyList := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"apiVersion": "kyverno.io/v1",
			"kind":       "kyvernoPolicyList",
		},
		Items: []unstructured.Unstructured{*policy1},
	}

	return policyList
}

func (badResource) MockPolicyAction() *unstructured.UnstructuredList {
	policy1 := &unstructured.Unstructured{}
	policy1.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "kyverno.io/v1",
		"kind":       "foo",
		"metadata": map[string]interface{}{
			"name":      "foo-1",
			"namespace": "demo",
			"labels": map[string]interface{}{
				"app.kubernetes.io/name": "kyverno",
			},
			"annotations": map[string]interface{}{
				"policies.kyverno.io/description": "invalid config",
			},
		},
		"spec": map[string]interface{}{
			"validationFailureAction": 1,
			"group":                   "kyverno.io",
		},
	})

	policyList := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"apiVersion": "kyverno.io/v1",
			"kind":       "kyvernoPolicyList",
		},
		Items: []unstructured.Unstructured{*policy1},
	}

	return policyList
}
