package docs

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Parser handles parsing of enhanced YAML manifests
type Parser struct {
	options *GenerationOptions
}

// NewParser creates a new Parser instance
func NewParser(options *GenerationOptions) *Parser {
	return &Parser{
		options: options,
	}
}

// ParseCommands reads and parses the enhanced commands.yaml file
func (p *Parser) ParseCommands() (*CommandsManifest, error) {
	data, err := os.ReadFile(p.options.CommandsYAMLPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read commands YAML file %s: %w", p.options.CommandsYAMLPath, err)
	}

	var commands CommandsManifest
	if err := yaml.Unmarshal(data, &commands); err != nil {
		return nil, fmt.Errorf("failed to parse commands YAML: %w", err)
	}

	return &commands, nil
}

// ParseServices reads and parses the services.yaml file
func (p *Parser) ParseServices() (*ServicesManifest, error) {
	data, err := os.ReadFile(p.options.ServicesYAMLPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read services YAML file %s: %w", p.options.ServicesYAMLPath, err)
	}

	var services ServicesManifest
	if err := yaml.Unmarshal(data, &services); err != nil {
		return nil, fmt.Errorf("failed to parse services YAML: %w", err)
	}

	return &services, nil
}

// ValidateCommandsManifest validates the enhanced commands manifest structure
func (p *Parser) ValidateCommandsManifest(commands *CommandsManifest) error {
	if commands == nil {
		return fmt.Errorf("commands manifest is nil")
	}

	// Validate metadata
	if commands.Metadata.Version == "" {
		return fmt.Errorf("metadata.version is required")
	}

	if commands.Metadata.CLIVersion == "" {
		return fmt.Errorf("metadata.cli_version is required")
	}

	if len(commands.Commands) == 0 {
		return fmt.Errorf("no commands defined")
	}

	// Validate each command
	for name, cmd := range commands.Commands {
		if cmd.Description == "" {
			return fmt.Errorf("command %s: description is required", name)
		}
		if cmd.Usage == "" {
			return fmt.Errorf("command %s: usage is required", name)
		}

		// Validate category exists if specified
		if cmd.Category != "" {
			if _, exists := commands.Categories[cmd.Category]; !exists {
				return fmt.Errorf("command %s: category %s does not exist", name, cmd.Category)
			}
		}

		// Validate flag types
		for flagName, flag := range cmd.Flags {
			if !isValidFlagType(flag.Type) {
				return fmt.Errorf("command %s, flag %s: invalid type %s", name, flagName, flag.Type)
			}
		}
	}

	// Validate categories reference existing commands
	for catName, category := range commands.Categories {
		for _, cmdName := range category.Commands {
			if _, exists := commands.Commands[cmdName]; !exists {
				return fmt.Errorf("category %s references undefined command %s", catName, cmdName)
			}
		}
	}

	// Validate workflows reference existing commands
	for workflowName, workflow := range commands.Workflows {
		for i, step := range workflow.Steps {
			// Basic validation - could be enhanced to parse actual commands
			if step.Command == "" {
				return fmt.Errorf("workflow %s, step %d: command is required", workflowName, i+1)
			}
		}
	}

	return nil
}

// isValidFlagType checks if a flag type is valid
func isValidFlagType(flagType string) bool {
	validTypes := []string{"bool", "string", "int", "float", "duration", "stringArray", "intArray"}
	for _, validType := range validTypes {
		if flagType == validType {
			return true
		}
	}
	return false
}

// ValidateServicesManifest validates the services manifest structure
func (p *Parser) ValidateServicesManifest(services *ServicesManifest) error {
	if services == nil {
		return fmt.Errorf("services manifest is nil")
	}

	if len(*services) == 0 {
		return fmt.Errorf("services manifest is empty")
	}

	for serviceName, info := range *services {
		if serviceName == "" {
			return fmt.Errorf("found empty service name")
		}
		if info.Description == "" {
			return fmt.Errorf("service %s has no description", serviceName)
		}
	}

	return nil
}

// ParseAll parses both commands and services manifests
func (p *Parser) ParseAll() (*CommandsManifest, *ServicesManifest, error) {
	commands, err := p.ParseCommands()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse commands: %w", err)
	}

	if err := p.ValidateCommandsManifest(commands); err != nil {
		return nil, nil, fmt.Errorf("commands validation failed: %w", err)
	}

	services, err := p.ParseServices()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse services: %w", err)
	}

	if err := p.ValidateServicesManifest(services); err != nil {
		return nil, nil, fmt.Errorf("services validation failed: %w", err)
	}

	return commands, services, nil
}
