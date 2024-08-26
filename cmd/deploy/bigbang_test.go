package deploy

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
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
	assert.Error(t, err)
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
	expectedOutput := `Release "bigbang" has been upgraded. Happy Helming!
NAME: bigbang
LAST DEPLOYED: Thu Aug 15 17:28:15 2024
NAMESPACE: bigbang
STATUS: deployed
REVISION: 3
TEST SUITE: None
NOTES: Thank you for supporting PlatformOne!
`

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

func TestBigBang_NewDeployBigBangCmd_WithK3d(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	factory.ResetPipe()

	// Create the pipe using the factory
	// Get the pipe reader and writer
	r, w, err := factory.GetPipe()
	assert.Nil(t, err)

	streams, _ := factory.GetIOStream()
	streams.In = r
	streams.Out = w

	out := new(bytes.Buffer)
	errOut := streams.ErrOut.(*bytes.Buffer)

	bigBangRepoLocation := "/tmp/big-bang"
	assert.Nil(t, os.MkdirAll(bigBangRepoLocation, 0755))
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", bigBangRepoLocation)
	v.Set("output-config.format", "yaml")
	expectedOutput := fmt.Sprintf(
		"message: 'Running command: helm upgrade -i bigbang %[1]v/chart -n bigbang\n  --create-namespace --set registryCredentials.username= --set registryCredentials.password=\n  -f %[1]v/chart/ingress-certs.yaml -f %[1]v/docs/assets/configs/example/policy-overrides-k3d.yaml '\nname: \"\"\nlastdeployed: \"\"\nnamespace: \"\"\nstatus: \"\"\nrevision: \"\"\ntestsuite: \"\"\nnotes: \"\"\n",
		bigBangRepoLocation,
	)

	cmd, err := NewDeployBigBangCmd(factory)
	assert.Nil(t, err)

	cmd.SetArgs([]string{"--k3d"})

	// Use a WaitGroup to synchronize the goroutine
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		err = cmd.Execute()
		assert.Nil(t, err)

		// Close the writer to signal the end of input
		w.Close()
	}()

	// Read the output from the pipe in the main goroutine
	_, err = io.Copy(out, r)
	assert.Nil(t, err)

	// Wait for the goroutine to finish
	wg.Wait()

	// Assert
	assert.NotNil(t, cmd)
	assert.Equal(t, "bigbang", cmd.Use)
	assert.Empty(t, errOut.String())

	// Check the output
	output := out.String()

	assert.Equal(t, expectedOutput, output)
}

func TestBigBang_NewDeployBigBangCmd_WithComponents(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	factory.ResetPipe()

	// Create the pipe using the factory
	// Get the pipe reader and writer
	r, w, err := factory.GetPipe()
	assert.Nil(t, err)

	streams, _ := factory.GetIOStream()
	streams.In = r
	streams.Out = w

	out := new(bytes.Buffer)
	errOut := streams.ErrOut.(*bytes.Buffer)

	bigBangRepoLocation := "/tmp/big-bang"
	assert.Nil(t, os.MkdirAll(bigBangRepoLocation, 0755))
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", bigBangRepoLocation)
	v.Set("output-config.format", "yaml")
	expectedOutput := fmt.Sprintf(
		"message: 'Running command: helm upgrade -i bigbang %[1]v/chart -n bigbang\n  --create-namespace --set registryCredentials.username= --set registryCredentials.password=\n  --set addons.baz.enabled=true --set addons.bar.enabled=true --set addons.foo.enabled=true '\nname: \"\"\nlastdeployed: \"\"\nnamespace: \"\"\nstatus: \"\"\nrevision: \"\"\ntestsuite: \"\"\nnotes: \"\"\n",
		bigBangRepoLocation,
	)

	cmd, err := NewDeployBigBangCmd(factory)
	assert.Nil(t, err)

	cmd.SetArgs([]string{"--addon=foo,bar", "--addon=baz"})

	// Use a WaitGroup to synchronize the goroutine
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		err = cmd.Execute()
		assert.Nil(t, err)

		// Close the writer to signal the end of input
		w.Close()
	}()

	// Read the output from the pipe in the main goroutine
	_, err = io.Copy(out, r)
	assert.Nil(t, err)

	// Wait for the goroutine to finish
	wg.Wait()

	// Assert
	assert.NotNil(t, cmd)
	assert.Equal(t, "bigbang", cmd.Use)
	assert.Empty(t, errOut.String())

	// Check the output
	output := out.String()

	assert.Equal(t, expectedOutput, output)
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
	assert.Error(t, err)
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
	assert.Error(t, err)
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
			expectedError:   "Dummy Error",
		},
		{
			name:            "Error on password",
			errorOnPassword: true,
			expectedError:   "Dummy Error",
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
			name:          "Error on command run",
			errorOnCmdRun: true,
			expectedError: "Failed to run command",
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
			originalOut := (*streams).Out
			assert.Nil(t, err)
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
				factory.SetCredentialHelper(func(s1, s2 string) (string, error) {
					if s1 == "username" {
						return "", fmt.Errorf("Dummy Error")
					}
					return "dummy", nil
				})
			}
			if tc.errorOnPassword {
				factory.SetCredentialHelper(func(s1, s2 string) (string, error) {
					if s1 == "password" {
						return "", fmt.Errorf("Dummy Error")
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
			if tc.errorOnCmdRun {
				factory.SetFail.SetCommandWrapperRunError = true
			}
			if tc.errorOnOutput {
				fakeWriter := bbTestApiWrappers.CreateFakeWriter(t, true)
				streams.Out = fakeWriter
				factory.SetIOStream(streams)
				originalOut = fakeWriter
			}
			cmd, _ := NewDeployBigBangCmd(factory)
			// Act
			err = deployBigBangToCluster(cmd, factory, []string{})
			// Assert
			assert.Error(t, err)
			if !assert.Contains(t, err.Error(), tc.expectedError) {
				t.Errorf("unexpected output: %s", err.Error())
			}
			if tc.errorOnOutput {
				assert.Empty(t, originalOut.(*bbTestApiWrappers.FakeWriter).ActualBuffer.(*bytes.Buffer).String())
			} else {
				result := originalOut.(*bytes.Buffer).String()
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
