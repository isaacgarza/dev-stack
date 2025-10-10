package version

import (
	"sort"
)

// DefaultVersionManager implements the VersionManager interface
type DefaultVersionManager struct {
	installDir string
	configDir  string
	detector   *VersionDetector
	installer  VersionInstaller
	registry   *VersionRegistry
	projectCfg *ProjectConfigManager
}

// NewDefaultVersionManager creates a new default version manager
func NewDefaultVersionManager(installDir, configDir string) *DefaultVersionManager {
	installer := NewGitHubVersionInstaller("isaacgarza", "dev-stack", installDir)
	registry := NewVersionRegistry(configDir)
	projectCfg := NewProjectConfigManager(configDir)

	return &DefaultVersionManager{
		installDir: installDir,
		configDir:  configDir,
		detector:   NewVersionDetector(),
		installer:  installer,
		registry:   registry,
		projectCfg: projectCfg,
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

// ListInstalledVersions lists all installed versions
func (m *DefaultVersionManager) ListInstalledVersions() ([]InstalledVersion, error) {
	return m.registry.ListInstalledVersions()
}

// GetActiveVersion returns the currently active version
func (m *DefaultVersionManager) GetActiveVersion() (*InstalledVersion, error) {
	return m.registry.GetActiveVersion()
}

// SetActiveVersion sets the active version
func (m *DefaultVersionManager) SetActiveVersion(version Version) error {
	return m.registry.SetActiveVersion(version)
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
			"no installed version satisfies constraint: "+constraint.Original, nil)
	}

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
	return m.projectCfg.GetProjectConfig(projectPath)
}

// SetProjectConfig sets the version configuration for a project
func (m *DefaultVersionManager) SetProjectConfig(config ProjectVersionConfig) error {
	return m.projectCfg.SetProjectConfig(config)
}

// ListProjectConfigs lists all project configurations
func (m *DefaultVersionManager) ListProjectConfigs() ([]ProjectVersionConfig, error) {
	return m.projectCfg.ListProjectConfigs()
}
