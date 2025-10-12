package init

import "github.com/isaacgarza/dev-stack/internal/pkg/constants"

// Test constants to eliminate magic strings and provide context
const (
	// Test directory patterns
	TestTempDirPattern = "dev-stack-test-*"

	// Test project names
	TestProjectName        = "test-project"
	TestProjectNameValid   = "valid-project"
	TestProjectNameInvalid = "invalid@project"

	// Test services
	TestServicePostgres = "postgres"
	TestServiceRedis    = "redis"
	TestServiceNginx    = "nginx"

	// Test file content
	TestConfigContent    = "test: config"
	TestReadmeContent    = "# Test Project"
	TestGitignoreContent = "*.log\n*.tmp"
	TestExistingContent  = "# Existing content"

	// Test validation messages
	MsgAlreadyInitialized = "already initialized"
	MsgRequiredTool       = "required tool"
	MsgNoServicesSelected = "no services selected"
	MsgInvalidService     = "invalid service"
	MsgDuplicateService   = "duplicate service"
)

// Use constants from the constants package
const (
	// Test environments (use actual constants)
	TestEnvironmentLocal = constants.DefaultEnvironment
	TestEnvironmentDev   = "development"
	TestEnvironmentProd  = "production"

	// Test CLI commands (use actual constants)
	CmdDevStackUp     = constants.CmdUp
	CmdDevStackDown   = constants.CmdDown
	CmdDevStackStatus = constants.CmdStatus

	// Test gitignore entries (use actual constants)
	TestGitignoreEntry = constants.DevStackDir + "/" + constants.EnvGeneratedFileName
)

// Test file paths (use actual constants for consistency)
var (
	TestConfigFilePath     = constants.DevStackDir + "/" + constants.ConfigFileName
	TestConfigFilePathYAML = constants.DevStackDir + "/" + constants.ConfigFileNameYAML
	TestReadmeFilePath     = constants.DevStackDir + "/" + constants.ReadmeFileName
)
