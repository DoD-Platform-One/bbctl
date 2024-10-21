package gitlab

import (
	"fmt"

	"github.com/xanzy/go-gitlab"
	bbGitLab "repo1.dso.mil/big-bang/product/packages/bbctl/util/gitlab"
)

type GetFileFunc func(repository string, path string, branch string) ([]byte, error)
type GetProjectFunc func(projectID string) (*gitlab.Project, error)
type GetReleaseArtifactFunc func(projectId int, releaseTag string, artifactPath string) ([]byte, error)

// NewFakeClient - returns a new Fake GitLab client with the provided options
func NewFakeClient(
	baseURL string,
	accessToken string,
	getFileFunc GetFileFunc,
	getProjectFunc GetProjectFunc,
	getReleaseArtifactFunc GetReleaseArtifactFunc,
) (bbGitLab.Client, error) {
	return &FakeClient{
		baseURL:                baseURL,
		token:                  accessToken,
		getFileFunc:            getFileFunc,
		getProjectFunc:         getProjectFunc,
		getReleaseArtifactFunc: getReleaseArtifactFunc,
	}, nil
}

// FakeClient
type FakeClient struct {
	baseURL string
	token   string

	getFileFunc            GetFileFunc
	getProjectFunc         GetProjectFunc
	getReleaseArtifactFunc GetReleaseArtifactFunc
}

func (c *FakeClient) GetFile(repository string, path string, branch string) ([]byte, error) {
	if c.getFileFunc != nil {
		return c.getFileFunc(repository, path, branch)
	}

	if branch != "main" {
		return nil, fmt.Errorf("failed to download file from GitLab branch: %v", branch)
	}
	return []byte("String file contents"), nil
}

func (c *FakeClient) GetReleaseArtifact(projectId int, releaseTag string, artifactPath string) ([]byte, error) {
	if c.getReleaseArtifactFunc != nil {
		return c.getReleaseArtifactFunc(projectId, releaseTag, artifactPath)
	}
	return []byte("grafana:1.0.3\ntempo:3.2.1\n"), nil
}

func (c *FakeClient) GetProject(projectID string) (*gitlab.Project, error) {
	if c.getProjectFunc != nil {
		return c.getProjectFunc(projectID)
	}

	return &gitlab.Project{
		ID:   1,
		Path: "/path/to/project",
	}, nil
}
