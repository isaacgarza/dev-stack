package version

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/isaacgarza/dev-stack/internal/pkg/constants"
)

// UpdateNotification represents an update notification
type UpdateNotification struct {
	CurrentVersion   Version   `json:"current_version"`
	LatestVersion    Version   `json:"latest_version"`
	UpdateType       string    `json:"update_type"` // constants.UpdateType*
	Severity         string    `json:"severity"`    // constants.UpdateSeverity*
	ReleaseDate      time.Time `json:"release_date"`
	ChangelogURL     string    `json:"changelog_url"`
	DownloadURL      string    `json:"download_url"`
	SecurityUpdate   bool      `json:"security_update"`
	BreakingChanges  bool      `json:"breaking_changes"`
	Message          string    `json:"message"`
	LastNotified     time.Time `json:"last_notified"`
	NotificationFreq string    `json:"notification_freq"` // constants.NotificationFrequency*
}

// NotificationConfig controls update notification behavior
type NotificationConfig struct {
	Enabled         bool          `json:"enabled"`
	Frequency       string        `json:"frequency"`      // constants.NotificationFrequency*
	MinSeverity     string        `json:"min_severity"`   // constants.UpdateSeverity*
	CheckInterval   time.Duration `json:"check_interval"` // How often to check for updates
	LastCheck       time.Time     `json:"last_check"`
	AutoCheck       bool          `json:"auto_check"`       // Check on command execution
	ShowPrerelease  bool          `json:"show_prerelease"`  // Include prerelease versions
	SuppressedUntil time.Time     `json:"suppressed_until"` // Suppress notifications until this time
}

// UpdateNotifier handles update notifications
type UpdateNotifier struct {
	manager    VersionManager
	config     NotificationConfig
	configPath string
}

// NewUpdateNotifier creates a new update notifier
func NewUpdateNotifier(manager VersionManager, configPath string) *UpdateNotifier {
	notifier := &UpdateNotifier{
		manager:    manager,
		configPath: configPath,
		config: NotificationConfig{
			Enabled:        constants.DefaultNotifyUpdates,
			Frequency:      constants.NotificationFrequencyDaily,
			MinSeverity:    constants.UpdateSeverityRecommended,
			CheckInterval:  constants.DefaultCheckInterval,
			AutoCheck:      constants.DefaultAutoCheck,
			ShowPrerelease: constants.DefaultShowPrerelease,
		},
	}

	notifier.loadConfig()
	return notifier
}

// CheckForUpdates checks for available updates
func (n *UpdateNotifier) CheckForUpdates() (*UpdateNotification, error) {
	if !n.config.Enabled {
		return nil, nil
	}

	// Check if we should skip based on frequency
	if !n.shouldCheck() {
		return nil, nil
	}

	currentVersion, err := n.manager.GetActiveVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get current version: %w", err)
	}

	availableVersions, err := n.manager.ListAvailableVersions()
	if err != nil {
		return nil, fmt.Errorf("failed to list available versions: %w", err)
	}

	// Find latest version
	var latestVersion *Version
	for _, version := range availableVersions {
		if !n.config.ShowPrerelease && version.PreRelease != "" {
			continue
		}

		if latestVersion == nil || version.Compare(*latestVersion) > 0 {
			latestVersion = &version
		}
	}

	if latestVersion == nil {
		return nil, nil
	}

	// Check if update is available
	if currentVersion.Version.Compare(*latestVersion) >= 0 {
		n.updateLastCheck()
		return nil, nil
	}

	// Create notification
	notification := &UpdateNotification{
		CurrentVersion: currentVersion.Version,
		LatestVersion:  *latestVersion,
		ReleaseDate:    time.Now(), // Would be fetched from release info
		LastNotified:   time.Now(),
	}

	// Determine update type and severity
	n.categorizeUpdate(notification)

	// Check if we should notify based on severity
	if !n.shouldNotify(notification) {
		return nil, nil
	}

	n.updateLastCheck()
	return notification, nil
}

// NotifyIfNeeded checks for updates and shows notification if needed
func (n *UpdateNotifier) NotifyIfNeeded() error {
	notification, err := n.CheckForUpdates()
	if err != nil {
		return err
	}

	if notification != nil {
		n.showNotification(notification)
	}

	return nil
}

// SuppressNotifications suppresses notifications for a duration
func (n *UpdateNotifier) SuppressNotifications(duration time.Duration) error {
	n.config.SuppressedUntil = time.Now().Add(duration)
	return n.saveConfig()
}

// SetConfig updates the notification configuration
func (n *UpdateNotifier) SetConfig(config NotificationConfig) error {
	n.config = config
	return n.saveConfig()
}

// GetConfig returns the current notification configuration
func (n *UpdateNotifier) GetConfig() NotificationConfig {
	return n.config
}

func (n *UpdateNotifier) shouldCheck() bool {
	if time.Now().Before(n.config.SuppressedUntil) {
		return false
	}

	switch n.config.Frequency {
	case constants.NotificationFrequencyNever:
		return false
	case constants.NotificationFrequencyAlways:
		return true
	case constants.NotificationFrequencyDaily:
		return time.Since(n.config.LastCheck) >= 24*time.Hour
	case constants.NotificationFrequencyWeekly:
		return time.Since(n.config.LastCheck) >= 7*24*time.Hour
	default:
		return time.Since(n.config.LastCheck) >= n.config.CheckInterval
	}
}

func (n *UpdateNotifier) shouldNotify(notification *UpdateNotification) bool {
	severityLevel := map[string]int{
		constants.UpdateSeverityOptional:    1,
		constants.UpdateSeverityRecommended: 2,
		constants.UpdateSeverityCritical:    3,
	}

	minLevel := severityLevel[n.config.MinSeverity]
	notificationLevel := severityLevel[notification.Severity]

	return notificationLevel >= minLevel
}

func (n *UpdateNotifier) categorizeUpdate(notification *UpdateNotification) {
	current := notification.CurrentVersion
	latest := notification.LatestVersion

	if latest.Major > current.Major {
		notification.UpdateType = constants.UpdateTypeMajor
		notification.Severity = constants.UpdateSeverityRecommended
		notification.BreakingChanges = true
		notification.Message = fmt.Sprintf("Major update available: %s → %s (may contain breaking changes)",
			current, latest)
	} else if latest.Minor > current.Minor {
		notification.UpdateType = constants.UpdateTypeMinor
		notification.Severity = constants.UpdateSeverityRecommended
		notification.Message = fmt.Sprintf("Minor update available: %s → %s (new features)",
			current, latest)
	} else if latest.Patch > current.Patch {
		notification.UpdateType = constants.UpdateTypePatch
		notification.Severity = constants.UpdateSeverityOptional
		notification.Message = fmt.Sprintf("Patch update available: %s → %s (bug fixes)",
			current, latest)
	} else {
		notification.UpdateType = constants.UpdateTypePrerelease
		notification.Severity = constants.UpdateSeverityOptional
		notification.Message = fmt.Sprintf("Prerelease update available: %s → %s",
			current, latest)
	}

	// Set URLs (would be populated from actual release data)
	notification.ChangelogURL = fmt.Sprintf("https://github.com/isaacgarza/dev-stack/releases/tag/v%s", latest)
	notification.DownloadURL = fmt.Sprintf("https://github.com/isaacgarza/dev-stack/releases/download/v%s/dev-stack", latest)
}

func (n *UpdateNotifier) showNotification(notification *UpdateNotification) {
	fmt.Printf("\n🔔 Update Available!\n")
	fmt.Printf("   Current: %s\n", notification.CurrentVersion)
	fmt.Printf("   Latest:  %s\n", notification.LatestVersion)
	fmt.Printf("   Type:    %s (%s)\n", notification.UpdateType, notification.Severity)

	if notification.BreakingChanges {
		fmt.Printf("   ⚠️  May contain breaking changes\n")
	}

	if notification.SecurityUpdate {
		fmt.Printf("   🔒 Security update\n")
	}

	fmt.Printf("\n   %s\n", notification.Message)
	fmt.Printf("\n   To update: dev-stack version install %s\n", notification.LatestVersion)
	fmt.Printf("   Changelog: %s\n", notification.ChangelogURL)
	fmt.Printf("\n   To suppress: dev-stack version suppress 7d\n\n")
}

func (n *UpdateNotifier) updateLastCheck() {
	n.config.LastCheck = time.Now()
	_ = n.saveConfig() // Ignore errors for non-critical operation
}

func (n *UpdateNotifier) loadConfig() {
	if n.configPath == "" {
		return
	}

	data, err := os.ReadFile(n.configPath)
	if err != nil {
		return // Use defaults
	}

	_ = json.Unmarshal(data, &n.config)
}

func (n *UpdateNotifier) saveConfig() error {
	if n.configPath == "" {
		return nil
	}

	// Ensure directory exists
	dir := filepath.Dir(n.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(n.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(n.configPath, data, 0644)
}
