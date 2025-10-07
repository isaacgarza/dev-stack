package docker

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"

	"github.com/isaacgarza/dev-stack/internal/pkg/types"
)

// Container state constants
const (
	StateRunning = "running"
)

// Client represents a Docker client with additional functionality for dev-stack
type Client struct {
	cli    *client.Client
	logger *slog.Logger
}

// NewClient creates a new Docker client instance
func NewClient(logger *slog.Logger) (*Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	return &Client{
		cli:    cli,
		logger: logger,
	}, nil
}

// Close closes the Docker client connection
func (c *Client) Close() error {
	return c.cli.Close()
}

// ContainerService provides container management operations
type ContainerService struct {
	client *Client
}

// VolumeService provides volume management operations
type VolumeService struct {
	client *Client
}

// NetworkService provides network management operations
type NetworkService struct {
	client *Client
}

// ImageService provides image management operations
type ImageService struct {
	client *Client
}

// Containers returns a service for container operations
func (c *Client) Containers() *ContainerService {
	return &ContainerService{client: c}
}

// Volumes returns a service for volume operations
func (c *Client) Volumes() *VolumeService {
	return &VolumeService{client: c}
}

// Networks returns a service for network operations
func (c *Client) Networks() *NetworkService {
	return &NetworkService{client: c}
}

// Images returns a service for image operations
func (c *Client) Images() *ImageService {
	return &ImageService{client: c}
}

// Container management operations

// List returns a list of containers matching the given filters
func (cs *ContainerService) List(ctx context.Context, projectName string, serviceNames []string) ([]types.ServiceStatus, error) {
	filters := filters.NewArgs()

	if projectName != "" {
		filters.Add("label", fmt.Sprintf("com.docker.compose.project=%s", projectName))
	}

	containers, err := cs.client.cli.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filters,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var services []types.ServiceStatus
	for _, c := range containers {
		serviceName := c.Labels["com.docker.compose.service"]

		// Filter by service names if specified
		if len(serviceNames) > 0 && !contains(serviceNames, serviceName) {
			continue
		}

		status := types.ServiceStatus{
			Name:      serviceName,
			State:     c.State,
			Health:    getHealthStatus(c.Status),
			CreatedAt: time.Unix(c.Created, 0),
		}

		if c.State == StateRunning {
			status.StartedAt = &status.CreatedAt
		}

		// Get container stats for running containers
		if c.State == StateRunning {
			stats, err := cs.getContainerStats(ctx, c.ID)
			if err == nil {
				status.CPUUsage = stats.CPUUsage
				status.Memory = stats.Memory
			}
		}

		// Map port bindings
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

// Start starts containers for the specified services
func (cs *ContainerService) Start(ctx context.Context, projectName string, serviceNames []string, options StartOptions) error {
	cs.client.logger.Info("Starting services", "project", projectName, "services", serviceNames)

	// Build docker-compose command
	args := []string{"compose", "-p", projectName, "up", "-d"}

	if options.Build {
		args = append(args, "--build")
	}

	if options.ForceRecreate {
		args = append(args, "--force-recreate")
	}

	if options.NoDeps {
		args = append(args, "--no-deps")
	}

	// Add specific services if provided
	args = append(args, serviceNames...)

	// Execute docker-compose command
	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		cs.client.logger.Error("Failed to start services", "error", err, "output", string(output))
		return fmt.Errorf("failed to start services: %w", err)
	}

	cs.client.logger.Info("Services started successfully", "services", serviceNames)
	return nil
}

// Stop stops containers for the specified services
func (cs *ContainerService) Stop(ctx context.Context, projectName string, serviceNames []string, options StopOptions) error {
	cs.client.logger.Info("Stopping services", "project", projectName, "services", serviceNames)

	filters := filters.NewArgs()
	filters.Add("label", fmt.Sprintf("com.docker.compose.project=%s", projectName))

	containers, err := cs.client.cli.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filters,
	})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	for _, c := range containers {
		serviceName := c.Labels["com.docker.compose.service"]

		// Filter by service names if specified
		if len(serviceNames) > 0 && !contains(serviceNames, serviceName) {
			continue
		}

		if c.State == StateRunning {
			timeoutSecs := options.Timeout
			if err := cs.client.cli.ContainerStop(ctx, c.ID, container.StopOptions{
				Timeout: &timeoutSecs,
			}); err != nil {
				cs.client.logger.Error("Failed to stop container", "container", c.ID, "service", serviceName, "error", err)
				continue
			}
			cs.client.logger.Info("Stopped container", "container", c.ID[:12], "service", serviceName)
		}

		if options.Remove {
			if err := cs.client.cli.ContainerRemove(ctx, c.ID, container.RemoveOptions{
				RemoveVolumes: options.RemoveVolumes,
				Force:         true,
			}); err != nil {
				cs.client.logger.Error("Failed to remove container", "container", c.ID, "service", serviceName, "error", err)
				continue
			}
			cs.client.logger.Info("Removed container", "container", c.ID[:12], "service", serviceName)
		}
	}

	return nil
}

// Exec executes a command in a running container
func (cs *ContainerService) Exec(ctx context.Context, projectName, serviceName string, cmd []string, options ExecOptions) error {
	containerID, err := cs.findServiceContainer(ctx, projectName, serviceName)
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

	exec, err := cs.client.cli.ContainerExecCreate(ctx, containerID, config)
	if err != nil {
		return fmt.Errorf("failed to create exec instance: %w", err)
	}

	resp, err := cs.client.cli.ContainerExecAttach(ctx, exec.ID, container.ExecAttachOptions{
		Tty: options.TTY,
	})
	if err != nil {
		return fmt.Errorf("failed to attach to exec instance: %w", err)
	}
	defer resp.Close()

	// Handle output
	if options.TTY {
		if _, err := io.Copy(os.Stdout, resp.Reader); err != nil {
			cs.client.logger.Error("Failed to copy output", "error", err)
		}
	} else {
		if _, err := stdcopy.StdCopy(os.Stdout, os.Stderr, resp.Reader); err != nil {
			cs.client.logger.Error("Failed to copy output", "error", err)
		}
	}

	return nil
}

// Logs retrieves logs from containers
func (cs *ContainerService) Logs(ctx context.Context, projectName string, serviceNames []string, options LogOptions) error {
	filters := filters.NewArgs()
	filters.Add("label", fmt.Sprintf("com.docker.compose.project=%s", projectName))

	containers, err := cs.client.cli.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filters,
	})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	for _, c := range containers {
		serviceName := c.Labels["com.docker.compose.service"]

		// Filter by service names if specified
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

		logs, err := cs.client.cli.ContainerLogs(ctx, c.ID, logOptions)
		if err != nil {
			cs.client.logger.Error("Failed to get logs", "container", c.ID, "service", serviceName, "error", err)
			continue
		}

		// Copy logs to stdout/stderr
		go func(serviceName string, logs io.ReadCloser) {
			defer func() {
				if closeErr := logs.Close(); closeErr != nil {
					cs.client.logger.Error("Failed to close logs", "error", closeErr)
				}
			}()
			if options.Follow {
				fmt.Printf("==> Following logs for %s <==\n", serviceName)
			}
			if _, err := stdcopy.StdCopy(os.Stdout, os.Stderr, logs); err != nil {
				cs.client.logger.Error("Failed to copy logs", "error", err)
			}
		}(serviceName, logs)
	}

	// Block indefinitely when following logs
	if options.Follow {
		select {}
	}

	return nil
}

// Volume management operations

// List returns a list of volumes for the project
func (vs *VolumeService) List(ctx context.Context, projectName string) ([]string, error) {
	filters := filters.NewArgs()
	filters.Add("label", fmt.Sprintf("com.docker.compose.project=%s", projectName))

	volumes, err := vs.client.cli.VolumeList(ctx, volume.ListOptions{
		Filters: filters,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list volumes: %w", err)
	}

	var volumeNames []string
	for _, v := range volumes.Volumes {
		volumeNames = append(volumeNames, v.Name)
	}

	return volumeNames, nil
}

// Remove removes volumes for the project
func (vs *VolumeService) Remove(ctx context.Context, projectName string) error {
	// Get all project volumes
	volumeNames, err := vs.List(ctx, projectName)
	if err != nil {
		return err
	}
	for _, volumeName := range volumeNames {
		if err := vs.client.cli.VolumeRemove(ctx, volumeName, false); err != nil {
			vs.client.logger.Error("Failed to remove volume", "volume", volumeName, "error", err)
			continue
		}
		vs.client.logger.Info("Removed volume", "volume", volumeName)
	}

	return nil
}

// Network management operations

// List returns a list of networks for the project
func (ns *NetworkService) List(ctx context.Context, projectName string) ([]string, error) {
	filters := filters.NewArgs()
	filters.Add("label", fmt.Sprintf("com.docker.compose.project=%s", projectName))

	networks, err := ns.client.cli.NetworkList(ctx, network.ListOptions{
		Filters: filters,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list networks: %w", err)
	}

	var networkNames []string
	for _, n := range networks {
		networkNames = append(networkNames, n.Name)
	}

	return networkNames, nil
}

// Remove removes networks for the project
func (ns *NetworkService) Remove(ctx context.Context, projectName string) error {
	// Get all project networks
	networkNames, err := ns.List(ctx, projectName)
	if err != nil {
		return err
	}
	for _, networkName := range networkNames {
		if err := ns.client.cli.NetworkRemove(ctx, networkName); err != nil {
			ns.client.logger.Error("Failed to remove network", "network", networkName, "error", err)
			continue
		}
		ns.client.logger.Info("Removed network", "network", networkName)
	}

	return nil
}

// Image management operations

// List returns a list of images for the project
func (is *ImageService) List(ctx context.Context, projectName string) ([]string, error) {
	filters := filters.NewArgs()
	filters.Add("label", fmt.Sprintf("com.docker.compose.project=%s", projectName))

	images, err := is.client.cli.ImageList(ctx, image.ListOptions{
		Filters: filters,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}

	var imageNames []string
	for _, img := range images {
		imageNames = append(imageNames, img.RepoTags...)
	}

	return imageNames, nil
}

// Remove removes images for the project
func (is *ImageService) Remove(ctx context.Context, projectName string) error {
	// Get all project images
	imageNames, err := is.List(ctx, projectName)
	if err != nil {
		return err
	}
	for _, imageName := range imageNames {
		if _, err := is.client.cli.ImageRemove(ctx, imageName, image.RemoveOptions{
			Force: true,
		}); err != nil {
			is.client.logger.Error("Failed to remove image", "image", imageName, "error", err)
			continue
		}
		is.client.logger.Info("Removed image", "image", imageName)
	}

	return nil
}

// Helper functions

func (cs *ContainerService) findServiceContainer(ctx context.Context, projectName, serviceName string) (string, error) {
	filters := filters.NewArgs()
	filters.Add("label", fmt.Sprintf("com.docker.compose.project=%s", projectName))
	filters.Add("label", fmt.Sprintf("com.docker.compose.service=%s", serviceName))

	// Only running containers
	containers, err := cs.client.cli.ContainerList(ctx, container.ListOptions{
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

func (cs *ContainerService) getContainerStats(ctx context.Context, containerID string) (*ContainerStats, error) {
	stats, err := cs.client.cli.ContainerStats(ctx, containerID, false)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := stats.Body.Close(); closeErr != nil {
			cs.client.logger.Error("Failed to close stats body", "error", closeErr)
		}
	}()

	// Parse stats response and calculate CPU/memory usage
	// This is a basic implementation that would need enhancement for production use
	return &ContainerStats{
		CPUUsage: 0.0,
		Memory: types.MemoryUsage{
			Used:  0,
			Limit: 0,
		},
	}, nil
}

func getHealthStatus(status string) string {
	if strings.Contains(status, "healthy") {
		return "healthy"
	}
	if strings.Contains(status, "unhealthy") {
		return "unhealthy"
	}
	if strings.Contains(status, "starting") {
		return "starting"
	}
	return "none"
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Option types for various operations

type StartOptions struct {
	Build         bool
	ForceRecreate bool
	NoDeps        bool
	Detach        bool
}

type StopOptions struct {
	Timeout       int
	Remove        bool
	RemoveVolumes bool
}

type ExecOptions struct {
	User        string
	WorkingDir  string
	Env         []string
	Interactive bool
	TTY         bool
	Detach      bool
}

type LogOptions struct {
	Follow     bool
	Timestamps bool
	Tail       string
	Since      string
}

type ContainerStats struct {
	CPUUsage float64
	Memory   types.MemoryUsage
}
