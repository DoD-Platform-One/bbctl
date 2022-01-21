package helm

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

// RESTClientGetter defines the values of a helm REST client
type RESTClientGetter struct {
	namespace  string
	restConfig *rest.Config
}

// NewRESTClientGetter returns a RESTClientGetter using the provided 'namespace' and 'restConfig'.
func NewRESTClientGetter(restConfig *rest.Config, namespace string) *RESTClientGetter {
	return &RESTClientGetter{
		namespace:  namespace,
		restConfig: restConfig,
	}
}

// ToRESTConfig returns a REST config build from a given kubeconfig
func (c *RESTClientGetter) ToRESTConfig() (*rest.Config, error) {
	return c.restConfig, nil
}

// ToDiscoveryClient returns a CachedDiscoveryInterface that can be used as a discovery client.
func (c *RESTClientGetter) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {

	config, err := c.ToRESTConfig()
	if err != nil {
		return nil, err
	}

	// The more API groups exist, the more discovery requests need to be made.
	// Given 25 API groups with about one version each, discovery needs to make 50 requests.
	// This setting is only used for discovery.
	config.Burst = 100

	discoveryClient, _ := discovery.NewDiscoveryClientForConfig(config)

	return memory.NewMemCacheClient(discoveryClient), nil
}

// ToRESTMapper - to rest mapper
func (c *RESTClientGetter) ToRESTMapper() (meta.RESTMapper, error) {
	discoveryClient, err := c.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(discoveryClient)
	expander := restmapper.NewShortcutExpander(mapper, discoveryClient)
	return expander, nil
}

// ToRawKubeConfigLoader - to raw kubeconfig loader
func (c *RESTClientGetter) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	return nil
}
