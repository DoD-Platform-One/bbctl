package gitlab

import (
	gitlab "github.com/xanzy/go-gitlab"
)

// Client holds the method signatures for a GitLab client.
type Client interface {
	GetFile(string, string, string) ([]byte, error)
}

// NewClient returns a new GitLab client with the provided configuration
func NewClient(baseURL string, accessToken string, getFileFunc GetFileFunc) (Client, error) {
	client, err := gitlab.NewClient(accessToken, gitlab.WithBaseURL(baseURL))
	if err != nil {
		return nil, err
	}
	return &gitlabClient{
		client:  client,
		getFile: getFileFunc,
	}, nil
}

// gitlabClient is composed of functions that interact with the GitLab v4 API
type gitlabClient struct {
	client  *gitlab.Client
	getFile GetFileFunc
}

type GetFileFunc func(*gitlab.Client, string, string, string) ([]byte, error)

// Downloads a single file from a specific branch of a GitLab repository
//
// Repository can be either the project_id or "group/repository" format
//
// Returns the file contents as a decoded bytes array
func (c *gitlabClient) GetFile(repository string, path string, branch string) ([]byte, error) {
	return c.getFile(c.client, repository, path, branch)
}
