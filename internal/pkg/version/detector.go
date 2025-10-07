package version

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// VersionDetector handles detection of version requirements from project files
type VersionDetector struct {
	searchPaths []string
	fileNames   []string
}

// NewVersionDetector creates a new version detector
func NewVersionDetector() *VersionDetector {
	return &VersionDetector{
		searchPaths: []string{
			".",
			".dev-stack",
			".config",
			"config",
		},
		fileNames: []string{
			".dev-stack-version",
			".dev-stack-version.yaml",
			".dev-stack-version.yml",
			"dev-stack-version",
			"dev-stack-version.yaml",
			"dev-stack-version.yml",
		},
	}
}

// DetectProjectVersion detects the required version for a project
func (d *VersionDetector) DetectProjectVersion(projectPath string) (*VersionConstraint, error) {
	// First, try to find a version file
	versionFile, err := d.findVersionFile(projectPath)
	if err != nil {
		return nil, err
	}

	if versionFile != nil {
		return d.parseVersionConstraintFromFile(versionFile)
	}

	// If no version file found, try to detect from other configuration files
	constraint, err := d.detectFromConfigFiles(projectPath)
	if err != nil {
		return nil, err
	}

	if constraint != nil {
		return constraint, nil
	}

	// Default to any version if no specific requirement found
	return &VersionConstraint{
		Operator: "*",
		Version:  Version{Major: 0, Minor: 0, Patch: 0},
		Original: "*",
	}, nil
}

// findVersionFile searches for version files in the project
func (d *VersionDetector) findVersionFile(projectPath string) (*VersionFile, error) {
	for _, searchPath := range d.searchPaths {
		fullSearchPath := filepath.Join(projectPath, searchPath)

		// Check if the search path exists
		if _, err := os.Stat(fullSearchPath); os.IsNotExist(err) {
			continue
		}

		for _, fileName := range d.fileNames {
			filePath := filepath.Join(fullSearchPath, fileName)

			if _, err := os.Stat(filePath); err == nil {
				versionFile, err := d.parseVersionFile(filePath)
				if err != nil {
					continue // Try next file
				}
				return versionFile, nil
			}
		}
	}

	return nil, nil // No version file found
}

// parseVersionFile parses a version file
func (d *VersionDetector) parseVersionFile(filePath string) (*VersionFile, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, NewVersionError(ErrProjectConfig,
			fmt.Sprintf("failed to read version file: %s", filePath), err)
	}

	// Try to determine file format based on extension or content
	if d.isYAMLFile(filePath, data) {
		return d.parseYAMLVersionFile(data)
	}

	return d.parseTextVersionFile(data)
}

// isYAMLFile determines if a file is YAML format
func (d *VersionDetector) isYAMLFile(filePath string, data []byte) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext == ".yaml" || ext == ".yml" {
		return true
	}

	// Check if content looks like YAML
	content := strings.TrimSpace(string(data))
	return strings.Contains(content, ":") || strings.Contains(content, "-")
}

// parseYAMLVersionFile parses a YAML version file
func (d *VersionDetector) parseYAMLVersionFile(data []byte) (*VersionFile, error) {
	var versionFile VersionFile
	if err := yaml.Unmarshal(data, &versionFile); err != nil {
		return nil, NewVersionError(ErrProjectConfig,
			"failed to parse YAML version file", err)
	}

	if versionFile.Version == "" {
		return nil, NewVersionError(ErrProjectConfig,
			"version field is required in version file", nil)
	}

	return &versionFile, nil
}

// parseTextVersionFile parses a plain text version file
func (d *VersionDetector) parseTextVersionFile(data []byte) (*VersionFile, error) {
	content := strings.TrimSpace(string(data))
	if content == "" {
		return nil, NewVersionError(ErrProjectConfig,
			"version file is empty", nil)
	}

	// Split by lines and take the first non-empty line as version
	lines := strings.Split(content, "\n")
	var version string
	var constraints []string

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if i == 0 || version == "" {
			version = line
		} else {
			constraints = append(constraints, line)
		}
	}

	if version == "" {
		return nil, NewVersionError(ErrProjectConfig,
			"no version found in version file", nil)
	}

	return &VersionFile{
		Version:     version,
		Constraints: constraints,
	}, nil
}

// parseVersionConstraintFromFile creates a version constraint from a version file
func (d *VersionDetector) parseVersionConstraintFromFile(versionFile *VersionFile) (*VersionConstraint, error) {
	// Primary constraint from version field
	constraint, err := ParseVersionConstraint(versionFile.Version)
	if err != nil {
		return nil, NewVersionError(ErrProjectConfig,
			fmt.Sprintf("invalid version constraint in file: %s", versionFile.Version), err)
	}

	// TODO: Handle additional constraints if needed
	// For now, we'll just use the primary version constraint

	return constraint, nil
}

// detectFromConfigFiles tries to detect version requirements from other config files
func (d *VersionDetector) detectFromConfigFiles(projectPath string) (*VersionConstraint, error) {
	// Check for dev-stack configuration files
	configPaths := []string{
		"dev-stack-config.yaml",
		"dev-stack-config.yml",
		".dev-stack.yaml",
		".dev-stack.yml",
		"dev-stack.yaml",
		"dev-stack.yml",
	}

	for _, configPath := range configPaths {
		fullPath := filepath.Join(projectPath, configPath)
		if _, err := os.Stat(fullPath); err == nil {
			constraint, err := d.parseVersionFromConfig(fullPath)
			if err != nil {
				continue // Try next config file
			}
			if constraint != nil {
				return constraint, nil
			}
		}
	}

	return nil, nil
}

// parseVersionFromConfig extracts version requirements from a config file
func (d *VersionDetector) parseVersionFromConfig(configPath string) (*VersionConstraint, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Look for version-related fields
	versionFields := []string{
		"dev_stack_version",
		"devStackVersion",
		"version",
		"required_version",
		"requiredVersion",
	}

	for _, field := range versionFields {
		if value, exists := config[field]; exists {
			if versionStr, ok := value.(string); ok && versionStr != "" {
				return ParseVersionConstraint(versionStr)
			}
		}
	}

	return nil, nil
}

// FindVersionFiles returns all version files found in a project
func (d *VersionDetector) FindVersionFiles(projectPath string) ([]string, error) {
	var files []string

	for _, searchPath := range d.searchPaths {
		fullSearchPath := filepath.Join(projectPath, searchPath)

		if _, err := os.Stat(fullSearchPath); os.IsNotExist(err) {
			continue
		}

		for _, fileName := range d.fileNames {
			filePath := filepath.Join(fullSearchPath, fileName)

			if _, err := os.Stat(filePath); err == nil {
				files = append(files, filePath)
			}
		}
	}

	return files, nil
}

// CreateVersionFile creates a version file for a project
func (d *VersionDetector) CreateVersionFile(projectPath, version string, format string) error {
	var fileName string
	var content []byte
	var err error

	switch strings.ToLower(format) {
	case "yaml", "yml":
		fileName = ".dev-stack-version.yaml"
		versionFile := VersionFile{
			Version: version,
			Metadata: map[string]string{
				"created_by": "dev-stack",
			},
		}
		content, err = yaml.Marshal(versionFile)
		if err != nil {
			return NewVersionError(ErrProjectConfig,
				"failed to marshal version file", err)
		}
	default:
		fileName = ".dev-stack-version"
		content = []byte(version + "\n")
	}

	filePath := filepath.Join(projectPath, fileName)

	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return NewVersionError(ErrProjectConfig,
			fmt.Sprintf("failed to write version file: %s", filePath), err)
	}

	return nil
}

// ValidateProjectVersion validates the version constraint in a project
func (d *VersionDetector) ValidateProjectVersion(projectPath string) error {
	constraint, err := d.DetectProjectVersion(projectPath)
	if err != nil {
		return err
	}

	if constraint == nil {
		return NewVersionError(ErrProjectConfig,
			"no version constraint found", nil)
	}

	// Validate the constraint format
	return ValidateConstraint(constraint.Original)
}

// UpdateProjectVersion updates the version requirement for a project
func (d *VersionDetector) UpdateProjectVersion(projectPath, newVersion string) error {
	// Find existing version file
	files, err := d.FindVersionFiles(projectPath)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		// Create new version file
		return d.CreateVersionFile(projectPath, newVersion, "text")
	}

	// Update the first found version file
	filePath := files[0]

	// Determine file format
	if d.isYAMLFile(filePath, nil) {
		return d.updateYAMLVersionFile(filePath, newVersion)
	}

	return d.updateTextVersionFile(filePath, newVersion)
}

// updateYAMLVersionFile updates a YAML version file
func (d *VersionDetector) updateYAMLVersionFile(filePath, newVersion string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var versionFile VersionFile
	if err := yaml.Unmarshal(data, &versionFile); err != nil {
		return err
	}

	versionFile.Version = newVersion

	updatedData, err := yaml.Marshal(versionFile)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, updatedData, 0644)
}

// updateTextVersionFile updates a text version file
func (d *VersionDetector) updateTextVersionFile(filePath, newVersion string) error {
	return os.WriteFile(filePath, []byte(newVersion+"\n"), 0644)
}
