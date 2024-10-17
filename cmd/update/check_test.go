package update

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"
)

func TestUpdate_CheckUsage(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	// Act
	cmd := NewUpdateCheckCmd(factory)
	// Assert
	assert.NotNil(t, cmd)
	assert.Equal(t, "check", cmd.Use)
}

func TestUpdateCheck(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	// Act
	cmd := NewUpdateCheckCmd(factory)
	// Assert
	var args []string
	err := cmd.RunE(cmd, args)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")
}
