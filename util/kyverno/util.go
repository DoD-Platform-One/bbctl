package kyverno

import (
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// FetchKyvernoCrds - Fetch Kyverno Policy CRDs
func FetchKyvernoCrds(client dynamic.Interface) (*unstructured.UnstructuredList, error) {

	var customResource = schema.GroupVersionResource{Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"}

	opts := metav1.ListOptions{LabelSelector: "app.kubernetes.io/name=kyverno"}

	kyvernoResources, err := client.Resource(customResource).List(context.TODO(), opts)
	if err != nil {
		return nil, fmt.Errorf("error getting kyverno crds: %s", err.Error())
	}

	items := make([]unstructured.Unstructured, 0)
	for _, crd := range kyvernoResources.Items {
		crdName, _, _ := unstructured.NestedString(crd.Object, "metadata", "name")
		if strings.HasSuffix(crdName, "policies.kyverno.io") {
			items = append(items, crd)
		}
	}

	kyvernoResources.Items = items

	return kyvernoResources, nil
}

// FetchKyvernoPolicies - Fetch Kyverno Policies
func FetchKyvernoPolicies(client dynamic.Interface, name string) (*unstructured.UnstructuredList, error) {

	resourceName := strings.Split(name, ".")[0]

	var policyResource = schema.GroupVersionResource{Group: "kyverno.io", Version: "v1", Resource: resourceName}

	resources, err := client.Resource(policyResource).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting kyverno policies: %s", err.Error())
	}

	return resources, nil
}
