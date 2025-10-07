package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Loader handles loading and parsing of command configurations
type Loader struct {
	configPath string
	cache      *CommandConfig
}

// NewLoader creates a new configuration loader
func NewLoader(configPath string) *Loader {
	return &Loader{
		configPath: configPath,
	}
}

// Load loads and parses the command configuration from YAML
func (l *Loader) Load() (*CommandConfig, error) {
	if l.cache != nil {
		return l.cache, nil
	}

	// Resolve config path
	configPath, err := l.resolveConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve config path: %w", err)
	}

	// Read the YAML file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	// Parse YAML
	var config CommandConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config YAML: %w", err)
	}

	// Validate configuration
	if err := l.validateConfig(&config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	// Post-process configuration
	if err := l.postProcessConfig(&config); err != nil {
		return nil, fmt.Errorf("configuration post-processing failed: %w", err)
	}

	l.cache = &config
	return &config, nil
}

// LoadFromPath loads configuration from a specific path
func (l *Loader) LoadFromPath(path string) (*CommandConfig, error) {
	oldPath := l.configPath
	l.configPath = path
	l.cache = nil // Clear cache

	config, err := l.Load()
	l.configPath = oldPath // Restore original path
	return config, err
}

// Reload clears cache and reloads configuration
func (l *Loader) Reload() (*CommandConfig, error) {
	l.cache = nil
	return l.Load()
}

// GetConfigPath returns the resolved configuration path
func (l *Loader) GetConfigPath() (string, error) {
	return l.resolveConfigPath()
}

// resolveConfigPath resolves the configuration file path
func (l *Loader) resolveConfigPath() (string, error) {
	if l.configPath == "" {
		// Try default locations
		candidates := []string{
			"scripts/commands.yaml",
			"config/commands.yaml",
			".dev-stack/commands.yaml",
		}

		for _, candidate := range candidates {
			if _, err := os.Stat(candidate); err == nil {
				return filepath.Abs(candidate)
			}
		}

		return "", fmt.Errorf("no commands.yaml found in default locations: %v", candidates)
	}

	// Use provided path
	if !filepath.IsAbs(l.configPath) {
		return filepath.Abs(l.configPath)
	}

	return l.configPath, nil
}

// validateConfig performs basic structural validation
func (l *Loader) validateConfig(config *CommandConfig) error {
	if config.Metadata.Version == "" {
		return fmt.Errorf("metadata.version is required")
	}

	if len(config.Commands) == 0 {
		return fmt.Errorf("no commands defined")
	}

	// Validate each command has required fields
	for name, cmd := range config.Commands {
		if cmd.Description == "" {
			return fmt.Errorf("command %s: description is required", name)
		}
		if cmd.Usage == "" {
			return fmt.Errorf("command %s: usage is required", name)
		}
	}

	return nil
}

// postProcessConfig performs post-processing tasks
func (l *Loader) postProcessConfig(config *CommandConfig) error {
	// Add commands to categories if not already present
	for cmdName, cmd := range config.Commands {
		if cmd.Category != "" {
			if category, exists := config.Categories[cmd.Category]; exists {
				// Check if command is already in category
				found := false
				for _, catCmd := range category.Commands {
					if catCmd == cmdName {
						found = true
						break
					}
				}
				if !found {
					category.Commands = append(category.Commands, cmdName)
					config.Categories[cmd.Category] = category
				}
			}
		}
	}

	// Set default values for flags
	for cmdName, cmd := range config.Commands {
		for flagName, flag := range cmd.Flags {
			if flag.Type == "" {
				flag.Type = "string" // Default type
				cmd.Flags[flagName] = flag
			}
		}
		config.Commands[cmdName] = cmd
	}

	return nil
}

// LoadDefault loads the default commands configuration
func LoadDefault() (*CommandConfig, error) {
	loader := NewLoader("")
	return loader.Load()
}

// LoadFromBytes loads configuration from byte data
func LoadFromBytes(data []byte) (*CommandConfig, error) {
	var config CommandConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config YAML: %w", err)
	}

	loader := &Loader{}
	if err := loader.validateConfig(&config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	if err := loader.postProcessConfig(&config); err != nil {
		return nil, fmt.Errorf("configuration post-processing failed: %w", err)
	}

	return &config, nil
}

// SaveConfig saves a configuration to a YAML file
func SaveConfig(config *CommandConfig, path string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file %s: %w", path, err)
	}

	return nil
}

// MergeConfigs merges multiple configurations, with later configs taking precedence
func MergeConfigs(configs ...*CommandConfig) (*CommandConfig, error) {
	if len(configs) == 0 {
		return nil, fmt.Errorf("no configurations to merge")
	}

	result := &CommandConfig{
		Metadata:   configs[0].Metadata,
		Global:     GlobalConfig{Flags: make(map[string]Flag)},
		Categories: make(map[string]Category),
		Commands:   make(map[string]Command),
		Workflows:  make(map[string]Workflow),
		Profiles:   make(map[string]Profile),
		Help:       make(map[string]string),
	}

	for _, config := range configs {
		// Merge metadata (last wins)
		result.Metadata = config.Metadata

		// Merge global flags
		for name, flag := range config.Global.Flags {
			result.Global.Flags[name] = flag
		}

		// Merge categories
		for name, category := range config.Categories {
			result.Categories[name] = category
		}

		// Merge commands
		for name, command := range config.Commands {
			result.Commands[name] = command
		}

		// Merge workflows
		for name, workflow := range config.Workflows {
			result.Workflows[name] = workflow
		}

		// Merge profiles
		for name, profile := range config.Profiles {
			result.Profiles[name] = profile
		}

		// Merge help
		for name, help := range config.Help {
			result.Help[name] = help
		}
	}

	return result, nil
}
