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

// UpHandler handles the up command
type UpHandler struct{}

// NewUpHandler creates a new up handler
func NewUpHandler() *UpHandler {
	return &UpHandler{}
}

// Handle executes the up command
func (h *UpHandler) Handle(ctx context.Context, cmd *cobra.Command, args []string, base *types.BaseCommand) error {
	ui.Header("Starting Dev Stack")

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

	// Parse flags
	build, _ := cmd.Flags().GetBool("build")
	forceRecreate, _ := cmd.Flags().GetBool("force-recreate")
	noDeps, _ := cmd.Flags().GetBool("no-deps")

	options := docker.StartOptions{
		Build:         build,
		ForceRecreate: forceRecreate,
		NoDeps:        noDeps,
	}

	// Determine services to start
	serviceNames := args
	if len(serviceNames) == 0 {
		serviceNames = cfg.Stack.Enabled
	}

	// Start services
	if err := dockerClient.Containers().Start(ctx, cfg.Project.Name, serviceNames, options); err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	}

	ui.Success("Dev stack started successfully")
	ui.Info("Run 'dev-stack status' to check service status")
	return nil
}

// ValidateArgs validates the command arguments
func (h *UpHandler) ValidateArgs(args []string) error {
	return nil
}

// GetRequiredFlags returns required flags for this command
func (h *UpHandler) GetRequiredFlags() []string {
	return []string{}
}
