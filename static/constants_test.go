package static

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadConstants(t *testing.T) {
	// Arrange & Act
	c, err := GetDefaultConstants()
	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, c)
}

func TestAssertConstants(t *testing.T) {
	// Arrange & Act
	c, err := GetDefaultConstants()
	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, c)
	assert.Equal(t, "bigbang", c.BigBangHelmReleaseName)
	assert.Equal(t, "bigbang", c.BigBangNamespace)
	assert.Equal(t, "0.7.2", c.BigBangCliVersion)
}

func TestErrorConstants(t *testing.T) {
	// Arrange
	c, err := GetDefaultConstants()
	assert.Nil(t, err)
	// Act
	c.readFileFunc = func(s string) ([]byte, error) {
		return nil, assert.AnError
	}
	err = c.readConstants()
	// Assert
	assert.NotNil(t, err)
}
