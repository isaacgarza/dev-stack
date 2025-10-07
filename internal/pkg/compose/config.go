package compose

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ConfigLoader handles loading and parsing project-specific configuration
type ConfigLoader struct {
	projectRoot string
	configFile  string
	loaded      bool
	config      *ProjectConfig
}

// ProjectConfig represents the complete project configuration
type ProjectConfig struct {
	Project   ProjectInfo        `yaml:"project"`
	Services  ServicesConfig     `yaml:"services"`
	Overrides ProjectOverrides   `yaml:"overrides"`
	Profiles  map[string]Profile `yaml:"profiles"`
	Networks  NetworksConfig     `yaml:"networks"`
	Volumes   VolumesConfig      `yaml:"volumes"`
}

// ProjectInfo contains basic project information
type ProjectInfo struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Version     string            `yaml:"version"`
	Environment string            `yaml:"environment"`
	Tags        []string          `yaml:"tags"`
	Metadata    map[string]string `yaml:"metadata"`
}

// ServicesConfig contains service-specific configuration
type ServicesConfig struct {
	Default  []string                   `yaml:"default"`
	Required []string                   `yaml:"required"`
	Optional []string                   `yaml:"optional"`
	Disabled []string                   `yaml:"disabled"`
	Custom   map[string]ServiceOverride `yaml:"custom"`
}

// Profile defines environment-specific configurations
type Profile struct {
	Name         string                     `yaml:"name"`
	Description  string                     `yaml:"description"`
	Services     ServicesConfig             `yaml:"services"`
	Environment  map[string]string          `yaml:"environment"`
	Overrides    map[string]ServiceOverride `yaml:"overrides"`
	Resources    bool                       `yaml:"resources"`
	HealthChecks bool                       `yaml:"health_checks"`
	Volumes      bool                       `yaml:"volumes"`
}

// NetworksConfig contains network configuration
type NetworksConfig struct {
	Default string                   `yaml:"default"`
	Custom  map[string]NetworkConfig `yaml:"custom"`
}

// VolumesConfig contains volume configuration
type VolumesConfig struct {
	Prefix string                  `yaml:"prefix"`
	Custom map[string]VolumeConfig `yaml:"custom"`
}

// NewConfigLoader creates a new configuration loader
func NewConfigLoader(projectRoot string) *ConfigLoader {
	configFile := filepath.Join(projectRoot, "dev-stack-config.yaml")

	// Check for alternative config file names
	alternatives := []string{
		"dev-stack.yaml",
		"dev-stack.yml",
		".dev-stack.yaml",
		".dev-stack.yml",
	}

	for _, alt := range alternatives {
		altPath := filepath.Join(projectRoot, alt)
		if _, err := os.Stat(altPath); err == nil {
			configFile = altPath
			break
		}
	}

	return &ConfigLoader{
		projectRoot: projectRoot,
		configFile:  configFile,
		loaded:      false,
	}
}

// Load loads the project configuration from the config file
func (c *ConfigLoader) Load() error {
	// Check if config file exists
	if _, err := os.Stat(c.configFile); os.IsNotExist(err) {
		// Create default configuration
		c.config = c.defaultConfig()
		c.loaded = true
		return nil
	}

	// Read config file
	data, err := os.ReadFile(c.configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", c.configFile, err)
	}

	// Parse YAML
	config := &ProjectConfig{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to parse config file %s: %w", c.configFile, err)
	}

	// Apply defaults for missing sections
	c.applyDefaults(config)

	c.config = config
	c.loaded = true
	return nil
}

// GetConfig returns the loaded configuration
func (c *ConfigLoader) GetConfig() (*ProjectConfig, error) {
	if !c.loaded {
		if err := c.Load(); err != nil {
			return nil, err
		}
	}
	return c.config, nil
}

// GetProfile returns a specific profile configuration
func (c *ConfigLoader) GetProfile(profileName string) (*Profile, error) {
	config, err := c.GetConfig()
	if err != nil {
		return nil, err
	}

	if profile, exists := config.Profiles[profileName]; exists {
		return &profile, nil
	}

	// Return default profile if specific one not found
	return c.getDefaultProfile(profileName), nil
}

// GetServiceOverrides returns service overrides for a specific profile
func (c *ConfigLoader) GetServiceOverrides(profileName string) (map[string]ServiceOverride, error) {
	config, err := c.GetConfig()
	if err != nil {
		return nil, err
	}

	overrides := make(map[string]ServiceOverride)

	// Apply global overrides
	for serviceName, override := range config.Overrides.Services {
		overrides[serviceName] = override
	}

	// Apply profile-specific overrides
	if profile, exists := config.Profiles[profileName]; exists {
		for serviceName, override := range profile.Overrides {
			// Merge with existing overrides
			if existing, existsInGlobal := overrides[serviceName]; existsInGlobal {
				merged := c.mergeServiceOverrides(existing, override)
				overrides[serviceName] = merged
			} else {
				overrides[serviceName] = override
			}
		}
	}

	return overrides, nil
}

// GetServicesForProfile returns the list of services for a specific profile
func (c *ConfigLoader) GetServicesForProfile(profileName string) ([]string, error) {
	config, err := c.GetConfig()
	if err != nil {
		return nil, err
	}

	var services []string

	// Start with default services
	services = append(services, config.Services.Default...)

	// Add required services
	services = append(services, config.Services.Required...)

	// Add profile-specific services
	if profile, exists := config.Profiles[profileName]; exists {
		services = append(services, profile.Services.Default...)
		services = append(services, profile.Services.Required...)
	}

	// Remove disabled services
	disabled := make(map[string]bool)
	for _, service := range config.Services.Disabled {
		disabled[service] = true
	}

	if profile, exists := config.Profiles[profileName]; exists {
		for _, service := range profile.Services.Disabled {
			disabled[service] = true
		}
	}

	// Filter out disabled services and deduplicate
	seen := make(map[string]bool)
	var result []string
	for _, service := range services {
		if !disabled[service] && !seen[service] {
			result = append(result, service)
			seen[service] = true
		}
	}

	return result, nil
}

// GetGlobalEnvironment returns global environment variables for a profile
func (c *ConfigLoader) GetGlobalEnvironment(profileName string) (map[string]string, error) {
	config, err := c.GetConfig()
	if err != nil {
		return nil, err
	}

	env := make(map[string]string)

	// Apply global environment
	for key, value := range config.Overrides.Global.Environment {
		env[key] = value
	}

	// Apply profile-specific environment
	if profile, exists := config.Profiles[profileName]; exists {
		for key, value := range profile.Environment {
			env[key] = value
		}
	}

	// Add project-specific environment
	env["PROJECT_NAME"] = config.Project.Name
	env["PROJECT_VERSION"] = config.Project.Version
	env["PROJECT_ENVIRONMENT"] = config.Project.Environment

	return env, nil
}

// ValidateConfig validates the loaded configuration
func (c *ConfigLoader) ValidateConfig() (*ConfigValidation, error) {
	config, err := c.GetConfig()
	if err != nil {
		return nil, err
	}

	validation := &ConfigValidation{
		Valid:         true,
		Errors:        []string{},
		Warnings:      []string{},
		PortConflicts: []PortConflict{},
		ServiceIssues: make(map[string]ValidationResult),
	}

	// Validate project name
	if config.Project.Name == "" {
		validation.Errors = append(validation.Errors, "project name is required")
		validation.Valid = false
	} else if !isValidProjectName(config.Project.Name) {
		validation.Errors = append(validation.Errors, "project name contains invalid characters")
		validation.Valid = false
	}

	// Validate service configurations
	for serviceName, serviceConfig := range config.Services.Custom {
		if err := c.validateServiceOverride(serviceName, serviceConfig); err != nil {
			validation.Errors = append(validation.Errors,
				fmt.Sprintf("service %s configuration invalid: %s", serviceName, err.Error()))
			validation.Valid = false
		}
	}

	// Validate profiles
	for profileName, profile := range config.Profiles {
		if err := c.validateProfile(profileName, profile); err != nil {
			validation.Errors = append(validation.Errors,
				fmt.Sprintf("profile %s invalid: %s", profileName, err.Error()))
			validation.Valid = false
		}
	}

	return validation, nil
}

// Save saves the current configuration to the config file
func (c *ConfigLoader) Save() error {
	if c.config == nil {
		return fmt.Errorf("no configuration loaded")
	}

	data, err := yaml.Marshal(c.config)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	// Add header comment
	header := `# dev-stack project configuration
# This file contains project-specific settings for the dev-stack framework
# Generated automatically - you can edit this file to customize your development environment

`

	content := header + string(data)

	if err := os.WriteFile(c.configFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// defaultConfig returns a default configuration
func (c *ConfigLoader) defaultConfig() *ProjectConfig {
	projectName := filepath.Base(c.projectRoot)
	if projectName == "." {
		projectName = "dev-stack"
	}

	return &ProjectConfig{
		Project: ProjectInfo{
			Name:        projectName,
			Description: fmt.Sprintf("Development environment for %s", projectName),
			Version:     "1.0.0",
			Environment: "development",
			Tags:        []string{"development"},
			Metadata:    make(map[string]string),
		},
		Services: ServicesConfig{
			Default:  []string{"postgres", "redis"},
			Required: []string{},
			Optional: []string{},
			Disabled: []string{},
			Custom:   make(map[string]ServiceOverride),
		},
		Overrides: ProjectOverrides{
			Services: make(map[string]ServiceOverride),
			Global: GlobalOverrides{
				Environment:    make(map[string]string),
				NetworkName:    fmt.Sprintf("%s-network", projectName),
				VolumePrefix:   projectName,
				ResourceLimits: false,
				HealthChecks:   true,
				RestartPolicy:  "unless-stopped",
			},
		},
		Profiles: map[string]Profile{
			"dev": {
				Name:         "development",
				Description:  "Development environment with debugging enabled",
				Resources:    false,
				HealthChecks: true,
				Volumes:      true,
				Environment: map[string]string{
					"LOG_LEVEL": "DEBUG",
					"DEBUG":     "true",
				},
			},
			"test": {
				Name:         "test",
				Description:  "Test environment for running automated tests",
				Resources:    true,
				HealthChecks: true,
				Volumes:      false,
				Environment: map[string]string{
					"LOG_LEVEL": "INFO",
					"TEST_MODE": "true",
				},
			},
			"prod": {
				Name:         "production",
				Description:  "Production-like environment with resource limits",
				Resources:    true,
				HealthChecks: true,
				Volumes:      true,
				Environment: map[string]string{
					"LOG_LEVEL": "WARN",
				},
			},
		},
		Networks: NetworksConfig{
			Default: fmt.Sprintf("%s-network", projectName),
			Custom:  make(map[string]NetworkConfig),
		},
		Volumes: VolumesConfig{
			Prefix: projectName,
			Custom: make(map[string]VolumeConfig),
		},
	}
}

// applyDefaults applies default values to missing configuration sections
func (c *ConfigLoader) applyDefaults(config *ProjectConfig) {
	if config.Project.Name == "" {
		config.Project.Name = filepath.Base(c.projectRoot)
	}

	if config.Services.Custom == nil {
		config.Services.Custom = make(map[string]ServiceOverride)
	}

	if config.Overrides.Services == nil {
		config.Overrides.Services = make(map[string]ServiceOverride)
	}

	if config.Overrides.Global.Environment == nil {
		config.Overrides.Global.Environment = make(map[string]string)
	}

	if config.Profiles == nil {
		config.Profiles = make(map[string]Profile)
	}
}

// getDefaultProfile returns a default profile configuration
func (c *ConfigLoader) getDefaultProfile(profileName string) *Profile {
	switch profileName {
	case "test":
		return &Profile{
			Name:         "test",
			Description:  "Test environment",
			Resources:    true,
			HealthChecks: true,
			Volumes:      false,
		}
	case "prod", "production":
		return &Profile{
			Name:         "production",
			Description:  "Production-like environment",
			Resources:    true,
			HealthChecks: true,
			Volumes:      true,
		}
	default:
		return &Profile{
			Name:         "development",
			Description:  "Development environment",
			Resources:    false,
			HealthChecks: true,
			Volumes:      true,
		}
	}
}

// mergeServiceOverrides merges two service override configurations
func (c *ConfigLoader) mergeServiceOverrides(base, override ServiceOverride) ServiceOverride {
	result := base

	// Merge environment variables
	if override.Environment != nil {
		if result.Environment == nil {
			result.Environment = make(map[string]string)
		}
		for key, value := range override.Environment {
			result.Environment[key] = value
		}
	}

	// Merge ports
	if override.Ports != nil {
		if result.Ports == nil {
			result.Ports = make(map[string]string)
		}
		for key, value := range override.Ports {
			result.Ports[key] = value
		}
	}

	// Merge labels
	if override.Labels != nil {
		if result.Labels == nil {
			result.Labels = make(map[string]string)
		}
		for key, value := range override.Labels {
			result.Labels[key] = value
		}
	}

	// Override other fields
	if override.Command != "" {
		result.Command = override.Command
	}
	if override.Entrypoint != "" {
		result.Entrypoint = override.Entrypoint
	}
	if override.Enabled != nil {
		result.Enabled = override.Enabled
	}
	if override.Profile != "" {
		result.Profile = override.Profile
	}

	// Merge volumes (append)
	if override.Volumes != nil {
		result.Volumes = append(result.Volumes, override.Volumes...)
	}

	// Merge networks (append)
	if override.Networks != nil {
		result.Networks = append(result.Networks, override.Networks...)
	}

	// Merge custom fields
	if override.Custom != nil {
		if result.Custom == nil {
			result.Custom = make(map[string]interface{})
		}
		for key, value := range override.Custom {
			result.Custom[key] = value
		}
	}

	return result
}

// validateServiceOverride validates a service override configuration
func (c *ConfigLoader) validateServiceOverride(serviceName string, override ServiceOverride) error {
	// Validate environment variables
	for key, value := range override.Environment {
		if key == "" {
			return fmt.Errorf("empty environment variable key")
		}
		if strings.Contains(key, " ") {
			return fmt.Errorf("environment variable key '%s' contains spaces", key)
		}
		if value == "" {
			// Warning, not error - empty values might be intentional
		}
	}

	// Validate port mappings
	for host, container := range override.Ports {
		if !isValidPortMapping(host) {
			return fmt.Errorf("invalid host port mapping: %s", host)
		}
		if !isValidPortMapping(container) {
			return fmt.Errorf("invalid container port mapping: %s", container)
		}
	}

	return nil
}

// validateProfile validates a profile configuration
func (c *ConfigLoader) validateProfile(profileName string, profile Profile) error {
	if profile.Name == "" {
		return fmt.Errorf("profile name is required")
	}

	// Validate environment variables
	for key := range profile.Environment {
		if key == "" {
			return fmt.Errorf("empty environment variable key in profile")
		}
	}

	return nil
}

// isValidProjectName checks if a project name is valid
func isValidProjectName(name string) bool {
	if name == "" {
		return false
	}

	// Check for valid characters (alphanumeric, hyphens, underscores)
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_') {
			return false
		}
	}

	return true
}

// isValidPortMapping checks if a port mapping string is valid
func isValidPortMapping(portStr string) bool {
	if portStr == "" {
		return false
	}

	// Handle port ranges (e.g., "8000-8010")
	if strings.Contains(portStr, "-") {
		parts := strings.Split(portStr, "-")
		if len(parts) != 2 {
			return false
		}
		// Both parts should be valid port numbers
		return isValidPort(parts[0]) && isValidPort(parts[1])
	}

	// Handle single port
	return isValidPort(portStr)
}

// isValidPort checks if a string represents a valid port number
func isValidPort(portStr string) bool {
	if portStr == "" {
		return false
	}

	// Simple numeric check (more sophisticated validation could be added)
	for _, char := range portStr {
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
}
