package aws

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientGetter_GetClient(t *testing.T) {
	// Arrange
	clientGetter := ClientGetter{}

	// Act
	client, err := clientGetter.GetClient()

	// Assert
	assert.NotNil(t, client)
	assert.Nil(t, err)
}
