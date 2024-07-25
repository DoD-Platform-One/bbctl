package gitlab

import (
	"encoding/base64"
	"fmt"

	gitlab "github.com/xanzy/go-gitlab"
)

// Downloads a single file from a specific branch of a GitLab repository
//
// Repository can be either the project_id or "group/repository" format
//
// Returns the file contents as a decoded bytes array
func getFile(client *gitlab.Client, repository string, branch string, path string) ([]byte, error) {
	opts := &gitlab.GetFileOptions{
		Ref: gitlab.Ptr(branch),
	}

	file, response, err := client.RepositoryFiles.GetFile(repository, path, opts)
	if err != nil {
		return nil, fmt.Errorf("GitLab file download request failed with status code %v: %w", response.StatusCode, err)
	}

	data, fileErr := base64.StdEncoding.DecodeString(file.Content)

	if fileErr != nil {
		return nil, fileErr
	}

	return data, nil
}
