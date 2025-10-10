package constants

// Brand constants for consistent naming
const (
	// Application name variations
	AppName      = "dev-stack" // CLI command name
	AppNameTitle = "Dev Stack" // Title case for headers
	AppNameLower = "dev stack" // Sentence case for messages

	// Common messages
	MsgInitializing = "Initializing " + AppNameTitle
	MsgStarting     = "Starting " + AppNameTitle
	MsgStopping     = "Stopping " + AppNameTitle
	MsgRestarting   = "Restarting " + AppNameTitle
	MsgStatus       = AppNameTitle + " Status"

	// Success messages
	MsgInitSuccess    = AppNameLower + " initialized successfully!"
	MsgStartSuccess   = AppNameLower + " started successfully"
	MsgStopSuccess    = AppNameLower + " stopped successfully"
	MsgRestartSuccess = AppNameLower + " restarted successfully"

	// Command references
	CmdUp     = AppName + " up"
	CmdDown   = AppName + " down"
	CmdStatus = AppName + " status"
	CmdInit   = AppName + " init"

	// Error messages
	ErrNotInitialized = AppName + " not initialized. Run '" + CmdInit + "' first"
)
