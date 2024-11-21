package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"
)

func TestUpdate_Usage(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	// Act
	cmd := NewUpdateCmd(factory)
	// Assert
	assert.NotNil(t, cmd)
	assert.Equal(t, "update", cmd.Use)
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

			cmd := NewUpdateCmd(factory)
			err := cmd.Execute()

			assert.Empty(t, errOut.String())
			assert.Empty(t, in.String())
			if tc.errorOnGetClient {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "failed to get output client")
				assert.Empty(t, out.String())
			} else {
				require.NoError(t, err)
				// TODO update this with an actual output
				assert.Contains(t, out.String(), "No update functionality has been implemented yet")
			}
		})
	}
}
