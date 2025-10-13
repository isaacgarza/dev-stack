package cli

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/isaacgarza/dev-stack/internal/core/services"
	"github.com/isaacgarza/dev-stack/internal/pkg/cli/handlers/completion"
	"github.com/isaacgarza/dev-stack/internal/pkg/cli/handlers/core"
	"github.com/isaacgarza/dev-stack/internal/pkg/cli/handlers/doctor"
	initHandler "github.com/isaacgarza/dev-stack/internal/pkg/cli/handlers/init"
	cliServices "github.com/isaacgarza/dev-stack/internal/pkg/cli/handlers/services"
	cliTypes "github.com/isaacgarza/dev-stack/internal/pkg/cli/types"
	"github.com/isaacgarza/dev-stack/internal/pkg/config"
	"github.com/isaacgarza/dev-stack/internal/pkg/constants"
	"github.com/spf13/cobra"
)

// BuildDynamicRootCommand creates commands from YAML configuration
func BuildDynamicRootCommand(config *config.CommandConfig) (*cobra.Command, error) {
	log := slog.Default()

	rootCmd := &cobra.Command{
		Use:     constants.AppName,
		Short:   config.Metadata.Description,
		Version: config.Metadata.CLIVersion,
		Long:    fmt.Sprintf("%s\n\nVersion: %s", config.Metadata.Description, config.Metadata.CLIVersion),
	}

	// Add global flags from config
	if err := addGlobalFlagsFromConfig(rootCmd, config); err != nil {
		return nil, fmt.Errorf("failed to add global flags: %w", err)
	}

	serviceManager, err := createServiceManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create service manager: %w", err)
	}

	// Build commands dynamically from config
	for cmdName, cmdConfig := range config.Commands {
		cmd, err := buildCommandFromConfig(cmdName, cmdConfig, serviceManager, log)
		if err != nil {
			return nil, fmt.Errorf("failed to build command %s: %w", cmdName, err)
		}
		rootCmd.AddCommand(cmd)
	}

	return rootCmd, nil
}

func buildCommandFromConfig(name string, cmdConfig config.Command, serviceManager *services.Manager, logger *slog.Logger) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   cmdConfig.Usage,
		Short: cmdConfig.Description,
		Long:  cmdConfig.LongDescription,
	}

	// Add aliases
	if len(cmdConfig.Aliases) > 0 {
		cmd.Aliases = cmdConfig.Aliases
	}

	// Add flags from config
	for flagName, flagConfig := range cmdConfig.Flags {
		addFlagFromConfig(cmd, flagName, flagConfig)
	}

	// Set up command handler based on name
	handler := getHandlerForCommand(name, serviceManager)
	if handler != nil {
		cmd.RunE = func(cmd *cobra.Command, args []string) error {
			base := &cliTypes.BaseCommand{
				Logger: &loggerAdapter{logger: logger},
			}
			return handler.Handle(context.Background(), cmd, args, base)
		}
	}

	// Add examples
	if len(cmdConfig.Examples) > 0 {
		cmd.Example = buildExamplesString(cmdConfig.Examples)
	}

	return cmd, nil
}

func getHandlerForCommand(name string, serviceManager *services.Manager) cliTypes.CommandHandler {
	switch name {
	case constants.CmdNameUp:
		return core.NewUpHandler()
	case constants.CmdNameDown:
		return core.NewDownHandler()
	case constants.CmdNameRestart:
		return core.NewRestartHandler()
	case constants.CmdNameStatus:
		return core.NewStatusHandler()
	case constants.CmdNameInit:
		return initHandler.NewInitHandler()
	case constants.CmdNameDoctor:
		return doctor.NewDoctorHandler()
	case constants.CmdNameCompletion:
		return completion.NewCompletionHandler()
	case constants.CmdNameServices:
		return cliServices.NewServicesHandler()
	case constants.CmdNameDeps:
		return cliServices.NewDepsHandler()
	case constants.CmdNameConflicts:
		return cliServices.NewConflictsHandler()
	default:
		return nil
	}
}

func addGlobalFlagsFromConfig(cmd *cobra.Command, config *config.CommandConfig) error {
	for name, flag := range config.Global.Flags {
		switch flag.Type {
		case "bool":
			cmd.PersistentFlags().Bool(name, flag.Default.(bool), flag.Description)
		case "string":
			defaultVal := ""
			if flag.Default != nil {
				defaultVal = flag.Default.(string)
			}
			cmd.PersistentFlags().String(name, defaultVal, flag.Description)
		case "int":
			defaultVal := 0
			if flag.Default != nil {
				defaultVal = flag.Default.(int)
			}
			cmd.PersistentFlags().Int(name, defaultVal, flag.Description)
		}

		if flag.Short != "" {
			if pf := cmd.PersistentFlags().Lookup(name); pf != nil {
				pf.Shorthand = flag.Short
			}
		}
	}
	return nil
}

func addFlagFromConfig(cmd *cobra.Command, name string, flag config.Flag) {
	switch flag.Type {
	case "bool":
		defaultVal := false
		if flag.Default != nil {
			defaultVal = flag.Default.(bool)
		}
		cmd.Flags().Bool(name, defaultVal, flag.Description)
	case "string":
		defaultVal := ""
		if flag.Default != nil {
			defaultVal = flag.Default.(string)
		}
		cmd.Flags().String(name, defaultVal, flag.Description)
	case "int":
		defaultVal := 0
		if flag.Default != nil {
			defaultVal = flag.Default.(int)
		}
		cmd.Flags().Int(name, defaultVal, flag.Description)
	}

	if flag.Short != "" {
		if f := cmd.Flags().Lookup(name); f != nil {
			f.Shorthand = flag.Short
		}
	}
}

func buildExamplesString(examples []config.Example) string {
	var result string
	for _, example := range examples {
		result += fmt.Sprintf("  %s\n    %s\n\n", example.Command, example.Description)
	}
	return result
}
