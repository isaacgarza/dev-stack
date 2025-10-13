package version

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/isaacgarza/dev-stack/internal/pkg/cli/handlers/utils"
	"github.com/isaacgarza/dev-stack/internal/pkg/cli/types"
	"github.com/isaacgarza/dev-stack/internal/pkg/constants"
	"github.com/isaacgarza/dev-stack/internal/pkg/version"
	"github.com/spf13/cobra"
)

// EnforcementHandler handles version enforcement commands
type EnforcementHandler struct {
	enforcer *version.VersionEnforcer
	notifier *version.UpdateNotifier
}

// NewEnforcementHandler creates a new enforcement handler
func NewEnforcementHandler(manager version.VersionManager) *EnforcementHandler {
	policy := version.EnforcementPolicy{
		StrictMode:       constants.DefaultStrictMode,
		AllowDrift:       constants.DefaultAllowDrift,
		MaxDriftDuration: constants.DefaultMaxDriftDuration,
		AutoSync:         constants.DefaultAutoSync,
		NotifyUpdates:    constants.DefaultNotifyUpdates,
	}

	enforcer := version.NewVersionEnforcer(manager, policy)

	// Use proper config path with constants
	configPath := filepath.Join(os.Getenv("HOME"), ".dev-stack", constants.NotificationConfigFile)
	notifier := version.NewUpdateNotifier(manager, configPath)

	return &EnforcementHandler{
		enforcer: enforcer,
		notifier: notifier,
	}
}

// HandleCheck handles the version check command
func (h *EnforcementHandler) HandleCheck(ctx context.Context, cmd *cobra.Command, args []string, base *types.BaseCommand) error {
	flags := utils.GetCIFlags(cmd)
	projectPath := "."

	if len(args) > 0 {
		projectPath = args[0]
	}

	result, err := h.enforcer.CheckCompliance(projectPath)
	if err != nil {
		utils.HandleError(flags, fmt.Errorf("compliance check failed: %w", err))
		return nil
	}

	if flags.JSON {
		return json.NewEncoder(os.Stdout).Encode(result)
	}

	h.displayComplianceResult(result)

	if result.ExitCode != constants.ExitSuccess {
		os.Exit(result.ExitCode)
	}

	return nil
}

// HandleEnforce handles the version enforce command
func (h *EnforcementHandler) HandleEnforce(ctx context.Context, cmd *cobra.Command, args []string, base *types.BaseCommand) error {
	flags := utils.GetCIFlags(cmd)
	projectPath := "."

	if len(args) > 0 {
		projectPath = args[0]
	}

	err := h.enforcer.EnforceCompliance(projectPath)
	if err != nil {
		utils.HandleError(flags, fmt.Errorf("enforcement failed: %w", err))
		return nil
	}

	if !flags.Quiet {
		fmt.Println("✅ Version compliance enforced")
	}

	return nil
}

// HandleDrift handles the version drift command
func (h *EnforcementHandler) HandleDrift(ctx context.Context, cmd *cobra.Command, args []string, base *types.BaseCommand) error {
	flags := utils.GetCIFlags(cmd)

	drifts, err := h.enforcer.DetectAllDrift()
	if err != nil {
		utils.HandleError(flags, fmt.Errorf("drift detection failed: %w", err))
		return nil
	}

	if flags.JSON {
		return json.NewEncoder(os.Stdout).Encode(drifts)
	}

	h.displayDriftResults(drifts)
	return nil
}

// HandleNotify handles update notification commands
func (h *EnforcementHandler) HandleNotify(ctx context.Context, cmd *cobra.Command, args []string, base *types.BaseCommand) error {
	if len(args) == 0 {
		return h.handleNotifyCheck(cmd)
	}

	switch args[0] {
	case "check":
		return h.handleNotifyCheck(cmd)
	case "config":
		return h.handleNotifyConfig(cmd, args[1:])
	case "suppress":
		return h.handleNotifySuppress(cmd, args[1:])
	default:
		return fmt.Errorf("unknown notify command: %s", args[0])
	}
}

func (h *EnforcementHandler) handleNotifyCheck(cmd *cobra.Command) error {
	flags := utils.GetCIFlags(cmd)

	notification, err := h.notifier.CheckForUpdates()
	if err != nil {
		utils.HandleError(flags, fmt.Errorf("update check failed: %w", err))
		return nil
	}

	if flags.JSON {
		return json.NewEncoder(os.Stdout).Encode(notification)
	}

	if notification == nil {
		if !flags.Quiet {
			fmt.Println("✅ No updates available")
		}
	} else {
		fmt.Printf("🔔 Update available: %s → %s (%s)\n",
			notification.CurrentVersion, notification.LatestVersion, notification.Severity)
		fmt.Printf("   %s\n", notification.Message)
	}

	return nil
}

func (h *EnforcementHandler) handleNotifyConfig(cmd *cobra.Command, args []string) error {
	flags := utils.GetCIFlags(cmd)

	if len(args) == 0 {
		// Show current config
		config := h.notifier.GetConfig()
		if flags.JSON {
			return json.NewEncoder(os.Stdout).Encode(config)
		}

		h.displayNotificationConfig(config)
		return nil
	}

	// Set config values
	config := h.notifier.GetConfig()

	for _, arg := range args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid config format: %s (expected key=value)", arg)
		}

		key, value := parts[0], parts[1]

		switch key {
		case "enabled":
			config.Enabled = value == "true"
		case "frequency":
			config.Frequency = value
		case "min-severity":
			config.MinSeverity = value
		case "auto-check":
			config.AutoCheck = value == "true"
		case "show-prerelease":
			config.ShowPrerelease = value == "true"
		default:
			return fmt.Errorf("unknown config key: %s", key)
		}
	}

	err := h.notifier.SetConfig(config)
	if err != nil {
		utils.HandleError(flags, fmt.Errorf("failed to save config: %w", err))
		return nil
	}

	if !flags.Quiet {
		fmt.Println("✅ Notification config updated")
	}

	return nil
}

func (h *EnforcementHandler) handleNotifySuppress(cmd *cobra.Command, args []string) error {
	flags := utils.GetCIFlags(cmd)

	if len(args) == 0 {
		return fmt.Errorf("suppress duration required (e.g., 1d, 1w, 1h)")
	}

	duration, err := parseDuration(args[0])
	if err != nil {
		utils.HandleError(flags, fmt.Errorf("invalid duration: %w", err))
		return nil
	}

	err = h.notifier.SuppressNotifications(duration)
	if err != nil {
		utils.HandleError(flags, fmt.Errorf("failed to suppress notifications: %w", err))
		return nil
	}

	if !flags.Quiet {
		fmt.Printf("✅ Notifications suppressed for %v\n", duration)
	}

	return nil
}

func (h *EnforcementHandler) displayComplianceResult(result *version.EnforcementResult) {
	if result.Compliant {
		fmt.Println("✅ Version compliance satisfied")
		return
	}

	fmt.Printf("❌ Version compliance failed\n")
	fmt.Printf("   %s\n", result.Message)

	if result.Drift != nil {
		fmt.Printf("   Required: %s\n", result.Drift.RequiredVersion)
		fmt.Printf("   Active:   %s\n", result.Drift.ActiveVersion)
		fmt.Printf("   Drift:    %s (%s)\n", result.Drift.DriftType, result.Drift.Severity)
		fmt.Printf("   Duration: %v\n", result.Drift.DriftDuration)
	}

	fmt.Printf("   Action:   %s\n", result.Action)
}

func (h *EnforcementHandler) displayDriftResults(drifts []version.DriftDetection) {
	if len(drifts) == 0 {
		fmt.Println("✅ No version drift detected")
		return
	}

	fmt.Printf("⚠️  Version drift detected in %d projects:\n\n", len(drifts))

	for _, drift := range drifts {
		fmt.Printf("📁 %s\n", drift.ProjectPath)
		fmt.Printf("   Required: %s\n", drift.RequiredVersion)
		fmt.Printf("   Active:   %s\n", drift.ActiveVersion)
		fmt.Printf("   Type:     %s (%s)\n", drift.DriftType, drift.Severity)
		fmt.Printf("   Duration: %v\n", drift.DriftDuration)
		fmt.Println()
	}
}

func (h *EnforcementHandler) displayNotificationConfig(config version.NotificationConfig) {
	fmt.Printf("Notification Configuration:\n")
	fmt.Printf("  Enabled:         %t\n", config.Enabled)
	fmt.Printf("  Frequency:       %s\n", config.Frequency)
	fmt.Printf("  Min Severity:    %s\n", config.MinSeverity)
	fmt.Printf("  Auto Check:      %t\n", config.AutoCheck)
	fmt.Printf("  Show Prerelease: %t\n", config.ShowPrerelease)
	fmt.Printf("  Last Check:      %s\n", config.LastCheck.Format(time.RFC3339))

	if !config.SuppressedUntil.IsZero() {
		fmt.Printf("  Suppressed Until: %s\n", config.SuppressedUntil.Format(time.RFC3339))
	}
}

func parseDuration(s string) (time.Duration, error) {
	// Handle simple formats like "1d", "2w", "3h"
	if len(s) < 2 {
		return 0, fmt.Errorf("invalid duration format")
	}

	unit := s[len(s)-1:]
	valueStr := s[:len(s)-1]

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return time.ParseDuration(s) // Fall back to standard parsing
	}

	switch unit {
	case "d":
		return time.Duration(value) * 24 * time.Hour, nil
	case "w":
		return time.Duration(value) * 7 * 24 * time.Hour, nil
	case "h":
		return time.Duration(value) * time.Hour, nil
	case "m":
		return time.Duration(value) * time.Minute, nil
	default:
		return time.ParseDuration(s)
	}
}
