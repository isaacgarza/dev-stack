package services

import (
	"log/slog"
	"os"
	"testing"

	"github.com/isaacgarza/dev-stack/internal/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T) (logger *slog.Logger, projectDir string)
		expectError bool
		errorMsg    string
	}{
		{
			name: "create manager with valid parameters",
			setupFunc: func(t *testing.T) (*slog.Logger, string) {
				logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
				projectDir := t.TempDir()
				return logger, projectDir
			},
			expectError: false,
		},
		{
			name: "create manager with nil logger",
			setupFunc: func(t *testing.T) (*slog.Logger, string) {
				projectDir := t.TempDir()
				return nil, projectDir
			},
			expectError: false, // Docker client should handle nil logger
		},
		{
			name: "create manager with empty project dir",
			setupFunc: func(t *testing.T) (*slog.Logger, string) {
				logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
				return logger, ""
			},
			expectError: false, // Empty project dir should be allowed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, projectDir := tt.setupFunc(t)

			manager, err := NewManager(logger, projectDir)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, manager)
			} else {
				if err != nil {
					// Docker might not be available in test environment
					t.Skipf("Docker not available: %v", err)
				}
				require.NotNil(t, manager)
				assert.Equal(t, projectDir, manager.projectDir)
				assert.NotNil(t, manager.operations)
				assert.NotNil(t, manager.cleanup)

				// Clean up
				_ = manager.Close()
			}
		})
	}
}

func TestManager_SetConfig(t *testing.T) {
	// Skip if Docker not available
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	projectDir := t.TempDir()

	manager, err := NewManager(logger, projectDir)
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer func() { _ = manager.Close() }()

	tests := []struct {
		name   string
		config *types.Config
	}{
		{
			name:   "set nil config",
			config: nil,
		},
		{
			name: "set valid config",
			config: &types.Config{
				Global: types.GlobalConfig{
					LogLevel:    "info",
					ColorOutput: true,
				},
				Projects: make(map[string]types.ProjectConfig),
				Profiles: make(map[string]types.Profile),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			assert.NotPanics(t, func() {
				manager.SetConfig(tt.config)
			})

			assert.Equal(t, tt.config, manager.config)
		})
	}
}

func TestManager_Close(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	projectDir := t.TempDir()

	manager, err := NewManager(logger, projectDir)
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	t.Run("close manager", func(t *testing.T) {
		err := manager.Close()
		// Should not error even if Docker client close fails
		assert.NoError(t, err)
	})

	t.Run("close manager twice", func(t *testing.T) {
		// Should not panic on double close
		assert.NotPanics(t, func() {
			_ = manager.Close()
		})
	})
}

func TestManager_GetDocker(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	projectDir := t.TempDir()

	manager, err := NewManager(logger, projectDir)
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer func() { _ = manager.Close() }()

	t.Run("get docker client", func(t *testing.T) {
		// Test that docker client is accessible (if the field is exported)
		assert.NotNil(t, manager.docker)
	})
}

func TestManager_GetLogger(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	projectDir := t.TempDir()

	manager, err := NewManager(logger, projectDir)
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer func() { _ = manager.Close() }()

	t.Run("get logger", func(t *testing.T) {
		// Test that logger is accessible (if the field is exported)
		assert.NotNil(t, manager.logger)
		assert.Equal(t, logger, manager.logger)
	})
}

func TestManager_GetProjectDir(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	projectDir := t.TempDir()

	manager, err := NewManager(logger, projectDir)
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer func() { _ = manager.Close() }()

	t.Run("get project directory", func(t *testing.T) {
		assert.Equal(t, projectDir, manager.projectDir)
	})
}

func TestManager_SubManagers(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	projectDir := t.TempDir()

	manager, err := NewManager(logger, projectDir)
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer func() { _ = manager.Close() }()

	t.Run("operations sub-manager initialized", func(t *testing.T) {
		assert.NotNil(t, manager.operations)
	})

	t.Run("cleanup sub-manager initialized", func(t *testing.T) {
		assert.NotNil(t, manager.cleanup)
	})
}
