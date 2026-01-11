package provider_source

import (
	"regexp"
)

// tagRegex matches version tags like v1.5.2, v1.5.2-beta1, etc.
var tagRegex = regexp.MustCompile(`^v([0-9]+\.[0-9]+\.[0-9]+(-[a-z0-9]+)?)$`)

// ReleaseArtifactMetadata stores release artifact metadata
// Python reference: provider_source/repository_release_metadata.py::ReleaseArtifactMetadata
type ReleaseArtifactMetadata struct {
	Name       string
	ProviderID string
}

// NewReleaseArtifactMetadata creates a new ReleaseArtifactMetadata
func NewReleaseArtifactMetadata(name, providerID string) *ReleaseArtifactMetadata {
	return &ReleaseArtifactMetadata{
		Name:       name,
		ProviderID: providerID,
	}
}

// Equals checks if two release artifacts are the same
// Python reference: ReleaseArtifactMetadata.__eq__
func (r *ReleaseArtifactMetadata) Equals(other *ReleaseArtifactMetadata) bool {
	if other == nil {
		return false
	}
	return r.Name == other.Name && r.ProviderID == other.ProviderID
}

// RepositoryReleaseMetadata stores repository release metadata
// Python reference: provider_source/repository_release_metadata.py::RepositoryReleaseMetadata
type RepositoryReleaseMetadata struct {
	Name            string
	Tag             string
	ArchiveURL      string
	CommitHash      string
	ProviderID      string
	ReleaseArtifacts []*ReleaseArtifactMetadata
}

// NewRepositoryReleaseMetadata creates a new RepositoryReleaseMetadata
func NewRepositoryReleaseMetadata(
	name, tag, archiveURL, commitHash, providerID string,
	releaseArtifacts []*ReleaseArtifactMetadata,
) *RepositoryReleaseMetadata {
	return &RepositoryReleaseMetadata{
		Name:            name,
		Tag:             tag,
		ArchiveURL:      archiveURL,
		CommitHash:      commitHash,
		ProviderID:      providerID,
		ReleaseArtifacts: releaseArtifacts,
	}
}

// Equals checks if two repository releases are the same
// Python reference: RepositoryReleaseMetadata.__eq__
func (r *RepositoryReleaseMetadata) Equals(other *RepositoryReleaseMetadata) bool {
	if other == nil {
		return false
	}
	if r.Name != other.Name ||
		r.Tag != other.Tag ||
		r.ProviderID != other.ProviderID ||
		r.ArchiveURL != other.ArchiveURL ||
		r.CommitHash != other.CommitHash ||
		len(r.ReleaseArtifacts) != len(other.ReleaseArtifacts) {
		return false
	}

	// Compare release artifacts
	for i, artifact := range r.ReleaseArtifacts {
		if !artifact.Equals(other.ReleaseArtifacts[i]) {
			return false
		}
	}

	return true
}

// TagToVersion converts a tag to a version string
// Returns nil if the tag doesn't match the expected format
// Python reference: RepositoryReleaseMetadata.tag_to_version
// Examples: "v1.5.2" -> "1.5.2", "v1.5.2-beta1" -> "1.5.2-beta1", "invalid" -> nil
func TagToVersion(tag string) *string {
	matches := tagRegex.FindStringSubmatch(tag)
	if len(matches) < 2 {
		return nil
	}
	version := matches[1]
	return &version
}

// VersionToTag converts a version string to a tag
// Python reference: RepositoryReleaseMetadata.version_to_tag
// Examples: "1.5.2" -> "v1.5.2", "1.5.2-beta1" -> "v1.5.2-beta1"
func VersionToTag(version string) string {
	return "v" + version
}

// Version returns the version string from the tag
// Returns nil if the tag doesn't match the expected format
// Python reference: RepositoryReleaseMetadata.version
func (r *RepositoryReleaseMetadata) Version() *string {
	return TagToVersion(r.Tag)
}
