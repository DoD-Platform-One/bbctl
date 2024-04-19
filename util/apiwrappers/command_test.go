package apiwrappers

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testReader struct {
	data        string
	shouldError bool
	hasBeenRead bool
}

func (r *testReader) Read(p []byte) (n int, err error) {
	if r.hasBeenRead {
		return 0, io.EOF
	}
	r.hasBeenRead = true
	if r.shouldError {
		return 0, io.ErrUnexpectedEOF
	}
	copy(p, r.data)
	return len(r.data), nil
}

func TestCommandNewExecRunner(t *testing.T) {
	// Arrange
	name := "ls"
	args := []string{"-l"}
	// Act
	cmd := NewExecRunner(name, args...)
	// Assert
	assert.NotNil(t, cmd)
	assert.Equal(t, args[0], cmd.args[0])
	assert.Equal(t, name, cmd.name)
	assert.Contains(t, cmd.cmd.(*ExecCommand).cmd.Path, name)
	assert.Equal(t, name, cmd.cmd.(*ExecCommand).cmd.Args[0])
	assert.Equal(t, args, cmd.cmd.(*ExecCommand).cmd.Args[1:])
	assert.Len(t, cmd.cmd.(*ExecCommand).cmd.Args, len(args)+1)
}

func TestCommandNewExecRunnerRunEcho(t *testing.T) {
	// Arrange
	name := "echo"
	msg := "hello"
	args := []string{msg}
	var stdout strings.Builder
	var stderr strings.Builder
	stdin := testReader{data: msg}
	// Act
	cmd := NewExecRunner(name, args...)
	cmd.SetStdout(&stdout)
	cmd.SetStderr(&stderr)
	cmd.SetStdin(&stdin)
	assert.Nil(t, cmd.Run())
	// Assert
	assert.NotNil(t, cmd)
	assert.Equal(t, msg, strings.TrimSpace(stdout.String()))
	assert.Empty(t, stderr.String())
	assert.True(t, stdin.hasBeenRead)
}
