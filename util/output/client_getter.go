package output

import (
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// ClientGetter is an interface for getting an BB output client.
type ClientGetter struct{}

// GetClient returns a new log client.
func (clientGetter *ClientGetter) GetClient(format OutputFormat, streams genericclioptions.IOStreams) Client {
	return NewOutputClient(format, streams)
}
