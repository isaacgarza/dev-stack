package cli

import (
	"fmt"
	"os"

	"github.com/isaacgarza/dev-stack/internal/pkg/cli"
	"github.com/isaacgarza/dev-stack/internal/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// CreateRootCommand creates the root command using the YAML-driven factory
func CreateRootCommand() (*cobra.Command, error) {
	// Load command configuration
	loader := config.NewLoader("")
	commandConfig, err := loader.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load command configuration: %w", err)
	}

	// Validate configuration
	validationResult := commandConfig.Validate()
	if !validationResult.Valid {
		fmt.Fprintf(os.Stderr, "Warning: Command configuration has validation errors:\n")
		for _, err := range validationResult.Errors {
			fmt.Fprintf(os.Stderr, "  - %s: %s\n", err.Field, err.Message)
		}
		if len(validationResult.Warnings) > 0 {
			fmt.Fprintf(os.Stderr, "Warnings:\n")
			for _, warning := range validationResult.Warnings {
				fmt.Fprintf(os.Stderr, "  - %s: %s\n", warning.Field, warning.Message)
			}
		}
	}

	// Create factory
	factory := cli.NewFactory(commandConfig)

	// Create root command using factory
	rootCmd, err := factory.CreateRootCommand()
	if err != nil {
		return nil, fmt.Errorf("failed to create root command: %w", err)
	}

	// Initialize configuration
	cobra.OnInitialize(func() {
		initFactoryConfig(commandConfig)
	})

	return rootCmd, nil
}

// ExecuteFactory executes the root command using the factory
func ExecuteFactory() error {
	rootCmd, err := CreateRootCommand()
	if err != nil {
		return fmt.Errorf("failed to create CLI: %w", err)
	}

	return rootCmd.Execute()
}

// initFactoryConfig reads in config file and ENV variables if set
func initFactoryConfig(commandConfig *config.CommandConfig) {
	var cfgFile string
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in multiple locations
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.AddConfigPath(".dev-stack")
		viper.SetConfigType("yaml")
		
		// Try to find config file with multiple names
		configNames := []string{"dev-stack-config", ".dev-stack"}
		var configFound bool
		for _, name := range configNames {
			viper.SetConfigName(name)
			if err := viper.ReadInConfig(); err == nil {
				configFound = true
				break
			}
		}
		
		// If no config found, don't call ReadInConfig again
		if configFound {
			return
		}
	}

	// read in environment variables that match
	viper.AutomaticEnv()

	// If a config file is found, read it in (only if not already read above)
	if err := viper.ReadInConfig(); err == nil {
		if viper.GetBool("verbose") {
			fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}
	}
}

// GetCommandConfig loads and returns the command configuration
func GetCommandConfig() (*config.CommandConfig, error) {
	loader := config.NewLoader("")
	return loader.Load()
}

// ValidateConfig validates the current configuration
func ValidateConfig() error {
	commandConfig, err := GetCommandConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	result := commandConfig.Validate()
	if !result.Valid {
		fmt.Printf("Configuration validation failed with %d errors:\n", len(result.Errors))
		for _, err := range result.Errors {
			fmt.Printf("  - %s: %s\n", err.Field, err.Message)
		}
		return fmt.Errorf("configuration validation failed")
	}

	fmt.Println("✅ Configuration is valid")
	if len(result.Warnings) > 0 {
		fmt.Printf("⚠️  %d warnings:\n", len(result.Warnings))
		for _, warning := range result.Warnings {
			fmt.Printf("  - %s: %s\n", warning.Field, warning.Message)
		}
	}

	return nil
}
