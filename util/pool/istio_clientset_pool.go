package pool

import (
	"k8s.io/client-go/rest"
	bbUtilApiWrappers "repo1.dso.mil/big-bang/product/packages/bbctl/util/apiwrappers"
)

// istioClientsetInstance is a struct that holds an Istio clientset and the rest config it is configured for
type istioClientsetInstance struct {
	restConfig *rest.Config
	clientset  bbUtilApiWrappers.IstioClientset
}

// istioClientsetPool is a slice of istioClientsetInstance structs
type istioClientsetPool []*istioClientsetInstance

// contains checks if an istioClientsetPool contains an Istio clientset for a given command and rest config
func (i istioClientsetPool) contains(restConfig *rest.Config) (bool, bbUtilApiWrappers.IstioClientset) {
	for _, client := range i {
		if client.restConfig == restConfig {
			return true, client.clientset
		}
	}
	return false, nil
}

// add adds an Istio clientset to the istioClientsetPool
func (i *istioClientsetPool) add(client bbUtilApiWrappers.IstioClientset, restConfig *rest.Config) {
	*i = append(*i, &istioClientsetInstance{
		restConfig: restConfig,
		clientset:  client,
	})
}
