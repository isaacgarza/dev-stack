package version

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// GetDefaultInstallDir returns the default installation directory for dev-stack versions
func GetDefaultInstallDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".dev-stack"), nil
}

// GetDefaultConfigDir returns the default configuration directory for dev-stack
func GetDefaultConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".config", "dev-stack"), nil
}

// EnsureDirectoryExists creates a directory if it doesn't exist
func EnsureDirectoryExists(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0755)
	}
	return nil
}

// GetPlatformBinaryName returns the binary name for the current platform
func GetPlatformBinaryName(baseName string) string {
	if runtime.GOOS == "windows" {
		return baseName + ".exe"
	}
	return baseName
}

// GetPlatformArchiveName returns the archive name for the current platform
func GetPlatformArchiveName(baseName, version string) string {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// Map Go architectures to common naming conventions
	switch goarch {
	case "amd64":
		goarch = "x86_64"
	case "386":
		goarch = "i386"
	case "arm64":
		goarch = "aarch64"
	}

	// Map OS names to common conventions
	switch goos {
	case "darwin":
		goos = "macos"
	}

	var extension string
	if runtime.GOOS == "windows" {
		extension = ".zip"
	} else {
		extension = ".tar.gz"
	}

	return fmt.Sprintf("%s-%s-%s-%s%s", baseName, version, goos, goarch, extension)
}

// IsVersionInstalled checks if a specific version is installed
func IsVersionInstalled(installDir string, version Version) bool {
	versionDir := filepath.Join(installDir, "versions", version.String())
	binaryName := GetPlatformBinaryName("dev-stack")
	binaryPath := filepath.Join(versionDir, binaryName)

	_, err := os.Stat(binaryPath)
	return err == nil
}

// GetVersionBinaryPath returns the path to a version's binary
func GetVersionBinaryPath(installDir string, version Version) string {
	versionDir := filepath.Join(installDir, "versions", version.String())
	binaryName := GetPlatformBinaryName("dev-stack")
	return filepath.Join(versionDir, binaryName)
}

// FindExecutableInPath finds an executable in the system PATH
func FindExecutableInPath(name string) (string, error) {
	pathEnv := os.Getenv("PATH")
	if pathEnv == "" {
		return "", fmt.Errorf("PATH environment variable not set")
	}

	pathSeparator := ":"
	if runtime.GOOS == "windows" {
		pathSeparator = ";"
	}

	paths := strings.Split(pathEnv, pathSeparator)
	execName := GetPlatformBinaryName(name)

	for _, path := range paths {
		if path == "" {
			continue
		}

		fullPath := filepath.Join(path, execName)
		if info, err := os.Stat(fullPath); err == nil && !info.IsDir() {
			// Check if file is executable
			if isExecutable(fullPath) {
				return fullPath, nil
			}
		}
	}

	return "", fmt.Errorf("executable %s not found in PATH", name)
}

// isExecutable checks if a file is executable
func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	if runtime.GOOS == "windows" {
		// On Windows, check file extension
		return strings.HasSuffix(strings.ToLower(path), ".exe")
	}

	// On Unix-like systems, check execute permission
	return info.Mode()&0111 != 0
}

// CleanVersionString normalizes a version string for consistent comparison
func CleanVersionString(version string) string {
	version = strings.TrimSpace(version)

	// Remove common prefixes
	version = strings.TrimPrefix(version, "v")
	if strings.HasPrefix(version, "version") {
		version = strings.TrimSpace(version[7:])
	}

	return version
}

// IsValidVersionString performs basic validation on a version string
func IsValidVersionString(version string) bool {
	if version == "" {
		return false
	}

	// Allow special versions
	if version == "latest" || version == "*" {
		return true
	}

	// Basic semantic version check
	_, err := ParseVersion(version)
	return err == nil
}

// FormatVersionsList formats a list of versions for display
func FormatVersionsList(versions []Version, activeVersion *Version) []string {
	var formatted []string

	for _, v := range versions {
		line := v.String()
		if activeVersion != nil && v.Compare(*activeVersion) == 0 {
			line += " (active)"
		}
		formatted = append(formatted, line)
	}

	return formatted
}

// GetCurrentExecutablePath returns the path of the currently running executable
func GetCurrentExecutablePath() (string, error) {
	return os.Executable()
}

// IsDevStackBinary checks if a path points to a dev-stack binary
func IsDevStackBinary(path string) bool {
	if path == "" {
		return false
	}

	base := filepath.Base(path)
	base = strings.TrimSuffix(base, ".exe") // Remove .exe on Windows

	return base == "dev-stack"
}

// CreateSymlink creates a symbolic link from source to target
func CreateSymlink(source, target string) error {
	// Remove existing symlink/file if it exists
	if _, err := os.Lstat(target); err == nil {
		if err := os.Remove(target); err != nil {
			return fmt.Errorf("failed to remove existing target: %w", err)
		}
	}

	// Create directory for target if it doesn't exist
	targetDir := filepath.Dir(target)
	if err := EnsureDirectoryExists(targetDir); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Create the symlink
	return os.Symlink(source, target)
}

// ResolveSymlink resolves a symbolic link to its target
func ResolveSymlink(path string) (string, error) {
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return path, err // Return original path if not a symlink
	}
	return resolved, nil
}

// CalculateDirectorySize calculates the total size of a directory
func CalculateDirectorySize(dirPath string) (int64, error) {
	var size int64

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

// FormatByteSize formats a byte size into a human-readable string
func FormatByteSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// GetVersionFromBinary tries to extract version information from a binary
func GetVersionFromBinary(binaryPath string) (*Version, error) {
	// This would typically run the binary with --version flag
	// For now, we'll return an error as this requires more complex implementation
	return nil, fmt.Errorf("version extraction from binary not implemented")
}

// BackupFile creates a backup of a file with a .bak extension
func BackupFile(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil // File doesn't exist, nothing to backup
	}

	backupPath := filePath + ".bak"

	source, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer func() {
		_ = source.Close()
	}()

	backup, err := os.Create(backupPath)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}
	defer func() {
		_ = backup.Close()
	}()

	if _, err := backup.ReadFrom(source); err != nil {
		return fmt.Errorf("failed to copy to backup: %w", err)
	}

	return nil
}

// RestoreFromBackup restores a file from its backup
func RestoreFromBackup(filePath string) error {
	backupPath := filePath + ".bak"

	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup file does not exist: %s", backupPath)
	}

	return os.Rename(backupPath, filePath)
}

// GetLatestVersionFromList returns the latest version from a list
func GetLatestVersionFromList(versions []Version) *Version {
	if len(versions) == 0 {
		return nil
	}

	latest := versions[0]
	for _, v := range versions[1:] {
		if v.Compare(latest) > 0 {
			latest = v
		}
	}

	return &latest
}

// FilterVersionsByPattern filters versions based on a pattern
func FilterVersionsByPattern(versions []Version, pattern string) []Version {
	if pattern == "" || pattern == "*" {
		return versions
	}

	var filtered []Version
	for _, v := range versions {
		if strings.Contains(v.String(), pattern) {
			filtered = append(filtered, v)
		}
	}

	return filtered
}
