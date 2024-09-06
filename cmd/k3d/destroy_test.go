package k3d

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"
	bbTestApiWrappers "repo1.dso.mil/big-bang/product/packages/bbctl/util/test/apiwrappers"
)

func TestNewDestroyClusterCmd(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	// Act
	cmd := NewDestroyClusterCmd(factory)
	// Assert
	assert.NotNil(t, cmd)
	assert.Equal(t, "destroy", cmd.Use)
}

func TestNewDestroyClusterCmd_RunWithMissingBigBangRepo(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "")
	// Act
	cmd := NewDestroyClusterCmd(factory)
	err := cmd.RunE(cmd, []string{})
	// Assert
	assert.NotNil(t, cmd)
	assert.Error(t, err)
	if !assert.Contains(
		t,
		err.Error(),
		"Error:Field validation for 'BigBangRepo' failed on the 'required' tag",
	) {
		t.Errorf("unexpected output: %s", err.Error())
	}
	assert.Equal(t, "destroy", cmd.Use)
}

func TestNewDestroyClusterCmd_Run(t *testing.T) {
	bigBangRepoLocation := "/tmp/big-bang"
	assert.Nil(t, os.MkdirAll(bigBangRepoLocation, 0755))

	testCases := []struct {
		name           string
		format         string
		expectedOutput string
	}{
		{
			name:   "JSON",
			format: "json",
			expectedOutput: fmt.Sprintf(
				"{\n  \"general_info\": null,\n  \"actions\": [\n    \"Running command: %s/docs/assets/scripts/developer/k3d-dev.sh -d\"\n  ],\n  \"warnings\": []\n}",
				bigBangRepoLocation,
			),
		},
		{
			name:   "YAML",
			format: "yaml",
			expectedOutput: fmt.Sprintf(
				"general_info: {}\nactions:\n- 'Running command: %s/docs/assets/scripts/developer/k3d-dev.sh -d'\nwarnings: []\n",
				bigBangRepoLocation,
			),
		},
		{
			name:   "TEXT",
			format: "text",
			expectedOutput: fmt.Sprintf(
				"Actions:\n  Running command: %s/docs/assets/scripts/developer/k3d-dev.sh -d\n\n",
				bigBangRepoLocation,
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			factory := bbTestUtil.GetFakeFactory()
			factory.ResetIOStream()
			streams, _ := factory.GetIOStream()

			// Set up the environment and configuration
			v, _ := factory.GetViper()
			v.Set("big-bang-repo", bigBangRepoLocation)
			v.Set("output-config.format", tc.format)

			// Expected output from the command
			cmd := NewDestroyClusterCmd(factory)

			// Act
			err := cmd.Execute()

			// Assert
			assert.NotNil(t, cmd)
			assert.Nil(t, err)
			assert.Equal(t, "destroy", cmd.Use)
			assert.Empty(t, streams.ErrOut.(*bytes.Buffer).String())

			// Check the output
			assert.Equal(
				t,
				tc.expectedOutput,
				streams.Out.(*bytes.Buffer).String(),
			)
		})
	}
}

func TestDestroyCluster_Failures(t *testing.T) {
	testCases := []struct {
		name                  string
		errorOnIOStream       bool
		errorOnConfigClient   bool
		errorOnConfig         bool
		errorOnOutputClient   bool
		errorOnCommandWrapper bool
		errorOnPipe           bool
		errorOnCopyBuffer     bool
		errorOnCmdRun         bool
		errorOnOutput         bool
		expectedError         string
	}{
		{
			name:            "Fail to get IO stream",
			errorOnIOStream: true,
			expectedError:   "failed to get streams",
		},
		{
			name:                "Fail to get config client",
			errorOnConfigClient: true,
			expectedError:       "failed to get config client",
		},
		{
			name:          "Fail to get config",
			errorOnConfig: true,
			expectedError: "error getting config:",
		},
		{
			name:                "Fail to get output client",
			errorOnOutputClient: true,
			expectedError:       "Unable to create output client:",
		},
		{
			name:                  "Fail to get command wrapper",
			errorOnCommandWrapper: true,
			expectedError:         "Unable to get command wrapper:",
		},
		{
			name:          "Fail to get pipe",
			errorOnPipe:   true,
			expectedError: "unable to get pipe:",
		},
		{
			name:              "Error on copy buffer alone",
			errorOnCopyBuffer: true,
			expectedError:     "(sole deferred error: FakeFile intentionally errored)",
		},
		{
			name:              "Error on copy buffer and output",
			errorOnCopyBuffer: true,
			errorOnCmdRun:     true,
			expectedError:     "(additional deferred error: FakeFile intentionally errored)",
		},
		{
			name:          "Fail to run command",
			errorOnCmdRun: true,
			expectedError: "Failed to run command",
		},
		{
			name:          "Fail to push output",
			errorOnOutput: true,
			expectedError: "unable to write human-readable output:",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			factory := bbTestUtil.GetFakeFactory()
			cmd := NewDestroyClusterCmd(factory)
			viper, _ := factory.GetViper()
			streams, _ := factory.GetIOStream()
			originalOut := (*streams).Out
			if tc.errorOnIOStream {
				factory.SetFail.GetIOStreams = 1
			}
			if tc.errorOnConfigClient {
				factory.SetFail.GetConfigClient = 1
			}
			if !tc.errorOnConfig {
				viper.Set("big-bang-repo", "/tmp/big-bang")
			}
			if tc.errorOnOutputClient {
				factory.SetFail.GetOutputClient = true
			}
			if tc.errorOnCommandWrapper {
				factory.SetFail.GetCommandWrapper = true
			}
			if tc.errorOnPipe {
				factory.SetFail.GetPipe = true
			}
			if tc.errorOnCopyBuffer {
				r, w, _ := bbTestApiWrappers.CreateFakeFileFromOSPipe(t, false, false)
				r.SetFail.WriteTo = true
				assert.Nil(t, factory.SetPipe(r, w))
			}
			if tc.errorOnCmdRun {
				factory.SetFail.SetCommandWrapperRunError = true
			}
			if tc.errorOnOutput {
				fakeWriter := bbTestApiWrappers.CreateFakeReaderWriter(t, false, true)
				streams.Out = fakeWriter
				factory.SetIOStream(streams)
				originalOut = fakeWriter
			}

			// Act
			err := destroyCluster(factory, cmd, []string{})

			// Assert
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedError)
			assert.Equal(t, "destroy", cmd.Use)
			if tc.errorOnOutput {
				assert.Empty(t, originalOut.(*bbTestApiWrappers.FakeReaderWriter).ActualBuffer.(*bytes.Buffer).String())
			} else {
				var result string
				if ff, ok := originalOut.(*bbTestApiWrappers.FakeReaderWriter); ok {
					buf := &bytes.Buffer{}
					_, _ = io.Copy(buf, ff.ActualBuffer)
					result = buf.String()
				} else {
					result = originalOut.(*bytes.Buffer).String()
				}
				// when running from the test, the output is empty, but when running from the command line, it is not
				if result != "" {
					assert.Contains(t, result, "Running command: /tmp/big-bang/docs/assets/scripts/developer/k3d-dev.sh -d")
				}
				assert.Empty(t, streams.ErrOut.(*bytes.Buffer).String())
				assert.Empty(t, streams.In.(*bytes.Buffer).String())
			}
		})
	}
}
