package version

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// RegistryManager handles the registry of installed versions
type RegistryManager struct {
	configDir string
}

// NewRegistryManager creates a new registry manager
func NewRegistryManager(configDir string) *RegistryManager {
	return &RegistryManager{configDir: configDir}
}

// ListInstalledVersions lists all installed versions
func (rm *RegistryManager) ListInstalledVersions() ([]InstalledVersion, error) {
	registryPath := filepath.Join(rm.configDir, "installed_versions.json")

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
		if err := rm.saveInstalledVersions(validVersions); err != nil {
			return nil, err
		}
	}

	return validVersions, nil
}

// GetActiveVersion returns the currently active version
func (rm *RegistryManager) GetActiveVersion() (*InstalledVersion, error) {
	installed, err := rm.ListInstalledVersions()
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
func (rm *RegistryManager) SetActiveVersion(version Version) error {
	installed, err := rm.ListInstalledVersions()
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

	return rm.saveInstalledVersions(installed)
}

// ResolveVersion finds the best installed version that matches a constraint
func (rm *RegistryManager) ResolveVersion(constraint VersionConstraint) (*InstalledVersion, error) {
	installed, err := rm.ListInstalledVersions()
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

// RegisterInstalledVersion adds a version to the installed versions registry
func (rm *RegistryManager) RegisterInstalledVersion(version Version, binaryPath string) error {
	installed, err := rm.ListInstalledVersions()
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
	return rm.saveInstalledVersions(installed)
}

// UnregisterInstalledVersion removes a version from the registry
func (rm *RegistryManager) UnregisterInstalledVersion(version Version) error {
	installed, err := rm.ListInstalledVersions()
	if err != nil {
		return err
	}

	var filtered []InstalledVersion
	for _, v := range installed {
		if v.Version.Compare(version) != 0 {
			filtered = append(filtered, v)
		}
	}

	return rm.saveInstalledVersions(filtered)
}

// saveInstalledVersions saves the installed versions registry
func (rm *RegistryManager) saveInstalledVersions(versions []InstalledVersion) error {
	if err := os.MkdirAll(rm.configDir, 0755); err != nil {
		return err
	}

	registryPath := filepath.Join(rm.configDir, "installed_versions.json")
	data, err := json.MarshalIndent(versions, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(registryPath, data, 0644)
}
