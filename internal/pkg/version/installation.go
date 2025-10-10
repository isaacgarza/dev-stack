package version

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// InstallVersion installs a specific version
func (m *DefaultVersionManager) InstallVersion(version Version) error {
	installed, err := m.ListInstalledVersions()
	if err != nil {
		return err
	}

	for _, v := range installed {
		if v.Version.Compare(version) == 0 {
			return nil
		}
	}

	downloadPath, err := m.installer.Download(version)
	if err != nil {
		return err
	}

	if githubInstaller, ok := m.installer.(*GitHubVersionInstaller); ok {
		checksum, err := githubInstaller.GetChecksum(version)
		if err == nil && checksum != "" {
			if err := m.installer.Verify(downloadPath, checksum); err != nil {
				return err
			}
		}
	}

	versionDir := filepath.Join(m.installDir, "versions", version.String())
	binaryPath := filepath.Join(versionDir, "dev-stack")
	if err := m.installer.Install(downloadPath, binaryPath); err != nil {
		return err
	}

	return m.registry.RegisterInstalledVersion(version, binaryPath)
}

// UninstallVersion removes a specific version
func (m *DefaultVersionManager) UninstallVersion(version Version) error {
	versionDir := filepath.Join(m.installDir, "versions", version.String())

	if err := os.RemoveAll(versionDir); err != nil {
		return NewVersionError(ErrVersionInstall,
			"failed to remove version directory: "+versionDir, err)
	}

	return m.registry.UnregisterInstalledVersion(version)
}

// VerifyVersion verifies an installed version
func (m *DefaultVersionManager) VerifyVersion(version Version) error {
	installed, err := m.ListInstalledVersions()
	if err != nil {
		return err
	}

	for _, v := range installed {
		if v.Version.Compare(version) == 0 {
			if _, err := os.Stat(v.Path); err != nil {
				return NewVersionError(ErrVersionNotFound,
					fmt.Sprintf("version %s binary not found at: %s", version.String(), v.Path), err)
			}
			return nil
		}
	}

	return NewVersionError(ErrVersionNotFound,
		"version "+version.String()+" is not installed", nil)
}

// CleanupOldVersions removes old versions, keeping only the specified number
func (m *DefaultVersionManager) CleanupOldVersions(keepCount int) error {
	installed, err := m.ListInstalledVersions()
	if err != nil {
		return err
	}

	if len(installed) <= keepCount {
		return nil
	}

	sort.Slice(installed, func(i, j int) bool {
		return installed[i].InstallDate.Before(installed[j].InstallDate)
	})

	var toKeep []InstalledVersion
	var toRemove []InstalledVersion

	for _, v := range installed {
		if v.Active {
			toKeep = append(toKeep, v)
		}
	}

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

	for _, v := range toRemove {
		if err := m.UninstallVersion(v.Version); err != nil {
			return err
		}
	}

	return nil
}

// GarbageCollect performs garbage collection of unused files
func (m *DefaultVersionManager) GarbageCollect() error {
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

	validDirs := make(map[string]bool)
	for _, v := range installed {
		validDirs[v.Version.String()] = true
	}

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
