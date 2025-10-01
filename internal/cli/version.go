package cli

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	// These will be set by build flags
	version   = "dev"
	commit    = "unknown"
	date      = "unknown"
	goVersion = runtime.Version()
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long: `Display version information for dev-stack including:
- Version number
- Git commit hash
- Build date
- Go version used to build
- Operating system and architecture

Examples:
  dev-stack version              # Show basic version
  dev-stack version --detailed   # Show detailed version info
  dev-stack version --format json # Output in JSON format`,
	RunE: func(cmd *cobra.Command, args []string) error {
		detailed, _ := cmd.Flags().GetBool("detailed")
		format, _ := cmd.Flags().GetString("format")

		switch format {
		case "json":
			return printVersionJSON(detailed)
		case "yaml":
			return printVersionYAML(detailed)
		default:
			return printVersionText(detailed)
		}
	},
}

func printVersionText(detailed bool) error {
	fmt.Printf("dev-stack version %s\n", version)

	if detailed {
		fmt.Printf("Git commit: %s\n", commit)
		fmt.Printf("Build date: %s\n", date)
		fmt.Printf("Go version: %s\n", goVersion)
		fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	}

	return nil
}

func printVersionJSON(detailed bool) error {
	if detailed {
		fmt.Printf(`{
  "version": "%s",
  "commit": "%s",
  "date": "%s",
  "goVersion": "%s",
  "os": "%s",
  "arch": "%s"
}
`, version, commit, date, goVersion, runtime.GOOS, runtime.GOARCH)
	} else {
		fmt.Printf(`{"version": "%s"}
`, version)
	}

	return nil
}

func printVersionYAML(detailed bool) error {
	fmt.Printf("version: %s\n", version)

	if detailed {
		fmt.Printf("commit: %s\n", commit)
		fmt.Printf("date: %s\n", date)
		fmt.Printf("goVersion: %s\n", goVersion)
		fmt.Printf("os: %s\n", runtime.GOOS)
		fmt.Printf("arch: %s\n", runtime.GOARCH)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(versionCmd)

	// Add flags for version command
	versionCmd.Flags().BoolP("detailed", "d", false, "Show detailed version information")
	versionCmd.Flags().StringP("format", "f", "text", "Output format (text, json, yaml)")
}
