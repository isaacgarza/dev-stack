package core

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/isaacgarza/dev-stack/internal/core/docker"
	cliTypes "github.com/isaacgarza/dev-stack/internal/pkg/cli/types"
	"github.com/isaacgarza/dev-stack/internal/pkg/constants"
	"github.com/isaacgarza/dev-stack/internal/pkg/types"
	"github.com/isaacgarza/dev-stack/internal/pkg/ui"
	"github.com/isaacgarza/dev-stack/internal/pkg/utils"
	"github.com/spf13/cobra"
)

// DownHandler handles the down command
type DownHandler struct{}

// NewDownHandler creates a new down handler
func NewDownHandler() *DownHandler {
	return &DownHandler{}
}

// Handle executes the down command
func (h *DownHandler) Handle(ctx context.Context, cmd *cobra.Command, args []string, base *cliTypes.BaseCommand) error {
	ui.Header(constants.MsgStopping)

	// Check if dev-stack is initialized
	configPath := filepath.Join(constants.DevStackDir, constants.ConfigFileName)
	if !utils.FileExists(configPath) {
		return errors.New(constants.ErrNotInitialized)
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
	timeout, _ := cmd.Flags().GetInt("timeout")

	options := types.StopOptions{
		Timeout: timeout,
		Remove:  true,
	}

	// Determine services to stop
	serviceNames := args
	if len(serviceNames) == 0 {
		serviceNames = cfg.Stack.Enabled
	}

	// Stop services
	if err := dockerClient.Containers().Stop(ctx, cfg.Project.Name, serviceNames, options); err != nil {
		return fmt.Errorf("failed to stop services: %w", err)
	}

	ui.Success(constants.MsgStopSuccess)
	return nil
}

// ValidateArgs validates the command arguments
func (h *DownHandler) ValidateArgs(args []string) error {
	return nil
}

// GetRequiredFlags returns required flags for this command
func (h *DownHandler) GetRequiredFlags() []string {
	return []string{}
}
