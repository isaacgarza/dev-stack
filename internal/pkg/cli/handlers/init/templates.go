package init

import (
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/isaacgarza/dev-stack/internal/config"
	"github.com/isaacgarza/dev-stack/internal/pkg/cli/handlers/utils"
	"github.com/isaacgarza/dev-stack/internal/pkg/cli/types"
	pkgConfig "github.com/isaacgarza/dev-stack/internal/pkg/config"
	"github.com/isaacgarza/dev-stack/internal/pkg/ui"
)

// generateConfig generates config using code generation
func (h *InitHandler) generateConfig(name, environment string, services []string, validation, advanced map[string]bool) (string, error) {
	return pkgConfig.GenerateConfig(name, environment, services, validation, advanced), nil
}

// generateInitialComposeFiles generates initial compose files during init
func (h *InitHandler) generateInitialComposeFiles(services []string, projectName, environment string, validation, advanced map[string]bool) error {
	// Create a temporary project config structure
	projectConfig := struct {
		Project struct {
			Name        string
			Environment string
		}
		Stack struct {
			Enabled []string
		}
	}{
		Project: struct {
			Name        string
			Environment string
		}{
			Name:        projectName,
			Environment: environment,
		},
		Stack: struct {
			Enabled []string
		}{
			Enabled: services,
		},
	}

	// Generate .env.generated
	if err := h.generateInitEnvFile(services, &projectConfig); err != nil {
		return fmt.Errorf("failed to generate .env file: %w", err)
	}

	// Generate docker-compose.yml
	if err := h.generateInitDockerCompose(services, &projectConfig); err != nil {
		return fmt.Errorf("failed to generate docker-compose.yml: %w", err)
	}

	ui.Success("Generated dev-stack/docker-compose.yml and dev-stack/.env.generated")
	return nil
}

// generateInitEnvFile generates .env.generated during init using template
func (h *InitHandler) generateInitEnvFile(services []string, projectConfig interface{}) error {
	pc := projectConfig.(*struct {
		Project struct {
			Name        string
			Environment string
		}
		Stack struct {
			Enabled []string
		}
	})

	// Load template
	var templateContent []byte
	candidates := []string{
		"internal/config/env.template",
		"config/env.template",
		"dev-stack/env.template",
	}

	if templatePath, err := h.findTemplateFile(candidates, "env template"); err == nil {
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
		Config *types.ServiceConfig
	}

	for _, serviceName := range services {
		serviceConfig, err := utils.NewServiceUtils().LoadServiceConfig(serviceName)
		if err != nil {
			ui.Warning("Failed to load config for %s: %v", serviceName, err)
			continue
		}
		templateServices = append(templateServices, struct {
			Name   string
			Config *types.ServiceConfig
		}{
			Name:   serviceName,
			Config: serviceConfig,
		})
	}

	data := struct {
		ProjectName string
		Environment string
		GeneratedAt string
		Services    []struct {
			Name   string
			Config *types.ServiceConfig
		}
	}{
		ProjectName: pc.Project.Name,
		Environment: pc.Project.Environment,
		GeneratedAt: time.Now().Format(time.RFC1123),
		Services:    templateServices,
	}

	// Execute template
	var result strings.Builder
	if err := tmpl.Execute(&result, data); err != nil {
		return fmt.Errorf("failed to execute env template: %w", err)
	}

	return os.WriteFile("dev-stack/.env.generated", []byte(result.String()), 0644)
}

// generateInitDockerCompose generates docker-compose.yml during init using template
func (h *InitHandler) generateInitDockerCompose(services []string, projectConfig interface{}) error {
	pc := projectConfig.(*struct {
		Project struct {
			Name        string
			Environment string
		}
		Stack struct {
			Enabled []string
		}
	})

	// Load template
	var templateContent []byte
	candidates := []string{
		"internal/config/docker-compose.template",
		"config/docker-compose.template",
		"dev-stack/docker-compose.template",
	}

	if templatePath, err := h.findTemplateFile(candidates, "docker-compose template"); err == nil {
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
		Config *types.ServiceConfig
	}
	var volumes []string

	for _, serviceName := range services {
		serviceConfig, err := utils.NewServiceUtils().LoadServiceConfig(serviceName)
		if err != nil {
			ui.Warning("Failed to load config for %s: %v", serviceName, err)
			continue
		}

		templateServices = append(templateServices, struct {
			Name   string
			Config *types.ServiceConfig
		}{
			Name:   serviceName,
			Config: serviceConfig,
		})

		// Collect volumes
		for _, volume := range serviceConfig.Volumes {
			volumeName := fmt.Sprintf("%s-%s", pc.Project.Name, volume.Name)
			volumes = append(volumes, volumeName)
		}
	}

	data := struct {
		ProjectName string
		Services    []struct {
			Name   string
			Config *types.ServiceConfig
		}
		Volumes []string
	}{
		ProjectName: pc.Project.Name,
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
