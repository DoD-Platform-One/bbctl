package deploy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"

	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"
)

func TestRoot_NewDeployCmd(t *testing.T) {
	// Arrange
	streams, _, _, _ := genericIOOptions.NewTestIOStreams()
	factory := bbTestUtil.GetFakeFactory()
	// Act
	cmd := NewDeployCmd(factory, streams)
	// Assert
	assert.NotNil(t, cmd)
	assert.Equal(t, "deploy", cmd.Use)
	commandsList := cmd.Commands()
	assert.Len(t, commandsList, 2)
	var commandUseNamesList []string
	for _, command := range commandsList {
		commandUseNamesList = append(commandUseNamesList, command.Use)
	}
	assert.Contains(t, commandUseNamesList, "flux")
	assert.Contains(t, commandUseNamesList, "bigbang")
}

func TestRoot_NewDeployCmd_NoSubcommand(t *testing.T) {
	// Arrange
	streams, in, out, errout := genericIOOptions.NewTestIOStreams()
	factory := bbTestUtil.GetFakeFactory()
	// Act
	cmd := NewDeployCmd(factory, streams)
	assert.Nil(t, cmd.Execute())
	// Assert
	assert.NotNil(t, cmd)
	assert.Equal(t, "deploy", cmd.Use)
	assert.Empty(t, in.String())
	assert.NotEmpty(t, out.String())
	assert.Empty(t, errout.String())
}
