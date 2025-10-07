package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	versionpkg "github.com/isaacgarza/dev-stack/internal/pkg/version"
	"github.com/spf13/cobra"
)

var (
	versionsCmd = &cobra.Command{
		Use:   "versions",
		Short: "Manage dev-stack versions",
		Long: `Manage different versions of dev-stack for different projects.
This allows you to have different versions of dev-stack for different projects
with automatic version switching based on project requirements.`,
	}

	versionListCmd = &cobra.Command{
		Use:   "list",
		Short: "List installed versions",
		Long:  `List all installed versions of dev-stack on this system.`,
		RunE:  runVersionList,
	}

	versionInstallCmd = &cobra.Command{
		Use:   "install <version>",
		Short: "Install a specific version",
		Long: `Install a specific version of dev-stack.
Version can be a semantic version like "1.2.3" or "latest".`,
		Args: cobra.ExactArgs(1),
		RunE: runVersionInstall,
	}

	versionUninstallCmd = &cobra.Command{
		Use:   "uninstall <version>",
		Short: "Uninstall a specific version",
		Long:  `Uninstall a specific version of dev-stack.`,
		Args:  cobra.ExactArgs(1),
		RunE:  runVersionUninstall,
	}

	versionUseCmd = &cobra.Command{
		Use:   "use <version>",
		Short: "Switch to a specific version",
		Long: `Switch to use a specific version of dev-stack.
This sets the global default version.`,
		Args: cobra.ExactArgs(1),
		RunE: runVersionUse,
	}

	versionAvailableCmd = &cobra.Command{
		Use:   "available",
		Short: "List available versions",
		Long:  `List all available versions that can be installed from GitHub releases.`,
		RunE:  runVersionAvailable,
	}

	versionDetectCmd = &cobra.Command{
		Use:   "detect [path]",
		Short: "Detect version requirements for a project",
		Long: `Detect version requirements for a project by reading version files.
If no path is provided, uses the current directory.`,
		RunE: runVersionDetect,
	}

	versionSetCmd = &cobra.Command{
		Use:   "set <version> [path]",
		Short: "Set version requirement for a project",
		Long: `Set the version requirement for a project by creating a version file.
If no path is provided, uses the current directory.`,
		Args: cobra.RangeArgs(1, 2),
		RunE: runVersionSet,
	}

	versionCleanupCmd = &cobra.Command{
		Use:   "cleanup",
		Short: "Clean up old versions",
		Long:  `Remove old versions to free up disk space. Keeps the most recent versions and the active version.`,
		RunE:  runVersionCleanup,
	}
)

func init() {
	rootCmd.AddCommand(versionsCmd)

	// Add subcommands
	versionsCmd.AddCommand(versionListCmd)
	versionsCmd.AddCommand(versionInstallCmd)
	versionsCmd.AddCommand(versionUninstallCmd)
	versionsCmd.AddCommand(versionUseCmd)
	versionsCmd.AddCommand(versionAvailableCmd)
	versionsCmd.AddCommand(versionDetectCmd)
	versionsCmd.AddCommand(versionSetCmd)
	versionsCmd.AddCommand(versionCleanupCmd)

	// Add flags
	versionListCmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	versionAvailableCmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	versionAvailableCmd.Flags().IntP("limit", "l", 20, "Limit number of versions to show")
	versionDetectCmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	versionSetCmd.Flags().StringP("format", "f", "text", "Format for version file (text, yaml)")
	versionCleanupCmd.Flags().IntP("keep", "k", 3, "Number of versions to keep")
	versionCleanupCmd.Flags().BoolP("dry-run", "n", false, "Show what would be cleaned up without actually doing it")
}

func getVersionManager() (*versionpkg.DefaultVersionManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	installDir := filepath.Join(homeDir, ".dev-stack")
	configDir := filepath.Join(homeDir, ".config", "dev-stack")

	return versionpkg.NewDefaultVersionManager(installDir, configDir), nil
}

func runVersionList(cmd *cobra.Command, args []string) error {
	vm, err := getVersionManager()
	if err != nil {
		return err
	}

	installed, err := vm.ListInstalledVersions()
	if err != nil {
		return fmt.Errorf("failed to list installed versions: %w", err)
	}

	jsonOutput, _ := cmd.Flags().GetBool("json")

	if jsonOutput {
		return printVersionsJSON(installed)
	}

	if len(installed) == 0 {
		fmt.Println("No versions installed.")
		fmt.Println("Run 'dev-stack versions install latest' to install the latest version.")
		return nil
	}

	// Create table writer
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "VERSION\tACTIVE\tINSTALLED\tSOURCE")

	for _, v := range installed {
		active := ""
		if v.Active {
			active = "*"
		}

		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			v.Version.String(),
			active,
			v.InstallDate.Format("2006-01-02"),
			v.Source,
		)
	}

	return w.Flush()
}

func runVersionInstall(cmd *cobra.Command, args []string) error {
	versionStr := args[0]

	vm, err := getVersionManager()
	if err != nil {
		return err
	}

	// Handle "latest" version
	if versionStr == "latest" {
		available, err := vm.ListAvailableVersions()
		if err != nil {
			return fmt.Errorf("failed to fetch available versions: %w", err)
		}

		if len(available) == 0 {
			return fmt.Errorf("no available versions found")
		}

		// Get the latest version
		latest := available[0]
		for _, v := range available {
			if v.Compare(latest) > 0 {
				latest = v
			}
		}
		versionStr = latest.String()
	}

	// Parse version
	ver, err := versionpkg.ParseVersion(versionStr)
	if err != nil {
		return fmt.Errorf("invalid version: %w", err)
	}

	fmt.Printf("Installing dev-stack version %s...\n", ver.String())

	if err := vm.InstallVersion(*ver); err != nil {
		return fmt.Errorf("failed to install version %s: %w", ver.String(), err)
	}

	fmt.Printf("Successfully installed dev-stack version %s\n", ver.String())
	return nil
}

func runVersionUninstall(cmd *cobra.Command, args []string) error {
	versionStr := args[0]

	vm, err := getVersionManager()
	if err != nil {
		return err
	}

	ver, err := versionpkg.ParseVersion(versionStr)
	if err != nil {
		return fmt.Errorf("invalid version: %w", err)
	}

	// Check if this is the active version
	active, err := vm.GetActiveVersion()
	if err == nil && active.Version.Compare(*ver) == 0 {
		return fmt.Errorf("cannot uninstall active version %s", ver.String())
	}

	fmt.Printf("Uninstalling dev-stack version %s...\n", ver.String())

	if err := vm.UninstallVersion(*ver); err != nil {
		return fmt.Errorf("failed to uninstall version %s: %w", ver.String(), err)
	}

	fmt.Printf("Successfully uninstalled dev-stack version %s\n", ver.String())
	return nil
}

func runVersionUse(cmd *cobra.Command, args []string) error {
	versionStr := args[0]

	vm, err := getVersionManager()
	if err != nil {
		return err
	}

	ver, err := versionpkg.ParseVersion(versionStr)
	if err != nil {
		return fmt.Errorf("invalid version: %w", err)
	}

	// Check if version is installed
	if err := vm.VerifyVersion(*ver); err != nil {
		return fmt.Errorf("version %s is not installed: %w", ver.String(), err)
	}

	if err := vm.SetActiveVersion(*ver); err != nil {
		return fmt.Errorf("failed to set active version: %w", err)
	}

	fmt.Printf("Switched to dev-stack version %s\n", ver.String())
	return nil
}

func runVersionAvailable(cmd *cobra.Command, args []string) error {
	vm, err := getVersionManager()
	if err != nil {
		return err
	}

	fmt.Println("Fetching available versions...")

	available, err := vm.ListAvailableVersions()
	if err != nil {
		return fmt.Errorf("failed to list available versions: %w", err)
	}

	jsonOutput, _ := cmd.Flags().GetBool("json")
	limit, _ := cmd.Flags().GetInt("limit")

	if jsonOutput {
		return printAvailableVersionsJSON(available, limit)
	}

	if len(available) == 0 {
		fmt.Println("No available versions found.")
		return nil
	}

	// Sort versions (latest first)
	for i := 0; i < len(available); i++ {
		for j := i + 1; j < len(available); j++ {
			if available[i].Compare(available[j]) < 0 {
				available[i], available[j] = available[j], available[i]
			}
		}
	}

	// Apply limit
	if limit > 0 && len(available) > limit {
		available = available[:limit]
	}

	fmt.Printf("Available versions (showing %d):\n", len(available))
	for _, v := range available {
		fmt.Printf("  %s\n", v.String())
	}

	return nil
}

func runVersionDetect(cmd *cobra.Command, args []string) error {
	projectPath := "."
	if len(args) > 0 {
		projectPath = args[0]
	}

	vm, err := getVersionManager()
	if err != nil {
		return err
	}

	constraint, err := vm.DetectProjectVersion(projectPath)
	if err != nil {
		return fmt.Errorf("failed to detect project version: %w", err)
	}

	jsonOutput, _ := cmd.Flags().GetBool("json")

	if jsonOutput {
		return printConstraintJSON(constraint)
	}

	if constraint.Original == "*" {
		fmt.Printf("No specific version requirement found for project: %s\n", projectPath)
		fmt.Println("Any version of dev-stack can be used.")
	} else {
		fmt.Printf("Project: %s\n", projectPath)
		fmt.Printf("Required version: %s\n", constraint.Original)
	}

	// Try to resolve to an installed version
	resolved, err := vm.ResolveVersion(*constraint)
	if err != nil {
		fmt.Printf("No installed version satisfies the requirement.\n")
		fmt.Printf("Run 'dev-stack versions install %s' to install a compatible version.\n", constraint.Original)
	} else {
		fmt.Printf("Resolved to installed version: %s\n", resolved.Version.String())
	}

	return nil
}

func runVersionSet(cmd *cobra.Command, args []string) error {
	versionStr := args[0]
	projectPath := "."
	if len(args) > 1 {
		projectPath = args[1]
	}

	format, _ := cmd.Flags().GetString("format")

	// Validate version format
	if _, err := versionpkg.ParseVersionConstraint(versionStr); err != nil {
		return fmt.Errorf("invalid version constraint: %w", err)
	}

	detector := versionpkg.NewVersionDetector()
	if err := detector.CreateVersionFile(projectPath, versionStr, format); err != nil {
		return fmt.Errorf("failed to create version file: %w", err)
	}

	fmt.Printf("Set version requirement '%s' for project: %s\n", versionStr, projectPath)
	return nil
}

func runVersionCleanup(cmd *cobra.Command, args []string) error {
	vm, err := getVersionManager()
	if err != nil {
		return err
	}

	keep, _ := cmd.Flags().GetInt("keep")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	if dryRun {
		installed, err := vm.ListInstalledVersions()
		if err != nil {
			return fmt.Errorf("failed to list installed versions: %w", err)
		}

		if len(installed) <= keep {
			fmt.Printf("No cleanup needed. %d versions installed, keeping %d.\n", len(installed), keep)
			return nil
		}

		fmt.Printf("Would remove %d old versions (keeping %d most recent + active version):\n", len(installed)-keep, keep)
		// Show which versions would be removed...
		return nil
	}

	if err := vm.CleanupOldVersions(keep); err != nil {
		return fmt.Errorf("failed to cleanup old versions: %w", err)
	}

	if err := vm.GarbageCollect(); err != nil {
		return fmt.Errorf("failed to run garbage collection: %w", err)
	}

	fmt.Printf("Successfully cleaned up old versions (kept %d most recent)\n", keep)
	return nil
}

// Helper functions for JSON output

func printVersionsJSON(versions []versionpkg.InstalledVersion) error {
	data, err := json.MarshalIndent(versions, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func printAvailableVersionsJSON(versions []versionpkg.Version, limit int) error {
	if limit > 0 && len(versions) > limit {
		versions = versions[:limit]
	}

	data, err := json.MarshalIndent(versions, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func printConstraintJSON(constraint *versionpkg.VersionConstraint) error {
	data, err := json.MarshalIndent(constraint, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
