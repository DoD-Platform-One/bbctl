package deploy

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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
	factory.GetViper().Set("big-bang-repo", "")
	// Act
	cmd, _ := NewDeployBigBangCmd(factory)
	// This does panic with a value, but that includes the stack trace so we can't compare it
	assert.Panics(t, func() { assert.Nil(t, cmd.Execute()) })
	// Assert
	assert.NotNil(t, cmd)
	assert.Equal(t, "bigbang", cmd.Use)
}

func TestBigBang_NewDeployBigBangCmd_WithK3d(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	streams := factory.GetIOStream()
	in := streams.In.(*bytes.Buffer)
	out := streams.Out.(*bytes.Buffer)
	errOut := streams.ErrOut.(*bytes.Buffer)
	bigBangRepoLocation := "/tmp/big-bang"
	assert.Nil(t, os.MkdirAll(bigBangRepoLocation, 0755))
	factory.GetViper().Set("big-bang-repo", bigBangRepoLocation)
	expectedCmdString := fmt.Sprintf("Running command: helm upgrade -i bigbang %[1]v/chart -n bigbang --create-namespace --set registryCredentials.username= --set registryCredentials.password= -f %[1]v/chart/ingress-certs.yaml -f %[1]v/docs/assets/configs/example/policy-overrides-k3d.yaml \n", bigBangRepoLocation)
	// Act
	cmd, _ := NewDeployBigBangCmd(factory)
	cmd.SetArgs([]string{"--k3d"})
	assert.Nil(t, cmd.Execute())
	// Assert
	assert.NotNil(t, cmd)
	assert.Equal(t, "bigbang", cmd.Use)
	assert.Empty(t, errOut.String())
	assert.Empty(t, in.String())
	assert.Equal(t, expectedCmdString, out.String())
}

func TestBigBang_NewDeployBigBangCmd_WithComponents(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	streams := factory.GetIOStream()
	in := streams.In.(*bytes.Buffer)
	out := streams.Out.(*bytes.Buffer)
	errOut := streams.ErrOut.(*bytes.Buffer)
	bigBangRepoLocation := "/tmp/big-bang"
	assert.Nil(t, os.MkdirAll(bigBangRepoLocation, 0755))
	factory.GetViper().Set("big-bang-repo", bigBangRepoLocation)
	// note that the order of the components is reversed
	expectedCmdString := fmt.Sprintf("Running command: helm upgrade -i bigbang %[1]v/chart -n bigbang --create-namespace --set registryCredentials.username= --set registryCredentials.password= --set addons.baz.enabled=true --set addons.bar.enabled=true --set addons.foo.enabled=true \n", bigBangRepoLocation)
	// Act
	cmd, _ := NewDeployBigBangCmd(factory)
	cmd.SetArgs([]string{"--addon=foo,bar", "--addon=baz"})
	assert.Nil(t, cmd.Execute())
	// Assert
	assert.NotNil(t, cmd)
	assert.Equal(t, "bigbang", cmd.Use)
	assert.Empty(t, errOut.String())
	assert.Empty(t, in.String())
	assert.Equal(t, expectedCmdString, out.String())
}

func TestGetBigBangCmdConfigClientError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	bigBangRepoLocation := "/tmp/big-bang"
	factory.GetViper().Set("big-bang-repo", bigBangRepoLocation)
	factory.SetFail.GetConfigClient = true
	// Act
	cmd, err := NewDeployBigBangCmd(factory)
	// Assert
	assert.Nil(t, cmd)
	assert.Error(t, err)
	if !assert.Contains(t, err.Error(), "Unable to get config client:") {
		t.Errorf("unexpected output: %s", err.Error())
	}
}
func TestDeployBigBangConfigClientError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	bigBangRepoLocation := "/tmp/big-bang"
	factory.GetViper().Set("big-bang-repo", bigBangRepoLocation)
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
