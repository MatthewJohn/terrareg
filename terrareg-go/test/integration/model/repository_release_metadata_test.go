package model

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source"
)

// TestReleaseArtifactMetadata_Init tests initialization of ReleaseArtifactMetadata
// Python reference: test_repository_release_metadata.py::TestReleaseArtifactMetadata::test_init
func TestReleaseArtifactMetadata_Init(t *testing.T) {
	obj := provider_source.NewReleaseArtifactMetadata("unittest-name", "unittest-provider-id")
	assert.Equal(t, "unittest-name", obj.Name)
	assert.Equal(t, "unittest-provider-id", obj.ProviderID)
}

// TestReleaseArtifactMetadata_Equals tests equality comparison
// Python reference: test_repository_release_metadata.py::TestReleaseArtifactMetadata::test_eq_method
func TestReleaseArtifactMetadata_Equals(t *testing.T) {
	first := provider_source.NewReleaseArtifactMetadata("first-name", "first-provider-id")

	t.Run("Equal artifacts", func(t *testing.T) {
		other := provider_source.NewReleaseArtifactMetadata("first-name", "first-provider-id")
		assert.True(t, first.Equals(other))
	})

	t.Run("Different name", func(t *testing.T) {
		other := provider_source.NewReleaseArtifactMetadata("other name", "first-provider-id")
		assert.False(t, first.Equals(other))
	})

	t.Run("Different provider_id", func(t *testing.T) {
		other := provider_source.NewReleaseArtifactMetadata("first-name", "other-provider-id")
		assert.False(t, first.Equals(other))
	})

	t.Run("Comparison with nil", func(t *testing.T) {
		assert.False(t, first.Equals(nil))
	})
}

// TestRepositoryReleaseMetadata_Init tests initialization of RepositoryReleaseMetadata
// Python reference: test_repository_release_metadata.py::TestRepositoryReleaseMetadata::test_init
func TestRepositoryReleaseMetadata_Init(t *testing.T) {
	releaseArtifacts := []*provider_source.ReleaseArtifactMetadata{
		provider_source.NewReleaseArtifactMetadata("unittest-release-art", "unittest-artifact-provider-id"),
	}

	obj := provider_source.NewRepositoryReleaseMetadata(
		"Unit Test Name",
		"v5.7.2",
		"https://example.com/unittest/example.zip",
		"abcdefgunittesthash",
		"unittestproviderreleaseid",
		releaseArtifacts,
	)

	assert.Equal(t, "Unit Test Name", obj.Name)
	assert.Equal(t, "v5.7.2", obj.Tag)
	assert.Equal(t, "https://example.com/unittest/example.zip", obj.ArchiveURL)
	assert.Equal(t, "abcdefgunittesthash", obj.CommitHash)
	assert.Equal(t, "unittestproviderreleaseid", obj.ProviderID)
	assert.Equal(t, releaseArtifacts, obj.ReleaseArtifacts)
}

// TestRepositoryReleaseMetadata_Equals tests equality comparison
// Python reference: test_repository_release_metadata.py::TestRepositoryReleaseMetadata::test_eq_method
func TestRepositoryReleaseMetadata_Equals(t *testing.T) {
	releaseArtifacts := []*provider_source.ReleaseArtifactMetadata{
		provider_source.NewReleaseArtifactMetadata("unittest-release-art", "unittest-artifact-provider-id"),
	}

	params := &provider_source.RepositoryReleaseMetadata{
		Name:            "Unit Test Name",
		Tag:             "v5.7.2",
		ArchiveURL:      "https://example.com/unittest/example.zip",
		CommitHash:      "abcdefgunittesthash",
		ProviderID:      "unittestproviderreleaseid",
		ReleaseArtifacts: releaseArtifacts,
	}

	otherReleaseArtifacts := []*provider_source.ReleaseArtifactMetadata{
		provider_source.NewReleaseArtifactMetadata("unittest-release-art", "unittest-artifact-provider-id"),
	}

	otherParams := &provider_source.RepositoryReleaseMetadata{
		Name:            "Unit Test Name",
		Tag:             "v5.7.2",
		ArchiveURL:      "https://example.com/unittest/example.zip",
		CommitHash:      "abcdefgunittesthash",
		ProviderID:      "unittestproviderreleaseid",
		ReleaseArtifacts: otherReleaseArtifacts,
	}

	t.Run("Equal objects", func(t *testing.T) {
		// Create a copy using NewRepositoryReleaseMetadata
		first := provider_source.NewRepositoryReleaseMetadata(
			params.Name,
			params.Tag,
			params.ArchiveURL,
			params.CommitHash,
			params.ProviderID,
			params.ReleaseArtifacts,
		)
		other := provider_source.NewRepositoryReleaseMetadata(
			otherParams.Name,
			otherParams.Tag,
			otherParams.ArchiveURL,
			otherParams.CommitHash,
			otherParams.ProviderID,
			otherParams.ReleaseArtifacts,
		)
		assert.True(t, first.Equals(other))
	})

	t.Run("Different name", func(t *testing.T) {
		first := provider_source.NewRepositoryReleaseMetadata(
			params.Name, params.Tag, params.ArchiveURL,
			params.CommitHash, params.ProviderID, params.ReleaseArtifacts,
		)
		other := provider_source.NewRepositoryReleaseMetadata(
			"other name", params.Tag, params.ArchiveURL,
			params.CommitHash, params.ProviderID, otherParams.ReleaseArtifacts,
		)
		assert.False(t, first.Equals(other))
	})

	t.Run("Different tag", func(t *testing.T) {
		first := provider_source.NewRepositoryReleaseMetadata(
			params.Name, params.Tag, params.ArchiveURL,
			params.CommitHash, params.ProviderID, params.ReleaseArtifacts,
		)
		other := provider_source.NewRepositoryReleaseMetadata(
			params.Name, "other-tag", params.ArchiveURL,
			params.CommitHash, params.ProviderID, otherParams.ReleaseArtifacts,
		)
		assert.False(t, first.Equals(other))
	})

	t.Run("Different provider_id", func(t *testing.T) {
		first := provider_source.NewRepositoryReleaseMetadata(
			params.Name, params.Tag, params.ArchiveURL,
			params.CommitHash, params.ProviderID, params.ReleaseArtifacts,
		)
		other := provider_source.NewRepositoryReleaseMetadata(
			params.Name, params.Tag, params.ArchiveURL,
			params.CommitHash, "other-provider-id", otherParams.ReleaseArtifacts,
		)
		assert.False(t, first.Equals(other))
	})

	t.Run("Different archive_url", func(t *testing.T) {
		first := provider_source.NewRepositoryReleaseMetadata(
			params.Name, params.Tag, params.ArchiveURL,
			params.CommitHash, params.ProviderID, params.ReleaseArtifacts,
		)
		other := provider_source.NewRepositoryReleaseMetadata(
			params.Name, params.Tag, "https://anotherurl.com",
			params.CommitHash, params.ProviderID, otherParams.ReleaseArtifacts,
		)
		assert.False(t, first.Equals(other))
	})

	t.Run("Different commit_hash", func(t *testing.T) {
		first := provider_source.NewRepositoryReleaseMetadata(
			params.Name, params.Tag, params.ArchiveURL,
			params.CommitHash, params.ProviderID, params.ReleaseArtifacts,
		)
		other := provider_source.NewRepositoryReleaseMetadata(
			params.Name, params.Tag, params.ArchiveURL,
			"zxcvbnn", params.ProviderID, otherParams.ReleaseArtifacts,
		)
		assert.False(t, first.Equals(other))
	})

	t.Run("Different release_artifacts", func(t *testing.T) {
		first := provider_source.NewRepositoryReleaseMetadata(
			params.Name, params.Tag, params.ArchiveURL,
			params.CommitHash, params.ProviderID, params.ReleaseArtifacts,
		)
		differentArtifacts := []*provider_source.ReleaseArtifactMetadata{
			provider_source.NewReleaseArtifactMetadata("other-artifact", "other-provider-id"),
		}
		other := provider_source.NewRepositoryReleaseMetadata(
			params.Name, params.Tag, params.ArchiveURL,
			params.CommitHash, params.ProviderID, differentArtifacts,
		)
		assert.False(t, first.Equals(other))
	})

	t.Run("Comparison with nil", func(t *testing.T) {
		first := provider_source.NewRepositoryReleaseMetadata(
			params.Name, params.Tag, params.ArchiveURL,
			params.CommitHash, params.ProviderID, params.ReleaseArtifacts,
		)
		assert.False(t, first.Equals(nil))
	})
}

// TestTagToVersion tests converting tags to versions
// Python reference: test_repository_release_metadata.py::TestRepositoryReleaseMetadata::test_tag_to_version
func TestTagToVersion(t *testing.T) {
	testCases := []struct {
		tag          string
		expectedVersion *string
	}{
		// Valid tags
		{"v1.5.2", strPtr("1.5.2")},
		{"v0.0.0", strPtr("0.0.0")},
		{"v999.998.997", strPtr("999.998.997")},
		// Invalid tags
		{"", nil},
		{"1.2.", nil},
		{"1.2", nil},
		{"1", nil},
		{"somethingelse", nil},
	}

	for _, tc := range testCases {
		t.Run(tc.tag, func(t *testing.T) {
			result := provider_source.TagToVersion(tc.tag)
			if tc.expectedVersion == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, *tc.expectedVersion, *result)
			}
		})
	}
}

// TestVersion tests the version property
// Python reference: test_repository_release_metadata.py::TestRepositoryReleaseMetadata::test_version
func TestVersion(t *testing.T) {
	testCases := []struct {
		tag             string
		expectedVersion *string
	}{
		// Valid tags
		{"v1.5.2", strPtr("1.5.2")},
		{"v0.0.0", strPtr("0.0.0")},
		{"v999.998.997", strPtr("999.998.997")},
		// Invalid tags
		{"", nil},
		{"1.2.", nil},
		{"1.2", nil},
		{"1", nil},
		{"somethingelse", nil},
	}

	for _, tc := range testCases {
		t.Run(tc.tag, func(t *testing.T) {
			metadata := provider_source.NewRepositoryReleaseMetadata(
				"test",
				tc.tag,
				"",
				"",
				"",
				nil,
			)
			result := metadata.Version()
			if tc.expectedVersion == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, *tc.expectedVersion, *result)
			}
		})
	}
}

// TestVersionToTag tests converting versions to tags
func TestVersionToTag(t *testing.T) {
	testCases := []struct {
		version     string
		expectedTag string
	}{
		{"1.5.2", "v1.5.2"},
		{"0.0.0", "v0.0.0"},
		{"999.998.997", "v999.998.997"},
		{"1.5.2-beta1", "v1.5.2-beta1"},
	}

	for _, tc := range testCases {
		t.Run(tc.version, func(t *testing.T) {
			result := provider_source.VersionToTag(tc.version)
			assert.Equal(t, tc.expectedTag, result)
		})
	}
}

// Helper function
func strPtr(s string) *string {
	return &s
}
