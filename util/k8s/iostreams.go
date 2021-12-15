package k8s

import (
	"os"

	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// GetIOStream - get io stream
func GetIOStream() genericclioptions.IOStreams {

	streams := genericclioptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}

	return streams
}
