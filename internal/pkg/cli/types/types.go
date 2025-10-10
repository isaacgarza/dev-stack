package types

// ServiceConfig represents the structure of service.yaml files
type ServiceConfig struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Defaults    struct {
		Image string `yaml:"image"`
		Port  int    `yaml:"port"`
	} `yaml:"defaults"`
	Environment map[string]string `yaml:"environment"`
	Docker      struct {
		// Single service configuration (legacy)
		Restart     string      `yaml:"restart,omitempty"`
		Command     interface{} `yaml:"command,omitempty"` // Can be string or []string
		Networks    []string    `yaml:"networks,omitempty"`
		MemoryLimit string      `yaml:"memory_limit,omitempty"`
		Environment []string    `yaml:"environment,omitempty"`
		ExtraHosts  []string    `yaml:"extra_hosts,omitempty"`
		HealthCheck struct {
			Test        []string `yaml:"test"`
			Interval    string   `yaml:"interval"`
			Timeout     string   `yaml:"timeout"`
			Retries     int      `yaml:"retries"`
			StartPeriod string   `yaml:"start_period"`
		} `yaml:"health_check,omitempty"`

		// Multi-service configuration (new)
		Services map[string]DockerService `yaml:"services,omitempty"`
	} `yaml:"docker"`
	Volumes []struct {
		Name  string `yaml:"name"`
		Mount string `yaml:"mount"`
	} `yaml:"volumes"`
}

// DockerService represents a single service in multi-service configuration
type DockerService struct {
	Image       string      `yaml:"image"`
	Restart     string      `yaml:"restart,omitempty"`
	Command     interface{} `yaml:"command,omitempty"` // Can be string or []string
	Networks    []string    `yaml:"networks,omitempty"`
	MemoryLimit string      `yaml:"memory_limit,omitempty"`
	Environment []string    `yaml:"environment,omitempty"`
	ExtraHosts  []string    `yaml:"extra_hosts,omitempty"`
	DependsOn   []string    `yaml:"depends_on,omitempty"`
	HealthCheck struct {
		Test        []string `yaml:"test"`
		Interval    string   `yaml:"interval"`
		Timeout     string   `yaml:"timeout"`
		Retries     int      `yaml:"retries"`
		StartPeriod string   `yaml:"start_period"`
	} `yaml:"health_check,omitempty"`
}

// ServiceInfo represents service information for display
type ServiceInfo struct {
	Name         string
	Description  string
	Category     string
	Dependencies []string
	Options      []string
	Examples     []string
	UsageNotes   string
	Links        []string
}

// ServiceStatus represents the status of a service
type ServiceStatus struct {
	Name   string
	Status string
	Health string
	Ports  []string
	Uptime string
}

// StartOptions represents options for starting services
type StartOptions struct {
	Services      []string
	Detached      bool
	Build         bool
	ForceRecreate bool
}

// StopOptions represents options for stopping services
type StopOptions struct {
	Timeout int
	Volumes bool
}
