package gatekeeper

import (
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// FetchGatekeeperConstraints - Fetch Gatekeeper Constraints
func FetchGatekeeperConstraints(client dynamic.Interface, name string) (*unstructured.UnstructuredList, error) {

	resourceName := strings.Split(name, ".")[0]

	var constraintResource = schema.GroupVersionResource{Group: "constraints.gatekeeper.sh", Version: "v1beta1", Resource: resourceName}

	resources, err := client.Resource(constraintResource).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting gatekeeper constraint: %s", err.Error())
	}

	return resources, nil
}

// FetchGatekeeperCrds - Fetch Gatekeeper Custom Resource Dedinitions
func FetchGatekeeperCrds(client dynamic.Interface) (*unstructured.UnstructuredList, error) {

	var customResource = schema.GroupVersionResource{Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"}

	opts := metav1.ListOptions{LabelSelector: "app.kubernetes.io/name=gatekeeper"}

	gkResources, err := client.Resource(customResource).List(context.TODO(), opts)
	if err != nil {
		return nil, fmt.Errorf("error getting gatekeeper crds: %s", err.Error())
	}

	return gkResources, nil
}
