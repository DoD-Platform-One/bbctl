package k3d

import (
	"testing"

	"github.com/stretchr/testify/assert"
	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"

	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"
)

func TestK3d_RootUsage(t *testing.T) {
	// Arrange
	streams, _, _, _ := genericIOOptions.NewTestIOStreams()
	factory := bbTestUtil.GetFakeFactory()
	// Act
	cmd := NewK3dCmd(factory, streams)
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
	streams, in, out, errout := genericIOOptions.NewTestIOStreams()
	factory := bbTestUtil.GetFakeFactory()
	// Act
	cmd := NewK3dCmd(factory, streams)
	// Assert
	assert.Nil(t, cmd.Execute())
	assert.Empty(t, in.String())
	assert.Empty(t, errout.String())
	assert.Contains(t, out.String(), "Please provide a subcommand for k3d (see help)")
}
