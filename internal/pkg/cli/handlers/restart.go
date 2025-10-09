package handlers

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// RestartHandler handles the restart command
type RestartHandler struct{}

// NewRestartHandler creates a new restart handler
func NewRestartHandler() *RestartHandler {
	return &RestartHandler{}
}

// Handle executes the restart command
func (h *RestartHandler) Handle(ctx context.Context, cmd *cobra.Command, args []string, base *BaseCommand) error {
	timeout, _ := cmd.Flags().GetInt("timeout")
	noDeps, _ := cmd.Flags().GetBool("no-deps")

	// If no specific services provided, read from config
	servicesToRestart := args
	if len(args) == 0 {
		config, err := h.loadProjectConfig()
		if err != nil {
			return fmt.Errorf("failed to load project configuration: %w", err)
		}
		servicesToRestart = config.Stack.Enabled
	}

	// Resolve dependencies unless --no-deps is specified
	if !noDeps {
		serviceUtils := NewServiceUtils()
		resolvedServices, err := serviceUtils.ResolveDependencies(servicesToRestart)
		if err != nil {
			fmt.Printf("⚠️  Warning: dependency resolution failed: %v\n", err)
		} else {
			servicesToRestart = resolvedServices
		}
	}

	// Validate services
	if len(servicesToRestart) > 0 {
		if err := base.ValidateServices(servicesToRestart); err != nil {
			return err
		}
	}

	// Display what we're restarting
	if len(args) == 0 {
		fmt.Printf("🔄 Restarting services from config: %v\n", servicesToRestart)
	} else {
		fmt.Printf("🔄 Restarting specified services: %v\n", servicesToRestart)
	}

	// Stop services first
	fmt.Println("⏹️  Stopping services...")
	stopOptions := StopOptions{
		Timeout: timeout,
	}
	if err := base.Manager.StopServices(ctx, servicesToRestart, stopOptions); err != nil {
		return fmt.Errorf("failed to stop services: %w", err)
	}

	// Start services
	fmt.Println("▶️  Starting services...")
	startOptions := StartOptions{
		Build:   false,
		Detach:  true,
		Timeout: 60,
	}
	return base.Manager.StartServices(ctx, servicesToRestart, startOptions)
}

// loadProjectConfig loads the dev-stack-config.yaml file
func (h *RestartHandler) loadProjectConfig() (*ProjectConfig, error) {
	configFile := "dev-stack-config.yaml"
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", configFile, err)
	}

	var config ProjectConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", configFile, err)
	}

	return &config, nil
}

// ValidateArgs validates the command arguments
func (h *RestartHandler) ValidateArgs(args []string) error {
	return nil
}

// GetRequiredFlags returns required flags for this command
func (h *RestartHandler) GetRequiredFlags() []string {
	return []string{}
}
