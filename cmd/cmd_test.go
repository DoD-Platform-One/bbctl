package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"
)

func TestGetHelpUsage(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	// Act
	cmd, _ := NewRootCmd(factory)
	// Assert
	assert.NotNil(t, cmd)
	assert.Equal(t, "bbctl", cmd.Use)
	assert.Contains(t, cmd.Example, "bbctl help")

	assert.False(t, cmd.CompletionOptions.DisableDefaultCmd)
	assert.False(t, cmd.CompletionOptions.DisableDescriptions)
	assert.True(t, cmd.CompletionOptions.DisableNoDescFlag)

	commandsList := cmd.Commands()
	assert.Len(t, commandsList, 12)
	var commandUseNamesList []string
	for _, command := range commandsList {
		commandUseNamesList = append(commandUseNamesList, command.Use)
	}
	assert.Contains(t, commandUseNamesList, "completion [bash|zsh|fish|powershell]")
	assert.Contains(t, commandUseNamesList, "config")
	assert.Contains(t, commandUseNamesList, "deploy")
	assert.Contains(t, commandUseNamesList, "k3d")
	assert.Contains(t, commandUseNamesList, "update")
	assert.Contains(t, commandUseNamesList, "list")
	assert.Contains(t, commandUseNamesList, "policy --PROVIDER CONSTRAINT_NAME")
	assert.Contains(t, commandUseNamesList, "preflight-check")
	assert.Contains(t, commandUseNamesList, "status")
	assert.Contains(t, commandUseNamesList, "values RELEASE_NAME")
	assert.Contains(t, commandUseNamesList, "version")
	assert.Contains(t, commandUseNamesList, "violations")
}

func TestNewRootCmdErrors(t *testing.T) {
	testCases := []struct {
		name              string
		errorOnCompletion bool
		errorOnVersion    bool
		errorOnViolations bool
		errorOnPolicies   bool
		errorOnPreflight  bool
		errorOnK3d        bool
		errorOnDeploy     bool
		errorOnUpdate     bool
		expectedError     string
	}{
		{
			name:              "error on completion",
			errorOnCompletion: true,
			expectedError:     "unable to get IO streams",
		},
		{
			name:           "error on version",
			errorOnVersion: true,
			expectedError:  "unable to get config client",
		},
		{
			name:              "error on violations",
			errorOnViolations: true,
			expectedError:     "unable to get config client",
		},
		{
			name:            "error on policies",
			errorOnPolicies: true,
			expectedError:   "unable to get config client",
		},
		{
			name:             "error on preflight",
			errorOnPreflight: true,
			expectedError:    "unable to get config client",
		},
		{
			name:          "error on k3d",
			errorOnK3d:    true,
			expectedError: "unable to get config client",
		},
		{
			name:          "error on deploy",
			errorOnDeploy: true,
			expectedError: "unable to get config client",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			factory := bbTestUtil.GetFakeFactory()
			if tc.errorOnCompletion {
				factory.SetFail.GetIOStreams = 1
			}
			if tc.errorOnVersion {
				factory.SetFail.GetConfigClient = 1
			}
			if tc.errorOnViolations {
				factory.SetFail.GetConfigClient = 2
			}
			if tc.errorOnPolicies {
				factory.SetFail.GetConfigClient = 3
			}
			if tc.errorOnPreflight {
				factory.SetFail.GetConfigClient = 4
			}
			if tc.errorOnK3d {
				factory.SetFail.GetConfigClient = 5
			}
			if tc.errorOnDeploy {
				factory.SetFail.GetConfigClient = 7
			}
			if tc.errorOnUpdate {
				factory.SetFail.GetIOStreams = 4
			}
			// Act
			cmd, err := NewRootCmd(factory)
			// Assert
			assert.Nil(t, cmd)
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), tc.expectedError)
		})
	}
}
