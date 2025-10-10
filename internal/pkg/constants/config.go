package constants

// Default configuration values
const (
	DefaultEnvironment = "local"
	DefaultProjectName = "dev-stack"
)

// Configuration sections
const (
	ProjectSection    = "project"
	StackSection      = "stack"
	OverridesSection  = "overrides"
	ValidationSection = "validation"
	AdvancedSection   = "advanced"
)

// Default configuration values
const (
	DefaultSkipWarnings      = false
	DefaultAllowMultipleDBs  = true
	DefaultAutoStart         = true
	DefaultPullLatestImages  = true
	DefaultCleanupOnRecreate = false
)
