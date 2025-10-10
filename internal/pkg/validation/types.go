package validation

// ValidationResult represents the result of validation
type ValidationResult struct {
	Valid       bool                `yaml:"valid"`
	Errors      []ValidationError   `yaml:"errors,omitempty"`
	Warnings    []ValidationWarning `yaml:"warnings,omitempty"`
	Summary     ValidationSummary   `yaml:"summary"`
	Suggestions []string            `yaml:"suggestions,omitempty"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Type         string `yaml:"type"`
	Field        string `yaml:"field"`
	Message      string `yaml:"message"`
	Code         string `yaml:"code"`
	Severity     string `yaml:"severity"`
	Suggestion   string `yaml:"suggestion,omitempty"`
	LineNumber   int    `yaml:"line_number,omitempty"`
	ColumnNumber int    `yaml:"column_number,omitempty"`
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Type       string `yaml:"type"`
	Field      string `yaml:"field"`
	Message    string `yaml:"message"`
	Code       string `yaml:"code"`
	Suggestion string `yaml:"suggestion,omitempty"`
}

// ValidationSummary provides a summary of validation results
type ValidationSummary struct {
	TotalCommands      int     `yaml:"total_commands"`
	TotalCategories    int     `yaml:"total_categories"`
	TotalWorkflows     int     `yaml:"total_workflows"`
	TotalProfiles      int     `yaml:"total_profiles"`
	ErrorCount         int     `yaml:"error_count"`
	WarningCount       int     `yaml:"warning_count"`
	CriticalErrors     int     `yaml:"critical_errors"`
	ConfigurationScore float64 `yaml:"configuration_score"`
}

// Helper methods for adding errors and warnings
func AddError(result *ValidationResult, errorType, field, message, code, severity, suggestion string) {
	result.Errors = append(result.Errors, ValidationError{
		Type:       errorType,
		Field:      field,
		Message:    message,
		Code:       code,
		Severity:   severity,
		Suggestion: suggestion,
	})
	result.Valid = false
}

func AddWarning(result *ValidationResult, warningType, field, message, code, suggestion string) {
	result.Warnings = append(result.Warnings, ValidationWarning{
		Type:       warningType,
		Field:      field,
		Message:    message,
		Code:       code,
		Suggestion: suggestion,
	})
}
