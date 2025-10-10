package constants

// Docker container states
const (
	StateRunning = "running"
	StateStopped = "exited"
	StateCreated = "created"
)

// Health statuses
const (
	HealthHealthy   = "healthy"
	HealthUnhealthy = "unhealthy"
	HealthStarting  = "starting"
	HealthNone      = "none"
)

// Docker Compose labels
const (
	ComposeProjectLabel = "com.docker.compose.project"
	ComposeServiceLabel = "com.docker.compose.service"
)

// Docker file paths
const (
	DockerComposeFile = DevStackDir + "/" + DockerComposeFileName
)
