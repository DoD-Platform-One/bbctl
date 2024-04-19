package aws

import bbLog "repo1.dso.mil/big-bang/product/packages/bbctl/util/log"

// ClientGetter is an interface for getting an BB AWS client.
type ClientGetter struct{}

// GetClient returns a new AWS client.
func (clientGetter *ClientGetter) GetClient(loggingClient bbLog.Client) (Client, error) {
	return NewClient(config, getClusterIPs, getSortedClusterIPs, getEc2Client, getIdentity, getStsClient, loggingClient)
}
