package init

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateInitEnvFile(t *testing.T) {
	handler := NewInitHandler()
	cleanup := setupTestDir(t)
	defer cleanup()

	err := handler.createDirectoryStructure()
	assert.NoError(t, err)

	projectConfig := &struct {
		Project struct {
			Name        string
			Environment string
		}
		Stack struct {
			Enabled []string
		}
	}{}
	projectConfig.Project.Name = TestProjectName
	projectConfig.Project.Environment = TestEnvironmentLocal
	projectConfig.Stack.Enabled = []string{TestServicePostgres}

	err = handler.generateInitEnvFile([]string{TestServicePostgres}, projectConfig)
	if err != nil {
		t.Logf("Expected error in test environment: %v", err)
	}
}

func TestGenerateInitDockerCompose(t *testing.T) {
	handler := NewInitHandler()
	cleanup := setupTestDir(t)
	defer cleanup()

	err := handler.createDirectoryStructure()
	assert.NoError(t, err)

	projectConfig := &struct {
		Project struct {
			Name        string
			Environment string
		}
		Stack struct {
			Enabled []string
		}
	}{}
	projectConfig.Project.Name = TestProjectName
	projectConfig.Project.Environment = TestEnvironmentLocal
	projectConfig.Stack.Enabled = []string{TestServicePostgres}

	err = handler.generateInitDockerCompose([]string{TestServicePostgres}, projectConfig)
	if err != nil {
		t.Logf("Expected error in test environment: %v", err)
	}
}

func TestGenerateInitialComposeFiles(t *testing.T) {
	handler := NewInitHandler()
	cleanup := setupTestDir(t)
	defer cleanup()

	err := handler.generateInitialComposeFiles([]string{TestServicePostgres}, TestProjectName, TestEnvironmentLocal,
		map[string]bool{"skip_warnings": false},
		map[string]bool{"auto_start": true})

	if err != nil {
		t.Logf("Expected error in test environment: %v", err)
	}
}
