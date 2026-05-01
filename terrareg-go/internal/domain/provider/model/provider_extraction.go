package model

// BinaryPlatform represents the operating system and architecture for a provider binary
type BinaryPlatform struct {
	// OS is the operating system (e.g., "linux", "windows", "darwin")
	OS string

	// Architecture is the CPU architecture (e.g., "amd64", "arm64")
	Architecture string
}

// DocumentationType represents the type of provider documentation
type DocumentationType string

const (
	// DocumentationTypeOverview represents provider overview documentation
	DocumentationTypeOverview DocumentationType = "overview"

	// DocumentationTypeResource represents resource documentation
	DocumentationTypeResource DocumentationType = "resource"

	// DocumentationTypeDataSource represents data source documentation
	DocumentationTypeDataSource DocumentationType = "data-source"

	// DocumentationTypeGuide represents guide documentation
	DocumentationTypeGuide DocumentationType = "guide"
)

// MarkdownMetadata represents parsed frontmatter metadata from markdown files
type MarkdownMetadata struct {
	// Title is the page_title from frontmatter
	Title *string

	// Subcategory is the subcategory from frontmatter
	Subcategory *string

	// Description is the description from frontmatter
	Description *string
}

// ManifestFile represents a terraform provider manifest file
// Reference: https://developer.hashicorp.com/terraform/registry/providers/publishing#terraform-registry-manifest-file
type ManifestFile struct {
	// Version is the manifest version (must be 1)
	Version int

	// Metadata contains the manifest metadata
	Metadata ManifestMetadata
}

// ManifestMetadata contains the metadata section of the manifest file
type ManifestMetadata struct {
	// ProtocolVersions is the list of supported Terraform protocol versions
	ProtocolVersions []string
}
