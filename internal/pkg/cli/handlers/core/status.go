package core

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/isaacgarza/dev-stack/internal/core/docker"
	"github.com/isaacgarza/dev-stack/internal/pkg/cli/handlers/utils"
	"github.com/isaacgarza/dev-stack/internal/pkg/cli/types"
	"github.com/isaacgarza/dev-stack/internal/pkg/constants"
	"github.com/isaacgarza/dev-stack/internal/pkg/ui"
	pkgUtils "github.com/isaacgarza/dev-stack/internal/pkg/utils"
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
	// Get CI-friendly flags
	ciFlags := utils.GetCIFlags(cmd)

	if !ciFlags.Quiet {
		ui.Header(constants.MsgStatus)
	}

	// Check if dev-stack is initialized
	configPath := filepath.Join(constants.DevStackDir, constants.ConfigFileName)
	if !pkgUtils.FileExists(configPath) {
		utils.HandleError(ciFlags, errors.New(constants.ErrNotInitialized))
		return nil
	}

	// Load project configuration
	cfg, err := LoadProjectConfig(configPath)
	if err != nil {
		utils.HandleError(ciFlags, fmt.Errorf("failed to load configuration: %w", err))
		return nil
	}

	// Create Docker client
	logger := base.Logger.(loggerAdapter)
	dockerClient, err := docker.NewClient(logger.SlogLogger())
	if err != nil {
		utils.HandleError(ciFlags, fmt.Errorf("failed to create Docker client: %w", err))
		return nil
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
		utils.HandleError(ciFlags, fmt.Errorf("failed to get service status: %w", err))
		return nil
	}

	// Handle CI-friendly output
	utils.OutputResult(ciFlags, map[string]interface{}{
		"services": statuses,
		"count":    len(statuses),
	}, constants.ExitSuccess)

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
