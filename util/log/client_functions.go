package log

import (
	"context"
	"fmt"
	"log/slog"
)

func debug(clientLogger Client, format string, args ...interface{}) {
	clientLogger.Logger().Debug(format, args...)
}

func debugContext(context context.Context, clientLogger Client, format string, args ...interface{}) {
	clientLogger.Logger().DebugContext(context, format, args...)
}

func enabled(context context.Context, clientLogger Client, level slog.Level) bool {
	return clientLogger.Logger().Enabled(context, level)
}

func errorOut(clientLogger Client, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	clientLogger.Logger().Error(msg)
}

func errorContext(context context.Context, clientLogger Client, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	clientLogger.Logger().ErrorContext(context, msg)
}

func handlerFunc(clientLogger Client) slog.Handler {
	return clientLogger.Logger().Handler()
}

func info(clientLogger Client, format string, args ...interface{}) {
	clientLogger.Logger().Info(format, args...)
}

func infoContext(context context.Context, clientLogger Client, format string, args ...interface{}) {
	clientLogger.Logger().InfoContext(context, format, args...)
}

func log(context context.Context, clientLogger Client, level slog.Level, msg string, args ...interface{}) {
	clientLogger.Logger().Log(context, level, msg, args...)
}

func logAttrs(context context.Context, clientLogger Client, level slog.Level, msg string, attrs ...slog.Attr) {
	clientLogger.Logger().LogAttrs(context, level, msg, attrs...)
}

func warn(clientLogger Client, format string, args ...interface{}) {
	clientLogger.Logger().Warn(format, args...)
}

func warnContext(context context.Context, clientLogger Client, format string, args ...interface{}) {
	clientLogger.Logger().WarnContext(context, format, args...)
}

func with(clientLogger Client, args ...interface{}) *Client {
	newLogger := clientLogger.Logger().With(args...)
	newClientLogger := clientLogger.CloneWithUpdates(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, newLogger, nil, nil, nil, nil)
	return &newClientLogger
}

func withGroup(clientLogger Client, group string) *Client {
	newLogger := clientLogger.Logger().WithGroup(group)
	newClientLogger := clientLogger.CloneWithUpdates(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, newLogger, nil, nil, nil, nil)
	return &newClientLogger
}
