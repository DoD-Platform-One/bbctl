package gitlab

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
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
		w.Write([]byte(fileResponse))
	}))
	defer server.Close()

	clientGetter := ClientGetter{}
	client, _ := clientGetter.GetClient(server.URL, "")

	// Act
	file, err := client.GetFile("12345", "FILE", "main")

	// Assert
	assert.Nil(t, err)
	assert.Equal(t, "String file contents", string(file))
}

func TestGetFileNotFound(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 Not Found"))
	}))
	defer server.Close()

	clientGetter := ClientGetter{}
	client, _ := clientGetter.GetClient(server.URL, "")

	// Act
	file, err := client.GetFile("12345", "main", "FILE")

	// Assert
	assert.Nil(t, file)
	assert.NotNil(t, err)
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

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fileResponse))
	}))
	defer server.Close()

	clientGetter := ClientGetter{}
	client, _ := clientGetter.GetClient(server.URL, "")

	// Act
	file, err := client.GetFile("12345", "FILE", "main")

	// Assert
	assert.Nil(t, file)
	assert.NotNil(t, err)
	assert.Equal(t, "illegal base64 data at input byte 17", err.Error())
}
