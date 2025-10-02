package docs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultGenerationOptions(t *testing.T) {
	opts := DefaultGenerationOptions()

	if opts.CommandsYAMLPath != "scripts/commands.yaml" {
		t.Errorf("Expected CommandsYAMLPath to be 'scripts/commands.yaml', got %s", opts.CommandsYAMLPath)
	}

	if opts.ServicesYAMLPath != "services/services.yaml" {
		t.Errorf("Expected ServicesYAMLPath to be 'services/services.yaml', got %s", opts.ServicesYAMLPath)
	}

	if opts.ReferenceMDPath != "docs/reference.md" {
		t.Errorf("Expected ReferenceMDPath to be 'docs/reference.md', got %s", opts.ReferenceMDPath)
	}

	if opts.ServicesMDPath != "docs/services.md" {
		t.Errorf("Expected ServicesMDPath to be 'docs/services.md', got %s", opts.ServicesMDPath)
	}

	if opts.Verbose != false {
		t.Errorf("Expected Verbose to be false, got %v", opts.Verbose)
	}

	if opts.DryRun != false {
		t.Errorf("Expected DryRun to be false, got %v", opts.DryRun)
	}
}

func TestParseCommands(t *testing.T) {
	// Create temporary commands YAML file
	tmpDir := t.TempDir()
	commandsFile := filepath.Join(tmpDir, "commands.yaml")

	commandsYAML := `
dev-stack:
  - backup
  - cleanup
  - init
global-flags:
  - --config
  - --help
`

	if err := os.WriteFile(commandsFile, []byte(commandsYAML), 0644); err != nil {
		t.Fatalf("Failed to create test commands file: %v", err)
	}

	opts := &GenerationOptions{
		CommandsYAMLPath: commandsFile,
	}

	parser := NewParser(opts)
	commands, err := parser.ParseCommands()

	if err != nil {
		t.Fatalf("Failed to parse commands: %v", err)
	}

	if len(*commands) != 2 {
		t.Errorf("Expected 2 command groups, got %d", len(*commands))
	}

	devStackCommands := (*commands)["dev-stack"]
	if len(devStackCommands) != 3 {
		t.Errorf("Expected 3 dev-stack commands, got %d", len(devStackCommands))
	}

	if devStackCommands[0] != "backup" {
		t.Errorf("Expected first command to be 'backup', got '%s'", devStackCommands[0])
	}
}

func TestParseServices(t *testing.T) {
	// Create temporary services YAML file
	tmpDir := t.TempDir()
	servicesFile := filepath.Join(tmpDir, "services.yaml")

	servicesYAML := `
redis:
  description: In-memory data store for caching and session storage.
  options:
    - port
    - password
  examples:
    - "redis-cli -h localhost -p 6379 ping"
  usage_notes: "Use Redis for caching, session storage, and pub/sub."
  links:
    - "https://redis.io/documentation"

postgres:
  description: Relational database (PostgreSQL) for structured data.
  options:
    - port
    - database
  examples:
    - 'psql -h localhost -U postgres -c "SELECT version();"'
  usage_notes: "Ideal for structured data and transactional workloads."
  links:
    - "https://www.postgresql.org/docs/"
`

	if err := os.WriteFile(servicesFile, []byte(servicesYAML), 0644); err != nil {
		t.Fatalf("Failed to create test services file: %v", err)
	}

	opts := &GenerationOptions{
		ServicesYAMLPath: servicesFile,
	}

	parser := NewParser(opts)
	services, err := parser.ParseServices()

	if err != nil {
		t.Fatalf("Failed to parse services: %v", err)
	}

	if len(*services) != 2 {
		t.Errorf("Expected 2 services, got %d", len(*services))
	}

	redis := (*services)["redis"]
	if redis.Description != "In-memory data store for caching and session storage." {
		t.Errorf("Unexpected redis description: %s", redis.Description)
	}

	if len(redis.Options) != 2 {
		t.Errorf("Expected 2 redis options, got %d", len(redis.Options))
	}

	if redis.Options[0] != "port" {
		t.Errorf("Expected first redis option to be 'port', got '%s'", redis.Options[0])
	}
}

func TestValidateCommandsManifest(t *testing.T) {
	parser := NewParser(&GenerationOptions{})

	// Test nil manifest
	err := parser.ValidateCommandsManifest(nil)
	if err == nil {
		t.Error("Expected error for nil manifest")
	}

	// Test empty manifest
	empty := make(CommandsManifest)
	err = parser.ValidateCommandsManifest(&empty)
	if err == nil {
		t.Error("Expected error for empty manifest")
	}

	// Test valid manifest
	valid := CommandsManifest{
		"dev-stack": []string{"init", "up", "down"},
	}
	err = parser.ValidateCommandsManifest(&valid)
	if err != nil {
		t.Errorf("Unexpected error for valid manifest: %v", err)
	}

	// Test manifest with empty script name
	invalidScript := CommandsManifest{
		"": []string{"init"},
	}
	err = parser.ValidateCommandsManifest(&invalidScript)
	if err == nil {
		t.Error("Expected error for empty script name")
	}

	// Test manifest with no commands for script
	noCommands := CommandsManifest{
		"dev-stack": []string{},
	}
	err = parser.ValidateCommandsManifest(&noCommands)
	if err == nil {
		t.Error("Expected error for script with no commands")
	}
}

func TestValidateServicesManifest(t *testing.T) {
	parser := NewParser(&GenerationOptions{})

	// Test nil manifest
	err := parser.ValidateServicesManifest(nil)
	if err == nil {
		t.Error("Expected error for nil manifest")
	}

	// Test empty manifest
	empty := make(ServicesManifest)
	err = parser.ValidateServicesManifest(&empty)
	if err == nil {
		t.Error("Expected error for empty manifest")
	}

	// Test valid manifest
	valid := ServicesManifest{
		"redis": ServiceInfo{
			Description: "Redis cache",
			Options:     []string{"port"},
		},
	}
	err = parser.ValidateServicesManifest(&valid)
	if err != nil {
		t.Errorf("Unexpected error for valid manifest: %v", err)
	}

	// Test manifest with empty service name
	invalidService := ServicesManifest{
		"": ServiceInfo{Description: "Test"},
	}
	err = parser.ValidateServicesManifest(&invalidService)
	if err == nil {
		t.Error("Expected error for empty service name")
	}

	// Test manifest with no description
	noDescription := ServicesManifest{
		"redis": ServiceInfo{Description: ""},
	}
	err = parser.ValidateServicesManifest(&noDescription)
	if err == nil {
		t.Error("Expected error for service with no description")
	}
}

func TestGenerateCommandReference(t *testing.T) {
	generator := NewGenerator(&GenerationOptions{})

	commands := &CommandsManifest{
		"dev-stack": []string{"init", "up", "down"},
		"flags":     []string{"--help", "--verbose"},
	}

	content, err := generator.GenerateCommandReference(commands)
	if err != nil {
		t.Fatalf("Failed to generate command reference: %v", err)
	}

	if !strings.Contains(content, "# Command Reference (dev-stack)") {
		t.Error("Generated content should contain title")
	}

	if !strings.Contains(content, "## dev-stack") {
		t.Error("Generated content should contain dev-stack section")
	}

	if !strings.Contains(content, "- `init`") {
		t.Error("Generated content should contain init command")
	}

	if !strings.Contains(content, "scripts/commands.yaml") {
		t.Error("Generated content should reference source file")
	}
}

func TestGenerateServicesGuide(t *testing.T) {
	generator := NewGenerator(&GenerationOptions{})

	services := &ServicesManifest{
		"redis": ServiceInfo{
			Description: "Redis cache",
			Options:     []string{"port", "password"},
			Examples:    []string{"redis-cli ping"},
			UsageNotes:  "Use for caching",
			Links:       []string{"https://redis.io"},
		},
	}

	content, err := generator.GenerateServicesGuide(services)
	if err != nil {
		t.Fatalf("Failed to generate services guide: %v", err)
	}

	if !strings.Contains(content, "# Services Guide (dev-stack)") {
		t.Error("Generated content should contain title")
	}

	if !strings.Contains(content, "## redis") {
		t.Error("Generated content should contain redis section")
	}

	if !strings.Contains(content, "Redis cache") {
		t.Error("Generated content should contain service description")
	}

	if !strings.Contains(content, "**Options:**") {
		t.Error("Generated content should contain options section")
	}

	if !strings.Contains(content, "- `port`") {
		t.Error("Generated content should contain port option")
	}

	if !strings.Contains(content, "**Examples:**") {
		t.Error("Generated content should contain examples section")
	}

	if !strings.Contains(content, "**Usage Notes:**") {
		t.Error("Generated content should contain usage notes")
	}

	if !strings.Contains(content, "**Links:**") {
		t.Error("Generated content should contain links section")
	}

	if !strings.Contains(content, "services/services.yaml") {
		t.Error("Generated content should reference source file")
	}
}

func TestUpdateAutoGenSection(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")

	// Create test file with auto-gen markers
	originalContent := `# Test Document

Some content before.

<!-- AUTO-GENERATED-START -->
Old generated content
<!-- AUTO-GENERATED-END -->

Some content after.
`

	if err := os.WriteFile(testFile, []byte(originalContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	opts := &GenerationOptions{
		ReferenceMDPath: testFile,
		Verbose:         false,
		DryRun:          false,
	}

	generator := NewGenerator(opts)
	newContent := "New generated content"

	err := generator.updateAutoGenSection(testFile, newContent)
	if err != nil {
		t.Fatalf("Failed to update auto-gen section: %v", err)
	}

	// Read updated file
	updatedBytes, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read updated file: %v", err)
	}

	updated := string(updatedBytes)

	if !strings.Contains(updated, "New generated content") {
		t.Error("Updated file should contain new generated content")
	}

	if strings.Contains(updated, "Old generated content") {
		t.Error("Updated file should not contain old generated content")
	}

	if !strings.Contains(updated, "Some content before.") {
		t.Error("Updated file should preserve content before markers")
	}

	if !strings.Contains(updated, "Some content after.") {
		t.Error("Updated file should preserve content after markers")
	}
}

func TestUpdateAutoGenSectionDryRun(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")

	// Create test file with auto-gen markers
	originalContent := `# Test Document

<!-- AUTO-GENERATED-START -->
Old content
<!-- AUTO-GENERATED-END -->
`

	if err := os.WriteFile(testFile, []byte(originalContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	opts := &GenerationOptions{
		ReferenceMDPath: testFile,
		Verbose:         true,
		DryRun:          true,
	}

	generator := NewGenerator(opts)
	newContent := "New generated content"

	err := generator.updateAutoGenSection(testFile, newContent)
	if err != nil {
		t.Fatalf("Failed to update auto-gen section in dry-run: %v", err)
	}

	// Read file - should be unchanged
	unchangedBytes, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read unchanged file: %v", err)
	}

	unchanged := string(unchangedBytes)

	if strings.Contains(unchanged, "New generated content") {
		t.Error("File should be unchanged in dry-run mode")
	}

	if !strings.Contains(unchanged, "Old content") {
		t.Error("File should still contain old content in dry-run mode")
	}
}

func TestCreateNewDocFile(t *testing.T) {
	tmpDir := t.TempDir()
	newFile := filepath.Join(tmpDir, "subdir", "new.md")

	opts := &GenerationOptions{
		Verbose: true,
		DryRun:  false,
	}

	generator := NewGenerator(opts)
	content := "Test generated content"

	err := generator.createNewDocFile(newFile, content)
	if err != nil {
		t.Fatalf("Failed to create new doc file: %v", err)
	}

	// Check that file was created
	if _, err := os.Stat(newFile); os.IsNotExist(err) {
		t.Error("New file should have been created")
	}

	// Read and verify content
	createdBytes, err := os.ReadFile(newFile)
	if err != nil {
		t.Fatalf("Failed to read created file: %v", err)
	}

	created := string(createdBytes)

	if !strings.Contains(created, StartMarker) {
		t.Error("Created file should contain start marker")
	}

	if !strings.Contains(created, EndMarker) {
		t.Error("Created file should contain end marker")
	}

	if !strings.Contains(created, "Test generated content") {
		t.Error("Created file should contain generated content")
	}
}

func TestGenerateCommandReferenceOnly(t *testing.T) {
	// Create temporary test files
	tmpDir := t.TempDir()
	commandsFile := filepath.Join(tmpDir, "commands.yaml")
	outputFile := filepath.Join(tmpDir, "reference.md")

	commandsYAML := `
dev-stack:
  - init
  - up
`

	if err := os.WriteFile(commandsFile, []byte(commandsYAML), 0644); err != nil {
		t.Fatalf("Failed to create test commands file: %v", err)
	}

	opts := &GenerationOptions{
		CommandsYAMLPath: commandsFile,
		ReferenceMDPath:  outputFile,
		Verbose:          false,
		DryRun:           false,
	}

	generator := NewGenerator(opts)

	err := generator.GenerateCommandReferenceOnly()
	if err != nil {
		t.Fatalf("Failed to generate command reference only: %v", err)
	}

	// Check that output file was created
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("Output file should have been created")
	}

	// Read and verify content
	outputBytes, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	output := string(outputBytes)

	if !strings.Contains(output, "# Command Reference (dev-stack)") {
		t.Error("Output should contain command reference title")
	}

	if !strings.Contains(output, "- `init`") {
		t.Error("Output should contain init command")
	}
}

func TestGenerateServicesGuideOnly(t *testing.T) {
	// Create temporary test files
	tmpDir := t.TempDir()
	servicesFile := filepath.Join(tmpDir, "services.yaml")
	outputFile := filepath.Join(tmpDir, "services.md")

	servicesYAML := `
redis:
  description: Redis cache
  options:
    - port
`

	if err := os.WriteFile(servicesFile, []byte(servicesYAML), 0644); err != nil {
		t.Fatalf("Failed to create test services file: %v", err)
	}

	opts := &GenerationOptions{
		ServicesYAMLPath: servicesFile,
		ServicesMDPath:   outputFile,
		Verbose:          false,
		DryRun:           false,
	}

	generator := NewGenerator(opts)

	err := generator.GenerateServicesGuideOnly()
	if err != nil {
		t.Fatalf("Failed to generate services guide only: %v", err)
	}

	// Check that output file was created
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("Output file should have been created")
	}

	// Read and verify content
	outputBytes, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	output := string(outputBytes)

	if !strings.Contains(output, "# Services Guide (dev-stack)") {
		t.Error("Output should contain services guide title")
	}

	if !strings.Contains(output, "## redis") {
		t.Error("Output should contain redis section")
	}

	if !strings.Contains(output, "Redis cache") {
		t.Error("Output should contain redis description")
	}
}
