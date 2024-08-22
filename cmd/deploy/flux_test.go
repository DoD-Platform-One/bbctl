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
	assert.Error(t, err)
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

func TestFlux_NewDeployFluxCmd_Run_Text(t *testing.T) {
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

	// Set up the environment and configuration
	bigBangRepoLocation := "/tmp/big-bang"
	assert.Nil(t, os.MkdirAll(bigBangRepoLocation, 0755))
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", bigBangRepoLocation)
	v.Set("output-config.format", "text")

	// Expected output from the command
	expectedOutput := "General Info:\n\nActions:\n  Running command: /tmp/big-bang/scripts/install_flux.sh -u  -p\n\n" // Replace with the actual expected output

	// Act
	cmd := NewDeployFluxCmd(factory)
	assert.Nil(t, err)

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
	assert.Equal(t, "flux", cmd.Use)
	assert.Empty(t, errOut.String())

	// Check the output
	output := out.String()
	assert.Equal(t, expectedOutput, output) // Ensure this matches your actual expected output
}

func TestFlux_NewDeployFluxCmd_Run_Json(t *testing.T) {
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

	// Set up the environment and configuration
	bigBangRepoLocation := "/tmp/big-bang"
	assert.Nil(t, os.MkdirAll(bigBangRepoLocation, 0755))
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", bigBangRepoLocation)
	v.Set("output-config.format", "json")

	// Expected output from the command
	expectedOutput := `{"general_info":{},"actions":["Running command: /tmp/big-bang/scripts/install_flux.sh -u  -p"],"warnings":[]}`

	// Act
	cmd := NewDeployFluxCmd(factory)
	assert.Nil(t, err)

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
	assert.Equal(t, "flux", cmd.Use)
	assert.Empty(t, errOut.String())

	// Check the output
	output := out.String()
	assert.Equal(t, expectedOutput, output) // Ensure this matches your actual expected output
}

func TestFlux_NewDeployFluxCmd_Run_Yaml(t *testing.T) {
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

	// Set up the environment and configuration
	bigBangRepoLocation := "/tmp/big-bang"
	assert.Nil(t, os.MkdirAll(bigBangRepoLocation, 0755))
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", bigBangRepoLocation)
	v.Set("output-config.format", "yaml")

	// Expected output from the command
	expectedOutput := "general_info: {}\nactions:\n- 'Running command: /tmp/big-bang/scripts/install_flux.sh -u  -p'\nwarnings: []\n" // Replace with the actual expected output

	// Act
	cmd := NewDeployFluxCmd(factory)
	assert.Nil(t, err)

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
	assert.Equal(t, "flux", cmd.Use)
	assert.Empty(t, errOut.String())

	// Check the output
	output := out.String()
	assert.Equal(t, expectedOutput, output) // Ensure this matches your actual expected output
}

func TestDeployFluxConfigClientError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	bigBangRepoLocation := "/tmp/big-bang"
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", bigBangRepoLocation)
	v.Set("output-config.format", "yaml")

	cmd := NewDeployFluxCmd(factory)
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

func TestFluxFailToGetConfig(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	loggingClient, _ := factory.GetLoggingClient()
	cmd := NewDeployFluxCmd(factory)
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
	err := deployFluxToCluster(factory, cmd, []string{})

	// Assert
	assert.Error(t, err)
	if !assert.Contains(t, err.Error(), "error getting config:") {
		t.Errorf("unexpected output: %s", err.Error())
	}
}

func TestFluxFailToGetStreams(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "/tmp/big-bang")
	v.Set("output-config.format", "yaml")
	factory.SetFail.GetIOStreams = true
	cmd := NewDeployFluxCmd(factory)

	// Act
	err := deployFluxToCluster(factory, cmd, []string{})

	// Assert
	assert.Error(t, err)
	if !assert.Contains(t, err.Error(), "unable to create IO streams:") {
		t.Errorf("unexpected output: %s", err.Error())
	}
}

func TestFluxFailToGetOutputClient(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "/tmp/big-bang")
	v.Set("output-config.format", "yaml")
	factory.SetFail.GetOutputClient = true
	cmd := NewDeployFluxCmd(factory)

	// Act
	err := deployFluxToCluster(factory, cmd, []string{})

	// Assert
	assert.Error(t, err)
	if !assert.Contains(t, err.Error(), "unable to create output client:") {
		t.Errorf("unexpected output: %s", err.Error())
	}
}

func TestFluxFailToGetCredentialHelper(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "/tmp/big-bang")
	v.Set("output-config.format", "yaml")
	factory.SetFail.GetCredentialHelper = true
	cmd := NewDeployFluxCmd(factory)

	// Act
	err := deployFluxToCluster(factory, cmd, []string{})

	// Assert
	assert.Error(t, err)
	if !assert.Contains(t, err.Error(), "unable to get credential helper:") {
		t.Errorf("unexpected output: %s", err.Error())
	}
}

func TestFluxFailToGetCredentials(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "/tmp/big-bang")
	v.Set("output-config.format", "yaml")
	factory.SetFail.GetCredentialFunction = true
	cmd := NewDeployFluxCmd(factory)

	// Act
	err := deployFluxToCluster(factory, cmd, []string{})

	// Assert
	assert.Error(t, err)
	if !assert.Contains(t, err.Error(), "unable to get username:") {
		t.Errorf("unexpected output: %s", err.Error())
	}
}

func TestFluxFailToGetCommandWrapper(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "/tmp/big-bang")
	v.Set("output-config.format", "yaml")
	factory.SetFail.GetCommandWrapper = true
	cmd := NewDeployFluxCmd(factory)

	// Act
	err := deployFluxToCluster(factory, cmd, []string{})

	// Assert
	assert.Error(t, err)
	if !assert.Contains(t, err.Error(), "unable to get command wrapper:") {
		t.Errorf("unexpected output: %s", err.Error())
	}
}

func TestFluxFailToGetPipe(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "/tmp/big-bang")
	v.Set("output-config.format", "yaml")
	factory.SetFail.CreatePipe = true
	cmd := NewDeployFluxCmd(factory)

	// Act
	err := deployFluxToCluster(factory, cmd, []string{})

	// Assert
	assert.Error(t, err)
	if !assert.Contains(t, err.Error(), "unable to create pipe:") {
		t.Errorf("unexpected output: %s", err.Error())
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
