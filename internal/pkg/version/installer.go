package version

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// GitHubVersionInstaller implements version installation from GitHub releases
type GitHubVersionInstaller struct {
	owner      string
	repo       string
	installDir string
	client     *http.Client

	binaryName string
}

// NewGitHubVersionInstaller creates a new GitHub version installer
func NewGitHubVersionInstaller(owner, repo, installDir string) *GitHubVersionInstaller {
	return &GitHubVersionInstaller{
		owner:      owner,
		repo:       repo,
		installDir: installDir,
		client: &http.Client{
			Timeout: 30 * time.Minute,
		},
		binaryName: "dev-stack",
	}
}

// Download downloads a specific version from GitHub releases
func (g *GitHubVersionInstaller) Download(version Version) (string, error) {
	// Create version-specific directory
	versionDir := filepath.Join(g.installDir, "versions", version.String())
	if err := os.MkdirAll(versionDir, 0755); err != nil {
		return "", NewVersionError(ErrVersionInstall,
			fmt.Sprintf("failed to create version directory: %s", versionDir), err)
	}

	// Determine download URL and file name
	downloadURL, fileName := g.buildDownloadURL(version)
	targetPath := filepath.Join(versionDir, fileName)

	// Check if already downloaded
	if _, err := os.Stat(targetPath); err == nil {
		return targetPath, nil
	}

	// Download the file
	if err := g.downloadFile(downloadURL, targetPath); err != nil {
		return "", NewVersionError(ErrVersionInstall,
			fmt.Sprintf("failed to download version %s", version.String()), err)
	}

	return targetPath, nil
}

// buildDownloadURL constructs the GitHub release download URL
func (g *GitHubVersionInstaller) buildDownloadURL(version Version) (string, string) {
	// Determine OS and architecture
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

	// Build file name
	var fileName string
	var extension string

	if runtime.GOOS == "windows" {
		extension = ".zip"
	} else {
		extension = ".tar.gz"
	}

	fileName = fmt.Sprintf("%s-%s-%s-%s%s",
		g.binaryName, version.String(), goos, goarch, extension)

	// Build download URL
	downloadURL := fmt.Sprintf("https://github.com/%s/%s/releases/download/v%s/%s",
		g.owner, g.repo, version.String(), fileName)

	return downloadURL, fileName
}

// downloadFile downloads a file from URL to target path
func (g *GitHubVersionInstaller) downloadFile(url, targetPath string) error {
	resp, err := g.client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download from %s: %w", url, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d: %s", resp.StatusCode, resp.Status)
	}

	// Create the target file
	out, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", targetPath, err)
	}
	defer func() {
		_ = out.Close()
	}()

	// Copy the response body to the file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write to file %s: %w", targetPath, err)
	}

	return nil
}

// Verify verifies the downloaded file against expected checksum
func (g *GitHubVersionInstaller) Verify(path string, expectedChecksum string) error {
	if expectedChecksum == "" {
		// Skip verification if no checksum provided
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		return NewVersionError(ErrVersionInstall,
			fmt.Sprintf("failed to open file for verification: %s", path), err)
	}
	defer func() {
		_ = file.Close()
	}()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return NewVersionError(ErrVersionInstall,
			fmt.Sprintf("failed to calculate checksum for: %s", path), err)
	}

	actualChecksum := hex.EncodeToString(hasher.Sum(nil))
	if actualChecksum != expectedChecksum {
		return NewVersionError(ErrVersionInstall,
			fmt.Sprintf("checksum mismatch for %s: expected %s, got %s",
				path, expectedChecksum, actualChecksum), nil)
	}

	return nil
}

// Install extracts and installs the downloaded version
func (g *GitHubVersionInstaller) Install(sourcePath, targetPath string) error {
	// Determine file type and extract accordingly
	if strings.HasSuffix(sourcePath, ".zip") {
		return g.extractZip(sourcePath, targetPath)
	} else if strings.HasSuffix(sourcePath, ".tar.gz") {
		return g.extractTarGz(sourcePath, targetPath)
	}

	// If it's a single binary, just copy it
	return g.copyBinary(sourcePath, targetPath)
}

// extractZip extracts a ZIP archive
func (g *GitHubVersionInstaller) extractZip(sourcePath, targetDir string) error {
	reader, err := zip.OpenReader(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open ZIP file: %w", err)
	}
	defer func() {
		_ = reader.Close()
	}()

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	for _, file := range reader.File {
		// Skip directories
		if file.FileInfo().IsDir() {
			continue
		}

		// Open file in ZIP
		rc, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open file in ZIP: %w", err)
		}

		// Create target file
		targetFile := filepath.Join(targetDir, filepath.Base(file.Name))
		out, err := os.OpenFile(targetFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			_ = rc.Close()
			return fmt.Errorf("failed to create target file: %w", err)
		}

		// Copy file content
		_, err = io.Copy(out, rc)
		_ = rc.Close()
		_ = out.Close()

		if err != nil {
			return fmt.Errorf("failed to extract file: %w", err)
		}

		// If this is the binary, make it executable
		if strings.Contains(file.Name, g.binaryName) || strings.HasSuffix(file.Name, ".exe") {
			if err := os.Chmod(targetFile, 0755); err != nil {
				return fmt.Errorf("failed to make binary executable: %w", err)
			}
		}
	}

	return nil
}

// extractTarGz extracts a tar.gz archive (simplified version)
func (g *GitHubVersionInstaller) extractTarGz(sourcePath, targetDir string) error {
	// For now, we'll implement a basic version
	// In a production system, you'd want to use archive/tar and compress/gzip
	return fmt.Errorf("tar.gz extraction not yet implemented")
}

// copyBinary copies a single binary file
func (g *GitHubVersionInstaller) copyBinary(sourcePath, targetPath string) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer func() {
		_ = source.Close()
	}()

	// Ensure target directory exists
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	target, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("failed to create target file: %w", err)
	}
	defer func() {
		_ = target.Close()
	}()

	if _, err := io.Copy(target, source); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	// Make it executable
	if err := os.Chmod(targetPath, 0755); err != nil {
		return fmt.Errorf("failed to make binary executable: %w", err)
	}

	return nil
}

// GetChecksum downloads and returns the checksum for a version
func (g *GitHubVersionInstaller) GetChecksum(version Version) (string, error) {
	// Build checksum URL
	checksumURL := fmt.Sprintf("https://github.com/%s/%s/releases/download/v%s/checksums.txt",
		g.owner, g.repo, version.String())

	resp, err := g.client.Get(checksumURL)
	if err != nil {
		return "", fmt.Errorf("failed to download checksums: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusNotFound {
		// Checksums not available for this version
		return "", nil
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download checksums: status %d", resp.StatusCode)
	}

	// Read and parse checksums
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read checksums: %w", err)
	}

	// Find the checksum for our platform
	_, fileName := g.buildDownloadURL(version)
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 && strings.Contains(line, fileName) {
			return parts[0], nil
		}
	}

	return "", fmt.Errorf("checksum not found for %s", fileName)
}

// ListAvailableVersions lists available versions from GitHub releases
func (g *GitHubVersionInstaller) ListAvailableVersions() ([]Version, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", g.owner, g.repo)

	resp, err := g.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch releases: status %d", resp.StatusCode)
	}

	// For simplicity, we'll parse this manually
	// In a production system, you'd want to use a proper JSON parser
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Extract version tags (simplified parsing)
	var versions []Version
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, `"tag_name"`) {
			// Extract tag name
			parts := strings.Split(line, `"`)
			if len(parts) >= 4 {
				tagName := parts[3]
				// Remove 'v' prefix if present
				tagName = strings.TrimPrefix(tagName, "v")

				version, err := ParseVersion(tagName)
				if err == nil {
					versions = append(versions, *version)
				}
			}
		}
	}

	return versions, nil
}

// Cleanup removes old downloaded files to free up space
func (g *GitHubVersionInstaller) Cleanup(keepVersions []Version) error {
	versionsDir := filepath.Join(g.installDir, "versions")

	// Get all version directories
	entries, err := os.ReadDir(versionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No versions directory
		}
		return fmt.Errorf("failed to read versions directory: %w", err)
	}

	// Convert keep versions to map for faster lookup
	keepMap := make(map[string]bool)
	for _, v := range keepVersions {
		keepMap[v.String()] = true
	}

	// Remove directories not in keep list
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		versionStr := entry.Name()
		if !keepMap[versionStr] {
			dirPath := filepath.Join(versionsDir, versionStr)
			if err := os.RemoveAll(dirPath); err != nil {
				return fmt.Errorf("failed to remove version directory %s: %w", dirPath, err)
			}
		}
	}

	return nil
}
