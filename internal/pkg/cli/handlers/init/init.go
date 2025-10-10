package init

import (
	"context"
	"fmt"

	"github.com/isaacgarza/dev-stack/internal/pkg/cli/handlers/utils"
	"github.com/isaacgarza/dev-stack/internal/pkg/cli/types"
	"github.com/isaacgarza/dev-stack/internal/pkg/constants"
	"github.com/isaacgarza/dev-stack/internal/pkg/ui"
	"github.com/spf13/cobra"
)

// InitHandler handles the init command
type InitHandler struct {
	serviceUtils *utils.ServiceUtils
}

// NewInitHandler creates a new InitHandler
func NewInitHandler() *InitHandler {
	return &InitHandler{
		serviceUtils: utils.NewServiceUtils(),
	}
}

// Handle executes the init command
func (h *InitHandler) Handle(ctx context.Context, cmd *cobra.Command, args []string, base *types.BaseCommand) error {
	force, _ := cmd.Flags().GetBool("force")

	ui.Header(constants.MsgInitializing)

	// Validate environment
	if err := h.validateInitEnvironment(); err != nil && !force {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Validate directory structure
	if err := h.validateDirectoryStructure(); err != nil && !force {
		return fmt.Errorf("directory validation failed: %w", err)
	}

	// Prompt for project details
	projectName, environment, err := h.promptForProjectDetails()
	if err != nil {
		return fmt.Errorf("failed to get project details: %w", err)
	}

	// Prompt for services
	services, err := h.promptForServices()
	if err != nil {
		return fmt.Errorf("failed to select services: %w", err)
	}

	// Validate selected services
	if err := h.validateServices(services); err != nil {
		return fmt.Errorf("service validation failed: %w", err)
	}

	// Prompt for advanced options
	validation, advanced, err := h.promptForAdvancedOptions()
	if err != nil {
		return fmt.Errorf("failed to get advanced options: %w", err)
	}

	// Confirm initialization
	confirmed, err := h.confirmInitialization(projectName, environment, services, validation, advanced)
	if err != nil {
		return fmt.Errorf("failed to get confirmation: %w", err)
	}

	if !confirmed {
		ui.Info("Initialization cancelled")
		return nil
	}

	// Create directory structure
	if err := h.createDirectoryStructure(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Create configuration file
	if err := h.createConfigFile(projectName, environment, services, validation, advanced); err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}

	// Generate initial compose files
	if err := h.generateInitialComposeFiles(services, projectName, environment, validation, advanced); err != nil {
		return fmt.Errorf("failed to generate compose files: %w", err)
	}

	// Create .gitignore entries
	if err := h.createGitignoreEntries(); err != nil {
		ui.Warning("Failed to update .gitignore: %v", err)
	}

	// Create README
	if err := h.createReadme(projectName, services); err != nil {
		ui.Warning("Failed to create README: %v", err)
	}

	ui.Success(constants.MsgInitSuccess)
	ui.Info("Next steps:")
	ui.Info("  1. Review the configuration in %s/%s", constants.DevStackDir, constants.ConfigFileName)
	ui.Info("  2. Start your stack with: %s", constants.CmdUp)
	ui.Info("  3. Check status with: %s", constants.CmdStatus)

	return nil
}

// ValidateArgs validates the command arguments
func (h *InitHandler) ValidateArgs(args []string) error {
	return nil
}

// GetRequiredFlags returns required flags for this command
func (h *InitHandler) GetRequiredFlags() []string {
	return []string{}
}
