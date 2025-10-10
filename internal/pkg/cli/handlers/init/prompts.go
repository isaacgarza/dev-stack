package init

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/isaacgarza/dev-stack/internal/pkg/cli/handlers/utils"
	"github.com/isaacgarza/dev-stack/internal/pkg/ui"
)

// promptForProjectDetails prompts user for project configuration
func (h *InitHandler) promptForProjectDetails() (string, string, error) {
	var projectName, environment string

	// Get current directory name as default project name
	currentDir, err := filepath.Abs(".")
	if err != nil {
		return "", "", fmt.Errorf("failed to get current directory: %w", err)
	}
	defaultName := filepath.Base(currentDir)

	// Project name prompt
	namePrompt := &survey.Input{
		Message: "Project name:",
		Default: defaultName,
		Help:    "Enter a name for your project (letters, numbers, hyphens, underscores only)",
	}

	if err := survey.AskOne(namePrompt, &projectName, survey.WithValidator(func(ans interface{}) error {
		return h.validateProjectName(ans.(string))
	})); err != nil {
		return "", "", fmt.Errorf("failed to get project name: %w", err)
	}

	// Environment prompt
	envPrompt := &survey.Select{
		Message: "Environment:",
		Options: []string{"local"},
		Default: "local",
		Help:    "Select the environment for this project",
	}

	if err := survey.AskOne(envPrompt, &environment); err != nil {
		return "", "", fmt.Errorf("failed to get environment: %w", err)
	}

	return projectName, environment, nil
}

// promptForServices prompts user to select services
func (h *InitHandler) promptForServices() ([]string, error) {
	serviceUtils := utils.NewServiceUtils()

	// Get available services by category
	categories, err := serviceUtils.GetServicesByCategory()
	if err != nil {
		return nil, fmt.Errorf("failed to load services: %w", err)
	}

	if len(categories) == 0 {
		return nil, fmt.Errorf("no services available")
	}

	var selectedServices []string

	// For each category, prompt for service selection
	for categoryName, services := range categories {
		if len(services) == 0 {
			continue
		}

		ui.Info("Select %s services:", categoryName)

		var serviceOptions []string
		for _, service := range services {
			description := service.Description
			if description == "" {
				description = "No description available"
			}
			serviceOptions = append(serviceOptions, fmt.Sprintf("%s - %s", service.Name, description))
		}

		var selected []string
		prompt := &survey.MultiSelect{
			Message: fmt.Sprintf("Choose %s services:", categoryName),
			Options: serviceOptions,
			Help:    "Use space to select/deselect, enter to confirm",
		}

		if err := survey.AskOne(prompt, &selected); err != nil {
			return nil, fmt.Errorf("failed to select %s services: %w", categoryName, err)
		}

		// Extract service names from selections
		for _, selection := range selected {
			serviceName := strings.Split(selection, " - ")[0]
			selectedServices = append(selectedServices, serviceName)
		}
	}

	if len(selectedServices) == 0 {
		return nil, fmt.Errorf("no services selected")
	}

	return selectedServices, nil
}

// promptForAdvancedOptions prompts for advanced configuration options
func (h *InitHandler) promptForAdvancedOptions() (map[string]bool, map[string]bool, error) {
	validation := make(map[string]bool)
	advanced := make(map[string]bool)

	// Ask if user wants advanced options
	var wantsAdvanced bool
	advancedPrompt := &survey.Confirm{
		Message: "Configure advanced options?",
		Default: false,
		Help:    "Enable additional configuration options for validation, monitoring, etc.",
	}

	if err := survey.AskOne(advancedPrompt, &wantsAdvanced); err != nil {
		return validation, advanced, fmt.Errorf("failed to get advanced options preference: %w", err)
	}

	if !wantsAdvanced {
		return validation, advanced, nil
	}

	// Validation options
	validationOptions := []string{
		"Enable schema validation",
		"Enable health checks",
		"Enable dependency validation",
	}

	var selectedValidation []string
	validationPrompt := &survey.MultiSelect{
		Message: "Validation options:",
		Options: validationOptions,
		Help:    "Select validation features to enable",
	}

	if err := survey.AskOne(validationPrompt, &selectedValidation); err != nil {
		return validation, advanced, fmt.Errorf("failed to get validation options: %w", err)
	}

	for _, option := range selectedValidation {
		switch option {
		case "Enable schema validation":
			validation["schema"] = true
		case "Enable health checks":
			validation["health"] = true
		case "Enable dependency validation":
			validation["dependencies"] = true
		}
	}

	// Advanced options
	advancedOptions := []string{
		"Enable monitoring",
		"Enable logging aggregation",
		"Enable development tools",
		"Enable testing framework",
	}

	var selectedAdvanced []string
	advancedPrompt2 := &survey.MultiSelect{
		Message: "Advanced features:",
		Options: advancedOptions,
		Help:    "Select additional features to enable",
	}

	if err := survey.AskOne(advancedPrompt2, &selectedAdvanced); err != nil {
		return validation, advanced, fmt.Errorf("failed to get advanced features: %w", err)
	}

	for _, option := range selectedAdvanced {
		switch option {
		case "Enable monitoring":
			advanced["monitoring"] = true
		case "Enable logging aggregation":
			advanced["logging"] = true
		case "Enable development tools":
			advanced["devtools"] = true
		case "Enable testing framework":
			advanced["testing"] = true
		}
	}

	return validation, advanced, nil
}

// confirmInitialization shows a summary and asks for confirmation
func (h *InitHandler) confirmInitialization(projectName, environment string, services []string, validation, advanced map[string]bool) (bool, error) {
	ui.Info("Initialization Summary:")
	ui.Info("  Project: %s", projectName)
	ui.Info("  Environment: %s", environment)
	ui.Info("  Services: %s", strings.Join(services, ", "))

	if len(validation) > 0 {
		var validationFeatures []string
		for feature := range validation {
			validationFeatures = append(validationFeatures, feature)
		}
		ui.Info("  Validation: %s", strings.Join(validationFeatures, ", "))
	}

	if len(advanced) > 0 {
		var advancedFeatures []string
		for feature := range advanced {
			advancedFeatures = append(advancedFeatures, feature)
		}
		ui.Info("  Advanced: %s", strings.Join(advancedFeatures, ", "))
	}

	var confirm bool
	confirmPrompt := &survey.Confirm{
		Message: "Proceed with initialization?",
		Default: true,
	}

	if err := survey.AskOne(confirmPrompt, &confirm); err != nil {
		return false, fmt.Errorf("failed to get confirmation: %w", err)
	}

	return confirm, nil
}
