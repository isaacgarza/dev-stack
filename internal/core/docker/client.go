package docker

import (
	"fmt"
	"log/slog"

	"github.com/docker/docker/client"
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

// Containers returns a service for container operations
func (c *Client) Containers() *ContainerService {
	return NewContainerService(c)
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
