package cli

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/isaacgarza/dev-stack/internal/core/services"
	"github.com/isaacgarza/dev-stack/internal/pkg/config"
	"github.com/isaacgarza/dev-stack/internal/pkg/constants"
	"github.com/isaacgarza/dev-stack/internal/pkg/logger"
	"github.com/spf13/cobra"
)

// BuildRootCommand creates the root command with all subcommands using YAML configuration
func BuildRootCommand(config *config.CommandConfig) (*cobra.Command, error) {
	// Use dynamic builder that reads from commands.yaml
	return BuildDynamicRootCommand(config)
}

// createServiceManager creates and initializes the service manager
func createServiceManager() (*services.Manager, error) {
	projectRoot := findProjectRoot(".")
	log := logger.New(slog.LevelInfo)

	return services.NewManager(log, projectRoot)
}

// findProjectRoot finds the project root directory using constants
func findProjectRoot(startDir string) string {
	configFiles := []string{
		constants.ConfigFileName,
		constants.ConfigFileNameYAML,
		constants.ConfigFileNameHidden,
		constants.ConfigFileNameHiddenYAML,
	}

	dir := startDir
	for {
		absDir, err := filepath.Abs(dir)
		if err != nil {
			break
		}

		for _, configFile := range configFiles {
			configPath := filepath.Join(absDir, configFile)
			if _, err := os.Stat(configPath); err == nil {
				return absDir
			}
		}

		parent := filepath.Dir(absDir)
		if parent == absDir {
			break
		}
		dir = parent
	}

	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}
