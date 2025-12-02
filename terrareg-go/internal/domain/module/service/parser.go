package service

// ModuleParser handles parsing of module files
type ModuleParser interface {
	// ParseModule parses a module directory and extracts metadata
	ParseModule(modulePath string) (*ParseResult, error)
	// DetectSubmodules finds submodules in the module directory
	DetectSubmodules(modulePath string) ([]string, error)
	// DetectExamples finds example directories in the module
	DetectExamples(modulePath string) ([]string, error)
}

// ParseResult contains the results of parsing a module
type ParseResult struct {
	Description      string
	ReadmeContent    string
	RawTerraformDocs []byte
	Owner            string
	Variables        []Variable
	Outputs          []Output
	ProviderVersions []ProviderVersion
	Resources        []Resource
}

// Variable represents a Terraform variable
type Variable struct {
	Name        string
	Type        string
	Description string
	Default     interface{}
	Required    bool
}

// Output represents a Terraform output
type Output struct {
	Name        string
	Description string
}

// ProviderVersion represents a required provider version
type ProviderVersion struct {
	Name    string
	Version string
}

// Resource represents a Terraform resource
type Resource struct {
	Type string
	Name string
}
