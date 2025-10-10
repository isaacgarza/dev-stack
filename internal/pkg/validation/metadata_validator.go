package validation

import (
	"strings"

	"github.com/isaacgarza/dev-stack/internal/pkg/config"
)

// MetadataValidator validates metadata and global configuration
type MetadataValidator struct {
	config *config.CommandConfig
}

// NewMetadataValidator creates a new metadata validator
func NewMetadataValidator(config *config.CommandConfig) *MetadataValidator {
	return &MetadataValidator{
		config: config,
	}
}

// ValidateMetadata validates the metadata section
func (v *MetadataValidator) ValidateMetadata(result *ValidationResult) {
	metadata := v.config.Metadata

	if metadata.Version == "" {
		AddError(result, "metadata", "metadata.version", "Version is required", "MISSING_VERSION", "critical", "Add a version field to metadata section")
	}

	if metadata.CLIVersion == "" {
		AddError(result, "metadata", "metadata.cli_version", "CLI version is required", "MISSING_CLI_VERSION", "critical", "Add a cli_version field to metadata section")
	}

	if metadata.Description == "" {
		AddWarning(result, "metadata", "metadata.description", "Description is recommended", "MISSING_DESCRIPTION", "Add a description field to metadata section")
	}

	if metadata.Version != "" && !isValidVersionFormat(metadata.Version) {
		AddError(result, "metadata", "metadata.version", "Invalid version format", "INVALID_VERSION_FORMAT", "high", "Use semantic versioning format (e.g., 2.0.0)")
	}
}

// ValidateGlobalConfiguration validates global configuration
func (v *MetadataValidator) ValidateGlobalConfiguration(result *ValidationResult) {
	global := v.config.Global

	requiredGlobalFlags := []string{"config", "verbose", "help"}
	for _, flagName := range requiredGlobalFlags {
		if _, exists := global.Flags[flagName]; !exists {
			AddWarning(result, "global", "global.flags."+flagName, "Recommended global flag missing", "MISSING_GLOBAL_FLAG", "Consider adding the "+flagName+" global flag")
		}
	}

	for flagName, flag := range global.Flags {
		v.validateFlagDefinition(result, "global.flags."+flagName, flagName, flag)
	}
}

// validateFlagDefinition validates a flag definition
func (v *MetadataValidator) validateFlagDefinition(result *ValidationResult, prefix, flagName string, flag config.Flag) {
	validTypes := []string{"bool", "string", "int", "float", "duration", "stringArray", "intArray"}
	if !contains(validTypes, flag.Type) {
		AddError(result, "flags", prefix+".type", "Invalid flag type '"+flag.Type+"'", "INVALID_FLAG_TYPE", "high", "Use one of: "+strings.Join(validTypes, ", "))
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
}

func isValidVersionFormat(version string) bool {
	parts := strings.Split(version, ".")
	return len(parts) >= 2 && len(parts) <= 3
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
