package config

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	commonInterfaces "repo1.dso.mil/big-bang/product/packages/bbctl/util/common_interfaces"
	bbConfig "repo1.dso.mil/big-bang/product/packages/bbctl/util/config"
	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"
	apiWrappers "repo1.dso.mil/big-bang/product/packages/bbctl/util/test/apiwrappers"
)

func TestGetConfigInit(t *testing.T) {
	testCases := []struct {
		name           string
		errorOnStream  bool
		expectedOutput string
	}{
		{
			name:           "Test Get Config Init",
			expectedOutput: "test1",
		},
		{
			name:           "Test Get Config Init Error On Stream",
			errorOnStream:  true,
			expectedOutput: "error getting IO streams",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			factory := bbTestUtil.GetFakeFactory()
			if tc.errorOnStream {
				factory.SetFail.GetIOStreams = 1
			}
			// Act
			cmd, _ := NewConfigInitCmd(factory)
			// Assert
			err := cmd.RunE(cmd, []string{})
			if tc.errorOnStream {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), tc.expectedOutput)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestWriteConfigFile(t *testing.T) {
	testCases := []struct {
		name           string
		errorOnMarshal bool
		errorOnHomeDir bool
		errorOnCreate  bool
		errorOnClose   bool
		errorOnWrite   bool
		expectedOutput string
	}{
		{
			name:           "Test Write Config File",
			expectedOutput: "test1",
		},
		{
			name:           "Test Write Config File Error On Marshal",
			errorOnMarshal: true,
			expectedOutput: "test2",
		},
		{
			name:           "Test Write Config File Error On Home Dir",
			errorOnHomeDir: true,
			expectedOutput: "test3",
		},
		{
			name:           "Test Write Config File Error On Create",
			errorOnCreate:  true,
			expectedOutput: "test4",
		},
		{
			name:           "Test Write Config File Error On Close",
			errorOnClose:   true,
			expectedOutput: "sole deferred error",
		},
		{
			name:           "Test Write Config File Error On Write",
			errorOnWrite:   true,
			expectedOutput: "FakeFile intentionally errored",
		},
		{
			name:           "Test Write Config File Error On Write And Close",
			errorOnWrite:   true,
			errorOnClose:   true,
			expectedOutput: "additional deferred error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			config := &map[string]interface{}{}
			err := fmt.Errorf(tc.expectedOutput)
			marshal := func(v interface{}) ([]byte, error) {
				if tc.errorOnMarshal {
					return nil, err
				}
				return []byte(tc.expectedOutput), nil
			}
			homeDir := "test"
			create := func(path string) (commonInterfaces.FileLike, error) {
				if tc.errorOnCreate {
					return nil, err
				}
				_, file, _ := apiWrappers.CreateFakeFileFromOSPipe(t, false, tc.errorOnWrite)
				if tc.errorOnClose {
					file.SetFail.Close = true
				}
				return file, nil
			}
			// Act
			err = writeConfigFile(config, marshal, homeDir, create)
			// Assert
			if tc.errorOnClose || tc.errorOnCreate || tc.errorOnMarshal || tc.errorOnWrite {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), tc.expectedOutput)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestConfigInitErrorBindingFlags(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "test")

	expectedError := fmt.Errorf("failed to set and bind flag")
	logClient, _ := factory.GetLoggingClient()

	tests := []struct {
		flagName       string
		failOnCallNum  int
		expectedCmd    bool
		expectedErrMsg string
	}{
		{
			flagName:       "output",
			failOnCallNum:  1,
			expectedCmd:    false,
			expectedErrMsg: fmt.Sprintf("error setting and binding output flag: %s", expectedError.Error()),
		},
		{
			flagName:       "bbctl-log-level",
			failOnCallNum:  2,
			expectedCmd:    false,
			expectedErrMsg: fmt.Sprintf("error setting and binding bbctl-log-level flag: %s", expectedError.Error()),
		},
		{
			flagName:       "bbctl-log-add-source",
			failOnCallNum:  3,
			expectedCmd:    false,
			expectedErrMsg: fmt.Sprintf("error setting and binding bbctl-log-add-source flag: %s", expectedError.Error()),
		},
		{
			flagName:       "bbctl-log-format",
			failOnCallNum:  4,
			expectedCmd:    false,
			expectedErrMsg: fmt.Sprintf("error setting and binding bbctl-log-format flag: %s", expectedError.Error()),
		},
		{
			flagName:       "big-bang-repo",
			failOnCallNum:  5,
			expectedCmd:    false,
			expectedErrMsg: fmt.Sprintf("error setting and binding big-bang-repo flag: %s", expectedError.Error()),
		},
		{
			flagName:       "bbctl-log-output",
			failOnCallNum:  6,
			expectedCmd:    false,
			expectedErrMsg: fmt.Sprintf("error setting and binding bbctl-log-output flag: %s", expectedError.Error()),
		},
		{
			flagName:       "big-bang-credential-helper",
			failOnCallNum:  7,
			expectedCmd:    false,
			expectedErrMsg: fmt.Sprintf("error setting and binding big-bang-credential-helper flag: %s", expectedError.Error()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.flagName, func(t *testing.T) {
			callCount := 0
			setAndBindFlagFunc := func(client *bbConfig.ConfigClient, name string, shortName string, value any, description string) error {
				callCount++
				if callCount == tt.failOnCallNum {
					return expectedError
				}
				return nil
			}

			configClient, err := bbConfig.NewClient(nil, setAndBindFlagFunc, &logClient, nil, v)
			assert.Nil(t, err)
			factory.SetConfigClient(configClient)

			// Act
			cmd, err := NewConfigInitCmd(factory)

			// Assert
			if tt.expectedCmd {
				assert.NotNil(t, cmd)
			} else {
				assert.Nil(t, cmd)
			}

			if tt.expectedErrMsg != "" {
				assert.NotNil(t, err)
				assert.Equal(t, tt.expectedErrMsg, err.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
