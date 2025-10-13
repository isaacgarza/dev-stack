package logger

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, LevelInfo, config.Level)
	assert.Equal(t, "text", config.Format)
	assert.Equal(t, "stdout", config.Output)
	assert.True(t, config.ColorOut)
	assert.Equal(t, time.RFC3339, config.TimeFormat)
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		input    LogLevel
		expected slog.Level
	}{
		{"debug", LevelDebug, slog.LevelDebug},
		{"info", LevelInfo, slog.LevelInfo},
		{"warn", LevelWarn, slog.LevelWarn},
		{"warning", LogLevel("warning"), slog.LevelWarn},
		{"error", LevelError, slog.LevelError},
		{"invalid", LogLevel("invalid"), slog.LevelInfo},
		{"uppercase", LogLevel("DEBUG"), slog.LevelDebug},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseLogLevel(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInit(t *testing.T) {
	// Reset default logger before each test
	originalLogger := defaultLogger
	defer func() { defaultLogger = originalLogger }()

	t.Run("stdout output", func(t *testing.T) {
		defaultLogger = nil
		config := Config{
			Level:      LevelDebug,
			Format:     "text",
			Output:     "stdout",
			ColorOut:   true,
			TimeFormat: time.RFC3339,
		}

		err := Init(config)
		assert.NoError(t, err)
		assert.NotNil(t, defaultLogger)
	})

	t.Run("stderr output", func(t *testing.T) {
		defaultLogger = nil
		config := Config{
			Level:      LevelInfo,
			Format:     "json",
			Output:     "stderr",
			ColorOut:   false,
			TimeFormat: time.Kitchen,
		}

		err := Init(config)
		assert.NoError(t, err)
		assert.NotNil(t, defaultLogger)
	})

	t.Run("file output", func(t *testing.T) {
		defaultLogger = nil
		tmpDir := t.TempDir()
		logFile := filepath.Join(tmpDir, "test.log")

		config := Config{
			Level:      LevelWarn,
			Format:     "json",
			Output:     logFile,
			ColorOut:   false,
			TimeFormat: time.RFC3339,
		}

		err := Init(config)
		assert.NoError(t, err)
		assert.NotNil(t, defaultLogger)

		// Write a test log to ensure file is created
		Warn("test message")

		// Verify file exists and has content
		info, err := os.Stat(logFile)
		assert.NoError(t, err)
		assert.True(t, info.Size() > 0, "Log file should have content")
	})

	t.Run("invalid file path", func(t *testing.T) {
		defaultLogger = nil
		config := Config{
			Level:  LevelInfo,
			Format: "text",
			Output: "/invalid/path/test.log",
		}

		err := Init(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to open log file")
	})
}

func TestInitFromViper(t *testing.T) {
	// Reset viper and default logger
	viper.Reset()
	originalLogger := defaultLogger
	defer func() {
		defaultLogger = originalLogger
		viper.Reset()
	}()

	t.Run("with viper config", func(t *testing.T) {
		defaultLogger = nil
		viper.Set("log.level", "debug")
		viper.Set("log.format", "json")
		viper.Set("log.output", "stderr")
		viper.Set("log.color", false)

		err := InitFromViper()
		assert.NoError(t, err)
		assert.NotNil(t, defaultLogger)
	})

	t.Run("without viper config", func(t *testing.T) {
		defaultLogger = nil
		viper.Reset()

		err := InitFromViper()
		assert.NoError(t, err)
		assert.NotNil(t, defaultLogger)
	})
}

func TestGetLogger(t *testing.T) {
	originalLogger := defaultLogger
	defer func() { defaultLogger = originalLogger }()

	t.Run("with initialized logger", func(t *testing.T) {
		defaultLogger = slog.Default()
		logger := GetLogger()
		assert.NotNil(t, logger)
		assert.Equal(t, defaultLogger, logger)
	})

	t.Run("without initialized logger", func(t *testing.T) {
		defaultLogger = nil
		logger := GetLogger()
		assert.NotNil(t, logger)
		assert.NotNil(t, defaultLogger)
	})
}

func TestLoggingFunctions(t *testing.T) {
	// Capture output
	var buf bytes.Buffer
	originalLogger := defaultLogger
	defer func() { defaultLogger = originalLogger }()

	// Create logger with buffer output
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	defaultLogger = slog.New(handler)

	t.Run("debug logging", func(t *testing.T) {
		buf.Reset()
		Debug("test debug message", "key", "value")
		output := buf.String()
		assert.Contains(t, output, "test debug message")
		assert.Contains(t, output, "key=value")
	})

	t.Run("info logging", func(t *testing.T) {
		buf.Reset()
		Info("test info message", "key", "value")
		output := buf.String()
		assert.Contains(t, output, "test info message")
		assert.Contains(t, output, "key=value")
	})

	t.Run("warn logging", func(t *testing.T) {
		buf.Reset()
		Warn("test warn message", "key", "value")
		output := buf.String()
		assert.Contains(t, output, "test warn message")
		assert.Contains(t, output, "key=value")
	})

	t.Run("error logging", func(t *testing.T) {
		buf.Reset()
		Error("test error message", "key", "value")
		output := buf.String()
		assert.Contains(t, output, "test error message")
		assert.Contains(t, output, "key=value")
	})
}

func TestWith(t *testing.T) {
	originalLogger := defaultLogger
	defer func() { defaultLogger = originalLogger }()

	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	defaultLogger = slog.New(handler)

	logger := With("component", "test")
	assert.NotNil(t, logger)

	logger.Info("test message")
	output := buf.String()
	assert.Contains(t, output, "component=test")
	assert.Contains(t, output, "test message")
}

func TestWithGroup(t *testing.T) {
	originalLogger := defaultLogger
	defer func() { defaultLogger = originalLogger }()

	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	defaultLogger = slog.New(handler)

	logger := WithGroup("testgroup")
	assert.NotNil(t, logger)

	logger.Info("test message", "key", "value")
	output := buf.String()
	assert.Contains(t, output, "test message")
	assert.Contains(t, output, "testgroup")
}

func TestLevelEnabled(t *testing.T) {
	originalLogger := defaultLogger
	defer func() { defaultLogger = originalLogger }()

	// Create logger with INFO level
	handler := slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	defaultLogger = slog.New(handler)

	assert.False(t, DebugEnabled())
	assert.True(t, InfoEnabled())
	assert.True(t, WarnEnabled())
	assert.True(t, ErrorEnabled())
}

func TestLogCommand(t *testing.T) {
	originalLogger := defaultLogger
	defer func() { defaultLogger = originalLogger }()

	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	defaultLogger = slog.New(handler)

	t.Run("successful command", func(t *testing.T) {
		buf.Reset()
		LogCommand("test-cmd", []string{"arg1", "arg2"}, time.Second, nil)
		output := buf.String()
		assert.Contains(t, output, "Command executed successfully")
		assert.Contains(t, output, "command=test-cmd")
		assert.Contains(t, output, "duration=1s")
	})

	t.Run("failed command", func(t *testing.T) {
		buf.Reset()
		testErr := errors.New("command failed")
		LogCommand("test-cmd", []string{"arg1"}, time.Millisecond*500, testErr)
		output := buf.String()
		assert.Contains(t, output, "Command failed")
		assert.Contains(t, output, "command=test-cmd")
		assert.Contains(t, output, "error=\"command failed\"")
	})
}

func TestLogServiceAction(t *testing.T) {
	originalLogger := defaultLogger
	defer func() { defaultLogger = originalLogger }()

	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	defaultLogger = slog.New(handler)

	LogServiceAction("start", "redis", "port", 6379)
	output := buf.String()
	assert.Contains(t, output, "Service action")
	assert.Contains(t, output, "action=start")
	assert.Contains(t, output, "service=redis")
	assert.Contains(t, output, "port=6379")
}

func TestLogProjectAction(t *testing.T) {
	originalLogger := defaultLogger
	defer func() { defaultLogger = originalLogger }()

	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	defaultLogger = slog.New(handler)

	LogProjectAction("init", "my-project", "template", "go")
	output := buf.String()
	assert.Contains(t, output, "Project action")
	assert.Contains(t, output, "action=init")
	assert.Contains(t, output, "project=my-project")
	assert.Contains(t, output, "template=go")
}

func TestLogError(t *testing.T) {
	originalLogger := defaultLogger
	defer func() { defaultLogger = originalLogger }()

	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	defaultLogger = slog.New(handler)

	testErr := errors.New("test error")
	LogError(testErr, "operation failed", "context", "test")
	output := buf.String()
	assert.Contains(t, output, "operation failed")
	assert.Contains(t, output, "error=\"test error\"")
	assert.Contains(t, output, "context=test")
}

func TestNew(t *testing.T) {
	logger := New(slog.LevelWarn)
	assert.NotNil(t, logger)

	// Test that it respects the level
	assert.False(t, logger.Enabled(context.Background(), slog.LevelDebug))
	assert.False(t, logger.Enabled(context.Background(), slog.LevelInfo))
	assert.True(t, logger.Enabled(context.Background(), slog.LevelWarn))
	assert.True(t, logger.Enabled(context.Background(), slog.LevelError))
}

func TestNewContextLogger(t *testing.T) {
	originalLogger := defaultLogger
	defer func() { defaultLogger = originalLogger }()

	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	defaultLogger = slog.New(handler)

	fields := map[string]any{
		"component": "test",
		"version":   "1.0.0",
		"env":       "testing",
	}

	logger := NewContextLogger(fields)
	assert.NotNil(t, logger)

	logger.Info("test message")
	output := buf.String()
	assert.Contains(t, output, "test message")
	assert.Contains(t, output, "component=test")
	assert.Contains(t, output, "version=1.0.0")
	assert.Contains(t, output, "env=testing")
}

func TestStartOperation(t *testing.T) {
	originalLogger := defaultLogger
	defer func() { defaultLogger = originalLogger }()

	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	defaultLogger = slog.New(handler)

	t.Run("successful operation", func(t *testing.T) {
		buf.Reset()
		complete := StartOperation("test-operation", "param", "value")

		// Check start message
		output := buf.String()
		assert.Contains(t, output, "Starting operation")
		assert.Contains(t, output, "operation=test-operation")
		assert.Contains(t, output, "param=value")

		// Complete the operation
		buf.Reset()
		complete(nil)

		output = buf.String()
		assert.Contains(t, output, "Operation completed")
		assert.Contains(t, output, "operation=test-operation")
		assert.Contains(t, output, "duration=")
	})

	t.Run("failed operation", func(t *testing.T) {
		buf.Reset()
		complete := StartOperation("test-operation", "param", "value")

		buf.Reset()
		testErr := errors.New("operation failed")
		complete(testErr)

		output := buf.String()
		assert.Contains(t, output, "Operation failed")
		assert.Contains(t, output, "operation=test-operation")
		assert.Contains(t, output, "error=\"operation failed\"")
		assert.Contains(t, output, "duration=")
	})
}

func TestLogLevels(t *testing.T) {
	tests := []struct {
		name  string
		level LogLevel
	}{
		{"debug level", LevelDebug},
		{"info level", LevelInfo},
		{"warn level", LevelWarn},
		{"error level", LevelError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, string(tt.level))
		})
	}
}

func TestConfigValidation(t *testing.T) {
	t.Run("valid formats", func(t *testing.T) {
		formats := []string{"text", "json"}
		for _, format := range formats {
			config := DefaultConfig()
			config.Format = format
			err := Init(config)
			assert.NoError(t, err)
		}
	})

	t.Run("valid outputs", func(t *testing.T) {
		outputs := []string{"stdout", "stderr"}
		for _, output := range outputs {
			config := DefaultConfig()
			config.Output = output
			err := Init(config)
			assert.NoError(t, err)
		}
	})
}
