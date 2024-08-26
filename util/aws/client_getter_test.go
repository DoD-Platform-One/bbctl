package aws

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientGetter_GetClient(t *testing.T) {
	// Arrange
	clientGetter := ClientGetter{}

	// Act
	client := clientGetter.GetClient()

	// Assert
	assert.NotNil(t, client)
}
