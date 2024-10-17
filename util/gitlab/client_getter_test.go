package gitlab

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetClient(t *testing.T) {
	// Arrange
	clientGetter := ClientGetter{}
	// Act
	client, err := clientGetter.GetClient("https://localhost", "")
	// Assert
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestGetBadClientURL(t *testing.T) {
	// Arrange
	clientGetter := ClientGetter{}
	// Act
	client, err := clientGetter.GetClient("%^&", "")
	// Assert
	assert.Nil(t, client)
	require.Error(t, err)
	assert.Equal(t, "parse \"%^&/\": invalid URL escape \"%^&\"", err.Error())
}
