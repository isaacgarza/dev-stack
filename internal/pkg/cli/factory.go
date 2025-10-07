package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/isaacgarza/dev-stack/internal/core/services"
	"github.com/isaacgarza/dev-stack/internal/pkg/cli/handlers"
	"github.com/isaacgarza/dev-stack/internal/pkg/config"
	"github.com/isaacgarza/dev-stack/internal/pkg/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

// Factory creates CLI commands using the handler registry pattern
type Factory struct {
	config   *config.CommandConfig
	logger   *slog.Logger
	registry *handlers.Registry
}

// NewFactory creates a new CLI factory with handler registry
func NewFactory(cfg *config.CommandConfig) *Factory {
	return &Factory{
		config:   cfg,
		logger:   logger.New(slog.LevelInfo),
		registry: handlers.NewRegistry(),
	}
}

// CreateRootCommand creates the root command with all subcommands using handlers
func (f *Factory) CreateRootCommand() (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use:     "dev-stack",
		Short:   f.config.Metadata.Description,
		Version: f.config.Metadata.CLIVersion,
		Long: fmt.Sprintf(`%s

Version: %s
Build: %s

Use "dev-stack help <command>" for more information about a command.`,
			f.config.Metadata.Description,
			f.config.Metadata.CLIVersion,
			f.config.Metadata.Version),
	}

	// Add global flags
	if err := f.addGlobalFlags(rootCmd.PersistentFlags()); err != nil {
		return nil, fmt.Errorf("failed to add global flags: %w", err)
	}

	// Add commands using handlers
	if err := f.addCommands(rootCmd); err != nil {
		return nil, fmt.Errorf("failed to add commands: %w", err)
	}

	return rootCmd, nil
}

// addGlobalFlags adds global flags to the root command
func (f *Factory) addGlobalFlags(flagSet *pflag.FlagSet) error {
	for name, flag := range f.config.Global.Flags {
		if err := f.addFlag(flagSet, name, flag); err != nil {
			return fmt.Errorf("failed to add global flag %s: %w", name, err)
		}
	}
	return nil
}

// addCommands adds all defined commands to the root command using handlers
func (f *Factory) addCommands(rootCmd *cobra.Command) error {
	for cmdName, cmdConfig := range f.config.Commands {
		cmd, err := f.createCommand(cmdName, cmdConfig)
		if err != nil {
			return fmt.Errorf("failed to create command %s: %w", cmdName, err)
		}
		rootCmd.AddCommand(cmd)
	}
	return nil
}

// createCommand creates a single command from configuration using handlers
func (f *Factory) createCommand(name string, cmdConfig config.Command) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:     name,
		Short:   cmdConfig.Description,
		Long:    f.formatLongDescription(cmdConfig),
		Aliases: cmdConfig.Aliases,
		RunE:    f.createCommandHandler(name),
	}

	// Add command-specific flags
	for flagName, flag := range cmdConfig.Flags {
		if err := f.addFlag(cmd.Flags(), flagName, flag); err != nil {
			return nil, fmt.Errorf("failed to add flag %s: %w", flagName, err)
		}
	}

	// Add completion
	if err := f.addCompletion(cmd, cmdConfig); err != nil {
		return nil, fmt.Errorf("failed to add completion: %w", err)
	}

	return cmd, nil
}

// createCommandHandler creates the actual command execution handler using the registry
func (f *Factory) createCommandHandler(name string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// Create execution context
		ctx := context.Background()

		// Create base command
		base, err := f.createBaseCommand()
		if err != nil {
			return fmt.Errorf("failed to create base command: %w", err)
		}
		defer func() {
			if closeErr := base.Close(); closeErr != nil {
				f.logger.Error("Failed to close base command", "error", closeErr)
			}
		}()

		// Get handler from registry
		handler, err := f.registry.GetHandler(name)
		if err != nil {
			return fmt.Errorf("handler not found for command %s: %w", name, err)
		}

		// Validate arguments using handler
		if err := handler.ValidateArgs(args); err != nil {
			return fmt.Errorf("argument validation failed: %w", err)
		}

		// Execute command using handler
		return handler.Handle(ctx, cmd, args, base)
	}
}

// createBaseCommand creates a base command with common initialization
func (f *Factory) createBaseCommand() (*handlers.BaseCommand, error) {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	// Find project directory
	projectDir := f.findProjectRoot(cwd)

	// Create service manager
	manager, err := services.NewManager(f.logger, projectDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create service manager: %w", err)
	}

	return &handlers.BaseCommand{
		ProjectDir: projectDir,
		Manager:    &serviceManagerAdapter{manager: manager},
		Logger:     &loggerAdapter{logger: f.logger},
	}, nil
}

// findProjectRoot attempts to find the project root directory
func (f *Factory) findProjectRoot(startDir string) string {
	dir := startDir
	for {
		// Check for dev-stack configuration files
		configFiles := []string{
			"dev-stack-config.yaml",
			"dev-stack-config.yml",
			".dev-stack.yaml",
			".dev-stack.yml",
		}

		for _, configFile := range configFiles {
			if _, err := os.Stat(filepath.Join(dir, configFile)); err == nil {
				return dir
			}
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root directory
			break
		}
		dir = parent
	}

	// Return starting directory if no project root found
	return startDir
}

// Adapter implementations to bridge between old and new interfaces

type serviceManagerAdapter struct {
	manager *services.Manager
}

func (s *serviceManagerAdapter) StartServices(ctx context.Context, serviceNames []string, options handlers.StartOptions) error {
	startOpts := services.StartOptions{
		Build:   options.Build,
		Detach:  options.Detach,
		Timeout: time.Duration(options.Timeout) * time.Second,
	}
	return s.manager.StartServices(ctx, serviceNames, startOpts)
}

func (s *serviceManagerAdapter) StopServices(ctx context.Context, serviceNames []string, options handlers.StopOptions) error {
	stopOpts := services.StopOptions{
		Timeout:       options.Timeout,
		RemoveVolumes: options.Volumes,
	}
	return s.manager.StopServices(ctx, serviceNames, stopOpts)
}

func (s *serviceManagerAdapter) GetServiceStatus(ctx context.Context, serviceNames []string) ([]handlers.ServiceStatus, error) {
	statuses, err := s.manager.GetServiceStatus(ctx, serviceNames)
	if err != nil {
		return nil, err
	}

	result := make([]handlers.ServiceStatus, len(statuses))
	for i, status := range statuses {
		// Convert PortMapping to string slice
		ports := make([]string, len(status.Ports))
		for j, port := range status.Ports {
			ports[j] = fmt.Sprintf("%s:%s", port.Host, port.Container)
		}

		result[i] = handlers.ServiceStatus{
			Name:      status.Name,
			State:     status.State,
			Health:    status.Health,
			Ports:     ports,
			CreatedAt: status.CreatedAt,
			UpdatedAt: status.CreatedAt, // Use CreatedAt as UpdatedAt since UpdatedAt doesn't exist
		}
	}
	return result, nil
}

func (s *serviceManagerAdapter) Close() error {
	return s.manager.Close()
}

type loggerAdapter struct {
	logger *slog.Logger
}

func (l *loggerAdapter) Info(msg string, args ...interface{}) {
	l.logger.Info(msg, args...)
}

func (l *loggerAdapter) Error(msg string, args ...interface{}) {
	l.logger.Error(msg, args...)
}

func (l *loggerAdapter) Debug(msg string, args ...interface{}) {
	l.logger.Debug(msg, args...)
}

// Helper methods (simplified versions from original factory)

func (f *Factory) addFlag(flagSet *pflag.FlagSet, name string, flag config.Flag) error {
	description := flag.Description
	if flag.Required {
		description += " (required)"
	}

	switch flag.Type {
	case "string":
		defaultVal := ""
		if flag.Default != nil {
			if str, ok := flag.Default.(string); ok {
				defaultVal = str
			}
		}
		flagSet.String(name, defaultVal, description)
	case "bool":
		defaultVal := false
		if flag.Default != nil {
			if b, ok := flag.Default.(bool); ok {
				defaultVal = b
			}
		}
		flagSet.Bool(name, defaultVal, description)
	case "int":
		defaultVal := 0
		if flag.Default != nil {
			if i, ok := flag.Default.(int); ok {
				defaultVal = i
			}
		}
		flagSet.Int(name, defaultVal, description)
	default:
		return fmt.Errorf("unsupported flag type: %s", flag.Type)
	}

	return nil
}

func (f *Factory) addCompletion(cmd *cobra.Command, cmdConfig config.Command) error {
	// Add service name completion for service-related commands
	if f.isServiceCommand(cmdConfig) {
		cmd.ValidArgsFunction = f.serviceNameCompletion
	}
	return nil
}

// isServiceCommand determines if a command accepts service arguments based on commands.yaml
func (f *Factory) isServiceCommand(cmdConfig config.Command) bool {
	// Check if usage pattern indicates service arguments
	usage := strings.ToLower(cmdConfig.Usage)

	// Look for patterns that indicate service arguments
	servicePatterns := []string{
		"[service",
		"service...",
		"[services",
		"services...",
	}

	for _, pattern := range servicePatterns {
		if strings.Contains(usage, pattern) {
			return true
		}
	}

	return false
}

func (f *Factory) serviceNameCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Load actual service names from services.yaml
	servicesFile := filepath.Join("services", "services.yaml")
	if _, err := os.Stat(servicesFile); os.IsNotExist(err) {
		return []string{}, cobra.ShellCompDirectiveError
	}

	data, err := os.ReadFile(servicesFile)
	if err != nil {
		return []string{}, cobra.ShellCompDirectiveError
	}

	var servicesConfig struct {
		Services map[string]interface{} `yaml:"services"`
	}
	if err := yaml.Unmarshal(data, &servicesConfig); err != nil {
		return []string{}, cobra.ShellCompDirectiveError
	}

	serviceNames := make([]string, 0, len(servicesConfig.Services))
	for name := range servicesConfig.Services {
		serviceNames = append(serviceNames, name)
	}

	return serviceNames, cobra.ShellCompDirectiveNoFileComp
}

func (f *Factory) formatLongDescription(cmdConfig config.Command) string {
	description := cmdConfig.Description
	if len(cmdConfig.Examples) > 0 {
		description += "\n\nExamples:\n"
		for _, example := range cmdConfig.Examples {
			description += fmt.Sprintf("  %s\n", example)
		}
	}
	return description
}
