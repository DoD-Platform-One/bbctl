package k3d

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"
)

func TestK3d_RootUsage(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	// Act
	cmd, _ := NewK3dCmd(factory)
	// Assert
	assert.NotNil(t, cmd)
	assert.Equal(t, "k3d", cmd.Use)
	commandsList := cmd.Commands()
	assert.Len(t, commandsList, 5)
	var commandUseNamesList []string
	for _, command := range commandsList {
		commandUseNamesList = append(commandUseNamesList, command.Use)
	}
	assert.Contains(t, commandUseNamesList, "create")
	assert.Contains(t, commandUseNamesList, "destroy")
	assert.Contains(t, commandUseNamesList, "hosts")
	assert.Contains(t, commandUseNamesList, "shellprofile")
	assert.Contains(t, commandUseNamesList, "ssh")
}

func TestK3d_RootNoSubcommand(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	streams := factory.GetIOStream()
	in := streams.In.(*bytes.Buffer)
	out := streams.Out.(*bytes.Buffer)
	errOut := streams.ErrOut.(*bytes.Buffer)
	// Act
	cmd, _ := NewK3dCmd(factory)
	// Assert
	assert.Nil(t, cmd.Execute())
	assert.Empty(t, in.String())
	assert.Empty(t, errOut.String())
	assert.Contains(t, out.String(), "Please provide a subcommand for k3d (see help)")
}

func TestK3d_RootSshError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	factory.SetFail.GetConfigClient = true

	// Act
	cmd, err := NewK3dCmd(factory)

	// Assert
	assert.Nil(t, cmd)
	if !assert.Contains(t, err.Error(), "Error retrieving ssh Command:") {
		t.Errorf("unexpected output: %s", err.Error())
	}
}
