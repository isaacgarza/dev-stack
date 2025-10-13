package constants

import "time"

// Version file names
const (
	VersionFileName        = ".dev-stack-version"
	VersionFileNameYAML    = ".dev-stack-version.yml"
	VersionFileNameYAML2   = ".dev-stack-version.yaml"
	NotificationConfigFile = "notifications.json"
)

// Version constraint operators
const (
	OperatorEqual          = "="
	OperatorEqualAlternate = "=="
	OperatorNotEqual       = "!="
	OperatorGreater        = ">"
	OperatorGreaterOrEqual = ">="
	OperatorLess           = "<"
	OperatorLessOrEqual    = "<="
	OperatorTilde          = "~"
	OperatorCaret          = "^"
	OperatorWildcard       = "*"
)

// Drift types
const (
	DriftTypeNone       = "none"
	DriftTypeMajor      = "major"
	DriftTypeMinor      = "minor"
	DriftTypePatch      = "patch"
	DriftTypePrerelease = "prerelease"
)

// Severity levels
const (
	SeverityInfo     = "info"
	SeverityWarning  = "warning"
	SeverityCritical = "critical"
)

// Update types
const (
	UpdateTypeMajor      = "major"
	UpdateTypeMinor      = "minor"
	UpdateTypePatch      = "patch"
	UpdateTypePrerelease = "prerelease"
)

// Update severity levels
const (
	UpdateSeverityOptional    = "optional"
	UpdateSeverityRecommended = "recommended"
	UpdateSeverityCritical    = "critical"
)

// Notification frequencies
const (
	NotificationFrequencyNever  = "never"
	NotificationFrequencyAlways = "always"
	NotificationFrequencyDaily  = "daily"
	NotificationFrequencyWeekly = "weekly"
)

// Enforcement actions
const (
	EnforcementActionNone    = "none"
	EnforcementActionWarn    = "warn"
	EnforcementActionSwitch  = "switch"
	EnforcementActionInstall = "install"
)

// Default durations
const (
	DefaultMaxDriftDuration = 7 * 24 * time.Hour // 1 week
	DefaultCheckInterval    = 24 * time.Hour     // 1 day
)

// Default enforcement policy
const (
	DefaultStrictMode     = false
	DefaultAllowDrift     = true
	DefaultAutoSync       = false
	DefaultNotifyUpdates  = true
	DefaultShowPrerelease = false
	DefaultAutoCheck      = true
)

// Version source types
const (
	VersionSourceGitHub = "github"
	VersionSourceLocal  = "local"
	VersionSourceBinary = "binary"
)
