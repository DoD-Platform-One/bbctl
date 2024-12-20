package pool

import (
	"bytes"
	"log/slog"
	"os"
	"reflect"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	fakeKubernetesDynamic "k8s.io/client-go/dynamic/fake"
	fakeK8s "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	fakeRuntimeClient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	bbAws "repo1.dso.mil/big-bang/product/packages/bbctl/util/aws"
	bbGitLab "repo1.dso.mil/big-bang/product/packages/bbctl/util/gitlab"
	bbHelm "repo1.dso.mil/big-bang/product/packages/bbctl/util/helm"
	bbLog "repo1.dso.mil/big-bang/product/packages/bbctl/util/log"
	bbOutput "repo1.dso.mil/big-bang/product/packages/bbctl/util/output"
	bbUtilTestApiWrappers "repo1.dso.mil/big-bang/product/packages/bbctl/util/test/apiwrappers"
)

func getCommonTestCases(t *testing.T) []struct {
	name        string
	errored     bool
	bubbleError bool
} {
	t.Helper()
	return []struct {
		name        string
		errored     bool
		bubbleError bool
	}{
		{
			name:        "success",
			errored:     false,
			bubbleError: false,
		},
		{
			name:        "top level error",
			errored:     true,
			bubbleError: false,
		},
		{
			name:        "underlying error",
			errored:     true,
			bubbleError: true,
		},
	}
}

func TestErrFactoryNotInitialized_Error(t *testing.T) {
	// arrange
	err := FactoryNotInitializedError{}
	// act
	result := err.Error()
	// assert
	assert.Equal(t, "factory not initialized", result)
}

func TestNewPooledFactory(t *testing.T) {
	// arrange
	// act
	result := NewPooledFactory()
	// assert
	assert.NotNil(t, result)
	assert.Nil(t, result.underlyingFactory)
}

func TestPooledFactory_SetUnderlyingFactory(t *testing.T) {
	// arrange
	factory1 := NewPooledFactory()
	factory2 := NewPooledFactory()
	// act
	factory1.SetUnderlyingFactory(factory2)
	// assert
	assert.Equal(t, factory2, factory1.underlyingFactory)
	assert.Nil(t, factory2.underlyingFactory)
	assert.NotEqual(t, factory1, factory2)
}

func TestPooledFactory_GetAWSClient(t *testing.T) {
	testCases := getCommonTestCases(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// arrange
			factory1 := NewPooledFactory()
			factory2 := NewPooledFactory()
			if !tc.errored {
				clientGetter := bbAws.ClientGetter{}
				awsClient := clientGetter.GetClient()
				factory2.awsClient = awsClient
			}
			if tc.bubbleError || !tc.errored {
				factory1.SetUnderlyingFactory(factory2)
			}
			// act
			result, err := factory1.GetAWSClient()
			cachedResult, cachedErr := factory1.GetAWSClient()
			// assert
			if tc.errored {
				assert.Nil(t, result)
				require.Error(t, err)
				assert.IsType(t, &FactoryNotInitializedError{}, err)
				assert.Nil(t, cachedResult)
				require.Error(t, cachedErr)
				assert.IsType(t, &FactoryNotInitializedError{}, cachedErr)
				assert.Nil(t, factory1.awsClient)
			} else {
				assert.NotNil(t, result)
				require.NoError(t, err)
				require.NoError(t, cachedErr)
				assert.Equal(t, factory1.awsClient, result)
				assert.Equal(t, factory1.awsClient, cachedResult)
				assert.Equal(t, factory1.awsClient, factory2.awsClient)
			}
		})
	}
}

func TestPooledFactory_GetGitLabClient(t *testing.T) {
	testCases := getCommonTestCases(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// arrange
			factory1 := NewPooledFactory()
			factory2 := NewPooledFactory()
			if !tc.errored {
				clientGetter := bbGitLab.ClientGetter{}
				gitLabClient, err := clientGetter.GetClient("", "")
				require.NoError(t, err)
				factory2.gitLabClient = gitLabClient
			}
			if tc.bubbleError || !tc.errored {
				factory1.SetUnderlyingFactory(factory2)
			}
			// act
			result, err := factory1.GetGitLabClient()
			cachedResult, cachedErr := factory1.GetGitLabClient()
			// assert
			if tc.errored {
				assert.Nil(t, result)
				require.Error(t, err)
				assert.IsType(t, &FactoryNotInitializedError{}, err)
				assert.Nil(t, cachedResult)
				require.Error(t, cachedErr)
				assert.IsType(t, &FactoryNotInitializedError{}, cachedErr)
				assert.Nil(t, factory1.gitLabClient)
			} else {
				assert.NotNil(t, result)
				require.NoError(t, err)
				require.NoError(t, cachedErr)
				assert.Equal(t, factory1.gitLabClient, result)
				assert.Equal(t, factory1.gitLabClient, cachedResult)
				assert.Equal(t, factory1.gitLabClient, factory2.gitLabClient)
			}
		})
	}
}

func TestPooledFactory_GetHelmClient(t *testing.T) {
	testCases := getCommonTestCases(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// arrange
			factory1 := NewPooledFactory()
			factory2 := NewPooledFactory()
			cmd := &cobra.Command{}
			namespace := "test"
			var helmClientPool helmClientPool
			if !tc.errored {
				restConfig, err := bbHelm.NewClient(nil, nil, nil)
				require.NoError(t, err)
				helmClientPool = []*helmClientInstance{
					{
						namespace: namespace,
						client:    restConfig,
					},
				}
				factory2.helmClients = helmClientPool
			}
			if tc.bubbleError || !tc.errored {
				factory1.SetUnderlyingFactory(factory2)
			}
			// act
			result, err := factory1.GetHelmClient(cmd, namespace)
			cachedResult, cachedErr := factory1.GetHelmClient(cmd, namespace)
			// assert
			if tc.errored {
				assert.Nil(t, result)
				require.Error(t, err)
				assert.IsType(t, &FactoryNotInitializedError{}, err)
				assert.Nil(t, cachedResult)
				require.Error(t, cachedErr)
				assert.IsType(t, &FactoryNotInitializedError{}, cachedErr)
				assert.Empty(t, factory1.helmClients)
			} else {
				assert.NotNil(t, result)
				require.NoError(t, err)
				require.NoError(t, cachedErr)
				assert.Equal(t, factory1.helmClients, helmClientPool)
				assert.Equal(t, factory1.helmClients, factory2.helmClients)
				assert.Equal(t, factory1.helmClients[0].client, result)
				assert.Equal(t, factory1.helmClients[0].client, cachedResult)
			}
		})
	}
}

func TestPooledFactory_GetK8sClientset(t *testing.T) {
	testCases := getCommonTestCases(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// arrange
			factory1 := NewPooledFactory()
			factory2 := NewPooledFactory()
			cmd := &cobra.Command{
				Use: "test",
			}
			client := fakeK8s.NewSimpleClientset()
			if !tc.errored {
				factory2.k8sClientset = client
			}
			if tc.bubbleError || !tc.errored {
				factory1.SetUnderlyingFactory(factory2)
			}
			// act
			result, err := factory1.GetK8sClientset(cmd)
			cachedResult, cachedErr := factory1.GetK8sClientset(cmd)
			// assert
			if tc.errored {
				assert.Nil(t, result)
				require.Error(t, err)
				assert.IsType(t, &FactoryNotInitializedError{}, err)
				assert.Nil(t, cachedResult)
				require.Error(t, cachedErr)
				assert.IsType(t, &FactoryNotInitializedError{}, cachedErr)
				assert.Nil(t, factory1.k8sClientset)
			} else {
				assert.NotNil(t, result)
				require.NoError(t, err)
				require.NoError(t, cachedErr)
				assert.Equal(t, factory1.k8sClientset, client)
				assert.Equal(t, factory1.k8sClientset, factory2.k8sClientset)
				assert.Equal(t, factory1.k8sClientset, result)
				assert.Equal(t, factory1.k8sClientset, cachedResult)
			}
		})
	}
}

func TestPooledFactory_GetLoggingClient(t *testing.T) {
	testCases := getCommonTestCases(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// arrange
			factory1 := NewPooledFactory()
			factory2 := NewPooledFactory()
			var loggerClientPool loggerClientPool
			if !tc.errored {
				clientGetter := bbLog.ClientGetter{}
				loggerClient := clientGetter.GetClient(slog.Default())
				loggerClientPool = []*loggerClientInstance{
					{
						client: loggerClient,
						logger: nil,
					},
				}
				factory2.loggerClients = loggerClientPool
			}
			if tc.bubbleError || !tc.errored {
				factory1.SetUnderlyingFactory(factory2)
			}
			// act
			result, err := factory1.GetLoggingClient()
			cachedResult, cachedErr := factory1.GetLoggingClient()
			// assert
			if tc.errored {
				assert.Nil(t, result)
				require.Error(t, err)
				assert.IsType(t, &FactoryNotInitializedError{}, err)
				assert.Nil(t, cachedResult)
				require.Error(t, cachedErr)
				assert.IsType(t, &FactoryNotInitializedError{}, cachedErr)
				assert.Empty(t, factory1.loggerClients)
			} else {
				assert.NotNil(t, result)
				require.NoError(t, err)
				require.NoError(t, cachedErr)
				assert.Equal(t, factory1.loggerClients, loggerClientPool)
				assert.Equal(t, factory1.loggerClients, factory2.loggerClients)
				assert.Equal(t, factory1.loggerClients[0].client, result)
				assert.Equal(t, factory1.loggerClients[0].client, cachedResult)
			}
		})
	}
}

func TestPooledFactory_GetLoggingClientWithLogger(t *testing.T) {
	testCases := getCommonTestCases(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// arrange
			factory1 := NewPooledFactory()
			factory2 := NewPooledFactory()
			logger := slog.Default()
			var loggerClientPool loggerClientPool
			if !tc.errored {
				clientGetter := bbLog.ClientGetter{}
				loggerClient := clientGetter.GetClient(logger)
				loggerClientPool = []*loggerClientInstance{
					{
						client: loggerClient,
						logger: logger,
					},
				}
				factory2.loggerClients = loggerClientPool
			}
			if tc.bubbleError || !tc.errored {
				factory1.SetUnderlyingFactory(factory2)
			}
			// act
			result, err := factory1.GetLoggingClientWithLogger(logger)
			cachedResult, cachedErr := factory1.GetLoggingClientWithLogger(logger)
			// assert
			if tc.errored {
				assert.Nil(t, result)
				require.Error(t, err)
				assert.IsType(t, &FactoryNotInitializedError{}, err)
				assert.Nil(t, cachedResult)
				require.Error(t, cachedErr)
				assert.IsType(t, &FactoryNotInitializedError{}, cachedErr)
				assert.Empty(t, factory1.loggerClients)
			} else {
				assert.NotNil(t, result)
				require.NoError(t, err)
				require.NoError(t, cachedErr)
				assert.Equal(t, factory1.loggerClients, loggerClientPool)
				assert.Equal(t, factory1.loggerClients, factory2.loggerClients)
				assert.Equal(t, factory1.loggerClients[0].client, result)
				assert.Equal(t, factory1.loggerClients[0].client, cachedResult)
			}
		})
	}
}

func TestPooledFactory_GetRuntimeClient(t *testing.T) {
	testCases := getCommonTestCases(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// arrange
			factory1 := NewPooledFactory()
			factory2 := NewPooledFactory()
			var runtimeClientPool runtimeClientPool
			goodScheme := runtime.NewScheme()
			require.NoError(t, goodScheme.SetVersionPriority(schema.GroupVersion{Group: "test", Version: "v1"}))
			if !tc.errored {
				goodClientBuilder := &fakeRuntimeClient.ClientBuilder{}
				goodClientBuilder.WithScheme(goodScheme)
				goodClient := goodClientBuilder.Build()
				runtimeClientPool = []*runtimeClientInstance{
					{
						client: goodClient,
						scheme: goodScheme,
					},
				}
				factory2.runtimeClients = runtimeClientPool
			}
			if tc.bubbleError || !tc.errored {
				factory1.SetUnderlyingFactory(factory2)
			}
			// act
			result, err := factory1.GetRuntimeClient(goodScheme)
			cachedResult, cachedErr := factory1.GetRuntimeClient(goodScheme)
			// assert
			if tc.errored {
				assert.Nil(t, result)
				require.Error(t, err)
				assert.Nil(t, cachedResult)
				require.Error(t, cachedErr)
				assert.IsType(t, &FactoryNotInitializedError{}, err)
				assert.IsType(t, &FactoryNotInitializedError{}, cachedErr)
				assert.Empty(t, factory1.runtimeClients)
			} else {
				assert.NotNil(t, result)
				assert.NotNil(t, cachedResult)
				require.NoError(t, err)
				require.NoError(t, cachedErr)
				assert.Equal(t, factory1.runtimeClients, factory2.runtimeClients)
				assert.Equal(t, factory1.runtimeClients, runtimeClientPool)
				assert.Equal(t, factory1.runtimeClients, factory2.runtimeClients)
				assert.Equal(t, factory1.runtimeClients[0].client, result)
				assert.Equal(t, factory1.runtimeClients[0].client, cachedResult)
			}
		})
	}
}

func TestPooledFactory_GetK8sDynamicClient(t *testing.T) {
	testCases := getCommonTestCases(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// arrange
			factory1 := NewPooledFactory()
			factory2 := NewPooledFactory()
			cmd := &cobra.Command{
				Use: "test",
			}
			client := fakeKubernetesDynamic.NewSimpleDynamicClient(&runtime.Scheme{})
			if !tc.errored {
				factory2.k8sDynamicClient = client
			}
			if tc.bubbleError || !tc.errored {
				factory1.SetUnderlyingFactory(factory2)
			}
			// act
			result, err := factory1.GetK8sDynamicClient(cmd)
			cachedResult, cachedErr := factory1.GetK8sDynamicClient(cmd)
			// assert
			if tc.errored {
				assert.Nil(t, result)
				require.Error(t, err)
				assert.IsType(t, &FactoryNotInitializedError{}, err)
				assert.Nil(t, cachedResult)
				require.Error(t, cachedErr)
				assert.IsType(t, &FactoryNotInitializedError{}, cachedErr)
				assert.Empty(t, factory1.k8sDynamicClient)
			} else {
				assert.NotNil(t, result)
				require.NoError(t, err)
				require.NoError(t, cachedErr)
				assert.Equal(t, factory1.k8sDynamicClient, client)
				assert.Equal(t, factory1.k8sDynamicClient, factory2.k8sDynamicClient)
				assert.Equal(t, factory1.k8sDynamicClient, result)
				assert.Equal(t, factory1.k8sDynamicClient, cachedResult)
			}
		})
	}
}

func TestPooledFactory_GetOutputClient(t *testing.T) {
	testCases := getCommonTestCases(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// arrange
			factory1 := NewPooledFactory()
			factory2 := NewPooledFactory()
			cmd := &cobra.Command{
				Use: "test",
			}
			streams := genericiooptions.IOStreams{
				In:     &bytes.Buffer{},
				Out:    &bytes.Buffer{},
				ErrOut: &bytes.Buffer{},
			}
			clientGetter := bbOutput.ClientGetter{}
			outputClient := clientGetter.GetClient("", streams)
			if !tc.errored {
				factory2.outputClient = &outputClient
			}
			if tc.bubbleError || !tc.errored {
				factory1.SetUnderlyingFactory(factory2)
			}
			// act
			result, err := factory1.GetOutputClient(cmd)
			cachedResult, cachedErr := factory1.GetOutputClient(cmd)
			// assert
			if tc.errored {
				assert.Nil(t, result)
				require.Error(t, err)
				assert.IsType(t, &FactoryNotInitializedError{}, err)
				assert.Nil(t, cachedResult)
				require.Error(t, cachedErr)
				assert.IsType(t, &FactoryNotInitializedError{}, cachedErr)
				assert.Empty(t, factory1.outputClient)
			} else {
				assert.NotNil(t, result)
				require.NoError(t, err)
				require.NoError(t, cachedErr)
				assert.Equal(t, factory1.outputClient, &outputClient)
				assert.Equal(t, factory1.outputClient, factory2.outputClient)
				assert.Equal(t, factory1.outputClient, &result)
				assert.Equal(t, factory1.outputClient, &cachedResult)
			}
		})
	}
}

func TestPooledFactory_GetRestConfig(t *testing.T) {
	testCases := getCommonTestCases(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// arrange
			factory1 := NewPooledFactory()
			factory2 := NewPooledFactory()
			cmd := &cobra.Command{
				Use: "test",
			}
			restConfig := &rest.Config{}
			if !tc.errored {
				factory2.restConfig = restConfig
			}
			if tc.bubbleError || !tc.errored {
				factory1.SetUnderlyingFactory(factory2)
			}
			// act
			result, err := factory1.GetRestConfig(cmd)
			cachedResult, cachedErr := factory1.GetRestConfig(cmd)
			// assert
			if tc.errored {
				assert.Nil(t, result)
				require.Error(t, err)
				assert.IsType(t, &FactoryNotInitializedError{}, err)
				assert.Nil(t, cachedResult)
				require.Error(t, cachedErr)
				assert.IsType(t, &FactoryNotInitializedError{}, cachedErr)
				assert.Empty(t, factory1.restConfig)
			} else {
				assert.NotNil(t, result)
				require.NoError(t, err)
				require.NoError(t, cachedErr)
				assert.Equal(t, factory1.restConfig, restConfig)
				assert.Equal(t, factory1.restConfig, factory2.restConfig)
				assert.Equal(t, factory1.restConfig, result)
				assert.Equal(t, factory1.restConfig, cachedResult)
			}
		})
	}
}

func TestPooledFactory_GetCommandExecutor(t *testing.T) {
	testCases := getCommonTestCases(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// arrange
			factory1 := NewPooledFactory()
			factory2 := NewPooledFactory()
			cmd := &cobra.Command{
				Use: "test",
			}
			pod := &coreV1.Pod{
				ObjectMeta: metaV1.ObjectMeta{
					Name: "test",
				},
			}
			container := "test"
			command := []string{"echo", "test"}
			stdout := &bytes.Buffer{}
			stdout.WriteString("test")
			stderr := &bytes.Buffer{}
			stderr.WriteString("testerr")
			if tc.bubbleError || !tc.errored {
				factory1.SetUnderlyingFactory(factory2)
			}
			// act
			result, err := factory1.GetCommandExecutor(cmd, pod, container, command, stdout, stderr)
			cachedResult, cachedErr := factory1.GetCommandExecutor(cmd, pod, container, command, stdout, stderr)
			// assert
			// all paths error if there isn't a non-pooled factory because there is no cache
			assert.Nil(t, result)
			require.Error(t, err)
			assert.IsType(t, &FactoryNotInitializedError{}, err)
			assert.Nil(t, cachedResult)
			require.Error(t, cachedErr)
			assert.IsType(t, &FactoryNotInitializedError{}, cachedErr)
		})
	}
}

func TestPooledFactory_GetCredentialHelper(t *testing.T) {
	testCases := getCommonTestCases(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// arrange
			factory1 := NewPooledFactory()
			factory2 := NewPooledFactory()
			credentialHelper := func(string, string) (string, error) {
				return "test", nil
			}
			if !tc.errored {
				factory2.credentialHelper = credentialHelper
			}
			if tc.bubbleError || !tc.errored {
				factory1.SetUnderlyingFactory(factory2)
			}
			// act
			result, err := factory1.GetCredentialHelper()
			cachedResult, cachedErr := factory1.GetCredentialHelper()
			// assert
			if tc.errored {
				assert.Nil(t, result)
				require.Error(t, err)
				assert.IsType(t, &FactoryNotInitializedError{}, err)
				assert.Nil(t, cachedResult)
				require.Error(t, cachedErr)
				assert.IsType(t, &FactoryNotInitializedError{}, cachedErr)
				assert.Nil(t, factory1.credentialHelper)
			} else {
				assert.NotNil(t, result)
				require.NoError(t, err)
				require.NoError(t, cachedErr)
				// you can't compare functions directly, so we'll just check that they're the same type
				assert.Equal(t, reflect.TypeOf(factory1.credentialHelper), reflect.TypeOf(result))
				assert.Equal(t, reflect.TypeOf(factory1.credentialHelper), reflect.TypeOf(cachedResult))
				// the closure isn't a CredentialHelper, so we can't compare types directly
				// assert.Equal(t, reflect.TypeOf(factory1.credentialHelper), reflect.TypeOf(credentialHelper))
			}
		})
	}
}

func TestPooledFactory_GetCommandWrapper(t *testing.T) {
	testCases := getCommonTestCases(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// arrange
			factory1 := NewPooledFactory()
			factory2 := NewPooledFactory()
			name := "test"
			args := []string{"echo", "test"}
			if tc.bubbleError || !tc.errored {
				factory1.SetUnderlyingFactory(factory2)
			}
			// act
			result, err := factory1.GetCommandWrapper(name, args...)
			cachedResult, cachedErr := factory1.GetCommandWrapper(name, args...)
			// assert
			// all paths error if there isn't a non-pooled factory, because there is no cache
			assert.Nil(t, result)
			require.Error(t, err)
			assert.IsType(t, &FactoryNotInitializedError{}, err)
			assert.Nil(t, cachedResult)
			require.Error(t, cachedErr)
			assert.IsType(t, &FactoryNotInitializedError{}, cachedErr)
		})
	}
}

func TestPooledFactory_GetIstioClientSet(t *testing.T) {
	testCases := getCommonTestCases(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// arrange
			factory1 := NewPooledFactory()
			factory2 := NewPooledFactory()
			cfg := &rest.Config{
				Host: "test",
			}
			var istioClientSetPool istioClientsetPool
			if !tc.errored {
				client := bbUtilTestApiWrappers.NewFakeIstioClientSet(nil, bbUtilTestApiWrappers.SetFail{GetList: false})
				istioClientSetPool = []*istioClientsetInstance{
					{
						restConfig: cfg,
						clientset:  client,
					},
				}
				factory2.istioClientSets = istioClientSetPool
			}
			if tc.bubbleError || !tc.errored {
				factory1.SetUnderlyingFactory(factory2)
			}
			// act
			result, err := factory1.GetIstioClientSet(cfg)
			cachedResult, cachedErr := factory1.GetIstioClientSet(cfg)
			// assert
			if tc.errored {
				assert.Nil(t, result)
				require.Error(t, err)
				assert.IsType(t, &FactoryNotInitializedError{}, err)
				assert.Nil(t, cachedResult)
				require.Error(t, cachedErr)
				assert.IsType(t, &FactoryNotInitializedError{}, cachedErr)
				assert.Empty(t, factory1.istioClientSets)
			} else {
				assert.NotNil(t, result)
				require.NoError(t, err)
				require.NoError(t, cachedErr)
				assert.Equal(t, factory1.istioClientSets, istioClientSetPool)
				assert.Equal(t, factory1.istioClientSets, factory2.istioClientSets)
				assert.Equal(t, factory1.istioClientSets[0].clientset, result)
				assert.Equal(t, factory1.istioClientSets[0].clientset, cachedResult)
			}
		})
	}
}

func TestPooledFactory_GetConfigClient(t *testing.T) {
	testCases := getCommonTestCases(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// arrange
			factory1 := NewPooledFactory()
			factory2 := NewPooledFactory()
			cmd := &cobra.Command{
				Use: "test",
			}
			if tc.bubbleError || !tc.errored {
				factory1.SetUnderlyingFactory(factory2)
			}
			// act
			result, err := factory1.GetConfigClient(cmd)
			cachedResult, cachedErr := factory1.GetConfigClient(cmd)
			// assert
			// this will always error because there is no caching
			assert.Nil(t, result)
			require.Error(t, err)
			assert.IsType(t, &FactoryNotInitializedError{}, err)
			assert.Nil(t, cachedResult)
			require.Error(t, cachedErr)
			assert.IsType(t, &FactoryNotInitializedError{}, cachedErr)
		})
	}
}

func TestPooledFactory_GetViper(t *testing.T) {
	testCases := getCommonTestCases(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// arrange
			v := viper.New()
			v.Set("test", "test")
			factory1 := NewPooledFactory()
			factory2 := NewPooledFactory()
			if !tc.errored {
				factory2.viper = v
			}
			if tc.bubbleError || !tc.errored {
				factory1.SetUnderlyingFactory(factory2)
			}
			// act
			result, err := factory1.GetViper()
			cachedResult, cachedErr := factory1.GetViper()
			// assert
			if tc.errored {
				assert.Nil(t, result)
				require.Error(t, err)
				assert.IsType(t, &FactoryNotInitializedError{}, err)
				assert.Nil(t, cachedResult)
				require.Error(t, cachedErr)
				assert.IsType(t, &FactoryNotInitializedError{}, cachedErr)
				assert.Nil(t, factory1.viper)
			} else {
				assert.NotNil(t, result)
				require.NoError(t, err)
				require.NoError(t, cachedErr)
				assert.Equal(t, factory1.viper, v)
				assert.Equal(t, factory1.viper, result)
				assert.Equal(t, factory1.viper, cachedResult)
			}
		})
	}
}

func TestPooledFactory_GetIOStream(t *testing.T) {
	testCases := getCommonTestCases(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v := viper.New()
			v.Set("test", "test")
			factory1 := NewPooledFactory()
			factory2 := NewPooledFactory()
			streams := &genericiooptions.IOStreams{
				In:     &bytes.Buffer{},
				Out:    &bytes.Buffer{},
				ErrOut: &bytes.Buffer{},
			}
			if !tc.errored {
				factory2.ioStream = streams
			}
			if tc.bubbleError || !tc.errored {
				factory1.SetUnderlyingFactory(factory2)
			}
			// act
			result, err := factory1.GetIOStream()
			cachedResult, cachedErr := factory1.GetIOStream()
			// assert
			if tc.errored {
				assert.Nil(t, result)
				require.Error(t, err)
				assert.IsType(t, &FactoryNotInitializedError{}, err)
				assert.Nil(t, cachedResult)
				require.Error(t, cachedErr)
				assert.IsType(t, &FactoryNotInitializedError{}, cachedErr)
				assert.Nil(t, factory1.ioStream)
			} else {
				assert.NotNil(t, result)
				require.NoError(t, err)
				require.NoError(t, cachedErr)
				assert.Equal(t, factory1.ioStream, streams)
				assert.Equal(t, factory1.ioStream, result)
				assert.Equal(t, factory1.ioStream, cachedResult)
			}
		})
	}
}

func TestPooledFactory_GetPipe(t *testing.T) {
	testCases := getCommonTestCases(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			factory1 := NewPooledFactory()
			factory2 := NewPooledFactory()
			fakeReader, fakeWriter, err := os.Pipe()
			require.NoError(t, err)
			if !tc.errored {
				factory2.pipeReader, factory2.pipeWriter = fakeReader, fakeWriter
			}
			if tc.bubbleError || !tc.errored {
				factory1.SetUnderlyingFactory(factory2)
			}
			// act
			reader, writer, err := factory1.GetPipe()
			cachedReader, cachedWriter, cachedErr := factory1.GetPipe()
			// assert
			if tc.errored {
				assert.Nil(t, reader)
				assert.Nil(t, writer)
				require.Error(t, err)
				assert.IsType(t, &FactoryNotInitializedError{}, err)
				assert.Nil(t, cachedReader)
				assert.Nil(t, cachedWriter)
				require.Error(t, cachedErr)
				assert.IsType(t, &FactoryNotInitializedError{}, cachedErr)
			} else {
				assert.NotNil(t, reader)
				assert.NotNil(t, writer)
				require.NoError(t, err)
				require.NoError(t, cachedErr)
				assert.Equal(t, factory1.pipeReader, reader)
				assert.Equal(t, factory1.pipeWriter, writer)
				assert.Equal(t, factory1.pipeReader, cachedReader)
				assert.Equal(t, factory1.pipeWriter, cachedWriter)
			}
		})
	}
}
