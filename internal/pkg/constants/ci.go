package constants

// Exit codes
const (
	ExitSuccess = 0
	ExitError   = 1
)

// Standard flag names (following cobra/viper conventions)
const (
	FlagQuiet          = "quiet"
	FlagJSON           = "json"
	FlagNoColor        = "no-color"
	FlagNonInteractive = "non-interactive"
	FlagStrict         = "strict"
)
