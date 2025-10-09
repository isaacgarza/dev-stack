package handlers

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/isaacgarza/dev-stack/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// InitHandler handles the init command
type InitHandler struct{}

// NewInitHandler creates a new init handler
func NewInitHandler() *InitHandler {
	return &InitHandler{}
}

// Handle executes the init command
func (h *InitHandler) Handle(ctx context.Context, cmd *cobra.Command, args []string, base *BaseCommand) error {
	force, _ := cmd.Flags().GetBool("force")

	// Check if dev-stack-config.yaml already exists
	configFile := "dev-stack-config.yaml"
	if _, err := os.Stat(configFile); err == nil && !force {
		return fmt.Errorf("dev-stack project already exists (use --force to overwrite)")
	}

	return h.runInteractive(configFile)
}

// runInteractive runs the interactive initialization
func (h *InitHandler) runInteractive(configFile string) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("🚀 Welcome to dev-stack interactive setup!")
	fmt.Println()

	// Get project name
	cwd, _ := os.Getwd()
	defaultName := filepath.Base(cwd)
	name := h.promptString(reader, fmt.Sprintf("Project name (%s)", defaultName), defaultName)

	// Get environment
	environment := h.promptString(reader, "Environment (local)", "local")

	// Select services
	services := h.promptServices(reader)

	// Configure validation settings
	validation := h.promptValidation(reader)

	// Configure advanced settings
	advanced := h.promptAdvanced(reader)

	// Generate config
	config, err := h.generateConfig(name, environment, services, validation, advanced)
	if err != nil {
		return fmt.Errorf("failed to generate config: %w", err)
	}

	// Write config file
	if err := os.WriteFile(configFile, []byte(config), 0644); err != nil {
		return fmt.Errorf("failed to create %s: %w", configFile, err)
	}

	fmt.Printf("✅ Created %s\n", configFile)

	// Create dev-stack directory for generated files
	if err := os.MkdirAll("dev-stack", 0755); err != nil {
		fmt.Printf("⚠️  Warning: failed to create dev-stack directory: %v\n", err)
	}

	// Generate initial compose files
	if err := h.generateInitialComposeFiles(services, name, environment, validation, advanced); err != nil {
		fmt.Printf("⚠️  Warning: failed to generate initial compose files: %v\n", err)
	}

	h.createGitignore()
	h.printSuccessMessage(name, configFile)

	return nil
}

// promptString prompts for a string value
func (h *InitHandler) promptString(reader *bufio.Reader, prompt, defaultValue string) string {
	fmt.Printf("%s: ", prompt)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue
	}
	return input
}

// promptServices prompts for service selection using category-first workflow
func (h *InitHandler) promptServices(reader *bufio.Reader) []string {
	fmt.Println("\n📦 Select service categories to configure:")

	// Load services by category using shared utilities
	serviceUtils := NewServiceUtils()
	servicesByCategory, err := serviceUtils.LoadServicesByCategory()
	if err != nil {
		fmt.Printf("❌ Error loading services: %v\n", err)
		fmt.Println("Please ensure you're running dev-stack from the correct location.")
		os.Exit(1)
	}

	if len(servicesByCategory) == 0 {
		fmt.Println("❌ No services found in any category")
		os.Exit(1)
	}

	// Show available categories with service counts
	var selectedCategories []string
	for category, services := range servicesByCategory {
		serviceNames := make([]string, len(services))
		for i, service := range services {
			serviceNames[i] = service.Name
		}
		
		fmt.Printf("  %s services (%s)\n", strings.Title(category), strings.Join(serviceNames, ", "))
		if h.promptYesOrNo(reader, fmt.Sprintf("Configure %s services?", category), false) {
			selectedCategories = append(selectedCategories, category)
		}
	}

	if len(selectedCategories) == 0 {
		fmt.Println("⚠️  No categories selected.")
		return []string{}
	}

	// For each selected category, show services and allow selection
	var selectedServices []string
	for _, category := range selectedCategories {
		services := servicesByCategory[category]
		fmt.Printf("\n%s services:\n", strings.Title(category))
		
		for _, service := range services {
			fmt.Printf("  %s - %s", service.Name, service.Description)
			if len(service.Dependencies) > 0 {
				fmt.Printf(" (requires: %s)", strings.Join(service.Dependencies, ", "))
			}
			fmt.Println()
			
			if h.promptYesOrNo(reader, fmt.Sprintf("  Enable %s?", service.Name), false) {
				selectedServices = append(selectedServices, service.Name)
				
				// Auto-add required dependencies and notify user
				if len(service.Dependencies) > 0 {
					fmt.Printf("    → Auto-adding required dependencies: %s\n", strings.Join(service.Dependencies, ", "))
					for _, dep := range service.Dependencies {
						// Check if dependency is not already selected
						found := false
						for _, selected := range selectedServices {
							if selected == dep {
								found = true
								break
							}
						}
						if !found {
							selectedServices = append(selectedServices, dep)
						}
					}
				}
			}
		}
	}

	if len(selectedServices) == 0 {
		fmt.Println("⚠️  No services selected.")
		return []string{}
	}

	// Auto-resolve dependencies (for any nested dependencies)
	resolvedServices, err := serviceUtils.ResolveDependencies(selectedServices)
	if err != nil {
		fmt.Printf("⚠️  Warning: dependency resolution failed: %v\n", err)
		return selectedServices
	}

	if len(resolvedServices) > len(selectedServices) {
		additionalDeps := []string{}
		for _, resolved := range resolvedServices {
			found := false
			for _, selected := range selectedServices {
				if selected == resolved {
					found = true
					break
				}
			}
			if !found {
				additionalDeps = append(additionalDeps, resolved)
			}
		}
		if len(additionalDeps) > 0 {
			fmt.Printf("\n🔗 Additional nested dependencies resolved: %s\n", strings.Join(additionalDeps, ", "))
		}
	}

	return resolvedServices
}

// promptValidation prompts for validation settings
func (h *InitHandler) promptValidation(reader *bufio.Reader) map[string]bool {
	fmt.Println("\n🔍 Validation settings:")
	return h.promptSettings(reader, "validation")
}

// promptAdvanced prompts for advanced settings
func (h *InitHandler) promptAdvanced(reader *bufio.Reader) map[string]bool {
	fmt.Println("\n⚙️  Advanced settings:")
	return h.promptSettings(reader, "advanced")
}

// promptSettings prompts for a specific settings section
func (h *InitHandler) promptSettings(reader *bufio.Reader, section string) map[string]bool {
	initSettings, err := h.loadInitSettings()
	if err != nil {
		fmt.Printf("⚠️  Warning: failed to load init settings: %v\n", err)
		// Return fallback defaults based on section
		if section == "validation" {
			return map[string]bool{"skip_warnings": false, "allow_multiple_databases": true}
		}
		return map[string]bool{"auto_start": true, "pull_latest_images": true, "cleanup_on_recreate": false}
	}

	var settingsMap map[string]struct {
		Description string `yaml:"description"`
		Default     bool   `yaml:"default"`
	}

	if section == "validation" {
		settingsMap = initSettings.Validation
	} else {
		settingsMap = initSettings.Advanced
	}

	settings := make(map[string]bool)
	for key, setting := range settingsMap {
		fmt.Printf("%s - %s\n", key, setting.Description)
		settings[key] = h.promptBool(reader, fmt.Sprintf("Set %s: ", key), setting.Default)
	}

	return settings
}

// promptYesOrNo prompts for a yes/no answer (for actions like "Enable service?")
func (h *InitHandler) promptYesOrNo(reader *bufio.Reader, prompt string, defaultValue bool) bool {
	defaultStr := "n"
	if defaultValue {
		defaultStr = "y"
	}

	fmt.Printf("%s (y/n, default %s): ", prompt, defaultStr)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "" {
		return defaultValue
	}

	return input == "y" || input == "yes"
}

// promptBool prompts for a boolean value (for settings like true/false)
func (h *InitHandler) promptBool(reader *bufio.Reader, prompt string, defaultValue bool) bool {
	defaultStr := "false"
	if defaultValue {
		defaultStr = "true"
	}

	fmt.Printf("%s (true/false, default %s): ", prompt, defaultStr)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "" {
		return defaultValue
	}

	return input == "true" || input == "t"
}

// generateConfig generates config from template
func (h *InitHandler) generateConfig(name, environment string, services []string, validation, advanced map[string]bool) (string, error) {
	candidates := []string{
		"internal/config/dev-stack-config.template",
		"config/dev-stack-config.template",
		".dev-stack/dev-stack-config.template",
	}

	var templateContent []byte
	var err error

	// Try to find template file
	if templatePath, findErr := h.findTemplateFile(candidates, "template file"); findErr == nil {
		templateContent, err = os.ReadFile(templatePath)
		if err != nil {
			return "", fmt.Errorf("failed to read template: %w", err)
		}
	} else {
		// Use embedded template as fallback
		templateContent = config.EmbeddedConfigTemplate
		if len(templateContent) == 0 {
			return "", fmt.Errorf("no template file found and no embedded template available")
		}
	}

	// Parse template
	tmpl, err := template.New("config").Parse(string(templateContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Prepare template data
	data := struct {
		ProjectName string
		Environment string
		Services    []string
		Validation  map[string]bool
		Advanced    map[string]bool
	}{
		ProjectName: name,
		Environment: environment,
		Services:    services,
		Validation:  validation,
		Advanced:    advanced,
	}

	// Execute template
	var result strings.Builder
	if err := tmpl.Execute(&result, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return result.String(), nil
}

// createGitignore creates a .gitignore file in dev-stack folder
func (h *InitHandler) createGitignore() {
	gitignoreFile := "dev-stack/.gitignore"

	candidates := []string{
		"internal/config/gitignore.txt",
		"config/gitignore.txt",
		"dev-stack/gitignore.txt",
	}

	var gitignoreContent []byte
	if gitignoreTemplatePath, err := h.findTemplateFile(candidates, "gitignore template"); err == nil {
		// Read from template file
		content, err := os.ReadFile(gitignoreTemplatePath)
		if err != nil {
			fmt.Printf("⚠️  Warning: failed to read gitignore template: %v\n", err)
			return
		}
		gitignoreContent = content
	} else {
		// Use embedded gitignore as fallback
		gitignoreContent = config.EmbeddedGitignore
		if len(gitignoreContent) == 0 {
			// Final fallback to minimal content
			gitignoreContent = []byte("# dev-stack\ndev-stack-config.yaml.bak\n*.log\n")
		}
	}

	if err := os.WriteFile(gitignoreFile, gitignoreContent, 0644); err != nil {
		fmt.Printf("⚠️  Warning: failed to create .gitignore: %v\n", err)
	} else {
		fmt.Printf("✅ Created %s\n", gitignoreFile)
	}
}

// printSuccessMessage prints the success message
func (h *InitHandler) printSuccessMessage(name, configFile string) {
	fmt.Printf("✅ Project '%s' initialized successfully!\n", name)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  1. Review %s and adjust settings as needed\n", configFile)
	fmt.Printf("  2. Run 'dev-stack up' to start your development stack\n")
}

// copyServiceConfigs copies selected service configurations to .dev-stack folder
func (h *InitHandler) copyServiceConfigs(services []string) error {
	// Create .dev-stack directory
	devStackDir := ".dev-stack"
	if err := os.MkdirAll(devStackDir, 0755); err != nil {
		return fmt.Errorf("failed to create %s directory: %w", devStackDir, err)
	}

	candidates := []string{
		"internal/config/services",
		"config/services",
	}

	// Try to find local services directory first
	servicesSourceDir, err := h.findTemplateFile(candidates, "services source directory")
	if err != nil {
		// Use embedded services filesystem
		return h.copyEmbeddedServices(services, devStackDir)
	}

	// Copy from local directory
	return h.copyLocalServices(services, servicesSourceDir, devStackDir)
}

// copyLocalServices copies services from local filesystem
func (h *InitHandler) copyLocalServices(services []string, servicesSourceDir, devStackDir string) error {
	for _, service := range services {
		sourceDir := filepath.Join(servicesSourceDir, service)
		targetDir := filepath.Join(devStackDir, service)

		if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
			fmt.Printf("⚠️  Warning: service directory not found for %s\n", service)
			continue
		}

		// Create target directory
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", targetDir, err)
		}

		// Copy all files from source to target
		if err := h.copyDir(sourceDir, targetDir); err != nil {
			return fmt.Errorf("failed to copy service %s: %w", service, err)
		}

		fmt.Printf("✅ Copied %s service configuration\n", service)
	}
	return nil
}

// copyEmbeddedServices copies services from embedded filesystem
func (h *InitHandler) copyEmbeddedServices(services []string, devStackDir string) error {
	for _, service := range services {
		serviceDir := filepath.Join("services", service)
		targetDir := filepath.Join(devStackDir, service)

		// Check if service exists in embedded FS
		if _, err := config.EmbeddedServicesFS.ReadDir(serviceDir); err != nil {
			fmt.Printf("⚠️  Warning: embedded service directory not found for %s\n", service)
			continue
		}

		// Create target directory
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", targetDir, err)
		}

		// Copy files from embedded FS
		if err := h.copyEmbeddedDir(serviceDir, targetDir); err != nil {
			return fmt.Errorf("failed to copy embedded service %s: %w", service, err)
		}

		fmt.Printf("✅ Copied %s service configuration\n", service)
	}
	return nil
}

// copyEmbeddedDir copies a directory from embedded filesystem
func (h *InitHandler) copyEmbeddedDir(srcPath, dstPath string) error {
	entries, err := config.EmbeddedServicesFS.ReadDir(srcPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcFile := filepath.Join(srcPath, entry.Name())
		dstFile := filepath.Join(dstPath, entry.Name())

		if entry.IsDir() {
			if err := os.MkdirAll(dstFile, 0755); err != nil {
				return err
			}
			if err := h.copyEmbeddedDir(srcFile, dstFile); err != nil {
				return err
			}
		} else {
			content, err := config.EmbeddedServicesFS.ReadFile(srcFile)
			if err != nil {
				return err
			}
			if err := os.WriteFile(dstFile, content, 0644); err != nil {
				return err
			}
		}
	}
	return nil
}

// copyDir recursively copies a directory
func (h *InitHandler) copyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := os.MkdirAll(dstPath, 0755); err != nil {
				return err
			}
			if err := h.copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := h.copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file
func (h *InitHandler) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// loadServicesByCategory loads services organized by category from embedded filesystem
func (h *InitHandler) loadServicesByCategory() (map[string][]struct {
	name         string
	description  string
	dependencies []string
}, error) {
	categories := []string{"database", "cache", "messaging", "observability", "cloud"}
	
	servicesByCategory := make(map[string][]struct {
		name         string
		description  string
		dependencies []string
	})

	for _, category := range categories {
		categoryPath := filepath.Join("services", category)
		
		// Use embedded filesystem
		entries, err := config.EmbeddedServicesFS.ReadDir(categoryPath)
		if err != nil {
			continue // Category doesn't exist, skip
		}

		var categoryServices []struct {
			name         string
			description  string
			dependencies []string
		}

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
				continue
			}

			serviceName := strings.TrimSuffix(entry.Name(), ".yaml")
			serviceFile := filepath.Join(categoryPath, entry.Name())
			
			data, err := config.EmbeddedServicesFS.ReadFile(serviceFile)
			if err != nil {
				continue
			}

			var serviceData map[string]interface{}
			if err := yaml.Unmarshal(data, &serviceData); err != nil {
				continue
			}

			description, _ := serviceData["description"].(string)
			var dependencies []string

			if deps, exists := serviceData["dependencies"]; exists {
				if depsMap, ok := deps.(map[string]interface{}); ok {
					if required, exists := depsMap["required"]; exists {
						if reqList, ok := required.([]interface{}); ok {
							for _, req := range reqList {
								if reqStr, ok := req.(string); ok {
									dependencies = append(dependencies, reqStr)
								}
							}
						}
					}
				}
			}

			categoryServices = append(categoryServices, struct {
				name         string
				description  string
				dependencies []string
			}{
				name:         serviceName,
				description:  description,
				dependencies: dependencies,
			})
		}

		if len(categoryServices) > 0 {
			servicesByCategory[category] = categoryServices
		}
	}

	return servicesByCategory, nil
}

// resolveDependencies resolves service dependencies and returns ordered list using embedded filesystem
func (h *InitHandler) resolveDependencies(selectedServices []string) ([]string, error) {
	serviceMap := make(map[string][]string)
	
	// Load all service dependencies from embedded filesystem
	categories := []string{"database", "cache", "messaging", "observability", "cloud"}
	for _, category := range categories {
		categoryPath := filepath.Join("services", category)
		
		entries, err := config.EmbeddedServicesFS.ReadDir(categoryPath)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
				continue
			}

			serviceName := strings.TrimSuffix(entry.Name(), ".yaml")
			serviceFile := filepath.Join(categoryPath, entry.Name())
			
			data, err := config.EmbeddedServicesFS.ReadFile(serviceFile)
			if err != nil {
				continue
			}

			var serviceData map[string]interface{}
			if err := yaml.Unmarshal(data, &serviceData); err != nil {
				continue
			}

			var dependencies []string
			if deps, exists := serviceData["dependencies"]; exists {
				if depsMap, ok := deps.(map[string]interface{}); ok {
					if required, exists := depsMap["required"]; exists {
						if reqList, ok := required.([]interface{}); ok {
							for _, req := range reqList {
								if reqStr, ok := req.(string); ok {
									dependencies = append(dependencies, reqStr)
								}
							}
						}
					}
				}
			}
			serviceMap[serviceName] = dependencies
		}
	}

	// Resolve dependencies using topological sort
	visited := make(map[string]bool)
	visiting := make(map[string]bool)
	var result []string

	var visit func(string) error
	visit = func(serviceName string) error {
		if visiting[serviceName] {
			return fmt.Errorf("circular dependency detected involving service: %s", serviceName)
		}
		if visited[serviceName] {
			return nil
		}

		visiting[serviceName] = true
		for _, dep := range serviceMap[serviceName] {
			if err := visit(dep); err != nil {
				return err
			}
		}
		visiting[serviceName] = false
		visited[serviceName] = true
		result = append(result, serviceName)
		return nil
	}

	for _, service := range selectedServices {
		if err := visit(service); err != nil {
			return selectedServices, err
		}
	}

	return result, nil
}

// loadInitSettings loads initialization settings from config
func (h *InitHandler) loadInitSettings() (struct {
	Validation map[string]struct {
		Description string `yaml:"description"`
		Default     bool   `yaml:"default"`
	} `yaml:"validation"`
	Advanced map[string]struct {
		Description string `yaml:"description"`
		Default     bool   `yaml:"default"`
	} `yaml:"advanced"`
}, error) {
	candidates := []string{
		"internal/config/init-settings.yaml",
		"config/init-settings.yaml",
		".dev-stack/init-settings.yaml",
	}

	data, err := h.loadConfigFile(candidates, config.EmbeddedInitSettingsYAML, "init-settings.yaml")
	if err != nil {
		return struct {
			Validation map[string]struct {
				Description string `yaml:"description"`
				Default     bool   `yaml:"default"`
			} `yaml:"validation"`
			Advanced map[string]struct {
				Description string `yaml:"description"`
				Default     bool   `yaml:"default"`
			} `yaml:"advanced"`
		}{}, err
	}

	var initSettings struct {
		Validation map[string]struct {
			Description string `yaml:"description"`
			Default     bool   `yaml:"default"`
		} `yaml:"validation"`
		Advanced map[string]struct {
			Description string `yaml:"description"`
			Default     bool   `yaml:"default"`
		} `yaml:"advanced"`
	}
	if err := yaml.Unmarshal(data, &initSettings); err != nil {
		return initSettings, fmt.Errorf("failed to parse init-settings.yaml: %w", err)
	}

	return initSettings, nil
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

	fmt.Println("✅ Generated dev-stack/docker-compose.yml and dev-stack/.env.generated")
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
		Config *ServiceConfig
	}

	for _, serviceName := range services {
		serviceConfig, err := NewServiceUtils().LoadServiceConfig(serviceName)
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
		ProjectName:  pc.Project.Name,
		Environment:  pc.Project.Environment,
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
		Config *ServiceConfig
	}
	var volumes []string

	for _, serviceName := range services {
		serviceConfig, err := NewServiceUtils().LoadServiceConfig(serviceName)
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
			volumeName := fmt.Sprintf("%s-%s", pc.Project.Name, volume.Name)
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

// loadInitServiceConfig loads a service configuration from embedded FS using category-based paths
func (h *InitHandler) loadInitServiceConfig(serviceName string) (*ServiceConfig, error) {
	// Search for the service in all categories
	categories := []string{"database", "cache", "messaging", "observability", "cloud"}
	
	for _, category := range categories {
		servicePath := fmt.Sprintf("services/%s/%s.yaml", category, serviceName)
		data, err := config.EmbeddedServicesFS.ReadFile(servicePath)
		if err != nil {
			continue // Try next category
		}

		var serviceConfig ServiceConfig
		if err := yaml.Unmarshal(data, &serviceConfig); err != nil {
			return nil, fmt.Errorf("failed to parse service config for %s: %w", serviceName, err)
		}

		return &serviceConfig, nil
	}
	
	return nil, fmt.Errorf("service %s not found in any category", serviceName)
}

// findTemplateFile finds a template file from candidate locations
func (h *InitHandler) findTemplateFile(candidates []string, description string) (string, error) {
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("%s not found in any of: %v", description, candidates)
}

// loadConfigFile loads a config file from candidates or embedded data
func (h *InitHandler) loadConfigFile(candidates []string, embeddedData []byte, description string) ([]byte, error) {
	// Try to find file in candidate locations
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			data, err := os.ReadFile(candidate)
			if err != nil {
				return nil, fmt.Errorf("failed to read %s: %w", candidate, err)
			}
			return data, nil
		}
	}

	// Use embedded data as fallback
	if len(embeddedData) == 0 {
		return nil, fmt.Errorf("no %s found in %v and no embedded config available", description, candidates)
	}
	return embeddedData, nil
}

// ValidateArgs validates the command arguments
func (h *InitHandler) ValidateArgs(args []string) error {
	return nil
}

// GetRequiredFlags returns required flags for this command
func (h *InitHandler) GetRequiredFlags() []string {
	return []string{}
}
