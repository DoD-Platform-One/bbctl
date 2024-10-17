package pool

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	bbLog "repo1.dso.mil/big-bang/product/packages/bbctl/util/log"
)

func TestLoggerClientPoolContains(t *testing.T) {
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
			logger := &slog.Logger{}
			clientGetter := bbLog.ClientGetter{}
			client := clientGetter.GetClient(logger)
			instance := &loggerClientInstance{
				logger: logger,
				client: client,
			}
			badInstance1 := &loggerClientInstance{
				logger: &slog.Logger{},
				client: clientGetter.GetClient(&slog.Logger{}),
			}
			pool := loggerClientPool{}
			if !tc.empty {
				pool = append(pool, badInstance1)
				if tc.found {
					pool = append(pool, instance)
				}
			}
			// act
			found, result := pool.contains(logger)
			// assert
			assert.Equal(t, tc.found, found)
			if tc.found {
				assert.Equal(t, client, result)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

func TestLoggerClientPoolAdd(t *testing.T) {
	// arrange
	logger := &slog.Logger{}
	clientGetter := bbLog.ClientGetter{}
	client := clientGetter.GetClient(logger)
	instance := &loggerClientInstance{
		logger: logger,
		client: client,
	}
	pool := loggerClientPool{}
	// act
	pool.add(client, logger)
	// assert
	assert.Len(t, pool, 1)
	assert.Equal(t, instance, pool[0])
}
