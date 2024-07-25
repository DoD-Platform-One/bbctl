package gitlab

import (
	"fmt"

	bbGitLab "repo1.dso.mil/big-bang/product/packages/bbctl/util/gitlab"
)

// NewFakeClient - returns a new Fake GitLab client with the provided options
func NewFakeClient(
	baseURL string,
	accessToken string,
) (bbGitLab.Client, error) {
	return &FakeClient{
		baseURL: baseURL,
		token:   accessToken,
	}, nil
}

// FakeClient
type FakeClient struct {
	baseURL string
	token   string
}

func (c *FakeClient) GetFile(repository string, branch string, path string) ([]byte, error) {
	if branch != "main" {
		return nil, fmt.Errorf("Failed to download file from GitLab branch: %v", branch)
	}
	return []byte("String file contents"), nil
}
