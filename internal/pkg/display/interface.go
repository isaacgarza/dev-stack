package display

import (
	"io"
	"time"
)

// Formatter defines the interface for different output formats
type Formatter interface {
	// FormatStatus formats service status information
	FormatStatus(services []ServiceStatus, options StatusOptions) error

	// FormatValidation formats validation results
	FormatValidation(result ValidationResult, options ValidationOptions) error

	// FormatVersion formats version information
	FormatVersion(info VersionInfo, options VersionOptions) error

	// FormatHealth formats health check results
	FormatHealth(report HealthReport, options HealthOptions) error
}

// FormatterFactory creates formatters based on output format
type FormatterFactory interface {
	CreateFormatter(format string, writer io.Writer) (Formatter, error)
}

// Common data types
type ServiceStatus struct {
	Name      string        `json:"name" yaml:"name"`
	State     string        `json:"state" yaml:"state"`
	Health    string        `json:"health" yaml:"health"`
	Ports     []string      `json:"ports" yaml:"ports"`
	CreatedAt time.Time     `json:"created_at" yaml:"created_at"`
	UpdatedAt time.Time     `json:"updated_at" yaml:"updated_at"`
	Uptime    time.Duration `json:"uptime" yaml:"uptime"`
}

type ValidationResult struct {
	Valid       bool                `json:"valid" yaml:"valid"`
	Errors      []ValidationError   `json:"errors,omitempty" yaml:"errors,omitempty"`
	Warnings    []ValidationWarning `json:"warnings,omitempty" yaml:"warnings,omitempty"`
	Summary     ValidationSummary   `json:"summary" yaml:"summary"`
	Suggestions []string            `json:"suggestions,omitempty" yaml:"suggestions,omitempty"`
}

type ValidationError struct {
	Type         string `json:"type" yaml:"type"`
	Field        string `json:"field" yaml:"field"`
	Message      string `json:"message" yaml:"message"`
	Code         string `json:"code" yaml:"code"`
	Severity     string `json:"severity" yaml:"severity"`
	Suggestion   string `json:"suggestion,omitempty" yaml:"suggestion,omitempty"`
	LineNumber   int    `json:"line_number,omitempty" yaml:"line_number,omitempty"`
	ColumnNumber int    `json:"column_number,omitempty" yaml:"column_number,omitempty"`
}

type ValidationWarning struct {
	Type       string `json:"type" yaml:"type"`
	Field      string `json:"field" yaml:"field"`
	Message    string `json:"message" yaml:"message"`
	Code       string `json:"code" yaml:"code"`
	Suggestion string `json:"suggestion,omitempty" yaml:"suggestion,omitempty"`
}

type ValidationSummary struct {
	TotalCommands      int `json:"total_commands" yaml:"total_commands"`
	ValidCommands      int `json:"valid_commands" yaml:"valid_commands"`
	ErrorCount         int `json:"error_count" yaml:"error_count"`
	WarningCount       int `json:"warning_count" yaml:"warning_count"`
	CoveragePercentage int `json:"coverage_percentage" yaml:"coverage_percentage"`
}

type VersionInfo struct {
	Version   string            `json:"version" yaml:"version"`
	BuildInfo map[string]string `json:"build_info" yaml:"build_info"`
	GoVersion string            `json:"go_version" yaml:"go_version"`
	Platform  string            `json:"platform" yaml:"platform"`
}

type HealthReport struct {
	Overall HealthStatus  `json:"overall" yaml:"overall"`
	Checks  []HealthCheck `json:"checks" yaml:"checks"`
}

type HealthStatus struct {
	Status  string `json:"status" yaml:"status"`
	Message string `json:"message" yaml:"message"`
}

type HealthCheck struct {
	Name       string `json:"name" yaml:"name"`
	Status     string `json:"status" yaml:"status"`
	Message    string `json:"message" yaml:"message"`
	Suggestion string `json:"suggestion,omitempty" yaml:"suggestion,omitempty"`
	CanAutoFix bool   `json:"can_auto_fix" yaml:"can_auto_fix"`
	Severity   string `json:"severity" yaml:"severity"`
	Category   string `json:"category" yaml:"category"`
	Duration   string `json:"duration,omitempty" yaml:"duration,omitempty"`
}

// Format options
type StatusOptions struct {
	Quiet   bool
	Compact bool
	NoLogs  bool
}

type ValidationOptions struct {
	Strict bool
	Fix    bool
}

type VersionOptions struct {
	Full         bool
	CheckUpdates bool
}

type HealthOptions struct {
	Verbose bool
	AutoFix bool
}
