package git

import (
	gitmodel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/git/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

func FromDBGitProvider(db *sqldb.GitProviderDB) *gitmodel.GitProvider {
	if db == nil {
		return nil
	}
	return &gitmodel.GitProvider{
		ID:                  db.ID,
		Name:                db.Name,
		BaseURLTemplate:     db.BaseURLTemplate,
		CloneURLTemplate:    db.CloneURLTemplate,
		BrowseURLTemplate:   db.BrowseURLTemplate,
		GitPathTemplate:     db.GitPathTemplate,
	}
}
