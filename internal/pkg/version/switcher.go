package version

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// VersionSwitcher handles automatic version delegation based on project requirements
type VersionSwitcher struct {
	manager  VersionManager
	detector *VersionDetector
}

// NewVersionSwitcher creates a new version switcher
func NewVersionSwitcher(manager VersionManager) *VersionSwitcher {
	return &VersionSwitcher{
		manager:  manager,
		detector: NewVersionDetector(),
	}
}

// DelegateToCorrectVersion determines the correct version and delegates execution
func (s *VersionSwitcher) DelegateToCorrectVersion(args []string) error {
	// Get current working directory for project context
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Find project root by looking for version files or git root
	projectRoot := s.findProjectRoot(cwd)

	// Detect required version for the project
	constraint, err := s.manager.DetectProjectVersion(projectRoot)
	if err != nil {
		// If we can't detect version requirements, use active version
		return s.delegateToActiveVersion(args)
	}

	// If no specific constraint, use active version
	if constraint.Original == "*" {
		return s.delegateToActiveVersion(args)
	}

	// Try to resolve to an installed version
	resolved, err := s.manager.ResolveVersion(*constraint)
	if err != nil {
		// No installed version matches, try to install one
		return s.handleMissingVersion(*constraint, args)
	}

	// Delegate to the resolved version
	return s.delegateToVersion(resolved, args)
}

// findProjectRoot finds the root directory of the project
func (s *VersionSwitcher) findProjectRoot(startPath string) string {
	currentPath := startPath

	for {
		// Check for version files
		versionFiles, _ := s.detector.FindVersionFiles(currentPath)
		if len(versionFiles) > 0 {
			return currentPath
		}

		// Check for git repository
		gitDir := filepath.Join(currentPath, ".git")
		if _, err := os.Stat(gitDir); err == nil {
			return currentPath
		}

		// Check for common project files
		projectFiles := []string{
			"go.mod",
			"package.json",
			"Cargo.toml",
			"requirements.txt",
			"pom.xml",
			"Makefile",
			"docker-compose.yml",
		}

		for _, file := range projectFiles {
			if _, err := os.Stat(filepath.Join(currentPath, file)); err == nil {
				return currentPath
			}
		}

		// Move up one directory
		parent := filepath.Dir(currentPath)
		if parent == currentPath || parent == "/" {
			// Reached filesystem root
			break
		}
		currentPath = parent
	}

	// Return original path if no project root found
	return startPath
}

// delegateToActiveVersion delegates to the currently active version
func (s *VersionSwitcher) delegateToActiveVersion(args []string) error {
	active, err := s.manager.GetActiveVersion()
	if err != nil {
		return fmt.Errorf("no active version set and no project-specific version found")
	}

	return s.delegateToVersion(active, args)
}

// delegateToVersion delegates execution to a specific version
func (s *VersionSwitcher) delegateToVersion(installedVersion *InstalledVersion, args []string) error {
	// Check if the binary exists
	if _, err := os.Stat(installedVersion.Path); err != nil {
		return NewVersionError(ErrVersionNotFound,
			fmt.Sprintf("version %s binary not found at: %s",
				installedVersion.Version.String(), installedVersion.Path), err)
	}

	// Check if we're already the correct version to avoid infinite recursion
	currentBinary, err := os.Executable()
	if err == nil {
		if currentBinary == installedVersion.Path {
			// We're already the correct version, this shouldn't happen
			return fmt.Errorf("infinite delegation loop detected")
		}
	}

	// Prepare arguments (remove the first argument which is the program name)
	execArgs := args
	if len(args) > 0 {
		execArgs = args[1:]
	}

	// Execute the correct version
	return s.executeVersion(installedVersion.Path, execArgs)
}

// executeVersion executes a specific version binary
func (s *VersionSwitcher) executeVersion(binaryPath string, args []string) error {
	// Use exec.Command for better control and error handling
	cmd := exec.Command(binaryPath, args...)

	// Pass through environment variables
	cmd.Env = os.Environ()

	// Connect stdin, stdout, stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set the working directory to current directory
	if cwd, err := os.Getwd(); err == nil {
		cmd.Dir = cwd
	}

	// Run the command and wait for completion
	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			// Exit with the same code as the delegated binary
			os.Exit(exitError.ExitCode())
		}
		return fmt.Errorf("failed to execute version %s: %w", binaryPath, err)
	}

	// Exit successfully
	os.Exit(0)
	return nil
}

// handleMissingVersion handles the case when no installed version matches requirements
func (s *VersionSwitcher) handleMissingVersion(constraint VersionConstraint, args []string) error {
	// Check if this is a version management command
	if s.isVersionManagementCommand(args) {
		// Allow version management commands to run with any available version
		return s.delegateToActiveVersion(args)
	}

	// For regular commands, suggest installing the required version
	fmt.Fprintf(os.Stderr, "Error: No installed version satisfies requirement: %s\n", constraint.Original)
	fmt.Fprintf(os.Stderr, "Run 'dev-stack versions install %s' to install a compatible version.\n", constraint.Original)

	// Try to suggest the best available version
	available, err := s.manager.ListAvailableVersions()
	if err == nil {
		bestMatch := s.findBestAvailableMatch(constraint, available)
		if bestMatch != nil {
			fmt.Fprintf(os.Stderr, "Suggested version: %s\n", bestMatch.String())
		}
	}

	os.Exit(1)
	return nil
}

// isVersionManagementCommand checks if the command is related to version management
func (s *VersionSwitcher) isVersionManagementCommand(args []string) bool {
	if len(args) < 2 {
		return false
	}

	versionCommands := []string{
		"versions",
		"version",
	}

	for _, cmd := range versionCommands {
		if args[1] == cmd {
			return true
		}
	}

	return false
}

// findBestAvailableMatch finds the best available version that matches a constraint
func (s *VersionSwitcher) findBestAvailableMatch(constraint VersionConstraint, available []Version) *Version {
	var candidates []Version

	for _, v := range available {
		if constraint.Satisfies(v) {
			candidates = append(candidates, v)
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	// Return the latest matching version
	best := candidates[0]
	for _, v := range candidates[1:] {
		if v.Compare(best) > 0 {
			best = v
		}
	}

	return &best
}

// AutoInstallAndDelegate automatically installs a missing version and delegates to it
func (s *VersionSwitcher) AutoInstallAndDelegate(constraint VersionConstraint, args []string) error {
	// Find the best available version
	available, err := s.manager.ListAvailableVersions()
	if err != nil {
		return fmt.Errorf("failed to list available versions: %w", err)
	}

	bestMatch := s.findBestAvailableMatch(constraint, available)
	if bestMatch == nil {
		return fmt.Errorf("no available version satisfies constraint: %s", constraint.Original)
	}

	fmt.Printf("Auto-installing dev-stack version %s...\n", bestMatch.String())

	// Install the version
	if err := s.manager.InstallVersion(*bestMatch); err != nil {
		return fmt.Errorf("failed to auto-install version %s: %w", bestMatch.String(), err)
	}

	fmt.Printf("Successfully installed version %s\n", bestMatch.String())

	// Resolve and delegate
	resolved, err := s.manager.ResolveVersion(constraint)
	if err != nil {
		return fmt.Errorf("failed to resolve version after installation: %w", err)
	}

	return s.delegateToVersion(resolved, args)
}

// CheckVersionCompatibility checks if the current version is compatible with project requirements
func (s *VersionSwitcher) CheckVersionCompatibility(projectPath string) (*CompatibilityResult, error) {
	constraint, err := s.manager.DetectProjectVersion(projectPath)
	if err != nil {
		return nil, err
	}

	// Get current version (the version of the binary being executed)
	currentVersion, err := s.getCurrentVersion()
	if err != nil {
		return nil, err
	}

	result := &CompatibilityResult{
		ProjectPath:        projectPath,
		RequiredConstraint: *constraint,
		CurrentVersion:     currentVersion,
		Compatible:         constraint.Satisfies(currentVersion),
	}

	if !result.Compatible {
		// Try to find a compatible installed version
		resolved, err := s.manager.ResolveVersion(*constraint)
		if err == nil {
			result.RecommendedVersion = &resolved.Version
		}
	}

	return result, nil
}

// getCurrentVersion gets the version of the currently executing binary
func (s *VersionSwitcher) getCurrentVersion() (Version, error) {
	// This would typically be set at build time
	// For now, we'll try to determine it from the active version
	active, err := s.manager.GetActiveVersion()
	if err != nil {
		// Fallback to a default version
		return Version{Major: 0, Minor: 1, Patch: 0, Original: "0.1.0"}, nil
	}

	return active.Version, nil
}

// CompatibilityResult represents the result of a version compatibility check
type CompatibilityResult struct {
	ProjectPath        string            `json:"project_path"`
	RequiredConstraint VersionConstraint `json:"required_constraint"`
	CurrentVersion     Version           `json:"current_version"`
	Compatible         bool              `json:"compatible"`
	RecommendedVersion *Version          `json:"recommended_version,omitempty"`
}

// ShouldDelegate determines if we should delegate to a different version
func (s *VersionSwitcher) ShouldDelegate(args []string) (bool, *InstalledVersion, error) {
	// Skip delegation for version management commands
	if s.isVersionManagementCommand(args) {
		return false, nil, nil
	}

	// Get current working directory for project context
	cwd, err := os.Getwd()
	if err != nil {
		return false, nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	// Find project root
	projectRoot := s.findProjectRoot(cwd)

	// Detect required version for the project
	constraint, err := s.manager.DetectProjectVersion(projectRoot)
	if err != nil || constraint.Original == "*" {
		// No specific requirements, don't delegate
		return false, nil, nil
	}

	// Get current version
	currentVersion, err := s.getCurrentVersion()
	if err != nil {
		return false, nil, err
	}

	// Check if current version satisfies the constraint
	if constraint.Satisfies(currentVersion) {
		return false, nil, nil
	}

	// Try to resolve to a different installed version
	resolved, err := s.manager.ResolveVersion(*constraint)
	if err != nil {
		return false, nil, err
	}

	// Check if the resolved version is different from current
	if resolved.Version.Compare(currentVersion) == 0 {
		return false, nil, nil
	}

	return true, resolved, nil
}
