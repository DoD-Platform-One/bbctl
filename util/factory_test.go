package util

import (
	"bytes"
	"fmt"
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
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/output"
)

// TestReadDefaultCredentialsFileMissing tests that a missing credentials file returns an error
func TestReadDefaultCredentialsFileMissing(t *testing.T) {
	// Arrange
	factory := NewFactory(nil)
	viperInstance, err := factory.GetViper()
	assert.Nil(t, err)
	// Set the big-bang-repo and kubeconfig to local test files to avoid reading the default credentials file
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "./test/data/kube-config.yaml")

	homeDir, _ := os.UserHomeDir()
	credsDir := path.Join(homeDir, ".bbctl")
	os.Rename(path.Join(credsDir, "credentials.yaml"), path.Join(credsDir, "old-credentials.yaml"))

	// Act & Assert
	value, err := factory.ReadCredentialsFile("", "")
	assert.Equal(t, "", value)
	assert.NotNil(t, err)
	expectedError := fmt.Sprintf(
		"unable to read credentials file %s: open %s: no such file or directory",
		path.Join(credsDir, "credentials.yaml"),
		path.Join(credsDir, "credentials.yaml"),
	)
	assert.Equal(t, expectedError, err.Error())

	// Cleanup
	os.Rename(path.Join(credsDir, "old-credentials.yaml"), path.Join(credsDir, "credentials.yaml"))
}

func TestReadCredentialsFile(t *testing.T) {
	// Arrange
	factory := NewFactory(nil)
	viperInstance, err := factory.GetViper()
	assert.Nil(t, err)
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "./test/data/kube-config.yaml")
	viperInstance.Set(
		"big-bang-credential-helper-credentials-file-path",
		"./test/data/test-credentials.yaml",
	)

	// Act
	// Test reading valid components
	username, err := factory.ReadCredentialsFile("username", "https://test.com:6443")
	if err != nil {
		t.Errorf("unexpected error getting username: %v", err)
	}
	password, err := factory.ReadCredentialsFile("password", "https://test.com:6443")
	if err != nil {
		t.Errorf("unexpected error getting password: %v", err)
	}

	// Assert
	assert.Equal(t, username, "username")
	assert.Equal(t, password, "password")

	// Test reading an invalid component
	invalid, err := factory.ReadCredentialsFile("invalidFieldName", "https://test.com:6443")
	assert.Equal(t, "", invalid)
	assert.NotNil(t, err)
	assert.Equal(t, "invalid component invalidFieldName", err.Error())

	// Test reading an invalid URI
	invalidURI, err := factory.ReadCredentialsFile("username", "invalidURI")
	assert.Equal(t, "", invalidURI)
	assert.NotNil(t, err)
	assert.Equal(
		t,
		"no credentials found for invalidURI in ./test/data/test-credentials.yaml",
		err.Error(),
	)

	// Force the viper instance to be nil to cause this to error downstream
	factory.SetViper(nil)
	username, err = factory.ReadCredentialsFile("username", "https://test.com:6443")
	assert.Empty(t, username)
	assert.NotNil(t, err)
	assert.Equal(t, "unable to get config client: viper instance is required", err.Error())
}

func TestGetCredentialHelper(t *testing.T) {
	// Arrange
	factory := NewFactory(nil)
	viperInstance, err := factory.GetViper()
	assert.Nil(t, err)
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "./test/data/kube-config.yaml")
	viperInstance.Set(
		"big-bang-credential-helper-credentials-file-path",
		"./test/data/test-credentials.yaml",
	)

	tests := []struct {
		name             string
		credentialHelper string
		field            string
		expected         string
		error            string
	}{
		{
			name:             "EmptyCredsHelper",
			credentialHelper: "",
			field:            "username",
			expected:         "",
			error:            "no credential helper defined (\"big-bang-credential-helper\")",
		},
		{
			name:             "CustomCredsHelperNonEmpty",
			credentialHelper: "./../scripts/factory-tests/fake-credentials-helper.sh",
			field:            "username",
			expected:         "username",
			error:            "",
		},
		{
			name:             "CustomCredsHelperEmpty",
			credentialHelper: "./../scripts/factory-tests/fake-credentials-helper.sh",
			field:            "password",
			expected:         "",
			error:            "no password found for https://invalidUri.com:6443 in ./../scripts/factory-tests/fake-credentials-helper.sh",
		},
		{
			name:             "InvalidCredsField",
			credentialHelper: "./../scripts/factory-tests/fake-credentials-helper.sh",
			field:            "invalidCredsField",
			expected:         "",
			error:            "unable to get invalidCredsField from https://invalidUri.com:6443 using ./../scripts/factory-tests/fake-credentials-helper.sh: exit status 1",
		},
		{
			name:             "InvalidUriForDefaultHelper",
			credentialHelper: "credentials-file",
			field:            "username",
			expected:         "username",
			error:            "",
		},
		{
			name:             "ValidUriForDefaultHelper",
			credentialHelper: "credentials-file",
			field:            "username",
			expected:         "username",
			error:            "",
		},
		{
			name:             "GetConfigClientError",
			credentialHelper: "credentials-file",
			field:            "username",
			expected:         "",
			error:            "unable to get config client: viper instance is required",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Act
			viperInstance.Set("big-bang-credential-helper", test.credentialHelper)
			helper, err := factory.GetCredentialHelper()
			assert.Nil(t, err)

			uri := "https://test.com:6443"
			if test.error != "" {
				uri = "https://invalidUri.com:6443"
			}

			// Force the viper instance to be nil to cause this to error downstream
			if test.name == "GetConfigClientError" {
				factory.SetViper(nil)
			}

			value, err := helper(test.field, uri)

			// Assert
			if test.error != "" {
				assert.Equal(t, test.error, err.Error())
			} else {
				assert.Nil(t, err)
			}

			assert.Equal(t, test.expected, strings.TrimSuffix(value, "\n"))
		})
	}
}

func TestGetCredentialHelperMissingFilePath(t *testing.T) {
	// Arrange
	factory := NewFactory(nil)
	viperInstance, err := factory.GetViper()
	assert.Nil(t, err)
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "./test/data/kube-config.yaml")
	viperInstance.Set("big-bang-credential-helper", "credentials-file")

	homeDir, _ := os.UserHomeDir()
	credsDir := path.Join(homeDir, ".bbctl")
	credsPath := path.Join(credsDir, "credentials.yaml")
	os.Rename(path.Join(credsDir, "credentials.yaml"), path.Join(credsDir, "old-credentials.yaml"))

	// Act
	helper, err := factory.GetCredentialHelper()
	assert.Nil(t, err)
	username, err := helper("username", "https://test.com:6443")

	// Assert
	assert.NotNil(t, helper)
	assert.Empty(t, username)
	assert.NotNil(t, err)
	assert.Equal(
		t,
		fmt.Sprintf(
			"unable to read credentials file: unable to read credentials file %s: open %s: no such file or directory",
			credsPath,
			credsPath,
		),
		err.Error(),
	)

	// Cleanup
	os.Rename(path.Join(credsDir, "old-credentials.yaml"), path.Join(credsDir, "credentials.yaml"))
}

func TestGetAWSClient(t *testing.T) {
	// Arrange
	factory := NewFactory(nil)
	// Act
	client, err := factory.GetAWSClient()
	if err != nil {
		t.Errorf("unexpected error getting AWS client: %v", err)
	}
	// Assert
	assert.NotNil(t, client)
}

func TestGetLoggingClient(t *testing.T) {
	// Arrange
	factory := NewFactory(nil)
	// Act
	client, err := factory.GetLoggingClient()
	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, client)
}

func TestGetLoggingClientWithNilLogger(t *testing.T) {
	// Arrange
	factory := NewFactory(nil)
	// Act
	client, err := factory.GetLoggingClientWithLogger(nil)
	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, client)
}

func TestGetLoggingClientWithLogger(t *testing.T) {
	// Arrange
	factory := NewFactory(nil)
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	// Act
	client, err := factory.GetLoggingClientWithLogger(logger)
	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, client.Logger(), logger)
}

func TestGetHelmConfig(t *testing.T) {
	// Arrange
	factory := NewFactory(nil)
	viperInstance, err := factory.GetViper()
	assert.Nil(t, err)
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
	// Assert
	assert.NotNil(t, config)
	assert.Nil(t, err)
}

func TestGetHelmConfigBadConfig(t *testing.T) {
	// Arrange
	factory := NewFactory(nil)
	factory.SetViper(nil)
	// Act
	config, err := factory.getHelmConfig(nil, "helmconfigtest")
	// Assert
	assert.Nil(t, config)
	assert.NotNil(t, err)
	assert.Equal(t, "viper instance is required", err.Error())
}

func TestGetCommandWrapper(t *testing.T) {
	// Arrange
	factory := NewFactory(nil)
	wrapper, err := factory.GetCommandWrapper("go", "help")
	assert.Nil(t, err)
	// Act
	err = wrapper.Run()
	// Assert
	assert.Nil(t, err)
}

func TestGetIstioClientSet(t *testing.T) {
	// Arrange
	factory := NewFactory(nil)
	viperInstance, err := factory.GetViper()
	assert.Nil(t, err)
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
	factory := NewFactory(nil)
	// Act
	client, err := factory.GetConfigClient(nil)
	// Assert
	// Actual contents of config are checked in the Client tests
	assert.Nil(t, err)
	assert.NotNil(t, client)
}

func TestGetHelmClient(t *testing.T) {
	// Arrange
	factory := NewFactory(nil)
	viperInstance, err := factory.GetViper()
	assert.Nil(t, err)
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
	factory := NewFactory(nil)
	viperInstance, err := factory.GetViper()
	assert.Nil(t, err)
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
	factory := NewFactory(nil)
	viperInstance, err := factory.GetViper()
	assert.Nil(t, err)
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
	factory := NewFactory(nil)
	viperInstance, err := factory.GetViper()
	assert.Nil(t, err)
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
	factory := NewFactory(nil)
	viperInstance, err := factory.GetViper()
	assert.Nil(t, err)
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
	factory := NewFactory(nil)
	viperInstance, err := factory.GetViper()
	assert.Nil(t, err)
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
	factory := NewFactory(nil)
	factory.SetViper(nil)
	// Act
	client, err := factory.GetK8sDynamicClient(nil)
	// Assert
	assert.Nil(t, client)
	assert.NotNil(t, err)
	assert.Equal(t, "viper instance is required", err.Error())
}

func TestGetRestConfig(t *testing.T) {
	// Arrange
	factory := NewFactory(nil)
	viperInstance, err := factory.GetViper()
	assert.Nil(t, err)
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
	factory := NewFactory(nil)
	viperInstance, err := factory.GetViper()
	assert.Nil(t, err)
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
	factory := NewFactory(nil)
	viperInstance, err := factory.GetViper()
	assert.Nil(t, err)
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
	executor, err := factory.GetCommandExecutor(
		nil,
		pod,
		"foo",
		[]string{"hello"},
		&stdout,
		&stderr,
	)
	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, executor)
}

func TestGetCommandExecutorMissingConfig(t *testing.T) {
	// Arrange
	factory := NewFactory(nil)
	viperInstance, err := factory.GetViper()
	assert.Nil(t, err)
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
	factory := NewFactory(nil)
	factory.SetViper(nil)
	// Act
	executor, err := factory.GetCommandExecutor(nil, nil, "", nil, nil, nil)
	// Assert
	assert.Nil(t, executor)
	assert.NotNil(t, err)
	assert.Equal(t, "viper instance is required", err.Error())
}

func TestGetRuntimeClient(t *testing.T) {
	// Arrange
	factory := NewFactory(nil)
	tests := []struct {
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

func TestGetIOStreams(t *testing.T) {
	// Arrange
	factory := NewFactory(nil)

	// Act
	ios, err := factory.GetIOStream()

	// Assert
	assert.Nil(t, err)
	assert.Equal(t, os.Stdin, ios.In)
	assert.Equal(t, os.Stdout, ios.Out)
	assert.Equal(t, os.Stderr, ios.ErrOut)
}

// TestGetOuputClient tests the GetOutputClient function using table-driven tests.
func TestGetOutputClient(t *testing.T) {
	// Define test cases using a table-driven approach
	tests := []struct {
		name         string
		outputFormat string
	}{
		{
			name:         "Test JSON output",
			outputFormat: "json",
		},
		{
			name:         "Test text output",
			outputFormat: "text",
		},
		{
			name:         "Test YAML output",
			outputFormat: "yaml",
		},
	}

	// Arrange
	factory := NewFactory(nil)
	fakeCommand := &cobra.Command{
		Use:     "testUse",
		Short:   "testShort",
		Long:    "testLong",
		Example: "testExample",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	// Iterate through test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Set the "output" format using Viper
			viperInstance, err := factory.GetViper()
			assert.Nil(t, err)
			viperInstance.Set("big-bang-repo", "test")
			viperInstance.Set("output", tc.outputFormat)

			// Act
			client, err := factory.GetOutputClient(fakeCommand)

			// Assert
			assert.Nil(t, err)
			assert.NotNil(t, client)

			// Check client output
			outputClient, ok := client.(output.Client)
			assert.True(t, ok, "Expected client to be of type output.Client")
			assert.NotNil(t, outputClient.Output)
		})
	}
}

func TestCreatePipe(t *testing.T) {
	// Arrange
	factory := NewFactory(nil)

	// Act
	err := factory.CreatePipe()

	// Assert
	assert.Nil(t, err)

	r, w := factory.GetPipe()
	assert.NotNil(t, r)
	assert.NotNil(t, w)

	// Verify that the pipe works by writing to the writer and reading from the reader
	testMessage := "Hello, Pipe!"
	go func() {
		w.Write([]byte(testMessage))
		w.Close()
	}()

	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	assert.Equal(t, testMessage, buf.String())
}

func TestSetPipe(t *testing.T) {
	// Arrange
	factory := NewFactory(nil)
	r1, w1, err := os.Pipe()
	assert.Nil(t, err)

	// Act
	factory.SetPipe(r1, w1)

	r2, w2 := factory.GetPipe()

	// Assert
	assert.Equal(t, r1, r2)
	assert.Equal(t, w1, w2)

	// Verify that the pipe set works correctly by writing and reading
	testMessage := "Testing SetPipe"
	go func() {
		w2.Write([]byte(testMessage))
		w2.Close()
	}()

	buf := new(bytes.Buffer)
	buf.ReadFrom(r2)
	assert.Equal(t, testMessage, buf.String())
}

func TestGetPipeWithoutCreate(t *testing.T) {
	// Arrange
	factory := NewFactory(nil)

	// Act
	r, w := factory.GetPipe()

	// Assert
	assert.Nil(t, r)
	assert.Nil(t, w)
}
