package config

import (
	"fmt"
	"strings"

	"github.com/isaacgarza/dev-stack/internal/pkg/constants"
)

// GenerateConfig generates a dev-stack configuration file
func GenerateConfig(projectName, environment string, services []string, validation, advanced map[string]bool) string {
	var builder strings.Builder

	// Header comment
	builder.WriteString(fmt.Sprintf("# %s Configuration\n", constants.AppNameTitle))
	builder.WriteString(fmt.Sprintf("# Generated for project: %s\n", projectName))
	builder.WriteString(fmt.Sprintf("# Documentation: %s\n\n", constants.ConfigDocsURL))

	// Project section
	builder.WriteString(fmt.Sprintf("%s:\n", constants.ProjectSection))
	builder.WriteString(fmt.Sprintf("  name: %s\n", projectName))
	builder.WriteString(fmt.Sprintf("  environment: %s\n\n", environment))

	// Stack section
	builder.WriteString(fmt.Sprintf("%s:\n", constants.StackSection))
	builder.WriteString("  enabled:\n")
	for _, service := range services {
		builder.WriteString(fmt.Sprintf("    - %s\n", service))
	}
	builder.WriteString("\n")

	// Overrides section
	builder.WriteString("# Service-specific overrides\n")
	builder.WriteString(fmt.Sprintf("# Service configuration options: %s\n", constants.ServiceConfigURL))
	builder.WriteString(fmt.Sprintf("%s: {}\n\n", constants.OverridesSection))

	// Validation section
	builder.WriteString(fmt.Sprintf("%s:\n", constants.ValidationSection))
	builder.WriteString(fmt.Sprintf("  skip_warnings: %t\n", getBoolValue(validation, "skip_warnings", constants.DefaultSkipWarnings)))
	builder.WriteString(fmt.Sprintf("  allow_multiple_databases: %t\n\n", getBoolValue(validation, "allow_multiple_databases", constants.DefaultAllowMultipleDBs)))

	// Advanced section
	builder.WriteString(fmt.Sprintf("%s:\n", constants.AdvancedSection))
	builder.WriteString(fmt.Sprintf("  auto_start: %t\n", getBoolValue(advanced, "auto_start", constants.DefaultAutoStart)))
	builder.WriteString(fmt.Sprintf("  pull_latest_images: %t\n", getBoolValue(advanced, "pull_latest_images", constants.DefaultPullLatestImages)))
	builder.WriteString(fmt.Sprintf("  cleanup_on_recreate: %t\n", getBoolValue(advanced, "cleanup_on_recreate", constants.DefaultCleanupOnRecreate)))

	return builder.String()
}

// getBoolValue gets a boolean value from map with fallback to default
func getBoolValue(m map[string]bool, key string, defaultValue bool) bool {
	if value, exists := m[key]; exists {
		return value
	}
	return defaultValue
}
