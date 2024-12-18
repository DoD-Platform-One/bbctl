package filesystem

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetClient tests the GetClient returns a non-nil client
func TestGetClient(t *testing.T) {
	// Arrange
	clientGetter := ClientGetter{}
	// Act
	client := clientGetter.GetClient()
	// Assert
	assert.NotNil(t, client)
}
