package docs

import "time"

// CommandsManifest represents the enhanced structure of commands.yaml
type CommandsManifest struct {
	Metadata   CommandMetadata            `yaml:"metadata"`
	Global     GlobalConfig               `yaml:"global"`
	Categories map[string]CommandCategory `yaml:"categories"`
	Commands   map[string]CommandInfo     `yaml:"commands"`
	Workflows  map[string]WorkflowInfo    `yaml:"workflows"`
	Profiles   map[string]ProfileInfo     `yaml:"profiles"`
	Help       map[string]string          `yaml:"help"`
}

// CommandMetadata contains version and generation information
type CommandMetadata struct {
	Version     string    `yaml:"version"`
	GeneratedAt time.Time `yaml:"generated_at"`
	CLIVersion  string    `yaml:"cli_version"`
	Description string    `yaml:"description"`
}

// GlobalConfig contains global CLI configuration
type GlobalConfig struct {
	Flags map[string]FlagInfo `yaml:"flags"`
}

// CommandCategory represents a command category for organization
type CommandCategory struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Icon        string   `yaml:"icon"`
	Commands    []string `yaml:"commands"`
}

// CommandInfo represents detailed command information
type CommandInfo struct {
	Category        string              `yaml:"category"`
	Description     string              `yaml:"description"`
	LongDescription string              `yaml:"long_description"`
	Usage           string              `yaml:"usage"`
	Aliases         []string            `yaml:"aliases"`
	Examples        []ExampleInfo       `yaml:"examples"`
	Flags           map[string]FlagInfo `yaml:"flags"`
	RelatedCommands []string            `yaml:"related_commands"`
	Tips            []string            `yaml:"tips"`
	Hidden          bool                `yaml:"hidden,omitempty"`
}

// FlagInfo represents command flag information
type FlagInfo struct {
	Short       string      `yaml:"short,omitempty"`
	Type        string      `yaml:"type"`
	Description string      `yaml:"description"`
	Default     interface{} `yaml:"default"`
	Options     []string    `yaml:"options,omitempty"`
	Completion  string      `yaml:"completion,omitempty"`
	Required    bool        `yaml:"required,omitempty"`
	Hidden      bool        `yaml:"hidden,omitempty"`
}

// ExampleInfo represents a command usage example
type ExampleInfo struct {
	Command     string `yaml:"command"`
	Description string `yaml:"description"`
}

// WorkflowInfo represents a workflow definition
type WorkflowInfo struct {
	Name        string             `yaml:"name"`
	Description string             `yaml:"description"`
	Steps       []WorkflowStepInfo `yaml:"steps"`
}

// WorkflowStepInfo represents a workflow step
type WorkflowStepInfo struct {
	Command     string `yaml:"command"`
	Description string `yaml:"description"`
	Optional    bool   `yaml:"optional,omitempty"`
}

// ProfileInfo represents a service profile
type ProfileInfo struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Services    []string `yaml:"services"`
}

// ServicesManifest represents the structure of services.yaml
type ServicesManifest map[string]ServiceInfo

// ServiceInfo represents the information for a single service
type ServiceInfo struct {
	Name         string   `yaml:"name"`
	Description  string   `yaml:"description"`
	Options      []string `yaml:"options"`
	Examples     []string `yaml:"examples"`
	UsageNotes   string   `yaml:"usage_notes"`
	Links        []string `yaml:"links"`
	Category     string   `yaml:"category,omitempty"`
	DefaultPort  int      `yaml:"default_port,omitempty"`
	Dependencies []string `yaml:"dependencies,omitempty"`
	Tags         []string `yaml:"tags,omitempty"`
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
	CommandsGenerated  bool
	ServicesGenerated  bool
	ReadmeSynced       bool
	ContributingSynced bool
	FilesUpdated       []string
	Errors             []error
	GeneratedAt        time.Time
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

// CommandReference represents generated command reference content
type CommandReference struct {
	Categories []CategoryReference   `yaml:"categories"`
	Commands   []CommandReferenceDoc `yaml:"commands"`
	Workflows  []WorkflowReference   `yaml:"workflows"`
	Profiles   []ProfileReference    `yaml:"profiles"`
}

// CategoryReference represents a category in the reference
type CategoryReference struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Icon        string   `yaml:"icon"`
	Commands    []string `yaml:"commands"`
}

// CommandReferenceDoc represents a command in the reference
type CommandReferenceDoc struct {
	Name            string          `yaml:"name"`
	Category        string          `yaml:"category"`
	Description     string          `yaml:"description"`
	LongDescription string          `yaml:"long_description"`
	Usage           string          `yaml:"usage"`
	Aliases         []string        `yaml:"aliases"`
	Examples        []ExampleInfo   `yaml:"examples"`
	Flags           []FlagReference `yaml:"flags"`
	RelatedCommands []string        `yaml:"related_commands"`
	Tips            []string        `yaml:"tips"`
}

// FlagReference represents a flag in the reference
type FlagReference struct {
	Name        string      `yaml:"name"`
	Short       string      `yaml:"short,omitempty"`
	Type        string      `yaml:"type"`
	Description string      `yaml:"description"`
	Default     interface{} `yaml:"default"`
	Required    bool        `yaml:"required"`
	Options     []string    `yaml:"options,omitempty"`
}

// WorkflowReference represents a workflow in the reference
type WorkflowReference struct {
	Name        string             `yaml:"name"`
	Description string             `yaml:"description"`
	Steps       []WorkflowStepInfo `yaml:"steps"`
}

// ProfileReference represents a profile in the reference
type ProfileReference struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Services    []string `yaml:"services"`
}

// DefaultGenerationOptions returns the default options for documentation generation
func DefaultGenerationOptions() *GenerationOptions {
	return &GenerationOptions{
		CommandsYAMLPath: "internal/config/commands.yaml",
		ServicesYAMLPath: "internal/config/services/services.yaml",
		ReferenceMDPath:  "docs-site/content/reference.md",
		ServicesMDPath:   "docs-site/content/services.md",
		HugoContentDir:   "docs-site/content",
		DocsSourceDir:    "docs-site/content",
		EnableHugoSync:   true,
		Verbose:          false,
		DryRun:           false,
	}
}
