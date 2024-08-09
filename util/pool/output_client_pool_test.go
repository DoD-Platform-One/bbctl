package pool

import (
	"testing"

	"github.com/stretchr/testify/assert"
	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"
	bbOutput "repo1.dso.mil/big-bang/product/packages/bbctl/util/output"
)

func TestOutputClientPoolContains(t *testing.T) {
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
			streams, in, out, errOut := genericIOOptions.NewTestIOStreams()
			badStreams, badIn, badOut, badErrOut := genericIOOptions.NewTestIOStreams()
			clientGetter := bbOutput.ClientGetter{}
			outputClient := clientGetter.GetClient("", streams)
			goodInstance := &outputClientInstance{
				client:  outputClient,
				streams: streams,
			}
			badInstance1 := &outputClientInstance{
				client:  outputClient,
				streams: badStreams,
			}
			pool := outputClientPool{}
			if !tc.empty {
				pool = append(pool, badInstance1)
				if tc.found {
					pool = append(pool, goodInstance)
				}
			}
			// act
			found, result := pool.contains(streams)
			// assert
			assert.Equal(t, tc.found, found)
			assert.Empty(t, in.String())
			assert.Empty(t, out.String())
			assert.Empty(t, errOut.String())
			assert.Empty(t, badIn.String())
			assert.Empty(t, badOut.String())
			assert.Empty(t, badErrOut.String())
			if tc.found {
				assert.Equal(t, outputClient, result)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

func TestOutputClientPoolAdd(t *testing.T) {
	// arrange
	streams, in, out, errOut := genericIOOptions.NewTestIOStreams()
	clientGetter := bbOutput.ClientGetter{}
	outputClient := clientGetter.GetClient("", streams)
	pool := outputClientPool{}
	// act
	pool.add(outputClient, streams)
	// assert
	assert.NotEmpty(t, pool)
	assert.Empty(t, in.String())
	assert.Empty(t, out.String())
	assert.Empty(t, errOut.String())
	assert.Equal(t, 1, len(pool))
	assert.Equal(t, outputClient, pool[0].client)
	assert.Equal(t, streams, pool[0].streams)
}
