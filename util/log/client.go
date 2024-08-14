package log

import (
	"context"
	"log/slog"

	"github.com/duaneking/coalesce"
)

// Client holds the functions to handle logging
type Client interface {
	// CloneWithUpdates - returns a new client with the given attributes in each output
	CloneWithUpdates(
		debugFunc DebugFunc,
		debugContextFunc DebugContextFunc,
		enabledFunc EnabledFunc,
		errorFunc ErrorFunc,
		errorContextFunc ErrorContextFunc,
		handleErrorFunc HandleErrorFunc,
		handlerFunc HandlerFunc,
		infoFunc InfoFunc,
		infoContextFunc InfoContextFunc,
		logFunc LogFunc,
		logAttrsFunc LogAttrsFunc,
		logger *slog.Logger,
		warnFunc WarnFunc,
		warnContextFunc WarnContextFunc,
		withFunc WithFunc,
		withGroupFunc WithGroupFunc,
	) Client
	// Debug - log a debug message
	Debug(format string, args ...interface{})
	// DebugContext - log a debug message with context
	DebugContext(context context.Context, format string, args ...interface{})
	// Enabled - check if logs will emit with given context and level
	Enabled(context context.Context, level slog.Level) bool
	// Error - log an error and panic
	Error(format string, args ...interface{})
	// ErrorContext - log an error with context and panic
	ErrorContext(context context.Context, format string, args ...interface{})
	// HandleError - check for an error, log and cleanly exit if there is one, else return. err gets appended to args before formatting
	HandleError(format string, err error, exitFunc ExitFunc, args ...interface{})
	// Handler - return a slog.Handler in use by the client
	Handler() slog.Handler
	// Info - log an info message
	Info(format string, args ...interface{})
	// InfoContext - log an info message with context
	InfoContext(context context.Context, format string, args ...interface{})
	// Log - log a message
	Log(context context.Context, level slog.Level, format string, args ...interface{})
	// LogAttrs - log a message with attributes
	LogAttrs(context context.Context, level slog.Level, msg string, attrs ...slog.Attr)
	// Logger - return the logger in use by the client
	Logger() *slog.Logger
	// Warn - log a warning message
	Warn(format string, args ...interface{})
	// WarnContext - log a warning message with context
	WarnContext(context context.Context, format string, args ...interface{})
	// With - returns a new client with the given attributes in each output
	With(args ...interface{}) *Client
	// WithGroup - returns a new client with the group for each output (affects all other attrs given the handler's processing)
	WithGroup(group string) *Client
}

// DebugFunc type
type DebugFunc func(Client, string, ...interface{})

// DebugContextFunc type
type DebugContextFunc func(context.Context, Client, string, ...interface{})

// EnabledFunc type
type EnabledFunc func(context.Context, Client, slog.Level) bool

// ErrorFunc type
type ErrorFunc func(Client, string, ...interface{})

// ErrorContextFunc type
type ErrorContextFunc func(context.Context, Client, string, ...interface{})

// exitFunc is a function that exits the program with the given exit code
type ExitFunc func(int)

// HandleErrorFunc type
type HandleErrorFunc func(Client, string, error, ExitFunc, ...interface{})

// HandlerFunc type
type HandlerFunc func(Client) slog.Handler

// InfoFunc type
type InfoFunc func(Client, string, ...interface{})

// InfoContextFunc type
type InfoContextFunc func(context.Context, Client, string, ...interface{})

// LogFunc type // golangci-lint complains about stuttering, but this is the best name for this type
type LogFunc func(context.Context, Client, slog.Level, string, ...interface{})

// LogAttrsFunc type // golangci-lint complains about stuttering, but this is the best name for this type
type LogAttrsFunc func(context.Context, Client, slog.Level, string, ...slog.Attr)

// WarnFunc type
type WarnFunc func(Client, string, ...interface{})

// WarnContextFunc type
type WarnContextFunc func(context.Context, Client, string, ...interface{})

// WithFunc type
type WithFunc func(Client, ...interface{}) *Client

// WithGroupFunc type
type WithGroupFunc func(Client, string) *Client

// loggingClient is composed of functions to handle logging
type loggingClient struct {
	debugFunc        DebugFunc
	debugContextFunc DebugContextFunc
	enabledFunc      EnabledFunc
	errorFunc        ErrorFunc
	errorContextFunc ErrorContextFunc
	handleErrorFunc  HandleErrorFunc
	handlerFunc      HandlerFunc
	infoFunc         InfoFunc
	infoContextFunc  InfoContextFunc
	logFunc          LogFunc
	logAttrsFunc     LogAttrsFunc
	logger           *slog.Logger
	warnFunc         WarnFunc
	warnContextFunc  WarnContextFunc
	withFunc         WithFunc
	withGroupFunc    WithGroupFunc
}

// NewClient returns a new Logging client with the provided configuration
func NewClient(
	debugFunc DebugFunc,
	debugContextFunc DebugContextFunc,
	enabledFunc EnabledFunc,
	errorFunc ErrorFunc,
	errorContextFunc ErrorContextFunc,
	handleErrorFunc HandleErrorFunc,
	handlerFunc HandlerFunc,
	infoFunc InfoFunc,
	infoContextFunc InfoContextFunc,
	logFunc LogFunc,
	logAttrsFunc LogAttrsFunc,
	logger *slog.Logger,
	warnFunc WarnFunc,
	warnContextFunc WarnContextFunc,
	withFunc WithFunc,
	withGroupFunc WithGroupFunc,
) Client {
	loggerToUse := logger
	if loggerToUse == nil {
		loggerToUse = slog.Default()
	}
	return &loggingClient{
		debugFunc:        debugFunc,
		debugContextFunc: debugContextFunc,
		enabledFunc:      enabledFunc,
		errorFunc:        errorFunc,
		errorContextFunc: errorContextFunc,
		handleErrorFunc:  handleErrorFunc,
		handlerFunc:      handlerFunc,
		infoFunc:         infoFunc,
		infoContextFunc:  infoContextFunc,
		logFunc:          logFunc,
		logAttrsFunc:     logAttrsFunc,
		logger:           loggerToUse,
		warnFunc:         warnFunc,
		warnContextFunc:  warnContextFunc,
		withFunc:         withFunc,
		withGroupFunc:    withGroupFunc,
	}
}

// CloneWithUpdates - returns a new client with the given attributes in each output
func (c *loggingClient) CloneWithUpdates(
	debugFunc DebugFunc,
	debugContextFunc DebugContextFunc,
	enabledFunc EnabledFunc,
	errorFunc ErrorFunc,
	errorContextFunc ErrorContextFunc,
	handleErrorFunc HandleErrorFunc,
	handlerFunc HandlerFunc,
	infoFunc InfoFunc,
	infoContextFunc InfoContextFunc,
	logFunc LogFunc,
	logAttrsFunc LogAttrsFunc,
	logger *slog.Logger,
	warnFunc WarnFunc,
	warnContextFunc WarnContextFunc,
	withFunc WithFunc,
	withGroupFunc WithGroupFunc,
) Client {
	debugFuncToUse := coalesce.Coalesce(&debugFunc, &c.debugFunc)
	debugContextFuncToUse := coalesce.Coalesce(&debugContextFunc, &c.debugContextFunc)
	enabledFuncToUse := coalesce.Coalesce(&enabledFunc, &c.enabledFunc)
	errorFuncToUse := coalesce.Coalesce(&errorFunc, &c.errorFunc)
	errorContextFuncToUse := coalesce.Coalesce(&errorContextFunc, &c.errorContextFunc)
	handleErrorFuncToUse := coalesce.Coalesce(&handleErrorFunc, &c.handleErrorFunc)
	handlerFuncToUse := coalesce.Coalesce(&handlerFunc, &c.handlerFunc)
	infoFuncToUse := coalesce.Coalesce(&infoFunc, &c.infoFunc)
	infoContextFuncToUse := coalesce.Coalesce(&infoContextFunc, &c.infoContextFunc)
	logFuncToUse := coalesce.Coalesce(&logFunc, &c.logFunc)
	logAttrsFuncToUse := coalesce.Coalesce(&logAttrsFunc, &c.logAttrsFunc)
	loggerToUse := coalesce.Coalesce(&logger, &c.logger)
	warnFuncToUse := coalesce.Coalesce(&warnFunc, &c.warnFunc)
	warnContextFuncToUse := coalesce.Coalesce(&warnContextFunc, &c.warnContextFunc)
	withFuncToUse := coalesce.Coalesce(&withFunc, &c.withFunc)
	withGroupFuncToUse := coalesce.Coalesce(&withGroupFunc, &c.withGroupFunc)
	return NewClient(
		*debugFuncToUse,
		*debugContextFuncToUse,
		*enabledFuncToUse,
		*errorFuncToUse,
		*errorContextFuncToUse,
		*handleErrorFuncToUse,
		*handlerFuncToUse,
		*infoFuncToUse,
		*infoContextFuncToUse,
		*logFuncToUse,
		*logAttrsFuncToUse,
		*loggerToUse,
		*warnFuncToUse,
		*warnContextFuncToUse,
		*withFuncToUse,
		*withGroupFuncToUse,
	)
}

// Debug - log a debug message
func (c *loggingClient) Debug(format string, args ...interface{}) {
	c.debugFunc(c, format, args...)
}

// DebugContext - log a debug message with context
func (c *loggingClient) DebugContext(context context.Context, format string, args ...interface{}) {
	c.debugContextFunc(context, c, format, args...)
}

// Enabled - check if logs will emit with given context and level
func (c *loggingClient) Enabled(context context.Context, level slog.Level) bool {
	return c.enabledFunc(context, c, level)
}

// Error - log an error
func (c *loggingClient) Error(format string, args ...interface{}) {
	c.errorFunc(c, format)
}

// ErrorContext - log an error with context and panic
func (c *loggingClient) ErrorContext(context context.Context, format string, args ...interface{}) {
	c.errorContextFunc(context, c, format, args...)
}

// HandleError - handle an error, execute the exitFunc with the given exit code if present
func (c *loggingClient) HandleError(format string, err error, exitFunc ExitFunc, args ...interface{}) {
	c.handleErrorFunc(c, format, err, exitFunc, args...)
}

// Handler - return a slog.Handler in use by the client
func (c *loggingClient) Handler() slog.Handler {
	return c.handlerFunc(c)
}

// Info - log an info message
func (c *loggingClient) Info(format string, args ...interface{}) {
	c.infoFunc(c, format, args...)
}

// InfoContext - log an info message with context
func (c *loggingClient) InfoContext(context context.Context, format string, args ...interface{}) {
	c.infoContextFunc(context, c, format, args...)
}

// Log - log a message
func (c *loggingClient) Log(context context.Context, level slog.Level, format string, args ...interface{}) {
	c.logFunc(context, c, level, format, args...)
}

// LogAttrs - log a message with attributes
func (c *loggingClient) LogAttrs(context context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	c.logAttrsFunc(context, c, level, msg, attrs...)
}

// Logger - return the logger in use by the client
func (c *loggingClient) Logger() *slog.Logger {
	return c.logger
}

// Warn - log a warning message
func (c *loggingClient) Warn(format string, args ...interface{}) {
	c.warnFunc(c, format, args...)
}

// WarnContext - log a warning message with context
func (c *loggingClient) WarnContext(context context.Context, format string, args ...interface{}) {
	c.warnContextFunc(context, c, format, args...)
}

// With - returns a new client with the given attributes in each output
func (c *loggingClient) With(args ...interface{}) *Client {
	return c.withFunc(c, args...)
}

// WithGroup - returns a new client with the group for each output (affects all other attrs given the handler's processing)
func (c *loggingClient) WithGroup(group string) *Client {
	return c.withGroupFunc(c, group)
}
