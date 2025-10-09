package compose

import (
	"time"
)

// ServiceDefinition represents a parsed service from a docker-compose.yml file
type ServiceDefinition struct {
	Name         string                 `yaml:"name"`
	Path         string                 `yaml:"-"` // Path to the service directory
	ComposeFile  string                 `yaml:"-"` // Path to docker-compose.yml
	Services     map[string]interface{} `yaml:"services"`
	Volumes      map[string]interface{} `yaml:"volumes"`
	Networks     map[string]interface{} `yaml:"networks"`
	Dependencies []string               `yaml:"dependencies"`
	Metadata     ServiceMetadata        `yaml:"metadata"`
	RawContent   []byte                 `yaml:"-"` // Original file content for debugging
}

// DependencyConfig defines service dependency relationships
type DependencyConfig struct {
	Required  []string `yaml:"required"`
	Soft      []string `yaml:"soft"`
	Conflicts []string `yaml:"conflicts"`
	Provides  []string `yaml:"provides"`
}

// ServiceMetadata contains metadata about a service from services.yaml
type ServiceMetadata struct {
	Description      string            `yaml:"description"`
	Options          []string          `yaml:"options"`
	Examples         []string          `yaml:"examples"`
	UsageNotes       string            `yaml:"usage_notes"`
	Links            []string          `yaml:"links"`
	Tags             []string          `yaml:"tags"`
	Category         string            `yaml:"category"`
	Dependencies     []string          `yaml:"dependencies"`
	Ports            map[string]string `yaml:"ports"`
	Environment      map[string]string `yaml:"environment"`
	DependencyConfig DependencyConfig  `yaml:"dependency_config"`
}

// ServicesManifest represents the services.yaml file structure
type ServicesManifest map[string]ServiceMetadata

// ServiceRegistry manages all available services and their definitions
type ServiceRegistry struct {
	services     map[string]*ServiceDefinition
	manifest     ServicesManifest
	servicesPath string
	manifestPath string
	loaded       bool
	lastLoaded   time.Time
}

// ComposeOptions contains options for generating docker-compose files
type ComposeOptions struct {
	ProjectName      string                 `yaml:"project_name"`
	OutputFile       string                 `yaml:"output_file"`
	Overwrite        bool                   `yaml:"overwrite"`
	IncludeDeps      bool                   `yaml:"include_deps"`
	Profile          string                 `yaml:"profile"` // dev, test, prod
	Environment      map[string]string      `yaml:"environment"`
	NetworkName      string                 `yaml:"network_name"`
	VolumePrefix     string                 `yaml:"volume_prefix"`
	ServiceOverrides map[string]interface{} `yaml:"service_overrides"`
	ProjectOverrides *ProjectOverrides      `yaml:"project_overrides"`
	DetectConflicts  bool                   `yaml:"detect_conflicts"`
	AutoFixPorts     bool                   `yaml:"auto_fix_ports"`
	Interactive      bool                   `yaml:"interactive"`
	ProjectConfig    string                 `yaml:"project_config"`
	PortMappings     map[string]int         `yaml:"port_mappings"`
	ResourceLimits   map[string]Resource    `yaml:"resource_limits"`
}

// ComposeFile represents the final generated docker-compose structure
type ComposeFile struct {
	Version  string                 `yaml:"version"`
	Services map[string]interface{} `yaml:"services"`
	Networks map[string]interface{} `yaml:"networks"`
	Volumes  map[string]interface{} `yaml:"volumes"`
	Metadata ComposeMetadata        `yaml:"x-metadata"`
}

// ComposeMetadata contains generation metadata
type ComposeMetadata struct {
	GeneratedBy      string    `yaml:"generated_by"`
	GeneratedAt      time.Time `yaml:"generated_at"`
	ProjectName      string    `yaml:"project_name"`
	Services         []string  `yaml:"services"`
	Profile          string    `yaml:"profile"`
	FrameworkVersion string    `yaml:"framework_version"`
}

// ServiceDependency represents a dependency relationship between services
type ServiceDependency struct {
	Service   string   `yaml:"service"`
	DependsOn []string `yaml:"depends_on"`
	Optional  bool     `yaml:"optional"`
	Condition string   `yaml:"condition"` // service_started, service_healthy, service_completed_successfully
}

// ValidationResult contains service validation results
type ValidationResult struct {
	Valid    bool     `yaml:"valid"`
	Errors   []string `yaml:"errors"`
	Warnings []string `yaml:"warnings"`
	Service  string   `yaml:"service"`
}

// RegistryOptions contains options for service registry initialization
type RegistryOptions struct {
	ServicesPath string `yaml:"services_path"`
	ManifestPath string `yaml:"manifest_path"`
	AutoReload   bool   `yaml:"auto_reload"`
	Validate     bool   `yaml:"validate"`
}

// NetworkConfig represents network configuration
type NetworkConfig struct {
	Name     string            `yaml:"name"`
	Driver   string            `yaml:"driver"`
	External bool              `yaml:"external"`
	Labels   map[string]string `yaml:"labels"`
}

// VolumeConfig represents volume configuration
type VolumeConfig struct {
	Name     string            `yaml:"name"`
	Driver   string            `yaml:"driver"`
	External bool              `yaml:"external"`
	Labels   map[string]string `yaml:"labels"`
}

// TransformationRule defines how to transform service definitions
type TransformationRule struct {
	Pattern     string `yaml:"pattern"`
	Replacement string `yaml:"replacement"`
	Target      string `yaml:"target"` // network, volume, environment, etc.
	Condition   string `yaml:"condition"`
}

// ProjectOverrides contains project-specific service customizations
type ProjectOverrides struct {
	Services map[string]ServiceOverride `yaml:"services"`
	Global   GlobalOverrides            `yaml:"global"`
}

// ServiceOverride contains override configuration for a specific service
type ServiceOverride struct {
	Environment map[string]string      `yaml:"environment"`
	Ports       map[string]string      `yaml:"ports"`
	Volumes     []string               `yaml:"volumes"`
	Labels      map[string]string      `yaml:"labels"`
	Resources   Resource               `yaml:"resources"`
	Networks    []string               `yaml:"networks"`
	Command     string                 `yaml:"command"`
	Entrypoint  string                 `yaml:"entrypoint"`
	Enabled     *bool                  `yaml:"enabled"`
	Profile     string                 `yaml:"profile"`
	Custom      map[string]interface{} `yaml:"custom"`
}

// GlobalOverrides contains global override settings
type GlobalOverrides struct {
	Environment    map[string]string `yaml:"environment"`
	NetworkName    string            `yaml:"network_name"`
	VolumePrefix   string            `yaml:"volume_prefix"`
	ResourceLimits bool              `yaml:"resource_limits"`
	HealthChecks   bool              `yaml:"health_checks"`
	RestartPolicy  string            `yaml:"restart_policy"`
}

// Resource defines resource limits and reservations
type Resource struct {
	Limits       ResourceSpec `yaml:"limits"`
	Reservations ResourceSpec `yaml:"reservations"`
}

// ResourceSpec defines CPU and memory specifications
type ResourceSpec struct {
	CPU    string `yaml:"cpus"`
	Memory string `yaml:"memory"`
}

// PortConflict represents a port conflict detection result
type PortConflict struct {
	Port     int      `yaml:"port"`
	Services []string `yaml:"services"`
	Severity string   `yaml:"severity"` // warning, error
}

// ConfigValidation contains validation results for the overall configuration
type ConfigValidation struct {
	Valid         bool                        `yaml:"valid"`
	Errors        []string                    `yaml:"errors"`
	Warnings      []string                    `yaml:"warnings"`
	PortConflicts []PortConflict              `yaml:"port_conflicts"`
	ServiceIssues map[string]ValidationResult `yaml:"service_issues"`
}

// DefaultComposeOptions returns default options for compose generation
func DefaultComposeOptions() ComposeOptions {
	return ComposeOptions{
		ProjectName:      "dev-stack",
		OutputFile:       "docker-compose.generated.yml",
		Overwrite:        false,
		IncludeDeps:      true,
		Profile:          "dev",
		Environment:      make(map[string]string),
		NetworkName:      "dev-stack-network",
		VolumePrefix:     "",
		ServiceOverrides: make(map[string]interface{}),
		DetectConflicts:  true,
		AutoFixPorts:     false,
		Interactive:      false,
	}
}

// DefaultRegistryOptions returns default options for registry initialization
func DefaultRegistryOptions() RegistryOptions {
	return RegistryOptions{
		ServicesPath: "services",
		ManifestPath: "internal/config/services/services.yaml",
		AutoReload:   false,
		Validate:     true,
	}
}
