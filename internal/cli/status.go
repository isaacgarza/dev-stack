package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"dev-stack/internal/core/services"
	"dev-stack/internal/pkg/logger"

	"dev-stack/internal/pkg/types"

	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status [service...]",
	Short: "Show status of development stack services",
	Long: `Show the current status of one or more services in the development stack.
If no services are specified, status for all configured services will be shown.

The status includes:
- Running state (up/down/starting/stopping)
- Container health (if health checks are configured)
- Port mappings
- Resource usage (CPU, memory)
- Last restart time

Examples:
  dev-stack status                    # Show status of all services
  dev-stack status postgres redis     # Show status of specific services
  dev-stack status --format json      # Output in JSON format
  dev-stack status --watch           # Watch for status changes`,
	RunE: func(cmd *cobra.Command, args []string) error {
		format, _ := cmd.Flags().GetString("format")
		watch, _ := cmd.Flags().GetBool("watch")
		quiet, _ := cmd.Flags().GetBool("quiet")
		filter, _ := cmd.Flags().GetString("filter")

		// Get current working directory
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		// Create logger
		log := logger.New(slog.LevelInfo)

		// Create service manager
		manager, err := services.NewManager(log, cwd)
		if err != nil {
			return fmt.Errorf("failed to create service manager: %w", err)
		}
		defer manager.Close()

		if len(args) == 0 {
			if !quiet {
				fmt.Println("📊 Checking status of all services...")
			}
		} else {
			if !quiet {
				fmt.Printf("📊 Checking status of services: %v\n", args)
			}
		}

		// Function to get and display status
		displayStatus := func() error {
			ctx := context.Background()
			services, err := manager.GetServiceStatus(ctx, args)
			if err != nil {
				return fmt.Errorf("failed to get service status: %w", err)
			}

			// Apply filter if specified
			if filter != "" {
				var filteredServices []types.ServiceStatus
				for _, service := range services {
					if strings.Contains(strings.ToLower(service.State), strings.ToLower(filter)) {
						filteredServices = append(filteredServices, service)
					}
				}
				services = filteredServices
			}

			switch format {
			case "json":
				return displayStatusJSON(services)
			case "yaml":
				return displayStatusYAML(services)
			case "table":
				fallthrough
			default:
				return displayStatusTable(services, quiet)
			}
		}

		// Display status once
		if err := displayStatus(); err != nil {
			return err
		}

		// Watch for changes if requested
		if watch {
			if !quiet {
				fmt.Println("\n🔄 Watching for status changes... (Press Ctrl+C to stop)")
			}
			for {
				time.Sleep(2 * time.Second)
				fmt.Print("\033[2J\033[H") // Clear screen and move cursor to top
				if err := displayStatus(); err != nil {
					fmt.Printf("Error getting status: %v\n", err)
				}
			}
		}

		return nil
	},
}

// Helper functions for status display

func displayStatusTable(services []types.ServiceStatus, quiet bool) error {
	if len(services) == 0 {
		fmt.Println("No services found")
		return nil
	}

	if quiet {
		for _, service := range services {
			fmt.Printf("%s: %s\n", service.Name, service.State)
		}
		return nil
	}

	// Display header
	fmt.Printf("%-20s %-10s %-12s %-10s %-15s %-10s\n",
		"SERVICE", "STATE", "HEALTH", "CPU %", "MEMORY", "UPTIME")
	fmt.Println(strings.Repeat("-", 80))

	// Display services
	for _, service := range services {
		state := getStateIcon(service.State) + " " + service.State
		health := getHealthIcon(service.Health) + " " + service.Health

		var cpuStr, memStr, uptimeStr string
		if service.State == "running" {
			cpuStr = fmt.Sprintf("%.1f%%", service.CPUUsage)
			if service.Memory.Limit > 0 {
				memStr = fmt.Sprintf("%.1fMB", float64(service.Memory.Used)/1024/1024)
			} else {
				memStr = "N/A"
			}
			if service.Uptime > 0 {
				uptimeStr = formatDuration(service.Uptime)
			}
		} else {
			cpuStr = "-"
			memStr = "-"
			uptimeStr = "-"
		}

		fmt.Printf("%-20s %-10s %-12s %-10s %-15s %-10s\n",
			service.Name, state, health, cpuStr, memStr, uptimeStr)
	}

	return nil
}

func displayStatusJSON(services []types.ServiceStatus) error {
	// TODO: Implement JSON output
	fmt.Println("JSON output not yet implemented")
	return nil
}

func displayStatusYAML(services []types.ServiceStatus) error {
	// TODO: Implement YAML output
	fmt.Println("YAML output not yet implemented")
	return nil
}

func getStateIcon(state string) string {
	switch state {
	case "running":
		return "🟢"
	case "stopped", "exited":
		return "🔴"
	case "starting":
		return "🟡"
	case "stopping":
		return "🟠"
	default:
		return "⚪"
	}
}

func getHealthIcon(health string) string {
	switch health {
	case "healthy":
		return "✅"
	case "unhealthy":
		return "❌"
	case "starting":
		return "🟡"
	default:
		return "➖"
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	} else if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	} else {
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}

func init() {
	rootCmd.AddCommand(statusCmd)

	// Add flags for status command
	statusCmd.Flags().StringP("format", "f", "table", "Output format (table, json, yaml)")
	statusCmd.Flags().BoolP("quiet", "q", false, "Only show service names and status")
	statusCmd.Flags().BoolP("watch", "w", false, "Watch for status changes")
	statusCmd.Flags().Bool("no-trunc", false, "Don't truncate output")
	statusCmd.Flags().StringP("filter", "", "", "Filter services by status (running, stopped, etc.)")
}
