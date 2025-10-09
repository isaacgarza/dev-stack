package docs

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"
)

// EnhancedGenerator generates rich documentation from enhanced YAML configuration
type EnhancedGenerator struct {
	options *GenerationOptions
	parser  *Parser
}

// NewEnhancedGenerator creates a new enhanced documentation generator
func NewEnhancedGenerator(options *GenerationOptions) *EnhancedGenerator {
	return &EnhancedGenerator{
		options: options,
		parser:  NewParser(options),
	}
}

// GenerateAll generates all documentation using enhanced structure
func (g *EnhancedGenerator) GenerateAll() (*GenerationResult, error) {
	result := &GenerationResult{
		GeneratedAt: time.Now(),
	}

	// Parse manifests
	commands, services, err := g.parser.ParseAll()
	if err != nil {
		return result, fmt.Errorf("failed to parse manifests: %w", err)
	}

	// Generate command reference
	if err := g.generateCommandReference(commands, result); err != nil {
		result.Errors = append(result.Errors, err)
	} else {
		result.CommandsGenerated = true
	}

	// Generate services guide
	if err := g.generateServicesGuide(services, result); err != nil {
		result.Errors = append(result.Errors, err)
	} else {
		result.ServicesGenerated = true
	}

	// Generate workflows guide
	if err := g.generateWorkflowsGuide(commands, result); err != nil {
		result.Errors = append(result.Errors, err)
	}

	// Generate profiles guide
	if err := g.generateProfilesGuide(commands, result); err != nil {
		result.Errors = append(result.Errors, err)
	}

	return result, nil
}

// generateCommandReference generates the enhanced command reference documentation
func (g *EnhancedGenerator) generateCommandReference(commands *CommandsManifest, result *GenerationResult) error {
	content, err := g.renderCommandReferenceTemplate(commands)
	if err != nil {
		return fmt.Errorf("failed to render command reference template: %w", err)
	}

	if err := g.writeDocumentationFile(g.options.ReferenceMDPath, content); err != nil {
		return fmt.Errorf("failed to write command reference: %w", err)
	}

	result.FilesUpdated = append(result.FilesUpdated, g.options.ReferenceMDPath)
	return nil
}

// generateServicesGuide generates the enhanced services guide
func (g *EnhancedGenerator) generateServicesGuide(services *ServicesManifest, result *GenerationResult) error {
	content, err := g.renderServicesGuideTemplate(services)
	if err != nil {
		return fmt.Errorf("failed to render services guide template: %w", err)
	}

	if err := g.writeDocumentationFile(g.options.ServicesMDPath, content); err != nil {
		return fmt.Errorf("failed to write services guide: %w", err)
	}

	result.FilesUpdated = append(result.FilesUpdated, g.options.ServicesMDPath)
	return nil
}

// generateWorkflowsGuide generates the workflows guide
func (g *EnhancedGenerator) generateWorkflowsGuide(commands *CommandsManifest, result *GenerationResult) error {
	workflowsPath := filepath.Join(filepath.Dir(g.options.ReferenceMDPath), "workflows.md")

	content, err := g.renderWorkflowsTemplate(commands)
	if err != nil {
		return fmt.Errorf("failed to render workflows template: %w", err)
	}

	if err := g.writeDocumentationFile(workflowsPath, content); err != nil {
		return fmt.Errorf("failed to write workflows guide: %w", err)
	}

	result.FilesUpdated = append(result.FilesUpdated, workflowsPath)
	return nil
}

// generateProfilesGuide generates the profiles guide
func (g *EnhancedGenerator) generateProfilesGuide(commands *CommandsManifest, result *GenerationResult) error {
	profilesPath := filepath.Join(filepath.Dir(g.options.ReferenceMDPath), "profiles.md")

	content, err := g.renderProfilesTemplate(commands)
	if err != nil {
		return fmt.Errorf("failed to render profiles template: %w", err)
	}

	if err := g.writeDocumentationFile(profilesPath, content); err != nil {
		return fmt.Errorf("failed to write profiles guide: %w", err)
	}

	result.FilesUpdated = append(result.FilesUpdated, profilesPath)
	return nil
}

// renderCommandReferenceTemplate renders the command reference using template
func (g *EnhancedGenerator) renderCommandReferenceTemplate(commands *CommandsManifest) (string, error) {
	tmpl := `---
title: "CLI Reference"
description: "Complete command reference for dev-stack CLI with all available commands and options"
lead: "Comprehensive reference for all dev-stack CLI commands and their usage"
date: {{.Date}}
lastmod: {{.Date}}
draft: false
weight: 50
toc: true
---

<!-- AUTO-GENERATED-START -->
# dev-stack CLI Reference

{{.Description}}

**Version:** {{.Version}}
**Generated:** {{.GeneratedAt}}

## Quick Reference

{{range .Categories}}
### {{.Icon}} {{.Name}}

{{.Description}}

{{range .Commands}}
- **[{{.}}](#{{. | lower}})** - {{index $.CommandDescriptions .}}
{{end}}

{{end}}

## Global Flags

{{range .GlobalFlags}}
- **--{{.Name}}**{{if .Short}}, **-{{.Short}}**{{end}} ({{.Type}}) - {{.Description}}{{if .Default}} (default: {{.Default}}){{end}}
{{end}}

## Commands

{{range .CommandsByCategory}}
{{$category := .Category}}
### {{$category.Icon}} {{$category.Name}}

{{$category.Description}}

{{range .Commands}}
#### {{.Name}}

**Usage:** ` + "`" + `{{.Usage}}` + "`" + `

{{.Description}}

{{if .LongDescription}}
{{.LongDescription}}
{{end}}

{{if .Aliases}}
**Aliases:** {{join .Aliases ", "}}
{{end}}

{{if .Flags}}
**Flags:**
{{range .Flags}}
- **--{{.Name}}**{{if .Short}}, **-{{.Short}}**{{end}} ({{.Type}}) - {{.Description}}{{if .Default}} (default: {{.Default}}){{end}}{{if .Required}} **[required]**{{end}}
{{end}}
{{end}}

{{if .Examples}}
**Examples:**
{{range .Examples}}
` + "```" + `bash
{{.Command}}
` + "```" + `
{{.Description}}

{{end}}
{{end}}

{{if .Tips}}
**Tips:**
{{range .Tips}}
- {{.}}
{{end}}
{{end}}

{{if .RelatedCommands}}
**See also:** {{join .RelatedCommands ", "}}
{{end}}

---

{{end}}
{{end}}

## Help and Support

{{if .Help.common_tasks}}
{{.Help.common_tasks}}
{{end}}

{{if .Help.troubleshooting}}
{{.Help.troubleshooting}}
{{end}}

<!-- AUTO-GENERATED-END -->`

	t, err := template.New("commandReference").Funcs(template.FuncMap{
		"join":  strings.Join,
		"lower": strings.ToLower,
	}).Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	data := g.prepareCommandReferenceData(commands)

	var content strings.Builder
	if err := t.Execute(&content, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return content.String(), nil
}

// renderServicesGuideTemplate renders the services guide using template
func (g *EnhancedGenerator) renderServicesGuideTemplate(services *ServicesManifest) (string, error) {
	tmpl := `---
title: "Services"
description: "Available services and configuration options for dev-stack"
lead: "Explore all the services you can use with dev-stack and how to configure them"
date: {{.Date}}
lastmod: {{.Date}}
draft: false
weight: 30
toc: true
---

<!-- AUTO-GENERATED-START -->
# Available Services

{{.TotalServices}} services available for your development stack.

## Service Categories

{{range .Categories}}
### {{.Name}}
{{range .Services}}
- **[{{.Name}}](#{{.Name | lower}})** - {{.Description}}
{{end}}

{{end}}

## Service Reference

{{range .Services}}
### {{.Name}}

{{.Description}}

{{if .Category}}
**Category:** {{.Category}}
{{end}}

{{if .DefaultPort}}
**Default Port:** {{.DefaultPort}}
{{end}}

{{if .Dependencies}}
**Dependencies:** {{join .Dependencies ", "}}
{{end}}

{{if .Tags}}
**Tags:** {{join .Tags ", "}}
{{end}}

{{if .Options}}
**Configuration Options:**
{{range .Options}}
- {{.}}
{{end}}
{{end}}

{{if .Examples}}
**Examples:**
{{range .Examples}}
` + "```" + `bash
{{.}}
` + "```" + `
{{end}}
{{end}}

{{if .UsageNotes}}
**Usage Notes:**

{{.UsageNotes}}
{{end}}

{{if .Links}}
**Links:**
{{range .Links}}
- [Documentation]({{.}})
{{end}}
{{end}}

---

{{end}}

<!-- AUTO-GENERATED-END -->`

	t, err := template.New("servicesGuide").Funcs(template.FuncMap{
		"join":  strings.Join,
		"lower": strings.ToLower,
	}).Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	data := g.prepareServicesGuideData(services)

	var content strings.Builder
	if err := t.Execute(&content, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return content.String(), nil
}

// renderWorkflowsTemplate renders the workflows guide
func (g *EnhancedGenerator) renderWorkflowsTemplate(commands *CommandsManifest) (string, error) {
	tmpl := `---
title: "Workflows"
description: "Common workflows and task sequences for dev-stack"
lead: "Step-by-step guides for common development tasks"
date: {{.Date}}
lastmod: {{.Date}}
draft: false
weight: 40
toc: true
---

<!-- AUTO-GENERATED-START -->
# Common Workflows

These workflows guide you through common development tasks using dev-stack.

{{range .Workflows}}
## {{.Name}}

{{.Description}}

**Steps:**
{{range $index, $step := .Steps}}
{{add $index 1}}. **{{.Description}}**
   ` + "```" + `bash
   {{.Command}}
   ` + "```" + `
   {{if .Optional}}*This step is optional.*{{end}}

{{end}}

---

{{end}}

## Creating Custom Workflows

You can create your own workflows by combining dev-stack commands. Use the ` + "`" + `dev-stack workflow` + "`" + ` command to see available workflows and execute them interactively.

<!-- AUTO-GENERATED-END -->`

	t, err := template.New("workflows").Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b },
	}).Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	data := struct {
		Date      string
		Workflows []WorkflowInfo
	}{
		Date:      time.Now().Format("2006-01-02T15:04:05Z07:00"),
		Workflows: g.prepareWorkflowsData(commands),
	}

	var content strings.Builder
	if err := t.Execute(&content, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return content.String(), nil
}

// renderProfilesTemplate renders the profiles guide
func (g *EnhancedGenerator) renderProfilesTemplate(commands *CommandsManifest) (string, error) {
	tmpl := `---
title: "Service Profiles"
description: "Predefined service combinations for common development scenarios"
lead: "Quickly start with predefined service combinations"
date: {{.Date}}
lastmod: {{.Date}}
draft: false
weight: 35
toc: true
---

<!-- AUTO-GENERATED-START -->
# Service Profiles

Service profiles are predefined combinations of services for common development scenarios. Use them to quickly start your development environment with the right services.

## Using Profiles

` + "```" + `bash
# Start services using a profile
dev-stack up --profile <profile-name>

# List available profiles
dev-stack up --profile <TAB>
` + "```" + `

## Available Profiles

{{range .Profiles}}
### {{.Name}}

{{.Description}}

**Services included:**
{{range .Services}}
- {{.}}
{{end}}

**Quick start:**
` + "```" + `bash
dev-stack up --profile {{.Name | lower}}
` + "```" + `

---

{{end}}

## Creating Custom Profiles

You can define custom profiles in your ` + "`" + `dev-stack-config.yaml` + "`" + ` file:

` + "```" + `yaml
profiles:
  my-profile:
    name: "My Custom Profile"
    description: "Custom services for my project"
    services:
      - postgres
      - redis
      - my-service
` + "```" + `

<!-- AUTO-GENERATED-END -->`

	t, err := template.New("profiles").Funcs(template.FuncMap{
		"lower": strings.ToLower,
	}).Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	data := struct {
		Date     string
		Profiles []ProfileInfo
	}{
		Date:     time.Now().Format("2006-01-02T15:04:05Z07:00"),
		Profiles: g.prepareProfilesData(commands),
	}

	var content strings.Builder
	if err := t.Execute(&content, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return content.String(), nil
}

// Data preparation methods

func (g *EnhancedGenerator) prepareCommandReferenceData(commands *CommandsManifest) map[string]interface{} {
	// Prepare command descriptions map
	commandDescriptions := make(map[string]string)
	for name, cmd := range commands.Commands {
		commandDescriptions[name] = cmd.Description
	}

	// Prepare global flags
	var globalFlags []map[string]interface{}
	for name, flag := range commands.Global.Flags {
		globalFlags = append(globalFlags, map[string]interface{}{
			"Name":        name,
			"Short":       flag.Short,
			"Type":        flag.Type,
			"Description": flag.Description,
			"Default":     flag.Default,
		})
	}

	// Sort global flags
	sort.Slice(globalFlags, func(i, j int) bool {
		return globalFlags[i]["Name"].(string) < globalFlags[j]["Name"].(string)
	})

	// Prepare commands by category
	var commandsByCategory []map[string]interface{}

	// Sort categories
	var categoryNames []string
	for name := range commands.Categories {
		categoryNames = append(categoryNames, name)
	}
	sort.Strings(categoryNames)

	for _, catName := range categoryNames {
		category := commands.Categories[catName]

		var categoryCommands []map[string]interface{}
		for _, cmdName := range category.Commands {
			if cmd, exists := commands.Commands[cmdName]; !exists || cmd.Hidden {
				continue
			} else {
				// Prepare flags
				var flags []map[string]interface{}
				for flagName, flag := range cmd.Flags {
					if flag.Hidden {
						continue
					}
					flags = append(flags, map[string]interface{}{
						"Name":        flagName,
						"Short":       flag.Short,
						"Type":        flag.Type,
						"Description": flag.Description,
						"Default":     flag.Default,
						"Required":    flag.Required,
					})
				}

				// Sort flags
				sort.Slice(flags, func(i, j int) bool {
					return flags[i]["Name"].(string) < flags[j]["Name"].(string)
				})

				categoryCommands = append(categoryCommands, map[string]interface{}{
					"Name":            cmdName,
					"Description":     cmd.Description,
					"LongDescription": cmd.LongDescription,
					"Usage":           cmd.Usage,
					"Aliases":         cmd.Aliases,
					"Examples":        cmd.Examples,
					"Flags":           flags,
					"RelatedCommands": cmd.RelatedCommands,
					"Tips":            cmd.Tips,
				})
			}
		}

		commandsByCategory = append(commandsByCategory, map[string]interface{}{
			"Category": category,
			"Commands": categoryCommands,
		})
	}

	return map[string]interface{}{
		"Date":                time.Now().Format("2006-01-02T15:04:05Z07:00"),
		"Description":         commands.Metadata.Description,
		"Version":             commands.Metadata.CLIVersion,
		"GeneratedAt":         commands.Metadata.GeneratedAt.Format("2006-01-02 15:04:05"),
		"Categories":          commands.Categories,
		"CommandDescriptions": commandDescriptions,
		"GlobalFlags":         globalFlags,
		"CommandsByCategory":  commandsByCategory,
		"Help":                commands.Help,
	}
}

func (g *EnhancedGenerator) prepareServicesGuideData(services *ServicesManifest) map[string]interface{} {
	// Group services by category
	categories := make(map[string][]map[string]interface{})
	var allServices []map[string]interface{}

	for name, service := range *services {
		serviceData := map[string]interface{}{
			"Name":         name,
			"Description":  service.Description,
			"Category":     service.Category,
			"DefaultPort":  service.DefaultPort,
			"Dependencies": service.Dependencies,
			"Tags":         service.Tags,
			"Options":      service.Options,
			"Examples":     service.Examples,
			"UsageNotes":   service.UsageNotes,
			"Links":        service.Links,
		}

		allServices = append(allServices, serviceData)

		category := service.Category
		if category == "" {
			category = "General"
		}
		categories[category] = append(categories[category], serviceData)
	}

	// Sort services within categories
	for category := range categories {
		sort.Slice(categories[category], func(i, j int) bool {
			return categories[category][i]["Name"].(string) < categories[category][j]["Name"].(string)
		})
	}

	// Prepare category list
	var categoryList []map[string]interface{}
	var categoryNames []string
	for name := range categories {
		categoryNames = append(categoryNames, name)
	}
	sort.Strings(categoryNames)

	for _, name := range categoryNames {
		categoryList = append(categoryList, map[string]interface{}{
			"Name":     name,
			"Services": categories[name],
		})
	}

	// Sort all services
	sort.Slice(allServices, func(i, j int) bool {
		return allServices[i]["Name"].(string) < allServices[j]["Name"].(string)
	})

	return map[string]interface{}{
		"Date":          time.Now().Format("2006-01-02T15:04:05Z07:00"),
		"TotalServices": len(allServices),
		"Categories":    categoryList,
		"Services":      allServices,
	}
}

func (g *EnhancedGenerator) prepareWorkflowsData(commands *CommandsManifest) []WorkflowInfo {
	var workflows []WorkflowInfo
	for _, workflow := range commands.Workflows {
		workflows = append(workflows, workflow)
	}

	// Sort workflows
	sort.Slice(workflows, func(i, j int) bool {
		return workflows[i].Name < workflows[j].Name
	})

	return workflows
}

func (g *EnhancedGenerator) prepareProfilesData(commands *CommandsManifest) []ProfileInfo {
	var profiles []ProfileInfo
	for _, profile := range commands.Profiles {
		profiles = append(profiles, profile)
	}

	// Sort profiles
	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].Name < profiles[j].Name
	})

	return profiles
}

// writeDocumentationFile writes content to a documentation file
func (g *EnhancedGenerator) writeDocumentationFile(path string, content string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write file
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	return nil
}
