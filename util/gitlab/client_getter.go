package gitlab

import (
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// ClientGetter is an interface for getting a GitLab client
type ClientGetter struct{}

// GetClient returns a new GitLab client
func (clientGetter *ClientGetter) GetClient(baseURL string, accessToken string, options ...gitlab.ClientOptionFunc) (Client, error) {
	return NewClient(baseURL, accessToken, getFile, getProject, getReleaseArtifact, options...)
}
