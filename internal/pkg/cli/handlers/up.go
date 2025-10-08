package handlers

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// UpHandler handles the "up" command for starting services
type UpHandler struct{}

// NewUpHandler creates a new up command handler
func NewUpHandler() *UpHandler {
	return &UpHandler{}
}

// Handle executes the up command
func (h *UpHandler) Handle(ctx context.Context, cmd *cobra.Command, args []string, base *BaseCommand) error {
	detach, _ := cmd.Flags().GetBool("detach")
	build, _ := cmd.Flags().GetBool("build")
	profile, _ := cmd.Flags().GetString("profile")

	// Validate services if specified
	if len(args) > 0 {
		if err := base.ValidateServices(args); err != nil {
			return err
		}
	}

	// Display what we're starting
	if len(args) == 0 {
		fmt.Println("🚀 Starting all services...")
	} else {
		fmt.Printf("🚀 Starting services: %v\n", args)
	}

	if profile != "" {
		fmt.Printf("📋 Using profile: %s\n", profile)
	}

	// Set up start options
	startOptions := StartOptions{
		Build:   build,
		Detach:  detach,
		Timeout: 30,
	}

	// Start services
	return base.Manager.StartServices(ctx, args, startOptions)
}

// ValidateArgs validates the command arguments
func (h *UpHandler) ValidateArgs(args []string) error {
	// No specific validation needed for up command
	// Service validation is done in Handle method
	return nil
}

// GetRequiredFlags returns the required flags for this command
func (h *UpHandler) GetRequiredFlags() []string {
	return []string{} // No required flags for up command
}
