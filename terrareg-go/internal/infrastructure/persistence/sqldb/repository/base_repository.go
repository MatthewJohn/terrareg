package repository

import (
	"context"

	"gorm.io/gorm"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/transaction"
)

// BaseRepository provides common database context handling for all repositories
// This eliminates duplicate getDBFromContext implementations across repositories
type BaseRepository struct {
	db *gorm.DB
	helper *transaction.SavepointHelper
}

// NewBaseRepository creates a new base repository
func NewBaseRepository(db *gorm.DB) *BaseRepository {
	return &BaseRepository{
		db: db,
		helper: transaction.NewSavepointHelper(db),
	}
}

// GetDBFromContext returns the appropriate database instance for the given context
// If a transaction is active in the context, it uses that transaction
// Otherwise, it returns the database instance with the context applied
func (r *BaseRepository) GetDBFromContext(ctx context.Context) *gorm.DB {
	// Check if transaction is active in context
	if r.helper.IsTransactionActive(ctx) {
		if tx, exists := ctx.Value(transaction.TransactionDBKey).(*gorm.DB); exists {
			return tx
		}
	}

	// No active transaction, use db with context
	return r.helper.WithContext(ctx)
}

// IsTransactionActive checks if a transaction is active in the context
func (r *BaseRepository) IsTransactionActive(ctx context.Context) bool {
	return r.helper.IsTransactionActive(ctx)
}

// GetDB returns the base database connection for operations that don't have context
func (r *BaseRepository) GetDB() *gorm.DB {
	return r.db
}