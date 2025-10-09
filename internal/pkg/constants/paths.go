package constants

// File names
const (
	ConfigFileName         = "dev-stack-config.yml"
	DockerComposeFileName  = "docker-compose.yml"
	EnvGeneratedFileName   = ".env.generated"
	GitignoreFileName      = ".gitignore"
	ReadmeFileName         = "README.md"
	ServiceConfigExtension = ".yaml"
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
	ConfigTemplate        = "dev-stack-config.template"
	EnvTemplate           = "env.template"
	DockerComposeTemplate = "docker-compose.template"
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
