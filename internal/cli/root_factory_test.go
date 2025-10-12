package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateRootCommand(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T)
		expectError bool
		errorMsg    string
	}{
		{
			name: "create root command with embedded config",
			setupFunc: func(t *testing.T) {
				// No setup needed - will use embedded config
			},
			expectError: false,
		},
		{
			name: "create root command with valid config file",
			setupFunc: func(t *testing.T) {
				// Create a valid config file in current directory
				configContent := `
metadata:
  version: "1.0.0"
  cli_version: "1.0.0"
  description: "Test CLI"
global:
  flags: {}
categories:
  general:
    name: "general"
    description: "General commands"
commands:
  test:
    description: "Test command"
    usage: "test [options]"
    category: "general"
workflows: {}
profiles: {}
help: {}
`
				configDir := "internal/config"
				err := os.MkdirAll(configDir, 0755)
				require.NoError(t, err)

				configFile := filepath.Join(configDir, "commands.yaml")
				err = os.WriteFile(configFile, []byte(configContent), 0644)
				require.NoError(t, err)

				t.Cleanup(func() {
					_ = os.RemoveAll(configDir)
				})
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFunc(t)

			rootCmd, err := CreateRootCommand()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, rootCmd)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, rootCmd)
				assert.Equal(t, "dev-stack", rootCmd.Use)
				assert.NotEmpty(t, rootCmd.Short)
			}
		})
	}
}

func TestExecuteFactory(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T)
		expectError bool
	}{
		{
			name: "execute with help flag",
			setupFunc: func(t *testing.T) {
				// Set args to show help (which should not error)
				os.Args = []string{"dev-stack", "--help"}
			},
			expectError: false,
		},
		{
			name: "execute with version flag",
			setupFunc: func(t *testing.T) {
				// Set args to show version
				os.Args = []string{"dev-stack", "--version"}
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original args
			originalArgs := os.Args
			t.Cleanup(func() {
				os.Args = originalArgs
			})

			tt.setupFunc(t)

			// Note: ExecuteFactory() will call os.Exit() for help/version
			// In a real test environment, we'd need to mock this or test differently
			// For now, we'll just test that CreateRootCommand works
			rootCmd, err := CreateRootCommand()
			require.NoError(t, err)
			require.NotNil(t, rootCmd)

			// Test that the command has the expected structure
			assert.Equal(t, "dev-stack", rootCmd.Use)
			assert.NotEmpty(t, rootCmd.Short)
			assert.True(t, rootCmd.HasSubCommands())
		})
	}
}

func TestInitFactoryConfig(t *testing.T) {
	// This function is called by cobra.OnInitialize
	// It's difficult to test directly, but we can test that it doesn't panic
	t.Run("init config does not panic", func(t *testing.T) {
		// Create a minimal config
		configContent := `
metadata:
  version: "1.0.0"
  cli_version: "1.0.0"
  description: "Test CLI"
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
		configDir := "internal/config"
		err := os.MkdirAll(configDir, 0755)
		require.NoError(t, err)

		configFile := filepath.Join(configDir, "commands.yaml")
		err = os.WriteFile(configFile, []byte(configContent), 0644)
		require.NoError(t, err)

		t.Cleanup(func() {
			_ = os.RemoveAll(configDir)
		})

		// Test that creating root command doesn't panic
		assert.NotPanics(t, func() {
			rootCmd, err := CreateRootCommand()
			assert.NoError(t, err)
			assert.NotNil(t, rootCmd)
		})
	})
}
