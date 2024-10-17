package log

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// create test client
func createTestClient(t *testing.T, writer io.Writer, level slog.Leveler) *Client {
	t.Helper()
	opts := slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	}
	jsonHandler := slog.NewJSONHandler(writer, &opts)
	logger := slog.New(jsonHandler)

	client := NewClient(
		debug,
		debugContext,
		enabled,
		errorOut,
		errorContext,
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
	assert.NotNil(t, client)
	return &client
}

// debug test
func TestDebugPass(t *testing.T) {
	// Arrange
	startTime := time.Now().UTC().Local()
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelDebug.Level())
	client := createTestClient(t, &stringBuilder, leveler)
	// Act
	debug(*client, "test")
	// Assert
	jsonResult := stringBuilder.String()
	var jsonObject TestLog
	err := json.Unmarshal([]byte(jsonResult), &jsonObject)
	require.NoError(t, err)
	logTime, err := time.Parse(time.RFC3339, jsonObject.Time)
	require.NoError(t, err)
	assert.False(t, logTime.Before(startTime))
	assert.Equal(t, "DEBUG", jsonObject.Level)
	assert.Contains(t, jsonObject.Source.Function, "util/log.debug")
	assert.Contains(t, jsonObject.Source.File, "util/log/client_functions.go")
	assert.Positive(t, jsonObject.Source.Line, 0)
	assert.Equal(t, "test", jsonObject.Message)
}

func TestDebugFail(t *testing.T) {
	// Arrange
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelInfo.Level())
	client := createTestClient(t, &stringBuilder, leveler)
	// Act
	debug(*client, "test")
	// Assert
	assert.Empty(t, stringBuilder.String())
}

func TestDebugContextPass(t *testing.T) {
	// Arrange
	startTime := time.Now().UTC().Local()
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelDebug.Level())
	client := createTestClient(t, &stringBuilder, leveler)
	context := context.TODO()
	// Act
	debugContext(context, *client, "test")
	// Assert
	jsonResult := stringBuilder.String()
	var jsonObject TestLog
	err := json.Unmarshal([]byte(jsonResult), &jsonObject)
	require.NoError(t, err)
	logTime, err := time.Parse(time.RFC3339, jsonObject.Time)
	require.NoError(t, err)
	assert.False(t, logTime.Before(startTime))
	assert.Equal(t, "DEBUG", jsonObject.Level)
	assert.Contains(t, jsonObject.Source.Function, "util/log.debugContext")
	assert.Contains(t, jsonObject.Source.File, "util/log/client_functions.go")
	assert.Positive(t, jsonObject.Source.Line, 0)
	assert.Equal(t, "test", jsonObject.Message)
}

func TestDebugContextFail(t *testing.T) {
	// Arrange
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelInfo.Level())
	client := createTestClient(t, &stringBuilder, leveler)
	context := context.TODO()
	// Act
	debugContext(context, *client, "test")
	// Assert
	assert.Empty(t, stringBuilder.String())
}

func TestEnabledPass(t *testing.T) {
	// Arrange
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelDebug.Level())
	client := createTestClient(t, &stringBuilder, leveler)
	context := context.TODO()
	// Act
	result := enabled(context, *client, slog.LevelDebug)
	// Assert
	assert.True(t, result)
}

func TestEnabledFail(t *testing.T) {
	// Arrange
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelInfo.Level())
	client := createTestClient(t, &stringBuilder, leveler)
	context := context.TODO()
	// Act
	result := enabled(context, *client, slog.LevelDebug)
	// Assert
	assert.False(t, result)
}

func TestErrorOut(t *testing.T) {
	// Arrange
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelError.Level())
	client := createTestClient(t, &stringBuilder, leveler)
	// Act
	errorOut(*client, "test")
	// Assert
	jsonResult := stringBuilder.String()
	var jsonObject TestLog
	err := json.Unmarshal([]byte(jsonResult), &jsonObject)
	require.NoError(t, err)
	assert.Equal(t, "ERROR", jsonObject.Level)
	assert.Contains(t, jsonObject.Source.Function, "util/log.errorOut")
	assert.Contains(t, jsonObject.Source.File, "util/log/client_functions.go")
	assert.Positive(t, jsonObject.Source.Line, 0)
	assert.Equal(t, "test", jsonObject.Message)
}

func TestErrorContext(t *testing.T) {
	// Arrange
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelError.Level())
	client := createTestClient(t, &stringBuilder, leveler)
	context := context.TODO()
	// Act
	errorContext(context, *client, "test")
	// Assert
	jsonResult := stringBuilder.String()
	var jsonObject TestLog
	err := json.Unmarshal([]byte(jsonResult), &jsonObject)
	require.NoError(t, err)
	assert.Equal(t, "ERROR", jsonObject.Level)
	assert.Contains(t, jsonObject.Source.Function, "util/log.errorContext")
	assert.Contains(t, jsonObject.Source.File, "util/log/client_functions.go")
	assert.Positive(t, jsonObject.Source.Line, 0)
	assert.Equal(t, "test", jsonObject.Message)
}

func TestHandlerPass(t *testing.T) {
	// Arrange
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelDebug.Level())
	client := createTestClient(t, &stringBuilder, leveler)
	// Act
	handler := handlerFunc(*client)
	// Assert
	assert.NotNil(t, handler)
}

func TestInfoPass(t *testing.T) {
	// Arrange
	startTime := time.Now().UTC().Local()
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelInfo.Level())
	client := createTestClient(t, &stringBuilder, leveler)
	// Act
	info(*client, "test")
	// Assert
	jsonResult := stringBuilder.String()
	var jsonObject TestLog
	err := json.Unmarshal([]byte(jsonResult), &jsonObject)
	require.NoError(t, err)
	logTime, err := time.Parse(time.RFC3339, jsonObject.Time)
	require.NoError(t, err)
	assert.False(t, logTime.Before(startTime))
	assert.Equal(t, "INFO", jsonObject.Level)
	assert.Contains(t, jsonObject.Source.Function, "util/log.info")
	assert.Contains(t, jsonObject.Source.File, "util/log/client_functions.go")
	assert.Positive(t, jsonObject.Source.Line, 0)
	assert.Equal(t, "test", jsonObject.Message)
}

func TestInfoFail(t *testing.T) {
	// Arrange
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelWarn.Level())
	client := createTestClient(t, &stringBuilder, leveler)
	// Act
	info(*client, "test")
	// Assert
	assert.Empty(t, stringBuilder.String())
}

func TestInfoContextPass(t *testing.T) {
	// Arrange
	startTime := time.Now().UTC().Local()
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelInfo.Level())
	client := createTestClient(t, &stringBuilder, leveler)
	context := context.TODO()
	// Act
	infoContext(context, *client, "test")
	// Assert
	jsonResult := stringBuilder.String()
	var jsonObject TestLog
	err := json.Unmarshal([]byte(jsonResult), &jsonObject)
	require.NoError(t, err)
	logTime, err := time.Parse(time.RFC3339, jsonObject.Time)
	require.NoError(t, err)
	assert.False(t, logTime.Before(startTime))
	assert.Equal(t, "INFO", jsonObject.Level)
	assert.Contains(t, jsonObject.Source.Function, "util/log.infoContext")
	assert.Contains(t, jsonObject.Source.File, "util/log/client_functions.go")
	assert.Positive(t, jsonObject.Source.Line, 0)
	assert.Equal(t, "test", jsonObject.Message)
}

func TestInfoContextFail(t *testing.T) {
	// Arrange
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelWarn.Level())
	client := createTestClient(t, &stringBuilder, leveler)
	context := context.TODO()
	// Act
	infoContext(context, *client, "test")
	// Assert
	assert.Empty(t, stringBuilder.String())
}

func TestLogPass(t *testing.T) {
	// Arrange
	startTime := time.Now().UTC().Local()
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelInfo.Level())
	client := createTestClient(t, &stringBuilder, leveler)
	context := context.TODO()
	// Act
	log(context, *client, slog.LevelInfo, "test")
	// Assert
	jsonResult := stringBuilder.String()
	var jsonObject TestLog
	err := json.Unmarshal([]byte(jsonResult), &jsonObject)
	require.NoError(t, err)
	logTime, err := time.Parse(time.RFC3339, jsonObject.Time)
	require.NoError(t, err)
	assert.False(t, logTime.Before(startTime))
	assert.Equal(t, "INFO", jsonObject.Level)
	assert.Contains(t, jsonObject.Source.Function, "util/log.log")
	assert.Contains(t, jsonObject.Source.File, "util/log/client_functions.go")
	assert.Positive(t, jsonObject.Source.Line, 0)
	assert.Equal(t, "test", jsonObject.Message)
}

func TestLogFail(t *testing.T) {
	// Arrange
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelWarn.Level())
	client := createTestClient(t, &stringBuilder, leveler)
	context := context.TODO()
	// Act
	log(context, *client, slog.LevelInfo, "test")
	// Assert
	assert.Empty(t, stringBuilder.String())
}

func TestLogAttrsPass(t *testing.T) {
	// Arrange
	startTime := time.Now().UTC().Local()
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelInfo.Level())
	client := createTestClient(t, &stringBuilder, leveler)
	context := context.TODO()
	// Act
	logAttrs(context, *client, slog.LevelInfo, "test", slog.Attr{Key: "key", Value: slog.StringValue("value")})
	// Assert
	jsonResult := stringBuilder.String()
	var jsonObject TestLog
	err := json.Unmarshal([]byte(jsonResult), &jsonObject)
	require.NoError(t, err)
	logTime, err := time.Parse(time.RFC3339, jsonObject.Time)
	require.NoError(t, err)
	assert.False(t, logTime.Before(startTime))
	assert.Equal(t, "INFO", jsonObject.Level)
	assert.Contains(t, jsonObject.Source.Function, "util/log.logAttrs")
	assert.Contains(t, jsonObject.Source.File, "util/log/client_functions.go")
	assert.Positive(t, jsonObject.Source.Line, 0)
	assert.Equal(t, "test", jsonObject.Message)
}

func TestLogAttrsFail(t *testing.T) {
	// Arrange
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelWarn.Level())
	client := createTestClient(t, &stringBuilder, leveler)
	context := context.TODO()
	// Act
	logAttrs(context, *client, slog.LevelInfo, "test", slog.Attr{Key: "key", Value: slog.StringValue("value")})
	// Assert
	assert.Empty(t, stringBuilder.String())
}

func TestWarnPass(t *testing.T) {
	// Arrange
	startTime := time.Now().UTC().Local()
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelWarn.Level())
	client := createTestClient(t, &stringBuilder, leveler)
	// Act
	warn(*client, "test")
	// Assert
	jsonResult := stringBuilder.String()
	var jsonObject TestLog
	err := json.Unmarshal([]byte(jsonResult), &jsonObject)
	require.NoError(t, err)
	logTime, err := time.Parse(time.RFC3339, jsonObject.Time)
	require.NoError(t, err)
	assert.False(t, logTime.Before(startTime))
	assert.Equal(t, "WARN", jsonObject.Level)
	assert.Contains(t, jsonObject.Source.Function, "util/log.warn")
	assert.Contains(t, jsonObject.Source.File, "util/log/client_functions.go")
	assert.Positive(t, jsonObject.Source.Line, 0)
	assert.Equal(t, "test", jsonObject.Message)
}

func TestWarnFail(t *testing.T) {
	// Arrange
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelError.Level())
	client := createTestClient(t, &stringBuilder, leveler)
	// Act
	warn(*client, "test")
	// Assert
	assert.Empty(t, stringBuilder.String())
}

func TestWarnContextPass(t *testing.T) {
	// Arrange
	startTime := time.Now().UTC().Local()
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelWarn.Level())
	client := createTestClient(t, &stringBuilder, leveler)
	context := context.TODO()
	// Act
	warnContext(context, *client, "test")
	// Assert
	jsonResult := stringBuilder.String()
	var jsonObject TestLog
	err := json.Unmarshal([]byte(jsonResult), &jsonObject)
	require.NoError(t, err)
	logTime, err := time.Parse(time.RFC3339, jsonObject.Time)
	require.NoError(t, err)
	assert.False(t, logTime.Before(startTime))
	assert.Equal(t, "WARN", jsonObject.Level)
	assert.Contains(t, jsonObject.Source.Function, "util/log.warnContext")
	assert.Contains(t, jsonObject.Source.File, "util/log/client_functions.go")
	assert.Positive(t, jsonObject.Source.Line, 0)
	assert.Equal(t, "test", jsonObject.Message)
}

func TestWarnContextFail(t *testing.T) {
	// Arrange
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelError.Level())
	client := createTestClient(t, &stringBuilder, leveler)
	context := context.TODO()
	// Act
	warnContext(context, *client, "test")
	// Assert
	assert.Empty(t, stringBuilder.String())
}

func TestWithPass(t *testing.T) {
	// Arrange
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelDebug.Level())
	client := createTestClient(t, &stringBuilder, leveler)
	// Act
	newClient := with(*client, "test")
	// Assert
	assert.NotNil(t, newClient)
	assert.NotEqual(t, client, newClient)
	originalLogger := (*client).Logger()
	newLogger := (*newClient).Logger()
	assert.NotEqual(t, originalLogger, newLogger)
}

func TestWithGroupPass(t *testing.T) {
	// Arrange
	var stringBuilder strings.Builder
	leveler := slog.Leveler(slog.LevelDebug.Level())
	client := createTestClient(t, &stringBuilder, leveler)
	// Act
	newClient := withGroup(*client, "test")
	// Assert
	assert.NotNil(t, newClient)
	assert.NotEqual(t, client, newClient)
	originalLogger := (*client).Logger()
	newLogger := (*newClient).Logger()
	assert.NotEqual(t, originalLogger, newLogger)
}
