package pool

import (
	"k8s.io/apimachinery/pkg/runtime"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

// runtimeClientInstance is a struct that holds a runtime client and the scheme it is configured for
type runtimeClientInstance struct {
	scheme *runtime.Scheme
	client runtimeClient.Client
}

// runtimeClientPool is a slice of runtimeClientInstance structs
type runtimeClientPool []*runtimeClientInstance

// contains checks if a runtimeClientPool contains a runtime client for a given scheme
func (r runtimeClientPool) contains(scheme *runtime.Scheme) (bool, runtimeClient.Client) {
	for _, client := range r {
		if client.scheme == scheme {
			return true, client.client
		}
	}
	return false, nil
}

// add adds a runtime client to the runtimeClientPool
func (r *runtimeClientPool) add(client runtimeClient.Client, scheme *runtime.Scheme) {
	*r = append(*r, &runtimeClientInstance{
		scheme: scheme,
		client: client,
	})
}
