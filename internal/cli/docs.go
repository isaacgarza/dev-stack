package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"log/slog"

	"github.com/isaacgarza/dev-stack/internal/pkg/docs"
	"github.com/isaacgarza/dev-stack/internal/pkg/logger"

	"github.com/spf13/cobra"
)

var (
	docsVerbose      bool
	docsDryRun       bool
	docsCommandsOnly bool
	docsServicesOnly bool
	docsHugoSync     bool
	docsNoHugoSync   bool
	docsHugoDir      string
)

// docsCmd represents the docs command
var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Generate documentation from YAML manifests",
	Long: `Generate documentation from YAML manifests.

This command generates documentation files from the YAML manifests:
- Commands reference from scripts/commands.yaml -> docs/reference.md
- Services guide from services/services.yaml -> docs/services.md

The generated documentation replaces content between AUTO-GENERATED markers
in existing files, or creates new files if they don't exist.

By default, documentation is synchronized to Hugo's content directory for
website generation. Use --no-hugo-sync to disable this behavior.

Examples:
  dev-stack docs                    Generate all documentation and sync to Hugo
  dev-stack docs --commands-only   Generate only command reference
  dev-stack docs --services-only   Generate only services guide
  dev-stack docs --no-hugo-sync    Generate docs without Hugo sync
  dev-stack docs --dry-run         Preview changes without writing files
  dev-stack docs --verbose         Show detailed progress information`,
	RunE: runDocsGeneration,
}

func init() {
	rootCmd.AddCommand(docsCmd)

	docsCmd.Flags().BoolVarP(&docsVerbose, "verbose", "v", false, "Show detailed progress information")
	docsCmd.Flags().BoolVar(&docsDryRun, "dry-run", false, "Preview changes without writing files")
	docsCmd.Flags().BoolVar(&docsCommandsOnly, "commands-only", false, "Generate only command reference documentation")
	docsCmd.Flags().BoolVar(&docsServicesOnly, "services-only", false, "Generate only services guide documentation")
	docsCmd.Flags().BoolVar(&docsHugoSync, "hugo-sync", false, "Force Hugo sync (enabled by default)")
	docsCmd.Flags().BoolVar(&docsNoHugoSync, "no-hugo-sync", false, "Disable Hugo content synchronization")
	docsCmd.Flags().StringVar(&docsHugoDir, "hugo-dir", "content", "Hugo content directory path")
}

func runDocsGeneration(cmd *cobra.Command, args []string) error {
	log := logger.GetLogger()

	// Validate flags
	if docsCommandsOnly && docsServicesOnly {
		return fmt.Errorf("cannot specify both --commands-only and --services-only")
	}

	if docsHugoSync && docsNoHugoSync {
		return fmt.Errorf("cannot specify both --hugo-sync and --no-hugo-sync")
	}

	// Setup generation options
	options := docs.DefaultGenerationOptions()
	options.Verbose = docsVerbose
	options.DryRun = docsDryRun

	// Handle Hugo sync logic
	if docsNoHugoSync {
		options.EnableHugoSync = false
	} else if docsHugoSync {
		options.EnableHugoSync = true
	}
	// else use default from DefaultGenerationOptions (which is true)

	options.HugoContentDir = docsHugoDir

	// Convert relative paths to absolute paths based on current working directory
	if err := makePathsAbsolute(options); err != nil {
		return fmt.Errorf("failed to resolve paths: %w", err)
	}

	// Validate that source files exist
	if err := validateSourceFiles(options); err != nil {
		return fmt.Errorf("source file validation failed: %w", err)
	}

	generator := docs.NewGenerator(options)

	if docsVerbose {
		log.Info("Starting documentation generation",
			"commands_yaml", options.CommandsYAMLPath,
			"services_yaml", options.ServicesYAMLPath,
			"reference_md", options.ReferenceMDPath,
			"services_md", options.ServicesMDPath,
			"hugo_sync", options.EnableHugoSync,
			"hugo_dir", docsHugoDir,
			"dry_run", docsDryRun)
	}

	// Generate specific documentation based on flags
	if docsCommandsOnly {
		return generateCommandsOnly(generator, log)
	}

	if docsServicesOnly {
		return generateServicesOnly(generator, log)
	}

	// Generate all documentation
	return generateAllDocs(generator, options, log)
}

func generateCommandsOnly(generator *docs.Generator, log *slog.Logger) error {
	log.Info("Generating command reference documentation only")

	if err := generator.GenerateCommandReferenceOnly(); err != nil {
		log.Error("Failed to generate command reference", "error", err)
		return fmt.Errorf("command reference generation failed: %w", err)
	}

	if !docsDryRun {
		fmt.Println("✅ Command reference documentation generated successfully")
	} else {
		fmt.Println("✅ Command reference documentation generation validated (dry-run)")
	}

	return nil
}

func generateServicesOnly(generator *docs.Generator, log *slog.Logger) error {
	log.Info("Generating services guide documentation only")

	if err := generator.GenerateServicesGuideOnly(); err != nil {
		log.Error("Failed to generate services guide", "error", err)
		return fmt.Errorf("services guide generation failed: %w", err)
	}

	if !docsDryRun {
		fmt.Println("✅ Services guide documentation generated successfully")
	} else {
		fmt.Println("✅ Services guide documentation generation validated (dry-run)")
	}

	return nil
}

func generateAllDocs(generator *docs.Generator, options *docs.GenerationOptions, log *slog.Logger) error {
	log.Info("Generating all documentation")

	var result *docs.GenerationResult
	var hugoResult *docs.HugoSyncResult
	var err error

	if options.EnableHugoSync {
		result, hugoResult, err = generator.GenerateAllWithHugo()
	} else {
		result, err = generator.GenerateAll()
	}

	if err != nil {
		log.Error("Documentation generation failed", "error", err)
		return fmt.Errorf("documentation generation failed: %w", err)
	}

	// Report results
	if len(result.Errors) > 0 {
		log.Warn("Documentation generation completed with errors", "error_count", len(result.Errors))
		for i, err := range result.Errors {
			fmt.Printf("❌ Error %d: %v\n", i+1, err)
		}
		return fmt.Errorf("documentation generation completed with %d errors", len(result.Errors))
	}

	// Success message
	if !docsDryRun {
		fmt.Println("✅ Documentation generation completed successfully")
		fmt.Printf("   • Command reference: %t\n", result.CommandsGenerated)
		fmt.Printf("   • Services guide: %t\n", result.ServicesGenerated)
		fmt.Printf("   • README synced: %t\n", result.ReadmeSynced)
		fmt.Printf("   • Files updated: %d\n", len(result.FilesUpdated))

		if docsVerbose {
			for _, file := range result.FilesUpdated {
				fmt.Printf("     - %s\n", file)
			}
		}

		// Report Hugo sync results
		if hugoResult != nil {
			fmt.Printf("   • Hugo sync: enabled\n")
			fmt.Printf("     - Files copied: %d\n", len(hugoResult.FilesCopied))
			fmt.Printf("     - Files updated: %d\n", len(hugoResult.FilesUpdated))
			fmt.Printf("     - Files skipped: %d\n", len(hugoResult.FilesSkipped))

			if len(hugoResult.Errors) > 0 {
				fmt.Printf("     - Sync errors: %d\n", len(hugoResult.Errors))
				for i, err := range hugoResult.Errors {
					fmt.Printf("       ❌ Hugo Error %d: %v\n", i+1, err)
				}
			}

			if docsVerbose {
				fmt.Println("     - Hugo files synced:")
				for _, file := range hugoResult.FilesCopied {
					fmt.Printf("       + %s (new)\n", file)
				}
				for _, file := range hugoResult.FilesUpdated {
					fmt.Printf("       ~ %s (updated)\n", file)
				}
			}
		}
	} else {
		fmt.Println("✅ Documentation generation validated (dry-run)")
		fmt.Printf("   • Would generate command reference: %t\n", result.CommandsGenerated)
		fmt.Printf("   • Would generate services guide: %t\n", result.ServicesGenerated)
		fmt.Printf("   • Would sync README: %t\n", result.ReadmeSynced)
		fmt.Printf("   • Would update files: %d\n", len(result.FilesUpdated))

		if hugoResult != nil {
			fmt.Printf("   • Would sync to Hugo: %d files\n", len(hugoResult.FilesUpdated)+len(hugoResult.FilesCopied))
		}
	}

	return nil
}

func makePathsAbsolute(options *docs.GenerationOptions) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Convert relative paths to absolute
	if !filepath.IsAbs(options.CommandsYAMLPath) {
		options.CommandsYAMLPath = filepath.Join(cwd, options.CommandsYAMLPath)
	}
	if !filepath.IsAbs(options.ServicesYAMLPath) {
		options.ServicesYAMLPath = filepath.Join(cwd, options.ServicesYAMLPath)
	}
	if !filepath.IsAbs(options.ReferenceMDPath) {
		options.ReferenceMDPath = filepath.Join(cwd, options.ReferenceMDPath)
	}
	if !filepath.IsAbs(options.ServicesMDPath) {
		options.ServicesMDPath = filepath.Join(cwd, options.ServicesMDPath)
	}
	if !filepath.IsAbs(options.HugoContentDir) {
		options.HugoContentDir = filepath.Join(cwd, options.HugoContentDir)
	}
	if !filepath.IsAbs(options.DocsSourceDir) {
		options.DocsSourceDir = filepath.Join(cwd, options.DocsSourceDir)
	}

	return nil
}

func validateSourceFiles(options *docs.GenerationOptions) error {
	// Check if commands.yaml exists
	if _, err := os.Stat(options.CommandsYAMLPath); os.IsNotExist(err) {
		return fmt.Errorf("commands YAML file not found: %s", options.CommandsYAMLPath)
	}

	// Check if services.yaml exists
	if _, err := os.Stat(options.ServicesYAMLPath); os.IsNotExist(err) {
		return fmt.Errorf("services YAML file not found: %s", options.ServicesYAMLPath)
	}

	return nil
}
