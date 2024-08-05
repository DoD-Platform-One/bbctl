package k3d

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	bbConfig "repo1.dso.mil/big-bang/product/packages/bbctl/util/config"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/config/schemas"
	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"
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

func TestK3d_NewCreateClusterCmd_RunWithMissingBigBangRepo(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	// Act
	cmd := NewCreateClusterCmd(factory)
	err := cmd.RunE(cmd, []string{})
	// Assert
	assert.NotNil(t, cmd)
	assert.Error(t, err)
	if !assert.Contains(t, err.Error(), "Error:Field validation for 'BigBangRepo' failed on the 'required' tag") {
		t.Errorf("unexpected output: %s", err.Error())
	}
	assert.Equal(t, "create", cmd.Use)
}

func TestK3d_NewCreateClusterCmd_Run(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	streams := factory.GetIOStream()
	in := streams.In.(*bytes.Buffer)
	out := streams.Out.(*bytes.Buffer)
	errOut := streams.ErrOut.(*bytes.Buffer)
	bigBangRepoLocation := "/tmp/big-bang"
	factory.GetViper().Set("big-bang-repo", bigBangRepoLocation)
	expectedCmdString := "Running command: /tmp/big-bang/docs/assets/scripts/developer/k3d-dev.sh \n"
	// Act
	cmd := NewCreateClusterCmd(factory)
	assert.Nil(t, cmd.Execute())
	// Assert
	assert.NotNil(t, cmd)
	assert.Equal(t, "create", cmd.Use)
	assert.Empty(t, errOut.String())
	assert.Empty(t, in.String())
	assert.Equal(t, expectedCmdString, out.String())
}

func TestK3d_CreateFailToGetConfigClient(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	bigBangRepoLocation := "/tmp/big-bang"
	factory.GetViper().Set("big-bang-repo", bigBangRepoLocation)
	cmd := NewCreateClusterCmd(factory)
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

func TestCreateFailToGetConfig(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	loggingClient := factory.GetLoggingClient()
	cmd := NewCreateClusterCmd(factory)
	viper := factory.GetViper()
	expected := ""
	getConfigFunc := func(client *bbConfig.ConfigClient) (*schemas.GlobalConfiguration, error) {
		return &schemas.GlobalConfiguration{
			BigBangRepo: expected,
		}, fmt.Errorf("Dummy Error")
	}
	client, _ := bbConfig.NewClient(getConfigFunc, nil, &loggingClient, cmd, viper)
	factory.SetConfigClient(client)

	// Act
	err := createCluster(factory, cmd, []string{})

	// Assert
	assert.Error(t, err)
	if !assert.Contains(t, err.Error(), "error getting config:") {
		t.Errorf("unexpected output: %s", err.Error())
	}
}
