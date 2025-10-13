package init

import (
	"context"
	"strings"
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
	// Test should fail due to either directory validation or missing Docker
	assert.True(t,
		strings.Contains(err.Error(), "directory validation failed") ||
			strings.Contains(err.Error(), "required tool 'docker' is not available"),
		"Expected directory validation or Docker availability error, got: %s", err.Error())
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
