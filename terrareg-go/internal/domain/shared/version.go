package shared

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Version represents a semantic version
type Version struct {
	major      int
	minor      int
	patch      int
	prerelease string
	build      string
	original   string
}

var versionRegex = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:-([a-z0-9]+))?$`)

// ParseVersion parses a version string into a Version
func ParseVersion(versionStr string) (*Version, error) {
	if versionStr == "" {
		return nil, fmt.Errorf("version string cannot be empty")
	}

	matches := versionRegex.FindStringSubmatch(versionStr)
	if matches == nil {
		return nil, fmt.Errorf("invalid version format: %s", versionStr)
	}

	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	patch, _ := strconv.Atoi(matches[3])

	return &Version{
		major:      major,
		minor:      minor,
		patch:      patch,
		prerelease: matches[4],
		build:      "", // Build metadata not supported
		original:   versionStr,
	}, nil
}

// NewVersion creates a new version
func NewVersion(major, minor, patch int) *Version {
	return &Version{
		major:    major,
		minor:    minor,
		patch:    patch,
		original: fmt.Sprintf("%d.%d.%d", major, minor, patch),
	}
}

// String returns the string representation
func (v *Version) String() string {
	if v.original != "" {
		return v.original
	}

	s := fmt.Sprintf("%d.%d.%d", v.major, v.minor, v.patch)
	if v.prerelease != "" {
		s += "-" + v.prerelease
	}
	if v.build != "" {
		s += "+" + v.build
	}
	return s
}

// Major returns the major version
func (v *Version) Major() int {
	return v.major
}

// Minor returns the minor version
func (v *Version) Minor() int {
	return v.minor
}

// Patch returns the patch version
func (v *Version) Patch() int {
	return v.patch
}

// Prerelease returns the prerelease version
func (v *Version) Prerelease() string {
	return v.prerelease
}

// IsPrerelease returns true if this is a prerelease version
func (v *Version) IsPrerelease() bool {
	return v.prerelease != ""
}

// Equal checks if two versions are equal
func (v *Version) Equal(other *Version) bool {
	if other == nil {
		return false
	}
	return v.major == other.major &&
		v.minor == other.minor &&
		v.patch == other.patch &&
		v.prerelease == other.prerelease
}

// GreaterThan checks if this version is greater than another
func (v *Version) GreaterThan(other *Version) bool {
	if other == nil {
		return true
	}

	if v.major != other.major {
		return v.major > other.major
	}
	if v.minor != other.minor {
		return v.minor > other.minor
	}
	if v.patch != other.patch {
		return v.patch > other.patch
	}

	// If one has prerelease and the other doesn't, the one without is greater
	if v.prerelease == "" && other.prerelease != "" {
		return true
	}
	if v.prerelease != "" && other.prerelease == "" {
		return false
	}

	// Both have prerelease, compare lexicographically
	return v.prerelease > other.prerelease
}

// LessThan checks if this version is less than another
func (v *Version) LessThan(other *Version) bool {
	if other == nil {
		return false
	}
	return !v.GreaterThan(other) && !v.Equal(other)
}

// Compare compares two versions (-1: less, 0: equal, 1: greater)
func (v *Version) Compare(other *Version) int {
	if v.Equal(other) {
		return 0
	}
	if v.GreaterThan(other) {
		return 1
	}
	return -1
}

// VersionConstraint represents a version constraint (e.g., ">= 1.0.0")
type VersionConstraint struct {
	operator string
	version  *Version
	raw      string
}

// ParseVersionConstraint parses a version constraint
func ParseVersionConstraint(constraintStr string) (*VersionConstraint, error) {
	constraintStr = strings.TrimSpace(constraintStr)
	if constraintStr == "" {
		return nil, fmt.Errorf("constraint string cannot be empty")
	}

	// Extract operator and version
	var operator, versionStr string
	if strings.HasPrefix(constraintStr, ">=") {
		operator = ">="
		versionStr = strings.TrimSpace(constraintStr[2:])
	} else if strings.HasPrefix(constraintStr, "<=") {
		operator = "<="
		versionStr = strings.TrimSpace(constraintStr[2:])
	} else if strings.HasPrefix(constraintStr, ">") {
		operator = ">"
		versionStr = strings.TrimSpace(constraintStr[1:])
	} else if strings.HasPrefix(constraintStr, "<") {
		operator = "<"
		versionStr = strings.TrimSpace(constraintStr[1:])
	} else if strings.HasPrefix(constraintStr, "=") {
		operator = "="
		versionStr = strings.TrimSpace(constraintStr[1:])
	} else if strings.HasPrefix(constraintStr, "~>") {
		operator = "~>"
		versionStr = strings.TrimSpace(constraintStr[2:])
	} else {
		operator = "="
		versionStr = constraintStr
	}

	version, err := ParseVersion(versionStr)
	if err != nil {
		return nil, fmt.Errorf("invalid version in constraint: %w", err)
	}

	return &VersionConstraint{
		operator: operator,
		version:  version,
		raw:      constraintStr,
	}, nil
}

// Matches checks if a version matches this constraint
func (vc *VersionConstraint) Matches(version *Version) bool {
	if version == nil {
		return false
	}

	switch vc.operator {
	case "=":
		return version.Equal(vc.version)
	case ">":
		return version.GreaterThan(vc.version)
	case ">=":
		return version.GreaterThan(vc.version) || version.Equal(vc.version)
	case "<":
		return version.LessThan(vc.version)
	case "<=":
		return version.LessThan(vc.version) || version.Equal(vc.version)
	case "~>":
		// Pessimistic constraint: ~> 1.2.3 means >= 1.2.3 and < 1.3.0
		if version.LessThan(vc.version) {
			return false
		}
		upper := NewVersion(vc.version.major, vc.version.minor+1, 0)
		return version.LessThan(upper)
	default:
		return false
	}
}

// String returns the string representation
func (vc *VersionConstraint) String() string {
	if vc.raw != "" {
		return vc.raw
	}
	return vc.operator + " " + vc.version.String()
}
