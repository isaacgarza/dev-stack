package services

import (
	"context"
	"fmt"

	"github.com/isaacgarza/dev-stack/internal/pkg/types"
)

// CleanupManager handles resource cleanup operations
type CleanupManager struct {
	manager *Manager
}

// NewCleanupManager creates a new cleanup manager
func NewCleanupManager(manager *Manager) *CleanupManager {
	return &CleanupManager{manager: manager}
}

// CleanupResources removes project resources
func (cm *CleanupManager) CleanupResources(ctx context.Context, options types.CleanupOptions) error {
	cm.manager.logger.Info("Cleaning up resources", "volumes", options.RemoveVolumes, "images", options.RemoveImages)

	projectName := cm.manager.getProjectName()

	// Stop and remove containers
	if err := cm.manager.docker.Containers().Stop(ctx, projectName, []string{}, types.StopOptions{
		Remove:        true,
		RemoveVolumes: options.RemoveVolumes,
	}); err != nil {
		return fmt.Errorf("failed to remove containers: %w", err)
	}

	// Remove volumes if requested
	if options.RemoveVolumes {
		if err := cm.manager.docker.Volumes().Remove(ctx, projectName); err != nil {
			cm.manager.logger.Error("Failed to remove volumes", "error", err)
		}
	}

	// Remove images if requested
	if options.RemoveImages {
		if err := cm.manager.docker.Images().Remove(ctx, projectName); err != nil {
			cm.manager.logger.Error("Failed to remove images", "error", err)
		}
	}

	// Remove networks if requested
	if options.RemoveNetworks {
		if err := cm.manager.docker.Networks().Remove(ctx, projectName); err != nil {
			cm.manager.logger.Error("Failed to remove networks", "error", err)
		}
	}

	cm.manager.logger.Info("Cleanup completed successfully")
	return nil
}
