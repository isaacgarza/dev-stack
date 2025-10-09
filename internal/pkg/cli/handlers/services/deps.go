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

// DepsHandler handles the deps command
type DepsHandler struct{}

// NewDepsHandler creates a new deps handler
func NewDepsHandler() *DepsHandler {
	return &DepsHandler{}
}

// Handle executes the deps command
func (h *DepsHandler) Handle(ctx context.Context, cmd *cobra.Command, args []string, base *types.BaseCommand) error {
	ui.Header("Service Dependencies")

	// Get output format
	format, _ := cmd.Flags().GetString("output")

	// Load service dependencies
	serviceUtils := utils.NewServiceUtils()
	dependencies, err := serviceUtils.LoadAllServiceDependencies()
	if err != nil {
		return fmt.Errorf("failed to load dependencies: %w", err)
	}

	if len(dependencies) == 0 {
		ui.Info("No service dependencies found")
		return nil
	}

	// Create display data
	var displayData []map[string]interface{}
	for serviceName, deps := range dependencies {
		if len(deps) == 0 {
			displayData = append(displayData, map[string]interface{}{
				"Service":      serviceName,
				"Dependencies": "None",
			})
		} else {
			for _, dep := range deps {
				displayData = append(displayData, map[string]interface{}{
					"Service":      serviceName,
					"Dependencies": dep,
				})
			}
		}
	}

	// Display results
	formatter, err := display.NewFactory().CreateFormatter(format, cmd.OutOrStdout())
	if err != nil {
		return fmt.Errorf("failed to create formatter: %w", err)
	}

	// Convert to ServiceStatus format for display
	var services []display.ServiceStatus
	for _, item := range displayData {
		services = append(services, display.ServiceStatus{
			Name:  item["Service"].(string),
			State: item["Dependencies"].(string),
		})
	}

	if err := formatter.FormatStatus(services, display.StatusOptions{}); err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	return nil
}

// ValidateArgs validates the command arguments
func (h *DepsHandler) ValidateArgs(args []string) error {
	return nil
}

// GetRequiredFlags returns required flags for this command
func (h *DepsHandler) GetRequiredFlags() []string {
	return []string{}
}
