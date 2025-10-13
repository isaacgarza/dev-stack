package constants

// Command names
const (
	CmdNameUp         = "up"
	CmdNameDown       = "down"
	CmdNameRestart    = "restart"
	CmdNameStatus     = "status"
	CmdNameInit       = "init"
	CmdNameDoctor     = "doctor"
	CmdNameCompletion = "completion"
	CmdNameServices   = "services"
	CmdNameDeps       = "deps"
	CmdNameConflicts  = "conflicts"
	CmdNameLogs       = "logs"
	CmdNameExec       = "exec"
	CmdNameConnect    = "connect"
	CmdNameBackup     = "backup"
	CmdNameRestore    = "restore"
	CmdNameCleanup    = "cleanup"
	CmdNameScale      = "scale"
	CmdNameMonitor    = "monitor"
	CmdNameValidate   = "validate"
	CmdNameVersion    = "version"
	CmdNameDocs       = "docs"
)

// Shell types for completion
const (
	ShellBash       = "bash"
	ShellZsh        = "zsh"
	ShellFish       = "fish"
	ShellPowerShell = "powershell"
)

// Docker commands
const (
	DockerCmd        = "docker"
	DockerInfoCmd    = "info"
	DockerComposeCmd = "compose"
	DockerVersionCmd = "version"
)

// URLs
const (
	DockerInstallURL = "https://docs.docker.com/get-docker/"
)

// Command reference builders
func CmdRef(cmdName string) string {
	return AppName + " " + cmdName
}

// Common command references
var (
	CmdUp     = CmdRef(CmdNameUp)
	CmdDown   = CmdRef(CmdNameDown)
	CmdStatus = CmdRef(CmdNameStatus)
	CmdInit   = CmdRef(CmdNameInit)
)

// Error messages
var (
	ErrNotInitialized = AppName + " not initialized. Run '" + CmdInit + "' first"
)
