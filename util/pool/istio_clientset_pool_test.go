package pool

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/rest"
	bbUtilTestApiWrappers "repo1.dso.mil/big-bang/product/packages/bbctl/util/test/apiwrappers"
)

func TestIstioClientsetPoolContains(t *testing.T) {
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
			restConfig := &rest.Config{
				Host: "goodTest",
			}
			clientset := &bbUtilTestApiWrappers.FakeIstioClientSet{}
			goodInstance := &istioClientsetInstance{
				clientset:  clientset,
				restConfig: restConfig,
			}
			badInstance1 := &istioClientsetInstance{
				clientset: clientset,
				restConfig: &rest.Config{
					Host: "badTest1",
				},
			}
			pool := istioClientsetPool{}
			if !tc.empty {
				pool = append(pool, badInstance1)
				if tc.found {
					pool = append(pool, goodInstance)
				}
			}
			// act
			found, result := pool.contains(restConfig)
			// assert
			assert.Equal(t, tc.found, found)
			if tc.found {
				assert.Equal(t, clientset, result)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

func TestIstioClientsetPoolAdd(t *testing.T) {
	// arrange
	restConfig := &rest.Config{
		Host: "goodTest",
	}
	clientset := &bbUtilTestApiWrappers.FakeIstioClientSet{}
	pool := istioClientsetPool{}
	// act
	pool.add(clientset, restConfig)
	// assert
	assert.Len(t, pool, 1)
	assert.Equal(t, clientset, pool[0].clientset)
	assert.Equal(t, restConfig, pool[0].restConfig)
}
