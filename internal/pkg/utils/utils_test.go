package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestFileExists(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test_file_exists")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	_ = tmpFile.Close()

	// Test file exists
	if !FileExists(tmpFile.Name()) {
		t.Errorf("FileExists should return true for existing file")
	}

	// Test file doesn't exist
	if FileExists("/nonexistent/file") {
		t.Errorf("FileExists should return false for non-existent file")
	}

	// Test directory (should return false)
	tmpDir, err := os.MkdirTemp("", "test_dir")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	if FileExists(tmpDir) {
		t.Errorf("FileExists should return false for directory")
	}
}

func TestDirExists(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "test_dir_exists")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Test directory exists
	if !DirExists(tmpDir) {
		t.Errorf("DirExists should return true for existing directory")
	}

	// Test directory doesn't exist
	if DirExists("/nonexistent/directory") {
		t.Errorf("DirExists should return false for non-existent directory")
	}

	// Test file (should return false)
	tmpFile, err := os.CreateTemp("", "test_file")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	_ = tmpFile.Close()

	if DirExists(tmpFile.Name()) {
		t.Errorf("DirExists should return false for file")
	}
}

func TestEnsureDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_ensure_dir")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Test creating new directory
	newDir := filepath.Join(tmpDir, "new", "nested", "directory")
	if err := EnsureDir(newDir); err != nil {
		t.Errorf("EnsureDir failed: %v", err)
	}

	if !DirExists(newDir) {
		t.Errorf("EnsureDir should create directory")
	}

	// Test existing directory (should not error)
	if err := EnsureDir(newDir); err != nil {
		t.Errorf("EnsureDir should not error on existing directory: %v", err)
	}
}

func TestGenerateRandomString(t *testing.T) {
	length := 16
	str, err := GenerateRandomString(length)
	if err != nil {
		t.Errorf("GenerateRandomString failed: %v", err)
	}

	if len(str) != length {
		t.Errorf("Expected string length %d, got %d", length, len(str))
	}

	// Generate another string and ensure they're different
	str2, err := GenerateRandomString(length)
	if err != nil {
		t.Errorf("GenerateRandomString failed: %v", err)
	}

	if str == str2 {
		t.Errorf("Generated strings should be different")
	}
}

func TestStringInSlice(t *testing.T) {
	slice := []string{"apple", "banana", "cherry"}

	if !StringInSlice("banana", slice) {
		t.Errorf("StringInSlice should return true for existing string")
	}

	if StringInSlice("orange", slice) {
		t.Errorf("StringInSlice should return false for non-existing string")
	}

	if StringInSlice("", slice) {
		t.Errorf("StringInSlice should return false for empty string")
	}
}

func TestRemoveStringFromSlice(t *testing.T) {
	slice := []string{"apple", "banana", "cherry", "banana"}
	result := RemoveStringFromSlice("banana", slice)

	expected := []string{"apple", "cherry"}
	if len(result) != len(expected) {
		t.Errorf("Expected length %d, got %d", len(expected), len(result))
	}

	for i, v := range expected {
		if result[i] != v {
			t.Errorf("Expected %s at index %d, got %s", v, i, result[i])
		}
	}
}

func TestUniqueStrings(t *testing.T) {
	slice := []string{"apple", "banana", "apple", "cherry", "banana"}
	result := UniqueStrings(slice)

	expected := []string{"apple", "banana", "cherry"}
	if len(result) != len(expected) {
		t.Errorf("Expected length %d, got %d", len(expected), len(result))
	}

	for _, v := range expected {
		if !StringInSlice(v, result) {
			t.Errorf("Expected %s to be in result", v)
		}
	}
}

func TestTrimQuotes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello"`, "hello"},
		{"'world'", "world"},
		{"no quotes", "no quotes"},
		{`"single quote'`, `"single quote'`},
		{`'double quote"`, `'double quote"`},
		{`""`, ""},
		{"''", ""},
		{`"`, `"`},
		{"'", "'"},
	}

	for _, test := range tests {
		result := TrimQuotes(test.input)
		if result != test.expected {
			t.Errorf("TrimQuotes(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestSplitAndTrim(t *testing.T) {
	input := "apple, banana , cherry,  ,orange"
	result := SplitAndTrim(input, ",")

	expected := []string{"apple", "banana", "cherry", "orange"}
	if len(result) != len(expected) {
		t.Errorf("Expected length %d, got %d", len(expected), len(result))
	}

	for i, v := range expected {
		if result[i] != v {
			t.Errorf("Expected %s at index %d, got %s", v, i, result[i])
		}
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    uint64
		expected string
	}{
		{0, "0 B"},
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, test := range tests {
		result := FormatBytes(test.bytes)
		if result != test.expected {
			t.Errorf("FormatBytes(%d) = %s, expected %s", test.bytes, result, test.expected)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		contains string
	}{
		{30 * time.Second, "s"},
		{2 * time.Minute, "m"},
		{2 * time.Hour, "h"},
		{25 * time.Hour, "d"},
	}

	for _, test := range tests {
		result := FormatDuration(test.duration)
		if !strings.Contains(result, test.contains) {
			t.Errorf("FormatDuration(%v) = %s, expected to contain %s", test.duration, result, test.contains)
		}
	}
}

func TestIsCommandAvailable(t *testing.T) {
	// Test with a command that should exist on most systems
	if !IsCommandAvailable("go") {
		t.Errorf("IsCommandAvailable should return true for 'go' command")
	}

	// Test with a command that shouldn't exist
	if IsCommandAvailable("nonexistentcommand12345") {
		t.Errorf("IsCommandAvailable should return false for non-existent command")
	}
}

func TestExpandPath(t *testing.T) {
	// Test environment variable expansion
	_ = os.Setenv("TEST_VAR", "test_value")
	defer func() { _ = os.Unsetenv("TEST_VAR") }()

	result := ExpandPath("$TEST_VAR/path")
	expected := "test_value/path"
	if result != expected {
		t.Errorf("ExpandPath($TEST_VAR/path) = %s, expected %s", result, expected)
	}

	// Test ~ expansion (if HOME is set)
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		result := ExpandPath("~/test")
		expected := filepath.Join(home, "test")
		if result != expected {
			t.Errorf("ExpandPath(~/test) = %s, expected %s", result, expected)
		}
	}
}

func TestIsAbsolutePath(t *testing.T) {
	// Test absolute path (platform-specific)
	var absolutePath string
	if runtime.GOOS == "windows" {
		absolutePath = "C:\\absolute\\path"
	} else {
		absolutePath = "/absolute/path"
	}

	if !IsAbsolutePath(absolutePath) {
		t.Errorf("IsAbsolutePath should return true for absolute path: %s", absolutePath)
	}

	// Test relative path
	if IsAbsolutePath("relative/path") {
		t.Errorf("IsAbsolutePath should return false for relative path")
	}

	// Test current directory
	if IsAbsolutePath(".") {
		t.Errorf("IsAbsolutePath should return false for current directory")
	}
}

func TestGetWorkingDir(t *testing.T) {
	wd, err := GetWorkingDir()
	if err != nil {
		t.Errorf("GetWorkingDir failed: %v", err)
	}

	if wd == "" {
		t.Errorf("GetWorkingDir should return non-empty string")
	}

	if !IsAbsolutePath(wd) {
		t.Errorf("GetWorkingDir should return absolute path")
	}
}

func TestRetry(t *testing.T) {
	// Test successful retry
	attempts := 0
	err := Retry(3, 10*time.Millisecond, func() error {
		attempts++
		if attempts < 2 {
			return &testError{"temporary error"}
		}
		return nil
	})

	if err != nil {
		t.Errorf("Retry should succeed after retries: %v", err)
	}

	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}

	// Test failure after all retries
	attempts = 0
	err = Retry(2, 10*time.Millisecond, func() error {
		attempts++
		return &testError{"persistent error"}
	})

	if err == nil {
		t.Errorf("Retry should fail after all attempts")
	}

	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}

func TestTimeout(t *testing.T) {
	// Test successful operation within timeout
	err := Timeout(100*time.Millisecond, func() error {
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	if err != nil {
		t.Errorf("Timeout should succeed for fast operation: %v", err)
	}

	// Test timeout
	err = Timeout(50*time.Millisecond, func() error {
		time.Sleep(100 * time.Millisecond)
		return nil
	})

	if err == nil {
		t.Errorf("Timeout should fail for slow operation")
	}

	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("Error should mention timeout: %v", err)
	}
}

func TestGetProcessPIDWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific test on non-Windows platform")
	}

	// Test with a process that should exist on Windows
	_, err := GetProcessPID("explorer")
	// We don't assert success because the process might not exist in CI
	// but we want to ensure the Windows code path is executed
	if err != nil {
		t.Logf("Expected error for non-existent or inaccessible process: %v", err)
	}
}

func TestGetProcessPIDUnix(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix-specific test on Windows platform")
	}

	// Test with a process that should exist on Unix systems
	_, err := GetProcessPID("kernel")
	// We don't assert success because the process might not exist in CI
	// but we want to ensure the Unix code path is executed
	if err != nil {
		t.Logf("Expected error for non-existent or inaccessible process: %v", err)
	}
}

func TestIsPortInUseWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific test on non-Windows platform")
	}

	// Test a port that's unlikely to be in use
	result := IsPortInUse(65432)
	// We just want to execute the Windows code path
	t.Logf("Port 65432 in use: %v", result)
}

func TestIsPortInUseUnix(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix-specific test on Windows platform")
	}

	// Test a port that's unlikely to be in use
	result := IsPortInUse(65432)
	// We just want to execute the Unix code path
	t.Logf("Port 65432 in use: %v", result)
}

func TestKillProcessWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific test on non-Windows platform")
	}

	// Test with an invalid PID to exercise the Windows code path
	err := KillProcess(999999)
	if err == nil {
		t.Error("Expected error when killing non-existent process")
	}
	t.Logf("Expected error killing invalid PID: %v", err)
}

func TestKillProcessUnix(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix-specific test on Windows platform")
	}

	// Test with an invalid PID to exercise the Unix code path
	err := KillProcess(999999)
	if err == nil {
		t.Error("Expected error when killing non-existent process")
	}
	t.Logf("Expected error killing invalid PID: %v", err)
}

func TestRunCommand(t *testing.T) {
	// Test a simple command that should work on all platforms
	var cmd string
	var args []string
	if runtime.GOOS == "windows" {
		cmd = "cmd"
		args = []string{"/c", "echo", "test"}
	} else {
		cmd = "echo"
		args = []string{"test"}
	}

	output, err := RunCommand(cmd, args...)
	if err != nil {
		t.Errorf("RunCommand failed: %v", err)
	}

	if len(output) == 0 {
		t.Error("RunCommand should return output")
	}

	// Check that output contains expected text
	if !strings.Contains(output, "test") {
		t.Errorf("RunCommand output should contain 'test', got: %s", output)
	}
}

func TestGetFreePort(t *testing.T) {
	port, err := GetFreePort(8000)
	if err != nil {
		t.Errorf("GetFreePort failed: %v", err)
	}

	if port < 8000 || port >= 8100 {
		t.Errorf("GetFreePort returned port %d outside expected range", port)
	}
}

// Helper error type for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
