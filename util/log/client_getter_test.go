package log

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetClient(t *testing.T) {
	// Arrange
	clientGetter := ClientGetter{}
	assert.NotNil(t, clientGetter)
	// Act
	client := clientGetter.GetClient(nil)
	// Assert
	assert.NotNil(t, client)
	assert.NotNil(t, client.Logger())
	assert.Equal(t, client.Logger(), slog.Default())
}
