package handlers

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// DownHandler handles the "down" command for stopping services
type DownHandler struct{}

// NewDownHandler creates a new down command handler
func NewDownHandler() *DownHandler {
	return &DownHandler{}
}

// Handle executes the down command
func (h *DownHandler) Handle(ctx context.Context, cmd *cobra.Command, args []string, base *BaseCommand) error {
	volumes, _ := cmd.Flags().GetBool("volumes")
	timeout, _ := cmd.Flags().GetInt("timeout")

	// Display what we're stopping
	if len(args) == 0 {
		fmt.Println("⏹️ Stopping all services...")
	} else {
		fmt.Printf("⏹️ Stopping services: %v\n", args)
	}

	// Set up stop options
	stopOptions := StopOptions{
		Timeout: timeout,
		Volumes: volumes,
	}

	// Stop services
	return base.Manager.StopServices(ctx, args, stopOptions)
}

// ValidateArgs validates the command arguments
func (h *DownHandler) ValidateArgs(args []string) error {
	// No specific validation needed for down command
	return nil
}

// GetRequiredFlags returns the required flags for this command
func (h *DownHandler) GetRequiredFlags() []string {
	return []string{} // No required flags for down command
}
