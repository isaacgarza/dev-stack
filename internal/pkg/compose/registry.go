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

// loadManifest loads the services.yaml manifest file
func (r *ServiceRegistry) loadManifest() error {
	if _, err := os.Stat(r.manifestPath); os.IsNotExist(err) {
		// Manifest file is optional, continue without it
		return nil
	}

	data, err := os.ReadFile(r.manifestPath)
	if err != nil {
		return fmt.Errorf("failed to read manifest file %s: %w", r.manifestPath, err)
	}

	if err := yaml.Unmarshal(data, &r.manifest); err != nil {
		return fmt.Errorf("failed to parse manifest file %s: %w", r.manifestPath, err)
	}

	return nil
}

// discoverServices scans the services directory and loads service definitions
func (r *ServiceRegistry) discoverServices() error {
	if _, err := os.Stat(r.servicesPath); os.IsNotExist(err) {
		return fmt.Errorf("services directory not found: %s\nPlease ensure you're in a dev-stack project directory or the services path is correct", r.servicesPath)
	}

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
