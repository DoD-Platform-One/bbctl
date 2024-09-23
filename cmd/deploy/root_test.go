package deploy

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"
	bbTestApiWrappers "repo1.dso.mil/big-bang/product/packages/bbctl/util/test/apiwrappers"
)

func TestRoot_NewDeployCmd(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	// Act
	cmd, _ := NewDeployCmd(factory)
	// Assert
	assert.NotNil(t, cmd)
	assert.Equal(t, "deploy", cmd.Use)
	commandsList := cmd.Commands()
	assert.Len(t, commandsList, 2)
	var commandUseNamesList []string
	for _, command := range commandsList {
		commandUseNamesList = append(commandUseNamesList, command.Use)
	}
	assert.Contains(t, commandUseNamesList, "flux")
	assert.Contains(t, commandUseNamesList, "bigbang")
}

func TestRoot_NewDeployCmd_NoSubcommand(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	factory.SetLoggingFunc(nil)
	streams, _ := factory.GetIOStream()
	in := streams.In.(*bytes.Buffer)
	out := streams.Out.(*bytes.Buffer)
	errOut := streams.ErrOut.(*bytes.Buffer)
	// Act
	cmd, _ := NewDeployCmd(factory)
	assert.Nil(t, cmd.RunE(cmd, []string{}))
	// Assert
	assert.NotNil(t, cmd)
	assert.Equal(t, "deploy", cmd.Use)
	assert.Empty(t, in.String())
	assert.Empty(t, out.String())
	assert.Contains(t, errOut.String(), "error: must specify one of:")
}

func TestRoot_NewDeployBigBang_CommandError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.SetFail.GetConfigClient = 1
	// Act
	cmd, err := NewDeployCmd(factory)
	// Assert
	assert.Nil(t, cmd)
	assert.Error(t, err)
	if !assert.Contains(t, err.Error(), "Error retrieving BigBang Command:") {
		t.Errorf("unexpected output: %s", err.Error())
	}
}

func TestRoot_NewDeployCmd_NoSubcommand_Error(t *testing.T) {
	testCases := []struct {
		name                 string
		errorOnLoggingClient bool
		errorOnWrite         bool
		errorOnHelp          bool
		expectedError        string
	}{
		{
			name:                 "LoggingClient",
			errorOnLoggingClient: true,
			expectedError:        "Unable to get logging client",
		},
		{
			name:          "Write",
			errorOnWrite:  true,
			expectedError: "Unable to write to output stream",
		},
		{
			name:          "Help",
			errorOnHelp:   true,
			expectedError: "Unable to write to output stream",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			factory := bbTestUtil.GetFakeFactory()
			cmd, _ := NewDeployCmd(factory)
			if tc.errorOnLoggingClient {
				factory.SetFail.GetLoggingClient = true
			}
			if tc.errorOnWrite {
				fakeWriter := bbTestApiWrappers.CreateFakeReaderWriter(t, false, true)
				streams, _ := factory.GetIOStream()
				streams.ErrOut = fakeWriter
				factory.SetIOStream(streams)
			}
			if tc.errorOnHelp {
				cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
					panic("Unable to write to output stream")
				})
			}
			// Act
			var err error
			if tc.errorOnWrite || tc.errorOnHelp {
				assert.Panics(t, func() {
					err = cmd.RunE(cmd, []string{})
				})
			} else {
				err = cmd.RunE(cmd, []string{})
			}
			// Assert
			assert.NotNil(t, cmd)
			assert.Equal(t, "deploy", cmd.Use)
			if tc.errorOnWrite || tc.errorOnHelp {
				assert.Nil(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			}
		})
	}
}
