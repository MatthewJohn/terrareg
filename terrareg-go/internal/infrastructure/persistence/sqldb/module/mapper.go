package module

import (
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// Mapper functions between domain models and database models

// toDBNamespace converts domain Namespace to database model
func toDBNamespace(ns *model.Namespace) sqldb.NamespaceDB {
	return sqldb.NamespaceDB{
		ID:            ns.ID(),
		Namespace:     ns.Name(),
		DisplayName:   ns.DisplayName(),
		NamespaceType: sqldb.NamespaceType(ns.Type()),
	}
}

// fromDBNamespace converts database model to domain Namespace
func fromDBNamespace(db *sqldb.NamespaceDB) *model.Namespace {
	return model.ReconstructNamespace(
		db.ID,
		db.Namespace,
		db.DisplayName,
		model.NamespaceType(db.NamespaceType),
	)
}

// toDBModuleProvider converts domain ModuleProvider to database model
func toDBModuleProvider(mp *model.ModuleProvider) sqldb.ModuleProviderDB {
	var verified *bool
	if mp.IsVerified() {
		v := true
		verified = &v
	}

	var latestVersionID *int
	if mp.GetLatestVersion() != nil {
		id := mp.GetLatestVersion().ID()
		latestVersionID = &id
	}

	return sqldb.ModuleProviderDB{
		ID:                    mp.ID(),
		NamespaceID:           mp.Namespace().ID(),
		Module:                mp.Module(),
		Provider:              mp.Provider(),
		Verified:              verified,
		GitProviderID:         mp.GitProviderID(),
		RepoBaseURLTemplate:   mp.RepoBaseURLTemplate(),
		RepoCloneURLTemplate:  mp.RepoCloneURLTemplate(),
		RepoBrowseURLTemplate: mp.RepoBrowseURLTemplate(),
		GitTagFormat:          mp.GitTagFormat(),
		GitPath:               mp.GitPath(),
		ArchiveGitPath:        mp.ArchiveGitPath(),
		LatestVersionID:       latestVersionID,
	}
}

// fromDBModuleProvider converts database model to domain ModuleProvider
func fromDBModuleProvider(db *sqldb.ModuleProviderDB, namespace *model.Namespace) *model.ModuleProvider {
	verified := false
	if db.Verified != nil {
		verified = *db.Verified
	}

	// Use current time for now since database doesn't store timestamps
	now := time.Now()
	mp := model.ReconstructModuleProvider(
		db.ID,
		namespace,
		db.Module,
		db.Provider,
		verified,
		db.GitProviderID,
		db.RepoBaseURLTemplate,
		db.RepoCloneURLTemplate,
		db.RepoBrowseURLTemplate,
		db.GitTagFormat,
		db.GitPath,
		db.ArchiveGitPath,
		now,
		now,
	)

	return mp
}

// fromDBModuleVersion converts database model to domain ModuleVersion
func fromDBModuleVersion(db *sqldb.ModuleVersionDB, details *model.ModuleDetails) (*model.ModuleVersion, error) {
	// Use empty details if none provided
	if details == nil {
		details = model.NewModuleDetails([]byte{})
	}

	// Determine if version is published
	published := db.Published != nil && *db.Published

	// Handle nil PublishedAt pointer - use current time if not set
	publishedAt := time.Now()
	if db.PublishedAt != nil {
		publishedAt = *db.PublishedAt
	}

	return model.ReconstructModuleVersion(
		db.ID,
		db.Version,
		details,
		db.Beta,
		db.Internal,
		published,
		db.PublishedAt,
		db.GitSHA,
		db.GitPath,
		db.ArchiveGitPath,
		db.RepoBaseURLTemplate,
		db.RepoCloneURLTemplate,
		db.RepoBrowseURLTemplate,
		db.Owner,
		db.Description,
		db.VariableTemplate,
		db.ExtractionVersion,
		publishedAt, // Use publishedAt as createdAt for simplicity
		publishedAt, // Use publishedAt as updatedAt for simplicity
	)
}

// toDBModuleVersion converts domain ModuleVersion to database model
func toDBModuleVersion(mv *model.ModuleVersion) sqldb.ModuleVersionDB {
	var moduleProviderID int
	if mv.ModuleProvider() != nil {
		moduleProviderID = mv.ModuleProvider().ID()
	}

	var detailsID *int
	// Details ID would be set by a separate save operation

	var published *bool
	if mv.IsPublished() {
		p := true
		published = &p
	}

	return sqldb.ModuleVersionDB{
		ID:                    mv.ID(),
		ModuleProviderID:      moduleProviderID,
		Version:               mv.Version().String(),
		Beta:                  mv.IsBeta(),
		Internal:              mv.IsInternal(),
		Published:             published,
		PublishedAt:           mv.PublishedAt(),
		GitSHA:                mv.GitSHA(),
		GitPath:               mv.GitPath(),
		ArchiveGitPath:        mv.ArchiveGitPath(),
		RepoBaseURLTemplate:   mv.RepoBaseURLTemplate(),
		RepoCloneURLTemplate:  mv.RepoCloneURLTemplate(),
		RepoBrowseURLTemplate: mv.RepoBrowseURLTemplate(),
		Owner:                 mv.Owner(),
		Description:           mv.Description(),
		VariableTemplate:      mv.VariableTemplate(),
		ExtractionVersion:     mv.ExtractionVersion(),
		ModuleDetailsID:       detailsID,
	}
}

// toDBModuleDetails converts domain ModuleDetails to database model
func toDBModuleDetails(md *model.ModuleDetails) sqldb.ModuleDetailsDB {
	return sqldb.ModuleDetailsDB{
		ReadmeContent:    md.ReadmeContent(),
		TerraformDocs:    md.TerraformDocs(),
		Tfsec:            md.Tfsec(),
		Infracost:        md.Infracost(),
		TerraformGraph:   md.TerraformGraph(),
		TerraformModules: md.TerraformModules(),
		TerraformVersion: []byte(md.TerraformVersion()),
	}
}

// fromDBModuleDetails converts database model to domain ModuleDetails
func fromDBModuleDetails(db *sqldb.ModuleDetailsDB) *model.ModuleDetails {
	if db == nil {
		return nil
	}

	terraformVersion := ""
	if len(db.TerraformVersion) > 0 {
		terraformVersion = string(db.TerraformVersion)
	}

	return model.NewCompleteModuleDetails(
		db.ReadmeContent,
		db.TerraformDocs,
		db.Tfsec,
		db.Infracost,
		db.TerraformGraph,
		db.TerraformModules,
		terraformVersion,
	)
}

