package config

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	commonInterfaces "repo1.dso.mil/big-bang/product/packages/bbctl/util/common_interfaces"
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
			cmd := NewConfigInitCmd(factory)
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
			homeDir := func() (string, error) {
				if tc.errorOnHomeDir {
					return "", err
				}
				return "test", nil
			}
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
			if tc.errorOnClose || tc.errorOnCreate || tc.errorOnHomeDir || tc.errorOnMarshal || tc.errorOnWrite {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), tc.expectedOutput)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
