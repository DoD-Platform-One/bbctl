package gitlab

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"time"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// Downloads a single file from a specific branch of a GitLab repository
//
// Repository can be either the project_id or "group/repository" format
//
// Returns the file contents as a decoded bytes array
func getFile(client *gitlab.Client, repository string, path string, branch string) ([]byte, error) {
	opts := &gitlab.GetFileOptions{
		Ref: gitlab.Ptr(branch),
	}

	file, response, err := client.RepositoryFiles.GetFile(repository, path, opts)
	if err != nil {
		return nil, fmt.Errorf("error downloading file from gitlab: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error downloading file from gitlab: %s", response.Status)
	}

	data, fileErr := base64.StdEncoding.DecodeString(file.Content)

	if fileErr != nil {
		return nil, fileErr
	}

	return data, nil
}

func getProject(client *gitlab.Client, projectPath string) (*gitlab.Project, error) {
	project, _, err := client.Projects.GetProject(projectPath, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting project from gitlab: %w", err)
	}

	return project, nil
}

func getReleaseArtifact(client *gitlab.Client, projectID int, releaseTag string, artifactPath string) ([]byte, error) {
	var data []byte

	httpClient := &http.Client{
		Timeout: time.Second * 10,
	}

	release, response, err := client.Releases.GetRelease(projectID, releaseTag)
	if err != nil {
		return nil, fmt.Errorf("error getting release from gitlab: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error getting release from gitlab: %s", response.Status)
	}

	for _, asset := range release.Assets.Links {
		if asset.Name == artifactPath {
			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, asset.URL, nil)
			if err != nil {
				return nil, fmt.Errorf("error creating request to download release artifact: %w", err)
			}

			resp, err := httpClient.Do(req)
			if err != nil {
				return nil, fmt.Errorf("error downloading release artifact: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return nil, fmt.Errorf("error downloading release artifact: %s", resp.Status)
			}

			data, err = io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("error reading release artifact: %w", err)
			}
			return data, nil
		}
	}

	return nil, fmt.Errorf("error finding release artifact: %s", artifactPath)
}
