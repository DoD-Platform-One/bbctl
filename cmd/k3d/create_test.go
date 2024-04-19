package k3d

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"

	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"
)

func TestK3d_NewCreateClusterCmd(t *testing.T) {
	// Arrange
	streams, _, _, _ := genericIOOptions.NewTestIOStreams()
	factory := bbTestUtil.GetFakeFactory()
	// Act
	cmd := NewCreateClusterCmd(factory, streams)
	// Assert
	assert.NotNil(t, cmd)
	assert.Equal(t, "create", cmd.Use)
}

func TestK3d_NewCreateClusterCmd_RunWithMissingBigBangRepo(t *testing.T) {
	// Arrange
	streams, _, _, _ := genericIOOptions.NewTestIOStreams()
	factory := bbTestUtil.GetFakeFactory()
	// Act
	cmd := NewCreateClusterCmd(factory, streams)
	assert.Panics(t, func() { assert.Nil(t, cmd.Execute()) })
	// Assert
	assert.NotNil(t, cmd)
	assert.Equal(t, "create", cmd.Use)
}

func TestK3d_NewCreateClusterCmd_Run(t *testing.T) {
	// Arrange
	streams, in, out, errout := genericIOOptions.NewTestIOStreams()
	factory := bbTestUtil.GetFakeFactory()
	bigBangRepoLocation := "/tmp/big-bang"
	viper.Set("big-bang-repo", bigBangRepoLocation)
	expectedCmdString := "Running command: /tmp/big-bang/docs/assets/scripts/developer/k3d-dev.sh \n"
	// Act
	cmd := NewCreateClusterCmd(factory, streams)
	assert.Nil(t, cmd.Execute())
	// Assert
	assert.NotNil(t, cmd)
	assert.Equal(t, "create", cmd.Use)
	assert.Empty(t, errout.String())
	assert.Empty(t, in.String())
	assert.Equal(t, expectedCmdString, out.String())
}
