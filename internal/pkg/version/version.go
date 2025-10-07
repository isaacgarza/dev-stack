package version

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"time"
)

var (
	// These variables are set by the build process using ldflags
	AppVersion = "dev"
	GitCommit  = "unknown"
	BuildDate  = "unknown"
	BuildBy    = "unknown"
)

// BuildInfo contains comprehensive build information
type BuildInfo struct {
	Version    string           `json:"version"`
	GitCommit  string           `json:"git_commit"`
	BuildDate  string           `json:"build_date"`
	BuildBy    string           `json:"build_by"`
	GoVersion  string           `json:"go_version"`
	Platform   string           `json:"platform"`
	Arch       string           `json:"arch"`
	BuildTime  time.Time        `json:"build_time"`
	ModuleInfo *debug.BuildInfo `json:"module_info,omitempty"`
}

// GetBuildInfo returns comprehensive build information
func GetBuildInfo() *BuildInfo {
	buildTime, _ := time.Parse(time.RFC3339, BuildDate)
	if buildTime.IsZero() {
		buildTime = time.Now()
	}

	info := &BuildInfo{
		Version:   AppVersion,
		GitCommit: GitCommit,
		BuildDate: BuildDate,
		BuildBy:   BuildBy,
		GoVersion: runtime.Version(),
		Platform:  runtime.GOOS,
		Arch:      runtime.GOARCH,
		BuildTime: buildTime,
	}

	// Try to get module information
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		info.ModuleInfo = buildInfo
	}

	return info
}

// GetAppVersion returns the application version string
func GetAppVersion() string {
	if AppVersion == "dev" {
		// Try to get version from module info
		if buildInfo, ok := debug.ReadBuildInfo(); ok {
			return buildInfo.Main.Version
		}
	}
	return AppVersion
}

// GetShortVersion returns a short version string
func GetShortVersion() string {
	version := GetAppVersion()
	if version == "(devel)" || version == "dev" {
		return "dev"
	}
	return version
}

// GetFullVersion returns a detailed version string
func GetFullVersion() string {
	info := GetBuildInfo()

	version := info.Version
	if version == "dev" || version == "(devel)" {
		version = "dev"
	}

	result := fmt.Sprintf("dev-stack %s", version)

	if info.GitCommit != "unknown" && info.GitCommit != "" {
		if len(info.GitCommit) > 7 {
			result += fmt.Sprintf(" (%s)", info.GitCommit[:7])
		} else {
			result += fmt.Sprintf(" (%s)", info.GitCommit)
		}
	}

	return result
}

// GetFormattedBuildInfo returns formatted build information
func GetFormattedBuildInfo() string {
	info := GetBuildInfo()

	result := fmt.Sprintf("Version:    %s\n", info.Version)
	result += fmt.Sprintf("Git Commit: %s\n", info.GitCommit)
	result += fmt.Sprintf("Build Date: %s\n", info.BuildDate)
	result += fmt.Sprintf("Built By:   %s\n", info.BuildBy)
	result += fmt.Sprintf("Go Version: %s\n", info.GoVersion)
	result += fmt.Sprintf("Platform:   %s/%s\n", info.Platform, info.Arch)

	return result
}

// IsDevBuild returns true if this is a development build
func IsDevBuild() bool {
	return AppVersion == "dev" || AppVersion == "(devel)"
}

// GetUserAgent returns a user agent string for HTTP requests
func GetUserAgent() string {
	return fmt.Sprintf("dev-stack/%s (%s/%s)", GetShortVersion(), runtime.GOOS, runtime.GOARCH)
}

// IsAppVersionCompatible checks if the current version is compatible with required version
func IsAppVersionCompatible(requiredVersion string) bool {
	currentVersion := GetShortVersion()

	// Development builds are always compatible
	if IsDevBuild() {
		return true
	}

	// For now, use simple string comparison
	// In production, you'd implement proper semantic versioning
	return currentVersion >= requiredVersion
}
