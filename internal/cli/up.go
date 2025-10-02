package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"dev-stack/internal/core/services"
	"dev-stack/internal/pkg/logger"

	"github.com/spf13/cobra"
)

// upCmd represents the up command
var upCmd = &cobra.Command{
	Use:   "up [service...]",
	Short: "Start development stack services",
	Long: `Start one or more services in the development stack.
If no services are specified, all configured services will be started.

Examples:
  dev-stack up                    # Start all services
  dev-stack up postgres redis     # Start specific services
  dev-stack up --detach          # Start services in background`,
	RunE: func(cmd *cobra.Command, args []string) error {
		detach, _ := cmd.Flags().GetBool("detach")
		build, _ := cmd.Flags().GetBool("build")
		profile, _ := cmd.Flags().GetString("profile")
		noDeps, _ := cmd.Flags().GetBool("no-deps")
		forceRecreate, _ := cmd.Flags().GetBool("force-recreate")

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
		defer func() {
			if closeErr := manager.Close(); closeErr != nil {
				log.Warn("failed to close manager", "error", closeErr)
			}
		}()

		// Display what we're starting
		if len(args) == 0 {
			fmt.Println("🚀 Starting all services...")
		} else {
			fmt.Printf("🚀 Starting services: %v\n", args)
		}

		if profile != "" {
			fmt.Printf("📋 Using profile: %s\n", profile)
		}

		if build {
			fmt.Println("🔨 Building services before starting...")
		}

		if detach {
			fmt.Println("🔄 Running in detached mode...")
		}

		// Set up start options
		startOptions := services.StartOptions{
			Build:         build,
			ForceRecreate: forceRecreate,
			NoDeps:        noDeps,
			Detach:        detach,
			Timeout:       30 * time.Second,
		}

		// Start services
		ctx := context.Background()
		if err := manager.StartServices(ctx, args, startOptions); err != nil {
			return fmt.Errorf("failed to start services: %w", err)
		}

		fmt.Println("✅ Services started successfully")

		if !detach {
			// Show service status
			fmt.Println("\n📊 Service Status:")
			services, err := manager.GetServiceStatus(ctx, args)
			if err != nil {
				fmt.Printf("Warning: failed to get service status: %v\n", err)
			} else {
				for _, service := range services {
					status := "❌"
					if service.State == "running" {
						switch service.Health {
						case "healthy":
							status = "✅"
						case "starting":
							status = "🟡"
						default:
							status = "🟢"
						}
					}
					fmt.Printf("  %s %s (%s)\n", status, service.Name, service.State)
				}
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(upCmd)

	// Add flags for up command
	upCmd.Flags().BoolP("detach", "d", false, "Run services in background")
	upCmd.Flags().BoolP("build", "b", false, "Build images before starting")
	upCmd.Flags().StringP("profile", "p", "", "Specify a profile to use")
	upCmd.Flags().Bool("no-deps", false, "Don't start linked services")
	upCmd.Flags().Bool("force-recreate", false, "Recreate containers even if config hasn't changed")
}
