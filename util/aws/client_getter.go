package aws

// ClientGetter is an interface for getting an BB AWS client.
type ClientGetter struct{}

// GetClient returns a new AWS client.
func (clientGetter *ClientGetter) GetClient() Client {
	return NewClient(config, getClusterIPs, getSortedClusterIPs, getEc2Client, getIdentity, getStsClient)
}
