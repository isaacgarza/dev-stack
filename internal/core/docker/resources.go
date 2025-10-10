package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"

	"github.com/isaacgarza/dev-stack/internal/pkg/constants"
)

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

// Volume operations

// List returns a list of volumes for the project
func (vs *VolumeService) List(ctx context.Context, projectName string) ([]string, error) {
	filters := filters.NewArgs()
	filters.Add("label", fmt.Sprintf("%s=%s", constants.ComposeProjectLabel, projectName))

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

// Network operations

// List returns a list of networks for the project
func (ns *NetworkService) List(ctx context.Context, projectName string) ([]string, error) {
	filters := filters.NewArgs()
	filters.Add("label", fmt.Sprintf("%s=%s", constants.ComposeProjectLabel, projectName))

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

// Image operations

// List returns a list of images for the project
func (is *ImageService) List(ctx context.Context, projectName string) ([]string, error) {
	filters := filters.NewArgs()
	filters.Add("label", fmt.Sprintf("%s=%s", constants.ComposeProjectLabel, projectName))

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
