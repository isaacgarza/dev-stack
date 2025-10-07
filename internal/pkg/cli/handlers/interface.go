package handlers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// CommandHandler defines the interface for all command handlers
type CommandHandler interface {
	// Handle executes the command with the given context, command, arguments, and base command
	Handle(ctx context.Context, cmd *cobra.Command, args []string, base *BaseCommand) error

	// ValidateArgs validates the command arguments before execution
	ValidateArgs(args []string) error

	// GetRequiredFlags returns a list of required flags for this command
	GetRequiredFlags() []string
}

// BaseCommand provides common functionality for all commands
type BaseCommand struct {
	ProjectDir string
	Manager    ServiceManager
	Logger     Logger
}

// Close cleans up resources
func (b *BaseCommand) Close() error {
	if b.Manager != nil {
		return b.Manager.Close()
	}
	return nil
}

// ServiceManager interface for service operations
type ServiceManager interface {
	StartServices(ctx context.Context, serviceNames []string, options StartOptions) error
	StopServices(ctx context.Context, serviceNames []string, options StopOptions) error
	GetServiceStatus(ctx context.Context, serviceNames []string) ([]ServiceStatus, error)
	Close() error
}

// Logger interface for logging operations
type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
}

// Common option types
type StartOptions struct {
	Detach  bool
	Build   bool
	Timeout int
}

type StopOptions struct {
	Timeout int
	Volumes bool
}

type ServiceStatus struct {
	Name      string
	State     string
	Health    string
	Ports     []string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ValidateServices validates service names against available services
func (b *BaseCommand) ValidateServices(serviceNames []string) error {
	// Load services.yaml to get available services
	servicesFile := filepath.Join(b.ProjectDir, "services", "services.yaml")
	if _, err := os.Stat(servicesFile); os.IsNotExist(err) {
		return fmt.Errorf("services.yaml not found at %s. Please ensure you're in a dev-stack project directory", servicesFile)
	}

	data, err := os.ReadFile(servicesFile)
	if err != nil {
		return fmt.Errorf("failed to read services.yaml: %w", err)
	}

	var servicesConfig struct {
		Services map[string]interface{} `yaml:"services"`
	}
	if err := yaml.Unmarshal(data, &servicesConfig); err != nil {
		return fmt.Errorf("failed to parse services.yaml: %w", err)
	}

	// Check each service name
	for _, serviceName := range serviceNames {
		if _, exists := servicesConfig.Services[serviceName]; !exists {
			availableServices := make([]string, 0, len(servicesConfig.Services))
			for name := range servicesConfig.Services {
				availableServices = append(availableServices, name)
			}
			return fmt.Errorf("unknown service '%s'. Available services: %v", serviceName, availableServices)
		}
	}

	return nil
}
