package k3d

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	bbConfig "repo1.dso.mil/big-bang/product/packages/bbctl/util/config"
	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"
)

func TestK3d_RootUsage(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	// Act
	cmd, _ := NewK3dCmd(factory)
	// Assert
	assert.NotNil(t, cmd)
	assert.Equal(t, "k3d", cmd.Use)
	commandsList := cmd.Commands()
	assert.Len(t, commandsList, 5)
	var commandUseNamesList []string
	for _, command := range commandsList {
		commandUseNamesList = append(commandUseNamesList, command.Use)
	}
	assert.Contains(t, commandUseNamesList, "create")
	assert.Contains(t, commandUseNamesList, "destroy")
	assert.Contains(t, commandUseNamesList, "hosts")
	assert.Contains(t, commandUseNamesList, "shellprofile")
	assert.Contains(t, commandUseNamesList, "ssh")
}

func TestK3d_RootIOStreamError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	bigBangRepoLocation := "test"
	viperInstance, _ := factory.GetViper()
	viperInstance.Set("big-bang-repo", bigBangRepoLocation)
	viperInstance.Set("kubeconfig", "../../util/test/data/kube-config.yaml")
	factory.SetFail.GetIOStreams = true

	// Act
	cmd, err := NewK3dCmd(factory)

	// Assert
	assert.Nil(t, cmd)
	assert.Error(t, err)
	if !assert.Contains(t, err.Error(), "unable to get IOStreams:") {
		t.Errorf("unexpected output: %s", err.Error())
	}
}

func TestK3d_RootNoSubcommand(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	streams, _ := factory.GetIOStream()
	in := streams.In.(*bytes.Buffer)
	out := streams.Out.(*bytes.Buffer)
	errOut := streams.ErrOut.(*bytes.Buffer)
	// Act
	cmd, _ := NewK3dCmd(factory)
	// Assert
	assert.Nil(t, cmd.Execute())
	assert.Empty(t, in.String())
	assert.Empty(t, errOut.String())
	assert.Contains(t, out.String(), "Please provide a subcommand for k3d (see help)")
}

func TestK3d_RootSshError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	viperInstance, viperErr := factory.GetViper()
	assert.Nil(t, viperErr)
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "../../util/test/data/kube-config.yaml")

	expectedError := fmt.Errorf("failed to set and bind flag")
	setAndBindFlagFunc := func(client *bbConfig.ConfigClient, name string, shorthand string, value interface{}, description string) error {
		if name == "ssh-username" {
			return expectedError
		}
		return nil
	}

	logClient, logClientErr := factory.GetLoggingClient()
	assert.Nil(t, logClientErr)
	configClient, err := bbConfig.NewClient(nil, setAndBindFlagFunc, &logClient, nil, viperInstance)
	assert.Nil(t, err)
	factory.SetConfigClient(configClient)

	// Act
	cmd, err := NewK3dCmd(factory)

	// Assert
	assert.Nil(t, cmd)
	if !assert.Contains(t, err.Error(), "error retrieving ssh command:") {
		t.Errorf("unexpected output: %s", err.Error())
	}
}

func TestK3d_RootHostsError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	viperInstance, viperErr := factory.GetViper()
	assert.Nil(t, viperErr)
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "../../util/test/data/kube-config.yaml")

	expectedError := fmt.Errorf("failed to set and bind flag")
	setAndBindFlagFunc := func(client *bbConfig.ConfigClient, name string, shorthand string, value interface{}, description string) error {
		if name == "private-ip" {
			return expectedError
		}
		return nil
	}

	logClient, logClientErr := factory.GetLoggingClient()
	assert.Nil(t, logClientErr)
	configClient, err := bbConfig.NewClient(nil, setAndBindFlagFunc, &logClient, nil, viperInstance)
	assert.Nil(t, err)
	factory.SetConfigClient(configClient)

	// Act
	cmd, err := NewK3dCmd(factory)

	// Assert
	assert.Nil(t, cmd)
	if !assert.Contains(t, err.Error(), "error retrieving hosts command:") {
		t.Errorf("unexpected output: %s", err.Error())
	}
}
