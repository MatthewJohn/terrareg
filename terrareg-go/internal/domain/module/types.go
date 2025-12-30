package module

// ImportModuleVersionRequest represents a module import request shared between domain and application layers
type ImportModuleVersionRequest struct {
	Namespace string
	Module    string
	Provider  string
	Version   *string // Optional - derived from git tag if not provided
	GitTag    *string // Optional - conflicts with Version
}