package pool

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	fakeRuntimeClient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestRuntimeClientPoolContains(t *testing.T) {
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
			goodScheme := runtime.NewScheme()
			assert.Nil(t, goodScheme.SetVersionPriority(schema.GroupVersion{Group: "test", Version: "v1"}))
			goodClientBuilder := &fakeRuntimeClient.ClientBuilder{}
			goodClientBuilder.WithScheme(goodScheme)
			goodClient := goodClientBuilder.Build()
			goodInstance := &runtimeClientInstance{
				scheme: goodScheme,
				client: goodClient,
			}
			badScheme1 := runtime.NewScheme()
			assert.Nil(t, goodScheme.SetVersionPriority(schema.GroupVersion{Group: "test", Version: "v0"}))
			badClientBuilder1 := &fakeRuntimeClient.ClientBuilder{}
			badClientBuilder1.WithScheme(badScheme1)
			badClient1 := badClientBuilder1.Build()
			badInstance1 := &runtimeClientInstance{
				scheme: badScheme1,
				client: badClient1,
			}
			pool := runtimeClientPool{}
			if !tc.empty {
				pool = append(pool, badInstance1)
				if tc.found {
					pool = append(pool, goodInstance)
				}
			}
			// act
			found, result := pool.contains(goodScheme)
			// assert
			assert.Equal(t, tc.found, found)
			if tc.found {
				assert.Equal(t, goodClient, result)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

func TestRuntimeClientPoolAdd(t *testing.T) {
	// arrange
	goodScheme := runtime.NewScheme()
	assert.Nil(t, goodScheme.SetVersionPriority(schema.GroupVersion{Group: "test", Version: "v1"}))
	goodClientBuilder := &fakeRuntimeClient.ClientBuilder{}
	goodClientBuilder.WithScheme(goodScheme)
	goodClient := goodClientBuilder.Build()
	pool := runtimeClientPool{}
	// act
	pool.add(goodClient, goodScheme)
	// assert
	assert.Len(t, pool, 1)
	assert.Equal(t, goodScheme, pool[0].scheme)
	assert.Equal(t, goodClient, pool[0].client)
}
