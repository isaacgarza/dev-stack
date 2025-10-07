package services

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// ServiceRegistry manages service definitions and validation
type ServiceRegistry struct {
	services   map[string]ServiceDefinition
	configPath string
}

// ServiceDefinition represents a complete service definition from services.yaml
type ServiceDefinition struct {
	Description string   `yaml:"description"`
	Options     []string `yaml:"options"`
	Examples    []string `yaml:"examples"`
	UsageNotes  string   `yaml:"usage_notes"`
	Links       []string `yaml:"links"`

	// Extended properties
	Category     string            `yaml:"category,omitempty"`
	DefaultPort  int               `yaml:"default_port,omitempty"`
	HealthCheck  HealthCheckConfig `yaml:"health_check,omitempty"`
	Dependencies []string          `yaml:"dependencies,omitempty"`
	Tags         []string          `yaml:"tags,omitempty"`
}

// HealthCheckConfig represents health check configuration
type HealthCheckConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Endpoint string `yaml:"endpoint,omitempty"`
	Interval string `yaml:"interval,omitempty"`
	Timeout  string `yaml:"timeout,omitempty"`
	Retries  int    `yaml:"retries,omitempty"`
}

// ServiceManifest represents the structure of services.yaml
type ServiceManifest map[string]ServiceDefinition

// NewServiceRegistry creates a new service registry
func NewServiceRegistry(configPath string) (*ServiceRegistry, error) {
	registry := &ServiceRegistry{
		services:   make(map[string]ServiceDefinition),
		configPath: configPath,
	}

	if err := registry.Load(); err != nil {
		return nil, fmt.Errorf("failed to load service registry: %w", err)
	}

	return registry, nil
}

// Load loads services from the configuration file
func (r *ServiceRegistry) Load() error {
	// Resolve config path
	configPath, err := r.resolveConfigPath()
	if err != nil {
		return fmt.Errorf("failed to resolve config path: %w", err)
	}

	// Read the YAML file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read services file %s: %w", configPath, err)
	}

	// Parse YAML
	var manifest ServiceManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("failed to parse services YAML: %w", err)
	}

	// Validate and store services
	for name, definition := range manifest {
		if err := r.validateServiceDefinition(name, definition); err != nil {
			return fmt.Errorf("invalid service definition for %s: %w", name, err)
		}
		r.services[name] = definition
	}

	return nil
}

// Reload reloads the service registry from the configuration file
func (r *ServiceRegistry) Reload() error {
	r.services = make(map[string]ServiceDefinition)
	return r.Load()
}

// GetService returns a service definition by name
func (r *ServiceRegistry) GetService(name string) (ServiceDefinition, bool) {
	service, exists := r.services[name]
	return service, exists
}

// GetAllServices returns all service definitions
func (r *ServiceRegistry) GetAllServices() map[string]ServiceDefinition {
	// Return a copy to prevent modification
	result := make(map[string]ServiceDefinition)
	for name, service := range r.services {
		result[name] = service
	}
	return result
}

// GetServiceNames returns all service names
func (r *ServiceRegistry) GetServiceNames() []string {
	names := make([]string, 0, len(r.services))
	for name := range r.services {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// GetServicesByCategory returns services in a specific category
func (r *ServiceRegistry) GetServicesByCategory(category string) []string {
	var services []string
	for name, service := range r.services {
		if service.Category == category {
			services = append(services, name)
		}
	}
	sort.Strings(services)
	return services
}

// GetServicesByTag returns services with a specific tag
func (r *ServiceRegistry) GetServicesByTag(tag string) []string {
	var services []string
	for name, service := range r.services {
		for _, serviceTag := range service.Tags {
			if serviceTag == tag {
				services = append(services, name)
				break
			}
		}
	}
	sort.Strings(services)
	return services
}

// ValidateService validates that a service exists
func (r *ServiceRegistry) ValidateService(name string) error {
	if _, exists := r.services[name]; !exists {
		available := r.GetServiceNames()
		return fmt.Errorf("unknown service '%s'. Available services: %v", name, available)
	}
	return nil
}

// ValidateServices validates multiple service names
func (r *ServiceRegistry) ValidateServices(names []string) error {
	for _, name := range names {
		if err := r.ValidateService(name); err != nil {
			return err
		}
	}
	return nil
}

// GetServiceDependencies returns the dependencies for a service
func (r *ServiceRegistry) GetServiceDependencies(name string) ([]string, error) {
	service, exists := r.GetService(name)
	if !exists {
		return nil, fmt.Errorf("service %s not found", name)
	}
	return service.Dependencies, nil
}

// GetAllCategories returns all unique service categories
func (r *ServiceRegistry) GetAllCategories() []string {
	categories := make(map[string]bool)
	for _, service := range r.services {
		if service.Category != "" {
			categories[service.Category] = true
		}
	}

	result := make([]string, 0, len(categories))
	for category := range categories {
		result = append(result, category)
	}
	sort.Strings(result)
	return result
}

// GetAllTags returns all unique service tags
func (r *ServiceRegistry) GetAllTags() []string {
	tags := make(map[string]bool)
	for _, service := range r.services {
		for _, tag := range service.Tags {
			tags[tag] = true
		}
	}

	result := make([]string, 0, len(tags))
	for tag := range tags {
		result = append(result, tag)
	}
	sort.Strings(result)
	return result
}

// SearchServices searches for services by name, description, or tags
func (r *ServiceRegistry) SearchServices(query string) []string {
	var matches []string
	query = strings.ToLower(query)

	for name, service := range r.services {
		// Check name
		if strings.Contains(strings.ToLower(name), query) {
			matches = append(matches, name)
			continue
		}

		// Check description
		if strings.Contains(strings.ToLower(service.Description), query) {
			matches = append(matches, name)
			continue
		}

		// Check tags
		for _, tag := range service.Tags {
			if strings.Contains(strings.ToLower(tag), query) {
				matches = append(matches, name)
				break
			}
		}
	}

	sort.Strings(matches)
	return matches
}

// GetServiceInfo returns formatted service information
func (r *ServiceRegistry) GetServiceInfo(name string) (string, error) {
	service, exists := r.GetService(name)
	if !exists {
		return "", fmt.Errorf("service %s not found", name)
	}

	var info strings.Builder
	info.WriteString(fmt.Sprintf("Service: %s\n", name))
	info.WriteString(fmt.Sprintf("Description: %s\n", service.Description))

	if service.Category != "" {
		info.WriteString(fmt.Sprintf("Category: %s\n", service.Category))
	}

	if service.DefaultPort > 0 {
		info.WriteString(fmt.Sprintf("Default Port: %d\n", service.DefaultPort))
	}

	if len(service.Dependencies) > 0 {
		info.WriteString(fmt.Sprintf("Dependencies: %s\n", strings.Join(service.Dependencies, ", ")))
	}

	if len(service.Tags) > 0 {
		info.WriteString(fmt.Sprintf("Tags: %s\n", strings.Join(service.Tags, ", ")))
	}

	if len(service.Options) > 0 {
		info.WriteString("\nConfiguration Options:\n")
		for _, option := range service.Options {
			info.WriteString(fmt.Sprintf("  - %s\n", option))
		}
	}

	if len(service.Examples) > 0 {
		info.WriteString("\nExamples:\n")
		for _, example := range service.Examples {
			info.WriteString(fmt.Sprintf("  %s\n", example))
		}
	}

	if service.UsageNotes != "" {
		info.WriteString(fmt.Sprintf("\nUsage Notes:\n%s\n", service.UsageNotes))
	}

	if len(service.Links) > 0 {
		info.WriteString("\nLinks:\n")
		for _, link := range service.Links {
			info.WriteString(fmt.Sprintf("  - %s\n", link))
		}
	}

	return info.String(), nil
}

// resolveConfigPath resolves the configuration file path
func (r *ServiceRegistry) resolveConfigPath() (string, error) {
	if r.configPath == "" {
		// Try default locations
		candidates := []string{
			"services/services.yaml",
			"config/services.yaml",
			".dev-stack/services.yaml",
		}

		for _, candidate := range candidates {
			if _, err := os.Stat(candidate); err == nil {
				return filepath.Abs(candidate)
			}
		}

		return "", fmt.Errorf("no services.yaml found in default locations: %v", candidates)
	}

	// Use provided path
	if !filepath.IsAbs(r.configPath) {
		return filepath.Abs(r.configPath)
	}

	return r.configPath, nil
}

// validateServiceDefinition validates a service definition
func (r *ServiceRegistry) validateServiceDefinition(name string, definition ServiceDefinition) error {
	if definition.Description == "" {
		return fmt.Errorf("description is required")
	}

	// Validate port if specified
	if definition.DefaultPort < 0 || definition.DefaultPort > 65535 {
		return fmt.Errorf("invalid default port: %d", definition.DefaultPort)
	}

	// Validate health check configuration
	if definition.HealthCheck.Enabled {
		if definition.HealthCheck.Endpoint == "" {
			return fmt.Errorf("health check endpoint is required when health check is enabled")
		}
	}

	return nil
}

// LoadDefault loads the service registry from the default location
func LoadDefaultServiceRegistry() (*ServiceRegistry, error) {
	return NewServiceRegistry("")
}

// LoadFromPath loads the service registry from a specific path
func LoadServiceRegistryFromPath(path string) (*ServiceRegistry, error) {
	return NewServiceRegistry(path)
}
