package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// downCmd represents the down command
var downCmd = &cobra.Command{
	Use:   "down [service...]",
	Short: "Stop development stack services",
	Long: `Stop one or more services in the development stack.
If no services are specified, all running services will be stopped.

Examples:
  dev-stack down                  # Stop all services
  dev-stack down postgres redis   # Stop specific services
  dev-stack down --volumes        # Stop services and remove volumes
  dev-stack down --remove-orphans # Remove orphaned containers`,
	RunE: func(cmd *cobra.Command, args []string) error {
		volumes, _ := cmd.Flags().GetBool("volumes")
		removeOrphans, _ := cmd.Flags().GetBool("remove-orphans")
		timeout, _ := cmd.Flags().GetInt("timeout")

		if len(args) == 0 {
			fmt.Println("Stopping all services...")
		} else {
			fmt.Printf("Stopping services: %v\n", args)
		}

		if volumes {
			fmt.Println("Removing volumes...")
		}

		if removeOrphans {
			fmt.Println("Removing orphaned containers...")
		}

		if timeout != 10 {
			fmt.Printf("Using timeout: %d seconds\n", timeout)
		}

		// TODO: Implement service shutdown logic
		// This will integrate with Docker Compose or similar orchestration
		return nil
	},
}

func init() {
	rootCmd.AddCommand(downCmd)

	// Add flags for down command
	downCmd.Flags().BoolP("volumes", "v", false, "Remove named volumes and anonymous volumes")
	downCmd.Flags().Bool("remove-orphans", false, "Remove containers for services not defined in compose file")
	downCmd.Flags().IntP("timeout", "t", 10, "Specify shutdown timeout in seconds")
	downCmd.Flags().Bool("rmi", false, "Remove images (type: 'all' to remove all images)")
}
