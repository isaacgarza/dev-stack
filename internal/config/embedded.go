package config

import (
	"embed"
)

//go:embed commands.yaml
var EmbeddedCommandsYAML []byte

//go:embed init-settings.yaml
var EmbeddedInitSettingsYAML []byte

//go:embed docker-compose.template
var EmbeddedDockerComposeTemplate []byte

//go:embed env.template
var EmbeddedEnvTemplate []byte

//go:embed services
var EmbeddedServicesFS embed.FS
