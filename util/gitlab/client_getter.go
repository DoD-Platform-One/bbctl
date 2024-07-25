package gitlab

// ClientGetter is an interface for getting a GitLab client
type ClientGetter struct{}

// GetClient returns a new GitLab client
func (clientGetter *ClientGetter) GetClient(baseURL string, accessToken string) (Client, error) {
	return NewClient(baseURL, accessToken, getFile)
}
