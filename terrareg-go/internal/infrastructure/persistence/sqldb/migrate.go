package sqldb

// GetAllModels returns a slice of all model types for migration.
// This is the canonical source for which models should be migrated.
func GetAllModels() []interface{} {
	return []interface{}{
		&SessionDB{},
		&TerraformIDPAuthorizationCodeDB{},
		&TerraformIDPAccessTokenDB{},
		&TerraformIDPSubjectIdentifierDB{},
		&UserGroupDB{},
		&NamespaceDB{},
		&UserGroupNamespacePermissionDB{},
		&GitProviderDB{},
		&NamespaceRedirectDB{},
		&ModuleDetailsDB{},
		&ModuleProviderDB{},
		&ModuleVersionDB{},
		&ModuleProviderRedirectDB{},
		&SubmoduleDB{},
		&AnalyticsDB{},
		&ProviderAnalyticsDB{},
		&ExampleFileDB{},
		&ModuleVersionFileDB{},
		&GPGKeyDB{},
		&ProviderSourceDB{},
		&ProviderCategoryDB{},
		&RepositoryDB{},
		&ProviderDB{},
		&ProviderVersionDB{},
		&ProviderVersionDocumentationDB{},
		&ProviderVersionBinaryDB{},
		&AuthenticationTokenDB{},
		&AuditHistoryDB{},
	}
}

// AutoMigrate runs GORM auto-migration for all models.
func (db *Database) AutoMigrate() error {
	return db.DB.AutoMigrate(GetAllModels()...)
}

// AutoMigrateModels runs GORM auto-migration for specific models.
// Use this for isolated unit tests that only need certain tables.
func (db *Database) AutoMigrateModels(models ...interface{}) error {
	return db.DB.AutoMigrate(models...)
}
