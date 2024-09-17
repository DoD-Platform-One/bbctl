package config

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"

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
			assert.Nil(t, os.WriteFile(dir+"/config.yaml", []byte("big-bang-repo: test"), 0644))
			assert.Nil(t, v.ReadInConfig())
			if tc.errorOnGetClient {
				factory.SetFail.GetOutputClient = true
			}
			if tc.errorOnOutput {
				v.Set("format", "garbage")
			}
			cmd := NewSetCmd(factory)
			if tc.errorOnSetConfigValue {
				factory.SetFail.GetViper = 3
			}
			// Act
			err := cmd.RunE(cmd, []string{"test", "stuff"})
			// Assert
			assert.NotNil(t, cmd)
			assert.Equal(t, "set [key] [value]", cmd.Use)
			assert.Equal(t, "Set a configuration value", cmd.Short)
			assert.Equal(t, "Example usage: bbctl config set KEY VALUE", cmd.Long)
			if tc.errorOnGetClient || tc.errorOnSetConfigValue || tc.errorOnOutput {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), tc.expectedOutput)
				assert.Empty(t, in)
				assert.Empty(t, out)
				assert.Empty(t, errOut)
			} else {
				assert.Nil(t, err)
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
				assert.Nil(t, os.WriteFile(dir+"/config.yaml", []byte("big-bang-repo: test"), 0644))
				assert.Nil(t, v.ReadInConfig())
			}
			// Act
			err := setConfigValue(factory, "test", "stuff")
			// Assert
			if tc.errorOnGetViper || tc.errorOnWriteConfig {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), tc.expectedOutput)
				if tc.errorOnGetViper {
					configContents, _ := os.ReadFile(dir + "/config.yaml")
					obj := map[string]interface{}{}
					assert.Nil(t, yaml.Unmarshal(configContents, &obj))
					assert.Equal(t, "test", obj["big-bang-repo"])
					assert.Len(t, obj, 1)
				} else {
					assert.NoFileExists(t, dir+"/config")
				}
			} else {
				assert.Nil(t, err)
				assert.Equal(t, "stuff", v.GetString("test"))
				configContents, _ := os.ReadFile(dir + "/config.yaml")
				obj := map[string]interface{}{}
				assert.Nil(t, yaml.Unmarshal(configContents, &obj))
				assert.Equal(t, "stuff", obj["test"])
				assert.Equal(t, "test", obj["big-bang-repo"])
				assert.Len(t, obj, 2)
			}
		})
	}
}
