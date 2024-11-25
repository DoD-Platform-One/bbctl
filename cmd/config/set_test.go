package config

import (
	"bytes"
	"errors"
	"os"
	"testing"

	"repo1.dso.mil/big-bang/apps/developer-tools/go-utils/yamler"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	bbConfig "repo1.dso.mil/big-bang/product/packages/bbctl/util/config"
	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"
)

func TestSet_NewSetCmd(t *testing.T) {
	testCases := []struct {
		name                  string
		errorOnGetClient      bool
		errorOnSetConfigValue bool
		errorOnOutput         bool
		expectedOutput        string
	}{
		{
			name:           "pass",
			expectedOutput: "Configuration updated",
		},
		{
			name:             "error on get client",
			errorOnGetClient: true,
			expectedOutput:   "failed to get output client:",
		},
		{
			name:                  "error on set config value",
			errorOnSetConfigValue: true,
			expectedOutput:        "failed to set config value:",
		},
		{
			name:           "error on output",
			errorOnOutput:  true,
			expectedOutput: "unsupported format",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			dir := t.TempDir()
			factory := bbTestUtil.GetFakeFactory()
			factory.ResetIOStream()
			streams, _ := factory.GetIOStream()
			in, out, errOut := streams.In.(*bytes.Buffer), streams.Out.(*bytes.Buffer), streams.ErrOut.(*bytes.Buffer)
			v, _ := factory.GetViper()
			v.AddConfigPath(dir)
			v.SetConfigName("config")
			assert.NoError(t, os.WriteFile(dir+"/config.yaml", []byte("big-bang-repo: test"), 0600))
			assert.NoError(t, v.ReadInConfig())
			if tc.errorOnGetClient {
				factory.SetFail.GetOutputClient = true
			}
			if tc.errorOnOutput {
				v.Set("format", "garbage")
			}
			cmd, _ := NewSetCmd(factory)
			if tc.errorOnSetConfigValue {
				factory.SetFail.GetViper = 4
			}
			// Act
			err := cmd.RunE(cmd, []string{"test", "stuff"})
			// Assert
			assert.NotNil(t, cmd)
			assert.Equal(t, "set [key] [value]", cmd.Use)
			assert.Equal(t, "Set a configuration value", cmd.Short)
			assert.Equal(t, "Example usage: bbctl config set KEY VALUE", cmd.Long)
			if tc.errorOnGetClient || tc.errorOnSetConfigValue || tc.errorOnOutput {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedOutput)
				assert.Empty(t, in)
				assert.Empty(t, out)
				assert.Empty(t, errOut)
			} else {
				require.NoError(t, err)
				assert.Empty(t, in)
				assert.Empty(t, errOut)
				assert.Contains(t, out.String(), tc.expectedOutput)
			}
		})
	}
}

func TestSet_SetConfigValue(t *testing.T) {
	testCases := []struct {
		name               string
		errorOnGetViper    bool
		errorOnWriteConfig bool
		expectedOutput     string
	}{
		{
			name:           "pass",
			expectedOutput: "Configuration updated",
		},
		{
			name:            "error on get viper",
			errorOnGetViper: true,
			expectedOutput:  "failed to get viper:",
		},
		{
			name:               "error on write config",
			errorOnWriteConfig: true,
			expectedOutput:     "Config File \"config\" Not Found in",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			dir := t.TempDir()
			factory := bbTestUtil.GetFakeFactory()
			v, _ := factory.GetViper()
			if tc.errorOnGetViper {
				factory.SetFail.GetViper = 1
			}
			if !tc.errorOnWriteConfig {
				v.AddConfigPath(dir)
				v.SetConfigName("config")
				assert.NoError(t, os.WriteFile(dir+"/config.yaml", []byte("big-bang-repo: test"), 0600))
				assert.NoError(t, v.ReadInConfig())
			}
			// Act
			err := setConfigValue(factory, "test", "stuff", "")
			// Assert
			if tc.errorOnGetViper || tc.errorOnWriteConfig {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedOutput)
				if tc.errorOnGetViper {
					configContents, _ := os.ReadFile(dir + "/config.yaml")
					obj := map[string]interface{}{}
					require.NoError(t, yamler.Unmarshal(configContents, &obj))
					assert.Equal(t, "test", obj["big-bang-repo"])
					assert.Len(t, obj, 1)
				} else {
					assert.NoFileExists(t, dir+"/config")
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, "stuff", v.GetString("test"))
				configContents, _ := os.ReadFile(dir + "/config.yaml")
				obj := map[string]interface{}{}
				require.NoError(t, yamler.Unmarshal(configContents, &obj))
				assert.Equal(t, "stuff", obj["test"])
				assert.Equal(t, "test", obj["big-bang-repo"])
				assert.Len(t, obj, 2)
			}
		})
	}
}

func TestConfigSetErrorBindingFlags(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "test")

	expectedError := errors.New("failed to set and bind flag")
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
			expectedErrMsg: "error setting and binding output flag: " + expectedError.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.flagName, func(t *testing.T) {
			callCount := 0
			setAndBindFlagFunc := func(_ *bbConfig.ConfigClient, _ string, _ string, _ any, _ string) error {
				callCount++
				if callCount == tt.failOnCallNum {
					return expectedError
				}
				return nil
			}

			configClient, err := bbConfig.NewClient(nil, setAndBindFlagFunc, &logClient, nil, v)
			require.NoError(t, err)
			factory.SetConfigClient(configClient)

			// Act
			cmd, err := NewSetCmd(factory)

			// Assert
			if tt.expectedCmd {
				assert.NotNil(t, cmd)
			} else {
				assert.Nil(t, cmd)
			}

			if tt.expectedErrMsg != "" {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErrMsg, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
