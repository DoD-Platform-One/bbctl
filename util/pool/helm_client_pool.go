package pool

import (
	helm "repo1.dso.mil/big-bang/product/packages/bbctl/util/helm"
)

// helmClientInstance is a struct that holds a Helm client and the namespace it is configured for
type helmClientInstance struct {
	namespace string
	client    helm.Client
}

// helmClientPool is a slice of helmClientInstance structs
type helmClientPool []*helmClientInstance

// contains checks if a helmClientPool contains a Helm client for a given command and namespace
func (h helmClientPool) contains(namespace string) (bool, helm.Client) {
	for _, client := range h {
		if client.namespace == namespace {
			return true, client.client
		}
	}
	return false, nil
}

// add adds a Helm client to the helmClientPool
func (h *helmClientPool) add(client helm.Client, namespace string) {
	*h = append(*h, &helmClientInstance{
		namespace: namespace,
		client:    client,
	})
}
