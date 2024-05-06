package helm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/release"
)

func TestGetList(t *testing.T) {
	var f GetListFunc = func() ([]*release.Release, error) {
		releaseFixture := []*release.Release{
			{
				Name:      "foo",
				Version:   1,
				Namespace: "bigbang",
			},
		}
		return releaseFixture, nil
	}

	client, _ := NewClient(nil, f, nil)

	releases, err := client.GetList()

	assert.Equal(t, len(releases), 1)
	assert.Equal(t, releases[0].Name, "foo")
	assert.Nil(t, err)
}

func TestGetRelease(t *testing.T) {
	var f GetReleaseFunc = func(name string) (*release.Release, error) {
		releaseFixture := &release.Release{

			Name:      "foo",
			Version:   1,
			Namespace: "bigbang",
		}
		return releaseFixture, nil
	}

	client, _ := NewClient(f, nil, nil)

	release, err := client.GetRelease("foo")

	assert.Equal(t, release.Name, "foo")
	assert.Nil(t, err)
}

func TestGetValues(t *testing.T) {
	var f GetValuesFunc = func(name string) (map[string]interface{}, error) {
		v := map[string]interface{}{"kind": "foo"}
		return v, nil
	}

	client, _ := NewClient(nil, nil, f)

	value, err := client.GetValues("foo")

	assert.Equal(t, value, map[string]interface{}{"kind": "foo"})
	assert.Nil(t, err)
}
