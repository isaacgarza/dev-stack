package validate

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/isaacgarza/dev-stack/internal/pkg/cli/handlers/utils"
	"github.com/isaacgarza/dev-stack/internal/pkg/cli/types"
	"github.com/isaacgarza/dev-stack/internal/pkg/config"
	"github.com/isaacgarza/dev-stack/internal/pkg/constants"
	"github.com/spf13/cobra"
)

// ValidateHandler handles the validate command for CI/CD integration
type ValidateHandler struct{}

// NewValidateHandler creates a new validate handler
func NewValidateHandler() *ValidateHandler {
	return &ValidateHandler{}
}

// Handle executes the validate command
func (h *ValidateHandler) Handle(ctx context.Context, cmd *cobra.Command, args []string, base *types.BaseCommand) error {
	flags := utils.GetCIFlags(cmd)

	// Load configuration
	loader := config.NewLoader("")
	commandConfig, err := loader.Load()
	if err != nil {
		utils.HandleError(flags, fmt.Errorf("failed to load configuration: %w", err))
		return nil
	}

	// Validate configuration
	result := commandConfig.Validate()

	// Handle CI exit codes
	exitCode := constants.ExitSuccess
	if !result.Valid {
		exitCode = constants.ExitError
	}
	if flags.Strict && len(result.Warnings) > 0 {
		exitCode = constants.ExitError
	}

	// Output results
	if flags.JSON {
		h.outputJSON(*result, exitCode)
	} else if !flags.Quiet {
		h.outputTable(*result, exitCode)
	} else if exitCode != constants.ExitSuccess {
		// Quiet mode with error
		os.Exit(exitCode)
	}

	return nil
}

func (h *ValidateHandler) outputJSON(result config.ValidationResult, exitCode int) {
	output := map[string]interface{}{
		"valid":     result.Valid,
		"errors":    h.formatErrors(result.Errors),
		"warnings":  h.formatWarnings(result.Warnings),
		"exit_code": exitCode,
	}

	_ = json.NewEncoder(os.Stdout).Encode(output)

	if exitCode != constants.ExitSuccess {
		os.Exit(exitCode)
	}
}

func (h *ValidateHandler) outputTable(result config.ValidationResult, exitCode int) {
	if result.Valid && len(result.Warnings) == 0 {
		fmt.Println("✅ Configuration is valid")
		return
	}

	if !result.Valid {
		fmt.Printf("❌ Configuration validation failed with %d errors:\n", len(result.Errors))
		for _, err := range result.Errors {
			fmt.Printf("  - %s: %s\n", err.Field, err.Message)
		}
	}

	if len(result.Warnings) > 0 {
		fmt.Printf("⚠️  %d warnings:\n", len(result.Warnings))
		for _, warning := range result.Warnings {
			fmt.Printf("  - %s: %s\n", warning.Field, warning.Message)
		}
	}

	if exitCode != constants.ExitSuccess {
		os.Exit(exitCode)
	}
}

func (h *ValidateHandler) formatErrors(errors []config.ValidationError) []map[string]string {
	result := make([]map[string]string, len(errors))
	for i, err := range errors {
		result[i] = map[string]string{
			"field":   err.Field,
			"message": err.Message,
		}
	}
	return result
}

func (h *ValidateHandler) formatWarnings(warnings []config.ValidationError) []map[string]string {
	result := make([]map[string]string, len(warnings))
	for i, warning := range warnings {
		result[i] = map[string]string{
			"field":   warning.Field,
			"message": warning.Message,
		}
	}
	return result
}

// ValidateArgs validates the command arguments
func (h *ValidateHandler) ValidateArgs(args []string) error {
	return nil
}

// GetRequiredFlags returns required flags for this command
func (h *ValidateHandler) GetRequiredFlags() []string {
	return []string{}
}
