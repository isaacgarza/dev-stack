package version

import (
	"fmt"
	"os"
	"path/filepath"
)

// InstallationManager handles version installation and uninstallation
type InstallationManager struct {
	installDir string
	installer  VersionInstaller
	registry   *RegistryManager
}

// NewInstallationManager creates a new installation manager
func NewInstallationManager(installDir string, registry *RegistryManager) *InstallationManager {
	installer := NewGitHubVersionInstaller("isaacgarza", "dev-stack", installDir)
	return &InstallationManager{
		installDir: installDir,
		installer:  installer,
		registry:   registry,
	}
}

// ListAvailableVersions lists all available versions from GitHub
func (im *InstallationManager) ListAvailableVersions() ([]Version, error) {
	if githubInstaller, ok := im.installer.(*GitHubVersionInstaller); ok {
		return githubInstaller.ListAvailableVersions()
	}
	return nil, NewVersionError(ErrVersionInstall, "unsupported installer type", nil)
}

// InstallVersion installs a specific version
func (im *InstallationManager) InstallVersion(version Version) error {
	// Check if already installed
	installed, err := im.registry.ListInstalledVersions()
	if err != nil {
		return err
	}

	for _, v := range installed {
		if v.Version.Compare(version) == 0 {
			return nil // Already installed
		}
	}

	// Download the version
	downloadPath, err := im.installer.Download(version)
	if err != nil {
		return err
	}

	// Verify checksum if available
	if githubInstaller, ok := im.installer.(*GitHubVersionInstaller); ok {
		checksum, err := githubInstaller.GetChecksum(version)
		if err == nil && checksum != "" {
			if err := im.installer.Verify(downloadPath, checksum); err != nil {
				return err
			}
		}
	}

	// Extract/install the binary
	versionDir := filepath.Join(im.installDir, "versions", version.String())
	binaryPath := filepath.Join(versionDir, "dev-stack")
	if err := im.installer.Install(downloadPath, binaryPath); err != nil {
		return err
	}

	// Update installed versions registry
	if err := im.registry.RegisterInstalledVersion(version, binaryPath); err != nil {
		return err
	}

	return nil
}

// UninstallVersion removes a specific version
func (im *InstallationManager) UninstallVersion(version Version) error {
	versionDir := filepath.Join(im.installDir, "versions", version.String())

	if err := os.RemoveAll(versionDir); err != nil {
		return NewVersionError(ErrVersionInstall,
			fmt.Sprintf("failed to remove version directory: %s", versionDir), err)
	}

	// Update installed versions registry
	if err := im.registry.UnregisterInstalledVersion(version); err != nil {
		return err
	}

	return nil
}

// VerifyVersion verifies an installed version
func (im *InstallationManager) VerifyVersion(version Version) error {
	installed, err := im.registry.ListInstalledVersions()
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
