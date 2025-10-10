package docker

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"

	"github.com/isaacgarza/dev-stack/internal/pkg/constants"
	"github.com/isaacgarza/dev-stack/internal/pkg/types"
)

// ContainerLister handles container discovery and status operations
type ContainerLister struct {
	client *Client
}

// NewContainerLister creates a new container lister
func NewContainerLister(client *Client) *ContainerLister {
	return &ContainerLister{
		client: client,
	}
}

// List returns a list of containers matching the given filters
func (cl *ContainerLister) List(ctx context.Context, projectName string, serviceNames []string) ([]types.ServiceStatus, error) {
	filters := filters.NewArgs()

	if projectName != "" {
		filters.Add("label", fmt.Sprintf("%s=%s", constants.ComposeProjectLabel, projectName))
	}

	containers, err := cl.client.cli.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filters,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var services []types.ServiceStatus
	for _, c := range containers {
		serviceName := c.Labels[constants.ComposeServiceLabel]

		if len(serviceNames) > 0 && !contains(serviceNames, serviceName) {
			continue
		}

		status := types.ServiceStatus{
			Name:      serviceName,
			State:     c.State,
			Health:    getHealthStatus(c.Status),
			CreatedAt: time.Unix(c.Created, 0),
		}

		if c.State == constants.StateRunning {
			status.StartedAt = &status.CreatedAt
		}

		if c.State == constants.StateRunning {
			stats, err := cl.getContainerStats(ctx, c.ID)
			if err == nil {
				status.CPUUsage = stats.CPUUsage
				status.Memory = stats.Memory
			}
		}

		for _, port := range c.Ports {
			if port.PublicPort > 0 {
				portMapping := types.PortMapping{
					Host:      fmt.Sprintf("%d", port.PublicPort),
					Container: fmt.Sprintf("%d", port.PrivatePort),
					Protocol:  port.Type,
				}
				status.Ports = append(status.Ports, portMapping)
			}
		}

		status.Labels = c.Labels
		services = append(services, status)
	}

	return services, nil
}

// getContainerStats retrieves container statistics
func (cl *ContainerLister) getContainerStats(ctx context.Context, containerID string) (*ContainerStats, error) {
	stats, err := cl.client.cli.ContainerStats(ctx, containerID, false)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := stats.Body.Close(); closeErr != nil {
			cl.client.logger.Error("Failed to close stats body", "error", closeErr)
		}
	}()

	return &ContainerStats{
		CPUUsage: 0.0,
		Memory: types.MemoryUsage{
			Used:  0,
			Limit: 0,
		},
	}, nil
}
