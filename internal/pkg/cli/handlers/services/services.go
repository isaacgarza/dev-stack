package services

import (
	"context"
	"fmt"

	"github.com/isaacgarza/dev-stack/internal/pkg/cli/handlers/utils"
	"github.com/isaacgarza/dev-stack/internal/pkg/cli/types"
	"github.com/isaacgarza/dev-stack/internal/pkg/display"
	"github.com/isaacgarza/dev-stack/internal/pkg/ui"
	"github.com/spf13/cobra"
)

// ServicesHandler handles the services command
type ServicesHandler struct{}

// NewServicesHandler creates a new services handler
func NewServicesHandler() *ServicesHandler {
	return &ServicesHandler{}
}

// Handle executes the services command
func (h *ServicesHandler) Handle(ctx context.Context, cmd *cobra.Command, args []string, base *types.BaseCommand) error {
	ui.Header("Available Services")

	// Get output format
	format, _ := cmd.Flags().GetString("output")

	// Load services by category
	serviceUtils := utils.NewServiceUtils()
	categories, err := serviceUtils.GetServicesByCategory()
	if err != nil {
		return fmt.Errorf("failed to load services: %w", err)
	}

	if len(categories) == 0 {
		ui.Info("No services available")
		return nil
	}

	// Create display data
	var displayData []map[string]interface{}
	for categoryName, services := range categories {
		for _, service := range services {
			displayData = append(displayData, map[string]interface{}{
				"Category":    categoryName,
				"Name":        service.Name,
				"Description": service.Description,
			})
		}
	}

	// Display results
	formatter, err := display.CreateFormatter(format, cmd.OutOrStdout())
	if err != nil {
		return fmt.Errorf("failed to create formatter: %w", err)
	}

	// Convert to ServiceStatus format for display
	var services []display.ServiceStatus
	for _, item := range displayData {
		services = append(services, display.ServiceStatus{
			Name:  item["Name"].(string),
			State: item["Description"].(string),
		})
	}

	if err := formatter.FormatStatus(services, display.StatusOptions{}); err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	return nil
}

// ValidateArgs validates the command arguments
func (h *ServicesHandler) ValidateArgs(args []string) error {
	return nil
}

// GetRequiredFlags returns required flags for this command
func (h *ServicesHandler) GetRequiredFlags() []string {
	return []string{}
}
