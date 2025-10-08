package config

import (
	"time"
)

// CommandConfig represents the root structure of the enhanced commands.yaml
type CommandConfig struct {
	Metadata   Metadata            `yaml:"metadata"`
	Global     GlobalConfig        `yaml:"global"`
	Categories map[string]Category `yaml:"categories"`
	Commands   map[string]Command  `yaml:"commands"`
	Workflows  map[string]Workflow `yaml:"workflows"`
	Profiles   map[string]Profile  `yaml:"profiles"`
	Help       map[string]string   `yaml:"help"`
}

// Metadata contains version and generation information
type Metadata struct {
	Version     string    `yaml:"version"`
	GeneratedAt time.Time `yaml:"generated_at"`
	CLIVersion  string    `yaml:"cli_version"`
	Description string    `yaml:"description"`
}

// GlobalConfig contains global CLI configuration
type GlobalConfig struct {
	Flags map[string]Flag `yaml:"flags"`
}

// Category represents a command category for organization
type Category struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Icon        string   `yaml:"icon"`
	Commands    []string `yaml:"commands"`
}

// Command represents a complete command definition
type Command struct {
	Category        string           `yaml:"category"`
	Description     string           `yaml:"description"`
	LongDescription string           `yaml:"long_description"`
	Usage           string           `yaml:"usage"`
	Aliases         []string         `yaml:"aliases"`
	Examples        []Example        `yaml:"examples"`
	Flags           map[string]Flag  `yaml:"flags"`
	RelatedCommands []string         `yaml:"related_commands"`
	Tips            []string         `yaml:"tips"`
	Hidden          bool             `yaml:"hidden,omitempty"`
	Deprecated      *DeprecationInfo `yaml:"deprecated,omitempty"`
}

// Flag represents a command line flag definition
type Flag struct {
	Short       string      `yaml:"short,omitempty"`
	Type        string      `yaml:"type"`
	Description string      `yaml:"description"`
	Default     interface{} `yaml:"default"`
	Options     []string    `yaml:"options,omitempty"`
	Completion  string      `yaml:"completion,omitempty"`
	Required    bool        `yaml:"required,omitempty"`
	Hidden      bool        `yaml:"hidden,omitempty"`
	Deprecated  string      `yaml:"deprecated,omitempty"`
}

// Example represents a command usage example
type Example struct {
	Command     string `yaml:"command"`
	Description string `yaml:"description"`
}

// DeprecationInfo contains deprecation details
type DeprecationInfo struct {
	Since       string `yaml:"since"`
	Reason      string `yaml:"reason"`
	Alternative string `yaml:"alternative"`
}

// Workflow represents a sequence of commands for common tasks
type Workflow struct {
	Name        string         `yaml:"name"`
	Description string         `yaml:"description"`
	Steps       []WorkflowStep `yaml:"steps"`
}

// WorkflowStep represents a single step in a workflow
type WorkflowStep struct {
	Command     string `yaml:"command"`
	Description string `yaml:"description"`
	Optional    bool   `yaml:"optional,omitempty"`
}

// Profile represents a predefined service combination
type Profile struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Services    []string `yaml:"services"`
}

// ValidationResult represents the result of configuration validation
type ValidationResult struct {
	Valid    bool              `yaml:"valid"`
	Errors   []ValidationError `yaml:"errors,omitempty"`
	Warnings []ValidationError `yaml:"warnings,omitempty"`
}

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string `yaml:"field"`
	Message string `yaml:"message"`
	Code    string `yaml:"code"`
}

// FlagType represents supported flag types
type FlagType string

const (
	FlagTypeBool        FlagType = "bool"
	FlagTypeString      FlagType = "string"
	FlagTypeInt         FlagType = "int"
	FlagTypeFloat       FlagType = "float"
	FlagTypeDuration    FlagType = "duration"
	FlagTypeStringArray FlagType = "stringArray"
	FlagTypeIntArray    FlagType = "intArray"
)

// CommandContext provides context for command execution
type CommandContext struct {
	Config     *CommandConfig
	WorkingDir string
	ConfigFile string
	Verbose    bool
	DryRun     bool
}

// GetCommand returns a command by name
func (c *CommandConfig) GetCommand(name string) (*Command, bool) {
	cmd, exists := c.Commands[name]
	return &cmd, exists
}

// GetCategory returns a category by name
func (c *CommandConfig) GetCategory(name string) (*Category, bool) {
	cat, exists := c.Categories[name]
	return &cat, exists
}

// GetProfile returns a profile by name
func (c *CommandConfig) GetProfile(name string) (*Profile, bool) {
	profile, exists := c.Profiles[name]
	return &profile, exists
}

// GetWorkflow returns a workflow by name
func (c *CommandConfig) GetWorkflow(name string) (*Workflow, bool) {
	workflow, exists := c.Workflows[name]
	return &workflow, exists
}

// GetAllCommandNames returns all command names
func (c *CommandConfig) GetAllCommandNames() []string {
	names := make([]string, 0, len(c.Commands))
	for name := range c.Commands {
		names = append(names, name)
	}
	return names
}

// GetCommandsByCategory returns commands in a specific category
func (c *CommandConfig) GetCommandsByCategory(categoryName string) []string {
	category, exists := c.Categories[categoryName]
	if !exists {
		return nil
	}
	return category.Commands
}

// GetAllCategories returns all category names
func (c *CommandConfig) GetAllCategories() []string {
	names := make([]string, 0, len(c.Categories))
	for name := range c.Categories {
		names = append(names, name)
	}
	return names
}

// GetAllProfiles returns all profile names
func (c *CommandConfig) GetAllProfiles() []string {
	names := make([]string, 0, len(c.Profiles))
	for name := range c.Profiles {
		names = append(names, name)
	}
	return names
}

// GetAllWorkflows returns all workflow names
func (c *CommandConfig) GetAllWorkflows() []string {
	names := make([]string, 0, len(c.Workflows))
	for name := range c.Workflows {
		names = append(names, name)
	}
	return names
}

// Validate performs comprehensive validation of the configuration
func (c *CommandConfig) Validate() *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Validate metadata
	if c.Metadata.Version == "" {
		result.addError("metadata.version", "Version is required", "MISSING_VERSION")
	}

	// Validate commands exist in categories
	for catName, category := range c.Categories {
		for _, cmdName := range category.Commands {
			if _, exists := c.Commands[cmdName]; !exists {
				result.addError(
					"categories."+catName+".commands",
					"Command '"+cmdName+"' referenced in category but not defined",
					"UNDEFINED_COMMAND",
				)
			}
		}
	}

	// Validate command categories exist
	for cmdName, command := range c.Commands {
		if command.Category != "" {
			if _, exists := c.Categories[command.Category]; !exists {
				result.addError(
					"commands."+cmdName+".category",
					"Category '"+command.Category+"' does not exist",
					"UNDEFINED_CATEGORY",
				)
			}
		}
	}

	// Validate related commands exist
	for cmdName, command := range c.Commands {
		for _, relatedCmd := range command.RelatedCommands {
			if _, exists := c.Commands[relatedCmd]; !exists {
				result.addWarning(
					"commands."+cmdName+".related_commands",
					"Related command '"+relatedCmd+"' does not exist",
					"UNDEFINED_RELATED_COMMAND",
				)
			}
		}
	}

	// Validate flag types
	for cmdName, command := range c.Commands {
		for flagName, flag := range command.Flags {
			if !isValidFlagType(flag.Type) {
				result.addError(
					"commands."+cmdName+".flags."+flagName+".type",
					"Invalid flag type '"+flag.Type+"'",
					"INVALID_FLAG_TYPE",
				)
			}
		}
	}

	result.Valid = len(result.Errors) == 0
	return result
}

// Helper methods for ValidationResult
func (v *ValidationResult) addError(field, message, code string) {
	v.Errors = append(v.Errors, ValidationError{
		Field:   field,
		Message: message,
		Code:    code,
	})
}

func (v *ValidationResult) addWarning(field, message, code string) {
	v.Warnings = append(v.Warnings, ValidationError{
		Field:   field,
		Message: message,
		Code:    code,
	})
}

// isValidFlagType checks if a flag type is supported
func isValidFlagType(flagType string) bool {
	validTypes := []FlagType{
		FlagTypeBool,
		FlagTypeString,
		FlagTypeInt,
		FlagTypeFloat,
		FlagTypeDuration,
		FlagTypeStringArray,
		FlagTypeIntArray,
	}

	for _, validType := range validTypes {
		if string(validType) == flagType {
			return true
		}
	}
	return false
}
