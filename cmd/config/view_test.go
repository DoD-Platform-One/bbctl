package config

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
	bbConfig "repo1.dso.mil/big-bang/product/packages/bbctl/util/config"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/config/schemas"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/output"
	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"
)

func TestGetConfigView(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	streams, _ := factory.GetIOStream()
	buf := streams.Out.(*bytes.Buffer)

	viper, _ := factory.GetViper()
	// Required value or the execution will fail
	viper.Set("big-bang-repo", "/path/to/repo")
	viper.Set("output-config.format", "yaml")

	cmd := NewConfigViewCmd(factory)
	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	response := strings.Split(buf.String(), "\n")

	// functionality is tested separately.
	// only checking for not nil to get code coverage for cobra cmd
	assert.NotNil(t, response)
}

func TestConfigGetAll(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	cmd := NewConfigViewCmd(factory)

	viper, _ := factory.GetViper()

	testValues := map[string]any{
		"big-bang-repo": "/path/to/repo",
		"log-level":     "testLogLevel",
		"log-output":    "testLogOutput",
		// The type [any]any is required here since the yaml unmarshaller erases the string type for nested keys
		"output-config": map[any]any{
			"format": "yaml",
		},
	}

	for key, value := range testValues {
		viper.Set(key, value)
	}

	err := getBBConfig(cmd, factory, []string{})
	if err != nil {
		t.Error(err)
	}

	streams, streamsErr := factory.GetIOStream()
	require.NoError(t, streamsErr)

	out := streams.Out.(*bytes.Buffer)

	// Parse output into another map[string]string as order
	// of outputcannot be guaranteed
	outputMap := make(map[string]any)

	yamlErr := yaml.Unmarshal(out.Bytes(), outputMap)
	require.NoError(t, yamlErr)

	for key, value := range testValues {
		got, ok := outputMap[key]
		if !ok {
			continue
		}

		if !assert.Equal(t, value, got) {
			t.Errorf("Value mismatch. Expected: %s, got: %s", value, got)
		}
	}
}

// TestConfigViewGetOne sets tests values and attempts to fetch only a single value.
// The expectation is that
func TestConfigViewGetOne(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	cmd := NewConfigViewCmd(factory)

	viper, _ := factory.GetViper()

	testValues := map[string]any{
		"big-bang-repo":    "/path/to/repo",
		"bbctl-log-level":  "testLogLevel",
		"bbctl-log-output": "testLogOutput",
		"policy": map[string]bool{
			"gatekeeper": false,
		},
		"k3d-ssh": map[string]string{
			"ssh-username": "testUser",
		},
		"output-config": map[string]string{
			"format": "yaml",
		},
	}

	for key, value := range testValues {
		viper.Set(key, value)
	}

	tt := []struct {
		key         string
		expected    string
		description string
		shouldErr   bool
	}{
		{
			key:         "bbctl-log-level",
			expected:    "bbctl-log-level: testLogLevel",
			description: "top-level string value",
		},
		{
			key:         "policy.gatekeeper",
			expected:    "policy.gatekeeper: \"false\"",
			description: "nested boolean value (stringified)",
		},
		{
			key:         "k3d-ssh.ssh-username",
			expected:    "k3d-ssh.ssh-username: testUser",
			description: "nested string value",
		},
	}

	for _, tc := range tt {
		t.Run(tc.key, func(t *testing.T) {
			factory.ResetIOStream()
			streams, _ := factory.GetIOStream()
			out := streams.Out.(*bytes.Buffer)

			err := getBBConfig(cmd, factory, []string{tc.key})
			if err != nil {
				t.Error(err)
			}

			output := strings.Trim(out.String(), "\n\t ")
			if output != tc.expected {
				t.Errorf("Value mismatch. Expected: %s, got: %s", tc.expected, output)
			}
		})
	}
}

// Test findRecurisve if called with n empty slice
func TestFindRecursiveNoKeys(t *testing.T) {
	in := []string{}

	result, err := findRecursive(reflect.Value{}, in)
	if result != "" {
		t.Errorf("Expected an zero-value string, got: %s", result)
	}
	expectedError := "invalid key"
	if err == nil || err.Error() != expectedError {
		t.Errorf("Expected error: %s, got %v", expectedError, err)
	}
}

// TestConfigMarshalError tests that when an invalid, unmarshalable configuration is created
// the code correctly panics.
func TestConfigViewMarshalError(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	loggingClient, _ := factory.GetLoggingClient()

	cmd := NewConfigViewCmd(factory)
	viper, _ := factory.GetViper()

	expected := ""
	getConfigFunc := func(_ *bbConfig.ConfigClient) (*schemas.GlobalConfiguration, error) {
		return &schemas.GlobalConfiguration{
			BigBangRepo:         expected,
			OutputConfiguration: schemas.OutputConfiguration{Format: output.YAML},
			ExampleConfiguration: schemas.ExampleConfiguration{
				ShouldFailToMarshal: func() *any { x := any(make(chan int)); return &x }(),
			},
		}, nil
	}

	// Get a configuration client and set our getConfigFunc
	client, err := bbConfig.NewClient(getConfigFunc, nil, &loggingClient, cmd, viper)
	if err != nil {
		t.Error()
	}
	factory.SetConfigClient(client)

	// Set required values or the execution will fail
	viper.Set("big-bang-repo", "/path/to/repo")
	viper.Set("output-config.format", "yaml")

	assert.Panics(t, func() {
		err = getBBConfig(cmd, factory, []string{})
		require.NoError(t, err)
	}, "did not panic marshaling unmarshalable type %w", err)
}

func TestConfigTooManyKeys(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	cmd := NewConfigViewCmd(factory)
	viper, _ := factory.GetViper()

	// Required value or the execution will fail
	viper.Set("big-bang-repo", "/path/to/repo")

	err := getBBConfig(cmd, factory, []string{"too", "many", "args"})
	expectedError := "too many arguments passed to bbctl config"
	if err == nil || err.Error() != expectedError {
		t.Errorf("Expected error: %s, got %v", expectedError, err)
	}
}

func TestConfigOutputClientError(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	factory.SetFail.GetIOStreams = 1

	cmd := NewConfigViewCmd(factory)
	viper, _ := factory.GetViper()

	// Required value or the execution will fail
	viper.Set("big-bang-repo", "/path/to/repo")

	err := getBBConfig(cmd, factory, []string{})
	expectedError := "error getting output client: failed to get streams"
	if err == nil || err.Error() != expectedError {
		t.Errorf("Expected error: %s, got %v", expectedError, err)
	}
}

func TestGlobalConfigFormat(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()

	cmd := NewConfigViewCmd(factory)
	viper, _ := factory.GetViper()
	streams, err := factory.GetIOStream()
	require.NoError(t, err)

	// Required value or the execution will fail
	viper.Set("big-bang-repo", "/path/to/repo")

	err = getBBConfig(cmd, factory, []string{})
	require.NoError(t, err)
	expectedStdOut := "BigBangRepo:\"/path/to/repo\","
	assert.Contains(t, streams.Out.(*bytes.Buffer).String(), expectedStdOut)
}

func TestSingleConfigFormat(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()

	cmd := NewConfigViewCmd(factory)
	viper, _ := factory.GetViper()
	streams, err := factory.GetIOStream()
	require.NoError(t, err)

	// Required value or the execution will fail
	viper.Set("big-bang-repo", "/path/to/repo")
	viper.Set("bbctl-log-level", "testLogLevel")

	err = getBBConfig(cmd, factory, []string{"bbctl-log-level"})
	require.NoError(t, err)
	expectedStdOut := "bbctl-log-level:testLogLevel"
	assert.Contains(t, streams.Out.(*bytes.Buffer).String(), expectedStdOut)
}

func TestConfigGetInvalidKey(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()

	cmd := NewConfigViewCmd(factory)
	viper, _ := factory.GetViper()

	// Required value or the execution will fail
	viper.Set("big-bang-repo", "/path/to/repo")

	err := getBBConfig(cmd, factory, []string{"invalid.key"})
	// The code splits keys at the dot, so it should look for a parent object "invalid" her
	expectedError := "error marshaling specific config: no such field: invalid"
	if err == nil || err.Error() != expectedError {
		t.Errorf("Expected error: %s, got %v", expectedError, err)
	}
}

func TestConfigFailToGetConfigClient(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()

	cmd := NewConfigViewCmd(factory)
	viper, _ := factory.GetViper()

	// Required value or the execution will fail
	viper.Set("big-bang-repo", "/path/to/repo")

	factory.SetFail.GetConfigClient = 1
	err := cmd.RunE(cmd, []string{"invalid.key"})

	// The code splits keys at the dot, so it should look for a parent object "invalid" her
	expectedError := "error getting config client: failed to get config client"
	if err == nil || err.Error() != expectedError {
		t.Errorf("Expected error: %s, got %v", expectedError, err)
	}
}

func TestConfigFailToGetConfig(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	loggingClient, _ := factory.GetLoggingClient()
	cmd := NewConfigViewCmd(factory)
	viper, _ := factory.GetViper()
	expected := ""
	getConfigFunc := func(_ *bbConfig.ConfigClient) (*schemas.GlobalConfiguration, error) {
		return &schemas.GlobalConfiguration{
			BigBangRepo: expected,
		}, errors.New("dummy error")
	}
	client, err := bbConfig.NewClient(getConfigFunc, nil, &loggingClient, cmd, viper)
	if err != nil {
		t.Error()
	}
	factory.SetConfigClient(client)

	// Act
	outputErr := getBBConfig(cmd, factory, []string{})

	// Assert
	require.Error(t, outputErr)
	if !assert.Contains(t, outputErr.Error(), "error getting config:") {
		t.Errorf("unexpected output: %s", outputErr.Error())
	}
}
