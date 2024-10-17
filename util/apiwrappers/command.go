package apiwrappers

import (
	"io"
	"os/exec"
)

// Runnable - describes a thing that can be run
type Runnable interface {
	Run() error
}

// Streamable - describes a thing that can have its stdout and stderr set
type Streamable interface {
	SetStdout(writer io.Writer)
	SetStderr(writer io.Writer)
	SetStdin(reader io.Reader)
}

// Commandable - describes a thing that can be run and have its stdout and stderr set
type Commandable interface {
	Runnable
	Streamable
}

// Command - a wrapper for a command to be run
type Command struct {
	stdout io.Writer
	stderr io.Writer
	stdin  io.Reader
	name   string
	args   []string
	cmd    Runnable
}

// ExecCommand - a wrapper for an `*exec.cmd` to be run
type ExecCommand struct {
	Command
	cmd *exec.Cmd
}

// NewCommand - create a new command
func NewCommand(runner Runnable, name string, args ...string) *Command {
	return &Command{
		name: name,
		args: args,
		cmd:  runner,
	}
}

// NewExecRunner - create a new exec runner
func NewExecRunner(name string, args ...string) *Command {
	return NewCommand(&ExecCommand{
		Command: Command{
			name: name,
			args: args,
		},
		cmd: exec.Command(name, args...),
	}, name, args...)
}

// SetStdout - set the stdout for the command and the underlying runner, if applicable
func (c *Command) SetStdout(writer io.Writer) {
	c.stdout = writer
	if s, ok := c.cmd.(Streamable); ok {
		s.SetStdout(writer)
	}
}

// SetStderr - set the stderr for the command and the underlying runner, if applicable
func (c *Command) SetStderr(writer io.Writer) {
	c.stderr = writer
	if s, ok := c.cmd.(Streamable); ok {
		s.SetStderr(writer)
	}
}

// SetStdin - set the stdin for the command and the underlying runner, if applicable
func (c *Command) SetStdin(reader io.Reader) {
	c.stdin = reader
	if s, ok := c.cmd.(Streamable); ok {
		s.SetStdin(reader)
	}
}

// Run - run the command
func (c *Command) Run() error {
	return c.cmd.Run()
}

// SetStdout - set the stdout for the command
func (c *ExecCommand) SetStdout(writer io.Writer) {
	c.stdout = writer
	c.cmd.Stdout = writer
}

// SetStderr - set the stderr for the command
func (c *ExecCommand) SetStderr(writer io.Writer) {
	c.stderr = writer
	c.cmd.Stderr = writer
}

// SetStdin - set the stdin for the command
func (c *ExecCommand) SetStdin(reader io.Reader) {
	c.stdin = reader
	c.cmd.Stdin = reader
}

// Run - run the command
func (c *ExecCommand) Run() error {
	return c.cmd.Run()
}
