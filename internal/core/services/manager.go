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
	"github.com/isaacgarza/dev-stack/internal/pkg/types"
)

// Database service constants
const (
	ServicePostgres   = "postgres"
	ServicePostgreSQL = "postgresql"
	ServiceMySQL      = "mysql"
	ServiceRedis      = "redis"
	ServiceMariaDB    = "mariadb"
	ServiceMongo      = "mongo"
	ServiceMongoDB    = "mongodb"
)

// Manager provides high-level service management operations
type Manager struct {
	docker     *docker.Client
	logger     *slog.Logger
	projectDir string
	config     *types.Config
}

// NewManager creates a new service manager instance
func NewManager(logger *slog.Logger, projectDir string) (*Manager, error) {
	dockerClient, err := docker.NewClient(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	return &Manager{
		docker:     dockerClient,
		logger:     logger,
		projectDir: projectDir,
	}, nil
}

// SetConfig sets the project configuration
func (m *Manager) SetConfig(config *types.Config) {
	m.config = config
}

// Close closes the service manager and its resources
func (m *Manager) Close() error {
	return m.docker.Close()
}

// StartServices starts the specified services or all services if none specified
func (m *Manager) StartServices(ctx context.Context, serviceNames []string, options StartOptions) error {
	m.logger.Info("Starting services", "services", serviceNames, "detach", options.Detach)

	// Get project name from directory or config
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
	dockerOptions := docker.StartOptions{
		Build:         options.Build,
		ForceRecreate: options.ForceRecreate,
		NoDeps:        options.NoDeps,
		Detach:        options.Detach,
	}

	if err := m.docker.Containers().Start(ctx, projectName, serviceNames, dockerOptions); err != nil {
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
func (m *Manager) StopServices(ctx context.Context, serviceNames []string, options StopOptions) error {
	m.logger.Info("Stopping services", "services", serviceNames, "timeout", options.Timeout)

	projectName := m.getProjectName()

	dockerOptions := docker.StopOptions{
		Timeout:       options.Timeout,
		Remove:        options.Remove,
		RemoveVolumes: options.RemoveVolumes,
	}

	if err := m.docker.Containers().Stop(ctx, projectName, serviceNames, dockerOptions); err != nil {
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
		if services[i].State == "running" && services[i].StartedAt != nil {
			services[i].Uptime = time.Since(*services[i].StartedAt)
		}
	}

	return services, nil
}

// ExecCommand executes a command in a service container
func (m *Manager) ExecCommand(ctx context.Context, serviceName string, cmd []string, options ExecOptions) error {
	projectName := m.getProjectName()

	dockerOptions := docker.ExecOptions{
		User:        options.User,
		WorkingDir:  options.WorkingDir,
		Env:         options.Env,
		Interactive: options.Interactive,
		TTY:         options.TTY,
		Detach:      options.Detach,
	}

	if err := m.docker.Containers().Exec(ctx, projectName, serviceName, cmd, dockerOptions); err != nil {
		return fmt.Errorf("failed to execute command in %s: %w", serviceName, err)
	}

	return nil
}

// GetLogs retrieves logs from services
func (m *Manager) GetLogs(ctx context.Context, serviceNames []string, options LogOptions) error {
	projectName := m.getProjectName()

	dockerOptions := docker.LogOptions{
		Follow:     options.Follow,
		Timestamps: options.Timestamps,
		Tail:       options.Tail,
		Since:      options.Since,
	}

	if err := m.docker.Containers().Logs(ctx, projectName, serviceNames, dockerOptions); err != nil {
		return fmt.Errorf("failed to get logs: %w", err)
	}

	return nil
}

// ConnectToService provides convenient connection to common services
func (m *Manager) ConnectToService(ctx context.Context, serviceName string, options ConnectOptions) error {
	projectName := m.getProjectName()

	// Map service names to connection commands
	var cmd []string
	switch strings.ToLower(serviceName) {
	case ServicePostgres, ServicePostgreSQL, "pg":
		cmd = []string{"psql"}
		if options.User != "" {
			cmd = append(cmd, "-U", options.User)
		} else {
			cmd = append(cmd, "-U", "postgres")
		}
		if options.Database != "" {
			cmd = append(cmd, "-d", options.Database)
		}

	case ServiceRedis:
		cmd = []string{"redis-cli"}
		if options.Host != "" {
			cmd = append(cmd, "-h", options.Host)
		}
		if options.Port != "" {
			cmd = append(cmd, "-p", options.Port)
		}

	case ServiceMySQL, ServiceMariaDB:
		cmd = []string{"mysql"}
		if options.User != "" {
			cmd = append(cmd, "-u", options.User)
		} else {
			cmd = append(cmd, "-u", "root")
		}
		cmd = append(cmd, "-p") // Prompt for password
		if options.Database != "" {
			cmd = append(cmd, options.Database)
		}

	case ServiceMongo, ServiceMongoDB:
		cmd = []string{"mongosh"}
		if options.Database != "" {
			cmd = append(cmd, options.Database)
		}

	case "elastic", "elasticsearch":
		cmd = []string{"curl", "localhost:9200"}

	default:
		return fmt.Errorf("unsupported service for connection: %s", serviceName)
	}

	// Execute the connection command
	execOptions := docker.ExecOptions{
		Interactive: true,
		TTY:         true,
		User:        options.User,
	}

	if err := m.docker.Containers().Exec(ctx, projectName, serviceName, cmd, execOptions); err != nil {
		return fmt.Errorf("failed to connect to %s: %w", serviceName, err)
	}

	return nil
}

// ScaleService scales a service to the specified number of replicas
func (m *Manager) ScaleService(ctx context.Context, serviceName string, replicas int, options ScaleOptions) error {
	m.logger.Info("Scaling service", "service", serviceName, "replicas", replicas)

	// TODO: Implement service scaling
	// This would typically involve:
	// 1. Updating the compose file or configuration
	// 2. Using docker-compose scale or docker service scale
	// 3. Waiting for containers to start/stop
	// 4. Verifying the desired replica count is achieved

	return fmt.Errorf("service scaling not yet implemented")
}

// BackupService creates a backup of service data
func (m *Manager) BackupService(ctx context.Context, serviceName, backupName string, options BackupOptions) error {
	m.logger.Info("Creating backup", "service", serviceName, "backup", backupName)

	projectName := m.getProjectName()
	backupDir := options.OutputDir
	if backupDir == "" {
		backupDir = "./backups"
	}

	// Ensure backup directory exists
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	backupPath := filepath.Join(backupDir, fmt.Sprintf("%s.%s", backupName, getBackupExtension(serviceName)))

	// Create backup based on service type
	var cmd []string
	switch strings.ToLower(serviceName) {
	case ServicePostgres, ServicePostgreSQL, "pg":
		cmd = []string{"pg_dump", "-U", "postgres", "-h", "localhost"}
		if options.Database != "" {
			cmd = append(cmd, options.Database)
		}

	case ServiceMySQL, ServiceMariaDB:
		cmd = []string{"mysqldump", "-u", "root", "-p"}
		if options.Database != "" {
			cmd = append(cmd, options.Database)
		} else {
			cmd = append(cmd, "--all-databases")
		}

	case ServiceRedis:
		// For Redis, we'll copy the RDB file
		return m.backupRedis(ctx, projectName, backupPath)

	case ServiceMongoDB, ServiceMongo:
		cmd = []string{"mongodump", "--out", "/tmp/backup"}

	default:
		return fmt.Errorf("unsupported service for backup: %s", serviceName)
	}

	// Execute backup command
	execOptions := docker.ExecOptions{
		User: options.User,
	}

	if err := m.docker.Containers().Exec(ctx, projectName, serviceName, cmd, execOptions); err != nil {
		return fmt.Errorf("failed to create backup for %s: %w", serviceName, err)
	}

	m.logger.Info("Backup created successfully", "service", serviceName, "backup", backupPath)
	return nil
}

// RestoreService restores service data from a backup
func (m *Manager) RestoreService(ctx context.Context, serviceName, backupFile string, options RestoreOptions) error {
	m.logger.Info("Restoring from backup", "service", serviceName, "backup", backupFile)

	projectName := m.getProjectName()

	// Validate backup file exists
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		return fmt.Errorf("backup file not found: %s", backupFile)
	}

	// Create restore command based on service type
	var cmd []string
	switch strings.ToLower(serviceName) {
	case "postgres", "postgresql", "pg":
		if options.Clean {
			// Drop and recreate database
			if options.Database != "" {
				dropCmd := []string{"dropdb", "-U", "postgres", "--if-exists", options.Database}
				if err := m.docker.Containers().Exec(ctx, projectName, serviceName, dropCmd, docker.ExecOptions{}); err != nil {
					return fmt.Errorf("failed to drop database: %w", err)
				}

				createCmd := []string{"createdb", "-U", "postgres", options.Database}
				if err := m.docker.Containers().Exec(ctx, projectName, serviceName, createCmd, docker.ExecOptions{}); err != nil {
					return fmt.Errorf("failed to create database: %w", err)
				}
			}
		}

		cmd = []string{"psql", "-U", "postgres"}
		if options.Database != "" {
			cmd = append(cmd, "-d", options.Database)
		}

	case "mysql", "mariadb":
		if options.Clean && options.Database != "" {
			dropCmd := []string{"mysql", "-u", "root", "-p", "-e", fmt.Sprintf("DROP DATABASE IF EXISTS %s;", options.Database)}
			if err := m.docker.Containers().Exec(ctx, projectName, serviceName, dropCmd, docker.ExecOptions{}); err != nil {
				return fmt.Errorf("failed to drop database: %w", err)
			}

			createCmd := []string{"mysql", "-u", "root", "-p", "-e", fmt.Sprintf("CREATE DATABASE %s;", options.Database)}
			if err := m.docker.Containers().Exec(ctx, projectName, serviceName, createCmd, docker.ExecOptions{}); err != nil {
				return fmt.Errorf("failed to create database: %w", err)
			}
		}

		cmd = []string{"mysql", "-u", "root", "-p"}
		if options.Database != "" {
			cmd = append(cmd, options.Database)
		}

	case "redis":
		return m.restoreRedis(ctx, projectName, backupFile, options.Clean)

	case "mongodb", "mongo":
		if options.Clean && options.Database != "" {
			dropCmd := []string{"mongosh", "--eval", "db.dropDatabase()", options.Database}
			if err := m.docker.Containers().Exec(ctx, projectName, serviceName, dropCmd, docker.ExecOptions{}); err != nil {
				return fmt.Errorf("failed to drop database: %w", err)
			}
		}

		cmd = []string{"mongorestore", "/tmp/backup"}

	default:
		return fmt.Errorf("unsupported service for restore: %s", serviceName)
	}

	// Execute restore command
	execOptions := docker.ExecOptions{
		User: options.User,
	}

	if err := m.docker.Containers().Exec(ctx, projectName, serviceName, cmd, execOptions); err != nil {
		return fmt.Errorf("failed to restore %s: %w", serviceName, err)
	}

	m.logger.Info("Restore completed successfully", "service", serviceName, "backup", backupFile)
	return nil
}

// CleanupResources removes project resources
func (m *Manager) CleanupResources(ctx context.Context, options CleanupOptions) error {
	m.logger.Info("Cleaning up resources", "volumes", options.RemoveVolumes, "images", options.RemoveImages)

	projectName := m.getProjectName()

	// Stop and remove containers
	if err := m.docker.Containers().Stop(ctx, projectName, []string{}, docker.StopOptions{
		Remove:        true,
		RemoveVolumes: options.RemoveVolumes,
	}); err != nil {
		return fmt.Errorf("failed to remove containers: %w", err)
	}

	// Remove volumes if requested
	if options.RemoveVolumes {
		if err := m.docker.Volumes().Remove(ctx, projectName, []string{}); err != nil {
			m.logger.Error("Failed to remove volumes", "error", err)
		}
	}

	// Remove images if requested
	if options.RemoveImages {
		if err := m.docker.Images().Remove(ctx, projectName, []string{}); err != nil {
			m.logger.Error("Failed to remove images", "error", err)
		}
	}

	// Remove networks if requested
	if options.RemoveNetworks {
		if err := m.docker.Networks().Remove(ctx, projectName, []string{}); err != nil {
			m.logger.Error("Failed to remove networks", "error", err)
		}
	}

	m.logger.Info("Cleanup completed successfully")
	return nil
}

// Helper methods

func (m *Manager) getProjectName() string {
	if m.config != nil && m.config.Global.DefaultProjectType != "" {
		return m.config.Global.DefaultProjectType
	}

	// Use directory name as project name
	return filepath.Base(m.projectDir)
}

func (m *Manager) validateServices(serviceNames []string) error {
	// TODO: Validate that services exist in the project configuration
	// For now, just check they're not empty
	for _, name := range serviceNames {
		if strings.TrimSpace(name) == "" {
			return fmt.Errorf("empty service name provided")
		}
	}
	return nil
}

func (m *Manager) checkPortConflicts(ctx context.Context, serviceNames []string) error {
	// TODO: Check for port conflicts before starting services
	// This would examine the compose file and check if ports are already in use
	return nil
}

func (m *Manager) waitForHealthy(ctx context.Context, projectName string, serviceNames []string, timeout time.Duration) error {
	// TODO: Wait for services to become healthy
	// This would poll service status until all are healthy or timeout is reached
	return nil
}

func (m *Manager) backupRedis(ctx context.Context, projectName, backupPath string) error {
	// TODO: Implement Redis backup by copying RDB file
	return fmt.Errorf("redis backup not yet implemented")
}

func (m *Manager) restoreRedis(ctx context.Context, projectName, backupFile string, clean bool) error {
	// TODO: Implement Redis restore by copying RDB file and restarting
	return fmt.Errorf("redis restore not yet implemented")
}

func getBackupExtension(serviceName string) string {
	switch strings.ToLower(serviceName) {
	case "postgres", "postgresql", "pg", "mysql", "mariadb":
		return "sql"
	case "redis":
		return "rdb"
	case "mongodb", "mongo":
		return "bson"
	default:
		return "backup"
	}
}

// Option types for service operations

type StartOptions struct {
	Build         bool
	ForceRecreate bool
	NoDeps        bool
	Detach        bool
	Timeout       time.Duration
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

type ConnectOptions struct {
	User     string
	Database string
	Host     string
	Port     string
	ReadOnly bool
}

type ScaleOptions struct {
	Detach     bool
	Timeout    time.Duration
	NoRecreate bool
}

type BackupOptions struct {
	OutputDir string
	Compress  bool
	Format    string
	Database  string
	User      string
	NoOwner   bool
	Clean     bool
}

type RestoreOptions struct {
	Database          string
	User              string
	Clean             bool
	CreateDB          bool
	DropDB            bool
	SingleTransaction bool
}

type CleanupOptions struct {
	RemoveVolumes  bool
	RemoveImages   bool
	RemoveNetworks bool
	All            bool
	DryRun         bool
}
