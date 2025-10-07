package handlers

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/isaacgarza/dev-stack/internal/pkg/display"
	"github.com/spf13/cobra"
)

// StatusHandler handles the "status" command for checking service status
type StatusHandler struct {
	formatterFactory display.FormatterFactory
}

// NewStatusHandler creates a new status command handler
func NewStatusHandler() *StatusHandler {
	return &StatusHandler{
		formatterFactory: display.NewFactory(),
	}
}

// Handle executes the status command
func (h *StatusHandler) Handle(ctx context.Context, cmd *cobra.Command, args []string, base *BaseCommand) error {
	format, _ := cmd.Flags().GetString("format")
	quiet, _ := cmd.Flags().GetBool("quiet")
	_, _ = cmd.Flags().GetBool("watch")

	if !quiet {
		if len(args) == 0 {
			fmt.Println("📊 Checking status of all services...")
		} else {
			fmt.Printf("📊 Checking status of services: %v\n", args)
		}
	}

	// Get service status
	services, err := base.Manager.GetServiceStatus(ctx, args)
	if err != nil {
		return fmt.Errorf("failed to get service status: %w", err)
	}

	// Convert to display format
	displayServices := h.convertToDisplayFormat(services)

	// Create formatter and display status
	formatter, err := h.formatterFactory.CreateFormatter(format, os.Stdout)
	if err != nil {
		return fmt.Errorf("failed to create formatter: %w", err)
	}

	options := display.StatusOptions{
		Quiet: quiet,
	}

	return formatter.FormatStatus(displayServices, options)
}

// ValidateArgs validates the command arguments
func (h *StatusHandler) ValidateArgs(args []string) error {
	// No specific validation needed for status command
	return nil
}

// GetRequiredFlags returns the required flags for this command
func (h *StatusHandler) GetRequiredFlags() []string {
	return []string{} // No required flags for status command
}

// convertToDisplayFormat converts service status to display format
func (h *StatusHandler) convertToDisplayFormat(services []ServiceStatus) []display.ServiceStatus {
	displayServices := make([]display.ServiceStatus, len(services))
	for i, service := range services {
		// Calculate uptime
		uptime := time.Since(service.CreatedAt)

		displayServices[i] = display.ServiceStatus{
			Name:      service.Name,
			State:     service.State,
			Health:    service.Health,
			Ports:     service.Ports,
			CreatedAt: service.CreatedAt,
			UpdatedAt: service.UpdatedAt,
			Uptime:    uptime,
		}
	}
	return displayServices
}
