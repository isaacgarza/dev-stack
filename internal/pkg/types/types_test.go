package types

import (
	"errors"
	"testing"
	"time"
)

func TestError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      Error
		expected string
	}{
		{
			name: "error without details",
			err: Error{
				Code:    "TEST_ERROR",
				Message: "test message",
			},
			expected: "TEST_ERROR: test message",
		},
		{
			name: "error with details",
			err: Error{
				Code:    "TEST_ERROR",
				Message: "test message",
				Details: "additional details",
			},
			expected: "TEST_ERROR: test message (additional details)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			if result != tt.expected {
				t.Errorf("Error() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestError_Unwrap(t *testing.T) {
	cause := errors.New("original error")
	err := Error{
		Code:    "WRAPPED_ERROR",
		Message: "wrapped message",
		Cause:   cause,
	}

	unwrapped := err.Unwrap()
	if unwrapped != cause {
		t.Errorf("Unwrap() = %v, expected %v", unwrapped, cause)
	}

	// Test nil cause
	errNoCause := Error{
		Code:    "NO_CAUSE",
		Message: "no cause",
	}
	if errNoCause.Unwrap() != nil {
		t.Errorf("Unwrap() should return nil for error without cause")
	}
}

func TestNewError(t *testing.T) {
	code := "TEST_CODE"
	message := "test message"

	err := NewError(code, message)

	if err.Code != code {
		t.Errorf("NewError() Code = %q, expected %q", err.Code, code)
	}
	if err.Message != message {
		t.Errorf("NewError() Message = %q, expected %q", err.Message, message)
	}
	if err.Details != "" {
		t.Errorf("NewError() Details = %q, expected empty string", err.Details)
	}
	if err.Cause != nil {
		t.Errorf("NewError() Cause = %v, expected nil", err.Cause)
	}
}

func TestNewErrorWithDetails(t *testing.T) {
	code := "TEST_CODE"
	message := "test message"
	details := "test details"

	err := NewErrorWithDetails(code, message, details)

	if err.Code != code {
		t.Errorf("NewErrorWithDetails() Code = %q, expected %q", err.Code, code)
	}
	if err.Message != message {
		t.Errorf("NewErrorWithDetails() Message = %q, expected %q", err.Message, message)
	}
	if err.Details != details {
		t.Errorf("NewErrorWithDetails() Details = %q, expected %q", err.Details, details)
	}
	if err.Cause != nil {
		t.Errorf("NewErrorWithDetails() Cause = %v, expected nil", err.Cause)
	}
}

func TestNewErrorWithCause(t *testing.T) {
	code := "TEST_CODE"
	message := "test message"
	cause := errors.New("original error")

	err := NewErrorWithCause(code, message, cause)

	if err.Code != code {
		t.Errorf("NewErrorWithCause() Code = %q, expected %q", err.Code, code)
	}
	if err.Message != message {
		t.Errorf("NewErrorWithCause() Message = %q, expected %q", err.Message, message)
	}
	if err.Details != "" {
		t.Errorf("NewErrorWithCause() Details = %q, expected empty string", err.Details)
	}
	if err.Cause != cause {
		t.Errorf("NewErrorWithCause() Cause = %v, expected %v", err.Cause, cause)
	}
}

func TestProjectStructure(t *testing.T) {
	// Test that Project struct can be created and populated
	now := time.Now()
	project := Project{
		Name: "test-project",
		Type: "web",
		Path: "/path/to/project",
		Environment: map[string]string{
			"ENV": "test",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if project.Name != "test-project" {
		t.Errorf("Project.Name = %q, expected %q", project.Name, "test-project")
	}
	if project.Type != "web" {
		t.Errorf("Project.Type = %q, expected %q", project.Type, "web")
	}
	if project.Environment["ENV"] != "test" {
		t.Errorf("Project.Environment[ENV] = %q, expected %q", project.Environment["ENV"], "test")
	}
}

func TestServiceStructure(t *testing.T) {
	service := Service{
		Name:  "test-service",
		Type:  "container",
		Image: "nginx:latest",
		Ports: []PortMapping{
			{
				Host:      "8080",
				Container: "80",
				Protocol:  "tcp",
			},
		},
		Environment: map[string]string{
			"DEBUG": "true",
		},
		DependsOn: []string{"database"},
	}

	if service.Name != "test-service" {
		t.Errorf("Service.Name = %q, expected %q", service.Name, "test-service")
	}
	if len(service.Ports) != 1 {
		t.Errorf("Service.Ports length = %d, expected 1", len(service.Ports))
	}
	if service.Ports[0].Host != "8080" {
		t.Errorf("Service.Ports[0].Host = %q, expected %q", service.Ports[0].Host, "8080")
	}
	if len(service.DependsOn) != 1 || service.DependsOn[0] != "database" {
		t.Errorf("Service.DependsOn = %v, expected [database]", service.DependsOn)
	}
}

func TestHealthCheckStructure(t *testing.T) {
	healthCheck := HealthCheck{
		Test:        []string{"CMD", "curl", "-f", "http://localhost/health"},
		Interval:    30 * time.Second,
		Timeout:     10 * time.Second,
		Retries:     3,
		StartPeriod: 60 * time.Second,
	}

	if len(healthCheck.Test) != 4 {
		t.Errorf("HealthCheck.Test length = %d, expected 4", len(healthCheck.Test))
	}
	if healthCheck.Interval != 30*time.Second {
		t.Errorf("HealthCheck.Interval = %v, expected %v", healthCheck.Interval, 30*time.Second)
	}
	if healthCheck.Retries != 3 {
		t.Errorf("HealthCheck.Retries = %d, expected 3", healthCheck.Retries)
	}
}

func TestServiceStatusStructure(t *testing.T) {
	now := time.Now()
	startTime := now.Add(-5 * time.Minute)

	status := ServiceStatus{
		Name:   "web-service",
		State:  "running",
		Health: "healthy",
		Uptime: 5 * time.Minute,
		Memory: MemoryUsage{
			Used:  1024 * 1024 * 100, // 100MB
			Limit: 1024 * 1024 * 512, // 512MB
		},
		CreatedAt: now,
		StartedAt: &startTime,
	}

	if status.Name != "web-service" {
		t.Errorf("ServiceStatus.Name = %q, expected %q", status.Name, "web-service")
	}
	if status.State != "running" {
		t.Errorf("ServiceStatus.State = %q, expected %q", status.State, "running")
	}
	if status.Memory.Used != 1024*1024*100 {
		t.Errorf("ServiceStatus.Memory.Used = %d, expected %d", status.Memory.Used, 1024*1024*100)
	}
	if status.StartedAt == nil {
		t.Error("ServiceStatus.StartedAt should not be nil")
	}
}

func TestTemplateStructure(t *testing.T) {
	template := Template{
		Name:        "web-app-template",
		Description: "A basic web application template",
		Type:        "web",
		Version:     "1.0.0",
		Tags:        []string{"web", "nodejs", "express"},
		Files: []TemplateFile{
			{
				Source:      "package.json.tmpl",
				Destination: "package.json",
				Template:    true,
			},
		},
		Variables: []TemplateVar{
			{
				Name:        "app_name",
				Description: "Application name",
				Type:        "string",
				Required:    true,
			},
		},
		PostInit: []string{"npm install"},
	}

	if template.Name != "web-app-template" {
		t.Errorf("Template.Name = %q, expected %q", template.Name, "web-app-template")
	}
	if len(template.Tags) != 3 {
		t.Errorf("Template.Tags length = %d, expected 3", len(template.Tags))
	}
	if len(template.Files) != 1 {
		t.Errorf("Template.Files length = %d, expected 1", len(template.Files))
	}
	if template.Files[0].Template != true {
		t.Error("Template.Files[0].Template should be true")
	}
	if len(template.Variables) != 1 {
		t.Errorf("Template.Variables length = %d, expected 1", len(template.Variables))
	}
	if template.Variables[0].Required != true {
		t.Error("Template.Variables[0].Required should be true")
	}
}

func TestConfigStructure(t *testing.T) {
	config := Config{
		Global: GlobalConfig{
			DefaultProjectType: "web",
			LogLevel:           "info",
			ColorOutput:        true,
			CheckUpdates:       true,
			TelemetryEnabled:   false,
		},
		Projects: map[string]ProjectConfig{
			"test-project": {
				Version: "3.8",
			},
		},
		Profiles: map[string]Profile{
			"development": {
				Name:        "development",
				Description: "Development profile",
				Services:    []string{"web", "database"},
			},
		},
	}

	if config.Global.DefaultProjectType != "web" {
		t.Errorf("Config.Global.DefaultProjectType = %q, expected %q", config.Global.DefaultProjectType, "web")
	}
	if !config.Global.ColorOutput {
		t.Error("Config.Global.ColorOutput should be true")
	}
	if config.Global.TelemetryEnabled {
		t.Error("Config.Global.TelemetryEnabled should be false")
	}
	if len(config.Projects) != 1 {
		t.Errorf("Config.Projects length = %d, expected 1", len(config.Projects))
	}
	if len(config.Profiles) != 1 {
		t.Errorf("Config.Profiles length = %d, expected 1", len(config.Profiles))
	}
}
