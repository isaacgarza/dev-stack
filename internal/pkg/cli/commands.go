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
	pkgTypes "github.com/isaacgarza/dev-stack/internal/pkg/types"
	"github.com/spf13/cobra"
)

// NewUpCommand creates the up command
func NewUpCommand(serviceManager *services.Manager, logger *slog.Logger) *cobra.Command {
	handler := core.NewUpHandler()

	cmd := &cobra.Command{
		Use:   "up [services...]",
		Short: "Start services",
		Long:  "Start the specified services or all services if none specified",
		RunE: func(cmd *cobra.Command, args []string) error {
			base := createBaseCommand(serviceManager, logger)
			return handler.Handle(context.Background(), cmd, args, base)
		},
	}

	cmd.Flags().Bool("build", false, "Build images before starting")
	cmd.Flags().Bool("force-recreate", false, "Force recreate containers")

	return cmd
}

// NewDownCommand creates the down command
func NewDownCommand(serviceManager *services.Manager, logger *slog.Logger) *cobra.Command {
	handler := core.NewDownHandler()

	cmd := &cobra.Command{
		Use:   "down [services...]",
		Short: "Stop services",
		Long:  "Stop the specified services or all services if none specified",
		RunE: func(cmd *cobra.Command, args []string) error {
			base := createBaseCommand(serviceManager, logger)
			return handler.Handle(context.Background(), cmd, args, base)
		},
	}

	cmd.Flags().Int("timeout", 10, "Timeout in seconds")
	cmd.Flags().Bool("remove", true, "Remove containers")
	cmd.Flags().Bool("volumes", false, "Remove volumes")

	return cmd
}

// NewStatusCommand creates the status command
func NewStatusCommand(serviceManager *services.Manager, logger *slog.Logger) *cobra.Command {
	handler := core.NewStatusHandler()

	cmd := &cobra.Command{
		Use:   "status [services...]",
		Short: "Show service status",
		Long:  "Show the status of specified services or all services if none specified",
		RunE: func(cmd *cobra.Command, args []string) error {
			base := createBaseCommand(serviceManager, logger)
			return handler.Handle(context.Background(), cmd, args, base)
		},
	}

	cmd.Flags().String("format", "table", "Output format (table, json, yaml)")

	return cmd
}

// NewRestartCommand creates the restart command
func NewRestartCommand(serviceManager *services.Manager, logger *slog.Logger) *cobra.Command {
	handler := core.NewRestartHandler()

	cmd := &cobra.Command{
		Use:   "restart [services...]",
		Short: "Restart services",
		Long:  "Restart the specified services or all services if none specified",
		RunE: func(cmd *cobra.Command, args []string) error {
			base := createBaseCommand(serviceManager, logger)
			return handler.Handle(context.Background(), cmd, args, base)
		},
	}

	cmd.Flags().Int("timeout", 10, "Timeout in seconds")

	return cmd
}

// NewLogsCommand creates the logs command
func NewLogsCommand(serviceManager *services.Manager, logger *slog.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs [services...]",
		Short: "Show service logs",
		Long:  "Show logs for specified services or all services if none specified",
		RunE: func(cmd *cobra.Command, args []string) error {
			follow, _ := cmd.Flags().GetBool("follow")
			tail, _ := cmd.Flags().GetString("tail")

			ctx := context.Background()
			options := pkgTypes.LogOptions{
				Follow:     follow,
				Tail:       tail,
				Timestamps: true,
			}

			return serviceManager.GetLogs(ctx, args, options)
		},
	}

	cmd.Flags().BoolP("follow", "f", false, "Follow log output")
	cmd.Flags().String("tail", "100", "Number of lines to show from end of logs")

	return cmd
}

// NewExecCommand creates the exec command
func NewExecCommand(serviceManager *services.Manager, logger *slog.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exec <service> <command>",
		Short: "Execute command in service container",
		Long:  "Execute a command in a running service container",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("requires service name and command")
			}

			serviceName := args[0]
			command := args[1:]

			interactive, _ := cmd.Flags().GetBool("interactive")
			tty, _ := cmd.Flags().GetBool("tty")

			ctx := context.Background()
			options := pkgTypes.ExecOptions{
				Interactive: interactive,
				TTY:         tty,
			}

			return serviceManager.ExecCommand(ctx, serviceName, command, options)
		},
	}

	cmd.Flags().BoolP("interactive", "i", false, "Interactive mode")
	cmd.Flags().BoolP("tty", "t", false, "Allocate a pseudo-TTY")

	return cmd
}

// NewInitCommand creates the init command
func NewInitCommand(logger *slog.Logger) *cobra.Command {
	handler := initHandler.NewInitHandler()

	cmd := &cobra.Command{
		Use:   "init [template]",
		Short: "Initialize a new project",
		Long:  "Initialize a new dev-stack project with optional template",
		RunE: func(cmd *cobra.Command, args []string) error {
			base := &cliTypes.BaseCommand{Logger: &loggerAdapter{logger: logger}}
			return handler.Handle(context.Background(), cmd, args, base)
		},
	}

	cmd.Flags().String("name", "", "Project name")
	cmd.Flags().String("template", "", "Template to use")

	return cmd
}

// NewConfigCommand creates the config command
func NewConfigCommand(logger *slog.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Show configuration",
		Long:  "Show the current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("dev-stack Configuration")
			fmt.Println("======================")
			fmt.Printf("Version: %s\n", cmd.Root().Version)
			fmt.Printf("Config file: %s\n", "dev-stack-config.yaml")
			fmt.Println("\nUse 'dev-stack services' to see available services")
			fmt.Println("Use 'dev-stack status' to see running services")
			return nil
		},
	}

	return cmd
}

// NewServicesCommand creates the services command
func NewServicesCommand(serviceManager *services.Manager, logger *slog.Logger) *cobra.Command {
	handler := cliServices.NewServicesHandler()

	cmd := &cobra.Command{
		Use:   "services",
		Short: "List available services",
		Long:  "List all available services and their descriptions",
		RunE: func(cmd *cobra.Command, args []string) error {
			base := &cliTypes.BaseCommand{
				Manager: &serviceManagerAdapter{manager: serviceManager},
				Logger:  &loggerAdapter{logger: logger},
			}
			return handler.Handle(context.Background(), cmd, args, base)
		},
	}

	return cmd
}

// createBaseCommand creates a base command with common dependencies
func createBaseCommand(serviceManager *services.Manager, logger *slog.Logger) *cliTypes.BaseCommand {
	return &cliTypes.BaseCommand{
		Manager: &serviceManagerAdapter{manager: serviceManager},
		Logger:  &loggerAdapter{logger: logger},
	}
}

// serviceManagerAdapter adapts services.Manager to types.ServiceManager interface
type serviceManagerAdapter struct {
	manager *services.Manager
}

func (s *serviceManagerAdapter) StartServices(ctx context.Context, serviceNames []string, options cliTypes.StartOptions) error {
	// Convert CLI types to internal types
	internalOptions := pkgTypes.StartOptions{
		Build:         options.Build,
		ForceRecreate: options.ForceRecreate,
		Detach:        true,
	}
	return s.manager.StartServices(ctx, serviceNames, internalOptions)
}

func (s *serviceManagerAdapter) StopServices(ctx context.Context, serviceNames []string, options cliTypes.StopOptions) error {
	internalOptions := pkgTypes.StopOptions{
		Timeout:       options.Timeout,
		Remove:        true,
		RemoveVolumes: options.Volumes,
	}
	return s.manager.StopServices(ctx, serviceNames, internalOptions)
}

func (s *serviceManagerAdapter) GetServiceStatus(ctx context.Context, serviceNames []string) ([]cliTypes.ServiceStatus, error) {
	// Get status from actual manager and convert types
	statuses, err := s.manager.GetServiceStatus(ctx, serviceNames)
	if err != nil {
		return nil, err
	}

	// Convert from internal types to CLI types
	result := make([]cliTypes.ServiceStatus, len(statuses))
	for i, status := range statuses {
		result[i] = cliTypes.ServiceStatus{
			Name:   status.Name,
			Status: status.State.String(),
			Health: status.Health.String(),
			Uptime: status.Uptime.String(),
		}
	}
	return result, nil
}

func (s *serviceManagerAdapter) Close() error {
	return s.manager.Close()
}

// loggerAdapter adapts slog.Logger to types.Logger interface
type loggerAdapter struct {
	logger *slog.Logger
}

func (l *loggerAdapter) Info(msg string, args ...interface{}) {
	l.logger.Info(msg, args...)
}

func (l *loggerAdapter) Error(msg string, args ...interface{}) {
	l.logger.Error(msg, args...)
}

func (l *loggerAdapter) Debug(msg string, args ...interface{}) {
	l.logger.Debug(msg, args...)
}

// NewDoctorCommand creates the doctor command
func NewDoctorCommand(logger *slog.Logger) *cobra.Command {
	handler := doctor.NewDoctorHandler()

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Run health checks",
		Long:  "Run comprehensive health checks on your dev-stack environment",
		RunE: func(cmd *cobra.Command, args []string) error {
			base := &cliTypes.BaseCommand{
				Logger: &loggerAdapter{logger: logger},
			}
			return handler.Handle(context.Background(), cmd, args, base)
		},
	}

	return cmd
}

// NewCompletionCommand creates the completion command
func NewCompletionCommand(logger *slog.Logger) *cobra.Command {
	handler := completion.NewCompletionHandler()

	cmd := &cobra.Command{
		Use:       "completion [bash|zsh|fish|powershell]",
		Short:     "Generate completion script",
		Long:      "Generate completion script for your shell",
		ValidArgs: pkgTypes.AllShellTypeStrings(),
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			base := &cliTypes.BaseCommand{
				Logger: &loggerAdapter{logger: logger},
			}
			return handler.Handle(context.Background(), cmd, args, base)
		},
	}

	return cmd
}
