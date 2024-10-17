package helm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	assert.Len(t, releases, 1)
	assert.Equal(t, "foo", releases[0].Name)
	require.NoError(t, err)
}

func TestGetRelease(t *testing.T) {
	var f GetReleaseFunc = func(_ string) (*release.Release, error) {
		releaseFixture := &release.Release{

			Name:      "foo",
			Version:   1,
			Namespace: "bigbang",
		}
		return releaseFixture, nil
	}

	client, _ := NewClient(f, nil, nil)

	release, err := client.GetRelease("foo")

	assert.Equal(t, "foo", release.Name)
	require.NoError(t, err)
}

func TestGetValues(t *testing.T) {
	var f GetValuesFunc = func(_ string) (map[string]interface{}, error) {
		v := map[string]interface{}{"kind": "foo"}
		return v, nil
	}

	client, _ := NewClient(nil, nil, f)

	value, err := client.GetValues("foo")

	assert.Equal(t, map[string]interface{}{"kind": "foo"}, value)
	require.NoError(t, err)
}
