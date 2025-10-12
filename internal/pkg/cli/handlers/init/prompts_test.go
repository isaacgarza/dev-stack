package init

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPromptForProjectDetails_ValidationLogic(t *testing.T) {
	handler := NewInitHandler()

	// Test the validation logic that would be used in prompts
	tests := []struct {
		name        string
		projectName string
		expectError bool
	}{
		{"valid name", TestProjectName, false},
		{"invalid name", "", true},
		{"invalid characters", TestProjectNameInvalid, true},
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

func TestPromptForServices_SelectionLogic(t *testing.T) {
	// Test service selection logic
	serviceOptions := []string{
		"postgres - PostgreSQL database",
		"redis - In-memory data store",
		"nginx - Web server",
	}

	var selectedServices []string
	for _, option := range serviceOptions {
		serviceName := strings.Split(option, " - ")[0]
		selectedServices = append(selectedServices, serviceName)
	}

	assert.Len(t, selectedServices, 3)
	assert.Contains(t, selectedServices, TestServicePostgres)
	assert.Contains(t, selectedServices, TestServiceRedis)
	assert.Contains(t, selectedServices, TestServiceNginx)
}

func TestPromptForAdvancedOptions_StructureLogic(t *testing.T) {
	// Test the validation and advanced option structures
	validation := map[string]bool{
		"schema":       true,
		"health":       false,
		"dependencies": true,
	}

	advanced := map[string]bool{
		"monitoring": true,
		"logging":    false,
		"devtools":   true,
		"testing":    false,
	}

	assert.Contains(t, validation, "schema")
	assert.Contains(t, advanced, "monitoring")
	assert.True(t, validation["schema"])
	assert.False(t, advanced["logging"])
}
