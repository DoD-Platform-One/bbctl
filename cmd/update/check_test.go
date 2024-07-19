package update

import (
	"testing"

	"github.com/stretchr/testify/assert"
	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"

	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"
)

func TestUpdate_CheckUsage(t *testing.T) {
	// Arrange
	streams, _, _, _ := genericIOOptions.NewTestIOStreams()
	factory := bbTestUtil.GetFakeFactory()
	// Act
	cmd := NewUpdateCheckCmd(factory, streams)
	// Assert
	assert.NotNil(t, cmd)
	assert.Equal(t, "check", cmd.Use)
}

func TestUpdateCheck(t *testing.T) {
	// Arrange
	streams, _, _, _ := genericIOOptions.NewTestIOStreams()
	factory := bbTestUtil.GetFakeFactory()
	// Act
	cmd := NewUpdateCheckCmd(factory, streams)
	// Assert
	var args []string
	err := cmd.RunE(cmd, args)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Not Implemented")
}
