package log

import "log/slog"

// ClientGetter is an interface for getting an BB log client.
type ClientGetter struct{}

// GetClient returns a new log client.
func (clientGetter *ClientGetter) GetClient(logger *slog.Logger) Client {
	return NewClient(
		debug,
		debugContext,
		enabled,
		errorOut,
		errorContext,
		handleError,
		handlerFunc,
		info,
		infoContext,
		log,
		logAttrs,
		logger,
		warn,
		warnContext,
		with,
		withGroup,
	)
}
