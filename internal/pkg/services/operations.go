package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/isaacgarza/dev-stack/internal/pkg/constants"
	"gopkg.in/yaml.v3"
)

// ServiceOperations defines operations available for a service
type ServiceOperations struct {
	Connect *ConnectOperation `yaml:"connect,omitempty"`
	Backup  *BackupOperation  `yaml:"backup,omitempty"`
	Restore *RestoreOperation `yaml:"restore,omitempty"`
}

// ConnectOperation defines how to connect to a service
type ConnectOperation struct {
	Command  []string            `yaml:"command"`
	Args     map[string][]string `yaml:"args,omitempty"`
	Defaults map[string]string   `yaml:"defaults,omitempty"`
}

// BackupOperation defines how to backup a service
type BackupOperation struct {
	Type      string              `yaml:"type"` // "command" or "custom"
	Command   []string            `yaml:"command,omitempty"`
	Commands  [][]string          `yaml:"commands,omitempty"` // for custom multi-step
	Args      map[string][]string `yaml:"args,omitempty"`
	Defaults  map[string]string   `yaml:"defaults,omitempty"`
	Extension string              `yaml:"extension"`
}

// RestoreOperation defines how to restore a service
type RestoreOperation struct {
	Type            string                `yaml:"type"`
	Command         []string              `yaml:"command,omitempty"`
	Commands        [][]string            `yaml:"commands,omitempty"`
	PreCommands     map[string][][]string `yaml:"pre_commands,omitempty"`
	Args            map[string][]string   `yaml:"args,omitempty"`
	Defaults        map[string]string     `yaml:"defaults,omitempty"`
	RequiresRestart bool                  `yaml:"requires_restart,omitempty"`
}

// ServiceConfig represents a service configuration with operations
type ServiceConfig struct {
	Name       string             `yaml:"name"`
	Operations *ServiceOperations `yaml:"operations,omitempty"`
}

// LoadServiceOperations loads operations for a service from its YAML file
func LoadServiceOperations(serviceName string) (*ServiceOperations, error) {
	// Find service YAML file
	serviceFile, err := findServiceFile(serviceName)
	if err != nil {
		return nil, fmt.Errorf("service file not found for %s: %w", serviceName, err)
	}

	data, err := os.ReadFile(serviceFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read service file %s: %w", serviceFile, err)
	}

	var config ServiceConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse service config %s: %w", serviceFile, err)
	}

	return config.Operations, nil
}

// BuildConnectCommand builds a connection command for a service
func (op *ConnectOperation) BuildCommand(options map[string]string) []string {
	if op == nil {
		return nil
	}

	cmd := make([]string, len(op.Command))
	copy(cmd, op.Command)

	// Apply defaults first
	params := make(map[string]string)
	for k, v := range op.Defaults {
		params[k] = v
	}
	// Override with provided options
	for k, v := range options {
		params[k] = v
	}

	// Add arguments based on parameters
	for param, value := range params {
		if argTemplate, exists := op.Args[param]; exists && value != "" {
			for _, arg := range argTemplate {
				rendered := strings.ReplaceAll(arg, "{{."+strings.ToUpper(param[:1])+param[1:]+"}}", value)
				cmd = append(cmd, rendered)
			}
		}
	}

	return cmd
}

// BuildBackupCommand builds a backup command for a service
func (op *BackupOperation) BuildCommand(options map[string]string) ([][]string, error) {
	if op == nil {
		return nil, fmt.Errorf("no backup operation defined")
	}

	params := make(map[string]string)
	for k, v := range op.Defaults {
		params[k] = v
	}
	for k, v := range options {
		params[k] = v
	}

	var commands [][]string

	if op.Type == "custom" && len(op.Commands) > 0 {
		// Multi-step custom commands
		for _, cmdTemplate := range op.Commands {
			cmd := make([]string, len(cmdTemplate))
			for i, part := range cmdTemplate {
				cmd[i] = renderTemplate(part, params)
			}
			commands = append(commands, cmd)
		}
	} else if len(op.Command) > 0 {
		// Single command
		cmd := make([]string, len(op.Command))
		copy(cmd, op.Command)

		// Add arguments
		for param, value := range params {
			if argTemplate, exists := op.Args[param]; exists && value != "" {
				for _, arg := range argTemplate {
					rendered := renderTemplate(arg, params)
					cmd = append(cmd, rendered)
				}
			}
		}
		commands = append(commands, cmd)
	}

	return commands, nil
}

// GetBackupExtension returns the file extension for backups
func (op *BackupOperation) GetBackupExtension() string {
	if op == nil || op.Extension == "" {
		return "backup"
	}
	return op.Extension
}

// findServiceFile finds the YAML file for a service
func findServiceFile(serviceName string) (string, error) {
	// Search in all service directories
	servicesDirs := []string{
		filepath.Join(constants.ServicesDir, "cache"),
		filepath.Join(constants.ServicesDir, "database"),
		filepath.Join(constants.ServicesDir, "messaging"),
		filepath.Join(constants.ServicesDir, "observability"),
		filepath.Join(constants.ServicesDir, "cloud"),
	}

	for _, dir := range servicesDirs {
		serviceFile := filepath.Join(dir, serviceName+".yaml")
		if _, err := os.Stat(serviceFile); err == nil {
			return serviceFile, nil
		}
	}

	return "", fmt.Errorf("service file not found for %s", serviceName)
}

// renderTemplate renders a template string with parameters
func renderTemplate(templateStr string, params map[string]string) string {
	tmpl, err := template.New("cmd").Parse(templateStr)
	if err != nil {
		return templateStr // Return original if parsing fails
	}

	var result strings.Builder
	if err := tmpl.Execute(&result, params); err != nil {
		return templateStr // Return original if execution fails
	}

	return result.String()
}
