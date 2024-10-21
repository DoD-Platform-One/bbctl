package gitlab

import (
	gitlab "github.com/xanzy/go-gitlab"
)

// Client holds the method signatures for a GitLab client.
type Client interface {
	GetFile(repository string, path string, branch string) ([]byte, error)
	GetProject(projectID string) (*gitlab.Project, error)
	GetReleaseArtifact(projectID int, releaseTag string, artifactPath string) ([]byte, error)
}

// NewClient returns a new GitLab client with the provided configuration
func NewClient(baseURL string, accessToken string, getFileFunc GetFileFunc, getProjectFunc GetProjectFunc, getReleaseArtifact GetReleaseArtifactFunc, options ...gitlab.ClientOptionFunc) (Client, error) {
	options = append(options, gitlab.WithBaseURL(baseURL))
	client, err := gitlab.NewClient(accessToken, options...)
	if err != nil {
		return nil, err
	}
	return &gitlabClient{
		client:             client,
		getFile:            getFileFunc,
		getProject:         getProjectFunc,
		getReleaseArtifact: getReleaseArtifact,
	}, nil
}

// gitlabClient is composed of functions that interact with the GitLab v4 API
type gitlabClient struct {
	client             *gitlab.Client
	getFile            GetFileFunc
	getProject         GetProjectFunc
	getReleaseArtifact GetReleaseArtifactFunc
}

type GetFileFunc func(*gitlab.Client, string, string, string) ([]byte, error)

type GetProjectFunc func(*gitlab.Client, string) (*gitlab.Project, error)

type GetReleaseArtifactFunc func(*gitlab.Client, int, string, string) ([]byte, error)

// GetFile downloads a single file from a specific branch of a GitLab repository
//
// Repository can be either the project_id or "group/repository" format
//
// Returns the file contents as a decoded bytes array
func (c *gitlabClient) GetFile(repository string, path string, branch string) ([]byte, error) {
	return c.getFile(c.client, repository, path, branch)
}

// GetProject fetches the project information for a given project on GitLab by project ID
func (c *gitlabClient) GetProject(projectID string) (*gitlab.Project, error) {
	return c.getProject(c.client, projectID)
}

// GetReleaseArtifact attempts to download a given release artifact from a GitLab repository.
//
// The artifactPath is the fully qualified path as made available in the `Links` field of the `Assets` field of the release, as returned by the API. The path argument must match the path in the link completely, at which point it is fetched with an unauthorized HTTP request.
func (c *gitlabClient) GetReleaseArtifact(projectID int, releaseTag string, artifactPath string) ([]byte, error) {
	return c.getReleaseArtifact(c.client, projectID, releaseTag, artifactPath)
}
