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
	_ bbUtilLog.DebugFunc,
	_ bbUtilLog.DebugContextFunc,
	_ bbUtilLog.EnabledFunc,
	_ bbUtilLog.ErrorFunc,
	_ bbUtilLog.ErrorContextFunc,
	_ bbUtilLog.HandlerFunc,
	_ bbUtilLog.InfoFunc,
	_ bbUtilLog.InfoContextFunc,
	_ bbUtilLog.LogFunc,
	_ bbUtilLog.LogAttrsFunc,
	_ *slog.Logger,
	_ bbUtilLog.WarnFunc,
	_ bbUtilLog.WarnContextFunc,
	_ bbUtilLog.WithFunc,
	_ bbUtilLog.WithGroupFunc,
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
func (c *fakeClient) DebugContext(_ context.Context, _ string, _ ...interface{}) {
	panic("unimplemented")
}

// Enabled not implemented
//
// Panics when called
func (c *fakeClient) Enabled(_ context.Context, _ slog.Level) bool {
	panic("unimplemented")
}

// Error formats the message using the provided format then calls the configured log function and panics
func (c *fakeClient) Error(format string, args ...interface{}) {
	msg := fmt.Sprintf("ERROR: "+format, args...)
	c.logFunction(msg)
}

// ErrorContext not implemented
//
// Panics when called
func (c *fakeClient) ErrorContext(_ context.Context, _ string, _ ...interface{}) {
	panic("unimplemented")
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
func (c *fakeClient) InfoContext(_ context.Context, _ string, _ ...interface{}) {
	panic("unimplemented")
}

// Log formats the message using the provided format then calls the configured log function
func (c *fakeClient) Log(_ context.Context, _ slog.Level, format string, args ...interface{}) {
	msg := fmt.Sprintf("LOG: "+format, args...)
	c.logFunction(msg)
}

// LogAttrs not implemented
//
// Panics when called
func (c *fakeClient) LogAttrs(_ context.Context, _ slog.Level, _ string, _ ...slog.Attr) {
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
func (c *fakeClient) WarnContext(_ context.Context, _ string, _ ...interface{}) {
	panic("unimplemented")
}

// With not implemented
//
// Panics when called
func (c *fakeClient) With(_ ...interface{}) *bbUtilLog.Client {
	panic("unimplemented")
}

// WithGroup not implemented
//
// Panics when called
func (c *fakeClient) WithGroup(_ string) *bbUtilLog.Client {
	panic("unimplemented")
}
