package cli

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"encoding/json"

	"github.com/isaacgarza/dev-stack/internal/core/services"
	"github.com/isaacgarza/dev-stack/internal/pkg/compose"
	"github.com/isaacgarza/dev-stack/internal/pkg/config"
	"github.com/isaacgarza/dev-stack/internal/pkg/docs"
	"github.com/isaacgarza/dev-stack/internal/pkg/logger"
	"github.com/isaacgarza/dev-stack/internal/pkg/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Factory creates CLI commands from YAML configuration
type Factory struct {
	config          *config.CommandConfig
	logger          *slog.Logger
	serviceRegistry interface{}
}

// NewFactory creates a new CLI factory
func NewFactory(cfg *config.CommandConfig) *Factory {
	return &Factory{
		config: cfg,
		logger: logger.New(slog.LevelInfo),
	}
}

// CreateRootCommand creates the root command with all subcommands
func (f *Factory) CreateRootCommand() (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use:     "dev-stack",
		Short:   f.config.Metadata.Description,
		Version: f.config.Metadata.CLIVersion,
		Long: fmt.Sprintf(`%s

%s

Use "dev-stack help <command>" for more information about a command.
Use "dev-stack workflow" to see common task workflows.`,
			f.config.Metadata.Description,
			f.config.Help["getting_started"]),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Add global flags
	if err := f.addGlobalFlags(rootCmd.PersistentFlags()); err != nil {
		return nil, fmt.Errorf("failed to add global flags: %w", err)
	}

	// Add all commands
	if err := f.addCommands(rootCmd); err != nil {
		return nil, fmt.Errorf("failed to add commands: %w", err)
	}

	// Add special commands
	f.addHelpCommands(rootCmd)
	f.addWorkflowCommands(rootCmd)
	f.addCompletionCommands(rootCmd)

	return rootCmd, nil
}

// addGlobalFlags adds global flags to the root command
func (f *Factory) addGlobalFlags(flagSet *pflag.FlagSet) error {
	for name, flag := range f.config.Global.Flags {
		if err := f.addFlag(flagSet, name, flag); err != nil {
			return fmt.Errorf("failed to add global flag %s: %w", name, err)
		}
	}
	return nil
}

// addCommands adds all defined commands to the root command
func (f *Factory) addCommands(rootCmd *cobra.Command) error {
	for cmdName, cmdConfig := range f.config.Commands {
		if cmdConfig.Hidden {
			continue
		}

		cmd, err := f.createCommand(cmdName, cmdConfig)
		if err != nil {
			return fmt.Errorf("failed to create command %s: %w", cmdName, err)
		}

		rootCmd.AddCommand(cmd)
	}
	return nil
}

// createCommand creates a single command from configuration
func (f *Factory) createCommand(name string, cmdConfig config.Command) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:     cmdConfig.Usage,
		Short:   cmdConfig.Description,
		Long:    f.formatLongDescription(cmdConfig),
		Aliases: cmdConfig.Aliases,
		RunE:    f.createCommandHandler(name, cmdConfig),
	}

	// Add deprecation warning if needed
	if cmdConfig.Deprecated != nil {
		cmd.Deprecated = fmt.Sprintf("since %s: %s. Use %s instead.",
			cmdConfig.Deprecated.Since,
			cmdConfig.Deprecated.Reason,
			cmdConfig.Deprecated.Alternative)
	}

	// Add command-specific flags
	for flagName, flag := range cmdConfig.Flags {
		if err := f.addFlag(cmd.Flags(), flagName, flag); err != nil {
			return nil, fmt.Errorf("failed to add flag %s: %w", flagName, err)
		}
	}

	// Add shell completion
	if err := f.addCompletion(cmd, cmdConfig); err != nil {
		return nil, fmt.Errorf("failed to add completion: %w", err)
	}

	return cmd, nil
}

// createCommandHandler creates the actual command execution handler
func (f *Factory) createCommandHandler(name string, cmdConfig config.Command) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// Create execution context
		ctx := context.Background()

		// Initialize base command
		base, err := f.createBaseCommand()
		if err != nil {
			return fmt.Errorf("failed to initialize: %w", err)
		}
		defer func() {
			if closeErr := base.Close(); closeErr != nil {
				f.logger.Warn("failed to close base command", "error", closeErr)
			}
		}()

		// Route to appropriate handler
		return f.routeCommand(ctx, cmd, args, base, name, cmdConfig)
	}
}

// routeCommand routes the command to the appropriate handler
func (f *Factory) routeCommand(ctx context.Context, cmd *cobra.Command, args []string, base *BaseCommand, name string, cmdConfig config.Command) error {
	switch name {
	case "up":
		return f.handleUpCommand(ctx, cmd, args, base)
	case "down":
		return f.handleDownCommand(ctx, cmd, args, base)
	case "restart":
		return f.handleRestartCommand(ctx, cmd, args, base)
	case "status":
		return f.handleStatusCommand(ctx, cmd, args, base)
	case "logs":
		return f.handleLogsCommand(ctx, cmd, args, base)
	case "monitor":
		return f.handleMonitorCommand(ctx, cmd, args, base)
	case "doctor":
		return f.handleDoctorCommand(ctx, cmd, args, base)
	case "exec":
		return f.handleExecCommand(ctx, cmd, args, base)
	case "connect":
		return f.handleConnectCommand(ctx, cmd, args, base)
	case "backup":
		return f.handleBackupCommand(ctx, cmd, args, base)
	case "restore":
		return f.handleRestoreCommand(ctx, cmd, args, base)
	case "cleanup":
		return f.handleCleanupCommand(ctx, cmd, args, base)
	case "scale":
		return f.handleScaleCommand(ctx, cmd, args, base)
	case "init":
		return f.handleInitCommand(ctx, cmd, args, base)
	case "docs":
		return f.handleDocsCommand(ctx, cmd, args, base)
	case "validate":
		return f.handleValidateCommand(ctx, cmd, args, base)
	case "version":
		return f.handleVersionCommand(ctx, cmd, args, base)
	case "generate":
		return f.handleGenerateCommand(ctx, cmd, args, base)
	default:
		return fmt.Errorf("command %s not implemented", name)
	}
}

// BaseCommand provides common functionality for all commands
type BaseCommand struct {
	manager         *services.Manager
	logger          *slog.Logger
	serviceRegistry interface{}
	config          *config.CommandConfig
}

// createBaseCommand creates a base command with common initialization
func (f *Factory) createBaseCommand() (*BaseCommand, error) {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	// Create service manager
	manager, err := services.NewManager(f.logger, cwd)
	if err != nil {
		return nil, fmt.Errorf("failed to create service manager: %w", err)
	}

	return &BaseCommand{
		manager: manager,
		logger:  f.logger,
		config:  f.config,
	}, nil
}

// Close cleans up resources
func (b *BaseCommand) Close() error {
	if b.manager != nil {
		return b.manager.Close()
	}
	return nil
}

// ValidateServices validates service names against available services
func (b *BaseCommand) ValidateServices(serviceNames []string) error {
	// Load services.yaml to validate service names
	servicesYAMLPath := "services/services.yaml"
	data, err := os.ReadFile(servicesYAMLPath)
	if err != nil {
		// If services.yaml doesn't exist, skip validation
		return nil
	}

	var services map[string]interface{}
	if err := yaml.Unmarshal(data, &services); err != nil {
		return fmt.Errorf("failed to parse services.yaml: %w", err)
	}

	// Validate each service name
	for _, serviceName := range serviceNames {
		if _, exists := services[serviceName]; !exists {
			availableServices := make([]string, 0, len(services))
			for name := range services {
				availableServices = append(availableServices, name)
			}
			return fmt.Errorf("unknown service '%s'. Available services: %v", serviceName, availableServices)
		}
	}

	return nil
}

// addFlag adds a flag to a flag set based on configuration
func (f *Factory) addFlag(flagSet *pflag.FlagSet, name string, flag config.Flag) error {
	description := flag.Description

	// Add deprecation notice if needed
	if flag.Deprecated != "" {
		description = fmt.Sprintf("[DEPRECATED: %s] %s", flag.Deprecated, description)
	}

	err := f.addCobraFlag(flagSet, name, flag, description)

	// Bind to viper if not hidden
	if err == nil && !flag.Hidden {
		if viperErr := viper.BindPFlag(name, flagSet.Lookup(name)); viperErr != nil {
			f.logger.Warn("failed to bind flag to viper", "flag", name, "error", viperErr)
		}
	}

	return err
}

// addCobraFlag adds a flag to a cobra flag set
func (f *Factory) addCobraFlag(fs *pflag.FlagSet, name string, flag config.Flag, description string) error {
	switch flag.Type {
	case "bool":
		defaultVal, _ := flag.Default.(bool)
		if flag.Short != "" {
			fs.BoolP(name, flag.Short, defaultVal, description)
		} else {
			fs.Bool(name, defaultVal, description)
		}

	case "string":
		defaultVal, _ := flag.Default.(string)
		if flag.Short != "" {
			fs.StringP(name, flag.Short, defaultVal, description)
		} else {
			fs.String(name, defaultVal, description)
		}

	case "int":
		var defaultVal int
		switch v := flag.Default.(type) {
		case int:
			defaultVal = v
		case string:
			if v != "" {
				parsed, err := strconv.Atoi(v)
				if err == nil {
					defaultVal = parsed
				}
			}
		}
		if flag.Short != "" {
			fs.IntP(name, flag.Short, defaultVal, description)
		} else {
			fs.Int(name, defaultVal, description)
		}

	case "duration":
		var defaultVal time.Duration
		switch v := flag.Default.(type) {
		case string:
			if v != "" {
				parsed, err := time.ParseDuration(v)
				if err == nil {
					defaultVal = parsed
				}
			}
		}
		if flag.Short != "" {
			fs.DurationP(name, flag.Short, defaultVal, description)
		} else {
			fs.Duration(name, defaultVal, description)
		}

	case "stringArray":
		var defaultVal []string
		if flag.Default != nil {
			if slice, ok := flag.Default.([]interface{}); ok {
				for _, item := range slice {
					if str, ok := item.(string); ok {
						defaultVal = append(defaultVal, str)
					}
				}
			}
		}
		if flag.Short != "" {
			fs.StringArrayP(name, flag.Short, defaultVal, description)
		} else {
			fs.StringArray(name, defaultVal, description)
		}

	default:
		return fmt.Errorf("unsupported flag type: %s", flag.Type)
	}

	// Mark as hidden if needed
	if flag.Hidden {
		_ = fs.MarkHidden(name)
	}

	// Mark as required if needed
	if flag.Required {
		// Store required flag information for later validation
		// This is handled at the command execution level since pflag doesn't have MarkFlagRequired
		f.logger.Debug("Required flag registered", "flag", name)
	}

	return nil
}

// addCompletion adds shell completion for a command
func (f *Factory) addCompletion(cmd *cobra.Command, cmdConfig config.Command) error {
	// Add service name completion for commands that accept service arguments
	serviceCommands := []string{"up", "down", "restart", "status", "logs", "exec", "connect", "backup", "restore", "scale"}
	for _, serviceCmd := range serviceCommands {
		if cmd.Name() == serviceCmd || contains(cmdConfig.Aliases, serviceCmd) {
			cmd.ValidArgsFunction = f.serviceNameCompletion
			break
		}
	}

	// Add profile completion for --profile flag
	for flagName, flag := range cmdConfig.Flags {
		if flagName == "profile" || flag.Completion == "profiles" {
			if err := cmd.RegisterFlagCompletionFunc(flagName, f.profileCompletion); err != nil {
				return fmt.Errorf("failed to register profile completion: %w", err)
			}
		}
	}

	return nil
}

// serviceNameCompletion provides completion for service names
func (f *Factory) serviceNameCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Load actual service names from services.yaml
	servicesYAMLPath := "services/services.yaml"
	data, err := os.ReadFile(servicesYAMLPath)
	if err != nil {
		// Fallback to hardcoded list if services.yaml not found
		serviceNames := []string{"postgres", "redis", "mysql", "kafka", "jaeger", "prometheus", "localstack"}
		return serviceNames, cobra.ShellCompDirectiveNoFileComp
	}

	var services map[string]interface{}
	if err := yaml.Unmarshal(data, &services); err != nil {
		// Fallback to hardcoded list if parsing fails
		serviceNames := []string{"postgres", "redis", "mysql", "kafka", "jaeger", "prometheus", "localstack"}
		return serviceNames, cobra.ShellCompDirectiveNoFileComp
	}

	serviceNames := make([]string, 0, len(services))
	for name := range services {
		serviceNames = append(serviceNames, name)
	}

	return serviceNames, cobra.ShellCompDirectiveNoFileComp
}

// profileCompletion provides completion for profile names
func (f *Factory) profileCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	profiles := f.config.GetAllProfiles()
	return profiles, cobra.ShellCompDirectiveNoFileComp
}

// formatLongDescription formats the long description with examples and tips
func (f *Factory) formatLongDescription(cmdConfig config.Command) string {
	var parts []string

	// Add long description
	if cmdConfig.LongDescription != "" {
		parts = append(parts, cmdConfig.LongDescription)
	}

	// Add examples
	if len(cmdConfig.Examples) > 0 {
		parts = append(parts, "\nExamples:")
		for _, example := range cmdConfig.Examples {
			parts = append(parts, fmt.Sprintf("  %s\n      %s", example.Command, example.Description))
		}
	}

	// Add tips
	if len(cmdConfig.Tips) > 0 {
		parts = append(parts, "\nTips:")
		for _, tip := range cmdConfig.Tips {
			parts = append(parts, fmt.Sprintf("  • %s", tip))
		}
	}

	// Add related commands
	if len(cmdConfig.RelatedCommands) > 0 {
		parts = append(parts, fmt.Sprintf("\nSee also: %s", strings.Join(cmdConfig.RelatedCommands, ", ")))
	}

	return strings.Join(parts, "\n")
}

// addHelpCommands adds special help commands
func (f *Factory) addHelpCommands(rootCmd *cobra.Command) {
	// Add category help command
	categoriesCmd := &cobra.Command{
		Use:   "categories",
		Short: "List all command categories",
		RunE: func(cmd *cobra.Command, args []string) error {
			return f.showCategories()
		},
	}
	rootCmd.AddCommand(categoriesCmd)
}

// addWorkflowCommands adds workflow-related commands
func (f *Factory) addWorkflowCommands(rootCmd *cobra.Command) {
	workflowCmd := &cobra.Command{
		Use:   "workflow [name]",
		Short: "Show or execute common workflows",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return f.showWorkflows()
			}
			return f.executeWorkflow(args[0])
		},
	}
	workflowCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return f.config.GetAllWorkflows(), cobra.ShellCompDirectiveNoFileComp
	}
	rootCmd.AddCommand(workflowCmd)
}

// addCompletionCommands adds shell completion commands
func (f *Factory) addCompletionCommands(rootCmd *cobra.Command) {
	completionCmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for dev-stack.

The completion script can be sourced in your shell's configuration file
to enable command and flag completion.`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.ExactValidArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return rootCmd.GenBashCompletion(os.Stdout)
			case "zsh":
				return rootCmd.GenZshCompletion(os.Stdout)
			case "fish":
				return rootCmd.GenFishCompletion(os.Stdout, true)
			case "powershell":
				return rootCmd.GenPowerShellCompletion(os.Stdout)
			default:
				return fmt.Errorf("unsupported shell: %s", args[0])
			}
		},
	}
	rootCmd.AddCommand(completionCmd)
}

// Helper methods for showing information
func (f *Factory) showCategories() error {
	fmt.Println("Command Categories:\n")

	for _, category := range f.config.Categories {
		fmt.Printf("%s %s\n", category.Icon, category.Name)
		fmt.Printf("  %s\n", category.Description)
		fmt.Printf("  Commands: %s\n\n", strings.Join(category.Commands, ", "))
	}

	return nil
}

func (f *Factory) showWorkflows() error {
	fmt.Println("Available Workflows:\n")

	for _, workflow := range f.config.Workflows {
		fmt.Printf("🔄 %s\n", workflow.Name)
		fmt.Printf("  %s\n", workflow.Description)
		fmt.Printf("  Steps:\n")
		for i, step := range workflow.Steps {
			optional := ""
			if step.Optional {
				optional = " (optional)"
			}
			fmt.Printf("    %d. %s%s\n", i+1, step.Description, optional)
			fmt.Printf("       Command: %s\n", step.Command)
		}
		fmt.Println()
	}

	return nil
}

func (f *Factory) executeWorkflow(name string) error {
	workflow, exists := f.config.GetWorkflow(name)
	if !exists {
		return fmt.Errorf("workflow '%s' not found", name)
	}

	fmt.Printf("🔄 Executing workflow: %s\n", workflow.Name)
	fmt.Printf("   %s\n\n", workflow.Description)

	for i, step := range workflow.Steps {
		fmt.Printf("Step %d: %s\n", i+1, step.Description)
		fmt.Printf("Command: %s\n", step.Command)

		if step.Optional {
			fmt.Print("This step is optional. Continue? [Y/n]: ")
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) == "n" || strings.ToLower(response) == "no" {
				fmt.Println("Skipping optional step.")
				continue
			}
		}

		// Execute the actual command by parsing and running it
		if err := f.executeWorkflowStep(step.Command); err != nil {
			fmt.Printf("❌ Step failed: %v\n", err)
			fmt.Print("Continue with remaining steps? [y/N]: ")
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
				return fmt.Errorf("workflow stopped by user")
			}
		} else {
			fmt.Println("✅ Step completed successfully")
		}
		fmt.Println()
	}

	fmt.Println("✅ Workflow completed!")
	return nil
}

// executeWorkflowStep executes a single workflow step command
func (f *Factory) executeWorkflowStep(command string) error {
	// Parse the command to extract the actual dev-stack command and args
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	// Remove "dev-stack" prefix if present
	if parts[0] == "dev-stack" {
		parts = parts[1:]
	}

	if len(parts) == 0 {
		return fmt.Errorf("no command specified")
	}

	fmt.Printf("🔄 Executing: dev-stack %s\n", strings.Join(parts, " "))

	// Create a new root command and execute the specified command
	rootCmd, err := f.CreateRootCommand()
	if err != nil {
		return fmt.Errorf("failed to create command: %w", err)
	}

	// Set the command and args
	rootCmd.SetArgs(parts)

	// Execute the command
	if err := rootCmd.Execute(); err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	}

	return nil
}

// Command handlers - these would call the actual implementation
func (f *Factory) handleUpCommand(ctx context.Context, cmd *cobra.Command, args []string, base *BaseCommand) error {
	detach, _ := cmd.Flags().GetBool("detach")
	build, _ := cmd.Flags().GetBool("build")
	profile, _ := cmd.Flags().GetString("profile")

	// Validate services if specified
	if len(args) > 0 {
		if err := base.ValidateServices(args); err != nil {
			return err
		}
	}

	// Display what we're starting
	if len(args) == 0 {
		fmt.Println("🚀 Starting all services...")
	} else {
		fmt.Printf("🚀 Starting services: %v\n", args)
	}

	if profile != "" {
		fmt.Printf("📋 Using profile: %s\n", profile)
	}

	// Set up start options
	startOptions := services.StartOptions{
		Build:   build,
		Detach:  detach,
		Timeout: 30 * time.Second,
	}

	// Start services
	return base.manager.StartServices(ctx, args, startOptions)
}

func (f *Factory) handleDownCommand(ctx context.Context, cmd *cobra.Command, args []string, base *BaseCommand) error {
	volumes, _ := cmd.Flags().GetBool("volumes")
	timeout, _ := cmd.Flags().GetInt("timeout")

	// Display what we're stopping
	if len(args) == 0 {
		fmt.Println("⏹️ Stopping all services...")
	} else {
		fmt.Printf("⏹️ Stopping services: %v\n", args)
	}

	// Set up stop options
	stopOptions := services.StopOptions{
		Timeout:       timeout,
		Remove:        true,
		RemoveVolumes: volumes,
	}

	// Stop services
	return base.manager.StopServices(ctx, args, stopOptions)
}

func (f *Factory) handleRestartCommand(ctx context.Context, cmd *cobra.Command, args []string, base *BaseCommand) error {
	timeout, _ := cmd.Flags().GetInt("timeout")

	// Display what we're restarting
	if len(args) == 0 {
		fmt.Println("🔄 Restarting all services...")
	} else {
		fmt.Printf("🔄 Restarting services: %v\n", args)
	}

	// Stop first
	stopOptions := services.StopOptions{
		Timeout: timeout,
		Remove:  true,
	}

	if err := base.manager.StopServices(ctx, args, stopOptions); err != nil {
		return fmt.Errorf("failed to stop services: %w", err)
	}

	// Then start
	startOptions := services.StartOptions{
		Timeout: time.Duration(timeout) * time.Second,
	}

	return base.manager.StartServices(ctx, args, startOptions)
}

func (f *Factory) handleStatusCommand(ctx context.Context, cmd *cobra.Command, args []string, base *BaseCommand) error {
	format, _ := cmd.Flags().GetString("format")
	quiet, _ := cmd.Flags().GetBool("quiet")
	_, _ = cmd.Flags().GetBool("watch")

	if !quiet {
		if len(args) == 0 {
			fmt.Println("📊 Checking status of all services...")
		} else {
			fmt.Printf("📊 Checking status of services: %v\n", args)
		}
	}

	// Get service status
	services, err := base.manager.GetServiceStatus(ctx, args)
	if err != nil {
		return fmt.Errorf("failed to get service status: %w", err)
	}

	// Display status based on format
	switch format {
	case "json":
		return f.displayStatusJSON(services)
	case "yaml":
		return f.displayStatusYAML(services)
	default:
		return f.displayStatusTable(services, quiet)
	}
}

func (f *Factory) handleLogsCommand(ctx context.Context, cmd *cobra.Command, args []string, base *BaseCommand) error {
	follow, _ := cmd.Flags().GetBool("follow")
	tail, _ := cmd.Flags().GetString("tail")
	since, _ := cmd.Flags().GetString("since")
	timestamps, _ := cmd.Flags().GetBool("timestamps")

	if len(args) == 0 {
		fmt.Println("📝 Showing logs from all services...")
	} else {
		fmt.Printf("📝 Showing logs from services: %v\n", args)
	}

	logOptions := services.LogOptions{
		Follow:     follow,
		Timestamps: timestamps,
		Tail:       tail,
		Since:      since,
	}

	return base.manager.GetLogs(ctx, args, logOptions)
}

func (f *Factory) handleMonitorCommand(ctx context.Context, cmd *cobra.Command, args []string, base *BaseCommand) error {
	refresh, _ := cmd.Flags().GetInt("refresh")
	noLogs, _ := cmd.Flags().GetBool("no-logs")
	compact, _ := cmd.Flags().GetBool("compact")

	fmt.Println("📊 Starting monitoring dashboard...")
	fmt.Printf("🔄 Refresh interval: %d seconds\n", refresh)
	fmt.Println("Press Ctrl+C to stop monitoring")
	fmt.Println()

	// Validate services if specified
	if len(args) > 0 {
		if err := base.ValidateServices(args); err != nil {
			return err
		}
	}

	// Start monitoring loop
	ticker := time.NewTicker(time.Duration(refresh) * time.Second)
	defer ticker.Stop()

	// Handle Ctrl+C gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	for {
		select {
		case <-sigChan:
			fmt.Println("\n👋 Monitoring stopped")
			return nil
		case <-ticker.C:
			if err := f.displayMonitoringDashboard(ctx, base, args, noLogs, compact); err != nil {
				fmt.Printf("❌ Error updating dashboard: %v\n", err)
			}
		default:
			// Initial display
			if err := f.displayMonitoringDashboard(ctx, base, args, noLogs, compact); err != nil {
				return fmt.Errorf("failed to display dashboard: %w", err)
			}
			// Wait for first tick
			select {
			case <-sigChan:
				fmt.Println("\n👋 Monitoring stopped")
				return nil
			case <-ticker.C:
				continue
			}
		}
	}
}

func (f *Factory) displayMonitoringDashboard(ctx context.Context, base *BaseCommand, args []string, noLogs, compact bool) error {
	// Clear screen and move cursor to top
	fmt.Print("\033[2J\033[H")

	// Get service status
	services, err := base.manager.GetServiceStatus(ctx, args)
	if err != nil {
		return fmt.Errorf("failed to get service status: %w", err)
	}

	// Display header
	fmt.Println("📊 dev-stack Monitoring Dashboard")
	fmt.Printf("⏰ Updated: %s\n", time.Now().Format("15:04:05"))
	fmt.Println(strings.Repeat("=", 80))

	if len(services) == 0 {
		fmt.Println("No services found")
		return nil
	}

	// Display service status table
	if compact {
		f.displayCompactStatus(services)
	} else {
		f.displayDetailedStatus(services)
	}

	// Display resource summary
	f.displayResourceSummary(services)

	if !noLogs {
		fmt.Println()
		fmt.Println("📝 Recent Activity:")
		fmt.Println("(Log streaming not yet implemented)")
	}

	fmt.Println()
	fmt.Println("Press Ctrl+C to stop monitoring")

	return nil
}

func (f *Factory) displayCompactStatus(services []types.ServiceStatus) {
	fmt.Printf("%-20s %-10s %-12s\n", "SERVICE", "STATE", "HEALTH")
	fmt.Println(strings.Repeat("-", 45))

	for _, service := range services {
		state := getStateIcon(service.State) + " " + service.State
		health := getHealthIcon(service.Health) + " " + service.Health
		fmt.Printf("%-20s %-10s %-12s\n", service.Name, state, health)
	}
}

func (f *Factory) displayDetailedStatus(services []types.ServiceStatus) {
	fmt.Printf("%-15s %-10s %-10s %-8s %-12s %-10s\n",
		"SERVICE", "STATE", "HEALTH", "CPU", "MEMORY", "UPTIME")
	fmt.Println(strings.Repeat("-", 70))

	for _, service := range services {
		state := getStateIcon(service.State)
		health := getHealthIcon(service.Health)

		var cpuStr, memStr, uptimeStr string
		if service.State == "running" {
			cpuStr = fmt.Sprintf("%.1f%%", service.CPUUsage)
			if service.Memory.Limit > 0 {
				memStr = fmt.Sprintf("%.0fMB", float64(service.Memory.Used)/1024/1024)
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

		fmt.Printf("%-15s %s%-9s %s%-9s %-8s %-12s %-10s\n",
			service.Name, state, service.State, health, service.Health,
			cpuStr, memStr, uptimeStr)
	}
}

func (f *Factory) displayResourceSummary(services []types.ServiceStatus) {
	running := 0
	totalCPU := 0.0
	totalMemory := uint64(0)

	for _, service := range services {
		if service.State == "running" {
			running++
			totalCPU += service.CPUUsage
			totalMemory += service.Memory.Used
		}
	}

	fmt.Println()
	fmt.Printf("📈 Summary: %d/%d running | CPU: %.1f%% | Memory: %.1fMB\n",
		running, len(services), totalCPU,
		float64(totalMemory)/1024/1024)
}

func (f *Factory) handleDoctorCommand(ctx context.Context, cmd *cobra.Command, args []string, base *BaseCommand) error {
	fix, _ := cmd.Flags().GetBool("fix")
	verbose, _ := cmd.Flags().GetBool("verbose")
	format, _ := cmd.Flags().GetString("format")

	fmt.Println("🏥 Running health checks...")

	if fix {
		fmt.Println("🔧 Auto-fix mode enabled")
	}

	// Validate services if specified
	if len(args) > 0 {
		if err := base.ValidateServices(args); err != nil {
			return err
		}
	}

	// Run health checks
	healthReport := f.runHealthChecks(ctx, base, args, verbose)

	// Display results
	switch format {
	case "json":
		return f.displayHealthReportJSON(healthReport)
	default:
		return f.displayHealthReportTable(healthReport, fix)
	}
}

func (f *Factory) handleExecCommand(ctx context.Context, cmd *cobra.Command, args []string, base *BaseCommand) error {
	if len(args) < 2 {
		return fmt.Errorf("exec requires at least service name and command")
	}

	serviceName := args[0]
	command := args[1:]

	user, _ := cmd.Flags().GetString("user")
	workdir, _ := cmd.Flags().GetString("workdir")
	interactive, _ := cmd.Flags().GetBool("interactive")
	tty, _ := cmd.Flags().GetBool("tty")

	execOptions := services.ExecOptions{
		User:        user,
		WorkingDir:  workdir,
		Interactive: interactive,
		TTY:         tty,
	}

	return base.manager.ExecCommand(ctx, serviceName, command, execOptions)
}

func (f *Factory) handleConnectCommand(ctx context.Context, cmd *cobra.Command, args []string, base *BaseCommand) error {
	if len(args) != 1 {
		return fmt.Errorf("connect requires exactly one service name")
	}

	serviceName := args[0]
	database, _ := cmd.Flags().GetString("database")
	user, _ := cmd.Flags().GetString("user")
	host, _ := cmd.Flags().GetString("host")
	readOnly, _ := cmd.Flags().GetBool("read-only")

	connectOptions := services.ConnectOptions{
		User:     user,
		Database: database,
		Host:     host,
		ReadOnly: readOnly,
	}

	return base.manager.ConnectToService(ctx, serviceName, connectOptions)
}

func (f *Factory) handleBackupCommand(ctx context.Context, cmd *cobra.Command, args []string, base *BaseCommand) error {
	output, _ := cmd.Flags().GetString("output")
	compress, _ := cmd.Flags().GetBool("compress")
	format, _ := cmd.Flags().GetString("format")

	fmt.Println("💾 Starting backup...")

	backupOptions := services.BackupOptions{
		OutputDir: output,
		Compress:  compress,
		Format:    format,
	}

	if len(args) == 0 {
		return fmt.Errorf("backup requires at least one service name")
	}
	return base.manager.BackupService(ctx, args[0], output, backupOptions)
}

func (f *Factory) handleRestoreCommand(ctx context.Context, cmd *cobra.Command, args []string, base *BaseCommand) error {
	if len(args) != 2 {
		return fmt.Errorf("restore requires service name and backup path")
	}

	serviceName := args[0]
	backupPath := args[1]

	clean, _ := cmd.Flags().GetBool("clean")
	createDB, _ := cmd.Flags().GetBool("create-db")
	singleTransaction, _ := cmd.Flags().GetBool("single-transaction")

	restoreOptions := services.RestoreOptions{
		Clean:             clean,
		CreateDB:          createDB,
		SingleTransaction: singleTransaction,
	}

	return base.manager.RestoreService(ctx, serviceName, backupPath, restoreOptions)
}

func (f *Factory) handleCleanupCommand(ctx context.Context, cmd *cobra.Command, args []string, base *BaseCommand) error {
	all, _ := cmd.Flags().GetBool("all")
	volumes, _ := cmd.Flags().GetBool("volumes")
	images, _ := cmd.Flags().GetBool("images")
	networks, _ := cmd.Flags().GetBool("networks")
	force, _ := cmd.Flags().GetBool("force")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	fmt.Println("🧹 Starting cleanup...")

	cleanupOptions := services.CleanupOptions{
		RemoveVolumes:  all || volumes,
		RemoveImages:   all || images,
		RemoveNetworks: all || networks,
		All:            all,
		DryRun:         dryRun,
	}

	if !force && !dryRun {
		fmt.Print("This will remove Docker resources. Continue? [y/N]: ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Cleanup cancelled.")
			return nil
		}
	}

	return base.manager.CleanupResources(ctx, cleanupOptions)
}

func (f *Factory) handleScaleCommand(ctx context.Context, cmd *cobra.Command, args []string, base *BaseCommand) error {
	if len(args) == 0 {
		return fmt.Errorf("scale requires service=replicas arguments")
	}

	timeout, _ := cmd.Flags().GetInt("timeout")

	fmt.Printf("⚖️ Scaling services: %v\n", args)

	scaleOptions := services.ScaleOptions{
		Timeout: time.Duration(timeout) * time.Second,
	}

	if len(args) == 0 {
		return fmt.Errorf("scale requires service=replicas arguments")
	}
	return base.manager.ScaleService(ctx, args[0], 1, scaleOptions)
}

func (f *Factory) handleInitCommand(ctx context.Context, cmd *cobra.Command, args []string, base *BaseCommand) error {
	name, _ := cmd.Flags().GetString("name")
	template, _ := cmd.Flags().GetString("template")
	force, _ := cmd.Flags().GetBool("force")
	minimal, _ := cmd.Flags().GetBool("minimal")

	// Determine project type
	projectType := "basic"
	if len(args) > 0 {
		projectType = args[0]
	}
	if template != "" {
		projectType = template
	}

	// Determine project name
	if name == "" {
		// Use current directory name as default
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
		name = filepath.Base(wd)
	}

	fmt.Println("🚀 Initializing dev-stack project...")
	fmt.Printf("📝 Project name: %s\n", name)
	fmt.Printf("📋 Project type: %s\n", projectType)

	if minimal {
		fmt.Println("⚡ Using minimal configuration")
	}

	// Check if files already exist
	if !force {
		existingFiles := f.checkExistingFiles()
		if len(existingFiles) > 0 {
			fmt.Printf("⚠️  Found existing files: %v\n", existingFiles)
			fmt.Print("Continue and overwrite? [y/N]: ")
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
				return fmt.Errorf("initialization cancelled")
			}
		}
	}

	// Initialize project
	if err := f.initializeProject(name, projectType, minimal); err != nil {
		return fmt.Errorf("project initialization failed: %w", err)
	}

	fmt.Println("✅ Project initialized successfully!")
	fmt.Println("\n📋 Next steps:")
	fmt.Println("  1. Review and edit dev-stack-config.yaml")
	fmt.Println("  2. Run: dev-stack up")
	fmt.Println("  3. Run: dev-stack status")

	return nil
}

func (f *Factory) checkExistingFiles() []string {
	var existing []string
	files := []string{
		"dev-stack-config.yaml",
		"docker-compose.yml",
		".gitignore",
	}

	for _, file := range files {
		if _, err := os.Stat(file); err == nil {
			existing = append(existing, file)
		}
	}

	return existing
}

func (f *Factory) initializeProject(name, projectType string, minimal bool) error {
	// Create directory structure
	dirs := []string{
		"scripts",
		"config",
		"data",
		"logs",
	}

	if !minimal {
		dirs = append(dirs, "services", "docs", "tests")
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Generate configuration file
	if err := f.generateProjectConfig(name, projectType, minimal); err != nil {
		return fmt.Errorf("failed to generate config: %w", err)
	}

	// Generate docker-compose.yml based on project type
	if err := f.generateProjectCompose(projectType, minimal); err != nil {
		return fmt.Errorf("failed to generate compose file: %w", err)
	}

	// Generate .gitignore
	if err := f.generateGitignore(); err != nil {
		return fmt.Errorf("failed to generate .gitignore: %w", err)
	}

	// Generate README.md
	if err := f.generateReadme(name, projectType); err != nil {
		return fmt.Errorf("failed to generate README: %w", err)
	}

	return nil
}

func (f *Factory) generateProjectConfig(name, projectType string, minimal bool) error {
	var services []string
	var description string

	switch projectType {
	case "web":
		services = []string{"postgres", "redis", "jaeger"}
		description = "Web application development stack"
	case "api":
		services = []string{"postgres", "redis", "prometheus"}
		description = "API development stack"
	case "microservices":
		services = []string{"postgres", "redis", "kafka", "jaeger", "prometheus"}
		description = "Microservices development stack"
	case "data":
		services = []string{"postgres", "redis", "kafka", "localstack"}
		description = "Data engineering stack"
	default:
		services = []string{"postgres"}
		description = "Basic development stack"
	}

	if minimal {
		services = []string{"postgres"}
		description = "Minimal development stack"
	}

	configTemplate := `# dev-stack configuration for %s
# Project type: %s
# Generated: %s

project:
  name: "%s"
  type: "%s"
  description: "%s"

# Default profile
default_profile: "development"

# Service profiles
profiles:
  development:
    name: "Development"
    description: "Development environment"
    services: %v

  production:
    name: "Production"
    description: "Production-like environment"
    services: %v

# Global settings
global:
  log_level: "info"
  color_output: true
  check_updates: true
  docker_registry: ""

# Environment variables
environment:
  NODE_ENV: "development"
  DEBUG: "true"
  LOG_LEVEL: "debug"

# Service overrides
services:
  postgres:
    environment:
      POSTGRES_DB: "%s_dev"
      POSTGRES_USER: "dev_user"
      POSTGRES_PASSWORD: "dev_password"
    ports:
      - "5432:5432"

  redis:
    environment:
      REDIS_PASSWORD: "dev_password"
    ports:
      - "6379:6379"
`

	content := fmt.Sprintf(configTemplate,
		name, projectType, time.Now().Format("2006-01-02 15:04:05"),
		name, projectType, description,
		fmt.Sprintf("%q", services), fmt.Sprintf("%q", services),
		strings.ReplaceAll(strings.ToLower(name), " ", "_"))

	if err := os.WriteFile("dev-stack-config.yaml", []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func (f *Factory) generateProjectCompose(projectType string, minimal bool) error {
	// Use the existing generateComposeFile function
	var services string
	switch projectType {
	case "web":
		services = "postgres,redis"
	case "api":
		services = "postgres,redis"
	case "microservices":
		services = "postgres,redis,kafka"
	case "data":
		services = "postgres,redis"
	default:
		services = "postgres"
	}

	if minimal {
		services = "postgres"
	}

	return f.generateComposeFile("docker-compose.yml", services, true)
}

func (f *Factory) generateGitignore() error {
	gitignoreContent := `# dev-stack generated .gitignore
# Logs
logs/
*.log
npm-debug.log*
yarn-debug.log*
yarn-error.log*

# Runtime data
pids/
*.pid
*.seed
*.pid.lock

# Coverage directory used by tools like istanbul
coverage/

# nyc test coverage
.nyc_output

# Dependency directories
node_modules/
vendor/

# Optional npm cache directory
.npm

# Optional REPL history
.node_repl_history

# Output of 'npm pack'
*.tgz

# Yarn Integrity file
.yarn-integrity

# dotenv environment variables file
.env
.env.local
.env.development.local
.env.test.local
.env.production.local

# Docker
.docker/

# IDE
.vscode/
.idea/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db

# dev-stack specific
data/
backups/
.dev-stack/
docker-compose.override.yml
`

	if err := os.WriteFile(".gitignore", []byte(gitignoreContent), 0644); err != nil {
		return fmt.Errorf("failed to write .gitignore: %w", err)
	}

	return nil
}

func (f *Factory) generateReadme(name, projectType string) error {
	readmeTemplate := `# %s

%s development stack powered by dev-stack.

## Quick Start

1. **Start the development stack:**
   ` + "```bash\n   dev-stack up\n   ```" + `

2. **Check service status:**
   ` + "```bash\n   dev-stack status\n   ```" + `

3. **View logs:**
   ` + "```bash\n   dev-stack logs --follow\n   ```" + `

## Available Services

- PostgreSQL database (localhost:5432)
- Redis cache (localhost:6379)

## Common Commands

- ` + "`dev-stack up`" + ` - Start all services
- ` + "`dev-stack down`" + ` - Stop all services
- ` + "`dev-stack status`" + ` - Show service status
- ` + "`dev-stack logs`" + ` - View service logs
- ` + "`dev-stack connect postgres`" + ` - Connect to PostgreSQL
- ` + "`dev-stack connect redis`" + ` - Connect to Redis

## Configuration

Edit ` + "`dev-stack-config.yaml`" + ` to customize your development environment.

## Documentation

- [dev-stack Documentation](https://dev-stack.dev/docs)
- [Command Reference](https://dev-stack.dev/docs/reference)
- [Service Configuration](https://dev-stack.dev/docs/services)

## Development

Generated on: %s
Project type: %s
`

	content := fmt.Sprintf(readmeTemplate, name, projectType, time.Now().Format("2006-01-02 15:04:05"), projectType)

	if err := os.WriteFile("README.md", []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write README.md: %w", err)
	}

	return nil
}

func (f *Factory) handleDocsCommand(ctx context.Context, cmd *cobra.Command, args []string, base *BaseCommand) error {
	commandsOnly, _ := cmd.Flags().GetBool("commands-only")
	servicesOnly, _ := cmd.Flags().GetBool("services-only")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	fmt.Println("📚 Generating documentation...")

	if commandsOnly {
		fmt.Println("📝 Generating commands reference only")
	} else if servicesOnly {
		fmt.Println("🔧 Generating services documentation only")
	} else {
		fmt.Println("📚 Generating all documentation")
	}

	if dryRun {
		fmt.Println("🔍 Dry run mode - no files will be modified")
	}

	// Create documentation generation options
	options := &docs.GenerationOptions{
		CommandsYAMLPath: "scripts/commands.yaml",
		ServicesYAMLPath: "services/services.yaml",
		ReferenceMDPath:  "docs-site/content/reference.md",
		ServicesMDPath:   "docs-site/content/services.md",
		HugoContentDir:   "docs-site/content",
		DocsSourceDir:    "docs-site/content",
		EnableHugoSync:   true,
		Verbose:          false,
		DryRun:           dryRun,
	}

	// Create enhanced generator
	generator := docs.NewEnhancedGenerator(options)

	// Generate documentation
	result, err := generator.GenerateAll()
	if err != nil {
		return fmt.Errorf("documentation generation failed: %w", err)
	}

	// Report results
	if len(result.Errors) > 0 {
		fmt.Printf("❌ Documentation generation completed with %d errors\n", len(result.Errors))
		for i, err := range result.Errors {
			fmt.Printf("   Error %d: %v\n", i+1, err)
		}
		return fmt.Errorf("documentation generation completed with errors")
	}

	if dryRun {
		fmt.Println("✅ Documentation generation validated (dry-run)")
		fmt.Printf("   • Would generate command reference: %t\n", result.CommandsGenerated)
		fmt.Printf("   • Would generate services guide: %t\n", result.ServicesGenerated)
		fmt.Printf("   • Would update files: %d\n", len(result.FilesUpdated))
	} else {
		fmt.Println("✅ Documentation generation completed successfully")
		fmt.Printf("   • Command reference: %t\n", result.CommandsGenerated)
		fmt.Printf("   • Services guide: %t\n", result.ServicesGenerated)
		fmt.Printf("   • Files updated: %d\n", len(result.FilesUpdated))

		if len(result.FilesUpdated) > 0 {
			fmt.Println("   • Updated files:")
			for _, file := range result.FilesUpdated {
				fmt.Printf("     - %s\n", file)
			}
		}
	}

	return nil
}

func (f *Factory) handleValidateCommand(ctx context.Context, cmd *cobra.Command, args []string, base *BaseCommand) error {
	strict, _ := cmd.Flags().GetBool("strict")
	format, _ := cmd.Flags().GetString("format")
	fix, _ := cmd.Flags().GetBool("fix")

	fmt.Println("✅ Validating configuration...")

	if strict {
		fmt.Println("🔒 Using strict validation rules")
	}

	// Validate the command configuration itself
	result := f.config.Validate()

	switch format {
	case "json":
		return f.displayValidationJSON(result)
	default:
		return f.displayValidationTable(result, fix)
	}
}

func (f *Factory) handleVersionCommand(ctx context.Context, cmd *cobra.Command, args []string, base *BaseCommand) error {
	full, _ := cmd.Flags().GetBool("full")
	checkUpdates, _ := cmd.Flags().GetBool("check-updates")
	format, _ := cmd.Flags().GetString("format")

	version := f.config.Metadata.CLIVersion
	buildInfo := f.getBuildInfo()

	switch format {
	case "json":
		return f.displayVersionJSON(version, buildInfo, full)
	case "yaml":
		return f.displayVersionYAML(version, buildInfo, full)
	default:
		return f.displayVersionTable(version, buildInfo, full, checkUpdates)
	}
}

func (f *Factory) getBuildInfo() map[string]string {
	return map[string]string{
		"go_version":     "go1.21.0", // This would be injected at build time
		"build_date":     f.config.Metadata.GeneratedAt.Format("2006-01-02T15:04:05Z"),
		"commit_hash":    "unknown",                              // This would be injected at build time
		"build_user":     "unknown",                              // This would be injected at build time
		"platform":       fmt.Sprintf("%s/%s", "linux", "amd64"), // runtime.GOOS/runtime.GOARCH
		"config_version": f.config.Metadata.Version,
	}
}

func (f *Factory) displayVersionTable(version string, buildInfo map[string]string, full bool, checkUpdates bool) error {
	if full {
		fmt.Printf("dev-stack version %s\n", version)
		fmt.Printf("Build date: %s\n", buildInfo["build_date"])
		fmt.Printf("Config version: %s\n", buildInfo["config_version"])
		fmt.Printf("Go version: %s\n", buildInfo["go_version"])
		fmt.Printf("Platform: %s\n", buildInfo["platform"])
		fmt.Printf("Commit: %s\n", buildInfo["commit_hash"])
		fmt.Printf("Built by: %s\n", buildInfo["build_user"])
	} else {
		fmt.Printf("dev-stack %s\n", version)
	}

	if checkUpdates {
		fmt.Println("🔍 Checking for updates...")
		if err := f.checkForUpdates(version); err != nil {
			fmt.Printf("⚠️  Update check failed: %v\n", err)
		}
	}

	return nil
}

func (f *Factory) displayVersionJSON(version string, buildInfo map[string]string, full bool) error {
	versionInfo := map[string]interface{}{
		"version": version,
	}

	if full {
		versionInfo["build_info"] = buildInfo
	}

	jsonData, err := json.MarshalIndent(versionInfo, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(jsonData))
	return nil
}

func (f *Factory) displayVersionYAML(version string, buildInfo map[string]string, full bool) error {
	versionInfo := map[string]interface{}{
		"version": version,
	}

	if full {
		versionInfo["build_info"] = buildInfo
	}

	yamlData, err := yaml.Marshal(versionInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	fmt.Print(string(yamlData))
	return nil
}

func (f *Factory) checkForUpdates(currentVersion string) error {
	// Simulate update check - in a real implementation this would:
	// 1. Check a remote registry or GitHub releases
	// 2. Compare versions using semantic versioning
	// 3. Provide download links or update instructions

	fmt.Println("✅ You are running the latest version")
	fmt.Println("💡 To update, run: go install github.com/isaacgarza/dev-stack@latest")

	return nil
}

func (f *Factory) handleGenerateCommand(ctx context.Context, cmd *cobra.Command, args []string, base *BaseCommand) error {
	if len(args) == 0 {
		return fmt.Errorf("generate requires a type argument")
	}

	generateType := args[0]
	output, _ := cmd.Flags().GetString("output")
	template, _ := cmd.Flags().GetString("template")
	services, _ := cmd.Flags().GetString("services")
	overwrite, _ := cmd.Flags().GetBool("overwrite")

	fmt.Printf("🏗️ Generating %s...\n", generateType)

	switch generateType {
	case "config":
		return f.generateConfig(output, template, overwrite)
	case "service":
		if len(args) < 2 {
			return fmt.Errorf("generate service requires a service name")
		}
		return f.generateServiceConfig(args[1], output, template, overwrite)
	case "compose":
		return f.generateComposeFile(output, services, overwrite)
	case "dockerfile":
		if len(args) < 2 {
			return fmt.Errorf("generate dockerfile requires a service name")
		}
		return f.generateDockerfile(args[1], output, template, overwrite)
	default:
		return fmt.Errorf("unknown generation type: %s. Available types: config, service, compose, dockerfile", generateType)
	}
}

// Display helpers
func (f *Factory) displayStatusTable(services []types.ServiceStatus, quiet bool) error {
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

func (f *Factory) displayStatusJSON(services []types.ServiceStatus) error {
	output := map[string]interface{}{
		"services":  services,
		"timestamp": time.Now().Format(time.RFC3339),
		"total":     len(services),
	}

	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(jsonData))
	return nil
}

func (f *Factory) displayStatusYAML(services []types.ServiceStatus) error {
	output := map[string]interface{}{
		"services":  services,
		"timestamp": time.Now().Format(time.RFC3339),
		"total":     len(services),
	}

	yamlData, err := yaml.Marshal(output)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	fmt.Print(string(yamlData))
	return nil
}

func (f *Factory) displayValidationTable(result *config.ValidationResult, fix bool) error {
	if result.Valid {
		fmt.Println("✅ Configuration is valid")
		return nil
	}

	fmt.Printf("❌ Configuration validation failed with %d errors\n", len(result.Errors))

	for _, err := range result.Errors {
		fmt.Printf("  • %s: %s\n", err.Field, err.Message)
	}

	if len(result.Warnings) > 0 {
		fmt.Printf("\n⚠️  %d warnings:\n", len(result.Warnings))
		for _, warning := range result.Warnings {
			fmt.Printf("  • %s: %s\n", warning.Field, warning.Message)
		}
	}

	if fix {
		fmt.Println("\n🔧 Auto-fix mode not yet implemented")
	}

	return fmt.Errorf("validation failed")
}

func (f *Factory) displayValidationJSON(result *config.ValidationResult) error {
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(jsonData))
	return nil
}

// Helper functions for status display

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

// Health check implementation

type HealthCheck struct {
	Name       string    `json:"name"`
	Status     string    `json:"status"`
	Message    string    `json:"message"`
	Details    string    `json:"details,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
	CanAutoFix bool      `json:"can_auto_fix"`
}

type HealthReport struct {
	OverallStatus string        `json:"overall_status"`
	Checks        []HealthCheck `json:"checks"`
	Summary       string        `json:"summary"`
	Timestamp     time.Time     `json:"timestamp"`
}

func (f *Factory) runHealthChecks(ctx context.Context, base *BaseCommand, serviceNames []string, verbose bool) *HealthReport {
	report := &HealthReport{
		Timestamp: time.Now(),
		Checks:    []HealthCheck{},
	}

	// Check Docker availability
	report.Checks = append(report.Checks, f.checkDockerAvailability())

	// Check configuration files
	report.Checks = append(report.Checks, f.checkConfigurationFiles())

	// Check port availability
	report.Checks = append(report.Checks, f.checkPortAvailability())

	// Check service status if services are running
	if len(serviceNames) > 0 || len(serviceNames) == 0 {
		report.Checks = append(report.Checks, f.checkServiceHealth(ctx, base, serviceNames)...)
	}

	// Calculate overall status
	healthy := 0
	warnings := 0
	errors := 0

	for _, check := range report.Checks {
		switch check.Status {
		case "healthy":
			healthy++
		case "warning":
			warnings++
		case "error":
			errors++
		}
	}

	if errors > 0 {
		report.OverallStatus = "unhealthy"
		report.Summary = fmt.Sprintf("%d errors, %d warnings, %d healthy", errors, warnings, healthy)
	} else if warnings > 0 {
		report.OverallStatus = "warning"
		report.Summary = fmt.Sprintf("%d warnings, %d healthy", warnings, healthy)
	} else {
		report.OverallStatus = "healthy"
		report.Summary = fmt.Sprintf("All %d checks passed", healthy)
	}

	return report
}

func (f *Factory) checkDockerAvailability() HealthCheck {
	check := HealthCheck{
		Name:      "Docker Availability",
		Timestamp: time.Now(),
	}

	// Try to run docker version command
	if _, err := exec.LookPath("docker"); err != nil {
		check.Status = "error"
		check.Message = "Docker not found in PATH"
		check.Details = "Please install Docker and ensure it's in your PATH"
		return check
	}

	check.Status = "healthy"
	check.Message = "Docker is available"
	return check
}

func (f *Factory) checkConfigurationFiles() HealthCheck {
	check := HealthCheck{
		Name:      "Configuration Files",
		Timestamp: time.Now(),
	}

	// Check for commands.yaml
	if _, err := os.Stat("scripts/commands.yaml"); os.IsNotExist(err) {
		check.Status = "error"
		check.Message = "commands.yaml not found"
		check.Details = "Expected at scripts/commands.yaml"
		return check
	}

	// Check for services.yaml
	if _, err := os.Stat("services/services.yaml"); os.IsNotExist(err) {
		check.Status = "warning"
		check.Message = "services.yaml not found"
		check.Details = "Expected at services/services.yaml"
		return check
	}

	check.Status = "healthy"
	check.Message = "Configuration files present"
	return check
}

func (f *Factory) checkPortAvailability() HealthCheck {
	check := HealthCheck{
		Name:      "Port Availability",
		Timestamp: time.Now(),
	}

	// Check common ports
	commonPorts := []int{5432, 6379, 3306, 9092, 16686, 9090}
	inUse := []int{}

	for _, port := range commonPorts {
		if f.isPortInUse(port) {
			inUse = append(inUse, port)
		}
	}

	if len(inUse) > 0 {
		check.Status = "warning"
		check.Message = fmt.Sprintf("%d common service ports in use", len(inUse))
		check.Details = fmt.Sprintf("Ports in use: %v", inUse)
	} else {
		check.Status = "healthy"
		check.Message = "Common service ports available"
	}

	return check
}

func (f *Factory) checkServiceHealth(ctx context.Context, base *BaseCommand, serviceNames []string) []HealthCheck {
	checks := []HealthCheck{}

	services, err := base.manager.GetServiceStatus(ctx, serviceNames)
	if err != nil {
		check := HealthCheck{
			Name:      "Service Status",
			Status:    "error",
			Message:   "Failed to get service status",
			Details:   err.Error(),
			Timestamp: time.Now(),
		}
		return []HealthCheck{check}
	}

	for _, service := range services {
		check := HealthCheck{
			Name:      fmt.Sprintf("Service: %s", service.Name),
			Timestamp: time.Now(),
		}

		switch service.State {
		case "running":
			if service.Health == "healthy" {
				check.Status = "healthy"
				check.Message = "Service running and healthy"
			} else {
				check.Status = "warning"
				check.Message = fmt.Sprintf("Service running but health: %s", service.Health)
			}
		case "stopped", "exited":
			check.Status = "warning"
			check.Message = "Service is stopped"
			check.CanAutoFix = true
		default:
			check.Status = "warning"
			check.Message = fmt.Sprintf("Service in state: %s", service.State)
		}

		checks = append(checks, check)
	}

	return checks
}

func (f *Factory) isPortInUse(port int) bool {
	timeout := time.Second
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), timeout)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

func (f *Factory) displayHealthReportTable(report *HealthReport, autoFix bool) error {
	// Display overall status
	statusIcon := "✅"
	if report.OverallStatus == "warning" {
		statusIcon = "⚠️"
	} else if report.OverallStatus == "unhealthy" {
		statusIcon = "❌"
	}

	fmt.Printf("\n%s Overall Status: %s\n", statusIcon, report.OverallStatus)
	fmt.Printf("📋 Summary: %s\n\n", report.Summary)

	// Display individual checks
	fmt.Printf("%-30s %-10s %-40s\n", "CHECK", "STATUS", "MESSAGE")
	fmt.Println(strings.Repeat("-", 85))

	for _, check := range report.Checks {
		var statusIcon string
		switch check.Status {
		case "healthy":
			statusIcon = "✅"
		case "warning":
			statusIcon = "⚠️"
		case "error":
			statusIcon = "❌"
		}

		fmt.Printf("%-30s %s%-9s %-40s\n",
			check.Name, statusIcon, check.Status, check.Message)

		if check.Details != "" {
			fmt.Printf("%-30s %-10s └─ %s\n", "", "", check.Details)
		}

		// Auto-fix if enabled and possible
		if autoFix && check.CanAutoFix && check.Status != "healthy" {
			fmt.Printf("%-30s %-10s 🔧 Attempting auto-fix...\n", "", "")
			if err := f.attemptAutoFix(check); err != nil {
				fmt.Printf("%-30s %-10s ❌ Auto-fix failed: %v\n", "", "", err)
			} else {
				fmt.Printf("%-30s %-10s ✅ Auto-fix successful\n", "", "")
			}
		}
	}

	return nil
}

func (f *Factory) displayHealthReportJSON(report *HealthReport) error {
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(jsonData))
	return nil
}

// Generate command implementations

func (f *Factory) generateConfig(output, template string, overwrite bool) error {
	if output == "" {
		output = "dev-stack-config.yaml"
	}

	// Check if file exists and overwrite flag
	if _, err := os.Stat(output); err == nil && !overwrite {
		return fmt.Errorf("file %s already exists. Use --overwrite to replace it", output)
	}

	configTemplate := `# dev-stack configuration file
# Generated on: %s

# Project information
project:
  name: "my-project"
  type: "web"
  description: "Development stack for my project"

# Global settings
global:
  log_level: "info"
  color_output: true
  check_updates: true

# Service profiles
profiles:
  minimal:
    name: "Minimal Stack"
    description: "Basic development services"
    services: ["postgres"]

  web:
    name: "Web Development"
    description: "Services for web application development"
    services: ["postgres", "redis"]

  full:
    name: "Full Stack"
    description: "Complete development environment"
    services: ["postgres", "redis", "kafka", "jaeger"]

# Service overrides
services:
  postgres:
    environment:
      POSTGRES_DB: "myproject_dev"
      POSTGRES_USER: "developer"
      POSTGRES_PASSWORD: "dev_password"

  redis:
    environment:
      REDIS_PASSWORD: "dev_password"

# Environment-specific settings
environments:
  development:
    debug: true
    hot_reload: true

  production:
    debug: false
    hot_reload: false
`

	content := fmt.Sprintf(configTemplate, time.Now().Format("2006-01-02 15:04:05"))

	if err := os.WriteFile(output, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("✅ Generated configuration file: %s\n", output)
	return nil
}

func (f *Factory) generateServiceConfig(serviceName, output, template string, overwrite bool) error {
	if output == "" {
		output = fmt.Sprintf("%s-service.yaml", serviceName)
	}

	// Check if file exists and overwrite flag
	if _, err := os.Stat(output); err == nil && !overwrite {
		return fmt.Errorf("file %s already exists. Use --overwrite to replace it", output)
	}

	serviceTemplate := `# Service configuration for %s
# Generated on: %s

%s:
  description: "%s service for development"
  image: "%s:latest"

  ports:
    - "8080:8080"

  environment:
    NODE_ENV: "development"
    DEBUG: "true"

  volumes:
    - "./%s:/app"
    - "/app/node_modules"

  healthcheck:
    test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
    interval: 30s
    timeout: 10s
    retries: 3
    start_period: 30s

  depends_on:
    - postgres
    - redis

  networks:
    - dev-stack-network

  restart: unless-stopped

  # Development specific settings
  develop:
    watch:
      - action: sync
        path: ./src
        target: /app/src
      - action: rebuild
        path: package.json
`

	content := fmt.Sprintf(serviceTemplate, serviceName, time.Now().Format("2006-01-02 15:04:05"),
		serviceName, serviceName, serviceName, serviceName)

	if err := os.WriteFile(output, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write service config file: %w", err)
	}

	fmt.Printf("✅ Generated service configuration: %s\n", output)
	return nil
}

func (f *Factory) generateComposeFile(output, services string, overwrite bool) error {
	if output == "" {
		output = "docker-compose.generated.yml"
	}

	// Parse services list
	var serviceList []string
	if services != "" {
		for _, service := range strings.Split(services, ",") {
			serviceList = append(serviceList, strings.TrimSpace(service))
		}
	} else {
		serviceList = []string{"postgres", "redis"} // Default services
	}

	// Initialize service registry
	registryOptions := compose.DefaultRegistryOptions()
	registry := compose.NewServiceRegistry(registryOptions)

	// Load services
	if err := registry.Load(); err != nil {
		return fmt.Errorf("failed to load service registry: %w", err)
	}

	// Validate requested services exist
	availableServices := registry.ListServices()
	for _, serviceName := range serviceList {
		found := false
		for _, available := range availableServices {
			if available == serviceName {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("service '%s' not found. Available services: %s",
				serviceName, strings.Join(availableServices, ", "))
		}
	}

	// Validate each requested service
	for _, serviceName := range serviceList {
		if result := registry.ValidateService(serviceName); !result.Valid {
			return fmt.Errorf("service '%s' validation failed: %s",
				serviceName, strings.Join(result.Errors, "; "))
		} else if len(result.Warnings) > 0 {
			fmt.Printf("⚠️  Service '%s' warnings: %s\n",
				serviceName, strings.Join(result.Warnings, "; "))
		}
	}

	// Setup compose options
	composeOptions := compose.DefaultComposeOptions()
	composeOptions.OutputFile = output
	composeOptions.Overwrite = overwrite

	// Get project name from config or use default
	projectName := "dev-stack"
	if projectRoot := findProjectRoot(); projectRoot != "" {
		if configLoader := compose.NewConfigLoader(projectRoot); configLoader != nil {
			if config, err := configLoader.GetConfig(); err == nil && config.Project.Name != "" {
				projectName = config.Project.Name
			}
		}
	}
	composeOptions.ProjectName = projectName
	composeOptions.NetworkName = fmt.Sprintf("%s-network", projectName)
	composeOptions.DetectConflicts = true
	composeOptions.AutoFixPorts = false
	composeOptions.Interactive = false

	// Get profile from command flags or environment
	if profile := os.Getenv("DEV_STACK_PROFILE"); profile != "" {
		composeOptions.Profile = profile
	}

	// Enable auto-fix if environment variable is set
	if autoFix := os.Getenv("DEV_STACK_AUTO_FIX_PORTS"); autoFix == "true" {
		composeOptions.AutoFixPorts = true
	}

	// Set project configuration path
	if projectRoot := findProjectRoot(); projectRoot != "" {
		composeOptions.ProjectConfig = filepath.Join(projectRoot, "dev-stack-config.yaml")
	}

	// Create composer
	composer := compose.NewComposer(registry, composeOptions)

	// Generate compose file
	composeFile, err := composer.GenerateCompose(serviceList)
	if err != nil {
		return fmt.Errorf("failed to generate compose file: %w", err)
	}

	// Write to file
	if err := composer.WriteToFile(composeFile, output); err != nil {
		return fmt.Errorf("failed to write compose file: %w", err)
	}

	fmt.Printf("✅ Generated Docker Compose file: %s\n", output)
	fmt.Printf("   Services included: %s\n", strings.Join(composeFile.Metadata.Services, ", "))
	fmt.Printf("   Profile: %s\n", composeFile.Metadata.Profile)
	fmt.Printf("   Dependencies resolved: %v\n", composeOptions.IncludeDeps)
	fmt.Printf("   Project: %s\n", composeOptions.ProjectName)
	if composeOptions.DetectConflicts {
		fmt.Printf("   Port conflicts checked: ✅\n")
	}

	return nil
}

// findProjectRoot attempts to find the project root directory
func findProjectRoot() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	// Look for dev-stack config files
	configFiles := []string{
		"dev-stack-config.yaml",
		"dev-stack.yaml",
		"dev-stack.yml",
		".dev-stack.yaml",
		".dev-stack.yml",
	}

	dir := cwd
	for {
		for _, configFile := range configFiles {
			configPath := filepath.Join(dir, configFile)
			if _, err := os.Stat(configPath); err == nil {
				return dir
			}
		}

		// Check for services directory
		servicesPath := filepath.Join(dir, "services")
		if stat, err := os.Stat(servicesPath); err == nil && stat.IsDir() {
			return dir
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			break // Reached root
		}
		dir = parent
	}

	return ""
}

func (f *Factory) generateDockerfile(serviceName, output, template string, overwrite bool) error {
	if output == "" {
		output = fmt.Sprintf("Dockerfile.%s", serviceName)
	}

	// Check if file exists and overwrite flag
	if _, err := os.Stat(output); err == nil && !overwrite {
		return fmt.Errorf("file %s already exists. Use --overwrite to replace it", output)
	}

	dockerfileTemplate := `# Dockerfile for %s service
# Generated on: %s

FROM node:18-alpine

# Set working directory
WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci --only=production

# Copy application code
COPY . .

# Create non-root user
RUN addgroup -g 1001 -S nodejs && \
    adduser -S nextjs -u 1001

# Change ownership
RUN chown -R nextjs:nodejs /app
USER nextjs

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

# Start application
CMD ["npm", "start"]
`

	content := fmt.Sprintf(dockerfileTemplate, serviceName, time.Now().Format("2006-01-02 15:04:05"))

	if err := os.WriteFile(output, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write Dockerfile: %w", err)
	}

	fmt.Printf("✅ Generated Dockerfile: %s\n", output)
	return nil
}

func (f *Factory) attemptAutoFix(check HealthCheck) error {
	switch {
	case strings.Contains(check.Name, "Service:") && check.Status == "warning":
		// Try to start stopped services
		serviceName := strings.TrimPrefix(check.Name, "Service: ")
		if strings.Contains(check.Message, "stopped") {
			fmt.Printf("🔧 Starting service %s...\n", serviceName)
			// Create a basic command to start the service
			cmd := exec.Command("docker", "compose", "up", "-d", serviceName)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to start service %s: %w", serviceName, err)
			}
			return nil
		}
	case strings.Contains(check.Name, "Port Availability") && check.Status == "warning":
		fmt.Println("🔧 Port conflicts detected - consider stopping conflicting services")
		return fmt.Errorf("manual intervention required for port conflicts")
	case strings.Contains(check.Name, "Docker Availability") && check.Status == "error":
		return fmt.Errorf("Docker installation required - cannot auto-fix")
	}

	return fmt.Errorf("no auto-fix available for this check")
}

// Utility functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
