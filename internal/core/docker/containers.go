package docker

import (
	"context"

	"github.com/isaacgarza/dev-stack/internal/pkg/types"
)

// ContainerService provides container management operations
type ContainerService struct {
	client    *Client
	lister    *ContainerLister
	lifecycle *ContainerLifecycle
	executor  *ContainerExecutor
}

// NewContainerService creates a new container service
func NewContainerService(client *Client) *ContainerService {
	return &ContainerService{
		client:    client,
		lister:    NewContainerLister(client),
		lifecycle: NewContainerLifecycle(client),
		executor:  NewContainerExecutor(client),
	}
}

// List returns a list of containers matching the given filters
func (cs *ContainerService) List(ctx context.Context, projectName string, serviceNames []string) ([]types.ServiceStatus, error) {
	return cs.lister.List(ctx, projectName, serviceNames)
}

// Start starts containers for the specified services
func (cs *ContainerService) Start(ctx context.Context, projectName string, serviceNames []string, options types.StartOptions) error {
	return cs.lifecycle.Start(ctx, projectName, serviceNames, options)
}

// Stop stops containers for the specified services
func (cs *ContainerService) Stop(ctx context.Context, projectName string, serviceNames []string, options types.StopOptions) error {
	return cs.lifecycle.Stop(ctx, projectName, serviceNames, options)
}

// Exec executes a command in a running container
func (cs *ContainerService) Exec(ctx context.Context, projectName, serviceName string, cmd []string, options types.ExecOptions) error {
	return cs.executor.Exec(ctx, projectName, serviceName, cmd, options)
}

// Logs retrieves logs from containers
func (cs *ContainerService) Logs(ctx context.Context, projectName string, serviceNames []string, options types.LogOptions) error {
	return cs.executor.Logs(ctx, projectName, serviceNames, options)
}
