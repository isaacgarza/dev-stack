package services

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/isaacgarza/dev-stack/internal/core/docker"
	"github.com/isaacgarza/dev-stack/internal/pkg/constants"
	"github.com/isaacgarza/dev-stack/internal/pkg/types"
	"gopkg.in/yaml.v3"
)

// Manager provides high-level service management operations
type Manager struct {
	docker     *docker.Client
	logger     *slog.Logger
	projectDir string
	config     *types.Config

	// Sub-managers
	operations *ServiceOperations
	cleanup    *CleanupManager
}

// NewManager creates a new service manager instance
func NewManager(logger *slog.Logger, projectDir string) (*Manager, error) {
	dockerClient, err := docker.NewClient(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	manager := &Manager{
		docker:     dockerClient,
		logger:     logger,
		projectDir: projectDir,
	}

	// Initialize sub-managers
	manager.operations = NewServiceOperations(manager)
	manager.cleanup = NewCleanupManager(manager)

	return manager, nil
}

// SetConfig sets the project configuration
func (m *Manager) SetConfig(config *types.Config) {
	m.config = config
}

// Close closes the service manager and its resources
func (m *Manager) Close() error {
	return m.docker.Close()
}

// Core service operations

// StartServices starts the specified services or all services if none specified
func (m *Manager) StartServices(ctx context.Context, serviceNames []string, options types.StartOptions) error {
	m.logger.Info("Starting services", "services", serviceNames, "detach", options.Detach)

	projectName := m.getProjectName()

	// Validate services exist
	if len(serviceNames) > 0 {
		if err := m.validateServices(serviceNames); err != nil {
			return err
		}
	}

	// Check for port conflicts before starting
	if err := m.checkPortConflicts(ctx, serviceNames); err != nil {
		return fmt.Errorf("port conflict detected: %w", err)
	}

	// Start services using Docker client
	if err := m.docker.Containers().Start(ctx, projectName, serviceNames, options); err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	}

	// Wait for services to be healthy if not detached
	if !options.Detach {
		if err := m.waitForHealthy(ctx, projectName, serviceNames, options.Timeout); err != nil {
			return fmt.Errorf("services failed to become healthy: %w", err)
		}
	}

	m.logger.Info("Services started successfully", "services", serviceNames)
	return nil
}

// StopServices stops the specified services or all services if none specified
func (m *Manager) StopServices(ctx context.Context, serviceNames []string, options types.StopOptions) error {
	m.logger.Info("Stopping services", "services", serviceNames, "timeout", options.Timeout)

	projectName := m.getProjectName()

	if err := m.docker.Containers().Stop(ctx, projectName, serviceNames, options); err != nil {
		return fmt.Errorf("failed to stop services: %w", err)
	}

	m.logger.Info("Services stopped successfully", "services", serviceNames)
	return nil
}

// GetServiceStatus returns the status of all services or specified services
func (m *Manager) GetServiceStatus(ctx context.Context, serviceNames []string) ([]types.ServiceStatus, error) {
	projectName := m.getProjectName()

	services, err := m.docker.Containers().List(ctx, projectName, serviceNames)
	if err != nil {
		return nil, fmt.Errorf("failed to get service status: %w", err)
	}

	// Calculate uptime for running services
	for i := range services {
		if services[i].State == constants.StateRunning && services[i].StartedAt != nil {
			services[i].Uptime = time.Since(*services[i].StartedAt)
		}
	}

	return services, nil
}

// ExecCommand executes a command in a service container
func (m *Manager) ExecCommand(ctx context.Context, serviceName string, cmd []string, options types.ExecOptions) error {
	projectName := m.getProjectName()

	if err := m.docker.Containers().Exec(ctx, projectName, serviceName, cmd, options); err != nil {
		return fmt.Errorf("failed to execute command in %s: %w", serviceName, err)
	}

	return nil
}

// GetLogs retrieves logs from services
func (m *Manager) GetLogs(ctx context.Context, serviceNames []string, options types.LogOptions) error {
	projectName := m.getProjectName()

	if err := m.docker.Containers().Logs(ctx, projectName, serviceNames, options); err != nil {
		return fmt.Errorf("failed to get logs: %w", err)
	}

	return nil
}

// Service operations (delegated to operations manager)

// ConnectToService provides convenient connection to services
func (m *Manager) ConnectToService(ctx context.Context, serviceName string, options types.ConnectOptions) error {
	return m.operations.ConnectToService(ctx, serviceName, options)
}

// BackupService creates a backup of service data
func (m *Manager) BackupService(ctx context.Context, serviceName, backupName string, options types.BackupOptions) error {
	return m.operations.BackupService(ctx, serviceName, backupName, options)
}

// RestoreService restores service data from a backup
func (m *Manager) RestoreService(ctx context.Context, serviceName, backupFile string, options types.RestoreOptions) error {
	return m.operations.RestoreService(ctx, serviceName, backupFile, options)
}

// ScaleService scales a service to the specified number of replicas
func (m *Manager) ScaleService(ctx context.Context, serviceName string, replicas int, options types.ScaleOptions) error {
	return m.operations.ScaleService(ctx, serviceName, replicas, options)
}

// Resource management (delegated to cleanup manager)

// CleanupResources removes project resources
func (m *Manager) CleanupResources(ctx context.Context, options types.CleanupOptions) error {
	return m.cleanup.CleanupResources(ctx, options)
}

// Helper methods (package-private for sub-managers)

func (m *Manager) getProjectName() string {
	if m.config != nil && m.config.Global.DefaultProjectType != "" {
		return m.config.Global.DefaultProjectType
	}
	return filepath.Base(m.projectDir)
}

func (m *Manager) validateServices(serviceNames []string) error {
	servicesYAMLPath := filepath.Join(constants.ServicesDir, "services.yaml")
	data, err := os.ReadFile(servicesYAMLPath)
	if err != nil {
		for _, name := range serviceNames {
			if strings.TrimSpace(name) == "" {
				return fmt.Errorf("empty service name provided")
			}
		}
		return nil
	}

	var services map[string]interface{}
	if err := yaml.Unmarshal(data, &services); err != nil {
		for _, name := range serviceNames {
			if strings.TrimSpace(name) == "" {
				return fmt.Errorf("empty service name provided")
			}
		}
		return nil
	}

	for _, name := range serviceNames {
		if strings.TrimSpace(name) == "" {
			return fmt.Errorf("empty service name provided")
		}
		if _, exists := services[name]; !exists {
			availableServices := make([]string, 0, len(services))
			for serviceName := range services {
				availableServices = append(availableServices, serviceName)
			}
			return fmt.Errorf("unknown service '%s'. Available services: %v", name, availableServices)
		}
	}
	return nil
}

func (m *Manager) checkPortConflicts(ctx context.Context, serviceNames []string) error {
	// Load service configurations dynamically to get ports
	conflicts := []string{}
	for _, serviceName := range serviceNames {
		_ = serviceName // TODO: implement dynamic port checking
	}

	if len(conflicts) > 0 {
		return fmt.Errorf("port conflicts detected for services: %v", conflicts)
	}
	return nil
}

func (m *Manager) waitForHealthy(ctx context.Context, projectName string, serviceNames []string, timeout time.Duration) error {
	m.logger.Info("Waiting for services to become healthy", "services", serviceNames, "timeout", timeout)

	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return fmt.Errorf("timeout waiting for services to become healthy")
			}

			allHealthy := true
			statuses, err := m.GetServiceStatus(ctx, serviceNames)
			if err != nil {
				m.logger.Warn("Failed to get service status during health check", "error", err)
				continue
			}

			for _, status := range statuses {
				if status.State != constants.StateRunning || (status.Health != constants.HealthHealthy && status.Health != "") {
					allHealthy = false
					break
				}
			}

			if allHealthy {
				m.logger.Info("All services are healthy")
				return nil
			}
		}
	}
}
