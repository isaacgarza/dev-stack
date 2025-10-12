package init

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/isaacgarza/dev-stack/internal/pkg/cli/types"
	"github.com/isaacgarza/dev-stack/internal/pkg/constants"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestHandle_ValidationFailure(t *testing.T) {
	cleanup := setupTestDir(t)
	defer cleanup()

	createTestConfig(t)

	handler := NewInitHandler()
	cmd := &cobra.Command{}
	cmd.Flags().Bool("force", false, "force initialization")

	err := handler.Handle(context.Background(), cmd, []string{}, &types.BaseCommand{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
}

func TestHandle_WithForceFlag(t *testing.T) {
	cleanup := setupTestDir(t)
	defer cleanup()

	handler := NewInitHandler()
	cmd := &cobra.Command{}
	cmd.Flags().Bool("force", true, "force initialization")

	err := handler.Handle(context.Background(), cmd, []string{}, &types.BaseCommand{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get project details")
}

func TestCreateDirectoryStructure(t *testing.T) {
	handler := NewInitHandler()
	cleanup := setupTestDir(t)
	defer cleanup()

	err := handler.createDirectoryStructure()
	assert.NoError(t, err)

	_, err = os.Stat(constants.DevStackDir)
	assert.NoError(t, err)
}

func TestCreateConfigFile(t *testing.T) {
	handler := NewInitHandler()
	cleanup := setupTestDir(t)
	defer cleanup()

	err := handler.createDirectoryStructure()
	assert.NoError(t, err)

	err = handler.createConfigFile(TestProjectName, TestEnvironmentLocal, []string{TestServicePostgres},
		map[string]bool{"skip_warnings": false},
		map[string]bool{"auto_start": true})
	assert.NoError(t, err)

	_, err = os.Stat(TestConfigFilePath)
	assert.NoError(t, err)
}

func TestCreateGitignoreEntries(t *testing.T) {
	handler := NewInitHandler()
	cleanup := setupTestDir(t)
	defer cleanup()

	err := handler.createGitignoreEntries()
	assert.NoError(t, err)

	_, err = os.Stat(constants.GitignoreFileName)
	assert.NoError(t, err)
}

func TestCreateReadme(t *testing.T) {
	handler := NewInitHandler()
	cleanup := setupTestDir(t)
	defer cleanup()

	err := handler.createDirectoryStructure()
	assert.NoError(t, err)

	err = handler.createReadme(TestProjectName, []string{TestServicePostgres, TestServiceRedis})
	assert.NoError(t, err)

	readmePath := filepath.Join(constants.DevStackDir, constants.ReadmeFileName)
	_, err = os.Stat(readmePath)
	assert.NoError(t, err)
}

func TestGenerateConfig(t *testing.T) {
	handler := NewInitHandler()

	config, err := handler.generateConfig(TestProjectName, TestEnvironmentLocal, []string{TestServicePostgres},
		map[string]bool{"skip_warnings": false},
		map[string]bool{"auto_start": true})

	assert.NoError(t, err)
	assert.Contains(t, config, TestProjectName)
	assert.Contains(t, config, TestServicePostgres)
}
