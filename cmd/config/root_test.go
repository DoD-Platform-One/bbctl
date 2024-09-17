package config

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"
)

func TestRoot_NewConfigCmd(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	// Act
	cmd := NewConfigCmd(factory)
	// Assert
	assert.NotNil(t, cmd)
	assert.Equal(t, "config", cmd.Use)
	commandsList := cmd.Commands()
	assert.Len(t, commandsList, 2)
	var commandUseNamesList []string
	for _, command := range commandsList {
		commandUseNamesList = append(commandUseNamesList, command.Use)
	}
	assert.Contains(t, commandUseNamesList, "view [key]")
	assert.Contains(t, commandUseNamesList, "init")
}

func TestRoot_NewConfigCmd_NoSubcommand(t *testing.T) {
	testCases := []struct {
		name             string
		errorOnGetClient bool
	}{
		{
			name:             "error on get client",
			errorOnGetClient: true,
		},
		{
			name:             "no error on get client",
			errorOnGetClient: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			factory := bbTestUtil.GetFakeFactory()
			factory.ResetIOStream()
			streams, _ := factory.GetIOStream()
			v, _ := factory.GetViper()
			v.Set("big-bang-repo", "/path/to/repo")
			in := streams.In.(*bytes.Buffer)
			out := streams.Out.(*bytes.Buffer)
			errOut := streams.ErrOut.(*bytes.Buffer)
			if tc.errorOnGetClient {
				factory.SetFail.GetOutputClient = true
			}
			// Act
			cmd := NewConfigCmd(factory)
			err := cmd.Execute()
			// Assert
			assert.Empty(t, errOut.String())
			assert.Empty(t, in.String())
			if tc.errorOnGetClient {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), "failed to get output client")
				assert.Empty(t, out.String())
			} else {
				assert.Nil(t, err)
				assert.Contains(t, out.String(), "Please provide a subcommand for config (see help)")
			}
		})
	}
}
