package docker

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/pkg/stdcopy"

	"github.com/isaacgarza/dev-stack/internal/pkg/constants"
	"github.com/isaacgarza/dev-stack/internal/pkg/types"
)

// ContainerExecutor handles container execution and logs operations
type ContainerExecutor struct {
	client *Client
}

// NewContainerExecutor creates a new container executor
func NewContainerExecutor(client *Client) *ContainerExecutor {
	return &ContainerExecutor{
		client: client,
	}
}

// Exec executes a command in a running container
func (ce *ContainerExecutor) Exec(ctx context.Context, projectName, serviceName string, cmd []string, options types.ExecOptions) error {
	containerID, err := ce.findServiceContainer(ctx, projectName, serviceName)
	if err != nil {
		return err
	}

	config := container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          options.TTY,
	}

	if options.Interactive {
		config.AttachStdin = true
	}

	if options.User != "" {
		config.User = options.User
	}

	if options.WorkingDir != "" {
		config.WorkingDir = options.WorkingDir
	}

	if len(options.Env) > 0 {
		config.Env = options.Env
	}

	exec, err := ce.client.cli.ContainerExecCreate(ctx, containerID, config)
	if err != nil {
		return fmt.Errorf("failed to create exec instance: %w", err)
	}

	resp, err := ce.client.cli.ContainerExecAttach(ctx, exec.ID, container.ExecAttachOptions{
		Tty: options.TTY,
	})
	if err != nil {
		return fmt.Errorf("failed to attach to exec instance: %w", err)
	}
	defer resp.Close()

	if options.TTY {
		if _, err := io.Copy(os.Stdout, resp.Reader); err != nil {
			ce.client.logger.Error("Failed to copy output", "error", err)
		}
	} else {
		if _, err := stdcopy.StdCopy(os.Stdout, os.Stderr, resp.Reader); err != nil {
			ce.client.logger.Error("Failed to copy output", "error", err)
		}
	}

	return nil
}

// Logs retrieves logs from containers
func (ce *ContainerExecutor) Logs(ctx context.Context, projectName string, serviceNames []string, options types.LogOptions) error {
	filters := filters.NewArgs()
	filters.Add("label", fmt.Sprintf("%s=%s", constants.ComposeProjectLabel, projectName))

	containers, err := ce.client.cli.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filters,
	})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	for _, c := range containers {
		serviceName := c.Labels[constants.ComposeServiceLabel]

		if len(serviceNames) > 0 && !contains(serviceNames, serviceName) {
			continue
		}

		logOptions := container.LogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     options.Follow,
			Timestamps: options.Timestamps,
		}

		if options.Since != "" {
			logOptions.Since = options.Since
		}

		if options.Tail != "" {
			logOptions.Tail = options.Tail
		}

		logs, err := ce.client.cli.ContainerLogs(ctx, c.ID, logOptions)
		if err != nil {
			ce.client.logger.Error("Failed to get logs", "container", c.ID, "service", serviceName, "error", err)
			continue
		}

		go func(serviceName string, logs io.ReadCloser) {
			defer func() {
				if closeErr := logs.Close(); closeErr != nil {
					ce.client.logger.Error("Failed to close logs", "error", closeErr)
				}
			}()
			if options.Follow {
				fmt.Printf("==> Following logs for %s <==\n", serviceName)
			}
			if _, err := stdcopy.StdCopy(os.Stdout, os.Stderr, logs); err != nil {
				ce.client.logger.Error("Failed to copy logs", "error", err)
			}
		}(serviceName, logs)
	}

	if options.Follow {
		select {}
	}

	return nil
}

// findServiceContainer finds a running container for a specific service
func (ce *ContainerExecutor) findServiceContainer(ctx context.Context, projectName, serviceName string) (string, error) {
	filters := filters.NewArgs()
	filters.Add("label", fmt.Sprintf("%s=%s", constants.ComposeProjectLabel, projectName))
	filters.Add("label", fmt.Sprintf("%s=%s", constants.ComposeServiceLabel, serviceName))

	containers, err := ce.client.cli.ContainerList(ctx, container.ListOptions{
		All:     false,
		Filters: filters,
	})
	if err != nil {
		return "", fmt.Errorf("failed to list containers: %w", err)
	}

	if len(containers) == 0 {
		return "", fmt.Errorf("no running container found for service %s", serviceName)
	}

	if len(containers) > 1 {
		return "", fmt.Errorf("multiple containers found for service %s", serviceName)
	}

	return containers[0].ID, nil
}
