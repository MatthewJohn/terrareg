package model

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// TestProviderTier_Values tests that provider tier enum values are defined correctly
func TestProviderTier_Values(t *testing.T) {
	// Test all valid provider tiers
	validTiers := []sqldb.ProviderTier{
		sqldb.ProviderTierOfficial,
		sqldb.ProviderTierPartner,
		sqldb.ProviderTierCommunity,
	}

	expectedValues := []string{
		"official",
		"partner",
		"community",
	}

	// Verify each tier has the correct string value
	for i, tier := range validTiers {
		assert.Equal(t, expectedValues[i], string(tier), "Tier should have correct string value")
	}
}

// TestProviderSourceType_Values tests that provider source type enum values are defined correctly
func TestProviderSourceType_Values(t *testing.T) {
	// Test all valid provider source types
	validTypes := []sqldb.ProviderSourceType{
		sqldb.ProviderSourceTypeGithub,
		sqldb.ProviderSourceTypeGitlab,
		sqldb.ProviderSourceTypeBitbucket,
	}

	expectedValues := []string{
		"github",
		"gitlab",
		"bitbucket",
	}

	// Verify each type has the correct string value
	for i, sourceType := range validTypes {
		assert.Equal(t, expectedValues[i], string(sourceType), "Source type should have correct string value")
	}
}

// TestNamespaceType_Values tests that namespace type enum values are defined correctly
func TestNamespaceType_Values(t *testing.T) {
	// Test all valid namespace types
	validTypes := []sqldb.NamespaceType{
		sqldb.NamespaceTypeNone,
		sqldb.NamespaceTypeGithubUser,
		sqldb.NamespaceTypeGithubOrg,
	}

	expectedValues := []string{
		"NONE",
		"GITHUB_USER",
		"GITHUB_ORGANISATION",
	}

	// Verify each type has the correct string value
	for i, nsType := range validTypes {
		assert.Equal(t, expectedValues[i], string(nsType), "Namespace type should have correct string value")
	}
}

// TestNamespaceType_DefaultValues tests namespace type default values
func TestNamespaceType_DefaultValues(t *testing.T) {
	// Test that zero value is empty string
	var nsType sqldb.NamespaceType
	assert.Equal(t, "", string(nsType))
	// The NONE constant has the value "NONE"
	assert.Equal(t, "NONE", string(sqldb.NamespaceTypeNone))
}

// TestEnum_StringConversion tests that enum types convert correctly to strings
func TestEnum_StringConversion(t *testing.T) {
	// ProviderTier to string conversion
	assert.Equal(t, "official", string(sqldb.ProviderTierOfficial))
	assert.Equal(t, "partner", string(sqldb.ProviderTierPartner))
	assert.Equal(t, "community", string(sqldb.ProviderTierCommunity))

	// ProviderSourceType to string conversion
	assert.Equal(t, "github", string(sqldb.ProviderSourceTypeGithub))
	assert.Equal(t, "gitlab", string(sqldb.ProviderSourceTypeGitlab))
	assert.Equal(t, "bitbucket", string(sqldb.ProviderSourceTypeBitbucket))

	// NamespaceType to string conversion
	assert.Equal(t, "NONE", string(sqldb.NamespaceTypeNone))
	assert.Equal(t, "GITHUB_USER", string(sqldb.NamespaceTypeGithubUser))
	assert.Equal(t, "GITHUB_ORGANISATION", string(sqldb.NamespaceTypeGithubOrg))
}

// TestProviderBinaryOperatingSystemType_Values tests provider binary OS type enum values
func TestProviderBinaryOperatingSystemType_Values(t *testing.T) {
	validTypes := []sqldb.ProviderBinaryOperatingSystemType{
		sqldb.OSLinux,
		sqldb.OSWindows,
		sqldb.OSDarwin,
		sqldb.OSFreeBSD,
	}

	expectedValues := []string{
		"linux",
		"windows",
		"darwin",
		"freebsd",
	}

	for i, osType := range validTypes {
		assert.Equal(t, expectedValues[i], string(osType), "OS type should have correct string value")
	}
}

// TestProviderBinaryArchitectureType_Values tests provider binary architecture type enum values
func TestProviderBinaryArchitectureType_Values(t *testing.T) {
	validTypes := []sqldb.ProviderBinaryArchitectureType{
		sqldb.ArchAMD64,
		sqldb.ArchARM,
		sqldb.ArchARM64,
		sqldb.Arch386,
	}

	expectedValues := []string{
		"amd64",
		"arm",
		"arm64",
		"386",
	}

	for i, archType := range validTypes {
		assert.Equal(t, expectedValues[i], string(archType), "Architecture type should have correct string value")
	}
}
