package pool

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"helm.sh/helm/v3/pkg/release"
	helm "repo1.dso.mil/big-bang/product/packages/bbctl/util/test/helm"
)

func TestHelmClientPoolContains(t *testing.T) {
	testCases := []struct {
		name  string
		found bool
		empty bool
	}{
		{
			name:  "found element",
			found: true,
			empty: false,
		},
		{
			name:  "not found element",
			found: false,
			empty: false,
		},
		{
			name:  "empty pool",
			found: false,
			empty: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// arrange
			client, err := helm.NewFakeClient(func(string) (*release.Release, error) { return nil, nil }, func() ([]*release.Release, error) { return nil, nil }, func(string) (map[string]interface{}, error) { return nil, nil }, nil)
			require.NoError(t, err)
			pool := helmClientPool{}
			if !tc.empty {
				pool = append(pool, &helmClientInstance{
					namespace: "not found",
					client:    client,
				})
				if tc.found {
					pool = append(pool, &helmClientInstance{
						namespace: "test",
						client:    client,
					})
				}
			}
			// act
			found, result := pool.contains("test")
			// assert
			if tc.found {
				assert.True(t, found)
				assert.Equal(t, client, result)
			} else {
				assert.False(t, found)
				assert.Nil(t, result)
			}
		})
	}
}

func TestHelmClientPoolAdd(t *testing.T) {
	// arrange
	client, err := helm.NewFakeClient(func(string) (*release.Release, error) { return nil, nil }, func() ([]*release.Release, error) { return nil, nil }, func(string) (map[string]interface{}, error) { return nil, nil }, nil)
	require.NoError(t, err)
	pool := helmClientPool{}
	// act
	pool.add(client, "test")
	// assert
	assert.Len(t, pool, 1)
	assert.Equal(t, client, pool[0].client)
	assert.Equal(t, "test", pool[0].namespace)
}
