package ironbank

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var credentialHelper = func(component, uri string) (string, error) {
	switch component {
	case "username":
		return "testuser", nil
	case "password":
		return "testpass", nil
	default:
		return "", fmt.Errorf("unknown credential component: %s", component)
	}
}

func TestGetClient(t *testing.T) {
	// Arrange
	clientGetter := ClientGetter{}
	// Act
	client, err := clientGetter.GetClient(credentialHelper)
	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, client)
}
