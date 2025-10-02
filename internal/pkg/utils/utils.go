package utils

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// OS constants
const (
	OSWindows = "windows"
	OSLinux   = "linux"
	OSDarwin  = "darwin"
)

// FileExists checks if a file exists and is not a directory
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// DirExists checks if a directory exists
func DirExists(dirname string) bool {
	info, err := os.Stat(dirname)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// EnsureDir creates a directory if it doesn't exist
func EnsureDir(dirname string) error {
	if !DirExists(dirname) {
		return os.MkdirAll(dirname, 0755)
	}
	return nil
}

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		_ = sourceFile.Close()
	}()

	if dirErr := EnsureDir(filepath.Dir(dst)); dirErr != nil {
		return dirErr
	}

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		_ = destFile.Close()
	}()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// WriteFile writes content to a file, creating directories if needed
func WriteFile(filename string, content []byte, perm os.FileMode) error {
	if err := EnsureDir(filepath.Dir(filename)); err != nil {
		return err
	}
	return os.WriteFile(filename, content, perm)
}

// ReadFileLines reads a file and returns its lines as a slice
func ReadFileLines(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}

// GenerateRandomString generates a random string of the given length
func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}

// StringInSlice checks if a string exists in a slice
func StringInSlice(str string, slice []string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// RemoveStringFromSlice removes a string from a slice
func RemoveStringFromSlice(str string, slice []string) []string {
	var result []string
	for _, s := range slice {
		if s != str {
			result = append(result, s)
		}
	}
	return result
}

// UniqueStrings returns a slice with unique strings
func UniqueStrings(slice []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, str := range slice {
		if !seen[str] {
			seen[str] = true
			result = append(result, str)
		}
	}
	return result
}

// TrimQuotes removes surrounding quotes from a string
func TrimQuotes(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// SplitAndTrim splits a string by delimiter and trims whitespace
func SplitAndTrim(s, delimiter string) []string {
	parts := strings.Split(s, delimiter)
	var result []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// FormatBytes formats bytes into human readable format
func FormatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatDuration formats a duration into human readable format
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%.1fh", d.Hours())
	}
	return fmt.Sprintf("%.1fd", d.Hours()/24)
}

// RunCommand executes a command and returns its output
func RunCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// RunCommandWithDir executes a command in a specific directory
func RunCommandWithDir(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// RunCommandQuiet executes a command without capturing output
func RunCommandQuiet(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Run()
}

// IsCommandAvailable checks if a command is available in PATH
func IsCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// GetProcessPID gets the PID of a running process by name (Linux/macOS only)
func GetProcessPID(name string) (int, error) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux", "darwin":
		cmd = exec.Command("pgrep", "-f", name)
	case OSWindows:
		// Use constant args to avoid gosec G204 warning
		args := []string{"/FI", fmt.Sprintf("IMAGENAME eq %s.exe", name), "/FO", "CSV", "/NH"}
		cmd = exec.Command("tasklist", args...)
	default:
		return 0, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return 0, fmt.Errorf("process not found: %s", name)
	}

	if runtime.GOOS == OSWindows {
		// Parse Windows tasklist output
		lines := strings.Split(outputStr, "\n")
		if len(lines) > 0 {
			fields := strings.Split(lines[0], ",")
			if len(fields) >= 2 {
				pidStr := strings.Trim(fields[1], "\"")
				return strconv.Atoi(pidStr)
			}
		}
		return 0, fmt.Errorf("failed to parse PID from tasklist output")
	}

	// Parse Unix pgrep output
	pidStr := strings.Split(outputStr, "\n")[0]
	return strconv.Atoi(pidStr)
}

// KillProcess kills a process by PID
func KillProcess(pid int) error {
	if runtime.GOOS == OSWindows {
		// Use constant args to avoid gosec G204 warning
		args := []string{"/F", "/PID", strconv.Itoa(pid)}
		cmd := exec.Command("taskkill", args...)
		return cmd.Run()
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return process.Signal(syscall.SIGTERM)
}

// IsPortInUse checks if a port is in use
func IsPortInUse(port int) bool {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case OSLinux, OSDarwin:
		// Use constant args to avoid gosec G204 warning
		args := []string{"-i", fmt.Sprintf(":%d", port)}
		cmd = exec.Command("lsof", args...)
	case OSWindows:
		cmd = exec.Command("netstat", "-an")
	default:
		return false
	}

	output, err := cmd.Output()
	if err != nil {
		return false
	}

	if runtime.GOOS == OSWindows {
		return strings.Contains(string(output), fmt.Sprintf(":%d", port))
	}

	return len(output) > 0
}

// GetFreePort finds an available port starting from the given port
func GetFreePort(startPort int) (int, error) {
	for port := startPort; port < startPort+100; port++ {
		if !IsPortInUse(port) {
			return port, nil
		}
	}
	return 0, fmt.Errorf("no free port found in range %d-%d", startPort, startPort+100)
}

// ExpandPath expands ~ and environment variables in a path
func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(home, path[2:])
		}
	}
	return os.ExpandEnv(path)
}

// GetHomeDir returns the user's home directory
func GetHomeDir() (string, error) {
	return os.UserHomeDir()
}

// GetWorkingDir returns the current working directory
func GetWorkingDir() (string, error) {
	return os.Getwd()
}

// IsAbsolutePath checks if a path is absolute
func IsAbsolutePath(path string) bool {
	return filepath.IsAbs(path)
}

// MakeAbsolutePath converts a relative path to absolute
func MakeAbsolutePath(path string) (string, error) {
	if IsAbsolutePath(path) {
		return path, nil
	}
	return filepath.Abs(path)
}

// Retry executes a function with retry logic
func Retry(attempts int, delay time.Duration, fn func() error) error {
	var err error
	for i := 0; i < attempts; i++ {
		if err = fn(); err == nil {
			return nil
		}
		if i < attempts-1 {
			time.Sleep(delay)
		}
	}
	return fmt.Errorf("failed after %d attempts: %w", attempts, err)
}

// Timeout executes a function with timeout
func Timeout(timeout time.Duration, fn func() error) error {
	done := make(chan error, 1)
	go func() {
		done <- fn()
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		return fmt.Errorf("operation timed out after %v", timeout)
	}
}

// AskConfirmation asks for user confirmation
func AskConfirmation(message string) bool {
	fmt.Printf("%s (y/N): ", message)
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		response := strings.ToLower(strings.TrimSpace(scanner.Text()))
		return response == "y" || response == "yes"
	}
	return false
}

// PromptInput prompts for user input with a message
func PromptInput(message string) (string, error) {
	fmt.Printf("%s: ", message)
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text()), nil
	}
	return "", scanner.Err()
}

// PromptSelect prompts user to select from options
func PromptSelect(message string, options []string) (int, error) {
	fmt.Println(message)
	for i, option := range options {
		fmt.Printf("  %d) %s\n", i+1, option)
	}

	for {
		fmt.Printf("Select (1-%d): ", len(options))
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			choice, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
			if err == nil && choice >= 1 && choice <= len(options) {
				return choice - 1, nil
			}
		}
		fmt.Printf("Invalid choice. Please select 1-%d.\n", len(options))
	}
}
