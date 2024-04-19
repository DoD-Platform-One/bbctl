package deploy

import (
	"fmt"
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"

	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"
)

func TestBigBang_NewDeployBigBangCmd(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	streams, _, _, _ := genericIOOptions.NewTestIOStreams()
	// Act
	cmd := NewDeployBigBangCmd(factory, streams)
	// Assert
	assert.NotNil(t, cmd)
	assert.Equal(t, "bigbang", cmd.Use)
}

func TestBigBang_NewDeployBigBangCmd_MissingBigBangRepo(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	streams, _, _, _ := genericIOOptions.NewTestIOStreams()
	viper.Set("big-bang-repo", "")
	// Act
	cmd := NewDeployBigBangCmd(factory, streams)
	// This does panic with a value, but that includes the stack trace so we can't compare it
	assert.Panics(t, func() { assert.Nil(t, cmd.Execute()) })
	// Assert
	assert.NotNil(t, cmd)
	assert.Equal(t, "bigbang", cmd.Use)
}

func TestBigBang_NewDeployBigBangCmd_WithK3d(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	streams, in, out, errout := genericIOOptions.NewTestIOStreams()
	bigBangRepoLocation := "/tmp/big-bang"
	assert.Nil(t, os.MkdirAll(bigBangRepoLocation, 0755))
	viper.Set("big-bang-repo", bigBangRepoLocation)
	expectedCmdString := fmt.Sprintf("Running command: helm upgrade -i bigbang %[1]v/chart -n bigbang --create-namespace --set registryCredentials.username= --set registryCredentials.password= -f %[1]v/chart/ingress-certs.yaml -f %[1]v/docs/assets/configs/example/policy-overrides-k3d.yaml \n", bigBangRepoLocation)
	// Act
	cmd := NewDeployBigBangCmd(factory, streams)
	cmd.SetArgs([]string{"--k3d"})
	assert.Nil(t, cmd.Execute())
	// Assert
	assert.NotNil(t, cmd)
	assert.Equal(t, "bigbang", cmd.Use)
	assert.Empty(t, errout.String())
	assert.Empty(t, in.String())
	assert.Equal(t, expectedCmdString, out.String())
}

func TestBigBang_NewDeployBigBangCmd_WithComponents(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	streams, in, out, errout := genericIOOptions.NewTestIOStreams()
	bigBangRepoLocation := "/tmp/big-bang"
	assert.Nil(t, os.MkdirAll(bigBangRepoLocation, 0755))
	viper.Set("big-bang-repo", bigBangRepoLocation)
	// note that the order of the components is reversed
	expectedCmdString := fmt.Sprintf("Running command: helm upgrade -i bigbang %[1]v/chart -n bigbang --create-namespace --set registryCredentials.username= --set registryCredentials.password= --set addons.baz.enabled=true --set addons.bar.enabled=true --set addons.foo.enabled=true \n", bigBangRepoLocation)
	// Act
	cmd := NewDeployBigBangCmd(factory, streams)
	cmd.SetArgs([]string{"--addon=foo,bar", "--addon=baz"})
	assert.Nil(t, cmd.Execute())
	// Assert
	assert.NotNil(t, cmd)
	assert.Equal(t, "bigbang", cmd.Use)
	assert.Empty(t, errout.String())
	assert.Empty(t, in.String())
	assert.Equal(t, expectedCmdString, out.String())
}
