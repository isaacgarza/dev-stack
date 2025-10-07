package version

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	// Semantic version regex pattern
	semverRegex = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z\-\.]+))?(?:\+([0-9A-Za-z\-\.]+))?$`)

	// Version constraint regex pattern
	constraintRegex = regexp.MustCompile(`^([~^<>=!]*)(.+)$`)
)

// ParseVersion parses a version string into a Version struct
func ParseVersion(versionStr string) (*Version, error) {
	if versionStr == "" {
		return nil, NewVersionError(ErrVersionInvalid, "version string cannot be empty", nil)
	}

	// Clean the version string
	versionStr = strings.TrimSpace(versionStr)

	// Handle special cases
	if versionStr == "latest" || versionStr == "*" {
		return &Version{
			Major:    999,
			Minor:    999,
			Patch:    999,
			Original: versionStr,
		}, nil
	}

	matches := semverRegex.FindStringSubmatch(versionStr)
	if matches == nil {
		return nil, NewVersionError(ErrVersionInvalid,
			fmt.Sprintf("invalid semantic version: %s", versionStr), nil)
	}

	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, NewVersionError(ErrVersionInvalid,
			fmt.Sprintf("invalid major version: %s", matches[1]), err)
	}

	minor, err := strconv.Atoi(matches[2])
	if err != nil {
		return nil, NewVersionError(ErrVersionInvalid,
			fmt.Sprintf("invalid minor version: %s", matches[2]), err)
	}

	patch, err := strconv.Atoi(matches[3])
	if err != nil {
		return nil, NewVersionError(ErrVersionInvalid,
			fmt.Sprintf("invalid patch version: %s", matches[3]), err)
	}

	version := &Version{
		Major:    major,
		Minor:    minor,
		Patch:    patch,
		Original: versionStr,
	}

	// Pre-release version
	if len(matches) > 4 && matches[4] != "" {
		version.PreRelease = matches[4]
	}

	// Build metadata
	if len(matches) > 5 && matches[5] != "" {
		version.Build = matches[5]
	}

	return version, nil
}

// ParseVersionConstraint parses a version constraint string
func ParseVersionConstraint(constraintStr string) (*VersionConstraint, error) {
	if constraintStr == "" {
		return nil, NewVersionError(ErrVersionConstraint, "constraint string cannot be empty", nil)
	}

	constraintStr = strings.TrimSpace(constraintStr)

	// Handle wildcards
	if constraintStr == "*" || constraintStr == "latest" {
		return &VersionConstraint{
			Operator: "*",
			Version:  Version{Major: 0, Minor: 0, Patch: 0},
			Original: constraintStr,
		}, nil
	}

	matches := constraintRegex.FindStringSubmatch(constraintStr)
	if matches == nil {
		// No operator, assume exact match
		version, err := ParseVersion(constraintStr)
		if err != nil {
			return nil, NewVersionError(ErrVersionConstraint,
				fmt.Sprintf("invalid version constraint: %s", constraintStr), err)
		}
		return &VersionConstraint{
			Operator: "=",
			Version:  *version,
			Original: constraintStr,
		}, nil
	}

	operator := matches[1]
	versionStr := matches[2]

	// Default operator is =
	if operator == "" {
		operator = "="
	}

	// Handle compound operators
	if strings.Contains(operator, "!") && strings.Contains(operator, "=") {
		operator = "!="
	} else if strings.Contains(operator, ">") && strings.Contains(operator, "=") {
		operator = ">="
	} else if strings.Contains(operator, "<") && strings.Contains(operator, "=") {
		operator = "<="
	} else if strings.Contains(operator, "~") {
		operator = "~"
	} else if strings.Contains(operator, "^") {
		operator = "^"
	} else if strings.Contains(operator, ">") {
		operator = ">"
	} else if strings.Contains(operator, "<") {
		operator = "<"
	} else if strings.Contains(operator, "=") {
		operator = "="
	}

	version, err := ParseVersion(versionStr)
	if err != nil {
		return nil, NewVersionError(ErrVersionConstraint,
			fmt.Sprintf("invalid version in constraint: %s", versionStr), err)
	}

	return &VersionConstraint{
		Operator: operator,
		Version:  *version,
		Original: constraintStr,
	}, nil
}

// ParseVersionRange parses a version range like ">=1.0.0 <2.0.0"
func ParseVersionRange(rangeStr string) ([]*VersionConstraint, error) {
	if rangeStr == "" {
		return nil, NewVersionError(ErrVersionConstraint, "range string cannot be empty", nil)
	}

	// Split by spaces or commas
	parts := regexp.MustCompile(`[\s,]+`).Split(strings.TrimSpace(rangeStr), -1)
	var constraints []*VersionConstraint

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		constraint, err := ParseVersionConstraint(part)
		if err != nil {
			return nil, err
		}

		constraints = append(constraints, constraint)
	}

	if len(constraints) == 0 {
		return nil, NewVersionError(ErrVersionConstraint, "no valid constraints found in range", nil)
	}

	return constraints, nil
}

// ValidateVersion checks if a version string is valid
func ValidateVersion(versionStr string) error {
	_, err := ParseVersion(versionStr)
	return err
}

// ValidateConstraint checks if a constraint string is valid
func ValidateConstraint(constraintStr string) error {
	_, err := ParseVersionConstraint(constraintStr)
	return err
}

// NormalizeVersion normalizes a version string to a standard format
func NormalizeVersion(versionStr string) (string, error) {
	version, err := ParseVersion(versionStr)
	if err != nil {
		return "", err
	}
	return version.String(), nil
}

// CompareVersions compares two version strings
func CompareVersions(v1Str, v2Str string) (int, error) {
	v1, err := ParseVersion(v1Str)
	if err != nil {
		return 0, NewVersionError(ErrVersionInvalid,
			fmt.Sprintf("invalid version v1: %s", v1Str), err)
	}

	v2, err := ParseVersion(v2Str)
	if err != nil {
		return 0, NewVersionError(ErrVersionInvalid,
			fmt.Sprintf("invalid version v2: %s", v2Str), err)
	}

	return v1.Compare(*v2), nil
}

// SortVersions sorts a slice of version strings in ascending order
func SortVersions(versions []string) ([]string, error) {
	type versionPair struct {
		original string
		parsed   *Version
	}

	var pairs []versionPair
	for _, v := range versions {
		parsed, err := ParseVersion(v)
		if err != nil {
			return nil, err
		}
		pairs = append(pairs, versionPair{original: v, parsed: parsed})
	}

	// Sort pairs by parsed version
	for i := 0; i < len(pairs); i++ {
		for j := i + 1; j < len(pairs); j++ {
			if pairs[i].parsed.Compare(*pairs[j].parsed) > 0 {
				pairs[i], pairs[j] = pairs[j], pairs[i]
			}
		}
	}

	result := make([]string, len(pairs))
	for i, pair := range pairs {
		result[i] = pair.original
	}

	return result, nil
}

// GetLatestVersion returns the latest version from a list of version strings
func GetLatestVersion(versions []string) (string, error) {
	if len(versions) == 0 {
		return "", NewVersionError(ErrVersionNotFound, "no versions provided", nil)
	}

	sorted, err := SortVersions(versions)
	if err != nil {
		return "", err
	}

	return sorted[len(sorted)-1], nil
}

// FilterVersionsByConstraint filters versions that satisfy a constraint
func FilterVersionsByConstraint(versions []string, constraintStr string) ([]string, error) {
	constraint, err := ParseVersionConstraint(constraintStr)
	if err != nil {
		return nil, err
	}

	var filtered []string
	for _, vStr := range versions {
		version, err := ParseVersion(vStr)
		if err != nil {
			continue // Skip invalid versions
		}

		if constraint.Satisfies(*version) {
			filtered = append(filtered, vStr)
		}
	}

	return filtered, nil
}
