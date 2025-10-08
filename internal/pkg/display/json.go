package display

import (
	"encoding/json"
	"io"
)

// JSONFormatter implements JSON output formatting
type JSONFormatter struct {
	writer io.Writer
}

// NewJSONFormatter creates a new JSON formatter
func NewJSONFormatter(writer io.Writer) *JSONFormatter {
	return &JSONFormatter{writer: writer}
}

// FormatStatus formats service status as JSON
func (f *JSONFormatter) FormatStatus(services []ServiceStatus, options StatusOptions) error {
	output := map[string]interface{}{
		"services": services,
		"summary": map[string]interface{}{
			"total":   len(services),
			"running": f.countByState(services, "running"),
			"healthy": f.countByHealth(services, "healthy"),
		},
	}

	return f.writeJSON(output)
}

// FormatValidation formats validation results as JSON
func (f *JSONFormatter) FormatValidation(result ValidationResult, options ValidationOptions) error {
	return f.writeJSON(result)
}

// FormatVersion formats version information as JSON
func (f *JSONFormatter) FormatVersion(info VersionInfo, options VersionOptions) error {
	return f.writeJSON(info)
}

// FormatHealth formats health check results as JSON
func (f *JSONFormatter) FormatHealth(report HealthReport, options HealthOptions) error {
	return f.writeJSON(report)
}

// Helper methods
func (f *JSONFormatter) writeJSON(data interface{}) error {
	encoder := json.NewEncoder(f.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func (f *JSONFormatter) countByState(services []ServiceStatus, state string) int {
	count := 0
	for _, service := range services {
		if service.State == state {
			count++
		}
	}
	return count
}

func (f *JSONFormatter) countByHealth(services []ServiceStatus, health string) int {
	count := 0
	for _, service := range services {
		if service.Health == health {
			count++
		}
	}
	return count
}
