package terrareg

// TerraregModuleSpecs contains Terraform module specification data.
// This matches Python's get_api_module_specs() with additional terrareg-specific 'modules' field.
type TerraregModuleSpecs struct {
	Path                 string                `json:"path"`
	Readme               string                `json:"readme"` // HTML sanitized
	Empty                bool                  `json:"empty"`
	Inputs               []TerraregInput       `json:"inputs"`
	Outputs              []TerraregOutput      `json:"outputs"`
	Dependencies         []TerraregDependency  `json:"dependencies"`
	ProviderDependencies []TerraregProviderDep `json:"provider_dependencies"`
	Resources            []TerraregResource    `json:"resources"`
	Modules              []TerraregModule      `json:"modules"` // Additional terrareg field (not in standard API)
}

// TerraregInput represents a Terraform input variable with additional terrareg fields.
type TerraregInput struct {
	Name           string      `json:"name"`
	Type           string      `json:"type"`
	Description    *string     `json:"description"`
	Required       bool        `json:"required"`
	Default        interface{} `json:"default"`
	AdditionalHelp *string     `json:"additional_help"` // Terrareg-specific field
	QuoteValue     bool        `json:"quote_value"`     // Terrareg-specific field
	Sensitive      bool        `json:"sensitive"`       // Terrareg-specific field
}

// TerraregOutput represents a Terraform output.
type TerraregOutput struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Type        *string `json:"type"` // Additional terrareg field
}

// TerraregDependency represents a Terraform module dependency.
type TerraregDependency struct {
	Module  string `json:"module"`
	Source  string `json:"source"`
	Version string `json:"version"`
}

// TerraregProviderDep represents a Terraform provider dependency.
type TerraregProviderDep struct {
	Provider string `json:"provider"`
	Source   string `json:"source"`
	Version  string `json:"version"`
}

// TerraregResource represents a Terraform resource.
type TerraregResource struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// TerraregModule represents a Terraform module dependency.
// This is the additional 'modules' field in terrareg API (not in standard API).
type TerraregModule struct {
	Name      string   `json:"name"`
	Source    string   `json:"source"`
	Version   string   `json:"version"`
	Key       string   `json:"key"`       // Terrareg-specific field
	Providers []string `json:"providers"` // Terrareg-specific field
}
