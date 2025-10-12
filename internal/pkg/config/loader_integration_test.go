//go:build integration

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoader_LoadWithRealFiles_Integration(t *testing.T) {
	t.Run("load config from real file system", func(t *testing.T) {
		// Create a temporary directory structure
		tmpDir := t.TempDir()
		configDir := filepath.Join(tmpDir, "internal", "config")
		err := os.MkdirAll(configDir, 0755)
		require.NoError(t, err)

		// Create a realistic config file
		configContent := `
metadata:
  version: "1.0.0"
  cli_version: "1.0.0"
  description: "Integration test config"
  generated_at: "2024-01-01T00:00:00Z"
global:
  flags:
    verbose:
      type: "bool"
      description: "Enable verbose output"
      default: false
categories:
  core:
    name: "core"
    description: "Core commands"
    weight: 10
  services:
    name: "services"  
    description: "Service management"
    weight: 20
commands:
  init:
    description: "Initialize a new project"
    usage: "init [options]"
    category: "core"
    flags:
      name:
        type: "string"
        description: "Project name"
        required: true
  up:
    description: "Start services"
    usage: "up [services...]"
    category: "services"
    aliases: ["start"]
workflows:
  dev:
    name: "Development"
    description: "Development workflow"
    steps:
      - command: "init"
      - command: "up"
profiles:
  minimal:
    name: "Minimal"
    description: "Minimal configuration"
    services: []
help:
  getting_started: "Run 'dev-stack init' to get started"
`
		configFile := filepath.Join(configDir, "commands.yaml")
		err = os.WriteFile(configFile, []byte(configContent), 0644)
		require.NoError(t, err)

		// Change to the temp directory so the loader can find the config
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		defer os.Chdir(originalDir)

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		// Test loading the config
		loader := NewLoader("")
		config, err := loader.Load()

		require.NoError(t, err)
		require.NotNil(t, config)

		// Verify the loaded configuration
		assert.Equal(t, "1.0.0", config.Metadata.Version)
		assert.Equal(t, "Integration test config", config.Metadata.Description)
		assert.Len(t, config.Categories, 2)
		assert.Len(t, config.Commands, 2)
		assert.Len(t, config.Workflows, 1)
		assert.Len(t, config.Profiles, 1)

		// Verify specific command details
		initCmd, exists := config.Commands["init"]
		assert.True(t, exists)
		assert.Equal(t, "Initialize a new project", initCmd.Description)
		assert.Equal(t, "core", initCmd.Category)

		upCmd, exists := config.Commands["up"]
		assert.True(t, exists)
		assert.Equal(t, "Start services", upCmd.Description)
		assert.Equal(t, "services", upCmd.Category)
		assert.Contains(t, upCmd.Aliases, "start")

		// Test caching works
		config2, err := loader.Load()
		require.NoError(t, err)
		assert.Same(t, config, config2)
	})
}
