package util

import (
	"bytes"
	"log/slog"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestReadDefaultCredentialsFile(t *testing.T) {
	// Arrange
	factory := NewFactory()
	viperInstance := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "./test/data/kube-config.yaml")

	homeDir, _ := os.UserHomeDir()
	credsDir := path.Join(homeDir, ".bbctl")
	os.Rename(path.Join(credsDir, "credentials.yaml"), path.Join(credsDir, "old-credentials.yaml"))

	// Act & Assert
	assert.Panics(t, func() {
		factory.ReadCredentialsFile("", "")
	})

	// Cleanup
	os.Rename(path.Join(credsDir, "old-credentials.yaml"), path.Join(credsDir, "credentials.yaml"))
}

func TestReadCredentialsFile(t *testing.T) {
	// Arrange
	factory := NewFactory()
	viperInstance := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "./test/data/kube-config.yaml")
	viperInstance.Set("big-bang-credential-helper-credentials-file-path", "./test/data/test-credentials.yaml")

	// Act
	username := factory.ReadCredentialsFile("username", "https://test.com:6443")
	password := factory.ReadCredentialsFile("password", "https://test.com:6443")

	// Assert
	assert.Equal(t, username, "username")
	assert.Equal(t, password, "password")
	assert.Panics(t, func() {
		factory.ReadCredentialsFile("invalidFieldName", "https://test.com:6443")
	})
	assert.Panics(t, func() {
		factory.ReadCredentialsFile("username", "invalidURI")
	})
}

func TestGetCredentialHelper(t *testing.T) {
	// Arrange
	factory := NewFactory()
	viperInstance := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "./test/data/kube-config.yaml")
	viperInstance.Set("big-bang-credential-helper-credentials-file-path", "./test/data/test-credentials.yaml")

	var tests = []struct {
		name             string
		credentialHelper string
		field            string
		expected         string
		panics           bool
	}{
		{
			name:             "EmptyCredsHelper",
			credentialHelper: "",
			field:            "username",
			expected:         "",
			panics:           true,
		},
		{
			name:             "CustomCredsHelperNonEmpty",
			credentialHelper: "./../scripts/factory-tests/fake-credentials-helper.sh",
			field:            "username",
			expected:         "username",
			panics:           false,
		},
		{
			name:             "CustomCredsHelperEmpty",
			credentialHelper: "./../scripts/factory-tests/fake-credentials-helper.sh",
			field:            "password",
			expected:         "",
			panics:           true,
		},
		{
			name:             "InvalidCredsField",
			credentialHelper: "./../scripts/factory-tests/fake-credentials-helper.sh",
			field:            "invalidCredsField",
			expected:         "",
			panics:           true,
		},
		{
			name:             "InvalidUriForDefaultHelper",
			credentialHelper: "credentials-file",
			field:            "username",
			expected:         "",
			panics:           true,
		},
		{
			name:             "ValidUriForDefaultHelper",
			credentialHelper: "credentials-file",
			field:            "username",
			expected:         "username",
			panics:           false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Act
			viperInstance.Set("big-bang-credential-helper", test.credentialHelper)
			helper := factory.GetCredentialHelper()
			// Assert
			if test.panics {
				assert.Panics(t, func() {
					helper(test.field, "https://invalidUri.com:6443")
				})
			} else {
				assert.Equal(t, test.expected, strings.TrimSuffix(helper(test.field, "https://test.com:6443"), "\n"))
			}
		})
	}
}

func TestGetAWSClient(t *testing.T) {
	// Arrange
	factory := NewFactory()
	// Act
	client := factory.GetAWSClient()
	// Assert
	assert.NotNil(t, client)
}

func TestGetLoggingClient(t *testing.T) {
	// Arrange
	factory := NewFactory()
	// Act
	client := factory.GetLoggingClient()
	// Assert
	assert.NotNil(t, client)
}

func TestGetLoggingClientWithNilLogger(t *testing.T) {
	// Arrange
	factory := NewFactory()
	// Act
	client := factory.GetLoggingClientWithLogger(nil)
	// Assert
	assert.NotNil(t, client)
}

func TestGetLoggingClientWithLogger(t *testing.T) {
	// Arrange
	factory := NewFactory()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	// Act
	client := factory.GetLoggingClientWithLogger(logger)
	// Assert
	assert.NotNil(t, client)
	assert.Equal(t, client.Logger(), logger)
}

func TestGetHelmConfig(t *testing.T) {
	// Arrange
	factory := NewFactory()
	viperInstance := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("bbctl-log-level", "debug")
	viperInstance.Set("kubeconfig", "./test/data/kube-config.yaml")

	fakeCommand := &cobra.Command{
		Use:     "testUse",
		Short:   "testShort",
		Long:    "testLong",
		Example: "testExample",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	// Act
	config, err := factory.getHelmConfig(fakeCommand, "helmconfigtest")
	config.Log("debug") // Required to cover the closure on line 277
	//Assert
	assert.NotNil(t, config)
	assert.Nil(t, err)
}

func TestGetHelmConfigBadConfig(t *testing.T) {
	// Arrange
	factory := NewFactory()
	factory.viperInstance = nil
	// Act
	config, err := factory.getHelmConfig(nil, "helmconfigtest")
	// Assert
	assert.Nil(t, config)
	assert.NotNil(t, err)
	assert.Equal(t, "viper instance is required", err.Error())
}

func TestGetCommandWrapper(t *testing.T) {
	// Arrange
	factory := NewFactory()
	wrapper := factory.GetCommandWrapper("go", "help")
	// Act
	err := wrapper.Run()
	// Assert
	assert.Nil(t, err)
}

func TestGetIstioClientSet(t *testing.T) {
	// Arrange
	factory := NewFactory()
	viperInstance := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "./test/data/kube-config.yaml")
	// Act
	config, configErr := factory.GetRestConfig(nil)
	client, clientErr := factory.GetIstioClientSet(config)
	// Assert
	assert.Nil(t, configErr)
	assert.Nil(t, clientErr)
	assert.NotNil(t, client)
}

func TestGetConfigClient(t *testing.T) {
	// Arrange
	factory := NewFactory()
	// Act
	client, err := factory.GetConfigClient(nil)
	// Assert
	// Actual contents of config are checked in the Client tests
	assert.Nil(t, err)
	assert.NotNil(t, client)
}

func TestGetHelmClient(t *testing.T) {
	// Arrange
	factory := NewFactory()
	viperInstance := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "./test/data/kube-config.yaml")
	// Act
	client, err := factory.GetHelmClient(nil, "foo")
	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, client.GetList)
	assert.NotNil(t, client.GetRelease)
	assert.NotNil(t, client.GetValues)
}

func TestGetHelmClientBadConfig(t *testing.T) {
	// Arrange
	factory := NewFactory()
	viperInstance := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "no-kube-config.yaml")
	// Act
	client, err := factory.GetHelmClient(nil, "foo")
	// Assert
	assert.NotNil(t, err)
	assert.Nil(t, client)
}

func TestGetK8sClientset(t *testing.T) {
	// Arrange
	factory := NewFactory()
	viperInstance := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "./test/data/kube-config.yaml")
	// Act
	client, err := factory.GetK8sClientset(nil)
	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, client)
}

func TestGetK8sClientsetBadConfig(t *testing.T) {
	// Arrange
	factory := NewFactory()
	viperInstance := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "no-kube-config.yaml")
	// Act
	client, err := factory.GetK8sClientset(nil)
	// Assert
	assert.Nil(t, client)
	assert.NotNil(t, err)
	assert.Equal(t, "stat no-kube-config.yaml: no such file or directory", err.Error())
}

func TestGetK8sDynamicClient(t *testing.T) {
	// Arrange
	factory := NewFactory()
	viperInstance := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "./test/data/kube-config.yaml")
	// Act
	client, err := factory.GetK8sDynamicClient(nil)
	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, client)
}

func TestGetK8sDynamicClientMissingConfig(t *testing.T) {
	// Arrange
	factory := NewFactory()
	viperInstance := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "no-kube-config.yaml")
	// Act
	client, err := factory.GetK8sDynamicClient(nil)
	// Assert
	assert.Nil(t, client)
	assert.NotNil(t, err)
	assert.Equal(t, "stat no-kube-config.yaml: no such file or directory", err.Error())
}

func TestGetK8sDynamicClientBadConfig(t *testing.T) {
	// Arrange
	factory := NewFactory()
	factory.viperInstance = nil
	// Act
	client, err := factory.GetK8sDynamicClient(nil)
	// Assert
	assert.Nil(t, client)
	assert.NotNil(t, err)
	assert.Equal(t, "viper instance is required", err.Error())
}

func TestGetRestConfig(t *testing.T) {
	// Arrange
	factory := NewFactory()
	viperInstance := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "./test/data/kube-config.yaml")
	// Act
	config, err := factory.GetRestConfig(nil)
	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, config)
}

func TestGetRestConfigMissingConfig(t *testing.T) {
	// Arrange
	factory := NewFactory()
	viperInstance := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "no-kube-config.yaml")
	// Act
	config, err := factory.GetRestConfig(nil)
	// Assert
	assert.Nil(t, config)
	assert.NotNil(t, err)
	assert.Equal(t, "stat no-kube-config.yaml: no such file or directory", err.Error())
}

func TestGetCommandExecutor(t *testing.T) {
	// Arrange
	factory := NewFactory()
	viperInstance := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "./test/data/kube-config.yaml")
	pod := &coreV1.Pod{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
	}
	var stdout, stderr bytes.Buffer
	// Act
	executor, err := factory.GetCommandExecutor(nil, pod, "foo", []string{"hello"}, &stdout, &stderr)
	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, executor)
}

func TestGetCommandExecutorMissingConfig(t *testing.T) {
	// Arrange
	factory := NewFactory()
	viperInstance := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "no-kube-config.yaml")
	// Act
	executor, err := factory.GetCommandExecutor(nil, nil, "", nil, nil, nil)
	// Assert
	assert.Nil(t, executor)
	assert.NotNil(t, err)
	assert.Equal(t, "stat no-kube-config.yaml: no such file or directory", err.Error())
}

func TestGetCommandExecutorBadConfig(t *testing.T) {
	// Arrange
	factory := NewFactory()
	factory.viperInstance = nil
	// Act
	executor, err := factory.GetCommandExecutor(nil, nil, "", nil, nil, nil)
	// Assert
	assert.Nil(t, executor)
	assert.NotNil(t, err)
	assert.Equal(t, "viper instance is required", err.Error())
}

func TestGetRuntimeClient(t *testing.T) {
	// Arrange
	factory := NewFactory()
	var tests = []struct {
		name             string
		scheme           *runtime.Scheme
		kubeconfig       string
		expectedErrorMsg string
	}{
		{
			name:             "WithNilScheme",
			scheme:           nil,
			kubeconfig:       "./test/data/kube-config.yaml",
			expectedErrorMsg: "",
		},
		{
			name:             "WithValidScheme",
			scheme:           runtime.NewScheme(),
			kubeconfig:       "./test/data/kube-config.yaml",
			expectedErrorMsg: "",
		},
		{
			name:             "WithErrorScheme",
			scheme:           runtime.NewScheme(),
			kubeconfig:       "./test/data/rest-error-kube-config.yaml",
			expectedErrorMsg: "no Auth Provider found for name \"oidc\"",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Act
			os.Setenv("KUBECONFIG", test.kubeconfig)
			client, err := factory.GetRuntimeClient(test.scheme)
			// Assert
			if test.expectedErrorMsg != "" {
				assert.NotNil(t, err)
				assert.Equal(t, err.Error(), test.expectedErrorMsg)
			} else {
				assert.NotNil(t, client)
			}
		})
	}
	// Cleanup
	os.Setenv("KUBECONFIG", "")
}
