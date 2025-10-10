package validation

import (
	"sort"

	"github.com/isaacgarza/dev-stack/internal/pkg/config"
	"github.com/spf13/cobra"
)

// CLIValidator validates CLI consistency
type CLIValidator struct {
	config *config.CommandConfig
}

// NewCLIValidator creates a new CLI validator
func NewCLIValidator(config *config.CommandConfig) *CLIValidator {
	return &CLIValidator{
		config: config,
	}
}

// ValidateAgainstCLI validates the configuration against actual CLI implementation
func (v *CLIValidator) ValidateAgainstCLI(rootCmd *cobra.Command) *ValidationResult {
	result := &ValidationResult{
		Valid: true,
		Summary: ValidationSummary{
			TotalCommands: len(v.config.Commands),
		},
	}

	cliCommands := v.extractCLICommands(rootCmd)

	for cmdName := range v.config.Commands {
		if !contains(cliCommands, cmdName) {
			AddError(result, "cli_consistency", "commands."+cmdName, "Command defined in YAML but not implemented in CLI", "CLI_MISSING_COMMAND", "high", "Implement command "+cmdName+" in CLI or remove from YAML")
		}
	}

	for _, cmdName := range cliCommands {
		if _, exists := v.config.Commands[cmdName]; !exists {
			AddWarning(result, "cli_consistency", "commands."+cmdName, "Command implemented in CLI but not defined in YAML", "YAML_MISSING_COMMAND", "Add command "+cmdName+" to YAML configuration")
		}
	}

	v.calculateSummary(result)
	return result
}

// extractCLICommands extracts command names from CLI structure
func (v *CLIValidator) extractCLICommands(rootCmd *cobra.Command) []string {
	var commands []string

	for _, cmd := range rootCmd.Commands() {
		if !cmd.Hidden {
			commands = append(commands, cmd.Name())
		}
	}

	sort.Strings(commands)
	return commands
}

// calculateSummary calculates validation summary statistics
func (v *CLIValidator) calculateSummary(result *ValidationResult) {
	result.Summary.ErrorCount = len(result.Errors)
	result.Summary.WarningCount = len(result.Warnings)

	for _, err := range result.Errors {
		if err.Severity == "critical" {
			result.Summary.CriticalErrors++
		}
	}

	totalIssues := result.Summary.ErrorCount + result.Summary.WarningCount
	if totalIssues == 0 {
		result.Summary.ConfigurationScore = 100.0
	} else {
		weightedScore := 100.0 - float64(result.Summary.ErrorCount*10+result.Summary.WarningCount*2)
		if weightedScore < 0 {
			weightedScore = 0
		}
		result.Summary.ConfigurationScore = weightedScore
	}

	result.Valid = result.Summary.CriticalErrors == 0 && result.Summary.ErrorCount < 5
}
