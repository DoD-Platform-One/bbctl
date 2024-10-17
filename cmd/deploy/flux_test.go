package deploy

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	bbConfig "repo1.dso.mil/big-bang/product/packages/bbctl/util/config"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/config/schemas"
	outputSchema "repo1.dso.mil/big-bang/product/packages/bbctl/util/output/schemas"
	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"
	bbTestApiWrappers "repo1.dso.mil/big-bang/product/packages/bbctl/util/test/apiwrappers"
)

func TestFlux_NewDeployFluxCmd(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()

	// Act
	cmd := NewDeployFluxCmd(factory)

	// Assert
	assert.NotNil(t, cmd)
	assert.Equal(t, "flux", cmd.Use)
}

func TestFlux_NewDeployFluxCmd_MissingBigBangRepo(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	streams, _ := factory.GetIOStream()
	in := streams.In.(*bytes.Buffer)
	out := streams.Out.(*bytes.Buffer)
	errOut := streams.ErrOut.(*bytes.Buffer)
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "")
	v.Set("output-config.format", "yaml")

	// Act
	cmd := NewDeployFluxCmd(factory)
	err := cmd.Execute()

	// Assert
	assert.NotNil(t, cmd)
	require.Error(t, err)
	if !assert.Contains(
		t,
		err.Error(),
		"Error:Field validation for 'BigBangRepo' failed on the 'required' tag",
	) {
		t.Errorf("unexpected output: %s", err.Error())
	}
	assert.Equal(t, "flux", cmd.Use)
	assert.Empty(t, in.String())
	assert.Empty(t, out.String())
	assert.Empty(t, errOut.String())
}

func TestParseOutput(t *testing.T) {
	// Arrange
	inputData := `REGISTRY_URL: https://registry.example.com
REGISTRY_USERNAME: user123
Starting deployment...
Warning: Disk space is low
Deployment complete`

	expectedOutput := outputSchema.Output{
		GeneralInfo: map[string]string{
			"REGISTRY_URL":      "https://registry.example.com",
			"REGISTRY_USERNAME": "user123",
		},
		Actions: []string{
			"Starting deployment...",
			"Deployment complete",
		},
		Warnings: []string{
			"Disk space is low",
		},
	}

	// Act
	parsedOutput := parseOutput(inputData)

	// Assert
	assert.Equal(t, expectedOutput, parsedOutput)
}

func TestFlux_NewDeployFluxCmd_Output(t *testing.T) {
	testCases := []struct {
		name           string
		format         string
		expectedOutput string
	}{
		{
			name:           "JSON",
			format:         "json",
			expectedOutput: `{"generalInfo":{},"actions":["Running command: /tmp/big-bang/scripts/install_flux.sh -u  -p"],"warnings":[]}`,
		},
		{
			name:           "YAML",
			format:         "yaml",
			expectedOutput: "generalInfo: {}\nactions:\n  - 'Running command: /tmp/big-bang/scripts/install_flux.sh -u  -p'\nwarnings: []\n",
		},
		{
			name:           "TEXT",
			format:         "text",
			expectedOutput: "General Info:\n\nActions:\n  Running command: /tmp/big-bang/scripts/install_flux.sh -u  -p\n\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			factory := bbTestUtil.GetFakeFactory()
			factory.ResetIOStream()
			streams, _ := factory.GetIOStream()
			// Set up the environment and configuration
			bigBangRepoLocation := "/tmp/big-bang"
			require.NoError(t, os.MkdirAll(bigBangRepoLocation, 0755))
			v, _ := factory.GetViper()
			v.Set("big-bang-repo", bigBangRepoLocation)
			v.Set("output-config.format", tc.format)
			// Expected output from the command
			cmd := NewDeployFluxCmd(factory)
			// Act
			err := cmd.Execute()
			// Assert
			assert.NotNil(t, cmd)
			require.NoError(t, err)
			assert.Equal(t, "flux", cmd.Use)
			assert.Empty(t, streams.ErrOut.(*bytes.Buffer).String())
			// Check the output
			assert.Equal(t, tc.expectedOutput, streams.Out.(*bytes.Buffer).String()) // Ensure this matches your actual expected output
		})
	}
}

func TestDeployFluxConfigClientError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	bigBangRepoLocation := "/tmp/big-bang"
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", bigBangRepoLocation)
	v.Set("output-config.format", "yaml")

	cmd := NewDeployFluxCmd(factory)
	factory.SetFail.GetConfigClient = 1
	// Act
	err := cmd.RunE(cmd, []string{})
	// Assert
	assert.NotNil(t, cmd)
	require.Error(t, err)
	if !assert.Contains(t, err.Error(), "failed to get config client") {
		t.Errorf("unexpected output: %s", err.Error())
	}
}

func TestFluxFailToGetConfig(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	loggingClient, _ := factory.GetLoggingClient()
	cmd := NewDeployFluxCmd(factory)
	viper, _ := factory.GetViper()
	expected := ""
	getConfigFunc := func(_ *bbConfig.ConfigClient) (*schemas.GlobalConfiguration, error) {
		return &schemas.GlobalConfiguration{
			BigBangRepo: expected,
		}, errors.New("dummy error")
	}
	client, _ := bbConfig.NewClient(getConfigFunc, nil, &loggingClient, cmd, viper)
	factory.SetConfigClient(client)

	// Act
	err := deployFluxToCluster(factory, cmd, []string{})

	// Assert
	require.Error(t, err)
	if !assert.Contains(t, err.Error(), "error getting config:") {
		t.Errorf("unexpected output: %s", err.Error())
	}
}

func TestDeployFluxToClusterErrors(t *testing.T) {
	testCases := []struct {
		name                     string
		errorOnConfigClient      bool
		errorOnConfig            bool
		errorOnIOStreams         bool
		errorOnOutputClient      bool
		errorOnCredential        bool
		errorOnUsername          bool
		errorOnPassword          bool
		errorOnGetCommandWrapper bool
		errorOnGetPipe           bool
		errorOnCopyBuffer        bool
		errorOnCmdRun            bool
		errorOnOutput            bool
		expectedError            string
		expectedOutput           string
	}{
		{
			name:                "Fail on Config Client",
			errorOnConfigClient: true,
			expectedError:       "failed to get config client",
		},
		{
			name:          "Fail on Config",
			errorOnConfig: true,
			expectedError: "error getting config",
		},
		{
			name:             "Fail on IO Streams",
			errorOnIOStreams: true,
			expectedError:    "unable to create IO streams",
		},
		{
			name:                "Fail on Output Client",
			errorOnOutputClient: true,
			expectedError:       "unable to create output client",
		},
		{
			name:              "Fail on Credential",
			errorOnCredential: true,
			expectedError:     "unable to get credential helper",
		},
		{
			name:            "Fail on Username",
			errorOnUsername: true,
			expectedError:   "unable to get username",
		},
		{
			name:            "Fail on Password",
			errorOnPassword: true,
			expectedError:   "unable to get password",
		},
		{
			name:                     "Fail on Get Command Wrapper",
			errorOnGetCommandWrapper: true,
			expectedError:            "unable to get command wrapper",
		},
		{
			name:           "Fail on Get Pipe",
			errorOnGetPipe: true,
			expectedError:  "unable to get pipe",
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
			name:          "Fail on Command Run",
			errorOnCmdRun: true,
			expectedError: "failed to run command",
		},
		{
			name:           "Fail on Output",
			errorOnOutput:  true,
			expectedError:  "FakeWriter intentionally errored",
			expectedOutput: "error: must specify one of: flux",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			factory := bbTestUtil.GetFakeFactory()
			factory.ResetIOStream()
			streams, err := factory.GetIOStream()
			// TODO: fix the flux client changing up the streams
			originalOut := streams.Out
			require.NoError(t, err)
			v, _ := factory.GetViper()
			v.Set("big-bang-repo", "/tmp/big-bang")
			v.Set("format", "yaml")
			if tc.errorOnConfigClient {
				factory.SetFail.GetConfigClient = 1
			}
			if tc.errorOnConfig {
				v.Set("big-bang-repo", "")
			}
			if tc.errorOnIOStreams {
				factory.SetFail.GetIOStreams = 1
			}
			if tc.errorOnOutputClient {
				factory.SetFail.GetOutputClient = true
			}
			if tc.errorOnCredential {
				factory.SetFail.GetCredentialHelper = true
			}
			if tc.errorOnUsername {
				factory.SetCredentialHelper(func(s1, _ string) (string, error) {
					if s1 == "username" {
						return "", errors.New("dummy error")
					}
					return "dummy", nil
				})
			}
			if tc.errorOnPassword {
				factory.SetCredentialHelper(func(s1, _ string) (string, error) {
					if s1 == "password" {
						return "", errors.New("dummy error")
					}
					return "dummy", nil
				})
			}
			if tc.errorOnGetCommandWrapper {
				factory.SetFail.GetCommandWrapper = true
			}
			if tc.errorOnGetPipe {
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
			cmd := NewDeployFluxCmd(factory)
			// Act
			err = deployFluxToCluster(factory, cmd, []string{})
			// Assert
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedError)
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
				assert.Contains(t, result, tc.expectedOutput)
			}
		})
	}
}

func TestFluxOutputParsing(t *testing.T) {
	// Arrange
	output := "\n\n Action 1 \n Warning: Warning 1\n invalid: warning: Action: 2\n key: value\n REGISTRY_URL: localhost\n REGISTRY_USERNAME: username"
	// Act
	schema := parseOutput(output)
	// Assert
	assert.Equal(t, map[string]string{"REGISTRY_URL": "localhost", "REGISTRY_USERNAME": "username"}, schema.GeneralInfo)
	assert.Equal(t, []string{"Action 1", "invalid: warning: Action: 2", "key: value"}, schema.Actions)
	assert.Equal(t, []string{"Warning 1"}, schema.Warnings)
}
