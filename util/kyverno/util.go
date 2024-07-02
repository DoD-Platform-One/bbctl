package kyverno

import (
	"context"
	"fmt"
	"strings"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// FetchKyvernoCrds - Fetches all Custom Resource Definitions(CRDs) related to Kyverno Policies from the cluster
// Filters CRDs with the group kyverno.io and returns the list
func FetchKyvernoCrds(client dynamic.Interface) (*unstructured.UnstructuredList, error) {
	var customResource = schema.GroupVersionResource{Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"}

	// Get all the CRDs
	opts := metaV1.ListOptions{LabelSelector: ""}

	allCrds, err := client.Resource(customResource).List(context.TODO(), opts)
	if err != nil {
		return nil, fmt.Errorf("error getting kyverno crds: %s", err.Error())
	}

	items := make([]unstructured.Unstructured, 0)
	for _, crd := range allCrds.Items {
		group, found, err := unstructured.NestedString(crd.Object, "spec", "group")
		if err != nil || !found {
			continue
		}
		if group == "kyverno.io" {
			items = append(items, crd)
		}
	}

	allCrds.Items = items

	return allCrds, nil
}

// FetchKyvernoPolicies - Fetches all Kyverno Policies for the specified resource name(crdname) across different API versions
// Iterates over each API version, collects all retrieved Kyverno Policies into a single list, and returns it
func FetchKyvernoPolicies(client dynamic.Interface, name string) (*unstructured.UnstructuredList, error) {
	resourceName := strings.Split(name, ".")[0]

	versions := []string{"v1", "v1beta1", "v2beta1", "v1alpha2"}

	allPolicies := &unstructured.UnstructuredList{}

	for _, version := range versions {
		var policyResource = schema.GroupVersionResource{Group: "kyverno.io", Version: version, Resource: resourceName}

		resources, err := client.Resource(policyResource).List(context.TODO(), metaV1.ListOptions{})
		if err != nil {
			// The resources call returns an error if the version isn't available for the CRD requested, but not every resource
			// is present in every version of the API so this needs to ignore "resource not found" errors
			if !strings.Contains(err.Error(), "the server could not find the requested resource") {
				return nil, fmt.Errorf("error getting kyverno policies: %s", err.Error())
			} else {
				continue
			}
		}

		allPolicies.Items = append(allPolicies.Items, resources.Items...)
	}

	return allPolicies, nil
}
