package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// execCmd represents the exec command
var execCmd = &cobra.Command{
	Use:   "exec <service> <command> [args...]",
	Short: "Execute a command in a running service container",
	Long: `Execute a command in a running service container.
The service must be running for this command to work.

Examples:
  dev-stack exec postgres psql -U postgres    # Connect to PostgreSQL
  dev-stack exec redis redis-cli              # Connect to Redis CLI
  dev-stack exec api bash                     # Open a bash shell in API container
  dev-stack exec --user root api apt update   # Run command as root user`,
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		service := args[0]
		command := args[1:]

		user, _ := cmd.Flags().GetString("user")
		workdir, _ := cmd.Flags().GetString("workdir")
		env, _ := cmd.Flags().GetStringSlice("env")
		detach, _ := cmd.Flags().GetBool("detach")
		interactive, _ := cmd.Flags().GetBool("interactive")
		tty, _ := cmd.Flags().GetBool("tty")

		fmt.Printf("Executing in %s: %v\n", service, command)

		if user != "" {
			fmt.Printf("Running as user: %s\n", user)
		}

		if workdir != "" {
			fmt.Printf("Working directory: %s\n", workdir)
		}

		if len(env) > 0 {
			fmt.Printf("Environment variables: %v\n", env)
		}

		if detach {
			fmt.Println("Running in detached mode...")
		}

		if interactive {
			fmt.Println("Running in interactive mode...")
		}

		if tty {
			fmt.Println("Allocating TTY...")
		}

		// TODO: Implement command execution logic
		// This will integrate with Docker Compose exec functionality
		return nil
	},
}

func init() {
	rootCmd.AddCommand(execCmd)

	// Add flags for exec command
	execCmd.Flags().StringP("user", "u", "", "Run the command as this user")
	execCmd.Flags().StringP("workdir", "w", "", "Path to workdir directory for this command")
	execCmd.Flags().StringSliceP("env", "e", []string{}, "Set environment variables")
	execCmd.Flags().BoolP("detach", "d", false, "Run command in background")
	execCmd.Flags().BoolP("interactive", "i", false, "Keep STDIN open even if not attached")
	execCmd.Flags().BoolP("tty", "t", false, "Allocate a pseudo-TTY")
	execCmd.Flags().Bool("privileged", false, "Give extended privileges to the command")
	execCmd.Flags().String("index", "", "Index of the container if there are multiple instances of a service")
}
