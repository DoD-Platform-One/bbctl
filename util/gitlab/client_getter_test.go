package gitlab

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetClient(t *testing.T) {
	// Arrange
	clientGetter := ClientGetter{}
	// Act
	client, err := clientGetter.GetClient("https://localhost", "")
	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, client)
}

func TestGetBadClientURL(t *testing.T) {
	// Arrange
	clientGetter := ClientGetter{}
	// Act
	client, err := clientGetter.GetClient("%^&", "")
	// Assert
	assert.Nil(t, client)
	assert.NotNil(t, err)
	assert.Equal(t, "parse \"%^&/\": invalid URL escape \"%^&\"", err.Error())
}
