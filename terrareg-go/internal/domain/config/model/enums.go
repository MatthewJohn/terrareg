package model

// ModuleVersionReindexMode represents the re-indexing mode for module versions
// Matches Python's ModuleVersionReindexMode enum
type ModuleVersionReindexMode string

const (
	ModuleVersionReindexModeLegacy     ModuleVersionReindexMode = "legacy"
	ModuleVersionReindexModeAutoPublish ModuleVersionReindexMode = "auto-publish"
	ModuleVersionReindexModeProhibit   ModuleVersionReindexMode = "prohibit"
)

// IsValid validates the module version reindex mode
func (m ModuleVersionReindexMode) IsValid() bool {
	switch m {
	case ModuleVersionReindexModeLegacy, ModuleVersionReindexModeAutoPublish, ModuleVersionReindexModeProhibit:
		return true
	default:
		return false
	}
}

// Product represents the Terraform product type to use for module extraction
// Matches Python's Product enum
type Product string

const (
	ProductTerraform Product = "terraform"
	ProductOpenTofu  Product = "opentofu"
)

// IsValid validates the product type
func (p Product) IsValid() bool {
	switch p {
	case ProductTerraform, ProductOpenTofu:
		return true
	default:
		return false
	}
}

// ServerType represents the server type for running the application
// Matches Python's ServerType enum
type ServerType string

const (
	ServerTypeBuiltin  ServerType = "builtin"
	ServerTypeWaitress ServerType = "waitress"
)

// IsValid validates the server type
func (s ServerType) IsValid() bool {
	switch s {
	case ServerTypeBuiltin, ServerTypeWaitress:
		return true
	default:
		return false
	}
}

// DefaultUiInputOutputView represents the default view type in UI for inputs and outputs
// Matches Python's DefaultUiInputOutputView enum
type DefaultUiInputOutputView string

const (
	DefaultUiInputOutputViewTable    DefaultUiInputOutputView = "table"
	DefaultUiInputOutputViewExpanded DefaultUiInputOutputView = "expanded"
)

// IsValid validates the UI view type
func (v DefaultUiInputOutputView) IsValid() bool {
	switch v {
	case DefaultUiInputOutputViewTable, DefaultUiInputOutputViewExpanded:
		return true
	default:
		return false
	}
}

// ModuleHostingMode represents the module hosting mode (from Python enum)
type ModuleHostingMode string

const (
	ModuleHostingModeAllow    ModuleHostingMode = "true"
	ModuleHostingModeDisallow ModuleHostingMode = "false"
	ModuleHostingModeEnforce  ModuleHostingMode = "enforce"
)

// IsValid validates the module hosting mode
func (m ModuleHostingMode) IsValid() bool {
	switch m {
	case ModuleHostingModeAllow, ModuleHostingModeDisallow, ModuleHostingModeEnforce:
		return true
	default:
		return false
	}
}