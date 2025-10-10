package validation

import (
	"fmt"

	"github.com/isaacgarza/dev-stack/internal/pkg/config"
)

// BestPracticesValidator validates adherence to best practices
type BestPracticesValidator struct {
	config *config.CommandConfig
}

// NewBestPracticesValidator creates a new best practices validator
func NewBestPracticesValidator(config *config.CommandConfig) *BestPracticesValidator {
	return &BestPracticesValidator{
		config: config,
	}
}

// ValidateBestPractices validates adherence to best practices
func (v *BestPracticesValidator) ValidateBestPractices(result *ValidationResult) {
	v.validateNamingConventions(result)
	v.validateCategoryBalance(result)
	v.validateDocumentationCompleteness(result)
}

// validateNamingConventions validates naming conventions
func (v *BestPracticesValidator) validateNamingConventions(result *ValidationResult) {
	for cmdName := range v.config.Commands {
		if !isValidCommandName(cmdName) {
			AddWarning(result, "commands", "commands."+cmdName, "Command name should be lowercase with hyphens", "NAMING_CONVENTION", "Use lowercase letters and hyphens for command names")
		}
	}

	for catName := range v.config.Categories {
		if !isValidCategoryName(catName) {
			AddWarning(result, "categories", "categories."+catName, "Category name should be lowercase", "NAMING_CONVENTION", "Use lowercase letters for category names")
		}
	}
}

// validateCategoryBalance validates that categories are reasonably balanced
func (v *BestPracticesValidator) validateCategoryBalance(result *ValidationResult) {
	if len(v.config.Categories) == 0 {
		return
	}

	for catName, category := range v.config.Categories {
		if len(category.Commands) > 10 {
			AddWarning(result, "categories", "categories."+catName, "Category has many commands, consider splitting", "LARGE_CATEGORY", "Consider splitting large categories for better organization")
		}
	}
}

// validateDocumentationCompleteness validates documentation completeness
func (v *BestPracticesValidator) validateDocumentationCompleteness(result *ValidationResult) {
	totalCommands := len(v.config.Commands)
	commandsWithExamples := 0
	commandsWithTips := 0
	commandsWithLongDescription := 0

	for _, command := range v.config.Commands {
		if len(command.Examples) > 0 {
			commandsWithExamples++
		}
		if len(command.Tips) > 0 {
			commandsWithTips++
		}
		if command.LongDescription != "" {
			commandsWithLongDescription++
		}
	}

	exampleCoverage := float64(commandsWithExamples) / float64(totalCommands) * 100
	tipsCoverage := float64(commandsWithTips) / float64(totalCommands) * 100
	longDescCoverage := float64(commandsWithLongDescription) / float64(totalCommands) * 100

	if exampleCoverage < 80 {
		AddWarning(result, "documentation", "commands.examples", fmt.Sprintf("Only %.1f%% of commands have examples", exampleCoverage), "LOW_EXAMPLE_COVERAGE", "Add examples to more commands")
	}

	if tipsCoverage < 50 {
		AddWarning(result, "documentation", "commands.tips", fmt.Sprintf("Only %.1f%% of commands have tips", tipsCoverage), "LOW_TIPS_COVERAGE", "Add helpful tips to more commands")
	}

	if longDescCoverage < 60 {
		AddWarning(result, "documentation", "commands.long_description", fmt.Sprintf("Only %.1f%% of commands have detailed descriptions", longDescCoverage), "LOW_DESCRIPTION_COVERAGE", "Add detailed descriptions to more commands")
	}
}

func isValidCommandName(name string) bool {
	for _, r := range name {
		if (r < 'a' || r > 'z') && r != '-' && (r < '0' || r > '9') {
			return false
		}
	}
	return name != "" && name[0] != '-' && name[len(name)-1] != '-'
}

func isValidCategoryName(name string) bool {
	for _, r := range name {
		if (r < 'a' || r > 'z') && r != '_' {
			return false
		}
	}
	return name != ""
}
