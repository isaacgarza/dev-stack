package core

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/isaacgarza/dev-stack/internal/core/docker"
	"github.com/isaacgarza/dev-stack/internal/pkg/cli/types"
	"github.com/isaacgarza/dev-stack/internal/pkg/constants"
	"github.com/isaacgarza/dev-stack/internal/pkg/ui"
	"github.com/isaacgarza/dev-stack/internal/pkg/utils"
	"github.com/spf13/cobra"
)

// StatusHandler handles the status command
type StatusHandler struct{}

// NewStatusHandler creates a new status handler
func NewStatusHandler() *StatusHandler {
	return &StatusHandler{}
}

// Handle executes the status command
func (h *StatusHandler) Handle(ctx context.Context, cmd *cobra.Command, args []string, base *types.BaseCommand) error {
	ui.Header("Dev Stack Status")

	// Check if dev-stack is initialized
	configPath := filepath.Join(constants.DevStackDir, constants.ConfigFileName)
	if !utils.FileExists(configPath) {
		return fmt.Errorf("dev-stack not initialized. Run 'dev-stack init' first")
	}

	// Load project configuration
	cfg, err := LoadProjectConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create Docker client
	logger := base.Logger.(loggerAdapter)
	dockerClient, err := docker.NewClient(logger.SlogLogger())
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer func() {
		if err := dockerClient.Close(); err != nil {
			base.Logger.Error("Failed to close Docker client", "error", err)
		}
	}()

	// Determine services to check
	serviceNames := args
	if len(serviceNames) == 0 {
		serviceNames = cfg.Stack.Enabled
	}

	// Get service status
	statuses, err := dockerClient.Containers().List(ctx, cfg.Project.Name, serviceNames)
	if err != nil {
		return fmt.Errorf("failed to get service status: %w", err)
	}

	// Display status
	if len(statuses) == 0 {
		ui.Info("No services found")
		return nil
	}

	ui.Info("Service Status:")
	for _, status := range statuses {
		statusIcon := "🔴"
		if status.State == "running" {
			statusIcon = "🟢"
		}
		ui.Info("  %s %s: %s", statusIcon, status.Name, status.State)
		if status.Health != "" {
			ui.Info("    Health: %s", status.Health)
		}
	}

	return nil
}

// ValidateArgs validates the command arguments
func (h *StatusHandler) ValidateArgs(args []string) error {
	return nil
}

// GetRequiredFlags returns required flags for this command
func (h *StatusHandler) GetRequiredFlags() []string {
	return []string{}
}
