package update

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"
)

func TestUpdate_RootUsage(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	// Act
	cmd := NewUpdateCmd(factory)
	// Assert
	assert.NotNil(t, cmd)
	assert.Equal(t, "update", cmd.Use)
	commandsList := cmd.Commands()
	assert.Len(t, commandsList, 1)
	var commandUseNamesList = make([]string, len(commandsList))
	for _, command := range commandsList {
		commandUseNamesList = append(commandUseNamesList, command.Use)
	}
	assert.Contains(t, commandUseNamesList, "check")
}

func TestUpdate_RootNoSubcommand(t *testing.T) {
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
			v, _ := factory.GetViper()
			v.Set("big-bang-repo", "/path/to/repo")
			streams, _ := factory.GetIOStream()
			in := streams.In.(*bytes.Buffer)
			out := streams.Out.(*bytes.Buffer)
			errOut := streams.ErrOut.(*bytes.Buffer)
			if tc.errorOnGetClient {
				factory.SetFail.GetOutputClient = true
			}
			// Act
			cmd := NewUpdateCmd(factory)
			err := cmd.Execute()
			// Assert
			assert.Empty(t, errOut.String())
			assert.Empty(t, in.String())
			if tc.errorOnGetClient {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "failed to get output client")
				assert.Empty(t, out.String())
			} else {
				require.NoError(t, err)
				assert.Contains(t, out.String(), "Please provide a subcommand for config (see help)")
			}
		})
	}
}
