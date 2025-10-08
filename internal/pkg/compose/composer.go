package compose

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/isaacgarza/dev-stack/internal/pkg/version"
	"gopkg.in/yaml.v3"
)

// Composer handles the composition and transformation of service definitions
type Composer struct {
	registry         *ServiceRegistry
	options          ComposeOptions
	configLoader     *ConfigLoader
	conflictDetector *ConflictDetector
	projectConfig    *ProjectConfig
}

// NewComposer creates a new composer with the given registry and options
func NewComposer(registry *ServiceRegistry, options ComposeOptions) *Composer {
	composer := &Composer{
		registry: registry,
		options:  options,
	}

	// Initialize configuration loader if project config is specified
	if options.ProjectConfig != "" {
		composer.configLoader = NewConfigLoader(filepath.Dir(options.ProjectConfig))
	} else {
		// Try to find project root
		if projectRoot := findProjectRoot(); projectRoot != "" {
			composer.configLoader = NewConfigLoader(projectRoot)
		}
	}

	return composer
}

// GenerateCompose generates a docker-compose file from the specified services
func (c *Composer) GenerateCompose(serviceNames []string) (*ComposeFile, error) {
	// Load project configuration if available
	if err := c.loadProjectConfiguration(); err != nil {
		fmt.Printf("Warning: failed to load project configuration: %v\n", err)
	}

	// Apply project-specific service selection if configured
	if c.projectConfig != nil {
		projectServices, err := c.configLoader.GetServicesForProfile(c.options.Profile)
		if err == nil && len(serviceNames) == 0 {
			serviceNames = projectServices
		}
	}

	// Use default services if none specified
	if len(serviceNames) == 0 {
		serviceNames = []string{"postgres", "redis"}
	}

	// Resolve dependencies if enabled
	var resolvedServices []string
	var err error

	if c.options.IncludeDeps {
		resolvedServices, err = c.registry.ResolveDependencies(serviceNames)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve dependencies: %w", err)
		}
	} else {
		resolvedServices = serviceNames
	}

	// Load service definitions
	services, err := c.registry.GetServices(resolvedServices)
	if err != nil {
		return nil, fmt.Errorf("failed to load services: %w", err)
	}

	// Initialize conflict detector if enabled
	if c.options.DetectConflicts {
		serviceMap := make(map[string]*ServiceDefinition)
		for _, service := range services {
			serviceMap[service.Name] = service
		}
		c.conflictDetector = NewConflictDetector(serviceMap, c.options.AutoFixPorts)
	}

	// Create compose file structure
	composeFile := &ComposeFile{
		Version:  "3.8",
		Services: make(map[string]interface{}),
		Networks: make(map[string]interface{}),
		Volumes:  make(map[string]interface{}),
		Metadata: ComposeMetadata{
			GeneratedBy:      "dev-stack",
			GeneratedAt:      time.Now(),
			ProjectName:      c.options.ProjectName,
			Services:         resolvedServices,
			Profile:          c.options.Profile,
			FrameworkVersion: version.GetShortVersion(),
		},
	}

	// Process each service
	for _, service := range services {
		if err := c.mergeService(service, composeFile); err != nil {
			return nil, fmt.Errorf("failed to merge service %s: %w", service.Name, err)
		}
	}

	// Add default network if not already present
	c.ensureDefaultNetwork(composeFile)

	// Apply transformations
	if err := c.applyTransformations(composeFile); err != nil {
		return nil, fmt.Errorf("failed to apply transformations: %w", err)
	}

	// Detect and resolve conflicts if enabled
	if c.conflictDetector != nil {
		if err := c.handleConflicts(composeFile); err != nil {
			return nil, fmt.Errorf("failed to handle conflicts: %w", err)
		}
	}

	return composeFile, nil
}

// mergeService merges a service definition into the compose file
func (c *Composer) mergeService(service *ServiceDefinition, composeFile *ComposeFile) error {
	// Merge services
	for serviceName, serviceConfig := range service.Services {
		// Transform service configuration
		transformedConfig, err := c.transformServiceConfig(serviceName, serviceConfig)
		if err != nil {
			return fmt.Errorf("failed to transform service %s: %w", serviceName, err)
		}

		composeFile.Services[serviceName] = transformedConfig
	}

	// Merge volumes
	for volumeName, volumeConfig := range service.Volumes {
		if _, exists := composeFile.Volumes[volumeName]; !exists {
			composeFile.Volumes[volumeName] = volumeConfig
		}
	}

	// Merge networks
	for networkName, networkConfig := range service.Networks {
		if _, exists := composeFile.Networks[networkName]; !exists {
			composeFile.Networks[networkName] = networkConfig
		}
	}

	return nil
}

// transformServiceConfig applies transformations to a service configuration
func (c *Composer) transformServiceConfig(serviceName string, config interface{}) (interface{}, error) {
	configMap, ok := config.(map[string]interface{})
	if !ok {
		return config, nil
	}

	// Create a deep copy to avoid modifying the original
	transformed := c.deepCopyMap(configMap)

	// Apply container name transformation
	c.transformContainerName(serviceName, transformed)

	// Apply network transformations
	c.transformNetworks(transformed)

	// Apply environment variable transformations
	c.transformEnvironment(transformed)

	// Apply volume transformations
	c.transformVolumes(transformed)

	// Apply service overrides from options
	if overrides, exists := c.options.ServiceOverrides[serviceName]; exists {
		if overrideMap, ok := overrides.(map[string]interface{}); ok {
			c.mergeOverrides(transformed, overrideMap)
		}
	}

	// Apply project-specific overrides
	if c.projectConfig != nil {
		if err := c.applyProjectOverrides(serviceName, transformed); err != nil {
			return nil, fmt.Errorf("failed to apply project overrides: %w", err)
		}
	}

	return transformed, nil
}

// transformContainerName ensures consistent container naming
func (c *Composer) transformContainerName(serviceName string, config map[string]interface{}) {
	if containerName, exists := config["container_name"]; exists {
		if nameStr, ok := containerName.(string); ok {
			// Replace generic project references with actual project name
			newName := strings.ReplaceAll(nameStr, "${PROJECT_NAME:-dev-stack}", c.options.ProjectName)
			newName = strings.ReplaceAll(newName, "dev-stack", c.options.ProjectName)
			config["container_name"] = newName
		}
	} else {
		// Set default container name if not specified
		config["container_name"] = fmt.Sprintf("%s-%s", c.options.ProjectName, serviceName)
	}
}

// transformNetworks transforms network references to use project-specific names
func (c *Composer) transformNetworks(config map[string]interface{}) {
	if networks, exists := config["networks"]; exists {
		switch networks := networks.(type) {
		case []interface{}:
			for i, network := range networks {
				if networkStr, ok := network.(string); ok {
					networks[i] = c.transformNetworkName(networkStr)
				}
			}
		case map[string]interface{}:
			for networkName := range networks {
				transformedName := c.transformNetworkName(networkName)
				if transformedName != networkName {
					networks[transformedName] = networks[networkName]
					delete(networks, networkName)
				}
			}
		}
	}
}

// transformNetworkName transforms a network name to use project-specific naming
func (c *Composer) transformNetworkName(networkName string) string {
	// Transform common network name patterns
	replacements := map[string]string{
		"local-dev": c.options.NetworkName,
		"dev-stack": c.options.NetworkName,
		"default":   c.options.NetworkName,
	}

	if replacement, exists := replacements[networkName]; exists {
		return replacement
	}

	return networkName
}

// transformEnvironment applies environment variable transformations
func (c *Composer) transformEnvironment(config map[string]interface{}) {
	if env, exists := config["environment"]; exists {
		switch env := env.(type) {
		case []interface{}:
			for i, envVar := range env {
				if envStr, ok := envVar.(string); ok {
					env[i] = c.expandEnvironmentVariable(envStr)
				}
			}
		case map[string]interface{}:
			for key, value := range env {
				if valueStr, ok := value.(string); ok {
					env[key] = c.expandEnvironmentVariable(valueStr)
				}
			}
		}
	}

	// Add global environment variables
	c.addGlobalEnvironment(config)
}

// expandEnvironmentVariable expands environment variable references
func (c *Composer) expandEnvironmentVariable(envVar string) string {
	// Replace PROJECT_NAME references
	expanded := strings.ReplaceAll(envVar, "${PROJECT_NAME:-dev-stack}", c.options.ProjectName)
	expanded = strings.ReplaceAll(expanded, "${PROJECT_NAME}", c.options.ProjectName)

	// Apply custom environment variables
	for key, value := range c.options.Environment {
		placeholder := fmt.Sprintf("${%s}", key)
		expanded = strings.ReplaceAll(expanded, placeholder, value)
	}

	return expanded
}

// addGlobalEnvironment adds global environment variables to the service
func (c *Composer) addGlobalEnvironment(config map[string]interface{}) {
	globalEnv := make(map[string]string)

	// Add options environment
	for key, value := range c.options.Environment {
		globalEnv[key] = value
	}

	// Add project configuration environment
	if c.projectConfig != nil {
		if projectEnv, err := c.configLoader.GetGlobalEnvironment(c.options.Profile); err == nil {
			for key, value := range projectEnv {
				globalEnv[key] = value
			}
		}
	}

	if len(globalEnv) == 0 {
		return
	}

	var envList []interface{}
	if env, exists := config["environment"]; exists {
		if envArray, ok := env.([]interface{}); ok {
			envList = envArray
		}
	}

	// Add global environment variables
	for key, value := range globalEnv {
		envVar := fmt.Sprintf("%s=%s", key, value)
		envList = append(envList, envVar)
	}

	if len(envList) > 0 {
		config["environment"] = envList
	}
}

// transformVolumes transforms volume references
func (c *Composer) transformVolumes(config map[string]interface{}) {
	if volumes, exists := config["volumes"]; exists {
		if volumeArray, ok := volumes.([]interface{}); ok {
			for i, volume := range volumeArray {
				if volumeStr, ok := volume.(string); ok {
					// Apply volume prefix if specified
					if c.options.VolumePrefix != "" && !strings.Contains(volumeStr, ":") {
						volumeArray[i] = fmt.Sprintf("%s_%s", c.options.VolumePrefix, volumeStr)
					}
				}
			}
		}
	}
}

// mergeOverrides merges service-specific overrides
func (c *Composer) mergeOverrides(config map[string]interface{}, overrides map[string]interface{}) {
	for key, value := range overrides {
		config[key] = value
	}
}

// ensureDefaultNetwork ensures a default network exists
func (c *Composer) ensureDefaultNetwork(composeFile *ComposeFile) {
	networkName := c.options.NetworkName
	if _, exists := composeFile.Networks[networkName]; !exists {
		composeFile.Networks[networkName] = map[string]interface{}{
			"driver": "bridge",
			"name":   networkName,
		}
	}
}

// applyTransformations applies global transformations to the compose file
func (c *Composer) applyTransformations(composeFile *ComposeFile) error {
	// Apply profile-specific transformations
	switch c.options.Profile {
	case "dev":
		c.applyDevelopmentTransformations(composeFile)
	case "test":
		c.applyTestTransformations(composeFile)
	case "prod":
		c.applyProductionTransformations(composeFile)
	}

	return nil
}

// applyDevelopmentTransformations applies development-specific settings
func (c *Composer) applyDevelopmentTransformations(composeFile *ComposeFile) {
	for serviceName, service := range composeFile.Services {
		if serviceMap, ok := service.(map[string]interface{}); ok {
			// Add restart policy for development
			if _, exists := serviceMap["restart"]; !exists {
				serviceMap["restart"] = "unless-stopped"
			}

			// Add development labels
			if labels, exists := serviceMap["labels"]; exists {
				if labelMap, ok := labels.(map[string]interface{}); ok {
					labelMap["dev-stack.profile"] = "development"
					labelMap["dev-stack.service"] = serviceName
				}
			} else {
				serviceMap["labels"] = map[string]interface{}{
					"dev-stack.profile": "development",
					"dev-stack.service": serviceName,
				}
			}
		}
	}
}

// applyTestTransformations applies test-specific settings
func (c *Composer) applyTestTransformations(composeFile *ComposeFile) {
	for serviceName, service := range composeFile.Services {
		if serviceMap, ok := service.(map[string]interface{}); ok {
			// Add test labels
			if labels, exists := serviceMap["labels"]; exists {
				if labelMap, ok := labels.(map[string]interface{}); ok {
					labelMap["dev-stack.profile"] = "test"
				}
			} else {
				serviceMap["labels"] = map[string]interface{}{
					"dev-stack.profile": "test",
				}
			}

			// Remove unnecessary ports for test environment
			if serviceName != "postgres" && serviceName != "redis" {
				delete(serviceMap, "ports")
			}
		}
	}
}

// applyProductionTransformations applies production-like settings
func (c *Composer) applyProductionTransformations(composeFile *ComposeFile) {
	for _, service := range composeFile.Services {
		if serviceMap, ok := service.(map[string]interface{}); ok {
			// Add production labels
			if labels, exists := serviceMap["labels"]; exists {
				if labelMap, ok := labels.(map[string]interface{}); ok {
					labelMap["dev-stack.profile"] = "production"
				}
			} else {
				serviceMap["labels"] = map[string]interface{}{
					"dev-stack.profile": "production",
				}
			}

			// Add resource limits
			if _, exists := serviceMap["deploy"]; !exists {
				serviceMap["deploy"] = map[string]interface{}{
					"resources": map[string]interface{}{
						"limits": map[string]interface{}{
							"memory": "512M",
						},
						"reservations": map[string]interface{}{
							"memory": "256M",
						},
					},
				}
			}
		}
	}
}

// WriteToFile writes the compose file to disk
func (c *Composer) WriteToFile(composeFile *ComposeFile, outputPath string) error {
	// Check if file exists and overwrite flag
	if _, err := os.Stat(outputPath); err == nil && !c.options.Overwrite {
		return fmt.Errorf("file %s already exists. Use overwrite option to replace it", outputPath)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(composeFile)
	if err != nil {
		return fmt.Errorf("failed to marshal compose file: %w", err)
	}

	// Add header comment
	header := fmt.Sprintf(`# Generated by dev-stack v%s
# Project: %s
# Services: %s
# Profile: %s
# Generated on: %s

`,
		composeFile.Metadata.FrameworkVersion,
		composeFile.Metadata.ProjectName,
		strings.Join(composeFile.Metadata.Services, ", "),
		composeFile.Metadata.Profile,
		composeFile.Metadata.GeneratedAt.Format("2006-01-02 15:04:05"),
	)

	finalContent := header + string(data)

	// Write to file
	if err := os.WriteFile(outputPath, []byte(finalContent), 0644); err != nil {
		return fmt.Errorf("failed to write compose file: %w", err)
	}

	return nil
}

// deepCopyMap creates a deep copy of a map
func (c *Composer) deepCopyMap(original map[string]interface{}) map[string]interface{} {
	copy := make(map[string]interface{})
	for key, value := range original {
		switch v := value.(type) {
		case map[string]interface{}:
			copy[key] = c.deepCopyMap(v)
		case []interface{}:
			copy[key] = c.deepCopySlice(v)
		default:
			copy[key] = value
		}
	}
	return copy
}

// loadProjectConfiguration loads project-specific configuration
func (c *Composer) loadProjectConfiguration() error {
	if c.configLoader == nil {
		return nil
	}

	config, err := c.configLoader.GetConfig()
	if err != nil {
		return err
	}

	c.projectConfig = config

	// Update composer options based on project config
	if config.Project.Name != "" {
		c.options.ProjectName = config.Project.Name
	}

	if config.Networks.Default != "" {
		c.options.NetworkName = config.Networks.Default
	}

	if config.Volumes.Prefix != "" {
		c.options.VolumePrefix = config.Volumes.Prefix
	}

	return nil
}

// applyProjectOverrides applies project-specific service overrides
func (c *Composer) applyProjectOverrides(serviceName string, config map[string]interface{}) error {
	overrides, err := c.configLoader.GetServiceOverrides(c.options.Profile)
	if err != nil {
		return err
	}

	serviceOverride, exists := overrides[serviceName]
	if !exists {
		return nil
	}

	// Apply environment overrides
	if len(serviceOverride.Environment) > 0 {
		c.applyEnvironmentOverrides(config, serviceOverride.Environment)
	}

	// Apply port overrides
	if len(serviceOverride.Ports) > 0 {
		c.applyPortOverrides(config, serviceOverride.Ports)
	}

	// Apply volume overrides
	if len(serviceOverride.Volumes) > 0 {
		c.applyVolumeOverrides(config, serviceOverride.Volumes)
	}

	// Apply label overrides
	if len(serviceOverride.Labels) > 0 {
		c.applyLabelOverrides(config, serviceOverride.Labels)
	}

	// Apply command overrides
	if serviceOverride.Command != "" {
		config["command"] = serviceOverride.Command
	}

	// Apply entrypoint overrides
	if serviceOverride.Entrypoint != "" {
		config["entrypoint"] = serviceOverride.Entrypoint
	}

	// Apply custom overrides
	if len(serviceOverride.Custom) > 0 {
		for key, value := range serviceOverride.Custom {
			config[key] = value
		}
	}

	return nil
}

// applyEnvironmentOverrides applies environment variable overrides
func (c *Composer) applyEnvironmentOverrides(config map[string]interface{}, envOverrides map[string]string) {
	var envList []interface{}
	if env, exists := config["environment"]; exists {
		if envArray, ok := env.([]interface{}); ok {
			envList = envArray
		}
	}

	// Add override environment variables
	for key, value := range envOverrides {
		envVar := fmt.Sprintf("%s=%s", key, value)
		envList = append(envList, envVar)
	}

	if len(envList) > 0 {
		config["environment"] = envList
	}
}

// applyPortOverrides applies port mapping overrides
func (c *Composer) applyPortOverrides(config map[string]interface{}, portOverrides map[string]string) {
	var portsList []interface{}
	if ports, exists := config["ports"]; exists {
		if portsArray, ok := ports.([]interface{}); ok {
			portsList = portsArray
		}
	}

	// Add override ports
	for host, container := range portOverrides {
		portMapping := fmt.Sprintf("%s:%s", host, container)
		portsList = append(portsList, portMapping)
	}

	if len(portsList) > 0 {
		config["ports"] = portsList
	}
}

// applyVolumeOverrides applies volume mount overrides
func (c *Composer) applyVolumeOverrides(config map[string]interface{}, volumeOverrides []string) {
	var volumesList []interface{}
	if volumes, exists := config["volumes"]; exists {
		if volumesArray, ok := volumes.([]interface{}); ok {
			volumesList = volumesArray
		}
	}

	// Add override volumes
	for _, volume := range volumeOverrides {
		volumesList = append(volumesList, volume)
	}

	if len(volumesList) > 0 {
		config["volumes"] = volumesList
	}
}

// applyLabelOverrides applies label overrides
func (c *Composer) applyLabelOverrides(config map[string]interface{}, labelOverrides map[string]string) {
	labels := make(map[string]interface{})
	if existingLabels, exists := config["labels"]; exists {
		if labelMap, ok := existingLabels.(map[string]interface{}); ok {
			labels = labelMap
		}
	}

	// Add override labels
	for key, value := range labelOverrides {
		labels[key] = value
	}

	if len(labels) > 0 {
		config["labels"] = labels
	}
}

// handleConflicts detects and resolves port conflicts
func (c *Composer) handleConflicts(composeFile *ComposeFile) error {
	// Detect conflicts
	conflicts, err := c.conflictDetector.DetectConflicts()
	if err != nil {
		return fmt.Errorf("failed to detect conflicts: %w", err)
	}

	if len(conflicts) > 0 {
		fmt.Printf("🔍 Port conflict analysis:\n%s", c.conflictDetector.GetConflictReport())

		// Auto-resolve if enabled
		if c.options.AutoFixPorts {
			resolutions, err := c.conflictDetector.ResolveConflicts()
			if err != nil {
				return fmt.Errorf("failed to resolve conflicts: %w", err)
			}

			if len(resolutions) > 0 {
				fmt.Printf("🔧 Auto-resolving port conflicts...\n")
				for serviceName, portMappings := range resolutions {
					for oldPort, newPort := range portMappings {
						fmt.Printf("   %s: %d → %d\n", serviceName, oldPort, newPort)
					}
				}

				// Apply resolutions to compose file
				if err := c.applyPortResolutions(composeFile, resolutions); err != nil {
					return fmt.Errorf("failed to apply port resolutions: %w", err)
				}
			}
		} else {
			// Provide manual resolution suggestions
			suggestions := c.conflictDetector.GetSuggestedResolutions()
			if len(suggestions) > 0 {
				fmt.Printf("\n💡 Suggested resolutions:\n")
				for _, suggestionList := range suggestions {
					for _, suggestion := range suggestionList {
						fmt.Printf("   %s\n", suggestion)
					}
					fmt.Println()
				}
			}
		}
	}

	return nil
}

// applyPortResolutions applies port conflict resolutions to the compose file
func (c *Composer) applyPortResolutions(composeFile *ComposeFile, resolutions map[string]map[int]int) error {
	for serviceName, portMappings := range resolutions {
		if serviceConfig, exists := composeFile.Services[serviceName]; exists {
			if serviceMap, ok := serviceConfig.(map[string]interface{}); ok {
				if err := c.updateServicePortsInConfig(serviceMap, portMappings); err != nil {
					return fmt.Errorf("failed to update ports for service %s: %w", serviceName, err)
				}
			}
		}
	}
	return nil
}

// updateServicePortsInConfig updates port mappings in a service configuration
func (c *Composer) updateServicePortsInConfig(serviceConfig map[string]interface{}, portMappings map[int]int) error {
	if portsConfig, exists := serviceConfig["ports"]; exists {
		if portsArray, ok := portsConfig.([]interface{}); ok {
			for i, portEntry := range portsArray {
				if portStr, ok := portEntry.(string); ok {
					updatedPort := c.updatePortString(portStr, portMappings)
					portsArray[i] = updatedPort
				}
			}
		}
	}
	return nil
}

// updatePortString updates a port string with new port mappings
func (c *Composer) updatePortString(portStr string, portMappings map[int]int) string {
	if strings.Contains(portStr, ":") {
		parts := strings.Split(portStr, ":")
		if len(parts) >= 2 {
			hostPortStr := strings.TrimSpace(parts[0])
			if hostPort, err := strconv.Atoi(hostPortStr); err == nil {
				if newPort, exists := portMappings[hostPort]; exists {
					parts[0] = strconv.Itoa(newPort)
					return strings.Join(parts, ":")
				}
			}
		}
	}
	return portStr
}

// findProjectRoot attempts to find the project root directory
func findProjectRoot() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	// Look for dev-stack config files
	configFiles := []string{
		"dev-stack-config.yaml",
		"dev-stack.yaml",
		"dev-stack.yml",
		".dev-stack.yaml",
		".dev-stack.yml",
	}

	dir := cwd
	for {
		for _, configFile := range configFiles {
			configPath := filepath.Join(dir, configFile)
			if _, err := os.Stat(configPath); err == nil {
				return dir
			}
		}

		// Check for services directory
		servicesPath := filepath.Join(dir, "services")
		if stat, err := os.Stat(servicesPath); err == nil && stat.IsDir() {
			return dir
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			break // Reached root
		}
		dir = parent
	}

	return ""
}

// deepCopySlice creates a deep copy of a slice
func (c *Composer) deepCopySlice(original []interface{}) []interface{} {
	copy := make([]interface{}, len(original))
	for i, value := range original {
		switch v := value.(type) {
		case map[string]interface{}:
			copy[i] = c.deepCopyMap(v)
		case []interface{}:
			copy[i] = c.deepCopySlice(v)
		default:
			copy[i] = value
		}
	}
	return copy
}
