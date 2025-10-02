package docs

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	// Markers for auto-generated sections
	StartMarker = "<!-- AUTO-GENERATED-START -->"
	EndMarker   = "<!-- AUTO-GENERATED-END -->"
)

// Generator handles markdown generation from parsed manifests
type Generator struct {
	options *GenerationOptions
	parser  *Parser
}

// NewGenerator creates a new Generator instance
func NewGenerator(options *GenerationOptions) *Generator {
	return &Generator{
		options: options,
		parser:  NewParser(options),
	}
}

// GenerateCommandReference generates the command reference documentation
func (g *Generator) GenerateCommandReference(commands *CommandsManifest) (string, error) {
	var content strings.Builder

	content.WriteString("# Command Reference (dev-stack)\n\n")
	content.WriteString("This section is auto-generated from `scripts/commands.yaml`.\n\n")

	for script, cmds := range *commands {
		content.WriteString(fmt.Sprintf("## %s\n\n", script))
		for _, cmd := range cmds {
			content.WriteString(fmt.Sprintf("- `%s`\n", cmd))
		}
		content.WriteString("\n")
	}

	return content.String(), nil
}

// GenerateServicesGuide generates the services guide documentation
func (g *Generator) GenerateServicesGuide(services *ServicesManifest) (string, error) {
	var content strings.Builder

	content.WriteString("# Services Guide (dev-stack)\n\n")
	content.WriteString("This section is auto-generated from `services/services.yaml`.\n\n")

	for serviceName, info := range *services {
		content.WriteString(fmt.Sprintf("## %s\n\n", serviceName))
		content.WriteString(fmt.Sprintf("%s\n\n", info.Description))

		if len(info.Options) > 0 {
			content.WriteString("**Options:**\n")
			for _, opt := range info.Options {
				content.WriteString(fmt.Sprintf("- `%s`\n", opt))
			}
			content.WriteString("\n")
		}

		if len(info.Examples) > 0 {
			content.WriteString("**Examples:**\n")
			for _, example := range info.Examples {
				content.WriteString(fmt.Sprintf("- `%s`\n", example))
			}
			content.WriteString("\n")
		}

		if info.UsageNotes != "" {
			content.WriteString(fmt.Sprintf("**Usage Notes:** %s\n\n", info.UsageNotes))
		}

		if len(info.Links) > 0 {
			content.WriteString("**Links:**\n")
			for _, link := range info.Links {
				content.WriteString(fmt.Sprintf("- [%s](%s)\n", link, link))
			}
			content.WriteString("\n")
		}
	}

	return content.String(), nil
}

// updateAutoGenSection updates the auto-generated section in an existing file
func (g *Generator) updateAutoGenSection(filePath, generatedContent string) error {
	// Read existing file
	content, err := os.ReadFile(filePath)
	if err != nil {
		// If file doesn't exist, create it with the generated content
		if os.IsNotExist(err) {
			return g.createNewDocFile(filePath, generatedContent)
		}
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	doc := string(content)
	startIndex := strings.Index(doc, StartMarker)
	endIndex := strings.Index(doc, EndMarker)

	if startIndex == -1 || endIndex == -1 || endIndex < startIndex {
		return fmt.Errorf("auto-generation markers not found or invalid in %s", filePath)
	}

	// Calculate the position after the start marker
	startPos := startIndex + len(StartMarker)

	// Build new document
	newDoc := doc[:startPos] + "\n" + strings.TrimSpace(generatedContent) + "\n" + doc[endIndex:]

	if g.options.DryRun {
		if g.options.Verbose {
			fmt.Printf("Would update auto-generated section in %s\n", filePath)
		}
		return nil
	}

	// Write updated content
	if err := os.WriteFile(filePath, []byte(newDoc), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	if g.options.Verbose {
		fmt.Printf("Updated auto-generated section in %s\n", filePath)
	}

	return nil
}

// createNewDocFile creates a new documentation file with the generated content
func (g *Generator) createNewDocFile(filePath, generatedContent string) error {
	// Create a basic template with auto-generation markers
	template := fmt.Sprintf(`%s
%s
%s
`, StartMarker, strings.TrimSpace(generatedContent), EndMarker)

	if g.options.DryRun {
		if g.options.Verbose {
			fmt.Printf("Would create new file %s\n", filePath)
		}
		return nil
	}

	// Ensure directory exists
	dir := strings.TrimSuffix(filePath, "/"+strings.Split(filePath, "/")[len(strings.Split(filePath, "/"))-1])
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	if err := os.WriteFile(filePath, []byte(template), 0644); err != nil {
		return fmt.Errorf("failed to create file %s: %w", filePath, err)
	}

	if g.options.Verbose {
		fmt.Printf("Created new documentation file %s\n", filePath)
	}

	return nil
}

// GenerateAll generates all documentation files
func (g *Generator) GenerateAll() (*GenerationResult, error) {
	result := &GenerationResult{
		GeneratedAt:  time.Now(),
		FilesUpdated: make([]string, 0),
		Errors:       make([]error, 0),
	}

	// Parse manifests
	commands, services, err := g.parser.ParseAll()
	if err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("failed to parse manifests: %w", err))
		return result, err
	}

	// Generate command reference
	if err := g.generateCommandReference(commands, result); err != nil {
		result.Errors = append(result.Errors, err)
	}

	// Generate services guide
	if err := g.generateServicesGuide(services, result); err != nil {
		result.Errors = append(result.Errors, err)
	}

	return result, nil
}

// generateCommandReference generates the command reference documentation
func (g *Generator) generateCommandReference(commands *CommandsManifest, result *GenerationResult) error {
	content, err := g.GenerateCommandReference(commands)
	if err != nil {
		return fmt.Errorf("failed to generate command reference: %w", err)
	}

	if err := g.updateAutoGenSection(g.options.ReferenceMDPath, content); err != nil {
		return fmt.Errorf("failed to update command reference file: %w", err)
	}

	result.CommandsGenerated = true
	result.FilesUpdated = append(result.FilesUpdated, g.options.ReferenceMDPath)
	return nil
}

// generateServicesGuide generates the services guide documentation
func (g *Generator) generateServicesGuide(services *ServicesManifest, result *GenerationResult) error {
	content, err := g.GenerateServicesGuide(services)
	if err != nil {
		return fmt.Errorf("failed to generate services guide: %w", err)
	}

	if err := g.updateAutoGenSection(g.options.ServicesMDPath, content); err != nil {
		return fmt.Errorf("failed to update services guide file: %w", err)
	}

	result.ServicesGenerated = true
	result.FilesUpdated = append(result.FilesUpdated, g.options.ServicesMDPath)
	return nil
}

// GenerateCommandReferenceOnly generates only the command reference
func (g *Generator) GenerateCommandReferenceOnly() error {
	commands, err := g.parser.ParseCommands()
	if err != nil {
		return fmt.Errorf("failed to parse commands: %w", err)
	}

	if err := g.parser.ValidateCommandsManifest(commands); err != nil {
		return fmt.Errorf("commands validation failed: %w", err)
	}

	content, err := g.GenerateCommandReference(commands)
	if err != nil {
		return fmt.Errorf("failed to generate command reference: %w", err)
	}

	return g.updateAutoGenSection(g.options.ReferenceMDPath, content)
}

// GenerateServicesGuideOnly generates only the services guide
func (g *Generator) GenerateServicesGuideOnly() error {
	services, err := g.parser.ParseServices()
	if err != nil {
		return fmt.Errorf("failed to parse services: %w", err)
	}

	if err := g.parser.ValidateServicesManifest(services); err != nil {
		return fmt.Errorf("services validation failed: %w", err)
	}

	content, err := g.GenerateServicesGuide(services)
	if err != nil {
		return fmt.Errorf("failed to generate services guide: %w", err)
	}

	return g.updateAutoGenSection(g.options.ServicesMDPath, content)
}
