//go:build integration
// +build integration

package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestDevStackBinaryExists(t *testing.T) {
	// Get the project root directory
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Navigate to project root (assuming tests/integration is two levels deep)
	projectRoot := filepath.Join(wd, "..", "..")
	binaryPath := filepath.Join(projectRoot, "build", "dev-stack")

	// Check if binary exists
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Fatalf("dev-stack binary does not exist at %s", binaryPath)
	}

	t.Logf("✓ dev-stack binary found at %s", binaryPath)
}

func TestDevStackVersionCommand(t *testing.T) {
	// Get the project root directory
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Navigate to project root
	projectRoot := filepath.Join(wd, "..", "..")
	binaryPath := filepath.Join(projectRoot, "build", "dev-stack")

	// Skip test if binary doesn't exist
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Skip("dev-stack binary not found, skipping version test")
	}

	// Run version command
	cmd := exec.Command(binaryPath, "version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run version command: %v\nOutput: %s", err, output)
	}

	if len(output) == 0 {
		t.Fatal("Version command returned empty output")
	}

	t.Logf("✓ Version command output: %s", string(output))
}

func TestDevStackHelpCommand(t *testing.T) {
	// Get the project root directory
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Navigate to project root
	projectRoot := filepath.Join(wd, "..", "..")
	binaryPath := filepath.Join(projectRoot, "build", "dev-stack")

	// Skip test if binary doesn't exist
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Skip("dev-stack binary not found, skipping help test")
	}

	// Run help command
	cmd := exec.Command(binaryPath, "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run help command: %v\nOutput: %s", err, output)
	}

	if len(output) == 0 {
		t.Fatal("Help command returned empty output")
	}

	// Check for common help indicators
	outputStr := string(output)
	if !contains(outputStr, "Usage") && !contains(outputStr, "Commands") && !contains(outputStr, "help") {
		t.Fatalf("Help output doesn't contain expected help content: %s", outputStr)
	}

	t.Logf("✓ Help command executed successfully")
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
