package docs

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Parser handles parsing of YAML manifests
type Parser struct {
	options *GenerationOptions
}

// NewParser creates a new Parser instance
func NewParser(options *GenerationOptions) *Parser {
	return &Parser{
		options: options,
	}
}

// ParseCommands reads and parses the commands.yaml file
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

// ValidateCommandsManifest validates the commands manifest structure
func (p *Parser) ValidateCommandsManifest(commands *CommandsManifest) error {
	if commands == nil {
		return fmt.Errorf("commands manifest is nil")
	}

	if len(*commands) == 0 {
		return fmt.Errorf("commands manifest is empty")
	}

	for script, cmds := range *commands {
		if script == "" {
			return fmt.Errorf("found empty script name")
		}
		if len(cmds) == 0 {
			return fmt.Errorf("script %s has no commands", script)
		}
	}

	return nil
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
