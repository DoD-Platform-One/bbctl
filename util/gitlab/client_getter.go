package gitlab

import (
	gitlab "github.com/xanzy/go-gitlab"
)

// ClientGetter is an interface for getting a GitLab client
type ClientGetter struct{}

// GetClient returns a new GitLab client
func (clientGetter *ClientGetter) GetClient(baseURL string, accessToken string, options ...gitlab.ClientOptionFunc) (Client, error) {
	return NewClient(baseURL, accessToken, getFile, options...)
}
