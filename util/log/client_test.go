package log

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewClientNilLoggerPass(t *testing.T) {
	// Arrange
	client := NewClient(
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
		nil,
		warn,
		warnContext,
		with,
		withGroup,
	)
	// Act
	// Assert
	assert.NotNil(t, client)
	assert.NotNil(t, client.Logger())
	assert.Equal(t, slog.Default(), client.Logger())
}

func TestClientDebugPass(t *testing.T) {
	// Arrange
	startTime := time.Now().UTC().Local()
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelDebug.Level())
	client := *createTestClient(t, &stringBuilder, leveler)
	// Act
	client.Debug("test")
	// Assert
	jsonResult := stringBuilder.String()
	var jsonObject TestLog
	err := json.Unmarshal([]byte(jsonResult), &jsonObject)
	assert.Nil(t, err)
	logTime, err := time.Parse(time.RFC3339, jsonObject.Time)
	assert.Nil(t, err)
	assert.True(t, logTime.After(startTime))
	assert.Equal(t, "DEBUG", jsonObject.Level)
	assert.Contains(t, jsonObject.Source.Function, "util/log.debug")
	assert.Contains(t, jsonObject.Source.File, "util/log/client_functions.go")
	assert.Greater(t, jsonObject.Source.Line, 0)
	assert.Equal(t, "test", jsonObject.Message)
}

func TestClientDebugFail(t *testing.T) {
	// Arrange
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelInfo.Level())
	client := *createTestClient(t, &stringBuilder, leveler)
	// Act
	client.Debug("test")
	// Assert
	jsonResult := stringBuilder.String()
	assert.Empty(t, jsonResult)
}

func TestClientDebugContextPass(t *testing.T) {
	// Arrange
	startTime := time.Now().UTC().Local()
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelDebug.Level())
	client := *createTestClient(t, &stringBuilder, leveler)
	context := context.TODO()
	// Act
	client.DebugContext(context, "test")
	// Assert
	jsonResult := stringBuilder.String()
	var jsonObject TestLog
	err := json.Unmarshal([]byte(jsonResult), &jsonObject)
	assert.Nil(t, err)
	logTime, err := time.Parse(time.RFC3339, jsonObject.Time)
	assert.Nil(t, err)
	assert.True(t, logTime.After(startTime))
	assert.Equal(t, "DEBUG", jsonObject.Level)
	assert.Contains(t, jsonObject.Source.Function, "util/log.debugContext")
	assert.Contains(t, jsonObject.Source.File, "util/log/client_functions.go")
	assert.Greater(t, jsonObject.Source.Line, 0)
	assert.Equal(t, "test", jsonObject.Message)
}

func TestClientDebugContextFail(t *testing.T) {
	// Arrange
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelInfo.Level())
	client := *createTestClient(t, &stringBuilder, leveler)
	context := context.TODO()
	// Act
	client.DebugContext(context, "test")
	// Assert
	jsonResult := stringBuilder.String()
	assert.Empty(t, jsonResult)
}

func TestClientEnabledPass(t *testing.T) {
	// Arrange
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelDebug.Level())
	client := *createTestClient(t, &stringBuilder, leveler)
	context := context.TODO()
	// Act
	result := client.Enabled(context, slog.LevelDebug)
	// Assert
	assert.True(t, result)
}

func TestClientEnabledFail(t *testing.T) {
	// Arrange
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelInfo.Level())
	client := *createTestClient(t, &stringBuilder, leveler)
	context := context.TODO()
	// Act
	result := client.Enabled(context, slog.LevelDebug)
	// Assert
	assert.False(t, result)
}

func TestClientErrorPass(t *testing.T) {
	// Arrange
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelError.Level())
	client := *createTestClient(t, &stringBuilder, leveler)
	// Act
	assert.PanicsWithValue(t, "test", func() {
		client.Error("test")
	})
	// Assert
	jsonResult := stringBuilder.String()
	var jsonObject TestLog
	err := json.Unmarshal([]byte(jsonResult), &jsonObject)
	assert.Nil(t, err)
	assert.Equal(t, "ERROR", jsonObject.Level)
	assert.Contains(t, jsonObject.Source.Function, "util/log.errorOut")
	assert.Contains(t, jsonObject.Source.File, "util/log/client_functions.go")
	assert.Greater(t, jsonObject.Source.Line, 0)
	assert.Equal(t, "test", jsonObject.Message)
}

func TestClientErrorContextPass(t *testing.T) {
	// Arrange
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelError.Level())
	client := *createTestClient(t, &stringBuilder, leveler)
	context := context.TODO()
	// Act
	assert.PanicsWithValue(t, "test", func() {
		client.ErrorContext(context, "test")
	})
	// Assert
	jsonResult := stringBuilder.String()
	var jsonObject TestLog
	err := json.Unmarshal([]byte(jsonResult), &jsonObject)
	assert.Nil(t, err)
	assert.Equal(t, "ERROR", jsonObject.Level)
	assert.Contains(t, jsonObject.Source.Function, "util/log.errorContext")
	assert.Contains(t, jsonObject.Source.File, "util/log/client_functions.go")
	assert.Greater(t, jsonObject.Source.Line, 0)
	assert.Equal(t, "test", jsonObject.Message)
}

func TestClientHandleErrorPass(t *testing.T) {
	// Arrange
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelError.Level())
	client := *createTestClient(t, &stringBuilder, leveler)
	err := errors.New("test error")
	// Act
	assert.PanicsWithValue(t, "test: test error", func() {
		client.HandleError("test: %v", err)
	})
	// Assert
	jsonResult := stringBuilder.String()
	var jsonObject TestLog
	err = json.Unmarshal([]byte(jsonResult), &jsonObject)
	assert.Nil(t, err)
	assert.Equal(t, "ERROR", jsonObject.Level)
	assert.Contains(t, jsonObject.Source.Function, "util/log.handleError")
	assert.Contains(t, jsonObject.Source.File, "util/log/client_functions.go")
	assert.Greater(t, jsonObject.Source.Line, 0)
	assert.Equal(t, "test: test error", jsonObject.Message)
}

func TestClientHandleErrorNil(t *testing.T) {
	// Arrange
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelError.Level())
	client := *createTestClient(t, &stringBuilder, leveler)
	// Act
	client.HandleError("test: %v", nil)
	// Assert
	jsonResult := stringBuilder.String()
	assert.Empty(t, jsonResult)
}

func TestClientHandlerPass(t *testing.T) {
	// Arrange
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelDebug.Level())
	client := *createTestClient(t, &stringBuilder, leveler)
	// Act
	handler := client.Handler()
	// Assert
	assert.NotNil(t, handler)
}

func TestClientInfoPass(t *testing.T) {
	// Arrange
	startTime := time.Now().UTC().Local()
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelInfo.Level())
	client := *createTestClient(t, &stringBuilder, leveler)
	// Act
	client.Info("test")
	// Assert
	jsonResult := stringBuilder.String()
	var jsonObject TestLog
	err := json.Unmarshal([]byte(jsonResult), &jsonObject)
	assert.Nil(t, err)
	logTime, err := time.Parse(time.RFC3339, jsonObject.Time)
	assert.Nil(t, err)
	assert.True(t, logTime.After(startTime))
	assert.Equal(t, "INFO", jsonObject.Level)
	assert.Contains(t, jsonObject.Source.Function, "util/log.info")
	assert.Contains(t, jsonObject.Source.File, "util/log/client_functions.go")
	assert.Greater(t, jsonObject.Source.Line, 0)
	assert.Equal(t, "test", jsonObject.Message)
}

func TestClientInfoFail(t *testing.T) {
	// Arrange
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelWarn.Level())
	client := *createTestClient(t, &stringBuilder, leveler)
	// Act
	client.Info("test")
	// Assert
	jsonResult := stringBuilder.String()
	assert.Empty(t, jsonResult)
}

func TestClientInfoContextPass(t *testing.T) {
	// Arrange
	startTime := time.Now().UTC().Local()
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelInfo.Level())
	client := *createTestClient(t, &stringBuilder, leveler)
	context := context.TODO()
	// Act
	client.InfoContext(context, "test")
	// Assert
	jsonResult := stringBuilder.String()
	var jsonObject TestLog
	err := json.Unmarshal([]byte(jsonResult), &jsonObject)
	assert.Nil(t, err)
	logTime, err := time.Parse(time.RFC3339, jsonObject.Time)
	assert.Nil(t, err)
	assert.True(t, logTime.After(startTime))
	assert.Equal(t, "INFO", jsonObject.Level)
	assert.Contains(t, jsonObject.Source.Function, "util/log.infoContext")
	assert.Contains(t, jsonObject.Source.File, "util/log/client_functions.go")
	assert.Greater(t, jsonObject.Source.Line, 0)
	assert.Equal(t, "test", jsonObject.Message)
}

func TestClientInfoContextFail(t *testing.T) {
	// Arrange
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelWarn.Level())
	client := *createTestClient(t, &stringBuilder, leveler)
	context := context.TODO()
	// Act
	client.InfoContext(context, "test")
	// Assert
	jsonResult := stringBuilder.String()
	assert.Empty(t, jsonResult)
}

func TestClientLogPass(t *testing.T) {
	// Arrange
	startTime := time.Now().UTC().Local()
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelInfo.Level())
	client := *createTestClient(t, &stringBuilder, leveler)
	context := context.TODO()
	// Act
	client.Log(context, slog.LevelInfo, "test")
	// Assert
	jsonResult := stringBuilder.String()
	var jsonObject TestLog
	err := json.Unmarshal([]byte(jsonResult), &jsonObject)
	assert.Nil(t, err)
	logTime, err := time.Parse(time.RFC3339, jsonObject.Time)
	assert.Nil(t, err)
	assert.True(t, logTime.After(startTime))
	assert.Equal(t, "INFO", jsonObject.Level)
	assert.Contains(t, jsonObject.Source.Function, "util/log.log")
	assert.Contains(t, jsonObject.Source.File, "util/log/client_functions.go")
	assert.Greater(t, jsonObject.Source.Line, 0)
	assert.Equal(t, "test", jsonObject.Message)
}

func TestClientLogFail(t *testing.T) {
	// Arrange
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelWarn.Level())
	client := *createTestClient(t, &stringBuilder, leveler)
	context := context.TODO()
	// Act
	client.Log(context, slog.LevelInfo, "test")
	// Assert
	jsonResult := stringBuilder.String()
	assert.Empty(t, jsonResult)
}

func TestClientLogAttrsPass(t *testing.T) {
	// Arrange
	startTime := time.Now().UTC().Local()
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelInfo.Level())
	client := *createTestClient(t, &stringBuilder, leveler)
	context := context.TODO()
	// Act
	client.LogAttrs(context, slog.LevelInfo, "test", slog.Attr{Key: "key", Value: slog.StringValue("value")})
	// Assert
	jsonResult := stringBuilder.String()
	var jsonObject TestLog
	err := json.Unmarshal([]byte(jsonResult), &jsonObject)
	assert.Nil(t, err)
	logTime, err := time.Parse(time.RFC3339, jsonObject.Time)
	assert.Nil(t, err)
	assert.True(t, logTime.After(startTime))
	assert.Equal(t, "INFO", jsonObject.Level)
	assert.Contains(t, jsonObject.Source.Function, "util/log.logAttrs")
	assert.Contains(t, jsonObject.Source.File, "util/log/client_functions.go")
	assert.Greater(t, jsonObject.Source.Line, 0)
	assert.Equal(t, "test", jsonObject.Message)
}

func TestClientLogAttrsFail(t *testing.T) {
	// Arrange
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelWarn.Level())
	client := *createTestClient(t, &stringBuilder, leveler)
	context := context.TODO()
	// Act
	client.LogAttrs(context, slog.LevelInfo, "test", slog.Attr{Key: "key", Value: slog.StringValue("value")})
	// Assert
	jsonResult := stringBuilder.String()
	assert.Empty(t, jsonResult)
}

func TestClientWarnPass(t *testing.T) {
	// Arrange
	startTime := time.Now().UTC().Local()
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelWarn.Level())
	client := *createTestClient(t, &stringBuilder, leveler)
	// Act
	client.Warn("test")
	// Assert
	jsonResult := stringBuilder.String()
	var jsonObject TestLog
	err := json.Unmarshal([]byte(jsonResult), &jsonObject)
	assert.Nil(t, err)
	logTime, err := time.Parse(time.RFC3339, jsonObject.Time)
	assert.Nil(t, err)
	assert.True(t, logTime.After(startTime))
	assert.Equal(t, "WARN", jsonObject.Level)
	assert.Contains(t, jsonObject.Source.Function, "util/log.warn")
	assert.Contains(t, jsonObject.Source.File, "util/log/client_functions.go")
	assert.Greater(t, jsonObject.Source.Line, 0)
	assert.Equal(t, "test", jsonObject.Message)
}

func TestClientWarnFail(t *testing.T) {
	// Arrange
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelError.Level())
	client := *createTestClient(t, &stringBuilder, leveler)
	// Act
	client.Warn("test")
	// Assert
	jsonResult := stringBuilder.String()
	assert.Empty(t, jsonResult)
}

func TestClientWarnContextPass(t *testing.T) {
	// Arrange
	startTime := time.Now().UTC().Local()
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelWarn.Level())
	client := *createTestClient(t, &stringBuilder, leveler)
	context := context.TODO()
	// Act
	client.WarnContext(context, "test")
	// Assert
	jsonResult := stringBuilder.String()
	var jsonObject TestLog
	err := json.Unmarshal([]byte(jsonResult), &jsonObject)
	assert.Nil(t, err)
	logTime, err := time.Parse(time.RFC3339, jsonObject.Time)
	assert.Nil(t, err)
	assert.True(t, logTime.After(startTime))
	assert.Equal(t, "WARN", jsonObject.Level)
	assert.Contains(t, jsonObject.Source.Function, "util/log.warnContext")
	assert.Contains(t, jsonObject.Source.File, "util/log/client_functions.go")
	assert.Greater(t, jsonObject.Source.Line, 0)
	assert.Equal(t, "test", jsonObject.Message)
}

func TestClientWarnContextFail(t *testing.T) {
	// Arrange
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelError.Level())
	client := *createTestClient(t, &stringBuilder, leveler)
	context := context.TODO()
	// Act
	client.WarnContext(context, "test")
	// Assert
	jsonResult := stringBuilder.String()
	assert.Empty(t, jsonResult)
}

func TestClientWithPass(t *testing.T) {
	// Arrange
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelDebug.Level())
	client := *createTestClient(t, &stringBuilder, leveler)
	// Act
	newClient := client.With("key", "value")
	// Assert
	assert.NotNil(t, newClient)
	assert.NotEqual(t, client, newClient)
	originalLogger := client.Logger()
	newLogger := (*newClient).Logger()
	assert.NotEqual(t, originalLogger, newLogger)
}

func TestClientWithGroupPass(t *testing.T) {
	// Arrange
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelDebug.Level())
	client := *createTestClient(t, &stringBuilder, leveler)
	// Act
	newClient := client.WithGroup("group")
	// Assert
	assert.NotNil(t, newClient)
	assert.NotEqual(t, client, newClient)
	originalLogger := client.Logger()
	newLogger := (*newClient).Logger()
	assert.NotEqual(t, originalLogger, newLogger)
}
