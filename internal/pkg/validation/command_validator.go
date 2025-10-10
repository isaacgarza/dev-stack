package validation

import (
	"reflect"

	"github.com/isaacgarza/dev-stack/internal/pkg/config"
)

// CommandValidator validates commands and categories
type CommandValidator struct {
	config *config.CommandConfig
}

// NewCommandValidator creates a new command validator
func NewCommandValidator(config *config.CommandConfig) *CommandValidator {
	return &CommandValidator{
		config: config,
	}
}

// ValidateCategories validates command categories
func (v *CommandValidator) ValidateCategories(result *ValidationResult) {
	if len(v.config.Categories) == 0 {
		AddWarning(result, "categories", "categories", "No categories defined", "NO_CATEGORIES", "Consider organizing commands into categories")
		return
	}

	for catName, category := range v.config.Categories {
		if category.Name == "" {
			AddError(result, "categories", "categories."+catName+".name", "Category name is required", "MISSING_CATEGORY_NAME", "high", "Add a name field to category "+catName)
		}

		if category.Description == "" {
			AddWarning(result, "categories", "categories."+catName+".description", "Category description is recommended", "MISSING_CATEGORY_DESCRIPTION", "Add a description to category "+catName)
		}

		if len(category.Commands) == 0 {
			AddWarning(result, "categories", "categories."+catName+".commands", "Category has no commands", "EMPTY_CATEGORY", "Add commands to category "+catName+" or remove it")
		}

		for _, cmdName := range category.Commands {
			if _, exists := v.config.Commands[cmdName]; !exists {
				AddError(result, "categories", "categories."+catName+".commands", "Command '"+cmdName+"' does not exist", "UNDEFINED_COMMAND", "high", "Define command '"+cmdName+"' or remove it from category")
			}
		}
	}
}

// ValidateCommands validates all command definitions
func (v *CommandValidator) ValidateCommands(result *ValidationResult) {
	if len(v.config.Commands) == 0 {
		AddError(result, "commands", "commands", "No commands defined", "NO_COMMANDS", "critical", "Define at least one command")
		return
	}

	for cmdName, command := range v.config.Commands {
		v.validateCommand(result, cmdName, command)
	}

	v.validateOrphanedCommands(result)
}

// validateCommand validates a single command definition
func (v *CommandValidator) validateCommand(result *ValidationResult, cmdName string, command config.Command) {
	prefix := "commands." + cmdName

	if command.Description == "" {
		AddError(result, "commands", prefix+".description", "Command description is required", "MISSING_DESCRIPTION", "high", "Add a description to command "+cmdName)
	}

	if command.Usage == "" {
		AddError(result, "commands", prefix+".usage", "Command usage is required", "MISSING_USAGE", "high", "Add usage information to command "+cmdName)
	}

	if command.Category != "" {
		if _, exists := v.config.Categories[command.Category]; !exists {
			AddError(result, "commands", prefix+".category", "Category '"+command.Category+"' does not exist", "UNDEFINED_CATEGORY", "high", "Define category '"+command.Category+"' or change command category")
		}
	} else {
		AddWarning(result, "commands", prefix+".category", "Command not assigned to category", "NO_CATEGORY", "Consider assigning command to a category for better organization")
	}

	for flagName, flag := range command.Flags {
		v.validateFlagDefinition(result, prefix+".flags."+flagName, flagName, flag)
	}

	if len(command.Examples) == 0 {
		AddWarning(result, "commands", prefix+".examples", "No examples provided", "NO_EXAMPLES", "Add usage examples to help users understand the command")
	}

	for _, relatedCmd := range command.RelatedCommands {
		if _, exists := v.config.Commands[relatedCmd]; !exists {
			AddWarning(result, "commands", prefix+".related_commands", "Related command '"+relatedCmd+"' does not exist", "UNDEFINED_RELATED_COMMAND", "Remove reference or define the related command")
		}
	}

	v.validateCommandAliases(result, cmdName, command)
}

// validateFlagDefinition validates a flag definition
func (v *CommandValidator) validateFlagDefinition(result *ValidationResult, prefix, flagName string, flag config.Flag) {
	validTypes := []string{"bool", "string", "int", "float", "duration", "stringArray", "intArray"}
	if !contains(validTypes, flag.Type) {
		AddError(result, "flags", prefix+".type", "Invalid flag type '"+flag.Type+"'", "INVALID_FLAG_TYPE", "high", "Use one of: bool, string, int, float, duration, stringArray, intArray")
	}

	if flag.Description == "" {
		AddError(result, "flags", prefix+".description", "Flag description is required", "MISSING_FLAG_DESCRIPTION", "medium", "Add description to flag "+flagName)
	}

	if flag.Short != "" && len(flag.Short) != 1 {
		AddError(result, "flags", prefix+".short", "Short flag must be single character", "INVALID_SHORT_FLAG", "medium", "Use single character for short flag")
	}

	if len(flag.Options) > 0 && flag.Type != "string" {
		AddWarning(result, "flags", prefix+".options", "Options should typically be used with string type flags", "OPTIONS_TYPE_MISMATCH", "Consider using string type for flags with options")
	}

	v.validateDefaultValueType(result, prefix, flag)
}

// validateDefaultValueType validates that default value matches flag type
func (v *CommandValidator) validateDefaultValueType(result *ValidationResult, prefix string, flag config.Flag) {
	if flag.Default == nil {
		return
	}

	defaultType := reflect.TypeOf(flag.Default).Kind()

	switch flag.Type {
	case "bool":
		if defaultType != reflect.Bool {
			AddError(result, "flags", prefix+".default", "Default value type mismatch for bool flag", "TYPE_MISMATCH", "medium", "Use boolean value for default")
		}
	case "int":
		if defaultType != reflect.Int && defaultType != reflect.Float64 {
			AddError(result, "flags", prefix+".default", "Default value type mismatch for int flag", "TYPE_MISMATCH", "medium", "Use integer value for default")
		}
	case "string":
		if defaultType != reflect.String {
			AddError(result, "flags", prefix+".default", "Default value type mismatch for string flag", "TYPE_MISMATCH", "medium", "Use string value for default")
		}
	}
}

// validateCommandAliases validates that command aliases don't conflict
func (v *CommandValidator) validateCommandAliases(result *ValidationResult, cmdName string, command config.Command) {
	for _, alias := range command.Aliases {
		if _, exists := v.config.Commands[alias]; exists {
			AddError(result, "commands", "commands."+cmdName+".aliases", "Alias '"+alias+"' conflicts with command name", "ALIAS_CONFLICT", "high", "Use a different alias that doesn't conflict with existing commands")
		}

		for otherCmdName, otherCmd := range v.config.Commands {
			if otherCmdName == cmdName {
				continue
			}
			if contains(otherCmd.Aliases, alias) {
				AddError(result, "commands", "commands."+cmdName+".aliases", "Alias '"+alias+"' conflicts with alias from command '"+otherCmdName+"'", "ALIAS_CONFLICT", "high", "Use a unique alias")
			}
		}
	}
}

// validateOrphanedCommands checks for commands not assigned to categories
func (v *CommandValidator) validateOrphanedCommands(result *ValidationResult) {
	categorizedCommands := make(map[string]bool)

	for _, category := range v.config.Categories {
		for _, cmdName := range category.Commands {
			categorizedCommands[cmdName] = true
		}
	}

	for cmdName := range v.config.Commands {
		if !categorizedCommands[cmdName] {
			AddWarning(result, "commands", "commands."+cmdName, "Command not assigned to any category", "ORPHANED_COMMAND", "Assign command to a category or create a new category")
		}
	}
}

// ValidateReferences validates cross-references between different sections
func (v *CommandValidator) ValidateReferences(result *ValidationResult) {
	// This method could be extended to validate references between
	// commands, workflows, profiles, etc.
}
