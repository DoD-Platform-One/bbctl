package util

import (
	"bytes"
	"errors"
	"log/slog"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gitlab "github.com/xanzy/go-gitlab"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"
	bbGitlab "repo1.dso.mil/big-bang/product/packages/bbctl/util/gitlab"
)

func TestReadCredentialsFile(t *testing.T) {
	testCases := []struct {
		name                      string
		useDefaultPath            bool
		shouldFail                bool
		shouldErrorOnConfigClient bool
		shouldErrorOnConfig       bool
		shouldErrorOnHomeDir      bool
		shouldErrorOnFindFile     bool
		shouldErrorOnUnmarshal    bool
		shouldErrorOnBadURI       bool
		shouldErrorOnBadComponent bool
		usernameErrorMessage      string
		passwordErrorMessage      string
	}{
		{
			name:                      "should return username and password with custom path",
			useDefaultPath:            false,
			shouldFail:                false,
			shouldErrorOnConfigClient: false,
			shouldErrorOnConfig:       false,
			shouldErrorOnHomeDir:      false,
			shouldErrorOnFindFile:     false,
			shouldErrorOnUnmarshal:    false,
			shouldErrorOnBadURI:       false,
			shouldErrorOnBadComponent: false,
			usernameErrorMessage:      "",
			passwordErrorMessage:      "",
		},
		{
			name:                      "should return empty with default path and no creds",
			useDefaultPath:            true,
			shouldFail:                true,
			shouldErrorOnConfigClient: false,
			shouldErrorOnConfig:       false,
			shouldErrorOnHomeDir:      false,
			shouldErrorOnFindFile:     false,
			shouldErrorOnUnmarshal:    false,
			shouldErrorOnBadURI:       false,
			shouldErrorOnBadComponent: false,
			usernameErrorMessage:      "no credentials found for https://test.com:6443 in",
			passwordErrorMessage:      "no credentials found for https://test.com:6443 in",
		},
		{
			name:                      "should return empty string for invalid component",
			useDefaultPath:            false,
			shouldFail:                true,
			shouldErrorOnConfigClient: false,
			shouldErrorOnConfig:       false,
			shouldErrorOnHomeDir:      false,
			shouldErrorOnFindFile:     false,
			shouldErrorOnUnmarshal:    false,
			shouldErrorOnBadURI:       false,
			shouldErrorOnBadComponent: true,
			usernameErrorMessage:      "invalid component invalidFieldName",
			passwordErrorMessage:      "invalid component invalidFieldName",
		},
		{
			name:                      "should return empty string for invalid URI",
			useDefaultPath:            false,
			shouldFail:                true,
			shouldErrorOnConfigClient: false,
			shouldErrorOnConfig:       false,
			shouldErrorOnHomeDir:      false,
			shouldErrorOnFindFile:     false,
			shouldErrorOnUnmarshal:    false,
			shouldErrorOnBadURI:       true,
			shouldErrorOnBadComponent: false,
			usernameErrorMessage:      "no credentials found for invalidURI in ./test/data/test-credentials.yaml",
			passwordErrorMessage:      "no credentials found for invalidURI in ./test/data/test-credentials.yaml",
		},
		{
			name:                      "should error on config client",
			useDefaultPath:            false,
			shouldFail:                true,
			shouldErrorOnConfigClient: true,
			shouldErrorOnConfig:       false,
			shouldErrorOnHomeDir:      false,
			shouldErrorOnFindFile:     false,
			shouldErrorOnUnmarshal:    false,
			shouldErrorOnBadURI:       false,
			shouldErrorOnBadComponent: false,
			usernameErrorMessage:      "unable to get config client: viper instance is required",
			passwordErrorMessage:      "unable to get config client: viper instance is required",
		},
		{
			name:                      "should error on config",
			useDefaultPath:            false,
			shouldFail:                true,
			shouldErrorOnConfigClient: false,
			shouldErrorOnConfig:       true,
			shouldErrorOnHomeDir:      false,
			shouldErrorOnFindFile:     false,
			shouldErrorOnUnmarshal:    false,
			shouldErrorOnBadURI:       false,
			shouldErrorOnBadComponent: false,
			usernameErrorMessage:      "unable to get client: error during validation for configuration: Key: 'GlobalConfiguration.BigBangRepo' Error:Field validation for 'BigBangRepo' failed on the 'required' tag",
			passwordErrorMessage:      "unable to get client: error during validation for configuration: Key: 'GlobalConfiguration.BigBangRepo' Error:Field validation for 'BigBangRepo' failed on the 'required' tag",
		},
		{
			name:                      "should error on home dir",
			useDefaultPath:            true,
			shouldFail:                true,
			shouldErrorOnConfigClient: false,
			shouldErrorOnConfig:       false,
			shouldErrorOnHomeDir:      true,
			shouldErrorOnFindFile:     false,
			shouldErrorOnUnmarshal:    false,
			shouldErrorOnBadURI:       false,
			shouldErrorOnBadComponent: false,
			usernameErrorMessage:      "unable to get home directory: $HOME is not defined",
			passwordErrorMessage:      "unable to get home directory: $HOME is not defined",
		},
		{
			name:                      "should error on find file",
			useDefaultPath:            false,
			shouldFail:                true,
			shouldErrorOnConfigClient: false,
			shouldErrorOnConfig:       false,
			shouldErrorOnHomeDir:      false,
			shouldErrorOnFindFile:     true,
			shouldErrorOnUnmarshal:    false,
			shouldErrorOnBadURI:       false,
			shouldErrorOnBadComponent: false,
			usernameErrorMessage:      "unable to read credentials file ./test/data/missing-credentials.yaml: open ./test/data/missing-credentials.yaml: no such file or directory",
			passwordErrorMessage:      "unable to read credentials file ./test/data/missing-credentials.yaml: open ./test/data/missing-credentials.yaml: no such file or directory",
		},
		{
			name:                      "should error on unmarshal",
			useDefaultPath:            false,
			shouldFail:                true,
			shouldErrorOnConfigClient: false,
			shouldErrorOnConfig:       false,
			shouldErrorOnHomeDir:      false,
			shouldErrorOnFindFile:     false,
			shouldErrorOnUnmarshal:    true,
			shouldErrorOnBadURI:       false,
			shouldErrorOnBadComponent: false,
			usernameErrorMessage:      "unable to unmarshal credentials file ./test/data/test-credentials.yaml: test failure",
			passwordErrorMessage:      "unable to unmarshal credentials file ./test/data/test-credentials.yaml: test failure",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			factory := NewFactory(nil)
			viperInstance, err := factory.GetViper()
			require.NoError(t, err)
			factory.getViperFunction = func() (*viper.Viper, error) {
				return viperInstance, nil
			}
			viperInstance.Set("big-bang-repo", "test")
			viperInstance.Set("kubeconfig", "./test/data/kube-config.yaml")
			viperInstance.Set("big-bang-credential-helper-credentials-file-path", "./test/data/test-credentials.yaml")
			if tc.useDefaultPath {
				viperInstance.Set("big-bang-credential-helper-credentials-file-path", "")
				homeDir, err := os.UserHomeDir()
				require.NoError(t, err)
				credentialsDir := path.Join(homeDir, ".bbctl")
				credentialsPath := path.Join(credentialsDir, "credentials.yaml")
				if _, err := os.Stat(credentialsPath); err != nil {
					err := os.MkdirAll(credentialsDir, os.ModePerm)
					require.NoError(t, err)
					_, err = os.Create(credentialsPath)
					require.NoError(t, err)
					defer func() {
						err := os.Remove(credentialsPath)
						require.NoError(t, err)
					}()
				}
			}
			if tc.shouldErrorOnConfigClient {
				factory.getViperFunction = func() (*viper.Viper, error) {
					return nil, nil
				}
			}
			if tc.shouldErrorOnConfig {
				viperInstance.Set("big-bang-repo", "")
			}
			if tc.shouldErrorOnHomeDir {
				t.Setenv("HOME", "")
				assert.Empty(t, os.Getenv("HOME"))
			}
			if tc.shouldErrorOnFindFile {
				viperInstance.Set("big-bang-credential-helper-credentials-file-path", "./test/data/missing-credentials.yaml")
			}
			var unmarshallFunc func(in []byte, out interface{}) error
			if tc.shouldErrorOnUnmarshal {
				unmarshallFunc = func(_ []byte, _ interface{}) error {
					return errors.New("test failure")
				}
			}
			host := "https://test.com:6443"
			if tc.shouldErrorOnBadURI {
				host = "invalidURI"
			}
			usernameComponent := "username"
			passwordComponent := "password"
			if tc.shouldErrorOnBadComponent {
				usernameComponent = "invalidFieldName"
				passwordComponent = "invalidFieldName"
			}
			// Act
			var username, password string
			var usernameErr, passwordErr error
			if tc.shouldErrorOnUnmarshal {
				username, usernameErr = factory.readCredentialsFile(usernameComponent, host, unmarshallFunc)
				password, passwordErr = factory.readCredentialsFile(passwordComponent, host, unmarshallFunc)
			} else {
				username, usernameErr = factory.ReadCredentialsFile(usernameComponent, host)
				password, passwordErr = factory.ReadCredentialsFile(passwordComponent, host)
			}
			// Assert
			if tc.shouldFail {
				assert.Empty(t, username)
				assert.Empty(t, password)
				require.Error(t, usernameErr)
				require.Error(t, passwordErr)
				assert.Contains(t, usernameErr.Error(), tc.usernameErrorMessage)
				assert.Contains(t, passwordErr.Error(), tc.passwordErrorMessage)
			} else {
				require.NoError(t, usernameErr)
				require.NoError(t, passwordErr)
				assert.Equal(t, "username", username)
				assert.Equal(t, "password", password)
			}
		})
	}
}

func TestGetCredentialHelper(t *testing.T) {
	var tests = []struct {
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
		{
			name:             "GetConfigError",
			credentialHelper: "credentials-file",
			field:            "username",
			expected:         "",
			error:            "unable to get client: error during validation for configuration: Key: 'GlobalConfiguration.BigBangRepo' Error:Field validation for 'BigBangRepo' failed on the 'required' tag",
		},
		{
			name:             "BadCredentialsFilePath",
			credentialHelper: "credentials-file",
			field:            "username",
			expected:         "",
			error:            "unable to read credentials file: unable to read credentials file ./test/data/missing-credentials.yaml: open ./test/data/missing-credentials.yaml: no such file or directory",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Arrange
			factory := NewFactory(nil)
			viperInstance, err := factory.GetViper()
			require.NoError(t, err)
			factory.getViperFunction = func() (*viper.Viper, error) {
				return viperInstance, nil
			}
			viperInstance.Set("big-bang-repo", "test")
			viperInstance.Set("kubeconfig", "./test/data/kube-config.yaml")
			viperInstance.Set("big-bang-credential-helper-credentials-file-path", "./test/data/test-credentials.yaml")
			if test.name == "BadCredentialsFilePath" {
				viperInstance.Set("big-bang-credential-helper-credentials-file-path", "./test/data/missing-credentials.yaml")
			}
			viperInstance.Set("big-bang-credential-helper", test.credentialHelper)
			uri := "https://test.com:6443"
			if test.error != "" {
				uri = "https://invalidUri.com:6443"
			}
			// Force the viper instance to be nil to cause this to error downstream
			if test.name == "GetConfigClientError" {
				factory.getViperFunction = func() (*viper.Viper, error) {
					return nil, nil
				}
			}
			if test.name == "GetConfigError" {
				viperInstance.Set("big-bang-repo", "")
			}
			// Act
			helper, err := factory.GetCredentialHelper()
			require.NoError(t, err)
			value, err := helper(test.field, uri)
			// Assert
			if test.error != "" {
				assert.Equal(t, test.error, err.Error())
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, test.expected, strings.TrimSuffix(value, "\n"))
		})
	}
}

func TestGetAWSClient(t *testing.T) {
	// Arrange
	factory := NewFactory(nil)
	// Act
	var client interface{}
	var err error
	client, err = factory.GetAWSClient()
	// Assert
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestGetLoggingClient(t *testing.T) {
	// Arrange
	factory := NewFactory(nil)
	// Act
	client, err := factory.GetLoggingClient()
	// Assert
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestGetLoggingClientWithNilLogger(t *testing.T) {
	// Arrange
	factory := NewFactory(nil)
	// Act
	client, err := factory.GetLoggingClientWithLogger(nil)
	// Assert
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestGetLoggingClientWithLogger(t *testing.T) {
	// Arrange
	factory := NewFactory(nil)
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	// Act
	client, err := factory.GetLoggingClientWithLogger(logger)
	// Assert
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, client.Logger(), logger)
}

func TestGetHelmConfig(t *testing.T) {
	// Arrange
	factory := NewFactory(nil)
	viperInstance, err := factory.GetViper()
	require.NoError(t, err)
	factory.getViperFunction = func() (*viper.Viper, error) {
		return viperInstance, nil
	}
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("bbctl-log-level", "debug")
	viperInstance.Set("kubeconfig", "./test/data/kube-config.yaml")

	fakeCommand := &cobra.Command{
		Use:     "testUse",
		Short:   "testShort",
		Long:    "testLong",
		Example: "testExample",
		RunE: func(_ *cobra.Command, _ []string) error {
			return nil
		},
	}
	// Act
	config, err := factory.getHelmConfig(fakeCommand, "helmconfigtest")
	config.Log("debug") // Required to cover the closure on line 277
	// Assert
	assert.NotNil(t, config)
	require.NoError(t, err)
}

func TestGetHelmConfigBadConfig(t *testing.T) {
	// Arrange
	factory := NewFactory(nil)
	factory.getViperFunction = func() (*viper.Viper, error) {
		return nil, nil
	}
	// Act
	config, err := factory.getHelmConfig(nil, "helmconfigtest")
	// Assert
	assert.Nil(t, config)
	require.Error(t, err)
	assert.Equal(t, "viper instance is required", err.Error())
}

func TestGetCommandWrapper(t *testing.T) {
	// Arrange
	factory := NewFactory(nil)
	wrapper, err := factory.GetCommandWrapper("go", "help")
	require.NoError(t, err)
	// Act				factory.getViperFunction = func() (*viper.Viper, error) {
	err = wrapper.Run()
	// Assert
	require.NoError(t, err)
}

func TestGetIstioClientSet(t *testing.T) {
	// Arrange
	factory := NewFactory(nil)
	viperInstance, err := factory.GetViper()
	require.NoError(t, err)
	factory.getViperFunction = func() (*viper.Viper, error) {
		return viperInstance, nil
	}
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "./test/data/kube-config.yaml")
	// Act
	config, configErr := factory.GetRestConfig(nil)
	client, clientErr := factory.GetIstioClientSet(config)
	// Assert
	require.NoError(t, configErr)
	require.NoError(t, clientErr)
	assert.NotNil(t, client)
}

func TestGetConfigClient(t *testing.T) {
	testCases := []struct {
		name                 string
		shouldFail           bool
		errorOnLoggingClient bool
		errorOnGetViper      bool
		errorOnGetClient     bool
		expectedErrorMessage string
	}{
		{
			name:                 "should not error",
			shouldFail:           false,
			errorOnLoggingClient: false,
			errorOnGetViper:      false,
			errorOnGetClient:     false,
			expectedErrorMessage: "",
		},
		{
			name:                 "should error on get client",
			shouldFail:           true,
			errorOnLoggingClient: false,
			errorOnGetViper:      false,
			errorOnGetClient:     true,
			expectedErrorMessage: "viper instance is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			factory := NewFactory(nil)
			if tc.errorOnGetClient {
				factory.getViperFunction = func() (*viper.Viper, error) {
					return nil, nil
				}
			}
			// Act
			client, err := factory.GetConfigClient(nil)
			// Assert
			if tc.shouldFail {
				assert.Nil(t, client)
				require.Error(t, err)
				assert.Equal(t, tc.expectedErrorMessage, err.Error())
			} else {
				// Actual contents of config are checked in the Client tests
				require.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

func TestGetGitLabClient(t *testing.T) {
	testCases := []struct {
		name                string
		errorOnConfigClient bool
		errorOnGetConfig    bool
		errorOnGetGitLab    bool
		callTopLevel        bool
	}{
		{
			name:                "should not error",
			errorOnConfigClient: false,
			errorOnGetConfig:    false,
			errorOnGetGitLab:    false,
			callTopLevel:        true,
		},
		{
			name:                "should error on config client",
			errorOnConfigClient: true,
			errorOnGetConfig:    false,
			errorOnGetGitLab:    false,
			callTopLevel:        true,
		},
		{
			name:                "should error on get config",
			errorOnConfigClient: false,
			errorOnGetConfig:    true,
			errorOnGetGitLab:    false,
			callTopLevel:        true,
		},
		{
			name:                "should error on get gitlab",
			errorOnConfigClient: false,
			errorOnGetConfig:    false,
			errorOnGetGitLab:    true,
			callTopLevel:        false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			factory := NewFactory(nil)
			v, _ := factory.GetViper()
			factory.getViperFunction = func() (*viper.Viper, error) {
				return v, nil
			}
			v.Set("big-bang-repo", "test")
			v.Set("base-url", "https://gitlab.com")
			v.Set("access-token", "test")
			var clientOptionsFuncs []gitlab.ClientOptionFunc
			if tc.errorOnConfigClient {
				factory.getViperFunction = func() (*viper.Viper, error) {
					return nil, nil
				}
			}
			if tc.errorOnGetConfig {
				v.Set("big-bang-repo", "")
			}
			if tc.errorOnGetGitLab {
				clientOptionsFuncs = append(clientOptionsFuncs, func(_ *gitlab.Client) error {
					return errors.New("error")
				})
			}
			var client bbGitlab.Client
			var err error
			// Act
			if tc.callTopLevel {
				client, err = factory.GetGitLabClient()
			} else {
				client, err = factory.getGitLabClient(clientOptionsFuncs...)
			}

			// Assert
			if tc.errorOnConfigClient || tc.errorOnGetConfig || tc.errorOnGetGitLab {
				assert.Nil(t, client)
				require.Error(t, err)
			} else {
				assert.NotNil(t, client)
				require.NoError(t, err)
			}
		})
	}
}

func TestGetHelmClient(t *testing.T) {
	testCases := []struct {
		name                 string
		shouldFail           bool
		errorOnConfigClient  bool
		errorOnConfig        bool
		errorOnLoggingClient bool
		errorOnKubeConfig    bool
		expectedErrorMessage string
	}{
		{
			name:                 "should not error",
			shouldFail:           false,
			errorOnConfigClient:  false,
			errorOnConfig:        false,
			errorOnLoggingClient: false,
			errorOnKubeConfig:    false,
			expectedErrorMessage: "",
		},
		{
			name:                 "should error on config client",
			shouldFail:           true,
			errorOnConfigClient:  true,
			errorOnConfig:        false,
			errorOnLoggingClient: false,
			errorOnKubeConfig:    false,
			expectedErrorMessage: "viper instance is required",
		},
		{
			name:                 "should error on config",
			shouldFail:           true,
			errorOnConfigClient:  false,
			errorOnConfig:        true,
			errorOnLoggingClient: false,
			errorOnKubeConfig:    false,
			expectedErrorMessage: "unable to get client: error during validation for configuration: Key: 'GlobalConfiguration.BigBangRepo' Error:Field validation for 'BigBangRepo' failed on the 'required' tag",
		},
		{
			name:                 "should error on kube config",
			shouldFail:           true,
			errorOnConfigClient:  false,
			errorOnConfig:        false,
			errorOnLoggingClient: false,
			errorOnKubeConfig:    true,
			expectedErrorMessage: "stat no-kube-config.yaml: no such file or directory",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			factory := NewFactory(nil)
			viperInstance, err := factory.GetViper()
			require.NoError(t, err)
			factory.getViperFunction = func() (*viper.Viper, error) {
				return viperInstance, nil
			}
			viperInstance.Set("big-bang-repo", "test")
			viperInstance.Set("kubeconfig", "./test/data/kube-config.yaml")
			if tc.errorOnConfigClient {
				factory.getViperFunction = func() (*viper.Viper, error) {
					return nil, nil
				}
			}
			if tc.errorOnConfig {
				viperInstance.Set("big-bang-repo", "")
			}
			if tc.errorOnKubeConfig {
				viperInstance.Set("kubeconfig", "no-kube-config.yaml")
			}
			// Act
			client, err := factory.GetHelmClient(nil, "foo")
			// Assert
			if tc.shouldFail {
				assert.Nil(t, client)
				require.Error(t, err)
				assert.Equal(t, tc.expectedErrorMessage, err.Error())
			} else {
				require.NoError(t, err)
				assert.NotNil(t, client.GetList)
				assert.NotNil(t, client.GetRelease)
				assert.NotNil(t, client.GetValues)
			}
		})
	}
}

func TestGetK8sClientset(t *testing.T) {
	// Arrange
	factory := NewFactory(nil)
	viperInstance, err := factory.GetViper()
	require.NoError(t, err)
	factory.getViperFunction = func() (*viper.Viper, error) {
		return viperInstance, nil
	}
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "./test/data/kube-config.yaml")
	// Act
	client, err := factory.GetK8sClientset(nil)
	// Assert
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestGetK8sClientsetBadConfig(t *testing.T) {
	// Arrange
	factory := NewFactory(nil)
	viperInstance, err := factory.GetViper()
	require.NoError(t, err)
	factory.getViperFunction = func() (*viper.Viper, error) {
		return viperInstance, nil
	}
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "no-kube-config.yaml")
	// Act
	client, err := factory.GetK8sClientset(nil)
	// Assert
	assert.Nil(t, client)
	require.Error(t, err)
	assert.Equal(t, "stat no-kube-config.yaml: no such file or directory", err.Error())
}

func TestGetK8sDynamicClient(t *testing.T) {
	testCases := []struct {
		name                 string
		shouldFail           bool
		errorOnConfigClient  bool
		errorOnConfig        bool
		errorOnMissingConfig bool
		errorOnBadConfig     bool
		expectedErrorMessage string
	}{
		{
			name:                 "should not error",
			shouldFail:           false,
			errorOnConfigClient:  false,
			errorOnConfig:        false,
			errorOnMissingConfig: false,
			errorOnBadConfig:     false,
			expectedErrorMessage: "",
		},
		{
			name:                 "should error on config client",
			shouldFail:           true,
			errorOnConfigClient:  true,
			errorOnConfig:        false,
			errorOnMissingConfig: false,
			errorOnBadConfig:     false,
			expectedErrorMessage: "viper instance is required",
		},
		{
			name:                 "should error on config",
			shouldFail:           true,
			errorOnConfigClient:  false,
			errorOnConfig:        true,
			errorOnMissingConfig: false,
			errorOnBadConfig:     false,
			expectedErrorMessage: "unable to get client: error during validation for configuration: Key: 'GlobalConfiguration.BigBangRepo' Error:Field validation for 'BigBangRepo' failed on the 'required' tag",
		},
		{
			name:                 "should error on missing config",
			shouldFail:           true,
			errorOnConfigClient:  false,
			errorOnConfig:        false,
			errorOnMissingConfig: true,
			errorOnBadConfig:     false,
			expectedErrorMessage: "stat no-kube-config.yaml: no such file or directory",
		},
		{
			name:                 "should error on bad config",
			shouldFail:           true,
			errorOnConfigClient:  false,
			errorOnConfig:        false,
			errorOnMissingConfig: false,
			errorOnBadConfig:     true,
			expectedErrorMessage: "no Auth Provider found for name \"oidc\"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			factory := NewFactory(nil)
			viperInstance, err := factory.GetViper()
			require.NoError(t, err)
			factory.getViperFunction = func() (*viper.Viper, error) {
				return viperInstance, nil
			}
			viperInstance.Set("big-bang-repo", "test")
			viperInstance.Set("kubeconfig", "./test/data/kube-config.yaml")
			if tc.errorOnConfigClient {
				factory.getViperFunction = func() (*viper.Viper, error) {
					return nil, nil
				}
			}
			if tc.errorOnConfig {
				viperInstance.Set("big-bang-repo", "")
			}
			if tc.errorOnMissingConfig {
				viperInstance.Set("kubeconfig", "no-kube-config.yaml")
			}
			if tc.errorOnBadConfig {
				viperInstance.Set("kubeconfig", "./test/data/rest-error-kube-config.yaml")
			}
			// Act
			client, err := factory.GetK8sDynamicClient(nil)
			// Assert
			if tc.shouldFail {
				assert.Nil(t, client)
				require.Error(t, err)
				assert.Equal(t, tc.expectedErrorMessage, err.Error())
			} else {
				require.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

func TestGetOutputClient(t *testing.T) {
	// Define test cases using a table-driven approach
	tests := []struct {
		name                      string
		outputFormat              string
		shouldFail                bool
		shouldErrorOnIOStreams    bool
		shouldErrorOnConfigClient bool
		shouldErrorOnConfig       bool
		expectedErrorMessage      string
	}{
		{
			name:                      "Test JSON output",
			outputFormat:              "json",
			shouldFail:                false,
			shouldErrorOnIOStreams:    false,
			shouldErrorOnConfigClient: false,
			shouldErrorOnConfig:       false,
			expectedErrorMessage:      "",
		},
		{
			name:                      "Test text output",
			outputFormat:              "text",
			shouldFail:                false,
			shouldErrorOnIOStreams:    false,
			shouldErrorOnConfigClient: false,
			shouldErrorOnConfig:       false,
			expectedErrorMessage:      "",
		},
		{
			name:                      "Test YAML output",
			outputFormat:              "yaml",
			shouldFail:                false,
			shouldErrorOnIOStreams:    false,
			shouldErrorOnConfigClient: false,
			shouldErrorOnConfig:       false,
			expectedErrorMessage:      "",
		},
		{
			name:                      "Should error on config client",
			outputFormat:              "json",
			shouldFail:                true,
			shouldErrorOnIOStreams:    false,
			shouldErrorOnConfigClient: true,
			shouldErrorOnConfig:       false,
			expectedErrorMessage:      "viper instance is required",
		},
		{
			name:                      "Should error on config",
			outputFormat:              "json",
			shouldFail:                true,
			shouldErrorOnIOStreams:    false,
			shouldErrorOnConfigClient: false,
			shouldErrorOnConfig:       true,
			expectedErrorMessage:      "error during validation for configuration: Key: 'GlobalConfiguration.BigBangRepo' Error:Field validation for 'BigBangRepo' failed on the 'required' tag",
		},
	}

	// Iterate through test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			factory := NewFactory(nil)
			fakeCommand := &cobra.Command{
				Use:     "testUse",
				Short:   "testShort",
				Long:    "testLong",
				Example: "testExample",
				RunE: func(_ *cobra.Command, _ []string) error {
					return nil
				},
			}
			// Set the "output" format using Viper
			viperInstance, err := factory.GetViper()
			require.NoError(t, err)
			factory.getViperFunction = func() (*viper.Viper, error) {
				return viperInstance, nil
			}
			viperInstance.Set("big-bang-repo", "test")
			viperInstance.Set("output", tc.outputFormat)
			if tc.shouldErrorOnConfigClient {
				factory.getViperFunction = func() (*viper.Viper, error) {
					return nil, nil
				}
			}
			if tc.shouldErrorOnConfig {
				viperInstance.Set("big-bang-repo", "")
			}
			// Act
			client, err := factory.GetOutputClient(fakeCommand)
			// Assert
			if tc.shouldFail {
				assert.Nil(t, client)
				require.Error(t, err)
				assert.Equal(t, tc.expectedErrorMessage, err.Error())
			} else {
				require.NoError(t, err)
				assert.NotNil(t, client)
				assert.NotNil(t, client.Output)
			}
		})
	}
}

func TestGetRestConfig(t *testing.T) {
	testCases := []struct {
		name                 string
		shouldFail           bool
		errorOnConfigClient  bool
		errorOnConfig        bool
		errorOnMissingConfig bool
		errorOnBadConfig     bool
		expectedErrorMessage string
	}{
		{
			name:                 "should not error",
			shouldFail:           false,
			errorOnConfigClient:  false,
			errorOnConfig:        false,
			errorOnMissingConfig: false,
			errorOnBadConfig:     false,
			expectedErrorMessage: "",
		},
		{
			name:                 "should error on config client",
			shouldFail:           true,
			errorOnConfigClient:  true,
			errorOnConfig:        false,
			errorOnMissingConfig: false,
			errorOnBadConfig:     false,
			expectedErrorMessage: "viper instance is required",
		},
		{
			name:                 "should error on config",
			shouldFail:           true,
			errorOnConfigClient:  false,
			errorOnConfig:        true,
			errorOnMissingConfig: false,
			errorOnBadConfig:     false,
			expectedErrorMessage: "unable to get client: error during validation for configuration: Key: 'GlobalConfiguration.BigBangRepo' Error:Field validation for 'BigBangRepo' failed on the 'required' tag",
		},
		{
			name:                 "should error on missing config",
			shouldFail:           true,
			errorOnConfigClient:  false,
			errorOnConfig:        false,
			errorOnMissingConfig: true,
			errorOnBadConfig:     false,
			expectedErrorMessage: "stat no-kube-config.yaml: no such file or directory",
		},
		{
			name:                 "should error on bad config",
			shouldFail:           true,
			errorOnConfigClient:  false,
			errorOnConfig:        false,
			errorOnMissingConfig: false,
			errorOnBadConfig:     true,
			expectedErrorMessage: "invalid configuration: [context was not found for specified context: bad, cluster has no server defined]",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			factory := NewFactory(nil)
			viperInstance, err := factory.GetViper()
			require.NoError(t, err)
			factory.getViperFunction = func() (*viper.Viper, error) {
				return viperInstance, nil
			}
			viperInstance.Set("big-bang-repo", "test")
			viperInstance.Set("kubeconfig", "./test/data/kube-config.yaml")
			if tc.errorOnConfigClient {
				factory.getViperFunction = func() (*viper.Viper, error) {
					return nil, nil
				}
			}
			if tc.errorOnConfig {
				viperInstance.Set("big-bang-repo", "")
			}
			if tc.errorOnMissingConfig {
				viperInstance.Set("kubeconfig", "no-kube-config.yaml")
			}
			if tc.errorOnBadConfig {
				viperInstance.Set("kubeconfig", "./test/data/bad-kube-config.yaml")
			}
			// Act
			config, err := factory.GetRestConfig(nil)
			// Assert
			if tc.shouldFail {
				assert.Nil(t, config)
				require.Error(t, err)
				assert.Equal(t, tc.expectedErrorMessage, err.Error())
			} else {
				require.NoError(t, err)
				assert.NotNil(t, config)
			}
		})
	}
}

func TestGetCommandExecutor(t *testing.T) {
	// Arrange
	factory := NewFactory(nil)
	viperInstance, err := factory.GetViper()
	require.NoError(t, err)
	factory.getViperFunction = func() (*viper.Viper, error) {
		return viperInstance, nil
	}
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
	require.NoError(t, err)
	assert.NotNil(t, executor)
}

func TestGetCommandExecutorMissingConfig(t *testing.T) {
	// Arrange
	factory := NewFactory(nil)
	viperInstance, err := factory.GetViper()
	require.NoError(t, err)
	factory.getViperFunction = func() (*viper.Viper, error) {
		return viperInstance, nil
	}
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "no-kube-config.yaml")
	// Act
	executor, err := factory.GetCommandExecutor(nil, nil, "", nil, nil, nil)
	// Assert
	assert.Nil(t, executor)
	require.Error(t, err)
	assert.Equal(t, "stat no-kube-config.yaml: no such file or directory", err.Error())
}

func TestGetCommandExecutorBadConfig(t *testing.T) {
	// Arrange
	factory := NewFactory(nil)
	factory.getViperFunction = func() (*viper.Viper, error) {
		return nil, nil
	}
	// Act
	executor, err := factory.GetCommandExecutor(nil, nil, "", nil, nil, nil)
	// Assert
	assert.Nil(t, executor)
	require.Error(t, err)
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
			t.Setenv("KUBECONFIG", test.kubeconfig)
			client, err := factory.GetRuntimeClient(test.scheme)
			// Assert
			if test.expectedErrorMsg != "" {
				require.Error(t, err)
				assert.Equal(t, test.expectedErrorMsg, err.Error())
			} else {
				assert.NotNil(t, client)
			}
		})
	}
	// Cleanup
	t.Setenv("KUBECONFIG", "")
}

func TestGetIOStreams(t *testing.T) {
	// Arrange
	factory := NewFactory(nil)
	var err error
	var ios *genericIOOptions.IOStreams
	ios, err = factory.GetIOStream()
	// Assert
	require.NoError(t, err)
	assert.Equal(t, os.Stdin, ios.In)
	assert.Equal(t, os.Stdout, ios.Out)
	assert.Equal(t, os.Stderr, ios.ErrOut)
}

func TestGetPipe(t *testing.T) {
	// Arrange
	factory := NewFactory(nil)

	// Act
	r, w, err := factory.GetPipe()

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, r)
	assert.NotNil(t, w)

	// Verify that the pipe works by writing to the writer and reading from the reader
	testMessage := "Hello, Pipe!"
	go func() {
		w.Write([]byte(testMessage))
		w.Close()
	}()

	buf := new(bytes.Buffer)
	_, readError := buf.ReadFrom(r)
	require.NoError(t, readError)
	assert.Equal(t, testMessage, buf.String())
}
