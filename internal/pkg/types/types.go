package types

import (
	"fmt"
	"time"
)

// Project represents a development stack project
type Project struct {
	Name        string            `yaml:"name" json:"name"`
	Type        string            `yaml:"type" json:"type"`
	Path        string            `yaml:"path" json:"path"`
	Services    []Service         `yaml:"services" json:"services"`
	Environment map[string]string `yaml:"environment" json:"environment"`
	Config      ProjectConfig     `yaml:"config" json:"config"`
	CreatedAt   time.Time         `yaml:"created_at" json:"created_at"`
	UpdatedAt   time.Time         `yaml:"updated_at" json:"updated_at"`
}

// Service represents a service within a development stack
type Service struct {
	Name          string            `yaml:"name" json:"name"`
	Type          string            `yaml:"type" json:"type"`
	Image         string            `yaml:"image,omitempty" json:"image,omitempty"`
	Build         BuildConfig       `yaml:"build,omitempty" json:"build,omitempty"`
	Ports         []PortMapping     `yaml:"ports,omitempty" json:"ports,omitempty"`
	Environment   map[string]string `yaml:"environment,omitempty" json:"environment,omitempty"`
	Volumes       []VolumeMapping   `yaml:"volumes,omitempty" json:"volumes,omitempty"`
	DependsOn     []string          `yaml:"depends_on,omitempty" json:"depends_on,omitempty"`
	HealthCheck   *HealthCheck      `yaml:"health_check,omitempty" json:"health_check,omitempty"`
	Networks      []string          `yaml:"networks,omitempty" json:"networks,omitempty"`
	Labels        map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
	RestartPolicy string            `yaml:"restart,omitempty" json:"restart,omitempty"`
	Command       []string          `yaml:"command,omitempty" json:"command,omitempty"`
	Entrypoint    []string          `yaml:"entrypoint,omitempty" json:"entrypoint,omitempty"`
}

// BuildConfig represents build configuration for a service
type BuildConfig struct {
	Context    string            `yaml:"context" json:"context"`
	Dockerfile string            `yaml:"dockerfile,omitempty" json:"dockerfile,omitempty"`
	Args       map[string]string `yaml:"args,omitempty" json:"args,omitempty"`
	Target     string            `yaml:"target,omitempty" json:"target,omitempty"`
}

// PortMapping represents a port mapping between host and container
type PortMapping struct {
	Host      string `yaml:"host" json:"host"`
	Container string `yaml:"container" json:"container"`
	Protocol  string `yaml:"protocol,omitempty" json:"protocol,omitempty"`
}

// VolumeMapping represents a volume mapping between host and container
type VolumeMapping struct {
	Host      string `yaml:"host" json:"host"`
	Container string `yaml:"container" json:"container"`
	Mode      string `yaml:"mode,omitempty" json:"mode,omitempty"`
}

// HealthCheck represents health check configuration
type HealthCheck struct {
	Test        []string      `yaml:"test" json:"test"`
	Interval    time.Duration `yaml:"interval,omitempty" json:"interval,omitempty"`
	Timeout     time.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	Retries     int           `yaml:"retries,omitempty" json:"retries,omitempty"`
	StartPeriod time.Duration `yaml:"start_period,omitempty" json:"start_period,omitempty"`
}

// ProjectConfig represents project-level configuration
type ProjectConfig struct {
	Version    string                 `yaml:"version" json:"version"`
	Networks   map[string]Network     `yaml:"networks,omitempty" json:"networks,omitempty"`
	Volumes    map[string]Volume      `yaml:"volumes,omitempty" json:"volumes,omitempty"`
	Secrets    map[string]Secret      `yaml:"secrets,omitempty" json:"secrets,omitempty"`
	Profiles   []string               `yaml:"profiles,omitempty" json:"profiles,omitempty"`
	Extensions map[string]interface{} `yaml:"x-*,omitempty" json:"x-*,omitempty"`
}

// Network represents a Docker network configuration
type Network struct {
	Driver     string            `yaml:"driver,omitempty" json:"driver,omitempty"`
	DriverOpts map[string]string `yaml:"driver_opts,omitempty" json:"driver_opts,omitempty"`
	External   bool              `yaml:"external,omitempty" json:"external,omitempty"`
	Name       string            `yaml:"name,omitempty" json:"name,omitempty"`
	Labels     map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
	IPAM       *IPAMConfig       `yaml:"ipam,omitempty" json:"ipam,omitempty"`
}

// IPAMConfig represents IP Address Management configuration
type IPAMConfig struct {
	Driver string     `yaml:"driver,omitempty" json:"driver,omitempty"`
	Config []IPAMPool `yaml:"config,omitempty" json:"config,omitempty"`
}

// IPAMPool represents an IPAM pool configuration
type IPAMPool struct {
	Subnet  string `yaml:"subnet,omitempty" json:"subnet,omitempty"`
	Gateway string `yaml:"gateway,omitempty" json:"gateway,omitempty"`
}

// Volume represents a Docker volume configuration
type Volume struct {
	Driver     string            `yaml:"driver,omitempty" json:"driver,omitempty"`
	DriverOpts map[string]string `yaml:"driver_opts,omitempty" json:"driver_opts,omitempty"`
	External   bool              `yaml:"external,omitempty" json:"external,omitempty"`
	Name       string            `yaml:"name,omitempty" json:"name,omitempty"`
	Labels     map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
}

// Secret represents a Docker secret configuration
type Secret struct {
	File     string            `yaml:"file,omitempty" json:"file,omitempty"`
	External bool              `yaml:"external,omitempty" json:"external,omitempty"`
	Name     string            `yaml:"name,omitempty" json:"name,omitempty"`
	Labels   map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
}

// ServiceStatus represents the runtime status of a service
type ServiceStatus struct {
	Name      string            `json:"name"`
	State     ServiceState      `json:"state"`  // running, stopped, starting, stopping, error
	Health    HealthStatus      `json:"health"` // healthy, unhealthy, starting, none
	Uptime    time.Duration     `json:"uptime"`
	CPUUsage  float64           `json:"cpu_usage"`
	Memory    MemoryUsage       `json:"memory"`
	Ports     []PortMapping     `json:"ports"`
	Labels    map[string]string `json:"labels"`
	CreatedAt time.Time         `json:"created_at"`
	StartedAt *time.Time        `json:"started_at,omitempty"`
}

// MemoryUsage represents memory usage statistics
type MemoryUsage struct {
	Used  uint64 `json:"used"`
	Limit uint64 `json:"limit"`
}

// StackStatus represents the overall status of a development stack
type StackStatus struct {
	Project     string          `json:"project"`
	Services    []ServiceStatus `json:"services"`
	TotalCPU    float64         `json:"total_cpu"`
	TotalMemory uint64          `json:"total_memory"`
	Networks    []string        `json:"networks"`
	Volumes     []string        `json:"volumes"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// Template represents a project template
type Template struct {
	Name        string         `yaml:"name" json:"name"`
	Description string         `yaml:"description" json:"description"`
	Type        string         `yaml:"type" json:"type"`
	Version     string         `yaml:"version" json:"version"`
	Author      string         `yaml:"author,omitempty" json:"author,omitempty"`
	Tags        []string       `yaml:"tags,omitempty" json:"tags,omitempty"`
	Files       []TemplateFile `yaml:"files" json:"files"`
	Variables   []TemplateVar  `yaml:"variables,omitempty" json:"variables,omitempty"`
	PostInit    []string       `yaml:"post_init,omitempty" json:"post_init,omitempty"`
}

// TemplateFile represents a file in a project template
type TemplateFile struct {
	Source      string `yaml:"source" json:"source"`
	Destination string `yaml:"destination" json:"destination"`
	Executable  bool   `yaml:"executable,omitempty" json:"executable,omitempty"`
	Template    bool   `yaml:"template,omitempty" json:"template,omitempty"`
}

// TemplateVar represents a template variable
type TemplateVar struct {
	Name        string      `yaml:"name" json:"name"`
	Description string      `yaml:"description" json:"description"`
	Type        string      `yaml:"type" json:"type"` // string, int, bool, choice
	Default     interface{} `yaml:"default,omitempty" json:"default,omitempty"`
	Required    bool        `yaml:"required,omitempty" json:"required,omitempty"`
	Choices     []string    `yaml:"choices,omitempty" json:"choices,omitempty"`
}

// Config represents the main application configuration
type Config struct {
	Global   GlobalConfig             `yaml:"global" json:"global"`
	Projects map[string]ProjectConfig `yaml:"projects,omitempty" json:"projects,omitempty"`
	Profiles map[string]Profile       `yaml:"profiles,omitempty" json:"profiles,omitempty"`
}

// GlobalConfig represents global application settings
type GlobalConfig struct {
	DefaultProjectType string            `yaml:"default_project_type" json:"default_project_type"`
	DockerRegistry     string            `yaml:"docker_registry,omitempty" json:"docker_registry,omitempty"`
	LogLevel           string            `yaml:"log_level" json:"log_level"`
	ColorOutput        bool              `yaml:"color_output" json:"color_output"`
	CheckUpdates       bool              `yaml:"check_updates" json:"check_updates"`
	TelemetryEnabled   bool              `yaml:"telemetry_enabled" json:"telemetry_enabled"`
	Environment        map[string]string `yaml:"environment,omitempty" json:"environment,omitempty"`
}

// Profile represents a configuration profile
type Profile struct {
	Name        string                 `yaml:"name" json:"name"`
	Description string                 `yaml:"description,omitempty" json:"description,omitempty"`
	Services    []string               `yaml:"services" json:"services"`
	Environment map[string]string      `yaml:"environment,omitempty" json:"environment,omitempty"`
	Overrides   map[string]interface{} `yaml:"overrides,omitempty" json:"overrides,omitempty"`
}

// Error represents an application error with context
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
	Cause   error  `json:"-"`
}

func (e Error) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e Error) Unwrap() error {
	return e.Cause
}

// NewError creates a new Error with the given code and message
func NewError(code, message string) Error {
	return Error{
		Code:    code,
		Message: message,
	}
}

// NewErrorWithDetails creates a new Error with code, message, and details
func NewErrorWithDetails(code, message, details string) Error {
	return Error{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// NewErrorWithCause creates a new Error with a wrapped cause
func NewErrorWithCause(code, message string, cause error) Error {
	return Error{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}
