package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

var (
	// Default logger instance
	defaultLogger *slog.Logger
)

// LogLevel represents available log levels
type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
)

// Config represents logger configuration
type Config struct {
	Level      LogLevel `yaml:"level" json:"level"`
	Format     string   `yaml:"format" json:"format"` // json, text
	Output     string   `yaml:"output" json:"output"` // stdout, stderr, file path
	ColorOut   bool     `yaml:"color" json:"color"`
	TimeFormat string   `yaml:"time_format" json:"time_format"`
}

// DefaultConfig returns default logger configuration
func DefaultConfig() Config {
	return Config{
		Level:      LevelInfo,
		Format:     "text",
		Output:     "stdout",
		ColorOut:   true,
		TimeFormat: time.RFC3339,
	}
}

// Init initializes the logger with the given configuration
func Init(config Config) error {
	level := parseLogLevel(config.Level)

	var writer io.Writer
	switch config.Output {
	case "stdout":
		writer = os.Stdout
	case "stderr":
		writer = os.Stderr
	default:
		// Treat as file path
		file, err := os.OpenFile(config.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("failed to open log file %s: %w", config.Output, err)
		}
		writer = file
	}

	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Customize time format
			if a.Key == slog.TimeKey {
				t := a.Value.Time()
				a.Value = slog.StringValue(t.Format(config.TimeFormat))
			}
			return a
		},
	}

	switch config.Format {
	case "json":
		handler = slog.NewJSONHandler(writer, opts)
	default:
		handler = slog.NewTextHandler(writer, opts)
	}

	defaultLogger = slog.New(handler)
	return nil
}

// InitFromViper initializes the logger from viper configuration
func InitFromViper() error {
	config := DefaultConfig()

	if viper.IsSet("log.level") {
		config.Level = LogLevel(viper.GetString("log.level"))
	}
	if viper.IsSet("log.format") {
		config.Format = viper.GetString("log.format")
	}
	if viper.IsSet("log.output") {
		config.Output = viper.GetString("log.output")
	}
	if viper.IsSet("log.color") {
		config.ColorOut = viper.GetBool("log.color")
	}

	return Init(config)
}

// parseLogLevel converts string to slog.Level
func parseLogLevel(level LogLevel) slog.Level {
	switch strings.ToLower(string(level)) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// GetLogger returns the default logger instance
func GetLogger() *slog.Logger {
	if defaultLogger == nil {
		// Initialize with default config if not already initialized
		if err := Init(DefaultConfig()); err != nil {
			// Fallback to basic logger
			defaultLogger = slog.Default()
		}
	}
	return defaultLogger
}

// Debug logs a debug message
func Debug(msg string, args ...any) {
	GetLogger().Debug(msg, args...)
}

// Info logs an info message
func Info(msg string, args ...any) {
	GetLogger().Info(msg, args...)
}

// Warn logs a warning message
func Warn(msg string, args ...any) {
	GetLogger().Warn(msg, args...)
}

// Error logs an error message
func Error(msg string, args ...any) {
	GetLogger().Error(msg, args...)
}

// With returns a logger with the given attributes
func With(args ...any) *slog.Logger {
	return GetLogger().With(args...)
}

// WithGroup returns a logger with the given group name
func WithGroup(name string) *slog.Logger {
	return GetLogger().WithGroup(name)
}

// Fatal logs an error message and exits the program
func Fatal(msg string, args ...any) {
	GetLogger().Error(msg, args...)
	os.Exit(1)
}

// DebugEnabled returns true if debug logging is enabled
func DebugEnabled() bool {
	return GetLogger().Enabled(context.Background(), slog.LevelDebug)
}

// InfoEnabled returns true if info logging is enabled
func InfoEnabled() bool {
	return GetLogger().Enabled(context.Background(), slog.LevelInfo)
}

// WarnEnabled returns true if warn logging is enabled
func WarnEnabled() bool {
	return GetLogger().Enabled(context.Background(), slog.LevelWarn)
}

// ErrorEnabled returns true if error logging is enabled
func ErrorEnabled() bool {
	return GetLogger().Enabled(context.Background(), slog.LevelError)
}

// LogCommand logs command execution with timing
func LogCommand(cmd string, args []string, duration time.Duration, err error) {
	logger := GetLogger().With(
		"command", cmd,
		"args", args,
		"duration", duration,
	)

	if err != nil {
		logger.Error("Command failed", "error", err)
	} else {
		logger.Info("Command executed successfully")
	}
}

// LogServiceAction logs service-related actions
func LogServiceAction(action, service string, args ...any) {
	allArgs := append([]any{"action", action, "service", service}, args...)
	GetLogger().Info("Service action", allArgs...)
}

// LogProjectAction logs project-related actions
func LogProjectAction(action, project string, args ...any) {
	allArgs := append([]any{"action", action, "project", project}, args...)
	GetLogger().Info("Project action", allArgs...)
}

// LogError logs an error with additional context
func LogError(err error, msg string, args ...any) {
	allArgs := append([]any{"error", err}, args...)
	GetLogger().Error(msg, allArgs...)
}

// New creates a new logger with the specified level
func New(level slog.Level) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: level,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	return slog.New(handler)
}

// NewContextLogger creates a new logger with context fields
func NewContextLogger(fields map[string]any) *slog.Logger {
	var args []any
	for k, v := range fields {
		args = append(args, k, v)
	}
	return GetLogger().With(args...)
}

// StartOperation logs the start of an operation and returns a function to log completion
func StartOperation(operation string, args ...any) func(error) {
	start := time.Now()
	allArgs := append([]any{"operation", operation}, args...)
	GetLogger().Info("Starting operation", allArgs...)

	return func(err error) {
		duration := time.Since(start)
		completeArgs := append(allArgs, "duration", duration)

		if err != nil {
			completeArgs = append(completeArgs, "error", err)
			GetLogger().Error("Operation failed", completeArgs...)
		} else {
			GetLogger().Info("Operation completed", completeArgs...)
		}
	}
}
