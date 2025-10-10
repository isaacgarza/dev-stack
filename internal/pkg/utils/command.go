package utils

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

// OS constants
const (
	OSWindows = "windows"
	OSLinux   = "linux"
	OSDarwin  = "darwin"
)

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

	pidStr := strings.Split(outputStr, "\n")[0]
	return strconv.Atoi(pidStr)
}

// KillProcess kills a process by PID
func KillProcess(pid int) error {
	if runtime.GOOS == OSWindows {
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
