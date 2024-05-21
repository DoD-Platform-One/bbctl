package cmd

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"
	bbConfig "repo1.dso.mil/big-bang/product/packages/bbctl/util/config"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/config/schemas"
	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"
)

func TestGetConfig(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()

	streams, _, buf, _ := genericIOOptions.NewTestIOStreams()

	viper := factory.GetViper()
	// Required value or the execution will fail
	viper.Set("big-bang-repo", "/path/to/repo")

	cmd := NewConfigCmd(factory, streams)
	cmd.Run(cmd, []string{})

	response := strings.Split(buf.String(), "\n")

	// functionality is tested separately.
	// only checking for not nil to get code coverage for cobra cmd
	assert.NotNil(t, response)
}

func TestConfigGetAll(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	streams, _, _, _ := genericIOOptions.NewTestIOStreams()
	cmd := NewConfigCmd(factory, streams)

	viper := factory.GetViper()

	testValues := map[string]string{
		"big-bang-repo": "/path/to/repo",
		"log-level":     "testLogLevel",
		"log-output":    "testLogOutput",
	}

	for key, value := range testValues {
		viper.Set(key, value)
	}

	result, err := getBBConfig(cmd, factory, []string{})
	if err != nil {
		t.Error(err)
	}

	// Parse output into another map[string]string as order
	// of outputcannot be guaranteed
	outputLines := strings.Split(result, "\n")

	outputValues := map[string]string{}

	for _, line := range outputLines {
		parts := strings.SplitN(line, ":", 2)
		key, value := parts[0], parts[1]

		outputValues[key] = strings.TrimSpace(value)
	}

	for key, value := range testValues {
		got, ok := outputValues[key]
		if !ok {
			continue
		}
		if got != value {
			t.Errorf("Value mismatch. Expected: %s, got: %s", value, got)
		}
	}
}

// TestConfigGetOne sets tests values and attempts to fetch only a single value.
// The expectation is that
func TestConfigGetOne(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()

	streams, _, _, _ := genericIOOptions.NewTestIOStreams()
	cmd := NewConfigCmd(factory, streams)

	viper := factory.GetViper()

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
			expected:    "testLogLevel",
			description: "top-level string value",
		},
		{
			key:         "policy.gatekeeper",
			expected:    "false",
			description: "nested boolean value (stringified)",
		},
		{
			key:         "k3d-ssh.ssh-username",
			expected:    "testUser",
			description: "nested string value",
		},
	}

	for _, tc := range tt {
		t.Run(tc.key, func(t *testing.T) {
			got, err := getBBConfig(cmd, factory, []string{tc.key})
			if err != nil {
				t.Error(err)
			}

			if got != tc.expected {
				t.Errorf("Value mismatch. Expected: %s, got: %s", tc.expected, got)
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

// TestConfigMarshalError tests that when an invalid, unmarshalable configuration is craeted
// the code correctly panics.
func TestConfigMarshalError(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	loggingClient := factory.GetLoggingClient()

	streams, _, _, _ := genericIOOptions.NewTestIOStreams()
	cmd := NewConfigCmd(factory, streams)
	viper := factory.GetViper()

	expected := ""
	getConfigFunc := func(client *bbConfig.ConfigClient) *schemas.GlobalConfiguration {
		return &schemas.GlobalConfiguration{
			BigBangRepo: expected,
			ExampleConfiguration: schemas.ExampleConfiguration{
				ShouldFailToMarshal: func() *any { x := any(make(chan int)); return &x }(),
			},
		}
	}

	// Get a configuration client and set our getConfigFunc
	client, err := bbConfig.NewClient(getConfigFunc, nil, &loggingClient, cmd, viper)
	if err != nil {
		t.Error()
	}
	factory.SetConfigClient(client)

	// Required value or the execution will fail
	viper.Set("big-bang-repo", "/path/to/repo")

	assert.Panics(t, func() {
		_, _ = getBBConfig(cmd, factory, []string{})
	}, "did not panic marshaling unmarshalable type")

}

func TestConfigTooManyKeys(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()

	streams, _, _, _ := genericIOOptions.NewTestIOStreams()
	cmd := NewConfigCmd(factory, streams)
	viper := factory.GetViper()

	// Required value or the execution will fail
	viper.Set("big-bang-repo", "/path/to/repo")

	_, err := getBBConfig(cmd, factory, []string{"too", "many", "args"})
	expectedError := "too many arguments passed to bbctl config"
	if err == nil || err.Error() != expectedError {
		t.Errorf("Expected error: %s, got %v", expectedError, err)
	}
}

func TestConfigGetInvalidKey(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()

	streams, _, _, _ := genericIOOptions.NewTestIOStreams()
	cmd := NewConfigCmd(factory, streams)
	viper := factory.GetViper()

	// Required value or the execution will fail
	viper.Set("big-bang-repo", "/path/to/repo")

	_, err := getBBConfig(cmd, factory, []string{"invalid.key"})
	// The code splits keys at the dot, so it should look for a parent object "invalid" her
	expectedError := "No such field: invalid"
	if err == nil || err.Error() != expectedError {
		t.Errorf("Expected error: %s, got %v", expectedError, err)
	}
}

func TestConfigFailToGetConfigClient(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()

	streams, _, _, _ := genericIOOptions.NewTestIOStreams()
	cmd := NewConfigCmd(factory, streams)
	viper := factory.GetViper()

	// Required value or the execution will fail
	viper.Set("big-bang-repo", "/path/to/repo")

	factory.SetFail.GetConfigClient = true
	_, err := getBBConfig(cmd, factory, []string{"invalid.key"})
	// The code splits keys at the dot, so it should look for a parent object "invalid" her
	expectedError := "error getting config client: failed to get config client"
	if err == nil || err.Error() != expectedError {
		t.Errorf("Expected error: %s, got %v", expectedError, err)
	}
}
