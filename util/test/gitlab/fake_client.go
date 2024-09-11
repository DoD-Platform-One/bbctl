package gitlab

import (
	"fmt"

	bbGitLab "repo1.dso.mil/big-bang/product/packages/bbctl/util/gitlab"
)

type GetFileFunc func(repository string, path string, branch string) ([]byte, error)

// NewFakeClient - returns a new Fake GitLab client with the provided options
func NewFakeClient(
	baseURL string,
	accessToken string,
	getFileFunc GetFileFunc,
) (bbGitLab.Client, error) {
	return &FakeClient{
		baseURL:     baseURL,
		token:       accessToken,
		getFileFunc: getFileFunc,
	}, nil
}

// FakeClient
type FakeClient struct {
	baseURL string
	token   string

	getFileFunc GetFileFunc
}

func (c *FakeClient) GetFile(repository string, path string, branch string) ([]byte, error) {
	if c.getFileFunc != nil {
		return c.getFileFunc(repository, path, branch)
	}

	if branch != "main" {
		return nil, fmt.Errorf("Failed to download file from GitLab branch: %v", branch)
	}
	return []byte("String file contents"), nil

}
