package model

// ModuleProcessingConfig represents module processing configuration settings
// Contains all settings related to how modules are processed and analyzed
type ModuleProcessingConfig struct {
	AutoPublishModuleVersions                   ModuleVersionReindexMode `json:"auto_publish_module_versions"`
	ModuleVersionReindexMode                    ModuleVersionReindexMode `json:"module_version_reindex_mode"`
	ModuleVersionUseGitCommit                   bool                     `json:"module_version_use_git_commit"`
	RequiredModuleMetadataAttributes            []string                 `json:"required_module_metadata_attributes"`
	DeleteExternallyHostedArtifacts             bool                     `json:"delete_externally_hosted_artifacts"`
	AutogenerateModuleProviderDescription       bool                     `json:"autogenerate_module_provider_description"`
	AutogenerateUsageBuilderVariables           bool                     `json:"autogenerate_usage_builder_variables"`
	AllowForcefulModuleProviderRedirectDeletion bool                     `json:"allow_forceful_module_provider_redirect_deletion"`
	RedirectDeletionLookbackDays                int                      `json:"redirect_deletion_lookback_days"`
}

// TerraformIntegrationConfig represents Terraform-specific integration settings
type TerraformIntegrationConfig struct {
	Product                 Product `json:"product"`
	DefaultTerraformVersion string  `json:"default_terraform_version"`
	TerraformArchiveMirror  string  `json:"terraform_archive_mirror"`
	ManageTerraformRcFile   bool    `json:"manage_terraform_rc_file"`
	ModulesDirectory        string  `json:"modules_directory"`
	ExamplesDirectory       string  `json:"examples_directory"`
}

// ExampleConfiguration represents example file and display settings
type ExampleConfiguration struct {
	ExampleFileExtensions                   []string `json:"example_file_extensions"`
	TerraformExampleVersionTemplate         string   `json:"terraform_example_version_template"`
	TerraformExampleVersionTemplatePreMajor string   `json:"terraform_example_version_template_pre_major"`
}

// IsValid validates the module processing configuration
func (m *ModuleProcessingConfig) IsValid() bool {
	return m.ModuleVersionReindexMode.IsValid() && m.RedirectDeletionLookbackDays >= -1
}

// IsValid validates the Terraform integration configuration
func (t *TerraformIntegrationConfig) IsValid() bool {
	return t.Product.IsValid() && t.DefaultTerraformVersion != ""
}

// IsValid validates the example configuration
func (e *ExampleConfiguration) IsValid() bool {
	return len(e.ExampleFileExtensions) > 0 && e.TerraformExampleVersionTemplate != ""
}

// GetDefaultReindexMode returns the default reindex mode if not set
func (m *ModuleProcessingConfig) GetDefaultReindexMode() ModuleVersionReindexMode {
	if m.ModuleVersionReindexMode == "" {
		return ModuleVersionReindexModeLegacy
	}
	return m.ModuleVersionReindexMode
}

// GetDefaultProduct returns the default product if not set
func (t *TerraformIntegrationConfig) GetDefaultProduct() Product {
	if t.Product == "" {
		return ProductTerraform
	}
	return t.Product
}

// ShouldAllowRedirectDeletion checks if redirect deletion should be allowed based on lookback days
func (m *ModuleProcessingConfig) ShouldAllowRedirectDeletion(lastAccessDays int) bool {
	// If lookback days is 0, always allow deletion
	if m.RedirectDeletionLookbackDays == 0 {
		return true
	}

	// If lookback days is -1, never allow forceful deletion
	if m.RedirectDeletionLookbackDays == -1 {
		return false
	}

	// Otherwise, allow if last access was before lookback period
	return lastAccessDays >= m.RedirectDeletionLookbackDays
}
