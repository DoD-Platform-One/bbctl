package gitlab

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

func TestGetValidFile(t *testing.T) {
	// Arrange
	route := "/api/v4/projects/12345/repository/files/FILE"
	fileResponse := `{
		"file_name": "FILE",
		"file_path": "FILE",
		"size": 25,
		"encoding": "base64",
		"content": "U3RyaW5nIGZpbGUgY29udGVudHM=",
		"content_sha256": "4c294617b60715c1d218e61164a3abd4808a4284cbc30e6728a01ad9aada4481",
		"ref": "main",
		"blob_id": "79f7bbd25901e8334750839545a9bd021f0e4c83",
		"commit_id": "d5a3ff139356ce33e37e73add446f16869741b50",
		"last_commit_id": "570e7b2abdd848b95f2f578043fc23bd6f6fd24d",
		"execute_filemode": false
	}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != route {
			t.Errorf("Expected request, got: %s", r.URL.Path)
		}

		if branch := r.URL.Query().Get("ref"); branch != "main" {
			t.Errorf("Expected branch to be main, got: %s", branch)
		}

		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(fileResponse))
		assert.NoError(t, err)
	}))
	defer server.Close()

	clientGetter := ClientGetter{}
	client, _ := clientGetter.GetClient(server.URL, "")

	// Act
	file, err := client.GetFile("12345", "FILE", "main")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "String file contents", string(file))
}

func TestGetFileNotFound(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, err := w.Write([]byte("404 Not Found"))
		assert.NoError(t, err)
	}))
	defer server.Close()

	clientGetter := ClientGetter{}
	client, _ := clientGetter.GetClient(server.URL, "")

	// Act
	file, err := client.GetFile("12345", "main", "FILE")

	// Assert
	assert.Nil(t, file)
	require.Error(t, err)
	assert.Equal(t, "error downloading file from gitlab: 404 Not Found", err.Error())
}

func TestGetFileEncodingError(t *testing.T) {
	// Arrange
	fileResponse := `{
		"file_name": "FILE",
		"file_path": "FILE",
		"size": 25,
		"encoding": "base64",
		"content": "U3RyaW5nIGZpbGUgY$@(*&^%#*(&^%*&29udGV",
		"content_sha256": "4c294617b60715c1d218e61164a3abd4808a4284cbc30e6728a01ad9aada4481",
		"ref": "main",
		"blob_id": "79f7bbd25901e8334750839545a9bd021f0e4c83",
		"commit_id": "d5a3ff139356ce33e37e73add446f16869741b50",
		"last_commit_id": "570e7b2abdd848b95f2f578043fc23bd6f6fd24d",
		"execute_filemode": false
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(fileResponse))
		assert.NoError(t, err)
	}))
	defer server.Close()

	clientGetter := ClientGetter{}
	client, _ := clientGetter.GetClient(server.URL, "")

	// Act
	file, err := client.GetFile("12345", "FILE", "main")

	// Assert
	assert.Nil(t, file)
	require.Error(t, err)
	assert.Equal(t, "illegal base64 data at input byte 17", err.Error())
}

func TestGetValidProject(t *testing.T) {
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request is for the correct endpoint
		if r.URL.Path != "/api/v4/projects/grafana" {
			t.Errorf("Expected to request '/api/v4/projects/grafana', got: %s", r.URL.Path)
		}

		// Check if the request method is GET
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got: %s", r.Method)
		}

		// Create a mock project response
		project := gitlab.Project{
			ID:                1,
			Name:              "Grafana",
			WebURL:            "https://repo1.dso.mil/big-bang/product/packages/grafana",
			NameWithNamespace: "Big Bang / Universe / Product / Grafana",
			PathWithNamespace: "big-bang/product/packages/grafana",
		}

		// Marshal the project to JSON
		json.NewEncoder(w).Encode(project)
	}))
	defer server.Close()

	clientGetter := ClientGetter{}
	gitlabClient, err := clientGetter.GetClient(server.URL, "")
	require.NoError(t, err)

	// Call the getProject function
	projectPath := "grafana"
	gotProject, err := gitlabClient.GetProject(projectPath)
	require.NoError(t, err)

	// Assert that the returned project matches the expected values
	assert.NotNil(t, gotProject)
	assert.Equal(t, 1, gotProject.ID)
	assert.Equal(t, "Grafana", gotProject.Name)
	assert.Equal(t, "https://repo1.dso.mil/big-bang/product/packages/grafana", gotProject.WebURL)
	assert.Equal(t, "big-bang/product/packages/grafana", gotProject.PathWithNamespace)
	assert.Equal(t, "Big Bang / Universe / Product / Grafana", gotProject.NameWithNamespace)
}

func TestGetInvalidProject(t *testing.T) {
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]string{
			"message": "404 Project Not Found",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	clientGetter := ClientGetter{}
	gitlabClient, err := clientGetter.GetClient(server.URL, "")
	require.NoError(t, err)

	// Call the getProject function
	projectPath := "invalid"
	gotProject, err := gitlabClient.GetProject(projectPath)

	// Assert that the returned project matches the expected values
	assert.Nil(t, gotProject)
	assert.Equal(t, "error getting project from gitlab: 404 Not Found", err.Error())
}

func TestGetReleaseArtifact(t *testing.T) {
	var serverURL string
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v4/projects/1/releases/v1.0.0":
			// Mock the response for getting the release
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			releaseJSON := fmt.Sprintf(`{
				"tag_name": "v1.0.0",
				"assets": {
					"links": [
						{
							"name": "images.txt",
							"url": "%s/download/images.txt"
						}
					]
				}
			}`, serverURL)
			w.Write([]byte(releaseJSON))
		case "/download/images.txt":
			// Mock the response for downloading the artifact
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("images"))
		default:
			http.Error(w, "Not found", http.StatusNotFound)
		}
	}))
	defer server.Close()

	serverURL = server.URL

	clientGetter := ClientGetter{}
	gitlabClient, err := clientGetter.GetClient(serverURL, "")
	require.NoError(t, err)

	artifact, err := gitlabClient.GetReleaseArtifact(1, "v1.0.0", "images.txt")
	require.NoError(t, err)

	assert.Equal(t, []byte("images"), artifact)
}

func TestGetReleaseArtifactErrorGettingRelease(t *testing.T) {
	var serverURL string
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v4/projects/1/releases/v1.0.0":
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("images"))
		default:
			http.Error(w, "Not found", http.StatusNotFound)
		}
	}))
	defer server.Close()

	serverURL = server.URL

	clientGetter := ClientGetter{}
	gitlabClient, err := clientGetter.GetClient(serverURL, "")
	require.NoError(t, err)

	artifact, err := gitlabClient.GetReleaseArtifact(1, "v1.0.0", "images.txt")
	assert.Nil(t, artifact)

	expected := fmt.Sprintf("error getting release from gitlab: GET %s/api/v4/projects/1/releases/v1.0.0: 500 failed to parse unknown error format: images", serverURL)

	assert.Equal(t, expected, err.Error())
}

func TestGetReleaseArtifactErrorDownloading(t *testing.T) {
	var serverURL string
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v4/projects/1/releases/v1.0.0":
			// Mock the response for getting the release
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			releaseJSON := fmt.Sprintf(`{
				"tag_name": "v1.0.0",
				"assets": {
					"links": [
						{
							"name": "images.txt",
							"url": "%s/download/images.txt"
						}
					]
				}
			}`, serverURL)
			w.Write([]byte(releaseJSON))
		default:
			http.Error(w, "Not found", http.StatusNotFound)
		}
	}))
	defer server.Close()

	serverURL = server.URL

	clientGetter := ClientGetter{}
	gitlabClient, err := clientGetter.GetClient(serverURL, "")
	require.NoError(t, err)

	artifact, err := gitlabClient.GetReleaseArtifact(1, "v1.0.0", "images.txt")
	assert.Nil(t, artifact)

	assert.Equal(t, "error downloading release artifact: 404 Not Found", err.Error())
}

func TestGetReleaseArtifactMissingArtifact(t *testing.T) {
	var serverURL string
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v4/projects/1/releases/v1.0.0":
			// Mock the response for getting the release
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			releaseJSON := fmt.Sprintf(`{
				"tag_name": "v1.0.0",
				"assets": {
					"links": [
						{
							"name": "images.txt",
							"url": "%s/download/images.txt"
						}
					]
				}
			}`, serverURL)
			w.Write([]byte(releaseJSON))
		default:
			http.Error(w, "Not found", http.StatusNotFound)
		}
	}))
	defer server.Close()

	serverURL = server.URL

	clientGetter := ClientGetter{}
	gitlabClient, err := clientGetter.GetClient(serverURL, "")
	require.NoError(t, err)

	artifact, err := gitlabClient.GetReleaseArtifact(1, "v1.0.0", "missing.txt")
	assert.Nil(t, artifact)

	assert.Equal(t, "error finding release artifact: missing.txt", err.Error())
}
