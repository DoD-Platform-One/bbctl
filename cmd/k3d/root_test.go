package k3d

import (
	"bytes"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
	var commandUseNamesList = make([]string, len(commandsList))
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

	viperInstance, _ := factory.GetViper()
	bigBangRepoLocation := "/tmp/big-bang"
	require.NoError(t, os.MkdirAll(bigBangRepoLocation, 0755))
	viperInstance.Set("big-bang-repo", bigBangRepoLocation)
	viperInstance.Set("kubeconfig", "../../util/test/data/kube-config.yaml")
	factory.SetFail.GetIOStreams = 1
	viperInstance.Set("output-config.format", "text")

	// Act
	cmd, err := NewK3dCmd(factory)

	// Assert
	assert.Nil(t, cmd)
	require.Error(t, err)
	if !assert.Contains(t, err.Error(), "unable to get IOStreams:") {
		t.Errorf("unexpected output: %s", err.Error())
	}
}

func TestK3d_RootNoSubcommand(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()

	viperInstance, _ := factory.GetViper()
	bigBangRepoLocation := "/tmp/big-bang"
	require.NoError(t, os.MkdirAll(bigBangRepoLocation, 0755))
	viperInstance.Set("big-bang-repo", bigBangRepoLocation)
	viperInstance.Set("kubeconfig", "../../util/test/data/kube-config.yaml")
	viperInstance.Set("output-config.format", "text")

	streams, _ := factory.GetIOStream()
	in := streams.In.(*bytes.Buffer)
	out := streams.Out.(*bytes.Buffer)
	errOut := streams.ErrOut.(*bytes.Buffer)
	// Act
	cmd, _ := NewK3dCmd(factory)
	// Assert
	require.NoError(t, cmd.Execute())
	assert.Empty(t, in.String())
	assert.Empty(t, errOut.String())
	assert.Contains(t, out.String(), "Please provide a subcommand for k3d (see help)")
}

func TestK3d_RootSshError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	viperInstance, viperErr := factory.GetViper()
	require.NoError(t, viperErr)
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "../../util/test/data/kube-config.yaml")
	viperInstance.Set("output-config.format", "text")

	expectedError := errors.New("failed to set and bind flag")
	setAndBindFlagFunc := func(_ *bbConfig.ConfigClient, name string, _ string, _ interface{}, _ string) error {
		if name == "ssh-username" {
			return expectedError
		}
		return nil
	}

	logClient, logClientErr := factory.GetLoggingClient()
	require.NoError(t, logClientErr)
	configClient, err := bbConfig.NewClient(nil, setAndBindFlagFunc, &logClient, nil, viperInstance)
	require.NoError(t, err)
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
	require.NoError(t, viperErr)
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("output-config.format", "text")
	viperInstance.Set("kubeconfig", "../../util/test/data/kube-config.yaml")

	expectedError := errors.New("failed to set and bind flag")
	setAndBindFlagFunc := func(_ *bbConfig.ConfigClient, name string, _ string, _ interface{}, _ string) error {
		if name == "private-ip" {
			return expectedError
		}
		return nil
	}

	logClient, logClientErr := factory.GetLoggingClient()
	require.NoError(t, logClientErr)
	configClient, err := bbConfig.NewClient(nil, setAndBindFlagFunc, &logClient, nil, viperInstance)
	require.NoError(t, err)
	factory.SetConfigClient(configClient)

	// Act
	cmd, err := NewK3dCmd(factory)

	// Assert
	assert.Nil(t, cmd)
	if !assert.Contains(t, err.Error(), "error retrieving hosts command:") {
		t.Errorf("unexpected output: %s", err.Error())
	}
}
