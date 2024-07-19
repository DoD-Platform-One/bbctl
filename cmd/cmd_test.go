package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"

	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"
)

func TestGetHelpUsage(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	streams, _, _, _ := genericIOOptions.NewTestIOStreams()
	// Act
	cmd, _ := NewRootCmd(factory, streams)
	// Assert
	assert.NotNil(t, cmd)
	assert.Equal(t, "bbctl", cmd.Use)
	assert.Contains(t, cmd.Example, "bbctl help")

	assert.False(t, cmd.CompletionOptions.DisableDefaultCmd)
	assert.False(t, cmd.CompletionOptions.DisableDescriptions)
	assert.True(t, cmd.CompletionOptions.DisableNoDescFlag)

	commandsList := cmd.Commands()
	assert.Len(t, commandsList, 11)
	var commandUseNamesList []string
	for _, command := range commandsList {
		commandUseNamesList = append(commandUseNamesList, command.Use)
	}
	assert.Contains(t, commandUseNamesList, "completion [bash|zsh|fish|powershell]")
	assert.Contains(t, commandUseNamesList, "config [key]")
	assert.Contains(t, commandUseNamesList, "deploy")
	assert.Contains(t, commandUseNamesList, "k3d")
	assert.Contains(t, commandUseNamesList, "list")
	assert.Contains(t, commandUseNamesList, "policy --PROVIDER CONSTRAINT_NAME")
	assert.Contains(t, commandUseNamesList, "preflight-check")
	assert.Contains(t, commandUseNamesList, "status")
	assert.Contains(t, commandUseNamesList, "values RELEASE_NAME")
	assert.Contains(t, commandUseNamesList, "version")
	assert.Contains(t, commandUseNamesList, "violations")
}
