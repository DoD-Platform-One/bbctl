package k3d

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	outputSchema "repo1.dso.mil/big-bang/product/packages/bbctl/util/output/schemas"
	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"
	bbTestApiWrappers "repo1.dso.mil/big-bang/product/packages/bbctl/util/test/apiwrappers"
)

func TestK3d_NewCreateClusterCmd(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	// Act
	cmd := NewCreateClusterCmd(factory)
	// Assert
	assert.NotNil(t, cmd)
	assert.Equal(t, "create", cmd.Use)
}

func TestK3d_NewCreateClusterCmd_Run(t *testing.T) {
	bigBangRepoLocation := "/tmp/big-bang"
	require.NoError(t, os.MkdirAll(bigBangRepoLocation, 0755))

	testCases := []struct {
		name           string
		format         string
		expectedOutput string
	}{
		{
			name:   "JSON",
			format: "json",
			expectedOutput: fmt.Sprintf(
				"{\"generalInfo\":null,\"actions\":[\"Running command: %s/docs/assets/scripts/developer/k3d-dev.sh\"],\"warnings\":[]}",
				bigBangRepoLocation,
			),
		},
		{
			name:   "YAML",
			format: "yaml",
			expectedOutput: fmt.Sprintf(
				"generalInfo: {}\nactions:\n- 'Running command: %s/docs/assets/scripts/developer/k3d-dev.sh'\nwarnings: []\n",
				bigBangRepoLocation,
			),
		},
		{
			name:   "TEXT",
			format: "text",
			expectedOutput: fmt.Sprintf(
				"Actions:\n  Running command: %s/docs/assets/scripts/developer/k3d-dev.sh\n\n",
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
			cmd := NewCreateClusterCmd(factory)
			// Act
			err := cmd.Execute()
			// Assert
			assert.NotNil(t, cmd)
			require.NoError(t, err)
			assert.Equal(t, "create", cmd.Use)
			assert.Empty(t, streams.ErrOut.(*bytes.Buffer).String())
			// Check the output
			assert.Equal(
				t,
				tc.expectedOutput,
				streams.Out.(*bytes.Buffer).String(),
			) // Ensure this matches your actual expected output
		})
	}
}

func TestK3d_Failures(t *testing.T) {
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
			expectedError:       "unable to create output client:",
		},
		{
			name:                  "Fail to get command wrapper",
			errorOnCommandWrapper: true,
			expectedError:         "failed to get command wrapper",
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
			expectedError: "failed to run command",
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
			cmd := NewCreateClusterCmd(factory)
			viper, _ := factory.GetViper()
			streams, _ := factory.GetIOStream()
			originalOut := streams.Out
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
				require.NoError(t, factory.SetPipe(r, w))
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
			err := createCluster(factory, cmd, []string{})

			// Assert
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedError)
			assert.Equal(t, "create", cmd.Use)
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
					assert.Contains(t, result, "Running command: /tmp/big-bang/docs/assets/scripts/developer/k3d-dev.sh")
				}
				assert.Empty(t, streams.ErrOut.(*bytes.Buffer).String())
				assert.Empty(t, streams.In.(*bytes.Buffer).String())
			}
		})
	}
}

func TestParseOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected outputSchema.Output
	}{
		{
			name:  "Single Action",
			input: "Action 1",
			expected: outputSchema.Output{
				Actions:  []string{"Action 1"},
				Warnings: []string{},
			},
		},
		{
			name:  "Single Warning",
			input: "Warning: First warning",
			expected: outputSchema.Output{
				Actions:  []string{},
				Warnings: []string{"Warning: First warning"},
			},
		},
		{
			name:  "Multiple Actions",
			input: "Action 1\nAction 2",
			expected: outputSchema.Output{
				Actions:  []string{"Action 1", "Action 2"},
				Warnings: []string{},
			},
		},
		{
			name:  "Multiple Warnings",
			input: "Warning: First warning\nWarning: Second warning",
			expected: outputSchema.Output{
				Actions:  []string{},
				Warnings: []string{"Warning: First warning", "Warning: Second warning"},
			},
		},
		{
			name:  "Actions and Warnings Mixed",
			input: "Action 1\nWarning: First warning\nAction 2\nWarning: Second warning",
			expected: outputSchema.Output{
				Actions:  []string{"Action 1", "Action 2"},
				Warnings: []string{"Warning: First warning", "Warning: Second warning"},
			},
		},
		{
			name:  "Empty Lines",
			input: "Action 1\n\nWarning: First warning\n\nAction 2\nWarning: Second warning\n",
			expected: outputSchema.Output{
				Actions:  []string{"Action 1", "Action 2"},
				Warnings: []string{"Warning: First warning", "Warning: Second warning"},
			},
		},
		{
			name:  "Warning without Action",
			input: "Warning: First warning",
			expected: outputSchema.Output{
				Actions:  []string{},
				Warnings: []string{"Warning: First warning"},
			},
		},
		{
			name:  "Warnings and Actions in Any Order",
			input: "Action 1\nWarning: First warning\nAction 2\nWarning: Second warning\nAction 3",
			expected: outputSchema.Output{
				Actions:  []string{"Action 1", "Action 2", "Action 3"},
				Warnings: []string{"Warning: First warning", "Warning: Second warning"},
			},
		},
		{
			name:  "No Warnings or Actions",
			input: "\n\n",
			expected: outputSchema.Output{
				Actions:  []string{},
				Warnings: []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseOutput(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
