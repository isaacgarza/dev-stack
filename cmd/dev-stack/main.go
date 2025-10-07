package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/isaacgarza/dev-stack/internal/cli"
	versionpkg "github.com/isaacgarza/dev-stack/internal/pkg/version"
)

func main() {
	// Check if we should delegate to a different version
	if shouldDelegate() {
		if err := delegateToCorrectVersion(); err != nil {
			fmt.Fprintf(os.Stderr, "Version delegation error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Execute normal CLI
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// shouldDelegate checks if we should delegate to a different version
func shouldDelegate() bool {
	// Skip delegation if this is a version management command
	if len(os.Args) >= 2 {
		cmd := os.Args[1]
		if cmd == "versions" || cmd == "version" {
			return false
		}
	}

	// Check if we have version requirements in the current project
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	installDir := filepath.Join(homeDir, ".dev-stack")
	configDir := filepath.Join(homeDir, ".config", "dev-stack")

	vm := versionpkg.NewDefaultVersionManager(installDir, configDir)
	switcher := versionpkg.NewVersionSwitcher(vm)

	shouldDelegate, _, err := switcher.ShouldDelegate(os.Args)
	return err == nil && shouldDelegate
}

// delegateToCorrectVersion delegates execution to the correct version
func delegateToCorrectVersion() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	installDir := filepath.Join(homeDir, ".dev-stack")
	configDir := filepath.Join(homeDir, ".config", "dev-stack")

	vm := versionpkg.NewDefaultVersionManager(installDir, configDir)
	switcher := versionpkg.NewVersionSwitcher(vm)

	return switcher.DelegateToCorrectVersion(os.Args)
}
