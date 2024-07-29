package deploy

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"
)

func TestRoot_NewDeployCmd(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	// Act
	cmd, _ := NewDeployCmd(factory)
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
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	streams := factory.GetIOStream()
	in := streams.In.(*bytes.Buffer)
	out := streams.Out.(*bytes.Buffer)
	errOut := streams.ErrOut.(*bytes.Buffer)
	// Act
	cmd, _ := NewDeployCmd(factory)
	assert.Nil(t, cmd.Execute())
	// Assert
	assert.NotNil(t, cmd)
	assert.Equal(t, "deploy", cmd.Use)
	assert.Empty(t, in.String())
	assert.NotEmpty(t, out.String())
	assert.Empty(t, errOut.String())
}
