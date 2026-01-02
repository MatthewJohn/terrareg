package model

import (
	"regexp"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/repository/model"
)

// tagRegex matches semantic version tags (v1.2.3 or v1.2.3-beta)
// Python reference: repository_release_metadata.py::_TAG_REGEX
var tagRegex = regexp.MustCompile(`^v([0-9]+\.[0-9]+\.[0-9]+(:?-[a-z0-9]+)?)$`)

// RepositoryReleaseMetadata holds metadata for a repository release
// Python reference: repository_release_metadata.py::RepositoryReleaseMetadata
type RepositoryReleaseMetadata struct {
	Name            string
	Tag             string
	ArchiveURL      string
	CommitHash      string
	ProviderID      int
	Repository      model.Repository
	ReleaseArtifacts []*ReleaseArtifactMetadata
}

// ReleaseArtifactMetadata holds metadata for a release artifact
// Python reference: repository_release_metadata.py::ReleaseArtifactMetadata
type ReleaseArtifactMetadata struct {
	Name      string
	ProviderID int
}

// Version converts the tag to a semantic version
// Returns nil if tag doesn't match expected format
// Python reference: repository_release_metadata.py::RepositoryReleaseMetadata.version
func (m *RepositoryReleaseMetadata) Version() *string {
	return TagToVersion(m.Tag)
}

// TagToVersion converts a git tag to a semantic version
// Returns nil if tag doesn't match expected format
// Python reference: repository_release_metadata.py::RepositoryReleaseMetadata.tag_to_version()
func TagToVersion(tag string) *string {
	if match := tagRegex.FindStringSubmatch(tag); len(match) > 1 {
		version := match[1]
		return &version
	}
	return nil
}

// VersionToTag converts a semantic version to a git tag
// Python reference: repository_release_metadata.py::RepositoryReleaseMetadata.version_to_tag()
func VersionToTag(version string) string {
	return "v" + version
}

// NewRepositoryReleaseMetadata creates a new repository release metadata
func NewRepositoryReleaseMetadata(
	name string,
	tag string,
	archiveURL string,
	commitHash string,
	providerID int,
	repository model.Repository,
	releaseArtifacts []*ReleaseArtifactMetadata,
) *RepositoryReleaseMetadata {
	return &RepositoryReleaseMetadata{
		Name:             name,
		Tag:              tag,
		ArchiveURL:       archiveURL,
		CommitHash:       commitHash,
		ProviderID:       providerID,
		Repository:       repository,
		ReleaseArtifacts: releaseArtifacts,
	}
}

// NewReleaseArtifactMetadata creates a new release artifact metadata
func NewReleaseArtifactMetadata(name string, providerID int) *ReleaseArtifactMetadata {
	return &ReleaseArtifactMetadata{
		Name:      name,
		ProviderID: providerID,
	}
}
