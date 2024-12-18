package filesystem

import (
	"fmt"
	"os"
	"path/filepath"

	commonInterfaces "repo1.dso.mil/big-bang/product/packages/bbctl/util/commoninterfaces"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/filesystem"
)

// FakeClient is a fake implementation of the FileSystem interface.
type FakeClient struct {
	HomeDir string
}

// NewFakeClient returns a new FakeClient
func NewFakeClient(homeDir string) filesystem.Client {
	return &FakeClient{
		HomeDir: homeDir,
	}
}

// UserHomeDir returns the home directory of the current user
//
// This behavior is configurable via the FakeClient.HomeDir field
// which is settable using the Setter on the FakeFactory this client
// is returned from.
func (fsc *FakeClient) UserHomeDir() (string, error) {
	return fsc.HomeDir, nil
}

// Create creates a temporary file
func (fsc *FakeClient) Create(name string) (commonInterfaces.FileLike, error) {
	tempDir, err := os.MkdirTemp("", "testdir-")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}

	tempFile, err := os.CreateTemp(tempDir, filepath.Base(name))
	if err != nil {
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	return tempFile, nil
}
