package display

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestTableFormatter(t *testing.T) {
	var buf bytes.Buffer
	formatter := NewTableFormatter(&buf)

	services := []ServiceStatus{
		{
			Name:      "redis",
			State:     "running",
			Health:    "healthy",
			Ports:     []string{"6379:6379"},
			CreatedAt: time.Now().Add(-time.Hour),
			UpdatedAt: time.Now(),
			Uptime:    time.Hour,
		},
		{
			Name:      "postgres",
			State:     "stopped",
			Health:    "unhealthy",
			Ports:     []string{"5432:5432"},
			CreatedAt: time.Now().Add(-30 * time.Minute),
			UpdatedAt: time.Now(),
			Uptime:    30 * time.Minute,
		},
	}

	options := StatusOptions{Quiet: false}
	err := formatter.FormatStatus(services, options)
	if err != nil {
		t.Errorf("FormatStatus failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "redis") {
		t.Error("Output should contain redis service")
	}
	if !strings.Contains(output, "postgres") {
		t.Error("Output should contain postgres service")
	}
	if !strings.Contains(output, "Resource Summary") {
		t.Error("Output should contain resource summary")
	}
}

func TestJSONFormatter(t *testing.T) {
	var buf bytes.Buffer
	formatter := NewJSONFormatter(&buf)

	services := []ServiceStatus{
		{
			Name:      "redis",
			State:     "running",
			Health:    "healthy",
			Ports:     []string{"6379:6379"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Uptime:    time.Hour,
		},
	}

	options := StatusOptions{Quiet: false}
	err := formatter.FormatStatus(services, options)
	if err != nil {
		t.Errorf("FormatStatus failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `"name": "redis"`) {
		t.Error("JSON output should contain redis service name")
	}
	if !strings.Contains(output, `"services"`) {
		t.Error("JSON output should contain services key")
	}
	if !strings.Contains(output, `"summary"`) {
		t.Error("JSON output should contain summary")
	}
}

func TestYAMLFormatter(t *testing.T) {
	var buf bytes.Buffer
	formatter := NewYAMLFormatter(&buf)

	services := []ServiceStatus{
		{
			Name:      "redis",
			State:     "running",
			Health:    "healthy",
			Ports:     []string{"6379:6379"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Uptime:    time.Hour,
		},
	}

	options := StatusOptions{Quiet: false}
	err := formatter.FormatStatus(services, options)
	if err != nil {
		t.Errorf("FormatStatus failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "name: redis") {
		t.Error("YAML output should contain redis service name")
	}
	if !strings.Contains(output, "services:") {
		t.Error("YAML output should contain services key")
	}
}

func TestFormatterFactory(t *testing.T) {
	var buf bytes.Buffer

	// Test table formatter
	formatter, err := CreateFormatter("table", &buf)
	if err != nil {
		t.Errorf("Failed to create table formatter: %v", err)
	}
	if _, ok := formatter.(*TableFormatter); !ok {
		t.Error("Expected TableFormatter")
	}

	// Test JSON formatter
	formatter, err = CreateFormatter("json", &buf)
	if err != nil {
		t.Errorf("Failed to create JSON formatter: %v", err)
	}
	if _, ok := formatter.(*JSONFormatter); !ok {
		t.Error("Expected JSONFormatter")
	}

	// Test YAML formatter
	formatter, err = CreateFormatter("yaml", &buf)
	if err != nil {
		t.Errorf("Failed to create YAML formatter: %v", err)
	}
	if _, ok := formatter.(*YAMLFormatter); !ok {
		t.Error("Expected YAMLFormatter")
	}

	// Test unsupported format
	_, err = CreateFormatter("xml", &buf)
	if err == nil {
		t.Error("Expected error for unsupported format")
	}

	// Test supported formats
	formats := GetSupportedFormats()
	expectedFormats := []string{"table", "json", "yaml"}
	if len(formats) != len(expectedFormats) {
		t.Errorf("Expected %d formats, got %d", len(expectedFormats), len(formats))
	}
}

func TestValidationFormatting(t *testing.T) {
	var buf bytes.Buffer
	formatter := NewTableFormatter(&buf)

	result := ValidationResult{
		Valid: false,
		Errors: []ValidationError{
			{
				Type:     "config",
				Field:    "services",
				Message:  "Invalid service configuration",
				Code:     "E001",
				Severity: "error",
			},
		},
		Warnings: []ValidationWarning{
			{
				Type:    "deprecation",
				Field:   "version",
				Message: "Using deprecated version format",
				Code:    "W001",
			},
		},
		Summary: ValidationSummary{
			TotalCommands:      10,
			ValidCommands:      8,
			ErrorCount:         1,
			WarningCount:       1,
			CoveragePercentage: 80,
		},
	}

	options := ValidationOptions{Strict: false}
	err := formatter.FormatValidation(result, options)
	if err != nil {
		t.Errorf("FormatValidation failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Configuration validation failed") {
		t.Error("Output should indicate validation failure")
	}
	if !strings.Contains(output, "Invalid service configuration") {
		t.Error("Output should contain error message")
	}
	if !strings.Contains(output, "Summary:") {
		t.Error("Output should contain summary")
	}
}
