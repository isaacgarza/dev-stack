package version

import (
	"fmt"
	"time"

	"github.com/isaacgarza/dev-stack/internal/pkg/constants"
)

// EnforcementPolicy defines how strict version enforcement should be
type EnforcementPolicy struct {
	StrictMode       bool          `json:"strict_mode"`
	AllowDrift       bool          `json:"allow_drift"`
	MaxDriftDuration time.Duration `json:"max_drift_duration"`
	AutoSync         bool          `json:"auto_sync"`
	NotifyUpdates    bool          `json:"notify_updates"`
}

// DriftDetection represents version drift information
type DriftDetection struct {
	ProjectPath     string        `json:"project_path"`
	RequiredVersion Version       `json:"required_version"`
	ActiveVersion   Version       `json:"active_version"`
	DriftDuration   time.Duration `json:"drift_duration"`
	DriftType       string        `json:"drift_type"` // constants.DriftType*
	Severity        string        `json:"severity"`   // constants.Severity*
}

// EnforcementResult represents the result of version enforcement
type EnforcementResult struct {
	Compliant bool            `json:"compliant"`
	Drift     *DriftDetection `json:"drift,omitempty"`
	Action    string          `json:"action"` // constants.EnforcementAction*
	Message   string          `json:"message"`
	ExitCode  int             `json:"exit_code"`
}

// VersionEnforcer handles strict version checking and enforcement
type VersionEnforcer struct {
	manager VersionManager
	policy  EnforcementPolicy
}

// NewVersionEnforcer creates a new version enforcer
func NewVersionEnforcer(manager VersionManager, policy EnforcementPolicy) *VersionEnforcer {
	return &VersionEnforcer{
		manager: manager,
		policy:  policy,
	}
}

// CheckCompliance checks if current version complies with project requirements
func (e *VersionEnforcer) CheckCompliance(projectPath string) (*EnforcementResult, error) {
	// Get project version requirements
	projectConfig, err := e.manager.GetProjectConfig(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get project config: %w", err)
	}

	// Get active version
	activeVersion, err := e.manager.GetActiveVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get active version: %w", err)
	}

	// Check if active version satisfies requirements
	if projectConfig.Required.Satisfies(activeVersion.Version) {
		return &EnforcementResult{
			Compliant: true,
			Action:    constants.EnforcementActionNone,
			Message:   "Version compliance satisfied",
			ExitCode:  constants.ExitSuccess,
		}, nil
	}

	// Detect drift
	drift := e.detectDrift(projectPath, projectConfig.Required.Version, activeVersion.Version, projectConfig.LastUsed)

	result := &EnforcementResult{
		Compliant: false,
		Drift:     drift,
		ExitCode:  constants.ExitError,
	}

	// Determine action based on policy
	if e.policy.StrictMode {
		result.Action = constants.EnforcementActionSwitch
		result.Message = fmt.Sprintf("Strict mode: must switch to version %s", projectConfig.Required.Version)
		result.ExitCode = constants.ExitError
	} else if !e.policy.AllowDrift {
		result.Action = constants.EnforcementActionWarn
		result.Message = fmt.Sprintf("Version drift detected: using %s, required %s",
			activeVersion.Version, projectConfig.Required.Version)
	} else if drift.DriftDuration > e.policy.MaxDriftDuration {
		result.Action = constants.EnforcementActionSwitch
		result.Message = fmt.Sprintf("Drift duration exceeded: %v > %v",
			drift.DriftDuration, e.policy.MaxDriftDuration)
		result.ExitCode = constants.ExitError
	} else {
		result.Action = constants.EnforcementActionNone
		result.Message = "Drift within acceptable limits"
		result.ExitCode = constants.ExitSuccess
	}

	return result, nil
}

// EnforceCompliance enforces version compliance based on policy
func (e *VersionEnforcer) EnforceCompliance(projectPath string) error {
	result, err := e.CheckCompliance(projectPath)
	if err != nil {
		return err
	}

	switch result.Action {
	case constants.EnforcementActionSwitch:
		if e.policy.AutoSync {
			return e.autoSwitchVersion(projectPath)
		}
		return fmt.Errorf("version enforcement required: %s", result.Message)
	case constants.EnforcementActionWarn:
		// Just log warning, don't fail
		return nil
	default:
		return nil
	}
}

// DetectAllDrift detects version drift across all projects
func (e *VersionEnforcer) DetectAllDrift() ([]DriftDetection, error) {
	projects, err := e.manager.ListProjectConfigs()
	if err != nil {
		return nil, fmt.Errorf("failed to list project configs: %w", err)
	}

	activeVersion, err := e.manager.GetActiveVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get active version: %w", err)
	}

	var drifts []DriftDetection
	for _, project := range projects {
		if !project.Required.Satisfies(activeVersion.Version) {
			drift := e.detectDrift(project.ProjectPath, project.Required.Version, activeVersion.Version, project.LastUsed)
			drifts = append(drifts, *drift)
		}
	}

	return drifts, nil
}

func (e *VersionEnforcer) detectDrift(projectPath string, required, active Version, lastUsed time.Time) *DriftDetection {
	drift := &DriftDetection{
		ProjectPath:     projectPath,
		RequiredVersion: required,
		ActiveVersion:   active,
		DriftDuration:   time.Since(lastUsed),
	}

	// Determine drift type and severity
	cmp := active.Compare(required)
	if cmp == 0 {
		drift.DriftType = constants.DriftTypeNone
		drift.Severity = constants.SeverityInfo
	} else if active.Major != required.Major {
		drift.DriftType = constants.DriftTypeMajor
		drift.Severity = constants.SeverityCritical
	} else if active.Minor != required.Minor {
		drift.DriftType = constants.DriftTypeMinor
		drift.Severity = constants.SeverityWarning
	} else if active.Patch != required.Patch {
		drift.DriftType = constants.DriftTypePatch
		drift.Severity = constants.SeverityInfo
	} else {
		drift.DriftType = constants.DriftTypePrerelease
		drift.Severity = constants.SeverityInfo
	}

	return drift
}

func (e *VersionEnforcer) autoSwitchVersion(projectPath string) error {
	projectConfig, err := e.manager.GetProjectConfig(projectPath)
	if err != nil {
		return fmt.Errorf("failed to get project config: %w", err)
	}

	// Try to resolve to an installed version first
	installedVersion, err := e.manager.ResolveVersion(projectConfig.Required)
	if err == nil {
		return e.manager.SwitchToVersion(installedVersion.Version)
	}

	// If not installed, install the required version
	availableVersions, err := e.manager.ListAvailableVersions()
	if err != nil {
		return fmt.Errorf("failed to list available versions: %w", err)
	}

	// Find best matching version
	var bestMatch *Version
	for _, version := range availableVersions {
		if projectConfig.Required.Satisfies(version) {
			if bestMatch == nil || version.Compare(*bestMatch) > 0 {
				bestMatch = &version
			}
		}
	}

	if bestMatch == nil {
		return fmt.Errorf("no available version satisfies constraint %s", projectConfig.Required.Original)
	}

	// Install and switch to the best match
	if err := e.manager.InstallVersion(*bestMatch); err != nil {
		return fmt.Errorf("failed to install version %s: %w", bestMatch, err)
	}

	return e.manager.SwitchToVersion(*bestMatch)
}
