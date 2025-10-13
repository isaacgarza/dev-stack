package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/isaacgarza/dev-stack/internal/pkg/services"
	"github.com/isaacgarza/dev-stack/internal/pkg/types"
)

// ServiceOperations handles service-specific operations like connect, backup, restore
type ServiceOperations struct {
	manager *Manager
}

// NewServiceOperations creates a new service operations handler
func NewServiceOperations(manager *Manager) *ServiceOperations {
	return &ServiceOperations{manager: manager}
}

// ConnectToService provides convenient connection to services using dynamic configuration
func (so *ServiceOperations) ConnectToService(ctx context.Context, serviceName string, options types.ConnectOptions) error {
	projectName := so.manager.getProjectName()

	// Load service operations
	ops, err := services.LoadServiceOperations(serviceName)
	if err != nil {
		return fmt.Errorf("failed to load service operations for %s: %w", serviceName, err)
	}

	if ops.Connect == nil {
		return fmt.Errorf("no connect operation defined for service %s", serviceName)
	}

	// Build connection parameters
	params := map[string]string{
		"user":     options.User,
		"database": options.Database,
		"host":     options.Host,
		"port":     options.Port,
	}

	// Build command
	cmd := ops.Connect.BuildCommand(params)
	if len(cmd) == 0 {
		return fmt.Errorf("failed to build connect command for %s", serviceName)
	}

	// Execute the connection command
	execOptions := types.ExecOptions{
		Interactive: true,
		TTY:         true,
		User:        options.User,
	}

	if err := so.manager.docker.Containers().Exec(ctx, projectName, serviceName, cmd, execOptions); err != nil {
		return fmt.Errorf("failed to connect to %s: %w", serviceName, err)
	}

	return nil
}

// BackupService creates a backup of service data using dynamic configuration
func (so *ServiceOperations) BackupService(ctx context.Context, serviceName, backupName string, options types.BackupOptions) error {
	so.manager.logger.Info("Creating backup", "service", serviceName, "backup", backupName)

	projectName := so.manager.getProjectName()
	backupDir := options.OutputDir
	if backupDir == "" {
		backupDir = "./backups"
	}

	// Ensure backup directory exists
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Load service operations
	ops, err := services.LoadServiceOperations(serviceName)
	if err != nil {
		return fmt.Errorf("failed to load service operations for %s: %w", serviceName, err)
	}

	if ops.Backup == nil {
		return fmt.Errorf("no backup operation defined for service %s", serviceName)
	}

	// Build backup file path
	extension := ops.Backup.GetBackupExtension()
	backupPath := filepath.Join(backupDir, fmt.Sprintf("%s.%s", backupName, extension))

	// Build backup parameters
	params := map[string]string{
		"database":   options.Database,
		"user":       options.User,
		"timestamp":  time.Now().Format("20060102_150405"),
		"backupFile": backupPath,
	}

	// Build and execute backup commands
	commands, err := ops.Backup.BuildCommand(params)
	if err != nil {
		return fmt.Errorf("failed to build backup command for %s: %w", serviceName, err)
	}

	execOptions := types.ExecOptions{
		User: options.User,
	}

	for _, cmd := range commands {
		if err := so.manager.docker.Containers().Exec(ctx, projectName, serviceName, cmd, execOptions); err != nil {
			return fmt.Errorf("failed to execute backup command for %s: %w", serviceName, err)
		}
	}

	so.manager.logger.Info("Backup created successfully", "service", serviceName, "backup", backupPath)
	return nil
}

// RestoreService restores service data from a backup using dynamic configuration
func (so *ServiceOperations) RestoreService(ctx context.Context, serviceName, backupFile string, options types.RestoreOptions) error {
	so.manager.logger.Info("Restoring from backup", "service", serviceName, "backup", backupFile)

	projectName := so.manager.getProjectName()

	// Validate backup file exists
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		return fmt.Errorf("backup file not found: %s", backupFile)
	}

	// Load service operations
	ops, err := services.LoadServiceOperations(serviceName)
	if err != nil {
		return fmt.Errorf("failed to load service operations for %s: %w", serviceName, err)
	}

	if ops.Restore == nil {
		return fmt.Errorf("no restore operation defined for service %s", serviceName)
	}

	// Build restore parameters
	params := map[string]string{
		"database":   options.Database,
		"user":       options.User,
		"backupFile": backupFile,
	}

	execOptions := types.ExecOptions{
		User: options.User,
	}

	// Execute pre-commands if clean is requested
	if options.Clean && ops.Restore.PreCommands != nil {
		if cleanCommands, exists := ops.Restore.PreCommands["clean"]; exists {
			for _, cmdTemplate := range cleanCommands {
				cmd := make([]string, len(cmdTemplate))
				for i, part := range cmdTemplate {
					cmd[i] = renderTemplate(part, params)
				}
				if err := so.manager.docker.Containers().Exec(ctx, projectName, serviceName, cmd, execOptions); err != nil {
					return fmt.Errorf("failed to execute pre-command for %s: %w", serviceName, err)
				}
			}
		}
	}

	// Execute restore commands
	if ops.Restore.Type == "custom" && len(ops.Restore.Commands) > 0 {
		for _, cmdTemplate := range ops.Restore.Commands {
			cmd := make([]string, len(cmdTemplate))
			for i, part := range cmdTemplate {
				cmd[i] = renderTemplate(part, params)
			}
			if err := so.manager.docker.Containers().Exec(ctx, projectName, serviceName, cmd, execOptions); err != nil {
				return fmt.Errorf("failed to execute restore command for %s: %w", serviceName, err)
			}
		}
	} else if len(ops.Restore.Command) > 0 {
		cmd := make([]string, len(ops.Restore.Command))
		copy(cmd, ops.Restore.Command)

		// Add arguments
		for param, value := range params {
			if argTemplate, exists := ops.Restore.Args[param]; exists && value != "" {
				for _, arg := range argTemplate {
					rendered := renderTemplate(arg, params)
					cmd = append(cmd, rendered)
				}
			}
		}

		if err := so.manager.docker.Containers().Exec(ctx, projectName, serviceName, cmd, execOptions); err != nil {
			return fmt.Errorf("failed to restore %s: %w", serviceName, err)
		}
	}

	// Restart service if required
	if ops.Restore.RequiresRestart {
		if err := so.manager.StopServices(ctx, []string{serviceName}, types.StopOptions{Timeout: 10}); err != nil {
			return fmt.Errorf("failed to stop %s for restart: %w", serviceName, err)
		}

		startOptions := types.StartOptions{
			Build:         false,
			ForceRecreate: false,
			Detach:        true,
			Timeout:       30 * time.Second,
		}

		if err := so.manager.StartServices(ctx, []string{serviceName}, startOptions); err != nil {
			return fmt.Errorf("failed to restart %s after restore: %w", serviceName, err)
		}
	}

	so.manager.logger.Info("Restore completed successfully", "service", serviceName, "backup", backupFile)
	return nil
}

// ScaleService scales a service to the specified number of replicas
func (so *ServiceOperations) ScaleService(ctx context.Context, serviceName string, replicas int, options types.ScaleOptions) error {
	so.manager.logger.Info("Scaling service", "service", serviceName, "replicas", replicas)

	if replicas < 0 {
		return fmt.Errorf("replica count cannot be negative")
	}

	// Validate service exists
	if err := so.manager.validateServices([]string{serviceName}); err != nil {
		return fmt.Errorf("service validation failed: %w", err)
	}

	// Scale to 0 means stop the service
	if replicas == 0 {
		stopOptions := types.StopOptions{
			Timeout:       int(options.Timeout.Seconds()),
			Remove:        true,
			RemoveVolumes: false,
		}
		return so.manager.StopServices(ctx, []string{serviceName}, stopOptions)
	}

	// For replicas > 0, ensure service is running
	statuses, err := so.manager.GetServiceStatus(ctx, []string{serviceName})
	if err != nil {
		return fmt.Errorf("failed to get service status: %w", err)
	}

	if len(statuses) == 0 || !statuses[0].State.IsRunning() {
		startOptions := types.StartOptions{
			Build:         false,
			ForceRecreate: options.NoRecreate,
			Detach:        true,
			Timeout:       options.Timeout,
		}

		if err := so.manager.StartServices(ctx, []string{serviceName}, startOptions); err != nil {
			return fmt.Errorf("failed to start service for scaling: %w", err)
		}
	}

	so.manager.logger.Info("Service scaling completed", "service", serviceName, "replicas", replicas)
	return nil
}

// renderTemplate renders a template string with parameters
func renderTemplate(templateStr string, params map[string]string) string {
	result := templateStr
	for key, value := range params {
		placeholder := "{{." + strings.ToUpper(key[:1]) + key[1:] + "}}"
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}
