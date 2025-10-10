package constants

// File names
const (
	ConfigFileName           = "dev-stack-config.yml"
	ConfigFileNameYAML       = "dev-stack-config.yaml"
	ConfigFileNameHidden     = ".dev-stack-config.yml"
	ConfigFileNameHiddenYAML = ".dev-stack-config.yaml"
	DockerComposeFileName    = "docker-compose.yml"
	EnvGeneratedFileName     = ".env.generated"
	GitignoreFileName        = ".gitignore"
	ReadmeFileName           = "README.md"
	ServiceConfigExtension   = ".yaml"
)

// Directory names
const (
	DevStackDir = "dev-stack"
	DataDir     = "data"
	LogsDir     = "logs"
	TmpDir      = "tmp"
	ServicesDir = "internal/config/services"
)

// Template file names
const (
	EnvTemplate           = "env.template"
	DockerComposeTemplate = "docker-compose.template"
)

// Configuration URLs
const (
	ConfigDocsURL    = "https://github.com/isaacgarza/dev-stack/tree/main/docs-site/content/configuration.md"
	ServiceConfigURL = "https://github.com/isaacgarza/dev-stack/tree/main/internal/config/services"
)

// Git entries
var GitignoreEntries = []string{
	"",
	"# Dev Stack",
	DevStackDir + "/" + EnvGeneratedFileName,
	DevStackDir + "/" + DataDir + "/",
	DevStackDir + "/" + LogsDir + "/",
	DevStackDir + "/" + TmpDir + "/",
}
