package gatekeeper

import (
	"context"
	"fmt"
	"strings"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// FetchGatekeeperConstraints - Fetches all Gatekeeper Constraints for the specified resource name (crdName)
func FetchGatekeeperConstraints(client dynamic.Interface, name string) (*unstructured.UnstructuredList, error) {
	resourceName := strings.Split(name, ".")[0]

	var constraintResource = schema.GroupVersionResource{Group: "constraints.gatekeeper.sh", Version: "v1beta1", Resource: resourceName}

	resources, err := client.Resource(constraintResource).List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting gatekeeper constraint: %s", err.Error())
	}

	return resources, nil
}

// FetchGatekeeperCrds - Fetches all Custom Resource Definitions(CRDs) related to Gatekeeper Constraints from the cluster
func FetchGatekeeperCrds(client dynamic.Interface) (*unstructured.UnstructuredList, error) {
	var customResource = schema.GroupVersionResource{Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"}

	opts := metaV1.ListOptions{LabelSelector: "app.kubernetes.io/name=gatekeeper"}

	gkResources, err := client.Resource(customResource).List(context.TODO(), opts)
	if err != nil {
		return nil, fmt.Errorf("error getting gatekeeper crds: %s", err.Error())
	}

	return gkResources, nil
}
