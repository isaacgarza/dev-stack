package doctor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/isaacgarza/dev-stack/internal/pkg/cli/types"
	"github.com/isaacgarza/dev-stack/internal/pkg/constants"
	"github.com/isaacgarza/dev-stack/internal/pkg/ui"
	"github.com/spf13/cobra"
)

type DoctorHandler struct {
	output *ui.Output
}

func NewDoctorHandler() *DoctorHandler {
	return &DoctorHandler{
		output: ui.NewOutput(),
	}
}

func (h *DoctorHandler) ValidateArgs(args []string) error {
	return nil
}

func (h *DoctorHandler) GetRequiredFlags() []string {
	return []string{}
}

func (h *DoctorHandler) Handle(ctx context.Context, cmd *cobra.Command, args []string, base *types.BaseCommand) error {
	h.output.Header("🩺 " + constants.AppNameTitle + " Health Check")

	allGood := true &&
		h.checkDocker() &&
		h.checkDockerCompose() &&
		h.checkProjectInit() &&
		h.checkConfiguration()

	if allGood {
		h.output.Success("All checks passed! Your %s is healthy.", constants.AppNameLower)
		return nil
	} else {
		h.output.Error("Some issues found. Please address them above.")
		return fmt.Errorf("health check failed")
	}
}

func (h *DoctorHandler) checkDocker() bool {
	h.output.Info("Checking Docker installation...")

	if !h.isCommandAvailable(constants.DockerCmd) {
		h.output.Error("Docker not found")
		h.output.Muted("Install Docker: %s", constants.DockerInstallURL)
		return false
	}

	// Check if Docker daemon is running
	cmd := exec.Command(constants.DockerCmd, constants.DockerInfoCmd)
	if err := cmd.Run(); err != nil {
		h.output.Error("Docker daemon not running")
		h.output.Muted("Start Docker daemon")
		return false
	}

	h.output.Success("Docker is available and running")
	return true
}

func (h *DoctorHandler) checkDockerCompose() bool {
	h.output.Info("Checking Docker Compose...")

	if !h.hasDockerComposePlugin() {
		h.output.Error("Docker Compose not found")
		h.output.Muted("Docker Compose is now integrated into Docker CLI")
		h.output.Muted("Update Docker to get 'docker compose' command")
		return false
	}

	h.output.Success("Docker Compose is available")
	return true
}

func (h *DoctorHandler) checkProjectInit() bool {
	h.output.Info("Checking project initialization...")

	configPath := filepath.Join(constants.DevStackDir, constants.ConfigFileName)
	configPathYAML := filepath.Join(constants.DevStackDir, constants.ConfigFileNameYAML)

	if _, err := os.Stat(configPath); err != nil {
		if _, err := os.Stat(configPathYAML); err != nil {
			h.output.Error("Project not initialized")
			h.output.Muted("Run '%s' to initialize", constants.CmdInit)
			return false
		}
	}

	h.output.Success("Project is initialized")
	return true
}

func (h *DoctorHandler) checkConfiguration() bool {
	h.output.Info("Checking configuration validity...")

	// Check if dev-stack directory exists
	if _, err := os.Stat(constants.DevStackDir); os.IsNotExist(err) {
		h.output.Error("Configuration directory missing")
		h.output.Muted("Run '%s' to initialize", constants.CmdInit)
		return false
	}

	// Check if docker-compose file exists
	composePath := filepath.Join(constants.DevStackDir, constants.DockerComposeFileName)
	if _, err := os.Stat(composePath); os.IsNotExist(err) {
		h.output.Error("Docker compose file missing")
		h.output.Muted("Configuration is incomplete")
		return false
	}

	h.output.Success("Configuration is valid")
	return true
}

func (h *DoctorHandler) isCommandAvailable(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

func (h *DoctorHandler) hasDockerComposePlugin() bool {
	cmd := exec.Command(constants.DockerCmd, constants.DockerComposeCmd, constants.DockerVersionCmd)
	return cmd.Run() == nil
}
