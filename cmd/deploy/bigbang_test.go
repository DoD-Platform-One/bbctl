package deploy

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	outputSchema "repo1.dso.mil/big-bang/product/packages/bbctl/util/output/schemas"
	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"
	bbTestApiWrappers "repo1.dso.mil/big-bang/product/packages/bbctl/util/test/apiwrappers"
)

func TestBigBang_NewDeployBigBangCmd(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	// Act
	cmd, _ := NewDeployBigBangCmd(factory)
	// Assert
	assert.NotNil(t, cmd)
	assert.Equal(t, "bigbang", cmd.Use)
}

func TestBigBang_NewDeployBigBangCmd_MissingBigBangRepo(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "")
	v.Set("output-config.format", "yaml")

	// Act
	cmd, _ := NewDeployBigBangCmd(factory)
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
	assert.Equal(t, "bigbang", cmd.Use)
}

func TestDeployBigBangToCluster_encodeHelmOpts(t *testing.T) {
	expectedOutput := "" +
		"Release \"bigbang\" has been upgraded. Happy Helming!\n" +
		"NAME: bigbang\n" +
		"LAST DEPLOYED: Thu Aug 15 17:28:15 2024\n" +
		"NAMESPACE: bigbang\n" +
		"STATUS: deployed\n" +
		"REVISION: 3\n" +
		"TEST SUITE: None\n" +
		"NOTES: Thank you for supporting PlatformOne!\n"

	parsedData := encodeHelmOpts(expectedOutput)

	expectedData := outputSchema.HelmOutput{
		Message:      `Release "bigbang" has been upgraded. Happy Helming!`,
		Name:         "bigbang",
		Namespace:    "bigbang",
		LastDeployed: "Thu Aug 15 17:28:15 2024",
		Status:       "deployed",
		Revision:     "3",
		TestSuite:    "None",
		Notes:        "Thank you for supporting PlatformOne!",
	}

	assert.Equal(t, expectedData, parsedData)
}

func TestBigBang_NewDeployBigBangCmd_Output(t *testing.T) {
	testCases := []struct {
		name           string
		format         string
		args           []string
		expectedOutput string
	}{
		{
			name:           "With Components YAML",
			format:         "yaml",
			args:           []string{"--addon=foo,bar", "--addon=baz"},
			expectedOutput: "message: 'Running command: helm upgrade -i bigbang /tmp/big-bang/chart -n bigbang\n  --create-namespace --set registryCredentials.username= --set registryCredentials.password=\n  --set addons.baz.enabled=true --set addons.bar.enabled=true --set addons.foo.enabled=true '\nname: \"\"\nlastDeployed: \"\"\nnamespace: \"\"\nstatus: \"\"\nrevision: \"\"\ntestSuite: \"\"\nnotes: \"\"\n",
		},
		{
			name:           "With Components JSON",
			format:         "json",
			args:           []string{"--addon=foo,bar", "--addon=baz"},
			expectedOutput: "{\"message\":\"Running command: helm upgrade -i bigbang /tmp/big-bang/chart -n bigbang --create-namespace --set registryCredentials.username= --set registryCredentials.password= --set addons.baz.enabled=true --set addons.bar.enabled=true --set addons.foo.enabled=true \",\"name\":\"\",\"lastDeployed\":\"\",\"namespace\":\"\",\"status\":\"\",\"revision\":\"\",\"testSuite\":\"\",\"notes\":\"\"}",
		},
		{
			name:           "With Components TEXT",
			format:         "text",
			args:           []string{"--addon=foo,bar", "--addon=baz"},
			expectedOutput: "Message: Running command: helm upgrade -i bigbang /tmp/big-bang/chart -n bigbang --create-namespace --set registryCredentials.username= --set registryCredentials.password= --set addons.baz.enabled=true --set addons.bar.enabled=true --set addons.foo.enabled=true \nName: \nLast Deployed: \nNamespace: \nStatus: \nRevision: \nTest Suite: \nNotes:\n\n\n",
		},
		{
			name:           "With K3d YAML",
			format:         "yaml",
			args:           []string{"--k3d"},
			expectedOutput: "message: 'Running command: helm upgrade -i bigbang /tmp/big-bang/chart -n bigbang\n  --create-namespace --set registryCredentials.username= --set registryCredentials.password=\n  --set addons.baz.enabled=true --set addons.bar.enabled=true --set addons.foo.enabled=true '\nname: \"\"\nlastDeployed: \"\"\nnamespace: \"\"\nstatus: \"\"\nrevision: \"\"\ntestSuite: \"\"\nnotes: \"\"\n",
		},
		{
			name:           "With K3d JSON",
			format:         "json",
			args:           []string{"--k3d"},
			expectedOutput: "{\"message\":\"Running command: helm upgrade -i bigbang /tmp/big-bang/chart -n bigbang --create-namespace --set registryCredentials.username= --set registryCredentials.password= --set addons.baz.enabled=true --set addons.bar.enabled=true --set addons.foo.enabled=true \",\"name\":\"\",\"lastDeployed\":\"\",\"namespace\":\"\",\"status\":\"\",\"revision\":\"\",\"testSuite\":\"\",\"notes\":\"\"}",
		},
		{
			name:           "With K3d TEXT",
			format:         "text",
			args:           []string{"--k3d"},
			expectedOutput: "Message: Running command: helm upgrade -i bigbang /tmp/big-bang/chart -n bigbang --create-namespace --set registryCredentials.username= --set registryCredentials.password= --set addons.baz.enabled=true --set addons.bar.enabled=true --set addons.foo.enabled=true \nName: \nLast Deployed: \nNamespace: \nStatus: \nRevision: \nTest Suite: \nNotes:\n\n\n",
		},
		{
			name:           "With K3d and Components YAML",
			format:         "yaml",
			args:           []string{"--k3d", "--addon=foo,bar", "--addon=baz"},
			expectedOutput: "message: 'Running command: helm upgrade -i bigbang /tmp/big-bang/chart -n bigbang\n  --create-namespace --set registryCredentials.username= --set registryCredentials.password=\n  --set addons.baz.enabled=true --set addons.bar.enabled=true --set addons.foo.enabled=true '\nname: \"\"\nlastDeployed: \"\"\nnamespace: \"\"\nstatus: \"\"\nrevision: \"\"\ntestSuite: \"\"\nnotes: \"\"\n",
		},
		{
			name:           "With K3d and Components JSON",
			format:         "json",
			args:           []string{"--k3d", "--addon=foo,bar", "--addon=baz"},
			expectedOutput: "{\"message\":\"Running command: helm upgrade -i bigbang /tmp/big-bang/chart -n bigbang --create-namespace --set registryCredentials.username= --set registryCredentials.password= --set addons.baz.enabled=true --set addons.bar.enabled=true --set addons.foo.enabled=true \",\"name\":\"\",\"lastDeployed\":\"\",\"namespace\":\"\",\"status\":\"\",\"revision\":\"\",\"testSuite\":\"\",\"notes\":\"\"}",
		},
		{
			name:           "With K3d and Components TEXT",
			format:         "text",
			args:           []string{"--k3d", "--addon=foo,bar", "--addon=baz"},
			expectedOutput: "Message: Running command: helm upgrade -i bigbang /tmp/big-bang/chart -n bigbang --create-namespace --set registryCredentials.username= --set registryCredentials.password= --set addons.baz.enabled=true --set addons.bar.enabled=true --set addons.foo.enabled=true \nName: \nLast Deployed: \nNamespace: \nStatus: \nRevision: \nTest Suite: \nNotes:\n\n\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			factory := bbTestUtil.GetFakeFactory()
			factory.ResetIOStream()
			streams, _ := factory.GetIOStream()
			bigBangRepoLocation := "/tmp/big-bang"
			require.NoError(t, os.MkdirAll(bigBangRepoLocation, 0755))
			v, _ := factory.GetViper()
			v.Set("big-bang-repo", bigBangRepoLocation)
			v.Set("output-config.format", tc.format)
			cmd, err := NewDeployBigBangCmd(factory)
			require.NoError(t, err)
			cmd.SetArgs([]string{"--addon=foo,bar", "--addon=baz"})
			// Act
			err = cmd.Execute()
			// Assert
			require.NoError(t, err)
			assert.NotNil(t, cmd)
			assert.Equal(t, "bigbang", cmd.Use)
			assert.Empty(t, streams.ErrOut.(*bytes.Buffer).String())
			// Check the output
			assert.Equal(t, tc.expectedOutput, streams.Out.(*bytes.Buffer).String())
		})
	}
}

func TestGetBigBangCmdConfigClientError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	bigBangRepoLocation := "/tmp/big-bang"
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", bigBangRepoLocation)
	v.Set("output-config.format", "yaml")

	factory.SetFail.GetConfigClient = 1
	// Act
	cmd, err := NewDeployBigBangCmd(factory)
	// Assert
	assert.Nil(t, cmd)
	require.Error(t, err)
	if !assert.Contains(t, err.Error(), "unable to get config client:") {
		t.Errorf("unexpected output: %s", err.Error())
	}
}

func TestDeployBigBangConfigClientError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	bigBangRepoLocation := "/tmp/big-bang"
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", bigBangRepoLocation)
	cmd, _ := NewDeployBigBangCmd(factory)
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

func TestDeployBigBangToClusterErrors(t *testing.T) {
	testCases := []struct {
		name                    string
		errorOnLoggingClient    bool
		errorOnConfigClient     bool
		errorOnConfig           bool
		errorOnIOStream         bool
		errorOnOutputClient     bool
		errorOnCredentialHelper bool
		errorOnUsername         bool
		errorOnPassword         bool
		errorOnCommandWrapper   bool
		errorOnGetPipe          bool
		errorOnCopyBuffer       bool
		errorOnCmdRun           bool
		errorOnOutput           bool
		expectedError           string
		expectedOutput          string
	}{
		{
			name:                 "Error on logging client",
			errorOnLoggingClient: true,
			expectedError:        "failed to get logging client",
		},
		{
			name:                "Error on config client",
			errorOnConfigClient: true,
			expectedError:       "failed to get config client",
		},
		{
			name:          "Error on config",
			errorOnConfig: true,
			expectedError: "error getting config",
		},
		{
			name:            "Error on IO stream",
			errorOnIOStream: true,
			expectedError:   "unable to create IO streams",
		},
		{
			name:                "Error on output client",
			errorOnOutputClient: true,
			expectedError:       "unable to create output client",
		},
		{
			name:                    "Error on credential helper",
			errorOnCredentialHelper: true,
			expectedError:           "unable to get credential helper",
		},
		{
			name:            "Error on username",
			errorOnUsername: true,
			expectedError:   "dummy error",
		},
		{
			name:            "Error on password",
			errorOnPassword: true,
			expectedError:   "dummy error",
		},
		{
			name:                  "Error on command wrapper",
			errorOnCommandWrapper: true,
			expectedError:         "failed to get command wrapper",
		},
		{
			name:           "Error on get pipe",
			errorOnGetPipe: true,
			expectedError:  "failed to get pipe",
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
			name:          "Error on command run",
			errorOnCmdRun: true,
			expectedError: "failed to run command",
		},
		{
			name:           "Error on output",
			errorOnOutput:  true,
			expectedError:  "unable to write YAML output: FakeWriter intentionally errored",
			expectedOutput: "Running command: helm upgrade -i bigbang /tmp/big-bang/chart -n bigbang\n  --create-namespace --set registryCredentials.username= --set registryCredentials.password=\n  -f /tmp/big-bang/chart/ingress-certs.yaml -f /tmp/big-bang/docs/assets/configs/example/policy-overrides.yaml ",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			factory := bbTestUtil.GetFakeFactory()
			streams, err := factory.GetIOStream()
			// TODO: fix the flux client changing up the streams
			originalOut := streams.Out
			require.NoError(t, err)
			v, _ := factory.GetViper()
			v.Set("big-bang-repo", "/tmp/big-bang")
			v.Set("format", "yaml")
			if tc.errorOnLoggingClient {
				factory.SetFail.GetLoggingClient = true
			}
			if tc.errorOnConfigClient {
				factory.SetFail.GetConfigClient = 1
			}
			if tc.errorOnConfig {
				v.Set("big-bang-repo", "")
			}
			if tc.errorOnIOStream {
				factory.SetFail.GetIOStreams = 1
			}
			if tc.errorOnOutputClient {
				factory.SetFail.GetOutputClient = true
			}
			if tc.errorOnCredentialHelper {
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
			if tc.errorOnCommandWrapper {
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
			cmd, _ := NewDeployBigBangCmd(factory)
			// Act
			err = deployBigBangToCluster(cmd, factory, []string{})
			// Assert
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedError)
			if tc.errorOnOutput {
				assert.Empty(t, originalOut.(*bbTestApiWrappers.FakeReaderWriter).ActualBuffer.(*bytes.Buffer).String())
			} else {
				var result string
				if ff, ok := originalOut.(*bbTestApiWrappers.FakeFile); ok {
					buf := &bytes.Buffer{}
					_, _ = io.Copy(buf, ff.File)
					result = buf.String()
				} else {
					result = originalOut.(*bytes.Buffer).String()
				}
				assert.Contains(t, result, tc.expectedOutput)
			}
		})
	}
}

func TestBigBangEncodeNotes(t *testing.T) {
	// Arrange
	output := "first line is the message\nNOTES: This is a multi-line note:\n that contains multiple : symbols which should cause it to fail parsing earlier\n in the execution of: encodeHelmOpts()"
	// Act
	schema := encodeHelmOpts(output)
	//Assert
	assert.Equal(t, "first line is the message", schema.Message)
	assert.Equal(t, "This is a multi-line note:\n that contains multiple : symbols which should cause it to fail parsing earlier\n in the execution of: encodeHelmOpts()", schema.Notes)
}
