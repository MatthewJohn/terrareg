package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseVersion tests version parsing
func TestParseVersion(t *testing.T) {
	tests := []struct {
		name        string
		versionStr  string
		expectError bool
		expectMajor int
		expectMinor int
		expectPatch int
		expectPre   string
	}{
		{"simple version", "1.2.3", false, 1, 2, 3, ""},
		{"with v prefix", "v1.2.3", false, 1, 2, 3, ""},
		{"with prerelease", "1.2.3-alpha", false, 1, 2, 3, "alpha"},
		{"with build", "1.2.3+build", true, 0, 0, 0, ""}, // Build metadata not supported
		{"with prerelease and build", "1.2.3-alpha+build", true, 0, 0, 0, ""}, // Build metadata not supported
		{"complex prerelease", "1.2.3-alpha.1.beta.2", true, 0, 0, 0, ""}, // Dots in prerelease not supported by current regex
		{"empty string", "", true, 0, 0, 0, ""},
		{"invalid format", "1.2", true, 0, 0, 0, ""},
		{"invalid format - text", "abc", true, 0, 0, 0, ""},
		{"zero version", "0.0.0", false, 0, 0, 0, ""},
		{"large version", "999.999.999", false, 999, 999, 999, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := ParseVersion(tt.versionStr)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, version)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, version)
				assert.Equal(t, tt.expectMajor, version.Major())
				assert.Equal(t, tt.expectMinor, version.Minor())
				assert.Equal(t, tt.expectPatch, version.Patch())
				assert.Equal(t, tt.expectPre, version.Prerelease())
				assert.Equal(t, tt.versionStr, version.String())
			}
		})
	}
}

// TestNewVersion tests creating a new version
func TestNewVersion(t *testing.T) {
	version := NewVersion(1, 2, 3)

	assert.Equal(t, 1, version.Major())
	assert.Equal(t, 2, version.Minor())
	assert.Equal(t, 3, version.Patch())
	assert.Equal(t, "1.2.3", version.String())
	assert.Empty(t, version.Prerelease())
	assert.False(t, version.IsPrerelease())
}

// TestVersion_Equal tests version equality
func TestVersion_Equal(t *testing.T) {
	v1, _ := ParseVersion("1.2.3")
	v2, _ := ParseVersion("1.2.3")
	v3, _ := ParseVersion("1.2.4")
	v4, _ := ParseVersion("1.2.3-alpha")

	assert.True(t, v1.Equal(v2))
	assert.True(t, v2.Equal(v1))
	assert.False(t, v1.Equal(v3))
	assert.False(t, v3.Equal(v1))
	assert.False(t, v1.Equal(v4))
	assert.False(t, v1.Equal(nil))
}

// TestVersion_GreaterThan tests version comparison
func TestVersion_GreaterThan(t *testing.T) {
	tests := []struct {
		name     string
		v1       string
		v2       string
		expected bool
	}{
		{"major greater", "2.0.0", "1.9.9", true},
		{"minor greater", "1.3.0", "1.2.9", true},
		{"patch greater", "1.2.4", "1.2.3", true},
		{"equal", "1.2.3", "1.2.3", false},
		{"less than", "1.2.3", "1.2.4", false},
		{"stable greater than prerelease", "1.2.3", "1.2.3-alpha", true},
		{"prerelease less than stable", "1.2.3-alpha", "1.2.3", false},
		{"both prerelease - compare lexically", "1.2.3-beta", "1.2.3-alpha", true},
		{"same prerelease", "1.2.3-alpha", "1.2.3-alpha", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v1, err := ParseVersion(tt.v1)
			require.NoError(t, err)
			v2, err := ParseVersion(tt.v2)
			require.NoError(t, err)

			assert.Equal(t, tt.expected, v1.GreaterThan(v2))
		})
	}

	t.Run("compare with nil", func(t *testing.T) {
		v, _ := ParseVersion("1.0.0")
		assert.True(t, v.GreaterThan(nil))
	})
}

// TestVersion_LessThan tests version comparison
func TestVersion_LessThan(t *testing.T) {
	tests := []struct {
		name     string
		v1       string
		v2       string
		expected bool
	}{
		{"major less", "1.0.0", "2.0.0", true},
		{"minor less", "1.2.0", "1.3.0", true},
		{"patch less", "1.2.3", "1.2.4", true},
		{"equal", "1.2.3", "1.2.3", false},
		{"greater than", "1.2.4", "1.2.3", false},
		{"prerelease less than stable", "1.2.3-alpha", "1.2.3", true},
		{"stable greater than prerelease", "1.2.3", "1.2.3-alpha", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v1, err := ParseVersion(tt.v1)
			require.NoError(t, err)
			v2, err := ParseVersion(tt.v2)
			require.NoError(t, err)

			assert.Equal(t, tt.expected, v1.LessThan(v2))
		})
	}

	t.Run("compare with nil", func(t *testing.T) {
		v, _ := ParseVersion("1.0.0")
		assert.False(t, v.LessThan(nil))
	})
}

// TestVersion_Compare tests version comparison
func TestVersion_Compare(t *testing.T) {
	v1, _ := ParseVersion("1.2.3")
	v2, _ := ParseVersion("1.2.3")
	v3, _ := ParseVersion("1.2.4")
	v4, _ := ParseVersion("1.2.2")

	assert.Equal(t, 0, v1.Compare(v2))
	assert.Equal(t, -1, v1.Compare(v3))
	assert.Equal(t, 1, v1.Compare(v4))
}

// TestVersion_IsPrerelease tests prerelease detection
func TestVersion_IsPrerelease(t *testing.T) {
	tests := []struct {
		version      string
		isPrerelease bool
	}{
		{"1.2.3", false},
		{"1.2.3-alpha", true},
		{"1.2.3-beta", true},
		// Note: rc.1 and alpha.1 not supported by current regex (no dots in prerelease)
		{"0.0.1", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			v, err := ParseVersion(tt.version)
			require.NoError(t, err)
			assert.Equal(t, tt.isPrerelease, v.IsPrerelease())
		})
	}
}

// TestVersionString tests version string representation
func TestVersionString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1.2.3", "1.2.3"},
		{"v1.2.3", "v1.2.3"},
		{"1.2.3-alpha", "1.2.3-alpha"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			v, err := ParseVersion(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.input, v.String())
		})
	}
}

// TestParseVersionConstraint tests constraint parsing
func TestParseVersionConstraint(t *testing.T) {
	tests := []struct {
		name           string
		constraint     string
		expectError    bool
		expectOperator string
		expectVersion  string
	}{
		{"exact version", "1.2.3", false, "=", "1.2.3"},
		{"greater than", ">1.2.3", false, ">", "1.2.3"},
		{"greater than or equal", ">=1.2.3", false, ">=", "1.2.3"},
		{"less than", "<1.2.3", false, "<", "1.2.3"},
		{"less than or equal", "<=1.2.3", false, "<=", "1.2.3"},
		{"pessimistic", "~>1.2.3", false, "~>", "1.2.3"},
		{"with spaces", ">= 1.2.3", false, ">=", "1.2.3"},
		{"exact with equals", "=1.2.3", false, "=", "1.2.3"},
		{"empty string", "", true, "", ""},
		{"invalid version", ">abc", true, "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint, err := ParseVersionConstraint(tt.constraint)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, constraint)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, constraint)
				assert.Equal(t, tt.expectOperator, constraint.operator)
				assert.Equal(t, tt.expectVersion, constraint.version.String())
				assert.Equal(t, tt.constraint, constraint.String())
			}
		})
	}
}

// TestVersionConstraint_Matches tests constraint matching
func TestVersionConstraint_Matches(t *testing.T) {
	tests := []struct {
		name       string
		constraint string
		version    string
		expected   bool
	}{
		{"exact match", "=1.2.3", "1.2.3", true},
		{"exact no match", "=1.2.3", "1.2.4", false},
		{"greater true", ">1.2.3", "1.2.4", true},
		{"greater false", ">1.2.3", "1.2.3", false},
		{"greater or equal true", ">=1.2.3", "1.2.3", true},
		{"greater or equal true 2", ">=1.2.3", "1.2.4", true},
		{"greater or equal false", ">=1.2.3", "1.2.2", false},
		{"less true", "<1.2.3", "1.2.2", true},
		{"less false", "<1.2.3", "1.2.3", false},
		{"less or equal true", "<=1.2.3", "1.2.3", true},
		{"less or equal true 2", "<=1.2.3", "1.2.2", true},
		{"less or equal false", "<=1.2.3", "1.2.4", false},
		{"pessimistic true", "~>1.2.3", "1.2.5", true},
		{"pessimistic upper bound", "~>1.2.3", "1.3.0", false},
		{"pessimistic lower bound", "~>1.2.3", "1.2.3", true},
		{"pessimistic below lower", "~>1.2.3", "1.2.2", false},
		{"implicit equals", "1.2.3", "1.2.3", true},
		{"implicit equals no match", "1.2.3", "1.2.4", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint, err := ParseVersionConstraint(tt.constraint)
			require.NoError(t, err)

			version, err := ParseVersion(tt.version)
			require.NoError(t, err)

			assert.Equal(t, tt.expected, constraint.Matches(version))
		})
	}

	t.Run("nil version", func(t *testing.T) {
		constraint, _ := ParseVersionConstraint(">=1.0.0")
		assert.False(t, constraint.Matches(nil))
	})
}

// TestVersionSort tests sorting versions
func TestVersionSort(t *testing.T) {
	versions := []string{
		"1.2.3",
		"1.10.0",
		"1.2.1",
		"2.0.0",
		"1.2.3-alpha",
		"1.2.4",
		"0.9.9",
	}

	parsedVersions := make([]*Version, len(versions))
	for i, v := range versions {
		parsed, err := ParseVersion(v)
		require.NoError(t, err)
		parsedVersions[i] = parsed
	}

	// Simple bubble sort to verify comparison logic works
	for i := 0; i < len(parsedVersions)-1; i++ {
		for j := 0; j < len(parsedVersions)-1-i; j++ {
			if parsedVersions[j].GreaterThan(parsedVersions[j+1]) {
				parsedVersions[j], parsedVersions[j+1] = parsedVersions[j+1], parsedVersions[j]
			}
		}
	}

	expected := []string{"0.9.9", "1.2.1", "1.2.3-alpha", "1.2.3", "1.2.4", "1.10.0", "2.0.0"}
	for i, v := range parsedVersions {
		assert.Equal(t, expected[i], v.String())
	}
}
