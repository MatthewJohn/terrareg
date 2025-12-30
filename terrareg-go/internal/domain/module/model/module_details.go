package model

import "bytes"

// ModuleDetails is a value object containing module metadata
// It is immutable - any updates create a new instance
type ModuleDetails struct {
	readmeContent    []byte
	terraformDocs    []byte
	tfsec            []byte
	infracost        []byte
	terraformGraph   []byte
	terraformModules []byte
	terraformVersion string
}

// NewModuleDetails creates a new module details
func NewModuleDetails(readme []byte) *ModuleDetails {
	return &ModuleDetails{
		readmeContent: readme,
	}
}

// NewCompleteModuleDetails creates module details with all fields
func NewCompleteModuleDetails(
	readme, terraformDocs, tfsec, infracost, terraformGraph, terraformModules []byte,
	terraformVersion string,
) *ModuleDetails {
	return &ModuleDetails{
		readmeContent:    readme,
		terraformDocs:    terraformDocs,
		tfsec:            tfsec,
		infracost:        infracost,
		terraformGraph:   terraformGraph,
		terraformModules: terraformModules,
		terraformVersion: terraformVersion,
	}
}

// Immutable updates - return new instances

// WithTerraformDocs returns a new instance with updated terraform docs
func (md *ModuleDetails) WithTerraformDocs(docs []byte) *ModuleDetails {
	return &ModuleDetails{
		readmeContent:    md.readmeContent,
		terraformDocs:    docs,
		tfsec:            md.tfsec,
		infracost:        md.infracost,
		terraformGraph:   md.terraformGraph,
		terraformModules: md.terraformModules,
		terraformVersion: md.terraformVersion,
	}
}

// WithTfsec returns a new instance with updated tfsec results
func (md *ModuleDetails) WithTfsec(tfsec []byte) *ModuleDetails {
	return &ModuleDetails{
		readmeContent:    md.readmeContent,
		terraformDocs:    md.terraformDocs,
		tfsec:            tfsec,
		infracost:        md.infracost,
		terraformGraph:   md.terraformGraph,
		terraformModules: md.terraformModules,
		terraformVersion: md.terraformVersion,
	}
}

// WithInfracost returns a new instance with updated infracost results
func (md *ModuleDetails) WithInfracost(infracost []byte) *ModuleDetails {
	return &ModuleDetails{
		readmeContent:    md.readmeContent,
		terraformDocs:    md.terraformDocs,
		tfsec:            md.tfsec,
		infracost:        infracost,
		terraformGraph:   md.terraformGraph,
		terraformModules: md.terraformModules,
		terraformVersion: md.terraformVersion,
	}
}

// WithTerraformGraph returns a new instance with updated terraform graph
func (md *ModuleDetails) WithTerraformGraph(graph []byte) *ModuleDetails {
	return &ModuleDetails{
		readmeContent:    md.readmeContent,
		terraformDocs:    md.terraformDocs,
		tfsec:            md.tfsec,
		infracost:        md.infracost,
		terraformGraph:   graph,
		terraformModules: md.terraformModules,
		terraformVersion: md.terraformVersion,
	}
}

// WithTerraformModules returns a new instance with updated terraform modules
func (md *ModuleDetails) WithTerraformModules(modules []byte) *ModuleDetails {
	return &ModuleDetails{
		readmeContent:    md.readmeContent,
		terraformDocs:    md.terraformDocs,
		tfsec:            md.tfsec,
		infracost:        md.infracost,
		terraformGraph:   md.terraformGraph,
		terraformModules: modules,
		terraformVersion: md.terraformVersion,
	}
}

// WithTerraformVersion returns a new instance with updated terraform version
func (md *ModuleDetails) WithTerraformVersion(version string) *ModuleDetails {
	return &ModuleDetails{
		readmeContent:    md.readmeContent,
		terraformDocs:    md.terraformDocs,
		tfsec:            md.tfsec,
		infracost:        md.infracost,
		terraformGraph:   md.terraformGraph,
		terraformModules: md.terraformModules,
		terraformVersion: version,
	}
}

// Getters

func (md *ModuleDetails) ReadmeContent() []byte {
	if md == nil {
		return []byte{}
	}
	return md.readmeContent
}

func (md *ModuleDetails) TerraformDocs() []byte {
	if md == nil {
		return nil
	}
	return md.terraformDocs
}

func (md *ModuleDetails) Tfsec() []byte {
	if md == nil {
		return nil
	}
	return md.tfsec
}

func (md *ModuleDetails) Infracost() []byte {
	if md == nil {
		return nil
	}
	return md.infracost
}

func (md *ModuleDetails) TerraformGraph() []byte {
	if md == nil {
		return nil
	}
	return md.terraformGraph
}

func (md *ModuleDetails) TerraformModules() []byte {
	if md == nil {
		return nil
	}
	return md.terraformModules
}

func (md *ModuleDetails) TerraformVersion() string {
	if md == nil {
		return ""
	}
	return md.terraformVersion
}

// Value object equality - compare by value, not identity

// Equals checks if two ModuleDetails are equal
func (md *ModuleDetails) Equals(other *ModuleDetails) bool {
	if other == nil {
		return false
	}

	return bytes.Equal(md.readmeContent, other.readmeContent) &&
		bytes.Equal(md.terraformDocs, other.terraformDocs) &&
		bytes.Equal(md.tfsec, other.tfsec) &&
		bytes.Equal(md.infracost, other.infracost) &&
		bytes.Equal(md.terraformGraph, other.terraformGraph) &&
		bytes.Equal(md.terraformModules, other.terraformModules) &&
		md.terraformVersion == other.terraformVersion
}

// HasReadme returns true if there is README content
func (md *ModuleDetails) HasReadme() bool {
	if md == nil {
		return false
	}
	return len(md.readmeContent) > 0
}

// HasTerraformDocs returns true if there are Terraform docs
func (md *ModuleDetails) HasTerraformDocs() bool {
	if md == nil {
		return false
	}
	return len(md.terraformDocs) > 0
}

// HasTfsec returns true if there are tfsec results
func (md *ModuleDetails) HasTfsec() bool {
	if md == nil {
		return false
	}
	return len(md.tfsec) > 0
}

// HasInfracost returns true if there are infracost results
func (md *ModuleDetails) HasInfracost() bool {
	if md == nil {
		return false
	}
	return len(md.infracost) > 0
}

// HasTerraformGraph returns true if there is a terraform graph
func (md *ModuleDetails) HasTerraformGraph() bool {
	if md == nil {
		return false
	}
	return len(md.terraformGraph) > 0
}

// HasTerraformModules returns true if there are terraform modules
func (md *ModuleDetails) HasTerraformModules() bool {
	if md == nil {
		return false
	}
	return len(md.terraformModules) > 0
}
