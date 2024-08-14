package log

import (
	"context"
	"fmt"
	"log/slog"

	bbUtilLog "repo1.dso.mil/big-bang/product/packages/bbctl/util/log"
)

// LoggingFunction type is a type hint representing an arbitrary function that accepts a collection of string parameters
type LoggingFunction func(...string)

// NewFakeClient initializes and returns a new Fake Log client with the provided logging function
func NewFakeClient(logFunction LoggingFunction) bbUtilLog.Client {
	return &fakeClient{logFunction: logFunction}
}

// fakeClient
type fakeClient struct {
	logFunction LoggingFunction
}

// Fake logger client functions provided to conform to the Logger interface

// CloneWithUpdates not implemented
//
// Panics when called
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

// Logger not implemented
//
// Panics when called
func (c *fakeClient) Logger() *slog.Logger {
	panic("unimplemented")
}

// Debug implements log.Client.
func (c *fakeClient) Debug(format string, args ...interface{}) {
	msg := fmt.Sprintf("DEBUG: "+format, args...)
	c.logFunction(msg)
}

// DebugContext not implemented
//
// Panics when called
func (c *fakeClient) DebugContext(context context.Context, format string, args ...interface{}) {
	panic("unimplemented")
}

// Enabled not implemented
//
// Panics when called
func (c *fakeClient) Enabled(context context.Context, level slog.Level) bool {
	panic("unimplemented")
}

// Error formats the message using the provided format then calls the configured log function and panics
func (c *fakeClient) Error(format string, args ...interface{}) {
	msg := fmt.Sprintf("ERROR: "+format, args...)
	c.logFunction(msg)
	panic(msg)
}

// ErrorContext not implemented
//
// Panics when called
func (c *fakeClient) ErrorContext(context context.Context, format string, args ...interface{}) {
	panic("unimplemented")
}

// HandleError formats the message using the provided format then calls the configured log function and panics
//
// Format string must not be empty
func (c *fakeClient) HandleError(format string, err error, exitFunc bbUtilLog.ExitFunc, args ...interface{}) {
	if format == "" {
		msg := "format string cannot be empty for HandleError"
		c.logFunction(msg)
		panic(msg)
	}
	if err != nil {
		newArgs := append(args, err)
		msg := fmt.Sprintf("HANDLE_ERROR: "+format, newArgs...)
		c.logFunction(msg)
		exitFunc(1)
	}
}

// Handler not implemented
//
// Panics when called
func (c *fakeClient) Handler() slog.Handler {
	panic("unimplemented")
}

// Info formats the message using the provided format then calls the configured log function
func (c *fakeClient) Info(format string, args ...interface{}) {
	msg := fmt.Sprintf("INFO: "+format, args...)
	c.logFunction(msg)
}

// InfoContext not implemented
//
// Panics when called
func (c *fakeClient) InfoContext(context context.Context, format string, args ...interface{}) {
	panic("unimplemented")
}

// Log formats the message using the provided format then calls the configured log function
func (c *fakeClient) Log(context context.Context, level slog.Level, format string, args ...interface{}) {
	msg := fmt.Sprintf("LOG: "+format, args...)
	c.logFunction(msg)
}

// LogAttrs not implemented
//
// Panics when called
func (c *fakeClient) LogAttrs(context context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	panic("unimplemented")
}

// Warn formats the message using the provided format then calls the configured log function
func (c *fakeClient) Warn(format string, args ...interface{}) {
	msg := fmt.Sprintf("WARN: "+format, args...)
	c.logFunction(msg)
}

// WarnContext not implemented
//
// Panics when called
func (c *fakeClient) WarnContext(context context.Context, format string, args ...interface{}) {
	panic("unimplemented")
}

// With not implemented
//
// Panics when called
func (c *fakeClient) With(attrs ...interface{}) *bbUtilLog.Client {
	panic("unimplemented")
}

// WithGroup not implemented
//
// Panics when called
func (c *fakeClient) WithGroup(group string) *bbUtilLog.Client {
	panic("unimplemented")
}
