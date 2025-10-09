package init

import (
	"fmt"
	"os"

	"github.com/isaacgarza/dev-stack/internal/pkg/constants"
	"github.com/isaacgarza/dev-stack/internal/pkg/ui"
)

// createDirectoryStructure creates the necessary directory structure
func (h *InitHandler) createDirectoryStructure() error {
	directories := []string{
		constants.DevStackDir,
	}

	for _, dir := range directories {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// createConfigFile creates the main configuration file
func (h *InitHandler) createConfigFile(projectName, environment string, services []string, validation, advanced map[string]bool) error {
	configContent, err := h.generateConfig(projectName, environment, services, validation, advanced)
	if err != nil {
		return fmt.Errorf("failed to generate config: %w", err)
	}

	configPath := constants.DevStackDir + "/" + constants.ConfigFileName
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	ui.Success("Created %s", configPath)
	return nil
}

// createGitignoreEntries adds dev-stack entries to .gitignore
func (h *InitHandler) createGitignoreEntries() error {
	gitignorePath := constants.GitignoreFileName

	// Check if .gitignore exists
	var existingContent []byte
	if content, err := os.ReadFile(gitignorePath); err == nil {
		existingContent = content
	}

	// Check if entries already exist
	existingStr := string(existingContent)
	hasDevStackEntries := false
	for _, entry := range constants.GitignoreEntries {
		if entry != "" && contains(existingStr, entry) {
			hasDevStackEntries = true
			break
		}
	}

	if hasDevStackEntries {
		ui.Info(".gitignore already contains dev-stack entries")
		return nil
	}

	// Append entries
	file, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open .gitignore: %w", err)
	}
	defer func() { _ = file.Close() }()

	for _, entry := range constants.GitignoreEntries {
		if _, err := file.WriteString(entry + "\n"); err != nil {
			return fmt.Errorf("failed to write to .gitignore: %w", err)
		}
	}

	ui.Success("Updated .gitignore with dev-stack entries")
	return nil
}

// createReadme creates a README file for the dev-stack
func (h *InitHandler) createReadme(projectName string, services []string) error {
	readmeContent := fmt.Sprintf(`# %s Dev Stack

This directory contains the development stack configuration for %s.

## Services

The following services are configured:

%s

## Quick Start

1. Start the stack:
   `+"```bash"+`
   dev-stack up
   `+"```"+`

2. Check status:
   `+"```bash"+`
   dev-stack status
   `+"```"+`

3. Stop the stack:
   `+"```bash"+`
   dev-stack down
   `+"```"+`

## Configuration

- Main config: `+"`%s`"+`
- Docker Compose: `+"`%s`"+`
- Environment: `+"`.env.generated`"+`

## Commands

Run `+"`dev-stack --help`"+` for a full list of available commands.
`, projectName, projectName, formatServicesList(services),
		constants.ConfigFileName, constants.DockerComposeFileName)

	readmePath := constants.DevStackDir + "/" + constants.ReadmeFileName
	if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
		return fmt.Errorf("failed to create README: %w", err)
	}

	ui.Success("Created %s", readmePath)
	return nil
}

// formatServicesList formats the services list for README
func formatServicesList(services []string) string {
	if len(services) == 0 {
		return "- No services configured"
	}

	result := ""
	for _, service := range services {
		result += fmt.Sprintf("- %s\n", service)
	}
	return result
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				containsSubstring(s, substr))))
}

// containsSubstring checks if string contains substring anywhere
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
