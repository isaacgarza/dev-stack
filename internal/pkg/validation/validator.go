package validation

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/isaacgarza/dev-stack/internal/pkg/config"
	"github.com/spf13/cobra"
)

// Validator provides comprehensive validation for CLI-YAML consistency
type Validator struct {
	config *config.CommandConfig
}

// ValidationResult represents the result of validation
type ValidationResult struct {
	Valid       bool                `yaml:"valid"`
	Errors      []ValidationError   `yaml:"errors,omitempty"`
	Warnings    []ValidationWarning `yaml:"warnings,omitempty"`
	Summary     ValidationSummary   `yaml:"summary"`
	Suggestions []string            `yaml:"suggestions,omitempty"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Type         string `yaml:"type"`
	Field        string `yaml:"field"`
	Message      string `yaml:"message"`
	Code         string `yaml:"code"`
	Severity     string `yaml:"severity"`
	Suggestion   string `yaml:"suggestion,omitempty"`
	LineNumber   int    `yaml:"line_number,omitempty"`
	ColumnNumber int    `yaml:"column_number,omitempty"`
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Type       string `yaml:"type"`
	Field      string `yaml:"field"`
	Message    string `yaml:"message"`
	Code       string `yaml:"code"`
	Suggestion string `yaml:"suggestion,omitempty"`
}

// ValidationSummary provides a summary of validation results
type ValidationSummary struct {
	TotalCommands      int     `yaml:"total_commands"`
	TotalCategories    int     `yaml:"total_categories"`
	TotalWorkflows     int     `yaml:"total_workflows"`
	TotalProfiles      int     `yaml:"total_profiles"`
	ErrorCount         int     `yaml:"error_count"`
	WarningCount       int     `yaml:"warning_count"`
	CriticalErrors     int     `yaml:"critical_errors"`
	ConfigurationScore float64 `yaml:"configuration_score"`
}

// NewValidator creates a new validator instance
func NewValidator(config *config.CommandConfig) *Validator {
	return &Validator{
		config: config,
	}
}

// ValidateAll performs comprehensive validation
func (v *Validator) ValidateAll() *ValidationResult {
	result := &ValidationResult{
		Valid: true,
		Summary: ValidationSummary{
			TotalCommands:   len(v.config.Commands),
			TotalCategories: len(v.config.Categories),
			TotalWorkflows:  len(v.config.Workflows),
			TotalProfiles:   len(v.config.Profiles),
		},
	}

	// Perform all validation checks
	v.validateMetadata(result)
	v.validateGlobalConfiguration(result)
	v.validateCategories(result)
	v.validateCommands(result)
	v.validateWorkflows(result)
	v.validateProfiles(result)
	v.validateReferences(result)
	v.validateBestPractices(result)

	// Calculate summary
	v.calculateSummary(result)

	// Generate suggestions
	v.generateSuggestions(result)

	return result
}

// validateMetadata validates the metadata section
func (v *Validator) validateMetadata(result *ValidationResult) {
	metadata := v.config.Metadata

	if metadata.Version == "" {
		v.addError(result, "metadata", "metadata.version", "Version is required", "MISSING_VERSION", "critical", "Add a version field to metadata section")
	}

	if metadata.CLIVersion == "" {
		v.addError(result, "metadata", "metadata.cli_version", "CLI version is required", "MISSING_CLI_VERSION", "critical", "Add a cli_version field to metadata section")
	}

	if metadata.Description == "" {
		v.addWarning(result, "metadata", "metadata.description", "Description is recommended", "MISSING_DESCRIPTION", "Add a description field to metadata section")
	}

	// Validate version format
	if metadata.Version != "" && !isValidVersionFormat(metadata.Version) {
		v.addError(result, "metadata", "metadata.version", "Invalid version format", "INVALID_VERSION_FORMAT", "high", "Use semantic versioning format (e.g., 2.0.0)")
	}
}

// validateGlobalConfiguration validates global configuration
func (v *Validator) validateGlobalConfiguration(result *ValidationResult) {
	global := v.config.Global

	// Check for required global flags
	requiredGlobalFlags := []string{"config", "verbose", "help"}
	for _, flagName := range requiredGlobalFlags {
		if _, exists := global.Flags[flagName]; !exists {
			v.addWarning(result, "global", "global.flags."+flagName, "Recommended global flag missing", "MISSING_GLOBAL_FLAG", "Consider adding the "+flagName+" global flag")
		}
	}

	// Validate global flag definitions
	for flagName, flag := range global.Flags {
		v.validateFlagDefinition(result, "global.flags."+flagName, flagName, flag)
	}
}

// validateCategories validates command categories
func (v *Validator) validateCategories(result *ValidationResult) {
	if len(v.config.Categories) == 0 {
		v.addWarning(result, "categories", "categories", "No categories defined", "NO_CATEGORIES", "Consider organizing commands into categories")
		return
	}

	for catName, category := range v.config.Categories {
		if category.Name == "" {
			v.addError(result, "categories", "categories."+catName+".name", "Category name is required", "MISSING_CATEGORY_NAME", "high", "Add a name field to category "+catName)
		}

		if category.Description == "" {
			v.addWarning(result, "categories", "categories."+catName+".description", "Category description is recommended", "MISSING_CATEGORY_DESCRIPTION", "Add a description to category "+catName)
		}

		if len(category.Commands) == 0 {
			v.addWarning(result, "categories", "categories."+catName+".commands", "Category has no commands", "EMPTY_CATEGORY", "Add commands to category "+catName+" or remove it")
		}

		// Validate category commands exist
		for _, cmdName := range category.Commands {
			if _, exists := v.config.Commands[cmdName]; !exists {
				v.addError(result, "categories", "categories."+catName+".commands", "Command '"+cmdName+"' does not exist", "UNDEFINED_COMMAND", "high", "Define command '"+cmdName+"' or remove it from category")
			}
		}
	}
}

// validateCommands validates all command definitions
func (v *Validator) validateCommands(result *ValidationResult) {
	if len(v.config.Commands) == 0 {
		v.addError(result, "commands", "commands", "No commands defined", "NO_COMMANDS", "critical", "Define at least one command")
		return
	}

	for cmdName, command := range v.config.Commands {
		v.validateCommand(result, cmdName, command)
	}

	// Check for orphaned commands (not in any category)
	v.validateOrphanedCommands(result)
}

// validateCommand validates a single command definition
func (v *Validator) validateCommand(result *ValidationResult, cmdName string, command config.Command) {
	prefix := "commands." + cmdName

	// Required fields
	if command.Description == "" {
		v.addError(result, "commands", prefix+".description", "Command description is required", "MISSING_DESCRIPTION", "high", "Add a description to command "+cmdName)
	}

	if command.Usage == "" {
		v.addError(result, "commands", prefix+".usage", "Command usage is required", "MISSING_USAGE", "high", "Add usage information to command "+cmdName)
	}

	// Validate category reference
	if command.Category != "" {
		if _, exists := v.config.Categories[command.Category]; !exists {
			v.addError(result, "commands", prefix+".category", "Category '"+command.Category+"' does not exist", "UNDEFINED_CATEGORY", "high", "Define category '"+command.Category+"' or change command category")
		}
	} else {
		v.addWarning(result, "commands", prefix+".category", "Command not assigned to category", "NO_CATEGORY", "Consider assigning command to a category for better organization")
	}

	// Validate flags
	for flagName, flag := range command.Flags {
		v.validateFlagDefinition(result, prefix+".flags."+flagName, flagName, flag)
	}

	// Validate examples
	if len(command.Examples) == 0 {
		v.addWarning(result, "commands", prefix+".examples", "No examples provided", "NO_EXAMPLES", "Add usage examples to help users understand the command")
	}

	// Validate related commands
	for _, relatedCmd := range command.RelatedCommands {
		if _, exists := v.config.Commands[relatedCmd]; !exists {
			v.addWarning(result, "commands", prefix+".related_commands", "Related command '"+relatedCmd+"' does not exist", "UNDEFINED_RELATED_COMMAND", "Remove reference or define the related command")
		}
	}

	// Validate aliases don't conflict
	v.validateCommandAliases(result, cmdName, command)
}

// validateFlagDefinition validates a flag definition
func (v *Validator) validateFlagDefinition(result *ValidationResult, prefix, flagName string, flag config.Flag) {
	// Validate flag type
	validTypes := []string{"bool", "string", "int", "float", "duration", "stringArray", "intArray"}
	if !contains(validTypes, flag.Type) {
		v.addError(result, "flags", prefix+".type", "Invalid flag type '"+flag.Type+"'", "INVALID_FLAG_TYPE", "high", "Use one of: "+strings.Join(validTypes, ", "))
	}

	// Validate description
	if flag.Description == "" {
		v.addError(result, "flags", prefix+".description", "Flag description is required", "MISSING_FLAG_DESCRIPTION", "medium", "Add description to flag "+flagName)
	}

	// Validate short flag format
	if flag.Short != "" && len(flag.Short) != 1 {
		v.addError(result, "flags", prefix+".short", "Short flag must be single character", "INVALID_SHORT_FLAG", "medium", "Use single character for short flag")
	}

	// Validate options for enum-like flags
	if len(flag.Options) > 0 && flag.Type != "string" {
		v.addWarning(result, "flags", prefix+".options", "Options should typically be used with string type flags", "OPTIONS_TYPE_MISMATCH", "Consider using string type for flags with options")
	}

	// Validate default value type consistency
	v.validateDefaultValueType(result, prefix, flag)
}

// validateDefaultValueType validates that default value matches flag type
func (v *Validator) validateDefaultValueType(result *ValidationResult, prefix string, flag config.Flag) {
	if flag.Default == nil {
		return
	}

	defaultType := reflect.TypeOf(flag.Default).Kind()

	switch flag.Type {
	case "bool":
		if defaultType != reflect.Bool {
			v.addError(result, "flags", prefix+".default", "Default value type mismatch for bool flag", "TYPE_MISMATCH", "medium", "Use boolean value for default")
		}
	case "int":
		if defaultType != reflect.Int && defaultType != reflect.Float64 {
			v.addError(result, "flags", prefix+".default", "Default value type mismatch for int flag", "TYPE_MISMATCH", "medium", "Use integer value for default")
		}
	case "string":
		if defaultType != reflect.String {
			v.addError(result, "flags", prefix+".default", "Default value type mismatch for string flag", "TYPE_MISMATCH", "medium", "Use string value for default")
		}
	}
}

// validateCommandAliases validates that command aliases don't conflict
func (v *Validator) validateCommandAliases(result *ValidationResult, cmdName string, command config.Command) {
	for _, alias := range command.Aliases {
		// Check if alias conflicts with another command name
		if _, exists := v.config.Commands[alias]; exists {
			v.addError(result, "commands", "commands."+cmdName+".aliases", "Alias '"+alias+"' conflicts with command name", "ALIAS_CONFLICT", "high", "Use a different alias that doesn't conflict with existing commands")
		}

		// Check if alias conflicts with other aliases
		for otherCmdName, otherCmd := range v.config.Commands {
			if otherCmdName == cmdName {
				continue
			}
			if contains(otherCmd.Aliases, alias) {
				v.addError(result, "commands", "commands."+cmdName+".aliases", "Alias '"+alias+"' conflicts with alias from command '"+otherCmdName+"'", "ALIAS_CONFLICT", "high", "Use a unique alias")
			}
		}
	}
}

// validateOrphanedCommands checks for commands not assigned to categories
func (v *Validator) validateOrphanedCommands(result *ValidationResult) {
	categorizedCommands := make(map[string]bool)

	for _, category := range v.config.Categories {
		for _, cmdName := range category.Commands {
			categorizedCommands[cmdName] = true
		}
	}

	for cmdName := range v.config.Commands {
		if !categorizedCommands[cmdName] {
			v.addWarning(result, "commands", "commands."+cmdName, "Command not assigned to any category", "ORPHANED_COMMAND", "Assign command to a category or create a new category")
		}
	}
}

// validateWorkflows validates workflow definitions
func (v *Validator) validateWorkflows(result *ValidationResult) {
	for workflowName, workflow := range v.config.Workflows {
		prefix := "workflows." + workflowName

		if workflow.Name == "" {
			v.addError(result, "workflows", prefix+".name", "Workflow name is required", "MISSING_WORKFLOW_NAME", "medium", "Add name to workflow "+workflowName)
		}

		if workflow.Description == "" {
			v.addWarning(result, "workflows", prefix+".description", "Workflow description is recommended", "MISSING_WORKFLOW_DESCRIPTION", "Add description to workflow "+workflowName)
		}

		if len(workflow.Steps) == 0 {
			v.addError(result, "workflows", prefix+".steps", "Workflow has no steps", "EMPTY_WORKFLOW", "medium", "Add steps to workflow "+workflowName)
		}

		// Validate workflow steps
		for i, step := range workflow.Steps {
			stepPrefix := fmt.Sprintf("%s.steps[%d]", prefix, i)

			if step.Command == "" {
				v.addError(result, "workflows", stepPrefix+".command", "Workflow step command is required", "MISSING_STEP_COMMAND", "medium", "Add command to workflow step")
			}

			if step.Description == "" {
				v.addWarning(result, "workflows", stepPrefix+".description", "Workflow step description is recommended", "MISSING_STEP_DESCRIPTION", "Add description to workflow step")
			}
		}
	}
}

// validateProfiles validates profile definitions
func (v *Validator) validateProfiles(result *ValidationResult) {
	for profileName, profile := range v.config.Profiles {
		prefix := "profiles." + profileName

		if profile.Name == "" {
			v.addError(result, "profiles", prefix+".name", "Profile name is required", "MISSING_PROFILE_NAME", "medium", "Add name to profile "+profileName)
		}

		if profile.Description == "" {
			v.addWarning(result, "profiles", prefix+".description", "Profile description is recommended", "MISSING_PROFILE_DESCRIPTION", "Add description to profile "+profileName)
		}

		if len(profile.Services) == 0 {
			v.addError(result, "profiles", prefix+".services", "Profile has no services", "EMPTY_PROFILE", "medium", "Add services to profile "+profileName)
		}

		// Note: We can't validate service names here without loading services.yaml
		// This would require dependency injection of the service registry
	}
}

// validateReferences validates cross-references between different sections
func (v *Validator) validateReferences(result *ValidationResult) {
	// This method could be extended to validate references between
	// commands, workflows, profiles, etc.
}

// validateBestPractices validates adherence to best practices
func (v *Validator) validateBestPractices(result *ValidationResult) {
	// Check for consistent naming conventions
	v.validateNamingConventions(result)

	// Check for balanced categories
	v.validateCategoryBalance(result)

	// Check for comprehensive documentation
	v.validateDocumentationCompleteness(result)
}

// validateNamingConventions validates naming conventions
func (v *Validator) validateNamingConventions(result *ValidationResult) {
	// Check command names are lowercase with hyphens
	for cmdName := range v.config.Commands {
		if !isValidCommandName(cmdName) {
			v.addWarning(result, "commands", "commands."+cmdName, "Command name should be lowercase with hyphens", "NAMING_CONVENTION", "Use lowercase letters and hyphens for command names")
		}
	}

	// Check category names are consistent
	for catName := range v.config.Categories {
		if !isValidCategoryName(catName) {
			v.addWarning(result, "categories", "categories."+catName, "Category name should be lowercase", "NAMING_CONVENTION", "Use lowercase letters for category names")
		}
	}
}

// validateCategoryBalance validates that categories are reasonably balanced
func (v *Validator) validateCategoryBalance(result *ValidationResult) {
	if len(v.config.Categories) == 0 {
		return
	}

	commandCounts := make([]int, 0, len(v.config.Categories))
	for _, category := range v.config.Categories {
		commandCounts = append(commandCounts, len(category.Commands))
	}

	// Check for categories with too many commands
	for catName, category := range v.config.Categories {
		if len(category.Commands) > 10 {
			v.addWarning(result, "categories", "categories."+catName, "Category has many commands, consider splitting", "LARGE_CATEGORY", "Consider splitting large categories for better organization")
		}
	}
}

// validateDocumentationCompleteness validates documentation completeness
func (v *Validator) validateDocumentationCompleteness(result *ValidationResult) {
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

	// Calculate documentation coverage
	exampleCoverage := float64(commandsWithExamples) / float64(totalCommands) * 100
	tipsCoverage := float64(commandsWithTips) / float64(totalCommands) * 100
	longDescCoverage := float64(commandsWithLongDescription) / float64(totalCommands) * 100

	if exampleCoverage < 80 {
		v.addWarning(result, "documentation", "commands.examples", fmt.Sprintf("Only %.1f%% of commands have examples", exampleCoverage), "LOW_EXAMPLE_COVERAGE", "Add examples to more commands")
	}

	if tipsCoverage < 50 {
		v.addWarning(result, "documentation", "commands.tips", fmt.Sprintf("Only %.1f%% of commands have tips", tipsCoverage), "LOW_TIPS_COVERAGE", "Add helpful tips to more commands")
	}

	if longDescCoverage < 60 {
		v.addWarning(result, "documentation", "commands.long_description", fmt.Sprintf("Only %.1f%% of commands have detailed descriptions", longDescCoverage), "LOW_DESCRIPTION_COVERAGE", "Add detailed descriptions to more commands")
	}
}

// calculateSummary calculates validation summary statistics
func (v *Validator) calculateSummary(result *ValidationResult) {
	result.Summary.ErrorCount = len(result.Errors)
	result.Summary.WarningCount = len(result.Warnings)

	// Count critical errors
	for _, err := range result.Errors {
		if err.Severity == "critical" {
			result.Summary.CriticalErrors++
		}
	}

	// Calculate configuration score (0-100)
	totalIssues := result.Summary.ErrorCount + result.Summary.WarningCount
	if totalIssues == 0 {
		result.Summary.ConfigurationScore = 100.0
	} else {
		// Weight errors more heavily than warnings
		weightedScore := 100.0 - float64(result.Summary.ErrorCount*10+result.Summary.WarningCount*2)
		if weightedScore < 0 {
			weightedScore = 0
		}
		result.Summary.ConfigurationScore = weightedScore
	}

	// Set overall validity
	result.Valid = result.Summary.CriticalErrors == 0 && result.Summary.ErrorCount < 5
}

// generateSuggestions generates improvement suggestions
func (v *Validator) generateSuggestions(result *ValidationResult) {
	if result.Summary.ErrorCount > 0 {
		result.Suggestions = append(result.Suggestions, "Fix validation errors to improve configuration quality")
	}

	if result.Summary.ConfigurationScore < 80 {
		result.Suggestions = append(result.Suggestions, "Consider improving documentation coverage and following best practices")
	}

	if len(v.config.Categories) == 0 {
		result.Suggestions = append(result.Suggestions, "Organize commands into categories for better CLI structure")
	}

	if len(v.config.Workflows) == 0 {
		result.Suggestions = append(result.Suggestions, "Add workflows to help users with common task sequences")
	}

	if len(v.config.Profiles) == 0 {
		result.Suggestions = append(result.Suggestions, "Define service profiles for quick environment setup")
	}
}

// Helper methods for adding errors and warnings
func (v *Validator) addError(result *ValidationResult, errorType, field, message, code, severity, suggestion string) {
	result.Errors = append(result.Errors, ValidationError{
		Type:       errorType,
		Field:      field,
		Message:    message,
		Code:       code,
		Severity:   severity,
		Suggestion: suggestion,
	})
	result.Valid = false
}

func (v *Validator) addWarning(result *ValidationResult, warningType, field, message, code, suggestion string) {
	result.Warnings = append(result.Warnings, ValidationWarning{
		Type:       warningType,
		Field:      field,
		Message:    message,
		Code:       code,
		Suggestion: suggestion,
	})
}

// Utility functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func isValidVersionFormat(version string) bool {
	// Simple semantic version validation
	parts := strings.Split(version, ".")
	return len(parts) >= 2 && len(parts) <= 3
}

func isValidCommandName(name string) bool {
	// Commands should be lowercase with optional hyphens
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || r == '-' || (r >= '0' && r <= '9')) {
			return false
		}
	}
	return name != "" && name[0] != '-' && name[len(name)-1] != '-'
}

func isValidCategoryName(name string) bool {
	// Categories should be lowercase
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || r == '_') {
			return false
		}
	}
	return name != ""
}

// ValidateAgainstCLI validates the configuration against actual CLI implementation
func (v *Validator) ValidateAgainstCLI(rootCmd *cobra.Command) *ValidationResult {
	result := &ValidationResult{
		Valid: true,
		Summary: ValidationSummary{
			TotalCommands: len(v.config.Commands),
		},
	}

	// Get all CLI commands
	cliCommands := v.extractCLICommands(rootCmd)

	// Check for commands in YAML but not in CLI
	for cmdName := range v.config.Commands {
		if !contains(cliCommands, cmdName) {
			v.addError(result, "cli_consistency", "commands."+cmdName, "Command defined in YAML but not implemented in CLI", "CLI_MISSING_COMMAND", "high", "Implement command "+cmdName+" in CLI or remove from YAML")
		}
	}

	// Check for commands in CLI but not in YAML
	for _, cmdName := range cliCommands {
		if _, exists := v.config.Commands[cmdName]; !exists {
			v.addWarning(result, "cli_consistency", "commands."+cmdName, "Command implemented in CLI but not defined in YAML", "YAML_MISSING_COMMAND", "Add command "+cmdName+" to YAML configuration")
		}
	}

	v.calculateSummary(result)
	return result
}

// extractCLICommands extracts command names from CLI structure
func (v *Validator) extractCLICommands(rootCmd *cobra.Command) []string {
	var commands []string

	for _, cmd := range rootCmd.Commands() {
		if !cmd.Hidden {
			commands = append(commands, cmd.Name())
		}
	}

	sort.Strings(commands)
	return commands
}
