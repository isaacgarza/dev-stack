package handlers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

// DocsHandler handles the docs command
type DocsHandler struct{}

// NewDocsHandler creates a new docs handler
func NewDocsHandler() *DocsHandler {
	return &DocsHandler{}
}

// Handle executes the docs command
func (h *DocsHandler) Handle(ctx context.Context, cmd *cobra.Command, args []string, base *BaseCommand) error {
	servicesOnly, _ := cmd.Flags().GetBool("services-only")
	commandsOnly, _ := cmd.Flags().GetBool("commands-only")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	if commandsOnly {
		return fmt.Errorf("commands documentation generation not implemented yet")
	}

	if !servicesOnly && !commandsOnly {
		// Generate both by default
		servicesOnly = true
	}

	if servicesOnly {
		return h.generateServicesDoc(dryRun)
	}

	return nil
}

// generateServicesDoc generates services documentation from service files
func (h *DocsHandler) generateServicesDoc(dryRun bool) error {
	fmt.Println("📚 Generating services documentation...")

	// Load services by category
	servicesByCategory, err := h.loadServicesByCategory()
	if err != nil {
		return fmt.Errorf("failed to load services: %w", err)
	}

	// Generate markdown content
	content, err := h.generateMarkdown(servicesByCategory)
	if err != nil {
		return fmt.Errorf("failed to generate markdown: %w", err)
	}

	if dryRun {
		fmt.Println("📋 Generated content preview:")
		fmt.Println(content)
		return nil
	}

	// Write to file
	outputFile := "docs-site/content/services.md"
	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		return fmt.Errorf("failed to create docs directory: %w", err)
	}

	if err := os.WriteFile(outputFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write documentation: %w", err)
	}

	fmt.Printf("✅ Generated services documentation: %s\n", outputFile)
	return nil
}

// generateMarkdown generates markdown content from services
func (h *DocsHandler) generateMarkdown(servicesByCategory map[string][]ServiceInfo) (string, error) {
	tmplText := `# Available Services

This documentation is automatically generated from service definitions.

{{range $category, $services := .ServicesByCategory}}
## {{title $category}} Services

{{range $services}}
### {{.Name}}

**Description:** {{.Description}}

{{if .Dependencies}}**Dependencies:** {{join .Dependencies ", "}}{{end}}

{{if .Options}}**Configuration Options:**
{{range .Options}}
- {{.}}
{{end}}
{{end}}

{{if .Examples}}**Examples:**
{{range .Examples}}
- {{.}}
{{end}}
{{end}}

{{if .UsageNotes}}**Usage Notes:** {{.UsageNotes}}{{end}}

{{if .Links}}**Links:**
{{range .Links}}
- {{.}}
{{end}}
{{end}}

---

{{end}}
{{end}}

## Service Categories

{{range $category, $services := .ServicesByCategory}}
- **{{title $category}}**: {{len $services}} service{{if ne (len $services) 1}}s{{end}}
{{end}}

*Documentation generated automatically from service definitions*
`

	tmpl, err := template.New("services").Funcs(template.FuncMap{
		"title": strings.Title,
		"join":  strings.Join,
	}).Parse(tmplText)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	data := struct {
		ServicesByCategory map[string][]ServiceInfo
	}{
		ServicesByCategory: servicesByCategory,
	}

	var result strings.Builder
	if err := tmpl.Execute(&result, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return result.String(), nil
}

// loadServicesByCategory loads services organized by category for documentation
func (h *DocsHandler) loadServicesByCategory() (map[string][]ServiceInfo, error) {
	return NewServiceUtils().LoadServicesByCategory()
}

// ValidateArgs validates the command arguments
func (h *DocsHandler) ValidateArgs(args []string) error {
	return nil
}

// GetRequiredFlags returns required flags for this command
func (h *DocsHandler) GetRequiredFlags() []string {
	return []string{}
}
