package filesystem

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"
)

// TestUserHomeDir tests the userHomeDir function.
func TestUserHomeDir(t *testing.T) {
	// Get the expected home directory using os.UserHomeDir
	expectedHomeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	clientGetter := ClientGetter{}
	client := clientGetter.GetClient()

	// Call the function being tested
	homeDir, err := client.UserHomeDir()
	require.NoError(t, err)

	assert.Equal(t, expectedHomeDir, homeDir)
}

// TestCreateFunc tests the createFunc function.
func TestCreateFunc(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "test-createfunc")
	require.NoError(t, err)

	clientGetter := ClientGetter{}
	client := clientGetter.GetClient()

	// Clean up the temporary directory
	defer os.RemoveAll(tempDir)

	// Define the test file path
	testFilePath := filepath.Join(tempDir, "testfile.txt")

	// Call the function being tested
	file, err := client.Create(testFilePath)
	require.NoError(t, err)

	// Ensure the file object is not nil
	if file == nil {
		t.Fatalf("createFunc(%q) returned a nil file object", testFilePath)
	}

	// Close the file
	err = file.Close()
	require.NoError(t, err)

	// Verify the file was created on disk
	if _, err := os.Stat(testFilePath); os.IsNotExist(err) {
		t.Errorf("File %q was not created", testFilePath)
	}
}
