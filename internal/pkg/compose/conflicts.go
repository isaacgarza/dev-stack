package compose

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// ConflictDetector handles detection and resolution of port conflicts
type ConflictDetector struct {
	services     map[string]*ServiceDefinition
	portMappings map[int][]string // port -> services using it
	conflicts    []PortConflict
	autoFix      bool
	nextFreePort int
}

// NewConflictDetector creates a new conflict detector
func NewConflictDetector(services map[string]*ServiceDefinition, autoFix bool) *ConflictDetector {
	return &ConflictDetector{
		services:     services,
		portMappings: make(map[int][]string),
		conflicts:    []PortConflict{},
		autoFix:      autoFix,
		nextFreePort: 8000, // Start looking for free ports from 8000
	}
}

// DetectConflicts analyzes services for port conflicts
func (cd *ConflictDetector) DetectConflicts() ([]PortConflict, error) {
	// Reset state
	cd.portMappings = make(map[int][]string)
	cd.conflicts = []PortConflict{}

	// Extract port mappings from all services
	for serviceName, service := range cd.services {
		ports, err := cd.extractPortsFromService(serviceName, service)
		if err != nil {
			return nil, fmt.Errorf("failed to extract ports from service %s: %w", serviceName, err)
		}

		for _, port := range ports {
			cd.portMappings[port] = append(cd.portMappings[port], serviceName)
		}
	}

	// Find conflicts
	for port, services := range cd.portMappings {
		if len(services) > 1 {
			conflict := PortConflict{
				Port:     port,
				Services: services,
				Severity: cd.determineSeverity(port, services),
			}
			cd.conflicts = append(cd.conflicts, conflict)
		}
	}

	// Sort conflicts by port number
	sort.Slice(cd.conflicts, func(i, j int) bool {
		return cd.conflicts[i].Port < cd.conflicts[j].Port
	})

	return cd.conflicts, nil
}

// ResolveConflicts attempts to resolve detected port conflicts
func (cd *ConflictDetector) ResolveConflicts() (map[string]map[int]int, error) {
	if !cd.autoFix {
		return nil, fmt.Errorf("auto-fix is disabled")
	}

	resolutions := make(map[string]map[int]int) // serviceName -> originalPort -> newPort

	for _, conflict := range cd.conflicts {
		if conflict.Severity == "error" {
			// Resolve error-level conflicts by reassigning ports
			for i, serviceName := range conflict.Services {
				if i == 0 {
					// Keep the first service on the original port
					continue
				}

				// Find a new port for the conflicting service
				newPort, err := cd.findFreePort(conflict.Port)
				if err != nil {
					return nil, fmt.Errorf("failed to find free port for service %s: %w", serviceName, err)
				}

				if resolutions[serviceName] == nil {
					resolutions[serviceName] = make(map[int]int)
				}
				resolutions[serviceName][conflict.Port] = newPort

				// Update our tracking
				cd.portMappings[newPort] = append(cd.portMappings[newPort], serviceName)
				cd.removeServiceFromPort(conflict.Port, serviceName)
			}
		}
	}

	return resolutions, nil
}

// ApplyResolutions applies port conflict resolutions to service definitions
func (cd *ConflictDetector) ApplyResolutions(resolutions map[string]map[int]int) error {
	for serviceName, portMappings := range resolutions {
		service, exists := cd.services[serviceName]
		if !exists {
			continue
		}

		if err := cd.updateServicePorts(service, portMappings); err != nil {
			return fmt.Errorf("failed to update ports for service %s: %w", serviceName, err)
		}
	}

	return nil
}

// GetConflictReport generates a human-readable conflict report
func (cd *ConflictDetector) GetConflictReport() string {
	if len(cd.conflicts) == 0 {
		return "✅ No port conflicts detected"
	}

	var report strings.Builder
	report.WriteString(fmt.Sprintf("⚠️  Found %d port conflict(s):\n\n", len(cd.conflicts)))

	for _, conflict := range cd.conflicts {
		severity := "🟡"
		if conflict.Severity == "error" {
			severity = "🔴"
		}

		report.WriteString(fmt.Sprintf("%s Port %d conflict:\n", severity, conflict.Port))
		report.WriteString(fmt.Sprintf("   Services: %s\n", strings.Join(conflict.Services, ", ")))
		report.WriteString(fmt.Sprintf("   Severity: %s\n", conflict.Severity))

		if cd.autoFix && conflict.Severity == "error" {
			report.WriteString("   Resolution: Will automatically reassign ports\n")
		} else {
			report.WriteString("   Resolution: Manual intervention required\n")
		}
		report.WriteString("\n")
	}

	return report.String()
}

// GetSuggestedResolutions provides manual resolution suggestions
func (cd *ConflictDetector) GetSuggestedResolutions() map[string][]string {
	suggestions := make(map[string][]string)

	for _, conflict := range cd.conflicts {
		key := fmt.Sprintf("port-%d", conflict.Port)
		suggestions[key] = []string{}

		suggestions[key] = append(suggestions[key],
			fmt.Sprintf("Port %d is used by: %s", conflict.Port, strings.Join(conflict.Services, ", ")))

		if conflict.Severity == "error" {
			suggestions[key] = append(suggestions[key],
				"Suggested actions:")
			suggestions[key] = append(suggestions[key],
				"1. Use different ports for each service")
			suggestions[key] = append(suggestions[key],
				"2. Configure port overrides in dev-stack-config.yaml")
			suggestions[key] = append(suggestions[key],
				"3. Enable auto-fix with --auto-fix-ports flag")

			// Suggest specific alternative ports
			for i, service := range conflict.Services {
				if i > 0 {
					suggestedPort, _ := cd.findFreePort(conflict.Port)
					suggestions[key] = append(suggestions[key],
						fmt.Sprintf("4. Move %s to port %d", service, suggestedPort))
				}
			}
		} else {
			suggestions[key] = append(suggestions[key],
				"This is a warning-level conflict and may not require action")
		}
	}

	return suggestions
}

// extractPortsFromService extracts port numbers from a service definition
func (cd *ConflictDetector) extractPortsFromService(serviceName string, service *ServiceDefinition) ([]int, error) {
	var ports []int

	for _, serviceConfig := range service.Services {
		if serviceMap, ok := serviceConfig.(map[string]interface{}); ok {
			if portsConfig, exists := serviceMap["ports"]; exists {
				servicePorts, err := cd.parsePortsConfig(portsConfig)
				if err != nil {
					return nil, fmt.Errorf("failed to parse ports for service %s: %w", serviceName, err)
				}
				ports = append(ports, servicePorts...)
			}
		}
	}

	return cd.deduplicatePorts(ports), nil
}

// parsePortsConfig parses various port configuration formats
func (cd *ConflictDetector) parsePortsConfig(portsConfig interface{}) ([]int, error) {
	var ports []int

	switch portsConfig := portsConfig.(type) {
	case []interface{}:
		for _, portEntry := range portsConfig {
			if portStr, ok := portEntry.(string); ok {
				portNums, err := cd.parsePortString(portStr)
				if err != nil {
					return nil, err
				}
				ports = append(ports, portNums...)
			}
		}
	case []string:
		for _, portStr := range portsConfig {
			portNums, err := cd.parsePortString(portStr)
			if err != nil {
				return nil, err
			}
			ports = append(ports, portNums...)
		}
	default:
		return nil, fmt.Errorf("unsupported ports configuration type: %T", portsConfig)
	}

	return ports, nil
}

// parsePortString parses port strings like "8080:8080", "8080", "${PORT:-8080}:8080"
func (cd *ConflictDetector) parsePortString(portStr string) ([]int, error) {
	// Handle environment variable substitution
	portStr = cd.expandPortVariables(portStr)

	// Handle port mapping formats
	if strings.Contains(portStr, ":") {
		parts := strings.Split(portStr, ":")
		if len(parts) >= 2 {
			// Extract host port (the first part)
			hostPort := strings.TrimSpace(parts[0])
			return cd.parsePortNumber(hostPort)
		}
	}

	// Handle simple port number
	return cd.parsePortNumber(portStr)
}

// parsePortNumber parses a port number string into integer(s)
func (cd *ConflictDetector) parsePortNumber(portStr string) ([]int, error) {
	portStr = strings.TrimSpace(portStr)

	// Handle port ranges
	if strings.Contains(portStr, "-") {
		return cd.parsePortRange(portStr)
	}

	// Handle single port
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid port number: %s", portStr)
	}

	if port < 1 || port > 65535 {
		return nil, fmt.Errorf("port number out of range: %d", port)
	}

	return []int{port}, nil
}

// parsePortRange parses port ranges like "8000-8010"
func (cd *ConflictDetector) parsePortRange(rangeStr string) ([]int, error) {
	parts := strings.Split(rangeStr, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid port range format: %s", rangeStr)
	}

	start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return nil, fmt.Errorf("invalid start port in range: %s", parts[0])
	}

	end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return nil, fmt.Errorf("invalid end port in range: %s", parts[1])
	}

	if start > end {
		return nil, fmt.Errorf("invalid port range: start port %d > end port %d", start, end)
	}

	var ports []int
	for port := start; port <= end; port++ {
		ports = append(ports, port)
	}

	return ports, nil
}

// expandPortVariables expands environment variables in port strings
func (cd *ConflictDetector) expandPortVariables(portStr string) string {
	// Handle ${VAR:-default} format
	envVarRegex := regexp.MustCompile(`\$\{([^}:]+)(:-([^}]+))?\}`)

	return envVarRegex.ReplaceAllStringFunc(portStr, func(match string) string {
		submatch := envVarRegex.FindStringSubmatch(match)
		if len(submatch) >= 4 && submatch[3] != "" {
			// Return the default value
			return submatch[3]
		}
		if len(submatch) >= 2 {
			// Try to get from common port mappings
			if defaultPort := cd.getCommonPortDefault(submatch[1]); defaultPort != "" {
				return defaultPort
			}
		}
		// Return as-is if no default found
		return match
	})
}

// getCommonPortDefault returns common default ports for well-known services
func (cd *ConflictDetector) getCommonPortDefault(varName string) string {
	commonPorts := map[string]string{
		"POSTGRES_PORT": "5432",
		"REDIS_PORT":    "6379",
		"MYSQL_PORT":    "3306",
		"KAFKA_PORT":    "9092",
		"HTTP_PORT":     "8080",
		"HTTPS_PORT":    "8443",
		"API_PORT":      "3000",
		"WEB_PORT":      "3000",
	}

	if port, exists := commonPorts[varName]; exists {
		return port
	}

	return ""
}

// determineSeverity determines the severity of a port conflict
func (cd *ConflictDetector) determineSeverity(port int, services []string) string {
	// Critical services that should not conflict
	criticalServices := map[string]bool{
		"postgres": true,
		"redis":    true,
		"mysql":    true,
		"kafka":    true,
	}

	// Well-known system ports
	if port < 1024 {
		return "error"
	}

	// Check if critical services are involved
	for _, service := range services {
		if criticalServices[service] {
			return "error"
		}
	}

	// Development ports (usually safe to conflict in some scenarios)
	if port >= 3000 && port <= 3010 {
		return "warning"
	}

	// Default to error for most conflicts
	return "error"
}

// findFreePort finds the next available port starting from a base port
func (cd *ConflictDetector) findFreePort(basePort int) (int, error) {
	startPort := basePort + 1
	if startPort < cd.nextFreePort {
		startPort = cd.nextFreePort
	}

	for port := startPort; port <= 65535; port++ {
		if _, exists := cd.portMappings[port]; !exists {
			cd.nextFreePort = port + 1
			return port, nil
		}
	}

	return 0, fmt.Errorf("no free ports available after %d", startPort)
}

// removeServiceFromPort removes a service from a port mapping
func (cd *ConflictDetector) removeServiceFromPort(port int, serviceName string) {
	services := cd.portMappings[port]
	for i, service := range services {
		if service == serviceName {
			cd.portMappings[port] = append(services[:i], services[i+1:]...)
			break
		}
	}

	// Remove the port entry if no services left
	if len(cd.portMappings[port]) == 0 {
		delete(cd.portMappings, port)
	}
}

// updateServicePorts updates port mappings in a service definition
func (cd *ConflictDetector) updateServicePorts(service *ServiceDefinition, portMappings map[int]int) error {
	for _, serviceConfig := range service.Services {
		if serviceMap, ok := serviceConfig.(map[string]interface{}); ok {
			if portsConfig, exists := serviceMap["ports"]; exists {
				updatedPorts, err := cd.updatePortsConfig(portsConfig, portMappings)
				if err != nil {
					return err
				}
				serviceMap["ports"] = updatedPorts
			}
		}
	}

	return nil
}

// updatePortsConfig updates a ports configuration with new port mappings
func (cd *ConflictDetector) updatePortsConfig(portsConfig interface{}, portMappings map[int]int) (interface{}, error) {
	switch portsConfig := portsConfig.(type) {
	case []interface{}:
		var updatedPorts []interface{}
		for _, portEntry := range portsConfig {
			if portStr, ok := portEntry.(string); ok {
				updatedPort, err := cd.updatePortString(portStr, portMappings)
				if err != nil {
					return nil, err
				}
				updatedPorts = append(updatedPorts, updatedPort)
			} else {
				updatedPorts = append(updatedPorts, portEntry)
			}
		}
		return updatedPorts, nil

	case []string:
		var updatedPorts []string
		for _, portStr := range portsConfig {
			updatedPort, err := cd.updatePortString(portStr, portMappings)
			if err != nil {
				return nil, err
			}
			updatedPorts = append(updatedPorts, updatedPort)
		}
		return updatedPorts, nil

	default:
		return portsConfig, nil
	}
}

// updatePortString updates a port string with new port mappings
func (cd *ConflictDetector) updatePortString(portStr string, portMappings map[int]int) (string, error) {
	if strings.Contains(portStr, ":") {
		parts := strings.Split(portStr, ":")
		if len(parts) >= 2 {
			hostPortStr := strings.TrimSpace(parts[0])

			// Try to parse the host port
			if hostPort, err := strconv.Atoi(hostPortStr); err == nil {
				if newPort, exists := portMappings[hostPort]; exists {
					// Replace the host port
					parts[0] = strconv.Itoa(newPort)
					return strings.Join(parts, ":"), nil
				}
			}
		}
	}

	return portStr, nil
}

// deduplicatePorts removes duplicate port numbers
func (cd *ConflictDetector) deduplicatePorts(ports []int) []int {
	seen := make(map[int]bool)
	var result []int

	for _, port := range ports {
		if !seen[port] {
			result = append(result, port)
			seen[port] = true
		}
	}

	return result
}

// GetPortUsage returns a summary of port usage across all services
func (cd *ConflictDetector) GetPortUsage() map[int][]string {
	return cd.portMappings
}

// IsPortConflicted checks if a specific port has conflicts
func (cd *ConflictDetector) IsPortConflicted(port int) bool {
	services, exists := cd.portMappings[port]
	return exists && len(services) > 1
}

// GetConflictedPorts returns a list of all conflicted port numbers
func (cd *ConflictDetector) GetConflictedPorts() []int {
	var ports []int
	for _, conflict := range cd.conflicts {
		ports = append(ports, conflict.Port)
	}
	return ports
}
