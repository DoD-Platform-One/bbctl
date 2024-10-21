package static

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadConstants(t *testing.T) {
	// Arrange & Act
	c, err := GetDefaultConstants()
	// Assert
	require.NoError(t, err)
	assert.NotNil(t, c)
}

func TestAssertConstants(t *testing.T) {
	// Arrange & Act
	c, err := GetDefaultConstants()
	// Assert
	require.NoError(t, err)
	assert.NotNil(t, c)
	assert.Equal(t, "bigbang", c.BigBangHelmReleaseName)
	assert.Equal(t, "bigbang", c.BigBangNamespace)
	assert.Equal(t, "0.7.5", c.BigBangCliVersion)
}

func TestErrorConstants(t *testing.T) {
	// Arrange
	c, err := GetDefaultConstants()
	require.NoError(t, err)
	// Act
	c.readFileFunc = func(_ string) ([]byte, error) {
		return nil, assert.AnError
	}
	err = c.readConstants()
	// Assert
	require.Error(t, err)
}
