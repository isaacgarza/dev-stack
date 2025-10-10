package docker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"

	"github.com/isaacgarza/dev-stack/internal/pkg/constants"
	"github.com/isaacgarza/dev-stack/internal/pkg/types"
)

// ContainerLifecycle handles container start/stop operations
type ContainerLifecycle struct {
	client *Client
}

// NewContainerLifecycle creates a new container lifecycle manager
func NewContainerLifecycle(client *Client) *ContainerLifecycle {
	return &ContainerLifecycle{
		client: client,
	}
}

// Start starts containers for the specified services
func (cl *ContainerLifecycle) Start(ctx context.Context, projectName string, serviceNames []string, options types.StartOptions) error {
	cl.client.logger.Info("Starting services", "project", projectName, "services", serviceNames)

	args := []string{"compose", "-f", constants.DockerComposeFile, "-p", projectName, "up", "-d"}

	if options.Build {
		args = append(args, "--build")
	}

	if options.ForceRecreate {
		args = append(args, "--force-recreate")
	}

	args = append(args, serviceNames...)

	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		cl.client.logger.Error("Failed to start services", "error", err, "output", string(output))

		if len(output) > 0 {
			fmt.Printf("\n🔍 Docker output:\n%s\n", string(output))
		}

		if saveErr := cl.saveErrorLogs(string(output)); saveErr != nil {
			cl.client.logger.Error("Failed to save error logs", "error", saveErr)
		}

		return fmt.Errorf("failed to start services: %w", err)
	}

	cl.client.logger.Info("Services started successfully", "services", serviceNames)
	return nil
}

// Stop stops containers for the specified services
func (cl *ContainerLifecycle) Stop(ctx context.Context, projectName string, serviceNames []string, options types.StopOptions) error {
	cl.client.logger.Info("Stopping services", "project", projectName, "services", serviceNames)

	filters := filters.NewArgs()
	filters.Add("label", fmt.Sprintf("%s=%s", constants.ComposeProjectLabel, projectName))

	containers, err := cl.client.cli.ContainerList(ctx, container.ListOptions{
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

		if c.State == constants.StateRunning {
			timeoutSecs := options.Timeout
			if err := cl.client.cli.ContainerStop(ctx, c.ID, container.StopOptions{
				Timeout: &timeoutSecs,
			}); err != nil {
				cl.client.logger.Error("Failed to stop container", "container", c.ID, "service", serviceName, "error", err)
				continue
			}
			cl.client.logger.Info("Stopped container", "container", c.ID[:12], "service", serviceName)
		}

		if options.Remove {
			if err := cl.client.cli.ContainerRemove(ctx, c.ID, container.RemoveOptions{
				RemoveVolumes: options.RemoveVolumes,
				Force:         true,
			}); err != nil {
				cl.client.logger.Error("Failed to remove container", "container", c.ID, "service", serviceName, "error", err)
				continue
			}
			cl.client.logger.Info("Removed container", "container", c.ID[:12], "service", serviceName)
		}
	}

	return nil
}

// saveErrorLogs saves error output to a log file
func (cl *ContainerLifecycle) saveErrorLogs(output string) error {
	logsDir := fmt.Sprintf("%s/%s", constants.DevStackDir, constants.LogsDir)
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return fmt.Errorf("failed to create logs directory: %w", err)
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logFile := fmt.Sprintf("%s/docker-error-%s.log", logsDir, timestamp)

	content := fmt.Sprintf("Docker Error Log - %s\n%s\n%s\n\n%s",
		time.Now().Format(time.RFC3339),
		strings.Repeat("=", 50),
		"Docker Compose Output:",
		output)

	if err := os.WriteFile(logFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write error log: %w", err)
	}

	fmt.Printf("📝 Error details saved to: %s\n", logFile)
	return nil
}
