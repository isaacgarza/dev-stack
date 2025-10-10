package cli

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/isaacgarza/dev-stack/internal/core/services"
	"github.com/isaacgarza/dev-stack/internal/pkg/config"
	"github.com/isaacgarza/dev-stack/internal/pkg/constants"
	"github.com/isaacgarza/dev-stack/internal/pkg/logger"
	"github.com/spf13/cobra"
)

// BuildRootCommand creates the root command with all subcommands using functional approach
func BuildRootCommand(config *config.CommandConfig) (*cobra.Command, error) {
	log := logger.New(slog.LevelInfo)

	rootCmd := &cobra.Command{
		Use:     "dev-stack",
		Short:   config.Metadata.Description,
		Version: config.Metadata.CLIVersion,
		Long: fmt.Sprintf(`%s

Version: %s

Use "dev-stack help <command>" for more information about a command.`,
			config.Metadata.Description,
			config.Metadata.CLIVersion),
	}

	if err := addGlobalFlags(rootCmd, config); err != nil {
		return nil, fmt.Errorf("failed to add global flags: %w", err)
	}

	serviceManager, err := createServiceManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create service manager: %w", err)
	}

	rootCmd.AddCommand(NewUpCommand(serviceManager, log))
	rootCmd.AddCommand(NewDownCommand(serviceManager, log))
	rootCmd.AddCommand(NewStatusCommand(serviceManager, log))
	rootCmd.AddCommand(NewRestartCommand(serviceManager, log))
	rootCmd.AddCommand(NewLogsCommand(serviceManager, log))
	rootCmd.AddCommand(NewExecCommand(serviceManager, log))
	rootCmd.AddCommand(NewInitCommand(log))
	rootCmd.AddCommand(NewConfigCommand(log))
	rootCmd.AddCommand(NewServicesCommand(serviceManager, log))

	return rootCmd, nil
}

// addGlobalFlags adds global flags to the root command
func addGlobalFlags(cmd *cobra.Command, config *config.CommandConfig) error {
	for name, flag := range config.Global.Flags {
		switch flag.Type {
		case "bool":
			cmd.PersistentFlags().Bool(name, false, flag.Description)
		case "string":
			cmd.PersistentFlags().String(name, "", flag.Description)
		case "int":
			cmd.PersistentFlags().Int(name, 0, flag.Description)
		}

		if flag.Short != "" {
			if pf := cmd.PersistentFlags().Lookup(name); pf != nil {
				pf.Shorthand = flag.Short
			}
		}
	}
	return nil
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
