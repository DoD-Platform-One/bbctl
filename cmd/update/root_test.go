package update

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"
)

func TestUpdate_RootUsage(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	// Act
	cmd, err := NewUpdateCmd(factory)
	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, cmd)
	assert.Equal(t, "update", cmd.Use)
	commandsList := cmd.Commands()
	assert.Len(t, commandsList, 1)
	var commandUseNamesList []string
	for _, command := range commandsList {
		commandUseNamesList = append(commandUseNamesList, command.Use)
	}
	assert.Contains(t, commandUseNamesList, "check")
}

func TestUpdate_RootNoSubcommand(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	streams, _ := factory.GetIOStream()
	in := streams.In.(*bytes.Buffer)
	out := streams.Out.(*bytes.Buffer)
	errOut := streams.ErrOut.(*bytes.Buffer)
	// Act
	cmd, err := NewUpdateCmd(factory)
	// Assert
	assert.Nil(t, err)
	assert.Nil(t, cmd.Execute())
	assert.Empty(t, in.String())
	assert.Empty(t, errOut.String())
	assert.Contains(t, out.String(), "Please provide a subcommand for update (see help)")
}
