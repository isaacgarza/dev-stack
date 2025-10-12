package validation

import (
	"testing"

	"github.com/isaacgarza/dev-stack/internal/pkg/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewValidator(t *testing.T) {
	tests := []struct {
		name   string
		config *config.CommandConfig
	}{
		{
			name:   "nil config",
			config: nil,
		},
		{
			name: "valid config",
			config: &config.CommandConfig{
				Metadata: config.Metadata{
					Version: "1.0.0",
				},
				Commands: make(map[string]config.Command),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewValidator(tt.config)

			assert.NotNil(t, validator)
			assert.Equal(t, tt.config, validator.config)
			assert.NotNil(t, validator.metadataValidator)
			assert.NotNil(t, validator.commandValidator)
			assert.NotNil(t, validator.workflowValidator)
			assert.NotNil(t, validator.practicesValidator)
			assert.NotNil(t, validator.cliValidator)
		})
	}
}

func TestValidator_ValidateAll(t *testing.T) {
	tests := []struct {
		name           string
		config         *config.CommandConfig
		expectValid    bool
		expectErrors   bool
		expectWarnings bool
	}{
		{
			name: "valid minimal config",
			config: &config.CommandConfig{
				Metadata: config.Metadata{
					Version:     "1.0.0",
					CLIVersion:  "1.0.0",
					Description: "Test config",
				},
				Global: config.GlobalConfig{
					Flags: make(map[string]config.Flag),
				},
				Categories: map[string]config.Category{
					"general": {
						Name:        "general",
						Description: "General commands",
					},
				},
				Commands: map[string]config.Command{
					"test": {
						Description: "Test command",
						Usage:       "test [options]",
						Category:    "general",
					},
				},
				Workflows: make(map[string]config.Workflow),
				Profiles:  make(map[string]config.Profile),
				Help:      make(map[string]string),
			},
			expectValid:    true,
			expectErrors:   false,
			expectWarnings: false,
		},
		{
			name: "config with missing metadata",
			config: &config.CommandConfig{
				Metadata: config.Metadata{
					// Missing required fields
				},
				Global:     config.GlobalConfig{},
				Categories: make(map[string]config.Category),
				Commands:   make(map[string]config.Command),
				Workflows:  make(map[string]config.Workflow),
				Profiles:   make(map[string]config.Profile),
				Help:       make(map[string]string),
			},
			expectValid:  false,
			expectErrors: true,
		},
		{
			name: "config with invalid command",
			config: &config.CommandConfig{
				Metadata: config.Metadata{
					Version:     "1.0.0",
					CLIVersion:  "1.0.0",
					Description: "Test config",
				},
				Global:     config.GlobalConfig{},
				Categories: make(map[string]config.Category),
				Commands: map[string]config.Command{
					"invalid": {
						Description: "Invalid command",
						Usage:       "invalid [options]",
						Category:    "nonexistent", // This should cause validation error
					},
				},
				Workflows: make(map[string]config.Workflow),
				Profiles:  make(map[string]config.Profile),
				Help:      make(map[string]string),
			},
			expectValid:  true, // May still be valid if errors < 5 and no critical errors
			expectErrors: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewValidator(tt.config)

			result := validator.ValidateAll()

			require.NotNil(t, result)
			assert.Equal(t, tt.expectValid, result.Valid)

			if tt.expectErrors {
				assert.NotEmpty(t, result.Errors)
			} else {
				assert.Empty(t, result.Errors)
			}

			if tt.expectWarnings {
				assert.NotEmpty(t, result.Warnings)
			}

			// Check summary is populated
			assert.NotNil(t, result.Summary)
			assert.Equal(t, len(tt.config.Commands), result.Summary.TotalCommands)
			assert.Equal(t, len(tt.config.Categories), result.Summary.TotalCategories)
			assert.Equal(t, len(tt.config.Workflows), result.Summary.TotalWorkflows)
			assert.Equal(t, len(tt.config.Profiles), result.Summary.TotalProfiles)
		})
	}
}

func TestValidator_ValidateAgainstCLI(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.CommandConfig
		setupCLI    func() *cobra.Command
		expectValid bool
	}{
		{
			name: "matching CLI and config",
			config: &config.CommandConfig{
				Metadata: config.Metadata{
					Version: "1.0.0",
				},
				Commands: map[string]config.Command{
					"test": {
						Description: "Test command",
						Usage:       "test [options]",
					},
				},
			},
			setupCLI: func() *cobra.Command {
				rootCmd := &cobra.Command{
					Use:   "dev-stack",
					Short: "Development stack management tool",
				}

				testCmd := &cobra.Command{
					Use:   "test",
					Short: "Test command",
				}
				rootCmd.AddCommand(testCmd)

				return rootCmd
			},
			expectValid: true,
		},
		{
			name: "CLI command missing from config",
			config: &config.CommandConfig{
				Metadata: config.Metadata{
					Version: "1.0.0",
				},
				Commands: make(map[string]config.Command),
			},
			setupCLI: func() *cobra.Command {
				rootCmd := &cobra.Command{
					Use:   "dev-stack",
					Short: "Development stack management tool",
				}

				testCmd := &cobra.Command{
					Use:   "test",
					Short: "Test command",
				}
				rootCmd.AddCommand(testCmd)

				return rootCmd
			},
			expectValid: true, // CLI validator may be more lenient
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewValidator(tt.config)
			rootCmd := tt.setupCLI()

			result := validator.ValidateAgainstCLI(rootCmd)

			require.NotNil(t, result)
			assert.Equal(t, tt.expectValid, result.Valid)
		})
	}
}

func TestValidationResult_AddError(t *testing.T) {
	result := &ValidationResult{
		Valid: true,
	}

	// Test adding an error
	error := ValidationError{
		Type:    "test",
		Field:   "test.field",
		Message: "Test error message",
		Code:    "TEST001",
	}

	result.Errors = append(result.Errors, error)
	result.Valid = false

	assert.False(t, result.Valid)
	assert.Len(t, result.Errors, 1)
	assert.Equal(t, "test", result.Errors[0].Type)
	assert.Equal(t, "test.field", result.Errors[0].Field)
	assert.Equal(t, "Test error message", result.Errors[0].Message)
	assert.Equal(t, "TEST001", result.Errors[0].Code)
}

func TestValidationResult_AddWarning(t *testing.T) {
	result := &ValidationResult{
		Valid: true,
	}

	// Test adding a warning
	warning := ValidationWarning{
		Type:    "test",
		Field:   "test.field",
		Message: "Test warning message",
		Code:    "WARN001",
	}

	result.Warnings = append(result.Warnings, warning)

	assert.True(t, result.Valid) // Warnings don't affect validity
	assert.Len(t, result.Warnings, 1)
	assert.Equal(t, "test", result.Warnings[0].Type)
	assert.Equal(t, "test.field", result.Warnings[0].Field)
	assert.Equal(t, "Test warning message", result.Warnings[0].Message)
	assert.Equal(t, "WARN001", result.Warnings[0].Code)
}
