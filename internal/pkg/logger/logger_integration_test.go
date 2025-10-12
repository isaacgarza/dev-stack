//go:build integration

package logger

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestIntegrationFileLogging(t *testing.T) {
	// Reset default logger
	originalLogger := defaultLogger
	defer func() { defaultLogger = originalLogger }()

	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "integration_test.log")

	t.Run("text format file logging", func(t *testing.T) {
		defaultLogger = nil
		config := Config{
			Level:      LevelDebug,
			Format:     "text",
			Output:     logFile,
			ColorOut:   false,
			TimeFormat: time.RFC3339,
		}

		err := Init(config)
		assert.NoError(t, err)

		// Log various messages
		Debug("debug message", "key", "debug_value")
		Info("info message", "key", "info_value")
		Warn("warn message", "key", "warn_value")
		Error("error message", "key", "error_value")

		// Read and verify file contents
		content, err := os.ReadFile(logFile)
		assert.NoError(t, err)

		logContent := string(content)
		assert.Contains(t, logContent, "debug message")
		assert.Contains(t, logContent, "info message")
		assert.Contains(t, logContent, "warn message")
		assert.Contains(t, logContent, "error message")
		assert.Contains(t, logContent, "key=debug_value")
		assert.Contains(t, logContent, "key=info_value")
		assert.Contains(t, logContent, "key=warn_value")
		assert.Contains(t, logContent, "key=error_value")
	})

	t.Run("json format file logging", func(t *testing.T) {
		defaultLogger = nil
		jsonLogFile := filepath.Join(tmpDir, "integration_json_test.log")

		config := Config{
			Level:      LevelInfo,
			Format:     "json",
			Output:     jsonLogFile,
			ColorOut:   false,
			TimeFormat: time.RFC3339,
		}

		err := Init(config)
		assert.NoError(t, err)

		// Log a message
		Info("json test message", "component", "integration_test", "count", 42)

		// Read and verify JSON format
		content, err := os.ReadFile(jsonLogFile)
		assert.NoError(t, err)

		// Parse as JSON
		var logEntry map[string]interface{}
		lines := strings.Split(strings.TrimSpace(string(content)), "\n")
		assert.Greater(t, len(lines), 0)

		err = json.Unmarshal([]byte(lines[0]), &logEntry)
		assert.NoError(t, err)

		assert.Equal(t, "json test message", logEntry["msg"])
		assert.Equal(t, "integration_test", logEntry["component"])
		assert.Equal(t, float64(42), logEntry["count"]) // JSON numbers are float64
		assert.Contains(t, logEntry, "time")
		assert.Contains(t, logEntry, "level")
	})
}

func TestIntegrationViperConfiguration(t *testing.T) {
	// Reset viper and default logger
	viper.Reset()
	originalLogger := defaultLogger
	defer func() {
		defaultLogger = originalLogger
		viper.Reset()
	}()

	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "viper_test.log")

	t.Run("full viper configuration", func(t *testing.T) {
		defaultLogger = nil

		// Set up viper configuration
		viper.Set("log.level", "warn")
		viper.Set("log.format", "json")
		viper.Set("log.output", logFile)
		viper.Set("log.color", false)

		err := InitFromViper()
		assert.NoError(t, err)

		// Test that debug and info are filtered out
		Debug("debug message - should not appear")
		Info("info message - should not appear")
		Warn("warn message - should appear", "test", "viper")
		Error("error message - should appear", "test", "viper")

		// Read and verify file contents
		content, err := os.ReadFile(logFile)
		assert.NoError(t, err)

		logContent := string(content)
		assert.NotContains(t, logContent, "debug message")
		assert.NotContains(t, logContent, "info message")
		assert.Contains(t, logContent, "warn message")
		assert.Contains(t, logContent, "error message")

		// Verify JSON format
		lines := strings.Split(strings.TrimSpace(logContent), "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				continue
			}
			var logEntry map[string]interface{}
			err := json.Unmarshal([]byte(line), &logEntry)
			assert.NoError(t, err, "Line should be valid JSON: %s", line)
		}
	})
}

func TestIntegrationLogRotation(t *testing.T) {
	// This test verifies that we can write to a file and append to it
	originalLogger := defaultLogger
	defer func() { defaultLogger = originalLogger }()

	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "rotation_test.log")

	config := Config{
		Level:      LevelInfo,
		Format:     "text",
		Output:     logFile,
		ColorOut:   false,
		TimeFormat: time.RFC3339,
	}

	// First initialization
	defaultLogger = nil
	err := Init(config)
	assert.NoError(t, err)

	Info("first message")

	// Second initialization (simulating restart)
	defaultLogger = nil
	err = Init(config)
	assert.NoError(t, err)

	Info("second message")

	// Verify both messages are in the file
	content, err := os.ReadFile(logFile)
	assert.NoError(t, err)

	logContent := string(content)
	assert.Contains(t, logContent, "first message")
	assert.Contains(t, logContent, "second message")

	// Count lines to ensure both messages are there
	scanner := bufio.NewScanner(strings.NewReader(logContent))
	lineCount := 0
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) != "" {
			lineCount++
		}
	}
	assert.Equal(t, 2, lineCount)
}

func TestIntegrationSpecializedLogging(t *testing.T) {
	originalLogger := defaultLogger
	defer func() { defaultLogger = originalLogger }()

	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "specialized_test.log")

	config := Config{
		Level:      LevelDebug,
		Format:     "json",
		Output:     logFile,
		ColorOut:   false,
		TimeFormat: time.RFC3339,
	}

	defaultLogger = nil
	err := Init(config)
	assert.NoError(t, err)

	// Test specialized logging functions
	LogCommand("docker", []string{"ps", "-a"}, time.Millisecond*250, nil)
	LogServiceAction("start", "redis", "port", 6379)
	LogProjectAction("init", "my-app", "template", "go")

	// Test operation logging
	complete := StartOperation("database-migration", "version", "1.2.3")
	time.Sleep(time.Millisecond * 10) // Small delay to show duration
	complete(nil)

	// Read and verify file contents
	content, err := os.ReadFile(logFile)
	assert.NoError(t, err)

	logContent := string(content)

	// Verify command logging
	assert.Contains(t, logContent, "Command executed successfully")
	assert.Contains(t, logContent, "docker")

	// Verify service action logging
	assert.Contains(t, logContent, "Service action")
	assert.Contains(t, logContent, "redis")

	// Verify project action logging
	assert.Contains(t, logContent, "Project action")
	assert.Contains(t, logContent, "my-app")

	// Verify operation logging
	assert.Contains(t, logContent, "Starting operation")
	assert.Contains(t, logContent, "Operation completed")
	assert.Contains(t, logContent, "database-migration")

	// Parse each line as JSON to ensure valid format
	lines := strings.Split(strings.TrimSpace(logContent), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var logEntry map[string]interface{}
		err := json.Unmarshal([]byte(line), &logEntry)
		assert.NoError(t, err, "Line should be valid JSON: %s", line)
	}
}

func TestIntegrationConcurrentLogging(t *testing.T) {
	originalLogger := defaultLogger
	defer func() { defaultLogger = originalLogger }()

	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "concurrent_test.log")

	config := Config{
		Level:      LevelInfo,
		Format:     "text",
		Output:     logFile,
		ColorOut:   false,
		TimeFormat: time.RFC3339,
	}

	defaultLogger = nil
	err := Init(config)
	assert.NoError(t, err)

	// Simulate concurrent logging from multiple goroutines
	done := make(chan bool, 3)

	for i := 0; i < 3; i++ {
		go func(id int) {
			for j := 0; j < 5; j++ {
				Info("concurrent message", "goroutine", id, "iteration", j)
				time.Sleep(time.Millisecond)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		<-done
	}

	// Read and verify file contents
	content, err := os.ReadFile(logFile)
	assert.NoError(t, err)

	logContent := string(content)

	// Count log entries
	scanner := bufio.NewScanner(strings.NewReader(logContent))
	lineCount := 0
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "concurrent message") {
			lineCount++
		}
	}

	// Should have 15 messages (3 goroutines * 5 messages each)
	assert.Equal(t, 15, lineCount)
}
