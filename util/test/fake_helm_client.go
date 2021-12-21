package test

import (
	"fmt"

	"helm.sh/helm/v3/pkg/release"
	helm "repo1.dso.mil/platform-one/big-bang/apps/product-tools/bbctl/util/helm"
)

// NewFakeClient - returns a new Fale Helm client with the provided options
func NewFakeClient(releases []*release.Release) (helm.Client, error) {
	return &FakeClient{releases: releases}, nil
}

// FakeClient - fake client
type FakeClient struct {
	releases []*release.Release
}

// GetRelease - returns a release specified by name.
func (c *FakeClient) GetRelease(name string) (*release.Release, error) {
	for _, r := range c.releases {
		if r.Name == name {
			return r, nil
		}
	}

	return nil, fmt.Errorf("release %s not found", name)
}

// GetList - returns a list of releases
func (c *FakeClient) GetList() ([]*release.Release, error) {
	return c.releases, nil
}

// GetValues - returns a list of releases
func (c *FakeClient) GetValues(name string, allValues bool) (interface{}, error) {
	for _, r := range c.releases {
		if r.Name == name {
			return r.Chart.Values, nil
		}
	}

	return nil, fmt.Errorf("release %s not found", name)
}
