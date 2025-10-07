package version

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// DefaultVersionManager implements the VersionManager interface
type DefaultVersionManager struct {
	installDir string
	configDir  string
	detector   *VersionDetector
	installer  VersionInstaller
}

// NewDefaultVersionManager creates a new default version manager
func NewDefaultVersionManager(installDir, configDir string) *DefaultVersionManager {
	installer := NewGitHubVersionInstaller("isaacgarza", "dev-stack", installDir)

	return &DefaultVersionManager{
		installDir: installDir,
		configDir:  configDir,
		detector:   NewVersionDetector(),
		installer:  installer,
	}
}

// DetectProjectVersion detects the required version for a project
func (m *DefaultVersionManager) DetectProjectVersion(projectPath string) (*VersionConstraint, error) {
	return m.detector.DetectProjectVersion(projectPath)
}

// ParseVersionFile parses a version file at the given path
func (m *DefaultVersionManager) ParseVersionFile(path string) (*VersionFile, error) {
	return m.detector.parseVersionFile(path)
}

// ParseVersionConstraint parses a version constraint string
func (m *DefaultVersionManager) ParseVersionConstraint(constraint string) (*VersionConstraint, error) {
	return ParseVersionConstraint(constraint)
}

// ListAvailableVersions lists all available versions from GitHub
func (m *DefaultVersionManager) ListAvailableVersions() ([]Version, error) {
	if githubInstaller, ok := m.installer.(*GitHubVersionInstaller); ok {
		return githubInstaller.ListAvailableVersions()
	}
	return nil, NewVersionError(ErrVersionInstall, "unsupported installer type", nil)
}

// InstallVersion installs a specific version
func (m *DefaultVersionManager) InstallVersion(version Version) error {
	// Check if already installed
	installed, err := m.ListInstalledVersions()
	if err != nil {
		return err
	}

	for _, v := range installed {
		if v.Version.Compare(version) == 0 {
			return nil // Already installed
		}
	}

	// Download the version
	downloadPath, err := m.installer.Download(version)
	if err != nil {
		return err
	}

	// Verify checksum if available
	if githubInstaller, ok := m.installer.(*GitHubVersionInstaller); ok {
		checksum, err := githubInstaller.GetChecksum(version)
		if err == nil && checksum != "" {
			if err := m.installer.Verify(downloadPath, checksum); err != nil {
				return err
			}
		}
	}

	// Extract/install the binary
	versionDir := filepath.Join(m.installDir, "versions", version.String())
	binaryPath := filepath.Join(versionDir, "dev-stack")
	if err := m.installer.Install(downloadPath, binaryPath); err != nil {
		return err
	}

	// Update installed versions registry
	if err := m.registerInstalledVersion(version, binaryPath); err != nil {
		return err
	}

	return nil
}

// UninstallVersion removes a specific version
func (m *DefaultVersionManager) UninstallVersion(version Version) error {
	versionDir := filepath.Join(m.installDir, "versions", version.String())

	if err := os.RemoveAll(versionDir); err != nil {
		return NewVersionError(ErrVersionInstall,
			fmt.Sprintf("failed to remove version directory: %s", versionDir), err)
	}

	// Update installed versions registry
	if err := m.unregisterInstalledVersion(version); err != nil {
		return err
	}

	return nil
}

// VerifyVersion verifies an installed version
func (m *DefaultVersionManager) VerifyVersion(version Version) error {
	installed, err := m.ListInstalledVersions()
	if err != nil {
		return err
	}

	for _, v := range installed {
		if v.Version.Compare(version) == 0 {
			// Check if binary exists and is executable
			if _, err := os.Stat(v.Path); err != nil {
				return NewVersionError(ErrVersionNotFound,
					fmt.Sprintf("version %s binary not found at: %s", version.String(), v.Path), err)
			}
			return nil
		}
	}

	return NewVersionError(ErrVersionNotFound,
		fmt.Sprintf("version %s is not installed", version.String()), nil)
}

// ListInstalledVersions lists all installed versions
func (m *DefaultVersionManager) ListInstalledVersions() ([]InstalledVersion, error) {
	registryPath := filepath.Join(m.configDir, "installed_versions.json")

	data, err := os.ReadFile(registryPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []InstalledVersion{}, nil
		}
		return nil, err
	}

	var versions []InstalledVersion
	if err := json.Unmarshal(data, &versions); err != nil {
		return nil, err
	}

	// Filter out versions that no longer exist on disk
	var validVersions []InstalledVersion
	for _, v := range versions {
		if _, err := os.Stat(v.Path); err == nil {
			validVersions = append(validVersions, v)
		}
	}

	// Update registry if we filtered any versions
	if len(validVersions) != len(versions) {
		if err := m.saveInstalledVersions(validVersions); err != nil {
			return nil, err
		}
	}

	return validVersions, nil
}

// GetActiveVersion returns the currently active version
func (m *DefaultVersionManager) GetActiveVersion() (*InstalledVersion, error) {
	installed, err := m.ListInstalledVersions()
	if err != nil {
		return nil, err
	}

	for _, v := range installed {
		if v.Active {
			return &v, nil
		}
	}

	return nil, NewVersionError(ErrVersionNotFound, "no active version set", nil)
}

// SetActiveVersion sets the active version
func (m *DefaultVersionManager) SetActiveVersion(version Version) error {
	installed, err := m.ListInstalledVersions()
	if err != nil {
		return err
	}

	var found bool
	for i := range installed {
		if installed[i].Version.Compare(version) == 0 {
			installed[i].Active = true
			found = true
		} else {
			installed[i].Active = false
		}
	}

	if !found {
		return NewVersionError(ErrVersionNotFound,
			fmt.Sprintf("version %s is not installed", version.String()), nil)
	}

	return m.saveInstalledVersions(installed)
}

// ResolveVersion finds the best installed version that matches a constraint
func (m *DefaultVersionManager) ResolveVersion(constraint VersionConstraint) (*InstalledVersion, error) {
	installed, err := m.ListInstalledVersions()
	if err != nil {
		return nil, err
	}

	var candidates []InstalledVersion
	for _, v := range installed {
		if constraint.Satisfies(v.Version) {
			candidates = append(candidates, v)
		}
	}

	if len(candidates) == 0 {
		return nil, NewVersionError(ErrVersionNotFound,
			fmt.Sprintf("no installed version satisfies constraint: %s", constraint.Original), nil)
	}

	// Sort candidates and return the latest
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Version.Compare(candidates[j].Version) > 0
	})

	return &candidates[0], nil
}

// SwitchToVersion switches to a specific version
func (m *DefaultVersionManager) SwitchToVersion(version Version) error {
	return m.SetActiveVersion(version)
}

// GetProjectConfig gets the version configuration for a project
func (m *DefaultVersionManager) GetProjectConfig(projectPath string) (*ProjectVersionConfig, error) {
	configsPath := filepath.Join(m.configDir, "project_configs.json")

	data, err := os.ReadFile(configsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, NewVersionError(ErrProjectConfig, "no project configuration found", nil)
		}
		return nil, err
	}

	var configs []ProjectVersionConfig
	if err := json.Unmarshal(data, &configs); err != nil {
		return nil, err
	}

	// Normalize project path
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		absPath = projectPath
	}

	for _, config := range configs {
		if config.ProjectPath == absPath {
			return &config, nil
		}
	}

	return nil, NewVersionError(ErrProjectConfig, "project configuration not found", nil)
}

// SetProjectConfig sets the version configuration for a project
func (m *DefaultVersionManager) SetProjectConfig(config ProjectVersionConfig) error {
	configsPath := filepath.Join(m.configDir, "project_configs.json")

	// Normalize project path
	absPath, err := filepath.Abs(config.ProjectPath)
	if err != nil {
		absPath = config.ProjectPath
	}
	config.ProjectPath = absPath
	config.LastUsed = time.Now()

	var configs []ProjectVersionConfig
	data, err := os.ReadFile(configsPath)
	if err == nil {
		_ = json.Unmarshal(data, &configs)
	}

	// Update or add configuration
	found := false
	for i := range configs {
		if configs[i].ProjectPath == config.ProjectPath {
			configs[i] = config
			found = true
			break
		}
	}

	if !found {
		configs = append(configs, config)
	}

	// Save updated configurations
	if err := os.MkdirAll(m.configDir, 0755); err != nil {
		return err
	}

	data, err = json.MarshalIndent(configs, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configsPath, data, 0644)
}

// ListProjectConfigs lists all project configurations
func (m *DefaultVersionManager) ListProjectConfigs() ([]ProjectVersionConfig, error) {
	configsPath := filepath.Join(m.configDir, "project_configs.json")

	data, err := os.ReadFile(configsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []ProjectVersionConfig{}, nil
		}
		return nil, err
	}

	var configs []ProjectVersionConfig
	if err := json.Unmarshal(data, &configs); err != nil {
		return nil, err
	}

	return configs, nil
}

// CleanupOldVersions removes old versions, keeping only the specified number
func (m *DefaultVersionManager) CleanupOldVersions(keepCount int) error {
	installed, err := m.ListInstalledVersions()
	if err != nil {
		return err
	}

	if len(installed) <= keepCount {
		return nil // Nothing to cleanup
	}

	// Sort by install date (oldest first)
	sort.Slice(installed, func(i, j int) bool {
		return installed[i].InstallDate.Before(installed[j].InstallDate)
	})

	// Keep the latest versions and the active version
	var toKeep []InstalledVersion
	var toRemove []InstalledVersion

	// Always keep the active version
	for _, v := range installed {
		if v.Active {
			toKeep = append(toKeep, v)
		}
	}

	// Keep the latest non-active versions
	nonActiveCount := 0
	for i := len(installed) - 1; i >= 0; i-- {
		v := installed[i]
		if !v.Active {
			if nonActiveCount < keepCount-len(toKeep) {
				toKeep = append(toKeep, v)
				nonActiveCount++
			} else {
				toRemove = append(toRemove, v)
			}
		}
	}

	// Remove old versions
	for _, v := range toRemove {
		if err := m.UninstallVersion(v.Version); err != nil {
			return err
		}
	}

	return nil
}

// GarbageCollect performs garbage collection of unused files
func (m *DefaultVersionManager) GarbageCollect() error {
	// Clean up orphaned version directories
	versionsDir := filepath.Join(m.installDir, "versions")
	entries, err := os.ReadDir(versionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	installed, err := m.ListInstalledVersions()
	if err != nil {
		return err
	}

	// Create map of valid version directories
	validDirs := make(map[string]bool)
	for _, v := range installed {
		validDirs[v.Version.String()] = true
	}

	// Remove orphaned directories
	for _, entry := range entries {
		if entry.IsDir() && !validDirs[entry.Name()] {
			dirPath := filepath.Join(versionsDir, entry.Name())
			if err := os.RemoveAll(dirPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// registerInstalledVersion adds a version to the installed versions registry
func (m *DefaultVersionManager) registerInstalledVersion(version Version, binaryPath string) error {
	installed, err := m.ListInstalledVersions()
	if err != nil {
		return err
	}

	// Check if already registered
	for _, v := range installed {
		if v.Version.Compare(version) == 0 {
			return nil // Already registered
		}
	}

	// Add new version
	newVersion := InstalledVersion{
		Version:     version,
		Path:        binaryPath,
		InstallDate: time.Now(),
		Source:      "github",
		Active:      len(installed) == 0, // First version is active by default
	}

	installed = append(installed, newVersion)
	return m.saveInstalledVersions(installed)
}

// unregisterInstalledVersion removes a version from the registry
func (m *DefaultVersionManager) unregisterInstalledVersion(version Version) error {
	installed, err := m.ListInstalledVersions()
	if err != nil {
		return err
	}

	var filtered []InstalledVersion
	for _, v := range installed {
		if v.Version.Compare(version) != 0 {
			filtered = append(filtered, v)
		}
	}

	return m.saveInstalledVersions(filtered)
}

// saveInstalledVersions saves the installed versions registry
func (m *DefaultVersionManager) saveInstalledVersions(versions []InstalledVersion) error {
	if err := os.MkdirAll(m.configDir, 0755); err != nil {
		return err
	}

	registryPath := filepath.Join(m.configDir, "installed_versions.json")
	data, err := json.MarshalIndent(versions, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(registryPath, data, 0644)
}
