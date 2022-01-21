package helm

import (
	"helm.sh/helm/v3/pkg/release"
)

// Client holds the method signatures for a Helm client.
type Client interface {
	GetRelease(name string) (*release.Release, error)
	GetList() ([]*release.Release, error)
	GetValues(name string, allValues bool) (interface{}, error)
}

type GetReleaseFunc func(string) (*release.Release, error)
type GetListFunc func() ([]*release.Release, error)
type GetValuesFunc func(string) (map[string]interface{}, error)

type HelmClient struct {
	getRelease GetReleaseFunc
	getList    GetListFunc
	getValues  GetValuesFunc
}

// New returns a new Helm client with the provided configuration
func NewClient(getRelease GetReleaseFunc, getList GetListFunc, getValues GetValuesFunc) (Client, error) {
	return &HelmClient{getRelease: getRelease, getList: getList, getValues: getValues}, nil
}

// GetRelease - GetRelease returns a release specified by name.
func (c *HelmClient) GetRelease(name string) (*release.Release, error) {
	return c.getRelease(name)
}

// GetList - getList returns a list of releases
func (c *HelmClient) GetList() ([]*release.Release, error) {
	return c.getList()
}

// GetValues - getValues returns release values
func (c *HelmClient) GetValues(name string, allValues bool) (interface{}, error) {
	return c.getValues(name)
}
