package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init [project-type]",
	Short: "Initialize a new development stack project",
	Long: `Initialize a new development stack project with the specified type.
Available project types will include common development stacks like:
- go: Go application with Docker
- node: Node.js application with Docker
- python: Python application with Docker
- fullstack: Full-stack application with multiple services`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectType := "default"
		if len(args) > 0 {
			projectType = args[0]
		}

		fmt.Printf("Initializing %s project...\n", projectType)
		// TODO: Implement project initialization logic
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Add flags for init command
	initCmd.Flags().StringP("name", "n", "", "Project name")
	initCmd.Flags().StringP("dir", "d", ".", "Target directory")
	initCmd.Flags().BoolP("force", "f", false, "Force overwrite existing files")
}
