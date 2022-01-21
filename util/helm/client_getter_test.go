package helm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/rest"
)

func TestToRESTConfig(t *testing.T) {
	restConfig := &rest.Config{Host: "localhost:8080"}
	restClientGetter := NewRESTClientGetter(restConfig, "default")
	config, err := restClientGetter.ToRESTConfig()
	assert.Equal(t, restConfig.Host, config.Host)
	assert.Nil(t, err)
}

func TestToDiscoveryClient(t *testing.T) {
	restConfig := &rest.Config{Host: "localhost:8080"}
	restClientGetter := NewRESTClientGetter(restConfig, "default")
	client, err := restClientGetter.ToDiscoveryClient()
	assert.NotNil(t, client)
	assert.Nil(t, err)
}

func TestToRestMapper(t *testing.T) {
	restConfig := &rest.Config{Host: "localhost:8080"}
	restClientGetter := NewRESTClientGetter(restConfig, "default")
	mapper, err := restClientGetter.ToRESTMapper()
	assert.NotNil(t, mapper)
	assert.Nil(t, err)
}

func TestToRawKubeConfigLoader(t *testing.T) {
	restClientGetter := NewRESTClientGetter(nil, "default")
	assert.Nil(t, restClientGetter.ToRawKubeConfigLoader())
}
