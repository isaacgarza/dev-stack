package compose

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// NewServiceRegistry creates a new service registry
func NewServiceRegistry(options RegistryOptions) *ServiceRegistry {
	return &ServiceRegistry{
		services:     make(map[string]*ServiceDefinition),
		manifest:     make(ServicesManifest),
		servicesPath: options.ServicesPath,
		manifestPath: options.ManifestPath,
		loaded:       false,
	}
}

// Load discovers and loads all services from the services directory
func (r *ServiceRegistry) Load() error {
	if err := r.loadManifest(); err != nil {
		return fmt.Errorf("failed to load services manifest: %w", err)
	}

	if err := r.discoverServices(); err != nil {
		return fmt.Errorf("failed to discover services: %w", err)
	}

	r.loaded = true
	r.lastLoaded = time.Now()
	return nil
}

// loadManifest loads the services.yaml manifest file (optional for backward compatibility)
func (r *ServiceRegistry) loadManifest() error {
	if _, err := os.Stat(r.manifestPath); os.IsNotExist(err) {
		// Manifest file is optional, continue without it
		return nil
	}

	data, err := os.ReadFile(r.manifestPath)
	if err != nil {
		// Don't fail if manifest can't be read, just continue without it
		return nil
	}

	if err := yaml.Unmarshal(data, &r.manifest); err != nil {
		// Don't fail if manifest can't be parsed, just continue without it
		return nil
	}

	return nil
}

// discoverServices scans the services directory and loads service definitions
func (r *ServiceRegistry) discoverServices() error {
	if _, err := os.Stat(r.servicesPath); os.IsNotExist(err) {
		return fmt.Errorf("services directory not found: %s\nPlease ensure you're in a dev-stack project directory or the services path is correct", r.servicesPath)
	}

	// Try category-based discovery first
	if err := r.discoverByCategory(); err == nil {
		return nil
	}

	// Fall back to legacy flat discovery
	return r.discoverFlat()
}

// discoverByCategory discovers services organized in category folders
func (r *ServiceRegistry) discoverByCategory() error {
	categories := []string{"database", "cache", "messaging", "observability", "cloud"}
	serviceCount := 0
	errorCount := 0

	for _, category := range categories {
		categoryPath := filepath.Join(r.servicesPath, category)
		if _, err := os.Stat(categoryPath); os.IsNotExist(err) {
			continue
		}

		services, err := r.scanCategory(category)
		if err != nil {
			errorCount++
			fmt.Printf("Warning: failed to scan category '%s': %v\n", category, err)
			continue
		}

		for _, serviceDef := range services {
			r.services[serviceDef.Name] = &serviceDef
			serviceCount++
		}
	}

	if serviceCount == 0 {
		return fmt.Errorf("no services found in category folders")
	}

	if errorCount > 0 {
		fmt.Printf("Loaded %d services successfully, %d categories had errors\n", serviceCount, errorCount)
	}

	return nil
}

// scanCategory scans a specific category folder for service definitions
func (r *ServiceRegistry) scanCategory(category string) ([]ServiceDefinition, error) {
	categoryPath := filepath.Join(r.servicesPath, category)
	entries, err := os.ReadDir(categoryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read category directory %s: %w", categoryPath, err)
	}

	var services []ServiceDefinition

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		serviceName := strings.TrimSuffix(entry.Name(), ".yaml")
		serviceDef, err := r.loadServiceFromFile(filepath.Join(categoryPath, entry.Name()))
		if err != nil {
			fmt.Printf("Warning: failed to load service '%s': %v\n", serviceName, err)
			continue
		}

		serviceDef.Name = serviceName
		serviceDef.Metadata.Category = category
		services = append(services, *serviceDef)
	}

	return services, nil
}

// loadServiceFromFile loads a service definition from a YAML file
func (r *ServiceRegistry) loadServiceFromFile(filePath string) (*ServiceDefinition, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read service file %s: %w", filePath, err)
	}

	var serviceData map[string]interface{}
	if err := yaml.Unmarshal(data, &serviceData); err != nil {
		return nil, fmt.Errorf("invalid YAML in %s: %w", filePath, err)
	}

	// Extract basic service info
	name, _ := serviceData["name"].(string)
	description, _ := serviceData["description"].(string)
	category, _ := serviceData["category"].(string)

	// Build metadata from the service file
	metadata := ServiceMetadata{
		Description: description,
		Category:    category,
	}

	// Extract CLI metadata
	if options, exists := serviceData["options"]; exists {
		if optionsList, ok := options.([]interface{}); ok {
			for _, opt := range optionsList {
				if optStr, ok := opt.(string); ok {
					metadata.Options = append(metadata.Options, optStr)
				}
			}
		}
	}

	if examples, exists := serviceData["examples"]; exists {
		if examplesList, ok := examples.([]interface{}); ok {
			for _, ex := range examplesList {
				if exStr, ok := ex.(string); ok {
					metadata.Examples = append(metadata.Examples, exStr)
				}
			}
		}
	}

	if usageNotes, exists := serviceData["usage_notes"]; exists {
		if notesStr, ok := usageNotes.(string); ok {
			metadata.UsageNotes = notesStr
		}
	}

	if links, exists := serviceData["links"]; exists {
		if linksList, ok := links.([]interface{}); ok {
			for _, link := range linksList {
				if linkStr, ok := link.(string); ok {
					metadata.Links = append(metadata.Links, linkStr)
				}
			}
		}
	}

	// Extract dependency configuration
	if deps, exists := serviceData["dependencies"]; exists {
		if depsMap, ok := deps.(map[string]interface{}); ok {
			if required, exists := depsMap["required"]; exists {
				if reqList, ok := required.([]interface{}); ok {
					for _, req := range reqList {
						if reqStr, ok := req.(string); ok {
							metadata.DependencyConfig.Required = append(metadata.DependencyConfig.Required, reqStr)
						}
					}
				}
			}
			if soft, exists := depsMap["soft"]; exists {
				if softList, ok := soft.([]interface{}); ok {
					for _, s := range softList {
						if sStr, ok := s.(string); ok {
							metadata.DependencyConfig.Soft = append(metadata.DependencyConfig.Soft, sStr)
						}
					}
				}
			}
			if conflicts, exists := depsMap["conflicts"]; exists {
				if conflictsList, ok := conflicts.([]interface{}); ok {
					for _, c := range conflictsList {
						if cStr, ok := c.(string); ok {
							metadata.DependencyConfig.Conflicts = append(metadata.DependencyConfig.Conflicts, cStr)
						}
					}
				}
			}
			if provides, exists := depsMap["provides"]; exists {
				if providesList, ok := provides.([]interface{}); ok {
					for _, p := range providesList {
						if pStr, ok := p.(string); ok {
							metadata.DependencyConfig.Provides = append(metadata.DependencyConfig.Provides, pStr)
						}
					}
				}
			}
		}
	}

	// For now, create empty docker-compose structure
	// The actual docker-compose generation will use the service.yaml data
	services := make(map[string]interface{})
	volumes := make(map[string]interface{})
	networks := make(map[string]interface{})

	// Extract dependencies (legacy + new format)
	dependencies := r.extractDependenciesFromMetadata(metadata)

	serviceDef := &ServiceDefinition{
		Name:         name,
		Path:         filepath.Dir(filePath),
		ComposeFile:  filePath,
		Services:     services,
		Volumes:      volumes,
		Networks:     networks,
		Dependencies: dependencies,
		Metadata:     metadata,
		RawContent:   data,
	}

	return serviceDef, nil
}

// extractDependenciesFromMetadata extracts dependencies from service metadata
func (r *ServiceRegistry) extractDependenciesFromMetadata(metadata ServiceMetadata) []string {
	var deps []string
	
	// Add required dependencies
	deps = append(deps, metadata.DependencyConfig.Required...)
	
	// Add legacy dependencies for backward compatibility
	deps = append(deps, metadata.Dependencies...)
	
	return deps
}

// discoverFlat performs legacy flat directory discovery
func (r *ServiceRegistry) discoverFlat() error {
	entries, err := os.ReadDir(r.servicesPath)
	if err != nil {
		return fmt.Errorf("failed to read services directory %s: %w", r.servicesPath, err)
	}

	serviceCount := 0
	errorCount := 0

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		serviceName := entry.Name()
		// Skip the services.yaml file and any hidden directories
		if serviceName == "services.yaml" || strings.HasPrefix(serviceName, ".") {
			continue
		}

		if err := r.loadService(serviceName); err != nil {
			errorCount++
			fmt.Printf("Warning: failed to load service '%s': %v\n", serviceName, err)
			continue
		}
		serviceCount++
	}

	if serviceCount == 0 {
		return fmt.Errorf("no valid services found in %s\nEnsure each service directory contains a docker-compose.yml file", r.servicesPath)
	}

	if errorCount > 0 {
		fmt.Printf("Loaded %d services successfully, %d services had errors\n", serviceCount, errorCount)
	}

	return nil
}

// loadService loads a specific service definition
func (r *ServiceRegistry) loadService(serviceName string) error {
	servicePath := filepath.Join(r.servicesPath, serviceName)
	composePath := filepath.Join(servicePath, "docker-compose.yml")

	// Check if docker-compose.yml exists
	if _, err := os.Stat(composePath); os.IsNotExist(err) {
		return fmt.Errorf("docker-compose.yml not found in %s\nEach service directory must contain a docker-compose.yml file", servicePath)
	}

	// Read the docker-compose.yml file
	data, err := os.ReadFile(composePath)
	if err != nil {
		return fmt.Errorf("failed to read docker-compose.yml in %s: %w", servicePath, err)
	}

	// Parse the YAML content
	var composeContent map[string]interface{}
	if err := yaml.Unmarshal(data, &composeContent); err != nil {
		return fmt.Errorf("invalid YAML in %s: %w\nPlease check the docker-compose.yml syntax", composePath, err)
	}

	// Extract sections
	services := make(map[string]interface{})
	volumes := make(map[string]interface{})
	networks := make(map[string]interface{})

	if servicesSection, ok := composeContent["services"].(map[string]interface{}); ok {
		services = servicesSection
	}

	if volumesSection, ok := composeContent["volumes"].(map[string]interface{}); ok {
		volumes = volumesSection
	}

	if networksSection, ok := composeContent["networks"].(map[string]interface{}); ok {
		networks = networksSection
	}

	// Get metadata from manifest
	metadata := ServiceMetadata{}
	if manifestEntry, exists := r.manifest[serviceName]; exists {
		metadata = manifestEntry
	}

	// Extract dependencies from docker-compose.yml
	dependencies := r.extractDependencies(services)

	// Create service definition
	serviceDef := &ServiceDefinition{
		Name:         serviceName,
		Path:         servicePath,
		ComposeFile:  composePath,
		Services:     services,
		Volumes:      volumes,
		Networks:     networks,
		Dependencies: dependencies,
		Metadata:     metadata,
		RawContent:   data,
	}

	r.services[serviceName] = serviceDef
	return nil
}

// extractDependencies extracts external service dependencies from docker-compose services
func (r *ServiceRegistry) extractDependencies(services map[string]interface{}) []string {
	var deps []string
	depSet := make(map[string]bool)

	// Get list of services defined in this compose file
	internalServices := make(map[string]bool)
	for serviceName := range services {
		internalServices[serviceName] = true
	}

	for _, serviceConfig := range services {
		if serviceMap, ok := serviceConfig.(map[string]interface{}); ok {
			// Check depends_on
			if dependsOn, exists := serviceMap["depends_on"]; exists {
				switch dependsOn := dependsOn.(type) {
				case []interface{}:
					for _, dep := range dependsOn {
						if depStr, ok := dep.(string); ok {
							// Only add external dependencies (not internal to this service)
							if !internalServices[depStr] && !depSet[depStr] {
								deps = append(deps, depStr)
								depSet[depStr] = true
							}
						}
					}
				case map[string]interface{}:
					for depName := range dependsOn {
						// Only add external dependencies (not internal to this service)
						if !internalServices[depName] && !depSet[depName] {
							deps = append(deps, depName)
							depSet[depName] = true
						}
					}
				}
			}
		}
	}

	return deps
}

// GetService returns a service definition by name
func (r *ServiceRegistry) GetService(name string) (*ServiceDefinition, error) {
	if !r.loaded {
		if err := r.Load(); err != nil {
			return nil, err
		}
	}

	service, exists := r.services[name]
	if !exists {
		return nil, fmt.Errorf("service %s not found", name)
	}

	return service, nil
}

// GetServices returns multiple service definitions by name
func (r *ServiceRegistry) GetServices(names []string) ([]*ServiceDefinition, error) {
	var services []*ServiceDefinition

	for _, name := range names {
		service, err := r.GetService(name)
		if err != nil {
			return nil, err
		}
		services = append(services, service)
	}

	return services, nil
}

// ListServices returns all available service names
func (r *ServiceRegistry) ListServices() []string {
	if !r.loaded {
		if err := r.Load(); err != nil {
			return []string{}
		}
	}

	var names []string
	for name := range r.services {
		names = append(names, name)
	}
	return names
}

// GetServicesByCategory returns all services in a specific category
func (r *ServiceRegistry) GetServicesByCategory(category string) ([]*ServiceDefinition, error) {
	if !r.loaded {
		if err := r.Load(); err != nil {
			return nil, err
		}
	}

	var services []*ServiceDefinition
	for _, service := range r.services {
		if service.Metadata.Category == category {
			services = append(services, service)
		}
	}

	return services, nil
}

// ListCategories returns all available categories
func (r *ServiceRegistry) ListCategories() []string {
	if !r.loaded {
		if err := r.Load(); err != nil {
			return []string{}
		}
	}

	categorySet := make(map[string]bool)
	for _, service := range r.services {
		if service.Metadata.Category != "" {
			categorySet[service.Metadata.Category] = true
		}
	}

	var categories []string
	for category := range categorySet {
		categories = append(categories, category)
	}

	return categories
}

// ResolveDependencies resolves service dependencies and returns an ordered list
func (r *ServiceRegistry) ResolveDependencies(serviceNames []string) ([]string, error) {
	if !r.loaded {
		if err := r.Load(); err != nil {
			return nil, err
		}
	}

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

		service, exists := r.services[serviceName]
		if !exists {
			return fmt.Errorf("service %s not found", serviceName)
		}

		visiting[serviceName] = true

		// Visit dependencies first
		for _, dep := range service.Dependencies {
			if err := visit(dep); err != nil {
				return err
			}
		}

		visiting[serviceName] = false
		visited[serviceName] = true
		result = append(result, serviceName)

		return nil
	}

	// Visit all requested services
	for _, serviceName := range serviceNames {
		if err := visit(serviceName); err != nil {
			return nil, err
		}
	}

	return result, nil
}

// ValidateService validates a service definition
func (r *ServiceRegistry) ValidateService(serviceName string) ValidationResult {
	result := ValidationResult{
		Service:  serviceName,
		Valid:    true,
		Errors:   []string{},
		Warnings: []string{},
	}

	service, err := r.GetService(serviceName)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("service not found: %s", err.Error()))
		return result
	}

	// Validate service has at least one service definition
	if len(service.Services) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, "no services defined in docker-compose.yml - ensure the 'services:' section contains at least one service")
	}

	// Validate each service definition has required fields
	for serviceName, serviceConfig := range service.Services {
		if serviceMap, ok := serviceConfig.(map[string]interface{}); ok {
			// Check for image or build
			if _, hasImage := serviceMap["image"]; !hasImage {
				if _, hasBuild := serviceMap["build"]; !hasBuild {
					result.Errors = append(result.Errors, fmt.Sprintf("service '%s' missing 'image' or 'build' configuration", serviceName))
					result.Valid = false
				}
			}
		}
	}

	// Validate dependencies exist
	for _, dep := range service.Dependencies {
		if _, exists := r.services[dep]; !exists {
			result.Warnings = append(result.Warnings, fmt.Sprintf("external dependency '%s' not available in current registry", dep))
		}
	}

	// Check for required environment variables
	for _, serviceConfig := range service.Services {
		if serviceMap, ok := serviceConfig.(map[string]interface{}); ok {
			if env, exists := serviceMap["environment"]; exists {
				r.validateEnvironmentVariables(env, &result)
			}
		}
	}

	return result
}

// validateEnvironmentVariables checks environment variable definitions
func (r *ServiceRegistry) validateEnvironmentVariables(env interface{}, result *ValidationResult) {
	switch env := env.(type) {
	case []interface{}:
		for _, envVar := range env {
			if envStr, ok := envVar.(string); ok {
				if strings.Contains(envStr, "${") && !strings.Contains(envStr, ":-") {
					result.Warnings = append(result.Warnings,
						fmt.Sprintf("environment variable %s has no default value", envStr))
				}
			}
		}
	case map[string]interface{}:
		for key, value := range env {
			if valueStr, ok := value.(string); ok {
				if strings.Contains(valueStr, "${") && !strings.Contains(valueStr, ":-") {
					result.Warnings = append(result.Warnings,
						fmt.Sprintf("environment variable %s=%s has no default value", key, valueStr))
				}
			}
		}
	}
}

// Reload reloads all service definitions
func (r *ServiceRegistry) Reload() error {
	r.services = make(map[string]*ServiceDefinition)
	r.manifest = make(ServicesManifest)
	r.loaded = false
	return r.Load()
}

// GetManifest returns the services manifest
func (r *ServiceRegistry) GetManifest() ServicesManifest {
	return r.manifest
}

// IsLoaded returns whether the registry has been loaded
func (r *ServiceRegistry) IsLoaded() bool {
	return r.loaded
}

// LastLoaded returns when the registry was last loaded
func (r *ServiceRegistry) LastLoaded() time.Time {
	return r.lastLoaded
}
