package version

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// VersionRegistry manages the registry of installed versions
type VersionRegistry struct {
	configDir string
}

// NewVersionRegistry creates a new version registry
func NewVersionRegistry(configDir string) *VersionRegistry {
	return &VersionRegistry{
		configDir: configDir,
	}
}

// ListInstalledVersions lists all installed versions
func (r *VersionRegistry) ListInstalledVersions() ([]InstalledVersion, error) {
	registryPath := filepath.Join(r.configDir, "installed_versions.json")

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

	var validVersions []InstalledVersion
	for _, v := range versions {
		if _, err := os.Stat(v.Path); err == nil {
			validVersions = append(validVersions, v)
		}
	}

	if len(validVersions) != len(versions) {
		if err := r.saveInstalledVersions(validVersions); err != nil {
			return nil, err
		}
	}

	return validVersions, nil
}

// GetActiveVersion returns the currently active version
func (r *VersionRegistry) GetActiveVersion() (*InstalledVersion, error) {
	installed, err := r.ListInstalledVersions()
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
func (r *VersionRegistry) SetActiveVersion(version Version) error {
	installed, err := r.ListInstalledVersions()
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
			"version "+version.String()+" is not installed", nil)
	}

	return r.saveInstalledVersions(installed)
}

// RegisterInstalledVersion adds a version to the installed versions registry
func (r *VersionRegistry) RegisterInstalledVersion(version Version, binaryPath string) error {
	installed, err := r.ListInstalledVersions()
	if err != nil {
		return err
	}

	for _, v := range installed {
		if v.Version.Compare(version) == 0 {
			return nil
		}
	}

	newVersion := InstalledVersion{
		Version:     version,
		Path:        binaryPath,
		InstallDate: time.Now(),
		Source:      "github",
		Active:      len(installed) == 0,
	}

	installed = append(installed, newVersion)
	return r.saveInstalledVersions(installed)
}

// UnregisterInstalledVersion removes a version from the registry
func (r *VersionRegistry) UnregisterInstalledVersion(version Version) error {
	installed, err := r.ListInstalledVersions()
	if err != nil {
		return err
	}

	var filtered []InstalledVersion
	for _, v := range installed {
		if v.Version.Compare(version) != 0 {
			filtered = append(filtered, v)
		}
	}

	return r.saveInstalledVersions(filtered)
}

// saveInstalledVersions saves the installed versions registry
func (r *VersionRegistry) saveInstalledVersions(versions []InstalledVersion) error {
	if err := os.MkdirAll(r.configDir, 0755); err != nil {
		return err
	}

	registryPath := filepath.Join(r.configDir, "installed_versions.json")
	data, err := json.MarshalIndent(versions, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(registryPath, data, 0644)
}
