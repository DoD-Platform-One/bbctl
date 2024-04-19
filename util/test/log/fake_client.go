package log

import (
	"context"
	"fmt"
	"log/slog"

	bbUtilLog "repo1.dso.mil/big-bang/product/packages/bbctl/util/log"
)

// LoggingFunction - log function
type LoggingFunction func(...string)

// NewFakeClient - returns a new Fake Log client with the provided options
func NewFakeClient(logFunction LoggingFunction) bbUtilLog.Client {
	return &fakeClient{logFunction: logFunction}
}

// fakeClient - fake client
type fakeClient struct {
	logFunction LoggingFunction
}

// CloneWithUpdates implements log.Client.
func (c *fakeClient) CloneWithUpdates(
	debugFunc bbUtilLog.DebugFunc,
	debugContextFunc bbUtilLog.DebugContextFunc,
	enabledFunc bbUtilLog.EnabledFunc,
	errorFunc bbUtilLog.ErrorFunc,
	errorContextFunc bbUtilLog.ErrorContextFunc,
	handleErrorFunc bbUtilLog.HandleErrorFunc,
	handlerFunc bbUtilLog.HandlerFunc,
	infoFunc bbUtilLog.InfoFunc,
	infoContextFunc bbUtilLog.InfoContextFunc,
	logFunc bbUtilLog.LogFunc,
	logAttrsFunc bbUtilLog.LogAttrsFunc,
	logger *slog.Logger,
	warnFunc bbUtilLog.WarnFunc,
	warnContextFunc bbUtilLog.WarnContextFunc,
	withFunc bbUtilLog.WithFunc,
	withGroupFunc bbUtilLog.WithGroupFunc,
) bbUtilLog.Client {
	panic("unimplemented")
}

// Logger implements log.Client.
func (c *fakeClient) Logger() *slog.Logger {
	panic("unimplemented")
}

// Debug implements log.Client.
func (c *fakeClient) Debug(format string, args ...interface{}) {
	msg := fmt.Sprintf("DEBUG: "+format, args...)
	c.logFunction(msg)
}

// DebugContext implements log.Client.
func (c *fakeClient) DebugContext(context context.Context, format string, args ...interface{}) {
	panic("unimplemented")
}

// Enabled implements log.Client.
func (c *fakeClient) Enabled(context context.Context, level slog.Level) bool {
	panic("unimplemented")
}

// Error - logs the error and panics
func (c *fakeClient) Error(format string, args ...interface{}) {
	msg := fmt.Sprintf("ERROR: "+format, args...)
	c.logFunction(msg)
	panic(msg)
}

// ErrorContext implements log.Client.
func (c *fakeClient) ErrorContext(context context.Context, format string, args ...interface{}) {
	panic("unimplemented")
}

// HandleError - logs the error and panics
func (c *fakeClient) HandleError(format string, err error, args ...interface{}) {
	if format == "" {
		msg := "format string cannot be empty for HandleError"
		c.logFunction(msg)
		panic(msg)
	}
	if err != nil {
		newArgs := append(args, err)
		msg := fmt.Sprintf("HANDLE_ERROR: "+format, newArgs...)
		c.logFunction(msg)
		panic(msg)
	}
}

// Handler implements log.Client.
func (c *fakeClient) Handler() slog.Handler {
	panic("unimplemented")
}

// Info implements log.Client.
func (c *fakeClient) Info(format string, args ...interface{}) {
	msg := fmt.Sprintf("INFO: "+format, args...)
	c.logFunction(msg)
}

// InfoContext implements log.Client.
func (c *fakeClient) InfoContext(context context.Context, format string, args ...interface{}) {
	panic("unimplemented")
}

// Log implements log.Client.
func (c *fakeClient) Log(context context.Context, level slog.Level, format string, args ...interface{}) {
	msg := fmt.Sprintf("LOG: "+format, args...)
	c.logFunction(msg)
}

// LogAttrs implements log.Client.
func (c *fakeClient) LogAttrs(context context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	panic("unimplemented")
}

// Warn implements log.Client.
func (c *fakeClient) Warn(format string, args ...interface{}) {
	msg := fmt.Sprintf("WARN: "+format, args...)
	c.logFunction(msg)
}

// WarnContext implements log.Client.
func (c *fakeClient) WarnContext(context context.Context, format string, args ...interface{}) {
	panic("unimplemented")
}

// With implements log.Client.
func (c *fakeClient) With(attrs ...interface{}) *bbUtilLog.Client {
	panic("unimplemented")
}

// WithGroup implements log.Client.
func (c *fakeClient) WithGroup(group string) *bbUtilLog.Client {
	panic("unimplemented")
}
