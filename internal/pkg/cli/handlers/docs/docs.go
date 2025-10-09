package docs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/isaacgarza/dev-stack/internal/pkg/cli/types"
	"github.com/isaacgarza/dev-stack/internal/pkg/ui"
	"github.com/spf13/cobra"
)

// DocsHandler handles the docs command
type DocsHandler struct{}

// NewDocsHandler creates a new docs handler
func NewDocsHandler() *DocsHandler {
	return &DocsHandler{}
}

// Handle executes the docs command
func (h *DocsHandler) Handle(ctx context.Context, cmd *cobra.Command, args []string, base *types.BaseCommand) error {
	outputDir, _ := cmd.Flags().GetString("output")
	if outputDir == "" {
		outputDir = "docs"
	}

	ui.Header("Generating Documentation")

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate documentation
	options := &GenerationOptions{
		ServicesMDPath: filepath.Join(outputDir, "services.md"),
	}
	generator := NewEnhancedGenerator(options)

	// Generate all documentation
	if _, err := generator.GenerateAll(); err != nil {
		return fmt.Errorf("failed to generate docs: %w", err)
	}

	ui.Success("Documentation generated in %s/", outputDir)
	return nil
}

// ValidateArgs validates the command arguments
func (h *DocsHandler) ValidateArgs(args []string) error {
	return nil
}

// GetRequiredFlags returns required flags for this command
func (h *DocsHandler) GetRequiredFlags() []string {
	return []string{}
}
