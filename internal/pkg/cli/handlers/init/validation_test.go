package init

import (
	"runtime"
	"strings"
	"testing"

	"github.com/isaacgarza/dev-stack/internal/pkg/constants"
	"github.com/stretchr/testify/assert"
)

func TestValidateProjectName(t *testing.T) {
	handler := NewInitHandler()

	tests := []struct {
		name        string
		projectName string
		expectError bool
	}{
		{"valid name", TestProjectName, false},
		{"valid with underscore", "my_project", false},
		{"valid with numbers", "project123", false},
		{"empty name", "", true},
		{"too short", "a", true},
		{"too long", strings.Repeat("a", 51), true},
		{"invalid characters", TestProjectNameInvalid, true},
		{"starts with hyphen", "-project", true},
		{"starts with underscore", "_project", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.validateProjectName(tt.projectName)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateServices(t *testing.T) {
	handler := NewInitHandler()

	tests := []struct {
		name        string
		services    []string
		expectError bool
	}{
		{"empty services", []string{}, true},
		{"nil services", nil, true},
		{"invalid service", []string{"nonexistent-service"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.validateServices(tt.services)
			if tt.expectError {
				assert.Error(t, err)
			}
		})
	}
}

func TestValidateInitEnvironment(t *testing.T) {
	handler := NewInitHandler()
	cleanup := setupTestDir(t)
	defer cleanup()

	// Test clean directory
	err := handler.validateInitEnvironment()
	if err != nil && !strings.Contains(err.Error(), MsgRequiredTool) {
		t.Errorf("Unexpected error: %v", err)
	}

	// Test with existing config
	createTestConfig(t)
	err = handler.validateInitEnvironment()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), MsgAlreadyInitialized)
}

func TestValidateDirectoryStructure(t *testing.T) {
	handler := NewInitHandler()
	cleanup := setupTestDir(t)
	defer cleanup()

	// Test clean directory
	err := handler.validateDirectoryStructure()
	assert.NoError(t, err)

	// Test with conflicting file
	createTestFile(t, constants.DockerComposeFileName, "version: '3'")
	err = handler.validateDirectoryStructure()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "conflicting file")
}

func TestIsCommandAvailable(t *testing.T) {
	handler := NewInitHandler()

	// Use commands that definitely exist on each platform
	existingCommand := "echo" // Available on Unix systems
	if runtime.GOOS == "windows" {
		existingCommand = "where" // Built-in Windows command
	}

	tests := []struct {
		name     string
		command  string
		expected bool
	}{
		{"empty command", "", false},
		{"existing command", existingCommand, true},
		{"nonexistent command", "nonexistent-command-12345", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.isCommandAvailable(tt.command)
			assert.Equal(t, tt.expected, result)
		})
	}
}
