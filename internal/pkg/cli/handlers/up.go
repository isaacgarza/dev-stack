package handlers

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/isaacgarza/dev-stack/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// UpHandler handles the "up" command for starting services
type UpHandler struct{}

// NewUpHandler creates a new up command handler
func NewUpHandler() *UpHandler {
	return &UpHandler{}
}

// Handle executes the up command
func (h *UpHandler) Handle(ctx context.Context, cmd *cobra.Command, args []string, base *BaseCommand) error {
	detach, _ := cmd.Flags().GetBool("detach")
	build, _ := cmd.Flags().GetBool("build")
	profile, _ := cmd.Flags().GetString("profile")

	// If no specific services provided, read from config
	servicesToStart := args
	if len(args) == 0 {
		config, err := h.loadProjectConfig()
		if err != nil {
			return fmt.Errorf("failed to load project configuration: %w", err)
		}
		servicesToStart = config.Stack.Enabled
	}

	// Validate services if specified
	if len(servicesToStart) > 0 {
		if err := base.ValidateServices(servicesToStart); err != nil {
			return err
		}
	}

	// Display what we're starting
	if len(args) == 0 {
		fmt.Printf("🚀 Starting services from config: %v\n", servicesToStart)
	} else {
		fmt.Printf("🚀 Starting specified services: %v\n", servicesToStart)
	}

	if profile != "" {
		fmt.Printf("📋 Using profile: %s\n", profile)
	}

	// Generate docker-compose.yml and .env.generated
	if err := h.generateComposeFiles(servicesToStart); err != nil {
		return fmt.Errorf("failed to generate compose files: %w", err)
	}

	// Set up start options
	startOptions := StartOptions{
		Build:   build,
		Detach:  detach,
		Timeout: 60, // Increased from 30 to 60 seconds
	}

	// Start services
	return base.Manager.StartServices(ctx, servicesToStart, startOptions)
}

// generateComposeFiles generates docker-compose.yml and .env.generated
func (h *UpHandler) generateComposeFiles(services []string) error {
	// Ensure dev-stack directory exists
	if err := os.MkdirAll("dev-stack", 0755); err != nil {
		return fmt.Errorf("failed to create dev-stack directory: %w", err)
	}

	// Load project config for overrides
	projectConfig, err := h.loadProjectConfig()
	if err != nil {
		return fmt.Errorf("failed to load project config: %w", err)
	}

	// Generate .env.generated
	if err := h.generateEnvFile(services, projectConfig); err != nil {
		return fmt.Errorf("failed to generate .env file: %w", err)
	}

	// Generate docker-compose.yml
	if err := h.generateDockerCompose(services, projectConfig); err != nil {
		return fmt.Errorf("failed to generate docker-compose.yml: %w", err)
	}

	fmt.Println("✅ Generated dev-stack/docker-compose.yml and dev-stack/.env.generated")
	return nil
}

// generateEnvFile generates .env.generated using template
func (h *UpHandler) generateEnvFile(services []string, projectConfig *ProjectConfig) error {
	// Load template
	var templateContent []byte
	candidates := []string{
		"internal/config/env.template",
		"config/env.template",
		"dev-stack/env.template",
	}

	if templatePath, err := h.findTemplate(candidates, "env template"); err == nil {
		content, err := os.ReadFile(templatePath)
		if err != nil {
			return fmt.Errorf("failed to read env template: %w", err)
		}
		templateContent = content
	} else {
		templateContent = config.EmbeddedEnvTemplate
		if len(templateContent) == 0 {
			return fmt.Errorf("no env template found and no embedded template available")
		}
	}

	// Parse template with custom functions
	tmpl, err := template.New("env").Funcs(template.FuncMap{
		"ToUpper": strings.ToUpper,
	}).Parse(string(templateContent))
	if err != nil {
		return fmt.Errorf("failed to parse env template: %w", err)
	}

	// Prepare template data
	var templateServices []struct {
		Name   string
		Config *ServiceConfig
	}

	for _, serviceName := range services {
		serviceConfig, err := h.loadServiceConfig(serviceName)
		if err != nil {
			fmt.Printf("⚠️  Warning: failed to load config for %s: %v\n", serviceName, err)
			continue
		}
		templateServices = append(templateServices, struct {
			Name   string
			Config *ServiceConfig
		}{
			Name:   serviceName,
			Config: serviceConfig,
		})
	}

	data := struct {
		ProjectName  string
		Environment  string
		GeneratedAt  string
		Services     []struct {
			Name   string
			Config *ServiceConfig
		}
	}{
		ProjectName:  projectConfig.Project.Name,
		Environment:  projectConfig.Project.Environment,
		GeneratedAt:  time.Now().Format(time.RFC1123),
		Services:     templateServices,
	}

	// Execute template
	var result strings.Builder
	if err := tmpl.Execute(&result, data); err != nil {
		return fmt.Errorf("failed to execute env template: %w", err)
	}

	return os.WriteFile("dev-stack/.env.generated", []byte(result.String()), 0644)
}

// generateDockerCompose generates docker-compose.yml using template
func (h *UpHandler) generateDockerCompose(services []string, projectConfig *ProjectConfig) error {
	// Load template
	var templateContent []byte
	candidates := []string{
		"internal/config/docker-compose.template",
		"config/docker-compose.template",
		"dev-stack/docker-compose.template",
	}

	if templatePath, err := h.findTemplate(candidates, "docker-compose template"); err == nil {
		content, err := os.ReadFile(templatePath)
		if err != nil {
			return fmt.Errorf("failed to read docker-compose template: %w", err)
		}
		templateContent = content
	} else {
		templateContent = config.EmbeddedDockerComposeTemplate
		if len(templateContent) == 0 {
			return fmt.Errorf("no docker-compose template found and no embedded template available")
		}
	}

	// Parse template with custom functions
	tmpl, err := template.New("docker-compose").Funcs(template.FuncMap{
		"toYamlArray": func(arr []string) string {
			if len(arr) == 0 {
				return "[]"
			}
			result := "["
			for i, item := range arr {
				if i > 0 {
					result += ", "
				}
				result += fmt.Sprintf(`"%s"`, item)
			}
			result += "]"
			return result
		},
	}).Parse(string(templateContent))
	if err != nil {
		return fmt.Errorf("failed to parse docker-compose template: %w", err)
	}

	// Prepare template data
	var templateServices []struct {
		Name   string
		Config *ServiceConfig
	}
	var volumes []string

	for _, serviceName := range services {
		serviceConfig, err := h.loadServiceConfig(serviceName)
		if err != nil {
			fmt.Printf("⚠️  Warning: failed to load config for %s: %v\n", serviceName, err)
			continue
		}

		templateServices = append(templateServices, struct {
			Name   string
			Config *ServiceConfig
		}{
			Name:   serviceName,
			Config: serviceConfig,
		})

		// Collect volumes
		for _, volume := range serviceConfig.Volumes {
			volumeName := fmt.Sprintf("%s-%s", projectConfig.Project.Name, volume.Name)
			volumes = append(volumes, volumeName)
		}
	}

	data := struct {
		ProjectName string
		Services    []struct {
			Name   string
			Config *ServiceConfig
		}
		Volumes []string
	}{
		ProjectName: projectConfig.Project.Name,
		Services:    templateServices,
		Volumes:     volumes,
	}

	// Execute template
	var result strings.Builder
	if err := tmpl.Execute(&result, data); err != nil {
		return fmt.Errorf("failed to execute docker-compose template: %w", err)
	}

	return os.WriteFile("dev-stack/docker-compose.yml", []byte(result.String()), 0644)
}

// findTemplate finds a template file from candidate locations
func (h *UpHandler) findTemplate(candidates []string, description string) (string, error) {
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("%s not found in any of: %v", description, candidates)
}

// loadServiceConfig loads a service configuration from embedded FS
func (h *UpHandler) loadServiceConfig(serviceName string) (*ServiceConfig, error) {
	servicePath := fmt.Sprintf("services/%s/service.yaml", serviceName)
	data, err := config.EmbeddedServicesFS.ReadFile(servicePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read service config for %s: %w", serviceName, err)
	}

	var serviceConfig ServiceConfig
	if err := yaml.Unmarshal(data, &serviceConfig); err != nil {
		return nil, fmt.Errorf("failed to parse service config for %s: %w", serviceName, err)
	}

	return &serviceConfig, nil
}

// ProjectConfig represents the dev-stack-config.yaml structure
type ProjectConfig struct {
	Project struct {
		Name        string `yaml:"name"`
		Environment string `yaml:"environment"`
	} `yaml:"project"`
	Stack struct {
		Enabled []string `yaml:"enabled"`
	} `yaml:"stack"`
}

// loadProjectConfig loads the dev-stack-config.yaml file
func (h *UpHandler) loadProjectConfig() (*ProjectConfig, error) {
	configFile := "dev-stack-config.yaml"
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("dev-stack-config.yaml not found in current directory. Run 'dev-stack init' first")
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", configFile, err)
	}

	var config ProjectConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", configFile, err)
	}

	return &config, nil
}

// ValidateArgs validates the command arguments
func (h *UpHandler) ValidateArgs(args []string) error {
	// No specific validation needed for up command
	// Service validation is done in Handle method
	return nil
}

// GetRequiredFlags returns the required flags for this command
func (h *UpHandler) GetRequiredFlags() []string {
	return []string{} // No required flags for up command
}
