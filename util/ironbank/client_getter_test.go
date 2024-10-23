package ironbank

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var credentialHelper = func(component, _ string) (string, error) {
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
	clientGetter := ClientGetter{}
	client, err := clientGetter.GetClient(credentialHelper)
	require.NoError(t, err)
	assert.NotNil(t, client)
}
