//go:build integration

package init

import (
	"context"
	"os"
	"testing"

	"github.com/isaacgarza/dev-stack/internal/pkg/cli/types"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitHandler_Handle_FullFlow(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "dev-stack-integration-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	handler := NewInitHandler()
	cmd := &cobra.Command{}
	cmd.Flags().Bool("force", true, "force initialization")

	base := &types.BaseCommand{}
	ctx := context.Background()

	// This would require mocking user input for full integration test
	// For now, just verify the handler can be created and basic validation works
	err = handler.Handle(ctx, cmd, []string{}, base)
	// Expected to fail at prompt stage without mocked input
	assert.Error(t, err)
}
