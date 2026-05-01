package repository

import (
	"context"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/transaction"
	"gorm.io/gorm"
)

// BaseRepository provides common database context handling for all repositories.
// This eliminates duplicate getDBFromContext implementations across repositories.
type BaseRepository struct {
	// db provides database access (required)
	db *gorm.DB
	// txManager provides transaction savepoint management (required)
	txManager *transaction.GormTransactionEngine
}

// NewBaseRepository creates a new base repository.
// Returns an error if db is nil.
func NewBaseRepository(db *gorm.DB) (*BaseRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("db cannot be nil")
	}

	txEngine, err := transaction.NewGormTransactionEngine(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction engine: %w", err)
	}

	// Type assert to get the concrete GormTransactionEngine
	// This is safe because NewGormTransactionEngine always returns *GormTransactionEngine
	engine, ok := txEngine.(*transaction.GormTransactionEngine)
	if !ok {
		return nil, fmt.Errorf("unexpected transaction manager type")
	}

	return &BaseRepository{
		db:        db,
		txManager: engine,
	}, nil
}

// NewBaseRepositoryWithoutTxManager creates a base repository without a transaction manager.
// Use this when you need a base repository but don't have a transaction manager yet.
// This is used internally by other repositories that need to create a BaseRepository.
func NewBaseRepositoryWithoutTxManager(db *gorm.DB) *BaseRepository {
	return &BaseRepository{
		db: db,
		// txManager will be initialized later
	}
}

// SetTxManager sets the transaction manager after creation.
// This is useful for repositories that are created in multiple steps.
func (r *BaseRepository) SetTxManager(txManager *transaction.GormTransactionEngine) {
	r.txManager = txManager
}

// GetDBFromContext returns the appropriate database instance for the given context.
// If a transaction is active in the context, it uses that transaction.
// Otherwise, it returns the database instance with the context applied.
func (r *BaseRepository) GetDBFromContext(ctx context.Context) *gorm.DB {
	// Check if transaction is active in context
	if r.txManager != nil && r.txManager.IsTransactionActive(ctx) {
		if tx := r.txManager.GetDBFromContext(ctx); tx != nil {
			return tx
		}
	}
	// No active transaction, use db with context
	return r.txManager.WithContext(ctx)
}

// IsTransactionActive checks if a transaction is active in the context.
func (r *BaseRepository) IsTransactionActive(ctx context.Context) bool {
	if r.txManager == nil {
		return false
	}
	return r.txManager.IsTransactionActive(ctx)
}

// GetTransactionManager returns the transaction manager for this repository.
// This allows other infrastructure code to access transaction management if needed.
func (r *BaseRepository) GetTransactionManager() *transaction.GormTransactionEngine {
	return r.txManager
}
