package helm

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/rest"
	clientCmdApi "k8s.io/client-go/tools/clientcmd/api"
)

func TestToRESTConfig(t *testing.T) {
	testCases := []struct {
		name      string
		shouldErr bool
	}{
		{
			name:      "should not error",
			shouldErr: false,
		},
		{
			name:      "should error",
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			restConfig := &rest.Config{Host: "localhost:8080"}
			restClientGetter := NewRESTClientGetter(restConfig, "default", nil)
			if tc.shouldErr {
				restClientGetter.toRESTConfigShouldErr = true
			}
			// Act
			config, err := restClientGetter.ToRESTConfig()
			// Assert
			if tc.shouldErr {
				assert.Nil(t, config)
				assert.NotNil(t, err)
				assert.Equal(t, "test error", err.Error())
			} else {
				assert.Equal(t, restConfig.Host, config.Host)
				assert.Nil(t, err)
			}
		})
	}
}

func TestToDiscoveryClient(t *testing.T) {
	testCases := []struct {
		name                       string
		shouldErrOnRESTConfig      bool
		shouldErrOnDiscoveryClient bool
	}{
		{
			name:                       "should not error",
			shouldErrOnRESTConfig:      false,
			shouldErrOnDiscoveryClient: false,
		},
		{
			name:                       "should error on REST config",
			shouldErrOnRESTConfig:      true,
			shouldErrOnDiscoveryClient: false,
		},
		{
			name:                       "should error on discovery client",
			shouldErrOnRESTConfig:      false,
			shouldErrOnDiscoveryClient: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			restConfig := &rest.Config{Host: "localhost:8080"}
			if tc.shouldErrOnDiscoveryClient {
				restConfig.ExecProvider = &clientCmdApi.ExecConfig{}
				restConfig.AuthProvider = &clientCmdApi.AuthProviderConfig{}
			}
			restClientGetter := NewRESTClientGetter(restConfig, "default", nil)
			if tc.shouldErrOnRESTConfig {
				restClientGetter.toRESTConfigShouldErr = true
			}
			// Act
			client, err := restClientGetter.ToDiscoveryClient()
			// Assert
			if tc.shouldErrOnRESTConfig {
				assert.Nil(t, client)
				assert.NotNil(t, err)
				assert.Equal(t, "test error", err.Error())
			} else if tc.shouldErrOnDiscoveryClient {
				assert.Nil(t, client)
				assert.NotNil(t, err)
				assert.Equal(t, "execProvider and authProvider cannot be used in combination", err.Error())
			} else {
				assert.NotNil(t, client)
				assert.Nil(t, err)
				assert.Equal(t, restConfig.Host, client.RESTClient().Delete().URL().Host)
			}
		})
	}
}

func TestToRestMapper(t *testing.T) {
	testCases := []struct {
		name      string
		shouldErr bool
	}{
		{
			name:      "should not error",
			shouldErr: false,
		},
		{
			name:      "should error",
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			restConfig := &rest.Config{Host: "localhost:8080"}
			restClientGetter := NewRESTClientGetter(restConfig, "default", nil)
			if tc.shouldErr {
				restClientGetter.toRESTConfigShouldErr = true
			}
			// Act
			mapper, err := restClientGetter.ToRESTMapper()
			// Assert
			if tc.shouldErr {
				assert.Nil(t, mapper)
				assert.NotNil(t, err)
				assert.Equal(t, "test error", err.Error())
			} else {
				assert.NotNil(t, mapper)
				assert.Nil(t, err)
			}
		})
	}
}

func TestToRawKubeConfigLoader(t *testing.T) {
	// Arrange
	// Act
	restClientGetter := NewRESTClientGetter(nil, "default", nil)
	// Assert
	assert.Nil(t, restClientGetter.ToRawKubeConfigLoader())
}

func TestSendWarning(t *testing.T) {
	testCases := []struct {
		name             string
		useCustomHandler bool
	}{
		{
			name:             "use default handler",
			useCustomHandler: false,
		},
		{
			name:             "use custom handler",
			useCustomHandler: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			restConfig := &rest.Config{Host: "localhost:8080"}
			buf := &bytes.Buffer{}
			var warningHandler func(string)
			if tc.useCustomHandler {
				warningHandler = func(s string) {
					buf.WriteString(s)
				}
			}
			restClientGetter := NewRESTClientGetter(restConfig, "default", warningHandler)
			// Act
			restClientGetter.SendWarning("test")
			// Assert
			if tc.useCustomHandler {
				assert.Equal(t, "test", buf.String())
			}
			// TODO: fmt.Print is not tested, it should eventually be tested when it uses streams to print instead of fmt.Print
		})
	}
}
