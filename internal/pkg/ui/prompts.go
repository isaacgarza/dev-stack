package ui

import (
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
)

// ServiceOption represents a service option for selection
type ServiceOption struct {
	Name         string
	Description  string
	Dependencies []string
	Category     string
}

// CategoryOption represents a category option
type CategoryOption struct {
	Name     string
	Services []ServiceOption
}

// PromptInput prompts for text input
func PromptInput(message, defaultValue string) (string, error) {
	var result string
	prompt := &survey.Input{
		Message: message,
		Default: defaultValue,
	}
	err := survey.AskOne(prompt, &result)
	return result, err
}

// PromptConfirm prompts for yes/no confirmation
func PromptConfirm(message string, defaultValue bool) (bool, error) {
	var result bool
	prompt := &survey.Confirm{
		Message: message,
		Default: defaultValue,
	}
	err := survey.AskOne(prompt, &result)
	return result, err
}

// PromptMultiSelect prompts for multiple selections
func PromptMultiSelect(message string, options []string) ([]string, error) {
	var result []string
	prompt := &survey.MultiSelect{
		Message: message,
		Options: options,
	}
	err := survey.AskOne(prompt, &result)
	return result, err
}

// PromptCategorySelection prompts for category selection with service previews
func PromptCategorySelection(categories map[string][]ServiceOption) ([]string, error) {
	// Build category options with service counts
	var categoryNames []string
	var categoryDescriptions []string

	for categoryName, services := range categories {
		serviceNames := make([]string, len(services))
		for i, service := range services {
			serviceNames[i] = service.Name
		}

		categoryNames = append(categoryNames, categoryName)
		description := fmt.Sprintf("%s (%s)",
			categoryName,
			strings.Join(serviceNames, ", "))
		categoryDescriptions = append(categoryDescriptions, description)
	}

	var selectedCategories []string
	prompt := &survey.MultiSelect{
		Message: "Select service categories to configure:",
		Options: categoryDescriptions,
	}

	err := survey.AskOne(prompt, &selectedCategories)
	if err != nil {
		return nil, err
	}

	// Map back to category names
	var result []string
	for _, selected := range selectedCategories {
		for i, desc := range categoryDescriptions {
			if desc == selected {
				result = append(result, categoryNames[i])
				break
			}
		}
	}

	return result, nil
}

// PromptServiceSelection prompts for service selection within categories
func PromptServiceSelection(categoryName string, services []ServiceOption) ([]string, error) {
	if len(services) == 0 {
		return []string{}, nil
	}

	// Build service options with descriptions
	var serviceOptions []string
	for _, service := range services {
		option := service.Name + " - " + service.Description
		if len(service.Dependencies) > 0 {
			option += fmt.Sprintf(" (requires: %s)", strings.Join(service.Dependencies, ", "))
		}
		serviceOptions = append(serviceOptions, option)
	}

	var selectedServices []string
	prompt := &survey.MultiSelect{
		Message: fmt.Sprintf("Select %s services:", categoryName),
		Options: serviceOptions,
	}

	err := survey.AskOne(prompt, &selectedServices)
	if err != nil {
		return nil, err
	}

	// Map back to service names
	var result []string
	for _, selected := range selectedServices {
		for i, option := range serviceOptions {
			if option == selected {
				result = append(result, services[i].Name)
				break
			}
		}
	}

	return result, nil
}

// PromptValidationSettings prompts for validation settings
func PromptValidationSettings(settings map[string]struct {
	Description string
	Default     bool
}) (map[string]bool, error) {
	result := make(map[string]bool)

	for key, setting := range settings {
		var value bool
		prompt := &survey.Confirm{
			Message: fmt.Sprintf("%s - %s", key, setting.Description),
			Default: setting.Default,
		}
		err := survey.AskOne(prompt, &value)
		if err != nil {
			return nil, err
		}
		result[key] = value
	}

	return result, nil
}

// PromptFinalConfirmation shows a summary and asks for final confirmation
func PromptFinalConfirmation(projectName, environment string, services []string, dependencies []string) (bool, error) {
	summary := fmt.Sprintf(`
Project Configuration Summary:
  Name: %s
  Environment: %s
  
Selected Services: %s

Auto-resolved Dependencies: %s

Total Services: %d`,
		projectName,
		environment,
		strings.Join(services, ", "),
		strings.Join(dependencies, ", "),
		len(services)+len(dependencies))

	Box("Configuration Summary", summary)

	return PromptConfirm("Continue with this configuration?", true)
}
