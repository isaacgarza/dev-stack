package docs

import "time"

// CommandsManifest represents the structure of commands.yaml
type CommandsManifest map[string][]string

// ServicesManifest represents the structure of services.yaml
type ServicesManifest map[string]ServiceInfo

// ServiceInfo represents the information for a single service
type ServiceInfo struct {
	Description string   `yaml:"description"`
	Options     []string `yaml:"options"`
	Examples    []string `yaml:"examples"`
	UsageNotes  string   `yaml:"usage_notes"`
	Links       []string `yaml:"links"`
}

// GenerationOptions represents options for documentation generation
type GenerationOptions struct {
	// Source file paths
	CommandsYAMLPath string
	ServicesYAMLPath string

	// Output file paths
	ReferenceMDPath string
	ServicesMDPath  string

	// Hugo integration paths
	HugoContentDir string
	DocsSourceDir  string
	EnableHugoSync bool

	// Generation settings
	Verbose bool
	DryRun  bool
}

// GenerationResult represents the result of documentation generation
type GenerationResult struct {
	CommandsGenerated bool
	ServicesGenerated bool
	ReadmeSynced      bool
	FilesUpdated      []string
	Errors            []error
	GeneratedAt       time.Time
}

// DocumentSection represents a section in a markdown document
type DocumentSection struct {
	Title       string
	Content     string
	Subsections []DocumentSection
}

// MarkdownDocument represents a complete markdown document
type MarkdownDocument struct {
	Title    string
	Sections []DocumentSection
}

// AutoGenSection represents an auto-generated section in existing markdown
type AutoGenSection struct {
	StartMarker string
	EndMarker   string
	Content     string
}

// HugoSyncResult represents the result of Hugo content synchronization
type HugoSyncResult struct {
	FilesCopied  []string
	FilesUpdated []string
	FilesSkipped []string
	Errors       []error
	SyncedAt     time.Time
}

// DefaultGenerationOptions returns the default options for documentation generation
func DefaultGenerationOptions() *GenerationOptions {
	return &GenerationOptions{
		CommandsYAMLPath: "scripts/commands.yaml",
		ServicesYAMLPath: "services/services.yaml",
		ReferenceMDPath:  "docs/reference.md",
		ServicesMDPath:   "docs/services.md",
		HugoContentDir:   "content",
		DocsSourceDir:    "docs",
		EnableHugoSync:   true,
		Verbose:          false,
		DryRun:           false,
	}
}
