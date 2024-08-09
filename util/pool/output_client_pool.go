package pool

import (
	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"
	bbOutput "repo1.dso.mil/big-bang/product/packages/bbctl/util/output"
)

// outputClientInstance is a struct that holds an output client and the command and streams it is configured for
type outputClientInstance struct {
	client  bbOutput.Client
	streams genericIOOptions.IOStreams
}

// outputClientPool is a slice of outputClientInstance structs
type outputClientPool []*outputClientInstance

// contains checks if an outputClientPool contains an output client for a given command and streams
func (o outputClientPool) contains(streams genericIOOptions.IOStreams) (bool, bbOutput.Client) {
	for _, client := range o {
		if client.streams == streams {
			return true, client.client
		}
	}
	return false, nil
}

// add adds an output client to the outputClientPool
func (o *outputClientPool) add(client bbOutput.Client, streams genericIOOptions.IOStreams) {
	*o = append(*o, &outputClientInstance{
		client:  client,
		streams: streams,
	})
}
