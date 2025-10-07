package version

import (
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected Version
		wantErr  bool
	}{
		{
			input: "1.2.3",
			expected: Version{
				Major:    1,
				Minor:    2,
				Patch:    3,
				Original: "1.2.3",
			},
			wantErr: false,
		},
		{
			input: "v1.2.3",
			expected: Version{
				Major:    1,
				Minor:    2,
				Patch:    3,
				Original: "v1.2.3",
			},
			wantErr: false,
		},
		{
			input: "1.2.3-alpha.1",
			expected: Version{
				Major:      1,
				Minor:      2,
				Patch:      3,
				PreRelease: "alpha.1",
				Original:   "1.2.3-alpha.1",
			},
			wantErr: false,
		},
		{
			input: "1.2.3+build.123",
			expected: Version{
				Major:    1,
				Minor:    2,
				Patch:    3,
				Build:    "build.123",
				Original: "1.2.3+build.123",
			},
			wantErr: false,
		},
		{
			input: "1.2.3-beta.2+build.456",
			expected: Version{
				Major:      1,
				Minor:      2,
				Patch:      3,
				PreRelease: "beta.2",
				Build:      "build.456",
				Original:   "1.2.3-beta.2+build.456",
			},
			wantErr: false,
		},
		{
			input: "latest",
			expected: Version{
				Major:    999,
				Minor:    999,
				Patch:    999,
				Original: "latest",
			},
			wantErr: false,
		},
		{
			input:   "",
			wantErr: true,
		},
		{
			input:   "invalid",
			wantErr: true,
		},
		{
			input:   "1.2",
			wantErr: true,
		},
		{
			input:   "1.2.3.4",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseVersion(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseVersion(%q) expected error, got nil", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseVersion(%q) unexpected error: %v", tt.input, err)
				return
			}

			if result.Major != tt.expected.Major ||
				result.Minor != tt.expected.Minor ||
				result.Patch != tt.expected.Patch ||
				result.PreRelease != tt.expected.PreRelease ||
				result.Build != tt.expected.Build ||
				result.Original != tt.expected.Original {
				t.Errorf("ParseVersion(%q) = %+v, want %+v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestVersionCompare(t *testing.T) {
	tests := []struct {
		v1       string
		v2       string
		expected int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.0.0", "1.0.1", -1},
		{"1.0.1", "1.0.0", 1},
		{"1.0.0", "1.1.0", -1},
		{"1.1.0", "1.0.0", 1},
		{"1.0.0", "2.0.0", -1},
		{"2.0.0", "1.0.0", 1},
		{"1.0.0", "1.0.0-alpha", 1},
		{"1.0.0-alpha", "1.0.0", -1},
		{"1.0.0-alpha", "1.0.0-beta", -1},
		{"1.0.0-beta", "1.0.0-alpha", 1},
		{"1.0.0-alpha.1", "1.0.0-alpha.2", -1},
		{"1.0.0-alpha.2", "1.0.0-alpha.1", 1},
	}

	for _, tt := range tests {
		t.Run(tt.v1+"_vs_"+tt.v2, func(t *testing.T) {
			v1, err := ParseVersion(tt.v1)
			if err != nil {
				t.Fatalf("Failed to parse v1 %q: %v", tt.v1, err)
			}

			v2, err := ParseVersion(tt.v2)
			if err != nil {
				t.Fatalf("Failed to parse v2 %q: %v", tt.v2, err)
			}

			result := v1.Compare(*v2)
			if result != tt.expected {
				t.Errorf("Compare(%q, %q) = %d, want %d", tt.v1, tt.v2, result, tt.expected)
			}
		})
	}
}

func TestParseVersionConstraint(t *testing.T) {
	tests := []struct {
		input    string
		expected VersionConstraint
		wantErr  bool
	}{
		{
			input: "1.2.3",
			expected: VersionConstraint{
				Operator: "=",
				Version:  Version{Major: 1, Minor: 2, Patch: 3, Original: "1.2.3"},
				Original: "1.2.3",
			},
			wantErr: false,
		},
		{
			input: ">=1.2.3",
			expected: VersionConstraint{
				Operator: ">=",
				Version:  Version{Major: 1, Minor: 2, Patch: 3, Original: "1.2.3"},
				Original: ">=1.2.3",
			},
			wantErr: false,
		},
		{
			input: ">1.2.3",
			expected: VersionConstraint{
				Operator: ">",
				Version:  Version{Major: 1, Minor: 2, Patch: 3, Original: "1.2.3"},
				Original: ">1.2.3",
			},
			wantErr: false,
		},
		{
			input: "<=1.2.3",
			expected: VersionConstraint{
				Operator: "<=",
				Version:  Version{Major: 1, Minor: 2, Patch: 3, Original: "1.2.3"},
				Original: "<=1.2.3",
			},
			wantErr: false,
		},
		{
			input: "<1.2.3",
			expected: VersionConstraint{
				Operator: "<",
				Version:  Version{Major: 1, Minor: 2, Patch: 3, Original: "1.2.3"},
				Original: "<1.2.3",
			},
			wantErr: false,
		},
		{
			input: "~1.2.3",
			expected: VersionConstraint{
				Operator: "~",
				Version:  Version{Major: 1, Minor: 2, Patch: 3, Original: "1.2.3"},
				Original: "~1.2.3",
			},
			wantErr: false,
		},
		{
			input: "^1.2.3",
			expected: VersionConstraint{
				Operator: "^",
				Version:  Version{Major: 1, Minor: 2, Patch: 3, Original: "1.2.3"},
				Original: "^1.2.3",
			},
			wantErr: false,
		},
		{
			input: "*",
			expected: VersionConstraint{
				Operator: "*",
				Version:  Version{Major: 0, Minor: 0, Patch: 0},
				Original: "*",
			},
			wantErr: false,
		},
		{
			input:   "",
			wantErr: true,
		},
		{
			input:   "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseVersionConstraint(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseVersionConstraint(%q) expected error, got nil", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseVersionConstraint(%q) unexpected error: %v", tt.input, err)
				return
			}

			if result.Operator != tt.expected.Operator ||
				result.Version.Compare(tt.expected.Version) != 0 ||
				result.Original != tt.expected.Original {
				t.Errorf("ParseVersionConstraint(%q) = %+v, want %+v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestVersionConstraintSatisfies(t *testing.T) {
	tests := []struct {
		constraint string
		version    string
		expected   bool
	}{
		{"1.2.3", "1.2.3", true},
		{"1.2.3", "1.2.4", false},
		{">=1.2.3", "1.2.3", true},
		{">=1.2.3", "1.2.4", true},
		{">=1.2.3", "1.2.2", false},
		{">1.2.3", "1.2.4", true},
		{">1.2.3", "1.2.3", false},
		{"<=1.2.3", "1.2.3", true},
		{"<=1.2.3", "1.2.2", true},
		{"<=1.2.3", "1.2.4", false},
		{"<1.2.3", "1.2.2", true},
		{"<1.2.3", "1.2.3", false},
		{"~1.2.3", "1.2.3", true},
		{"~1.2.3", "1.2.4", true},
		{"~1.2.3", "1.3.0", false},
		{"^1.2.3", "1.2.3", true},
		{"^1.2.3", "1.3.0", true},
		{"^1.2.3", "2.0.0", false},
		{"*", "1.2.3", true},
		{"*", "0.0.1", true},
	}

	for _, tt := range tests {
		t.Run(tt.constraint+"_"+tt.version, func(t *testing.T) {
			constraint, err := ParseVersionConstraint(tt.constraint)
			if err != nil {
				t.Fatalf("Failed to parse constraint %q: %v", tt.constraint, err)
			}

			version, err := ParseVersion(tt.version)
			if err != nil {
				t.Fatalf("Failed to parse version %q: %v", tt.version, err)
			}

			result := constraint.Satisfies(*version)
			if result != tt.expected {
				t.Errorf("Constraint %q.Satisfies(%q) = %v, want %v", tt.constraint, tt.version, result, tt.expected)
			}
		})
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		v1       string
		v2       string
		expected int
		wantErr  bool
	}{
		{"1.0.0", "1.0.0", 0, false},
		{"1.0.0", "1.0.1", -1, false},
		{"1.0.1", "1.0.0", 1, false},
		{"invalid", "1.0.0", 0, true},
		{"1.0.0", "invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.v1+"_vs_"+tt.v2, func(t *testing.T) {
			result, err := CompareVersions(tt.v1, tt.v2)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CompareVersions(%q, %q) expected error, got nil", tt.v1, tt.v2)
				}
				return
			}

			if err != nil {
				t.Errorf("CompareVersions(%q, %q) unexpected error: %v", tt.v1, tt.v2, err)
				return
			}

			if result != tt.expected {
				t.Errorf("CompareVersions(%q, %q) = %d, want %d", tt.v1, tt.v2, result, tt.expected)
			}
		})
	}
}

func TestSortVersions(t *testing.T) {
	input := []string{"2.0.0", "1.0.0", "1.1.0", "1.0.1"}
	expected := []string{"1.0.0", "1.0.1", "1.1.0", "2.0.0"}

	result, err := SortVersions(input)
	if err != nil {
		t.Fatalf("SortVersions() unexpected error: %v", err)
	}

	if len(result) != len(expected) {
		t.Fatalf("SortVersions() returned %d versions, want %d", len(result), len(expected))
	}

	for i, v := range result {
		if v != expected[i] {
			t.Errorf("SortVersions()[%d] = %q, want %q", i, v, expected[i])
		}
	}
}

func TestGetLatestVersion(t *testing.T) {
	tests := []struct {
		input    []string
		expected string
		wantErr  bool
	}{
		{
			input:    []string{"1.0.0", "1.1.0", "2.0.0"},
			expected: "2.0.0",
			wantErr:  false,
		},
		{
			input:    []string{"2.0.0", "1.0.0", "1.1.0"},
			expected: "2.0.0",
			wantErr:  false,
		},
		{
			input:   []string{},
			wantErr: true,
		},
		{
			input:   []string{"invalid"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result, err := GetLatestVersion(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetLatestVersion() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("GetLatestVersion() unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("GetLatestVersion() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestFilterVersionsByConstraint(t *testing.T) {
	input := []string{"1.0.0", "1.1.0", "1.2.0", "2.0.0", "2.1.0"}

	tests := []struct {
		constraint string
		expected   []string
	}{
		{
			constraint: ">=1.1.0",
			expected:   []string{"1.1.0", "1.2.0", "2.0.0", "2.1.0"},
		},
		{
			constraint: "~1.1.0",
			expected:   []string{"1.1.0"},
		},
		{
			constraint: "^1.1.0",
			expected:   []string{"1.1.0", "1.2.0"},
		},
		{
			constraint: "*",
			expected:   input,
		},
	}

	for _, tt := range tests {
		t.Run(tt.constraint, func(t *testing.T) {
			result, err := FilterVersionsByConstraint(input, tt.constraint)
			if err != nil {
				t.Fatalf("FilterVersionsByConstraint() unexpected error: %v", err)
			}

			if len(result) != len(tt.expected) {
				t.Fatalf("FilterVersionsByConstraint() returned %d versions, want %d", len(result), len(tt.expected))
			}

			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("FilterVersionsByConstraint()[%d] = %q, want %q", i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestVersionString(t *testing.T) {
	tests := []struct {
		version  Version
		expected string
	}{
		{
			version:  Version{Major: 1, Minor: 2, Patch: 3},
			expected: "1.2.3",
		},
		{
			version:  Version{Major: 1, Minor: 2, Patch: 3, PreRelease: "alpha"},
			expected: "1.2.3-alpha",
		},
		{
			version:  Version{Major: 1, Minor: 2, Patch: 3, Build: "build.123"},
			expected: "1.2.3+build.123",
		},
		{
			version:  Version{Major: 1, Minor: 2, Patch: 3, PreRelease: "beta", Build: "build.456"},
			expected: "1.2.3-beta+build.456",
		},
		{
			version:  Version{Major: 1, Minor: 2, Patch: 3, Original: "v1.2.3"},
			expected: "v1.2.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.version.String()
			if result != tt.expected {
				t.Errorf("Version.String() = %q, want %q", result, tt.expected)
			}
		})
	}
}
