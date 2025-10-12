package init

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// setupTestDir creates a temporary directory and changes to it
func setupTestDir(t *testing.T) (cleanup func()) {
	tempDir, err := os.MkdirTemp("", TestTempDirPattern)
	require.NoError(t, err)

	originalDir, err := os.Getwd()
	require.NoError(t, err)

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	return func() {
		os.Chdir(originalDir)
		os.RemoveAll(tempDir)
	}
}

// createTestFile creates a file with given content
func createTestFile(t *testing.T, path, content string) {
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)
}

// createTestConfig creates a dev-stack config file
func createTestConfig(t *testing.T) {
	err := os.MkdirAll("dev-stack", 0755)
	require.NoError(t, err)
	createTestFile(t, TestConfigFilePath, TestConfigContent)
}
