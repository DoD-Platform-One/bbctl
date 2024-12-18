package filesystem

import (
	commonInterfaces "repo1.dso.mil/big-bang/product/packages/bbctl/util/commoninterfaces"
)

// FileSystem is an interface for a client that interacts with the local file system
type Client interface {
	UserHomeDir() (string, error)
	Create(name string) (commonInterfaces.FileLike, error)
}

// UserHomeDirFunc is a function that returns the user's home directory
type UserHomeDirFunc func() (string, error)

// CreateFunc is a function that creates a new file with the given name
type CreateFunc func(name string) (commonInterfaces.FileLike, error)

// fileSystemClient is the implementation of the FileSystem interface
type fileSystemClient struct {
	UserHomeDirFunc UserHomeDirFunc
	CreateFunc      CreateFunc
}

// UserHomeDir returns the user's home directory
func (fsc *fileSystemClient) UserHomeDir() (string, error) {
	return fsc.UserHomeDirFunc()
}

// Create creates a new file with the given name
func (fsc *fileSystemClient) Create(name string) (commonInterfaces.FileLike, error) {
	return fsc.CreateFunc(name)
}

// NewClient injects the behavioral functions and returns a new FileSystem client configured with them
func NewClient(
	userHomeDirFunc UserHomeDirFunc,
	createFunc CreateFunc,
) Client {
	return &fileSystemClient{
		UserHomeDirFunc: userHomeDirFunc,
		CreateFunc:      createFunc,
	}
}
