package init

import (
	"context"
	"testing"

	"github.com/isaacgarza/dev-stack/internal/pkg/cli/types"
	"github.com/isaacgarza/dev-stack/internal/pkg/constants"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestHandle_DirectoryValidation(t *testing.T) {
	cleanup := setupTestDir(t)
	defer cleanup()

	// Create conflicting docker-compose.yml file
	createTestFile(t, constants.DockerComposeFileName, "version: '3'")

	handler := NewInitHandler()
	cmd := &cobra.Command{}
	cmd.Flags().Bool("force", false, "force initialization")

	err := handler.Handle(context.Background(), cmd, []string{}, &types.BaseCommand{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "directory validation failed")
}

func TestHandle_AlreadyInitialized(t *testing.T) {
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
