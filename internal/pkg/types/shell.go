package types

import "github.com/isaacgarza/dev-stack/internal/pkg/constants"

// ShellType represents supported shell types for completion
type ShellType string

const (
	ShellTypeBash       ShellType = constants.ShellBash
	ShellTypeZsh        ShellType = constants.ShellZsh
	ShellTypeFish       ShellType = constants.ShellFish
	ShellTypePowerShell ShellType = constants.ShellPowerShell
)

// String returns the string representation of the shell type
func (s ShellType) String() string {
	return string(s)
}

// IsValid returns true if the shell type is supported
func (s ShellType) IsValid() bool {
	switch s {
	case ShellTypeBash, ShellTypeZsh, ShellTypeFish, ShellTypePowerShell:
		return true
	default:
		return false
	}
}

// AllShellTypes returns all supported shell types
func AllShellTypes() []ShellType {
	return []ShellType{
		ShellTypeBash,
		ShellTypeZsh,
		ShellTypeFish,
		ShellTypePowerShell,
	}
}

// AllShellTypeStrings returns all supported shell types as strings
func AllShellTypeStrings() []string {
	shells := AllShellTypes()
	result := make([]string, len(shells))
	for i, shell := range shells {
		result[i] = shell.String()
	}
	return result
}
