package types

import "github.com/isaacgarza/dev-stack/internal/pkg/constants"

// ServiceState represents the state of a service
type ServiceState string

const (
	ServiceStateRunning ServiceState = constants.StateRunning
	ServiceStateStopped ServiceState = constants.StateStopped
	ServiceStateCreated ServiceState = constants.StateCreated
)

// String returns the string representation of the service state
func (s ServiceState) String() string {
	return string(s)
}

// IsRunning returns true if the service is running
func (s ServiceState) IsRunning() bool {
	return s == ServiceStateRunning
}

// IsStopped returns true if the service is stopped
func (s ServiceState) IsStopped() bool {
	return s == ServiceStateStopped
}

// HealthStatus represents the health status of a service
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = constants.HealthHealthy
	HealthStatusUnhealthy HealthStatus = constants.HealthUnhealthy
	HealthStatusStarting  HealthStatus = constants.HealthStarting
	HealthStatusNone      HealthStatus = constants.HealthNone
)

// String returns the string representation of the health status
func (h HealthStatus) String() string {
	return string(h)
}

// IsHealthy returns true if the service is healthy
func (h HealthStatus) IsHealthy() bool {
	return h == HealthStatusHealthy
}

// IsUnhealthy returns true if the service is unhealthy
func (h HealthStatus) IsUnhealthy() bool {
	return h == HealthStatusUnhealthy
}

// IsStarting returns true if the service is starting
func (h HealthStatus) IsStarting() bool {
	return h == HealthStatusStarting
}
