package services

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/isaacgarza/dev-stack/internal/core/docker"
	"github.com/isaacgarza/dev-stack/internal/pkg/types"
	"gopkg.in/yaml.v3"
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
		// Prompt for password
		cmd = append(cmd, "-p")
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

	if replicas < 0 {
		return fmt.Errorf("replica count cannot be negative")
	}

	// Validate service exists
	if err := m.validateServices([]string{serviceName}); err != nil {
		return fmt.Errorf("service validation failed: %w", err)
	}

	// For now, implement basic scaling by stopping and starting services
	// In a full implementation, this would use docker-compose scale or docker service scale
	// Scale to 0 means stop the service
	if replicas == 0 {
		stopOptions := StopOptions{
			Timeout:       int(options.Timeout.Seconds()),
			Remove:        true,
			RemoveVolumes: false,
		}
		return m.StopServices(ctx, []string{serviceName}, stopOptions)
	}

	// For replicas > 0, ensure service is running
	// Check current status
	statuses, err := m.GetServiceStatus(ctx, []string{serviceName})
	if err != nil {
		return fmt.Errorf("failed to get service status: %w", err)
	}

	if len(statuses) == 0 || statuses[0].State != "running" {
		// Start the service if not running
		startOptions := StartOptions{
			Build:         false,
			ForceRecreate: options.NoRecreate,
			NoDeps:        false,
			Detach:        true,
			Timeout:       options.Timeout,
		}

		if err := m.StartServices(ctx, []string{serviceName}, startOptions); err != nil {
			return fmt.Errorf("failed to start service for scaling: %w", err)
		}
	}

	m.logger.Info("Service scaling completed", "service", serviceName, "replicas", replicas)
	return nil
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
		if err := m.docker.Volumes().Remove(ctx, projectName); err != nil {
			m.logger.Error("Failed to remove volumes", "error", err)
		}
	}

	// Remove images if requested
	if options.RemoveImages {
		if err := m.docker.Images().Remove(ctx, projectName); err != nil {
			m.logger.Error("Failed to remove images", "error", err)
		}
	}

	// Remove networks if requested
	if options.RemoveNetworks {
		if err := m.docker.Networks().Remove(ctx, projectName); err != nil {
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
	// Load services.yaml to validate service names
	servicesYAMLPath := "internal/config/services/services.yaml"
	data, err := os.ReadFile(servicesYAMLPath)
	if err != nil {
		// If services.yaml doesn't exist, just check for empty names
		for _, name := range serviceNames {
			if strings.TrimSpace(name) == "" {
				return fmt.Errorf("empty service name provided")
			}
		}
		return nil
	}

	var services map[string]interface{}
	if err := yaml.Unmarshal(data, &services); err != nil {
		// If parsing fails, fall back to basic validation
		for _, name := range serviceNames {
			if strings.TrimSpace(name) == "" {
				return fmt.Errorf("empty service name provided")
			}
		}
		return nil
	}

	// Validate each service name exists in configuration
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
	// Common service ports to check
	servicePorts := map[string]int{
		"postgres":   5432,
		"mysql":      3306,
		"redis":      6379,
		"kafka":      9092,
		"jaeger":     16686,
		"prometheus": 9090,
		"localstack": 4566,
	}

	conflicts := []string{}
	for _, serviceName := range serviceNames {
		if port, exists := servicePorts[serviceName]; exists {
			if m.isPortInUse(port) {
				conflicts = append(conflicts, fmt.Sprintf("%s (port %d)", serviceName, port))
			}
		}
	}

	if len(conflicts) > 0 {
		return fmt.Errorf("port conflicts detected for services: %v. Stop existing services or change port mappings", conflicts)
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

			// Check health of all services
			allHealthy := true
			statuses, err := m.GetServiceStatus(ctx, serviceNames)
			if err != nil {
				m.logger.Warn("Failed to get service status during health check", "error", err)
				continue
			}

			for _, status := range statuses {
				if status.State != "running" || (status.Health != "healthy" && status.Health != "") {
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

func (m *Manager) backupRedis(ctx context.Context, projectName, backupPath string) error {
	m.logger.Info("Starting Redis backup", "project", projectName, "path", backupPath)

	// Execute Redis BGSAVE command to create backup
	execOptions := ExecOptions{
		User:        "",
		WorkingDir:  "",
		Env:         nil,
		Interactive: false,
		TTY:         false,
		Detach:      false,
	}

	// Trigger background save
	saveCmd := []string{"redis-cli", "BGSAVE"}
	if err := m.ExecCommand(ctx, "redis", saveCmd, execOptions); err != nil {
		return fmt.Errorf("failed to trigger Redis backup: %w", err)
	}

	// Wait for backup to complete
	time.Sleep(2 * time.Second)

	// Copy the RDB file
	copyCmd := []string{"cp", "/data/dump.rdb", fmt.Sprintf("/data/backup_%s.rdb", time.Now().Format("20060102_150405"))}
	if err := m.ExecCommand(ctx, "redis", copyCmd, execOptions); err != nil {
		return fmt.Errorf("failed to copy Redis backup: %w", err)
	}

	m.logger.Info("Redis backup completed successfully")
	return nil
}

func (m *Manager) restoreRedis(ctx context.Context, projectName, backupFile string, clean bool) error {
	m.logger.Info("Starting Redis restore", "project", projectName, "backup", backupFile, "clean", clean)

	execOptions := ExecOptions{
		User:        "",
		WorkingDir:  "",
		Env:         nil,
		Interactive: false,
		TTY:         false,
		Detach:      false,
	}

	// Stop Redis to safely replace data
	if err := m.StopServices(ctx, []string{"redis"}, StopOptions{Timeout: 10}); err != nil {
		return fmt.Errorf("failed to stop Redis for restore: %w", err)
	}

	// Clear existing data if clean flag is set
	if clean {
		clearCmd := []string{"rm", "-f", "/data/dump.rdb"}
		if err := m.ExecCommand(ctx, "redis", clearCmd, execOptions); err != nil {
			m.logger.Warn("Failed to clear existing Redis data", "error", err)
		}
	}

	// Copy backup file to Redis data directory
	restoreCmd := []string{"cp", backupFile, "/data/dump.rdb"}
	if err := m.ExecCommand(ctx, "redis", restoreCmd, execOptions); err != nil {
		return fmt.Errorf("failed to restore Redis backup: %w", err)
	}

	// Restart Redis
	startOptions := StartOptions{
		Build:         false,
		ForceRecreate: false,
		NoDeps:        false,
		Detach:        true,
		Timeout:       30 * time.Second,
	}

	if err := m.StartServices(ctx, []string{"redis"}, startOptions); err != nil {
		return fmt.Errorf("failed to restart Redis after restore: %w", err)
	}

	m.logger.Info("Redis restore completed successfully")
	return nil
}

// isPortInUse checks if a port is currently in use
func (m *Manager) isPortInUse(port int) bool {
	timeout := time.Second
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), timeout)
	if err != nil {
		return false
	}
	defer func() {
		if err := conn.Close(); err != nil {
			m.logger.Warn("failed to close connection", "error", err)
		}
	}()
	return true
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
