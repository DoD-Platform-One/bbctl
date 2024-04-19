package aws

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	bbUtilTestLog "repo1.dso.mil/big-bang/product/packages/bbctl/util/test/log"
)

func TestClientGetter_GetClient(t *testing.T) {
	// Arrange
	clientGetter := ClientGetter{}
	var stringBuilder strings.Builder
	logFunc := func(args ...string) {
		for _, arg := range args {
			stringBuilder.WriteString(arg)
		}
	}
	loggingClient := bbUtilTestLog.NewFakeClient(logFunc)
	// Act
	client, err := clientGetter.GetClient(loggingClient)
	// Assert
	assert.NotNil(t, client)
	assert.Nil(t, err)
	assert.Empty(t, stringBuilder.String())
}
