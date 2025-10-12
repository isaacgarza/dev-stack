package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLoader(t *testing.T) {
	tests := []struct {
		name       string
		configPath string
	}{
		{
			name:       "empty path",
			configPath: "",
		},
		{
			name:       "valid path",
			configPath: "/path/to/config.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewLoader(tt.configPath)
			assert.NotNil(t, loader)
			assert.Equal(t, tt.configPath, loader.configPath)
			assert.Nil(t, loader.cache)
		})
	}
}

func TestLoader_Load(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T) string
		expectError bool
		errorMsg    string
	}{
		{
			name: "load valid YAML file",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configFile := filepath.Join(tmpDir, "config.yaml")

				validYAML := `
metadata:
  version: "1.0.0"
  cli_version: "1.0.0"
  description: "Test config"
global:
  flags: {}
categories: {}
commands:
  test:
    description: "Test command"
    usage: "test [options]"
workflows: {}
profiles: {}
help: {}
`
				err := os.WriteFile(configFile, []byte(validYAML), 0644)
				require.NoError(t, err)
				return configFile
			},
			expectError: false,
		},
		{
			name: "invalid YAML file",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configFile := filepath.Join(tmpDir, "config.yaml")

				invalidYAML := `
metadata:
  version: "1.0.0"
  invalid_yaml: [unclosed
`
				err := os.WriteFile(configFile, []byte(invalidYAML), 0644)
				require.NoError(t, err)
				return configFile
			},
			expectError: true,
			errorMsg:    "failed to parse config YAML",
		},
		{
			name: "nonexistent file",
			setupFunc: func(t *testing.T) string {
				return "/nonexistent/path/config.yaml"
			},
			expectError: true,
			errorMsg:    "failed to read config file",
		},
		{
			name: "missing required metadata version",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configFile := filepath.Join(tmpDir, "config.yaml")

				invalidYAML := `
metadata:
  cli_version: "1.0.0"
  description: "Test config"
global:
  flags: {}
categories: {}
commands:
  test:
    description: "Test command"
    usage: "test [options]"
workflows: {}
profiles: {}
help: {}
`
				err := os.WriteFile(configFile, []byte(invalidYAML), 0644)
				require.NoError(t, err)
				return configFile
			},
			expectError: true,
			errorMsg:    "configuration validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := tt.setupFunc(t)
			loader := NewLoader(configPath)

			config, err := loader.Load()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, config)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, config)
				assert.NotNil(t, config.Metadata)
				assert.NotEmpty(t, config.Metadata.Version)
			}
		})
	}
}

func TestLoader_Load_Caching(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	validYAML := `
metadata:
  version: "1.0.0"
  cli_version: "1.0.0"
  description: "Test config"
global:
  flags: {}
categories: {}
commands:
  test:
    description: "Test command"
    usage: "test [options]"
workflows: {}
profiles: {}
help: {}
`
	err := os.WriteFile(configFile, []byte(validYAML), 0644)
	require.NoError(t, err)

	loader := NewLoader(configFile)

	// First load
	config1, err := loader.Load()
	require.NoError(t, err)
	require.NotNil(t, config1)

	// Second load should return cached result
	config2, err := loader.Load()
	require.NoError(t, err)
	require.NotNil(t, config2)

	// Should be the same instance (cached)
	assert.Same(t, config1, config2)
}

func TestLoader_resolveConfigPath(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T) *Loader
		expectError bool
	}{
		{
			name: "absolute path exists",
			setupFunc: func(t *testing.T) *Loader {
				tmpDir := t.TempDir()
				configFile := filepath.Join(tmpDir, "config.yaml")
				err := os.WriteFile(configFile, []byte("test"), 0644)
				require.NoError(t, err)
				return NewLoader(configFile)
			},
			expectError: false,
		},
		{
			name: "absolute path does not exist",
			setupFunc: func(t *testing.T) *Loader {
				return NewLoader("/nonexistent/config.yaml")
			},
			expectError: false, // resolveConfigPath doesn't check file existence for absolute paths
		},
		{
			name: "relative path in current directory",
			setupFunc: func(t *testing.T) *Loader {
				// Create config in current directory
				configFile := "test-config.yaml"
				err := os.WriteFile(configFile, []byte("test"), 0644)
				require.NoError(t, err)
				t.Cleanup(func() {
					_ = os.Remove(configFile)
				})
				return NewLoader(configFile)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := tt.setupFunc(t)

			path, err := loader.resolveConfigPath()

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, path)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, path)
			}
		})
	}
}
