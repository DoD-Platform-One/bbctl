package helm

import (
	"helm.sh/helm/v3/pkg/release"
)

// Client holds the method signatures for a Helm client.
type Client interface {
	GetRelease(name string) (*release.Release, error)
	GetList() ([]*release.Release, error)
	GetValues(name string) (map[string]interface{}, error)
}

// GetReleaseFunc type
type GetReleaseFunc func(string) (*release.Release, error)

// GetListFunc type
type GetListFunc func() ([]*release.Release, error)

// GetValuesFunc type
type GetValuesFunc func(string) (map[string]interface{}, error)

// helmClient is composed of functions to interact with Helm API
type helmClient struct {
	getRelease GetReleaseFunc
	getList    GetListFunc
	getValues  GetValuesFunc
}

// NewClient returns a new Helm client with the provided configuration
func NewClient(getRelease GetReleaseFunc, getList GetListFunc, getValues GetValuesFunc) (Client, error) {
	return &helmClient{getRelease: getRelease, getList: getList, getValues: getValues}, nil
}

// GetRelease - GetRelease returns a release specified by name.
func (c *helmClient) GetRelease(name string) (*release.Release, error) {
	return c.getRelease(name)
}

// GetList - getList returns a list of releases
func (c *helmClient) GetList() ([]*release.Release, error) {
	return c.getList()
}

// GetValues - getValues returns release values
func (c *helmClient) GetValues(name string) (map[string]interface{}, error) {
	return c.getValues(name)
}
