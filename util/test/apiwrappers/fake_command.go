package apiwrappers

import (
	"io"
	"strings"

	bbUtilApiWrappers "repo1.dso.mil/big-bang/product/packages/bbctl/util/apiwrappers"
)

// FakeCommandRunner - a fake command runner
type FakeCommandRunner struct {
	stdout io.Writer
	stderr io.Writer
	stdin  io.Reader
	name   string
	args   []string
}

// NewFakeCommand - create a new fake command
func NewFakeCommand(name string, args ...string) *bbUtilApiWrappers.Command {
	return bbUtilApiWrappers.NewCommand(
		&FakeCommandRunner{
			name: name,
			args: args,
		},
		name,
		args...,
	)
}

// SetStdout - set the stdout for the command and the underlying runner, if applicable
func (c *FakeCommandRunner) SetStdout(writer io.Writer) {
	c.stdout = writer
}

// SetStderr - set the stderr for the command and the underlying runner, if applicable
func (c *FakeCommandRunner) SetStderr(writer io.Writer) {
	c.stderr = writer
}

// SetStdin - set the stdin for the command and the underlying runner, if applicable
func (c *FakeCommandRunner) SetStdin(reader io.Reader) {
	c.stdin = reader
}

func (c *FakeCommandRunner) Run() error {
	var argsStr strings.Builder
	for _, arg := range c.args {
		argsStr.WriteString(arg)
		argsStr.WriteString(" ")
	}
	_, err := c.stdout.Write([]byte("Running command: " + c.name + " " + argsStr.String() + "\n"))
	return err
}
