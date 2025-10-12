package core

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	cliTypes "github.com/isaacgarza/dev-stack/internal/pkg/cli/types"
)

// MockLogger implements the Logger interface for testing
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Info(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *MockLogger) Error(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *MockLogger) Debug(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func TestNewUpHandler(t *testing.T) {
	handler := NewUpHandler()
	assert.NotNil(t, handler)
	assert.IsType(t, &UpHandler{}, handler)
}

func TestUpHandler_ValidateArgs(t *testing.T) {
	handler := NewUpHandler()

	t.Run("no args", func(t *testing.T) {
		err := handler.ValidateArgs([]string{})
		assert.NoError(t, err)
	})

	t.Run("with args", func(t *testing.T) {
		err := handler.ValidateArgs([]string{"service1", "service2"})
		assert.NoError(t, err)
	})
}

func TestUpHandler_GetRequiredFlags(t *testing.T) {
	handler := NewUpHandler()
	flags := handler.GetRequiredFlags()
	assert.Empty(t, flags)
}

func TestUpHandler_Handle_ConfigNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	handler := NewUpHandler()
	mockLogger := &MockLogger{}

	cmd := &cobra.Command{}
	cmd.Flags().Bool("build", false, "")
	cmd.Flags().Bool("force-recreate", false, "")

	base := &cliTypes.BaseCommand{
		ProjectDir: tmpDir,
		Logger:     mockLogger,
	}

	err := handler.Handle(context.Background(), cmd, []string{}, base)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestLoadProjectConfig(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("valid config", func(t *testing.T) {
		configContent := `project:
  name: test-project
  environment: development
stack:
  enabled:
    - redis
    - postgres`

		configPath := filepath.Join(tmpDir, "config.yml")
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		assert.NoError(t, err)

		cfg, err := LoadProjectConfig(configPath)
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "test-project", cfg.Project.Name)
		assert.Equal(t, "development", cfg.Project.Environment)
		assert.Equal(t, []string{"redis", "postgres"}, cfg.Stack.Enabled)
	})

	t.Run("file not found", func(t *testing.T) {
		cfg, err := LoadProjectConfig("/nonexistent/config.yml")
		assert.Error(t, err)
		assert.Nil(t, cfg)
	})

	t.Run("invalid YAML", func(t *testing.T) {
		configPath := filepath.Join(tmpDir, "invalid.yml")
		err := os.WriteFile(configPath, []byte("invalid: yaml: [unclosed"), 0644)
		assert.NoError(t, err)

		cfg, err := LoadProjectConfig(configPath)
		assert.Error(t, err)
		assert.Nil(t, cfg)
		assert.Contains(t, err.Error(), "failed to parse config")
	})

	t.Run("empty config", func(t *testing.T) {
		configPath := filepath.Join(tmpDir, "empty.yml")
		err := os.WriteFile(configPath, []byte(""), 0644)
		assert.NoError(t, err)

		cfg, err := LoadProjectConfig(configPath)
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Empty(t, cfg.Project.Name)
		assert.Empty(t, cfg.Stack.Enabled)
	})
}

func TestProjectConfig_Structure(t *testing.T) {
	cfg := ProjectConfig{}

	// Test that struct fields exist and can be set
	cfg.Project.Name = "test"
	cfg.Project.Environment = "test"
	cfg.Stack.Enabled = []string{"service1"}

	assert.Equal(t, "test", cfg.Project.Name)
	assert.Equal(t, "test", cfg.Project.Environment)
	assert.Equal(t, []string{"service1"}, cfg.Stack.Enabled)
}

func TestMockLogger(t *testing.T) {
	mockLogger := &MockLogger{}

	mockLogger.On("Info", "test info", mock.Anything).Return()
	mockLogger.On("Error", "test error", mock.Anything).Return()
	mockLogger.On("Debug", "test debug", mock.Anything).Return()

	// Test interface methods
	mockLogger.Info("test info", "key", "value")
	mockLogger.Error("test error", "key", "value")
	mockLogger.Debug("test debug", "key", "value")

	mockLogger.AssertExpectations(t)
}
