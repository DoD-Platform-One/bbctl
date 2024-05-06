package k3d

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"

	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"
)

func TestK3d_NewCreateClusterCmd(t *testing.T) {
	// Arrange
	streams, _, _, _ := genericIOOptions.NewTestIOStreams()
	factory := bbTestUtil.GetFakeFactory()
	// Act
	cmd := NewCreateClusterCmd(factory, streams)
	// Assert
	assert.NotNil(t, cmd)
	assert.Equal(t, "create", cmd.Use)
}

func TestK3d_NewCreateClusterCmd_RunWithMissingBigBangRepo(t *testing.T) {
	// Arrange
	streams, _, _, _ := genericIOOptions.NewTestIOStreams()
	factory := bbTestUtil.GetFakeFactory()
	// Act
	cmd := NewCreateClusterCmd(factory, streams)
	assert.Panics(t, func() { assert.Nil(t, cmd.Execute()) })
	// Assert
	assert.NotNil(t, cmd)
	assert.Equal(t, "create", cmd.Use)
}

func TestK3d_NewCreateClusterCmd_Run(t *testing.T) {
	// Arrange
	streams, in, out, errout := genericIOOptions.NewTestIOStreams()
	factory := bbTestUtil.GetFakeFactory()
	bigBangRepoLocation := "/tmp/big-bang"
	factory.GetViper().Set("big-bang-repo", bigBangRepoLocation)
	expectedCmdString := "Running command: /tmp/big-bang/docs/assets/scripts/developer/k3d-dev.sh \n"
	// Act
	cmd := NewCreateClusterCmd(factory, streams)
	assert.Nil(t, cmd.Execute())
	// Assert
	assert.NotNil(t, cmd)
	assert.Equal(t, "create", cmd.Use)
	assert.Empty(t, errout.String())
	assert.Empty(t, in.String())
	assert.Equal(t, expectedCmdString, out.String())
}

func TestK3d_CreateFailToGetConfigClient(t *testing.T) {
	// Arrange
	streams, in, out, errOut := genericIOOptions.NewTestIOStreams()
	factory := bbTestUtil.GetFakeFactory()
	bigBangRepoLocation := "/tmp/big-bang"
	factory.GetViper().Set("big-bang-repo", bigBangRepoLocation)
	factory.SetFail.GetConfigClient = true

	// Act
	if os.Getenv("BE_CRASHER") == "1" {
		cmd := NewCreateClusterCmd(factory, streams)
		cmd.Run(cmd, []string{})
		return
	}
	runCrasherCommand := exec.Command(os.Args[0], "-test.run=TestK3d_CreateFailToGetConfigClient")
	runCrasherCommand.Env = append(os.Environ(), "BE_CRASHER=1")
	runCrasherCommand.Stderr = errOut
	runCrasherCommand.Stdout = out
	runCrasherCommand.Stdin = in
	err := runCrasherCommand.Run()

	// Assert
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		assert.Equal(t, 1, e.ExitCode())
		assert.NotNil(t, runCrasherCommand)
		assert.Equal(t, "exit status 1", e.Error())
		assert.Equal(t, "error: failed to get config client\n", errOut.String())
		assert.Empty(t, in.String())
		assert.Empty(t, out.String())
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}
