package helm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/rest"
)

func TestToRESTConfig(t *testing.T) {
	restConfig := &rest.Config{Host: "localhost:8080"}
	restClientGetter := NewRESTClientGetter(restConfig, "default", nil)
	config, err := restClientGetter.ToRESTConfig()
	assert.Equal(t, restConfig.Host, config.Host)
	assert.Nil(t, err)
}

func TestToDiscoveryClient(t *testing.T) {
	restConfig := &rest.Config{Host: "localhost:8080"}
	restClientGetter := NewRESTClientGetter(restConfig, "default", nil)
	client, err := restClientGetter.ToDiscoveryClient()
	assert.NotNil(t, client)
	assert.Nil(t, err)
}

func TestToRestMapper(t *testing.T) {
	restConfig := &rest.Config{Host: "localhost:8080"}
	restClientGetter := NewRESTClientGetter(restConfig, "default", nil)
	mapper, err := restClientGetter.ToRESTMapper()
	assert.NotNil(t, mapper)
	assert.Nil(t, err)
}

func TestToRawKubeConfigLoader(t *testing.T) {
	restClientGetter := NewRESTClientGetter(nil, "default", nil)
	assert.Nil(t, restClientGetter.ToRawKubeConfigLoader())
}

func TestAlternateWarningHandler(t *testing.T) {
	restConfig := &rest.Config{Host: "localhost:8080"}
	valueToModify := "test"
	expectedValue := "something else"
	restClientGetter := NewRESTClientGetter(restConfig, "default", func(warning string) {
		valueToModify = warning
	})
	restClientGetter.SendWarning(expectedValue)
	assert.Equal(t, expectedValue, valueToModify)
}
