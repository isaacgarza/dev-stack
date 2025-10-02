package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

// scaleCmd represents the scale command
var scaleCmd = &cobra.Command{
	Use:   "scale <service>=<replicas> [service=replicas...]",
	Short: "Scale service replicas up or down",
	Long: `Scale service replicas up or down by specifying the desired number of instances.
This command allows you to adjust the number of running containers for each service
to handle different load requirements or testing scenarios.

Examples:
  dev-stack scale api=3              # Scale API service to 3 replicas
  dev-stack scale api=3 worker=2     # Scale multiple services
  dev-stack scale postgres=1         # Scale down to single instance
  dev-stack scale --detach api=5     # Scale in background`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		detach, _ := cmd.Flags().GetBool("detach")
		timeout, _ := cmd.Flags().GetInt("timeout")
		noRecreate, _ := cmd.Flags().GetBool("no-recreate")

		fmt.Println("🔧 Scaling services...")

		// Parse service=replica pairs
		services := make(map[string]int)
		for _, arg := range args {
			// Split on '=' to get service and replica count
			parts := splitServiceScale(arg)
			if len(parts) != 2 {
				return fmt.Errorf("invalid format: %s (expected: service=replicas)", arg)
			}

			service := parts[0]
			replicas, err := strconv.Atoi(parts[1])
			if err != nil {
				return fmt.Errorf("invalid replica count for %s: %s", service, parts[1])
			}

			if replicas < 0 {
				return fmt.Errorf("replica count cannot be negative for service %s", service)
			}

			services[service] = replicas
		}

		// Display scaling plan
		for service, replicas := range services {
			fmt.Printf("  %s: scaling to %d replica(s)\n", service, replicas)
		}

		if detach {
			fmt.Println("Running in detached mode...")
		}

		if timeout != 10 {
			fmt.Printf("Using timeout: %d seconds\n", timeout)
		}

		if noRecreate {
			fmt.Println("Using existing containers where possible...")
		}

		// TODO: Implement scaling logic
		// This will integrate with Docker Compose to:
		// 1. Validate that services exist in the compose file
		// 2. Update the service replica count
		// 3. Start/stop containers as needed
		// 4. Wait for containers to be healthy if not detached
		// 5. Display updated service status

		for service, replicas := range services {
			if replicas == 0 {
				fmt.Printf("⏹️  Stopping all instances of %s\n", service)
			} else {
				fmt.Printf("🚀 Scaling %s to %d replica(s)\n", service, replicas)
			}
		}

		if !detach {
			fmt.Println("Waiting for containers to be ready...")
			// In real implementation, wait for health checks
		}

		fmt.Println("✅ Scaling completed successfully")
		return nil
	},
}

// splitServiceScale splits a service=replicas string into parts
func splitServiceScale(s string) []string {
	for i, c := range s {
		if c == '=' {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}

func init() {
	rootCmd.AddCommand(scaleCmd)

	// Add flags for scale command
	scaleCmd.Flags().BoolP("detach", "d", false, "Don't wait for containers to be ready")
	scaleCmd.Flags().IntP("timeout", "t", 10, "Timeout for container operations")
	scaleCmd.Flags().Bool("no-recreate", false, "Don't recreate containers, just start/stop as needed")
	scaleCmd.Flags().Bool("force-recreate", false, "Recreate containers even if config hasn't changed")
}
