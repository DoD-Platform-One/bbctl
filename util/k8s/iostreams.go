package k8s

import (
	"os"

	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"
)

// GetIOStream initializes and returns a new IOStreams object used to interact with console input, output, and error output
func GetIOStream() genericIOOptions.IOStreams {
	streams := genericIOOptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}

	return streams
}
