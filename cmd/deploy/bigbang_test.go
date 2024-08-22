package deploy

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	bbConfig "repo1.dso.mil/big-bang/product/packages/bbctl/util/config"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/config/schemas"
	outputSchema "repo1.dso.mil/big-bang/product/packages/bbctl/util/output/schemas"
	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"
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
	err := factory.CreatePipe()
	assert.Nil(t, err)

	// Get the pipe reader and writer
	r, w := factory.GetPipe()

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
	io.Copy(out, r)

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
	err := factory.CreatePipe()
	assert.Nil(t, err)

	// Get the pipe reader and writer
	r, w := factory.GetPipe()

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
	io.Copy(out, r)

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

	factory.SetFail.GetConfigClient = true
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
	factory.SetFail.GetConfigClient = true
	// Act
	err := cmd.RunE(cmd, []string{})

	// Assert
	assert.NotNil(t, cmd)
	assert.Error(t, err)
	if !assert.Contains(t, err.Error(), "failed to get config client") {
		t.Errorf("unexpected output: %s", err.Error())
	}
}

func TestBigBangFailToGetConfig(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	loggingClient, _ := factory.GetLoggingClient()
	cmd, _ := NewDeployBigBangCmd(factory)
	viper, _ := factory.GetViper()
	expected := ""
	getConfigFunc := func(client *bbConfig.ConfigClient) (*schemas.GlobalConfiguration, error) {
		return &schemas.GlobalConfiguration{
			BigBangRepo: expected,
		}, fmt.Errorf("Dummy Error")
	}
	client, _ := bbConfig.NewClient(getConfigFunc, nil, &loggingClient, cmd, viper)
	factory.SetConfigClient(client)

	// Act
	err := deployBigBangToCluster(cmd, factory, []string{})

	// Assert
	assert.Error(t, err)
	if !assert.Contains(t, err.Error(), "error getting config:") {
		t.Errorf("unexpected output: %s", err.Error())
	}
}

func TestBigBangFailToGetStreams(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "/tmp/big-bang")
	v.Set("output-config.format", "yaml")
	factory.SetFail.GetIOStreams = true
	cmd, cmdErr := NewDeployBigBangCmd(factory)

	// Act
	err := deployBigBangToCluster(cmd, factory, []string{})

	// Assert
	assert.NoError(t, cmdErr)
	assert.Error(t, err)
	if !assert.Contains(t, err.Error(), "unable to create IO streams:") {
		t.Errorf("unexpected output: %s", err.Error())
	}
}

func TestBigBangFailToGetOutputClient(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "/tmp/big-bang")
	v.Set("output-config.format", "yaml")
	factory.SetFail.GetOutputClient = true
	cmd, cmdErr := NewDeployBigBangCmd(factory)

	// Act
	err := deployBigBangToCluster(cmd, factory, []string{})

	// Assert
	assert.NoError(t, cmdErr)
	assert.Error(t, err)
	if !assert.Contains(t, err.Error(), "unable to create output client:") {
		t.Errorf("unexpected output: %s", err.Error())
	}
}

func TestBigBangFailToGetCredentialHelper(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "/tmp/big-bang")
	v.Set("output-config.format", "yaml")
	factory.SetFail.GetCredentialHelper = true
	cmd, cmdErr := NewDeployBigBangCmd(factory)

	// Act
	err := deployBigBangToCluster(cmd, factory, []string{})

	// Assert
	assert.NoError(t, cmdErr)
	assert.Error(t, err)
	if !assert.Contains(t, err.Error(), "unable to get credential helper:") {
		t.Errorf("unexpected output: %s", err.Error())
	}
}

func TestBigBangFailToGetCredentials(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "/tmp/big-bang")
	v.Set("output-config.format", "yaml")
	factory.SetFail.GetCredentialFunction = true
	cmd, cmdErr := NewDeployBigBangCmd(factory)

	// Act
	err := deployBigBangToCluster(cmd, factory, []string{})

	// Assert
	assert.NoError(t, cmdErr)
	assert.Error(t, err)
	if !assert.Contains(t, err.Error(), "unable to get username:") {
		t.Errorf("unexpected output: %s", err.Error())
	}
}

func TestBigBangFailToGetCommandWrapper(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "/tmp/big-bang")
	v.Set("output-config.format", "yaml")
	factory.SetFail.GetCommandWrapper = true
	cmd, cmdErr := NewDeployBigBangCmd(factory)

	// Act
	err := deployBigBangToCluster(cmd, factory, []string{})

	// Assert
	assert.NoError(t, cmdErr)
	assert.Error(t, err)
	if !assert.Contains(t, err.Error(), "unable to get command wrapper:") {
		t.Errorf("unexpected output: %s", err.Error())
	}
}

func TestBigBangFailToGetPipe(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "/tmp/big-bang")
	v.Set("output-config.format", "yaml")
	factory.SetFail.CreatePipe = true
	cmd, cmdErr := NewDeployBigBangCmd(factory)

	// Act
	err := deployBigBangToCluster(cmd, factory, []string{})

	// Assert
	assert.NoError(t, cmdErr)
	assert.Error(t, err)
	if !assert.Contains(t, err.Error(), "unable to create pipe:") {
		t.Errorf("unexpected output: %s", err.Error())
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
