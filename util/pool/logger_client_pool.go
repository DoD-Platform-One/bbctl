package pool

import (
	"log/slog"

	bbLog "repo1.dso.mil/big-bang/product/packages/bbctl/util/log"
)

// loggerClientInstance is a struct that holds a logging client and the slog.Logger it is configured for
type loggerClientInstance struct {
	logger *slog.Logger
	client bbLog.Client
}

// loggerClientPool is a slice of loggerClientInstance structs
type loggerClientPool []*loggerClientInstance

// contains checks if a loggerClientPool contains a logging client for a given slog.Logger
func (l loggerClientPool) contains(logger *slog.Logger) (bool, bbLog.Client) {
	for _, client := range l {
		if client.logger == logger {
			return true, client.client
		}
	}
	return false, nil
}

// add adds a logging client to the loggerClientPool
func (l *loggerClientPool) add(client bbLog.Client, logger *slog.Logger) {
	*l = append(*l, &loggerClientInstance{
		logger: logger,
		client: client,
	})
}
