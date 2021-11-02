package helmclient

import (
	"helm.sh/helm/v3/pkg/release"
)

// Client holds the method signatures for a Helm client.
type Client interface {
	GetRelease(name string) (*release.Release, error)
	GetList() ([]*release.Release, error)
}
