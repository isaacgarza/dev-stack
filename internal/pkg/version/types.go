package version

import (
	"fmt"
	"time"
)

// Version represents a semantic version
type Version struct {
	Major      int    `json:"major"`
	Minor      int    `json:"minor"`
	Patch      int    `json:"patch"`
	PreRelease string `json:"prerelease,omitempty"`
	Build      string `json:"build,omitempty"`
	Original   string `json:"original"`
}

// String returns the string representation of the version
func (v Version) String() string {
	if v.Original != "" {
		return v.Original
	}

	version := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.PreRelease != "" {
		version += "-" + v.PreRelease
	}
	if v.Build != "" {
		version += "+" + v.Build
	}
	return version
}

// Compare compares two versions, returns:
// -1 if v < other
//
//	0 if v == other
//	1 if v > other
func (v Version) Compare(other Version) int {
	if v.Major != other.Major {
		if v.Major < other.Major {
			return -1
		}
		return 1
	}

	if v.Minor != other.Minor {
		if v.Minor < other.Minor {
			return -1
		}
		return 1
	}

	if v.Patch != other.Patch {
		if v.Patch < other.Patch {
			return -1
		}
		return 1
	}

	// Handle pre-release versions
	if v.PreRelease == "" && other.PreRelease != "" {
		return 1 // release > pre-release
	}
	if v.PreRelease != "" && other.PreRelease == "" {
		return -1 // pre-release < release
	}
	if v.PreRelease != other.PreRelease {
		if v.PreRelease < other.PreRelease {
			return -1
		}
		return 1
	}

	return 0
}

// VersionConstraint represents a version constraint like ">=1.0.0", "~1.2.3", etc.
type VersionConstraint struct {
	Operator string  `json:"operator"`
	Version  Version `json:"version"`
	Original string  `json:"original"`
}

// Satisfies checks if a version satisfies the constraint
func (c VersionConstraint) Satisfies(version Version) bool {
	cmp := version.Compare(c.Version)

	switch c.Operator {
	case "=", "==":
		return cmp == 0
	case "!=":
		return cmp != 0
	case ">":
		return cmp > 0
	case ">=":
		return cmp >= 0
	case "<":
		return cmp < 0
	case "<=":
		return cmp <= 0
	case "~":
		// Tilde allows patch-level changes within the same minor version
		// ~1.2.3 allows >=1.2.3 but <1.3.0
		// ~1.2.0 allows >=1.2.0 but <1.3.0
		return version.Major == c.Version.Major &&
			version.Minor == c.Version.Minor &&
			cmp >= 0
	case "^":
		// Caret allows minor-level changes
		return version.Major == c.Version.Major && cmp >= 0
	case "*":
		return true
	default:
		return false
	}
}

// IsSatisfiedBy checks if a version satisfies the constraint (alias for Satisfies)
func (c VersionConstraint) IsSatisfiedBy(version Version) bool {
	return c.Satisfies(version)
}

// InstalledVersion represents an installed version of dev-stack
type InstalledVersion struct {
	Version     Version   `json:"version"`
	Path        string    `json:"path"`
	InstallDate time.Time `json:"install_date"`
	Source      string    `json:"source"`   // "github", "local", etc.
	Checksum    string    `json:"checksum"` // SHA256 checksum
	Active      bool      `json:"active"`   // Is this the currently active version
}

// ProjectVersionConfig represents version configuration for a specific project
type ProjectVersionConfig struct {
	ProjectPath string            `json:"project_path"`
	Required    VersionConstraint `json:"required"`
	Preferred   *Version          `json:"preferred,omitempty"`
	LastUsed    time.Time         `json:"last_used"`
}

// VersionFile represents the contents of a .dev-stack-version file
type VersionFile struct {
	Version        string            `yaml:"version" json:"version"`
	MinimumVersion string            `yaml:"minimum_version,omitempty" json:"minimum_version,omitempty"`
	MaximumVersion string            `yaml:"maximum_version,omitempty" json:"maximum_version,omitempty"`
	Constraints    []string          `yaml:"constraints,omitempty" json:"constraints,omitempty"`
	Metadata       map[string]string `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

// VersionManager interface defines the contract for version management
type VersionManager interface {
	// Version Detection
	DetectProjectVersion(projectPath string) (*VersionConstraint, error)
	ParseVersionFile(path string) (*VersionFile, error)
	ParseVersionConstraint(constraint string) (*VersionConstraint, error)

	// Version Installation
	ListAvailableVersions() ([]Version, error)
	InstallVersion(version Version) error
	UninstallVersion(version Version) error
	VerifyVersion(version Version) error

	// Version Management
	ListInstalledVersions() ([]InstalledVersion, error)
	GetActiveVersion() (*InstalledVersion, error)
	SetActiveVersion(version Version) error

	// Version Switching
	ResolveVersion(constraint VersionConstraint) (*InstalledVersion, error)
	SwitchToVersion(version Version) error

	// Multi-Project Support
	GetProjectConfig(projectPath string) (*ProjectVersionConfig, error)
	SetProjectConfig(config ProjectVersionConfig) error
	ListProjectConfigs() ([]ProjectVersionConfig, error)

	// Cleanup
	CleanupOldVersions(keepCount int) error
	GarbageCollect() error
}

// VersionResolver interface for resolving version conflicts
type VersionResolver interface {
	ResolveConflict(constraints []VersionConstraint) (*Version, error)
	FindBestMatch(constraints []VersionConstraint, available []Version) (*Version, error)
}

// VersionInstaller interface for installing versions
type VersionInstaller interface {
	Download(version Version) (string, error)
	Verify(path string, expectedChecksum string) error
	Install(sourcePath, targetPath string) error
}

// Error types for version management
type VersionError struct {
	Type    string
	Message string
	Cause   error
}

func (e VersionError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Common error types
var (
	ErrVersionNotFound      = "VERSION_NOT_FOUND"
	ErrVersionInvalid       = "VERSION_INVALID"
	ErrVersionAlreadyExists = "VERSION_ALREADY_EXISTS"
	ErrVersionConstraint    = "VERSION_CONSTRAINT"
	ErrVersionInstall       = "VERSION_INSTALL"
	ErrVersionSwitch        = "VERSION_SWITCH"
	ErrProjectConfig        = "PROJECT_CONFIG"
)

// NewVersionError creates a new version error
func NewVersionError(errorType, message string, cause error) *VersionError {
	return &VersionError{
		Type:    errorType,
		Message: message,
		Cause:   cause,
	}
}
