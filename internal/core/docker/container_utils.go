package docker

import (
	"strings"

	"github.com/isaacgarza/dev-stack/internal/pkg/constants"
	"github.com/isaacgarza/dev-stack/internal/pkg/types"
)

// ContainerStats holds container statistics
type ContainerStats struct {
	CPUUsage float64
	Memory   types.MemoryUsage
}

// getHealthStatus extracts health status from container status string
func getHealthStatus(status string) string {
	if strings.Contains(status, constants.HealthHealthy) {
		return constants.HealthHealthy
	}
	if strings.Contains(status, constants.HealthUnhealthy) {
		return constants.HealthUnhealthy
	}
	if strings.Contains(status, constants.HealthStarting) {
		return constants.HealthStarting
	}
	return constants.HealthNone
}

// contains checks if a slice contains a specific item
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
