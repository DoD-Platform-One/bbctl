package deploy

import (
	"bytes"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
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
	streams := factory.GetIOStream()
	in := streams.In.(*bytes.Buffer)
	out := streams.Out.(*bytes.Buffer)
	errOut := streams.ErrOut.(*bytes.Buffer)
	factory.GetViper().Set("big-bang-repo", "")

	// Act
	cmd := NewDeployFluxCmd(factory)
	// This does panic with a value, but that includes the stack trace so we can't compare it
	assert.Panics(t, func() { assert.Nil(t, cmd.Execute()) })

	// Assert
	assert.NotNil(t, cmd)
	assert.Equal(t, "flux", cmd.Use)
	assert.Empty(t, in.String())
	assert.Empty(t, out.String())
	assert.Empty(t, errOut.String())
}

func TestFlux_NewDeployFluxCmd_Run(t *testing.T) {
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
	expectedCmdString := "Running command: /tmp/big-bang/scripts/install_flux.sh -u  -p  \n"

	// Act
	cmd := NewDeployFluxCmd(factory)
	assert.Nil(t, cmd.Execute())

	// Assert
	assert.NotNil(t, cmd)
	assert.Equal(t, "flux", cmd.Use)
	assert.Empty(t, errOut.String())
	assert.Empty(t, in.String())
	assert.Equal(t, expectedCmdString, out.String())
}

func TestFlux_NewDeployFluxFailToGetConfigClient(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	streams := factory.GetIOStream()
	in := streams.In.(*bytes.Buffer)
	out := streams.Out.(*bytes.Buffer)
	errOut := streams.ErrOut.(*bytes.Buffer)
	bigBangRepoLocation := "/tmp/big-bang"
	factory.GetViper().Set("big-bang-repo", bigBangRepoLocation)

	// Act
	if os.Getenv("BE_CRASHER") == "1" {
		cmd := NewDeployFluxCmd(factory)
		factory.SetFail.GetConfigClient = true
		cmd.Run(cmd, []string{})
		return
	}
	runCrasherCommand := exec.Command(os.Args[0], "-test.run=TestFlux_NewDeployFluxFailToGetConfigClient")
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
