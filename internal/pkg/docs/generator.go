package docs

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
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

	// Sort script names for consistent output
	scripts := make([]string, 0, len(*commands))
	for script := range *commands {
		scripts = append(scripts, script)
	}
	sort.Strings(scripts)

	for _, script := range scripts {
		cmds := (*commands)[script]
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

	// Sort service names for consistent output
	serviceNames := make([]string, 0, len(*services))
	for serviceName := range *services {
		serviceNames = append(serviceNames, serviceName)
	}
	sort.Strings(serviceNames)

	for _, serviceName := range serviceNames {
		info := (*services)[serviceName]
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

	// Only write if content has actually changed
	if newDoc == doc {
		if g.options.Verbose {
			fmt.Printf("Auto-generated section in %s is up to date\n", filePath)
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
	dir := filepath.Dir(filePath)
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

	// Sync README.md to docs/_index.md (part of docs generation, not Hugo sync)
	if err := g.syncReadmeToDocsIndex(result); err != nil {
		result.Errors = append(result.Errors, err)
	}

	// Sync CONTRIBUTING.md to docs-site/content/contributing.md (part of docs generation, not Hugo sync)
	if err := g.syncContributingToDocsContributing(result); err != nil {
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

// SyncToHugo synchronizes docs folder content to Hugo content directory
func (g *Generator) SyncToHugo() (*HugoSyncResult, error) {
	result := &HugoSyncResult{
		FilesCopied:  make([]string, 0),
		FilesUpdated: make([]string, 0),
		FilesSkipped: make([]string, 0),
		Errors:       make([]error, 0),
		SyncedAt:     time.Now(),
	}

	if !g.options.EnableHugoSync {
		if g.options.Verbose {
			fmt.Println("Hugo sync is disabled")
		}
		return result, nil
	}

	// Ensure Hugo content directory exists
	if err := os.MkdirAll(g.options.HugoContentDir, 0755); err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("failed to create Hugo content directory: %w", err))
		return result, err
	}

	// Walk through docs source directory
	err := filepath.Walk(g.options.DocsSourceDir, func(srcPath string, info os.FileInfo, err error) error {
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("error walking path %s: %w", srcPath, err))
			return nil // Continue walking
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only sync markdown files
		if filepath.Ext(srcPath) != ".md" {
			result.FilesSkipped = append(result.FilesSkipped, srcPath)
			return nil
		}

		// Calculate relative path from docs source
		relPath, err := filepath.Rel(g.options.DocsSourceDir, srcPath)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to get relative path for %s: %w", srcPath, err))
			return nil
		}

		// Destination path in Hugo content directory
		destPath := filepath.Join(g.options.HugoContentDir, relPath)

		// Ensure destination directory exists
		destDir := filepath.Dir(destPath)
		if err := os.MkdirAll(destDir, 0755); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to create directory %s: %w", destDir, err))
			return nil
		}

		// Check if file needs updating
		needsUpdate, err := g.needsUpdate(srcPath, destPath)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to check if %s needs update: %w", destPath, err))
			return nil
		}

		if !needsUpdate {
			result.FilesSkipped = append(result.FilesSkipped, relPath)
			if g.options.Verbose {
				fmt.Printf("Skipping %s (up to date)\n", relPath)
			}
			return nil
		}

		// Copy/update file
		if g.options.DryRun {
			if g.options.Verbose {
				fmt.Printf("Would sync %s -> %s\n", srcPath, destPath)
			}
			result.FilesUpdated = append(result.FilesUpdated, relPath)
			return nil
		}

		if err := g.copyFile(srcPath, destPath); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to copy %s to %s: %w", srcPath, destPath, err))
			return nil
		}

		// Check if this was a new file or update
		if _, err := os.Stat(destPath); err == nil {
			result.FilesUpdated = append(result.FilesUpdated, relPath)
		} else {
			result.FilesCopied = append(result.FilesCopied, relPath)
		}

		if g.options.Verbose {
			fmt.Printf("Synced %s -> %s\n", srcPath, destPath)
		}

		return nil
	})

	if err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("failed to walk docs directory: %w", err))
	}

	return result, nil
}

// needsUpdate checks if source file is newer than destination file
func (g *Generator) needsUpdate(srcPath, destPath string) (bool, error) {
	destInfo, err := os.Stat(destPath)
	if os.IsNotExist(err) {
		return true, nil // Destination doesn't exist, needs update
	}
	if err != nil {
		return false, err
	}

	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return false, err
	}

	// Compare modification times
	return srcInfo.ModTime().After(destInfo.ModTime()), nil
}

// copyFile copies a file from src to dst
func (g *Generator) copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := srcFile.Close(); err != nil {
			log.Printf("Error closing source file: %v", err)
		}
	}()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		if err := dstFile.Close(); err != nil {
			log.Printf("Error closing destination file: %v", err)
		}
	}()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	// Copy file permissions
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, srcInfo.Mode())
}

// syncReadmeToDocsIndex syncs README.md to docs/_index.md if README is newer
func (g *Generator) syncReadmeToDocsIndex(result *GenerationResult) error {
	readmePath := "README.md"
	indexPath := filepath.Join(g.options.DocsSourceDir, "_index.md")

	// Check if README is newer than _index.md
	needsUpdate, err := g.needsUpdate(readmePath, indexPath)
	if err != nil {
		return fmt.Errorf("failed to check if README needs sync: %w", err)
	}
	if !needsUpdate {
		if g.options.Verbose {
			fmt.Printf("Skipping README sync (up to date)\n")
		}
		return nil
	}

	if g.options.DryRun {
		result.FilesUpdated = append(result.FilesUpdated, indexPath)
		result.ReadmeSynced = true
		if g.options.Verbose {
			fmt.Printf("Would sync README.md -> %s\n", indexPath)
		}
		return nil
	}

	// Transform README to index format
	if err := g.transformReadmeToIndex(readmePath, indexPath); err != nil {
		return fmt.Errorf("failed to sync README: %w", err)
	}

	result.FilesUpdated = append(result.FilesUpdated, indexPath)
	result.ReadmeSynced = true
	if g.options.Verbose {
		fmt.Printf("Synced README.md -> %s\n", indexPath)
	}

	return nil
}

// transformReadmeToIndex converts README.md to docs/_index.md with Hugo frontmatter
func (g *Generator) transformReadmeToIndex(readmePath, indexPath string) error {
	// Read README content
	content, err := os.ReadFile(readmePath)
	if err != nil {
		return fmt.Errorf("failed to read README.md: %w", err)
	}

	// Add Hugo frontmatter
	frontmatter := `---
title: "dev-stack"
description: "A powerful development stack management tool built in Go for streamlined local development automation"
lead: "Streamline your local development with powerful CLI tools and automated service management"
date: 2024-01-01T00:00:00+00:00
lastmod: 2024-01-01T00:00:00+00:00
draft: false
weight: 50
toc: true
---

`

	// Convert markdown links to Hugo relref format
	processed := string(content)
	processed = strings.ReplaceAll(processed, "](docs/setup.md)", "]({{< relref \"setup\" >}})")
	processed = strings.ReplaceAll(processed, "](docs/usage.md)", "]({{< relref \"usage\" >}})")
	processed = strings.ReplaceAll(processed, "](docs/services.md)", "]({{< relref \"services\" >}})")
	processed = strings.ReplaceAll(processed, "](docs/configuration.md)", "]({{< relref \"configuration\" >}})")
	processed = strings.ReplaceAll(processed, "](docs/reference.md)", "]({{< relref \"reference\" >}})")
	processed = strings.ReplaceAll(processed, "](CONTRIBUTING.md)", "]({{< relref \"contributing\" >}})")
	processed = strings.ReplaceAll(processed, "](docs/)", "]({{< relref \"/\" >}})")

	// Combine frontmatter with processed content
	result := frontmatter + processed

	// Ensure directory exists
	dir := filepath.Dir(indexPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write the result
	return os.WriteFile(indexPath, []byte(result), 0644)
}

// syncContributingToDocsContributing syncs CONTRIBUTING.md to docs-site/content/contributing.md if CONTRIBUTING is newer
func (g *Generator) syncContributingToDocsContributing(result *GenerationResult) error {
	contributingPath := "CONTRIBUTING.md"
	docsContributingPath := filepath.Join(g.options.DocsSourceDir, "contributing.md")

	// Check if CONTRIBUTING is newer than docs/contributing.md
	needsUpdate, err := g.needsUpdate(contributingPath, docsContributingPath)
	if err != nil {
		return fmt.Errorf("failed to check if CONTRIBUTING needs sync: %w", err)
	}
	if !needsUpdate {
		if g.options.Verbose {
			fmt.Printf("Skipping CONTRIBUTING sync (up to date)\n")
		}
		return nil
	}

	if g.options.DryRun {
		result.FilesUpdated = append(result.FilesUpdated, docsContributingPath)
		result.ContributingSynced = true
		if g.options.Verbose {
			fmt.Printf("Would sync CONTRIBUTING.md -> %s\n", docsContributingPath)
		}
		return nil
	}

	// Transform CONTRIBUTING to docs contributing format
	if err := g.transformContributingToDocsContributing(contributingPath, docsContributingPath); err != nil {
		return fmt.Errorf("failed to sync CONTRIBUTING: %w", err)
	}

	result.FilesUpdated = append(result.FilesUpdated, docsContributingPath)
	result.ContributingSynced = true
	if g.options.Verbose {
		fmt.Printf("Synced CONTRIBUTING.md -> %s\n", docsContributingPath)
	}

	return nil
}

// transformContributingToDocsContributing converts CONTRIBUTING.md to docs-site/content/contributing.md with Hugo frontmatter
func (g *Generator) transformContributingToDocsContributing(contributingPath, docsContributingPath string) error {
	// Read CONTRIBUTING content
	content, err := os.ReadFile(contributingPath)
	if err != nil {
		return fmt.Errorf("failed to read CONTRIBUTING.md: %w", err)
	}

	// Add Hugo frontmatter
	frontmatter := `---
title: "Contributing"
description: "How to contribute to dev-stack development and documentation"
lead: "Join the dev-stack community and help make it better for everyone"
date: 2024-01-01T00:00:00+00:00
lastmod: 2024-01-01T00:00:00+00:00
draft: false
weight: 60
toc: true
---

`

	// Convert markdown links to Hugo relref format
	processed := string(content)
	processed = strings.ReplaceAll(processed, "](docs-site/content/setup.md)", "]({{< relref \"setup\" >}})")
	processed = strings.ReplaceAll(processed, "](docs-site/content/usage.md)", "]({{< relref \"usage\" >}})")
	processed = strings.ReplaceAll(processed, "](docs-site/content/services.md)", "]({{< relref \"services\" >}})")
	processed = strings.ReplaceAll(processed, "](docs-site/content/configuration.md)", "]({{< relref \"configuration\" >}})")
	processed = strings.ReplaceAll(processed, "](docs-site/content/reference.md)", "]({{< relref \"reference\" >}})")
	processed = strings.ReplaceAll(processed, "](docs-site/content/troubleshooting.md)", "]({{< relref \"troubleshooting\" >}})")
	processed = strings.ReplaceAll(processed, "](docs-site/)", "]({{< relref \"/\" >}})")

	// Combine frontmatter with processed content
	result := frontmatter + processed

	// Ensure directory exists
	dir := filepath.Dir(docsContributingPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write the result
	return os.WriteFile(docsContributingPath, []byte(result), 0644)
}

// GenerateAllWithHugo generates all documentation and optionally syncs to Hugo
func (g *Generator) GenerateAllWithHugo() (*GenerationResult, *HugoSyncResult, error) {
	// Generate documentation first
	docResult, err := g.GenerateAll()
	if err != nil {
		return docResult, nil, err
	}

	// Sync to Hugo if enabled
	var hugoResult *HugoSyncResult
	if g.options.EnableHugoSync {
		hugoResult, err = g.SyncToHugo()
		if err != nil {
			return docResult, hugoResult, fmt.Errorf("hugo sync failed: %w", err)
		}
	}

	return docResult, hugoResult, nil
}
