package helm

import (
	"fmt"

	helm "repo1.dso.mil/big-bang/product/packages/bbctl/util/helm"

	"helm.sh/helm/v3/pkg/release"
)

// NewFakeClient returns a new Helm client with the provided configuration
func NewFakeClient(getRelease helm.GetReleaseFunc, getList helm.GetListFunc, getValues helm.GetValuesFunc, releases []*release.Release) (helm.Client, error) {
	return &FakeClient{getRelease: getRelease, getList: getList, getValues: getValues, releases: releases}, nil
}

// FakeClient - fake client
type FakeClient struct {
	releases []*release.Release

	getRelease helm.GetReleaseFunc
	getList    helm.GetListFunc
	getValues  helm.GetValuesFunc
}

// GetRelease - returns a release specified by name.
func (c *FakeClient) GetRelease(name string) (*release.Release, error) {

	if c.getRelease != nil {
		return c.getRelease(name)
	}
	for _, r := range c.releases {
		if r.Name == name {
			return r, nil
		}
	}

	return nil, fmt.Errorf("release %s not found", name)
}

// GetList - returns a list of releases
func (c *FakeClient) GetList() ([]*release.Release, error) {
	if c.getList != nil {
		return c.getList()
	}
	return c.releases, nil
}

// GetValues - returns a list of releases
func (c *FakeClient) GetValues(name string) (interface{}, error) {
	if c.getValues != nil {
		return c.getValues(name)
	}

	for _, r := range c.releases {
		if r.Name == name {
			return r.Chart.Values, nil
		}
	}

	return nil, fmt.Errorf("release %s not found", name)
}
