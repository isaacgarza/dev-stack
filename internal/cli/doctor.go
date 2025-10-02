package cli

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

// Status constants
const (
	StatusFail = "FAIL"
)

// doctorCmd represents the doctor command
var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run system health checks",
	Long: `Run comprehensive system health checks to ensure dev-stack can operate properly.
This command checks for:
- Required dependencies (Docker, Docker Compose, Git)
- System permissions
- Network connectivity
- Configuration file validity
- Version compatibility
- Available disk space and memory

Examples:
  dev-stack doctor              # Run all health checks
  dev-stack doctor --fix        # Attempt to fix issues automatically
  dev-stack doctor --verbose    # Show detailed output
  dev-stack doctor --check docker # Run specific checks only`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fix, _ := cmd.Flags().GetBool("fix")
		verbose, _ := cmd.Flags().GetBool("verbose")
		check, _ := cmd.Flags().GetString("check")

		if verbose {
			fmt.Println("Running comprehensive system health checks...")
		}

		var checks []HealthCheck
		if check != "" {
			checks = getSpecificChecks(check)
		} else {
			checks = getAllChecks()
		}

		results := runHealthChecks(checks, verbose)
		displayResults(results, verbose)

		if fix {
			fmt.Println("\nAttempting to fix issues...")
			fixIssues(results, verbose)
		}

		// Return error if any critical checks failed
		for _, result := range results {
			if result.Status == StatusFail && result.Critical {
				return fmt.Errorf("critical health checks failed")
			}
		}

		return nil
	},
}

type HealthCheck struct {
	Name        string
	Description string
	Critical    bool
	CheckFunc   func(bool) CheckResult
	FixFunc     func(bool) error
}

type CheckResult struct {
	Name     string
	Status   string // PASS, WARN, FAIL
	Message  string
	Details  string
	Critical bool
	Fixable  bool
}

func getAllChecks() []HealthCheck {
	return []HealthCheck{
		{
			Name:        "docker",
			Description: "Docker daemon availability",
			Critical:    true,
			CheckFunc:   checkDocker,
			FixFunc:     fixDocker,
		},
		{
			Name:        "docker-compose",
			Description: "Docker Compose availability",
			Critical:    true,
			CheckFunc:   checkDockerCompose,
			FixFunc:     fixDockerCompose,
		},
		{
			Name:        "git",
			Description: "Git availability",
			Critical:    false,
			CheckFunc:   checkGit,
			FixFunc:     fixGit,
		},
		{
			Name:        "permissions",
			Description: "Docker permissions",
			Critical:    true,
			CheckFunc:   checkDockerPermissions,
			FixFunc:     fixDockerPermissions,
		},
		{
			Name:        "disk-space",
			Description: "Available disk space",
			Critical:    false,
			CheckFunc:   checkDiskSpace,
			FixFunc:     nil,
		},
		{
			Name:        "memory",
			Description: "Available memory",
			Critical:    false,
			CheckFunc:   checkMemory,
			FixFunc:     nil,
		},
		{
			Name:        "network",
			Description: "Network connectivity",
			Critical:    false,
			CheckFunc:   checkNetwork,
			FixFunc:     nil,
		},
		{
			Name:        "config",
			Description: "Configuration file validity",
			Critical:    false,
			CheckFunc:   checkConfig,
			FixFunc:     fixConfig,
		},
	}
}

func getSpecificChecks(checkName string) []HealthCheck {
	allChecks := getAllChecks()
	for _, check := range allChecks {
		if check.Name == checkName {
			return []HealthCheck{check}
		}
	}
	return []HealthCheck{}
}

func runHealthChecks(checks []HealthCheck, verbose bool) []CheckResult {
	var results []CheckResult

	for _, check := range checks {
		if verbose {
			fmt.Printf("Checking %s... ", check.Description)
		}

		result := check.CheckFunc(verbose)
		result.Critical = check.Critical
		result.Fixable = check.FixFunc != nil

		if verbose {
			fmt.Printf("[%s]\n", result.Status)
			if result.Details != "" {
				fmt.Printf("  %s\n", result.Details)
			}
		}

		results = append(results, result)
	}

	return results
}

func displayResults(results []CheckResult, verbose bool) {
	if !verbose {
		fmt.Println("\nHealth Check Results:")
		fmt.Println("====================")
	}

	passed := 0
	warnings := 0
	failed := 0

	for _, result := range results {
		status := result.Status
		marker := "✓"
		switch result.Status {
		case "WARN":
			marker = "⚠"
			warnings++
		case StatusFail:
			marker = "✗"
			failed++
		default:
			passed++
		}

		if result.Critical && result.Status == StatusFail {
			status += " (CRITICAL)"
		}

		fmt.Printf("%s %s: %s\n", marker, result.Name, status)
		if result.Message != "" {
			fmt.Printf("  %s\n", result.Message)
		}
	}

	fmt.Printf("\nSummary: %d passed, %d warnings, %d failed\n", passed, warnings, failed)
}

func fixIssues(results []CheckResult, verbose bool) {
	// TODO: Implement automatic fixes for common issues
	fmt.Println("Automatic fixes not yet implemented")
}

// Health check implementations
func checkDocker(verbose bool) CheckResult {
	cmd := exec.Command("docker", "version", "--format", "{{.Server.Version}}")
	output, err := cmd.Output()
	if err != nil {
		return CheckResult{
			Name:    "docker",
			Status:  "FAIL",
			Message: "Docker is not installed or not running",
			Details: err.Error(),
		}
	}

	version := strings.TrimSpace(string(output))
	return CheckResult{
		Name:    "docker",
		Status:  "PASS",
		Message: fmt.Sprintf("Docker version %s is running", version),
	}
}

func checkDockerCompose(verbose bool) CheckResult {
	// Try docker compose (new syntax)
	cmd := exec.Command("docker", "compose", "version")
	if err := cmd.Run(); err == nil {
		return CheckResult{
			Name:    "docker-compose",
			Status:  "PASS",
			Message: "Docker Compose v2 is available",
		}
	}

	// Try docker-compose (legacy syntax)
	cmd = exec.Command("docker-compose", "--version")
	if err := cmd.Run(); err == nil {
		return CheckResult{
			Name:    "docker-compose",
			Status:  "PASS",
			Message: "Docker Compose v1 is available",
		}
	}

	return CheckResult{
		Name:    "docker-compose",
		Status:  "FAIL",
		Message: "Docker Compose is not installed",
	}
}

func checkGit(verbose bool) CheckResult {
	cmd := exec.Command("git", "--version")
	output, err := cmd.Output()
	if err != nil {
		return CheckResult{
			Name:    "git",
			Status:  "WARN",
			Message: "Git is not installed (optional for some features)",
		}
	}

	version := strings.TrimSpace(string(output))
	return CheckResult{
		Name:    "git",
		Status:  "PASS",
		Message: version,
	}
}

func checkDockerPermissions(verbose bool) CheckResult {
	cmd := exec.Command("docker", "ps")
	if err := cmd.Run(); err != nil {
		message := "Cannot run Docker commands"
		if runtime.GOOS == "linux" {
			message += " (try adding user to docker group)"
		}
		return CheckResult{
			Name:    "permissions",
			Status:  "FAIL",
			Message: message,
			Details: err.Error(),
		}
	}

	return CheckResult{
		Name:    "permissions",
		Status:  "PASS",
		Message: "Docker permissions are correct",
	}
}

func checkDiskSpace(verbose bool) CheckResult {
	// TODO: Implement disk space checking
	return CheckResult{
		Name:    "disk-space",
		Status:  "PASS",
		Message: "Disk space check not yet implemented",
	}
}

func checkMemory(verbose bool) CheckResult {
	// TODO: Implement memory checking
	return CheckResult{
		Name:    "memory",
		Status:  "PASS",
		Message: "Memory check not yet implemented",
	}
}

func checkNetwork(verbose bool) CheckResult {
	// TODO: Implement network connectivity checking
	return CheckResult{
		Name:    "network",
		Status:  "PASS",
		Message: "Network check not yet implemented",
	}
}

func checkConfig(verbose bool) CheckResult {
	// Check if config file exists and is valid
	if _, err := os.Stat("dev-stack-config.yaml"); err != nil {
		if _, err := os.Stat(".dev-stack.yaml"); err != nil {
			return CheckResult{
				Name:    "config",
				Status:  "WARN",
				Message: "No configuration file found (will use defaults)",
			}
		}
	}

	return CheckResult{
		Name:    "config",
		Status:  "PASS",
		Message: "Configuration file found and readable",
	}
}

// Fix functions (stubs for now)
func fixDocker(verbose bool) error {
	return fmt.Errorf("automatic Docker installation not supported")
}

func fixDockerCompose(verbose bool) error {
	return fmt.Errorf("automatic Docker Compose installation not supported")
}

func fixGit(verbose bool) error {
	return fmt.Errorf("automatic Git installation not supported")
}

func fixDockerPermissions(verbose bool) error {
	if runtime.GOOS == "linux" {
		return fmt.Errorf("please add your user to the docker group: sudo usermod -aG docker $USER")
	}
	return fmt.Errorf("docker permission fix not available for this platform")
}

func fixConfig(verbose bool) error {
	// TODO: Create default config file
	return fmt.Errorf("automatic config creation not yet implemented")
}

func init() {
	rootCmd.AddCommand(doctorCmd)

	// Add flags for doctor command
	doctorCmd.Flags().BoolP("fix", "f", false, "Attempt to fix issues automatically")
	doctorCmd.Flags().BoolP("verbose", "v", false, "Show detailed output")
	doctorCmd.Flags().StringP("check", "c", "", "Run specific check only (docker, git, permissions, etc.)")
	doctorCmd.Flags().Bool("no-color", false, "Disable colored output")
}
