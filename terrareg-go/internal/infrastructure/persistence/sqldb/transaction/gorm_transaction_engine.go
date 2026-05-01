package transaction

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"gorm.io/gorm"

	domainTransaction "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/transaction"
)

// Context key for storing transaction state
type contextKey string

const (
	TransactionDBKey contextKey = "transaction_db"
)

// GormTransactionEngine provides GORM transaction operations with savepoint functionality.
// It implements the domain.TransactionManager interface, enabling domain services to use
// transaction management without depending on GORM directly.
//
// For infrastructure code that needs GORM access (like repositories), use GetDBFromContext().
type GormTransactionEngine struct {
	// db provides database access (required)
	db *gorm.DB
}

// sanitizeSavepointName converts a string to a SQL-safe identifier
// Replaces invalid characters with underscores and ensures it starts with a letter/underscore
func sanitizeSavepointName(name string) string {
	// Replace invalid SQL identifier characters with underscores
	// Valid characters: letters, digits, underscores (cannot start with digit)
	re := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	sanitized := re.ReplaceAllString(name, "_")

	// Ensure it doesn't start with a digit (SQL identifiers can't start with digits)
	if len(sanitized) > 0 && sanitized[0] >= '0' && sanitized[0] <= '9' {
		sanitized = "sp_" + sanitized
	}

	// Ensure it's not empty
	if sanitized == "" {
		sanitized = fmt.Sprintf("sp_%d", time.Now().UnixNano())
	}

	// Truncate if too long (SQL identifiers have length limits)
	if len(sanitized) > 64 {
		sanitized = sanitized[:61] + fmt.Sprintf("_%d", time.Now().UnixNano()%1000)
	}

	return sanitized
}

// NewGormTransactionEngine creates a new GORM transaction engine.
// Returns an error if db is nil.
//
// The returned value implements domain.TransactionManager.
func NewGormTransactionEngine(db *gorm.DB) (domainTransaction.TransactionManager, error) {
	if db == nil {
		return nil, fmt.Errorf("db cannot be nil")
	}
	return &GormTransactionEngine{
		db: db,
	}, nil
}

// Ensure GormTransactionEngine implements the domain interface
var _ domainTransaction.TransactionManager = (*GormTransactionEngine)(nil)

// WithTransaction executes a function within a database transaction or savepoint.
// This is the unified method that handles all transaction scenarios:
// - Creates new transaction if none exists
// - Creates savepoint if transaction already exists
// - Ensures consistent context propagation with transaction
//
// Part of the domain.TransactionManager interface.
func (e *GormTransactionEngine) WithTransaction(ctx context.Context, fn func(context.Context) error) error {
	return e.WithNamedTransaction(ctx, "", fn)
}

// WithNamedTransaction executes a function within a transaction or savepoint with a specific name.
// If savepointName is empty, generates a unique name.
//
// Part of the domain.TransactionManager interface.
func (e *GormTransactionEngine) WithNamedTransaction(ctx context.Context, savepointName string, fn func(context.Context) error) error {
	if savepointName == "" {
		savepointName = fmt.Sprintf("tx_%d", time.Now().UnixNano())
	}

	// Check if transaction already exists in context
	if tx, exists := e.getTransactionFromContext(ctx); exists {
		// Use existing transaction and create savepoint
		ctxWithTx := context.WithValue(ctx, TransactionDBKey, tx)
		return e.withSavepointAndTx(tx, savepointName, func(db *gorm.DB) error {
			return fn(ctxWithTx)
		})
	}

	// Create new transaction
	return e.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ctxWithTx := context.WithValue(ctx, TransactionDBKey, tx)
		return e.withSavepointAndTx(tx, savepointName, func(db *gorm.DB) error {
			return fn(ctxWithTx)
		})
	})
}

// IsTransactionActive checks if a transaction is active in the context.
//
// Part of the domain.TransactionManager interface.
func (e *GormTransactionEngine) IsTransactionActive(ctx context.Context) bool {
	_, exists := e.getTransactionFromContext(ctx)
	return exists
}

// GetDBFromContext retrieves the GORM DB instance from the context if a transaction is active.
// This is an infrastructure-level method for repository implementations that need direct GORM access.
// This method is NOT part of the domain interface and should only be used in the infrastructure layer.
func (e *GormTransactionEngine) GetDBFromContext(ctx context.Context) *gorm.DB {
	if tx, exists := e.getTransactionFromContext(ctx); exists {
		return tx
	}
	return nil
}

// WithContext returns the GORM DB instance with the given context applied.
// This is an infrastructure-level method for repository implementations.
func (e *GormTransactionEngine) WithContext(ctx context.Context) *gorm.DB {
	return e.db.WithContext(ctx)
}

// getTransactionFromContext retrieves the transaction from context
func (e *GormTransactionEngine) getTransactionFromContext(ctx context.Context) (*gorm.DB, bool) {
	if tx, ok := ctx.Value(TransactionDBKey).(*gorm.DB); ok {
		return tx, true
	}
	return nil, false
}

// withSavepointAndTx creates a savepoint using a specific transaction instance
func (e *GormTransactionEngine) withSavepointAndTx(tx *gorm.DB, savepointName string, fn func(*gorm.DB) error) error {
	// Sanitize savepoint name to be SQL-safe
	safeName := sanitizeSavepointName(savepointName)

	// Create savepoint using the specific transaction instance
	if err := tx.Exec(fmt.Sprintf("SAVEPOINT `%s`", safeName)).Error; err != nil {
		return fmt.Errorf("failed to create savepoint %s: %w", safeName, err)
	}

	// Execute the function
	err := fn(tx)

	if err != nil {
		// Rollback to savepoint on error
		if rollbackErr := tx.Exec(fmt.Sprintf("ROLLBACK TO SAVEPOINT `%s`", safeName)).Error; rollbackErr != nil {
			return fmt.Errorf("failed to rollback to savepoint %s: %w (original error: %v)", safeName, rollbackErr, err)
		}
		return err
	}

	// Release the savepoint on success
	if releaseErr := tx.Exec(fmt.Sprintf("RELEASE SAVEPOINT `%s`", safeName)).Error; releaseErr != nil {
		return fmt.Errorf("failed to release savepoint %s: %w", safeName, releaseErr)
	}

	return nil
}

// WithLegacyTransaction provides backward compatibility for services using old callback signature.
// DEPRECATED: Use WithTransaction instead which provides context in callback.
func (e *GormTransactionEngine) WithLegacyTransaction(ctx context.Context, savepointName string, fn func(*gorm.DB) error) error {
	return e.WithNamedTransaction(ctx, savepointName, func(ctx context.Context) error {
		return fn(e.GetDBFromContext(ctx))
	})
}

// WithSmartSavepointOrTransaction is an alias for WithLegacyTransaction for backward compatibility.
// DEPRECATED: Use WithTransaction or WithLegacyTransaction instead.
func (e *GormTransactionEngine) WithSmartSavepointOrTransaction(ctx context.Context, savepointName string, fn func(*gorm.DB) error) error {
	return e.WithLegacyTransaction(ctx, savepointName, fn)
}

// Savepoint represents a database savepoint
type Savepoint struct {
	Name      string
	CreatedAt time.Time
	Active    bool
}

// BatchOperation represents an operation in a batch with results
type BatchOperation struct {
	ID       string
	Function func(*gorm.DB) error
}

// BatchResult represents the result of a batch operation
type BatchResult struct {
	Total   int
	Success []BatchOperationResult
	Failed  []BatchOperationResult
}

// BatchOperationResult represents the result of a single batch operation
type BatchOperationResult struct {
	ID       string
	Success  bool
	Error    error
	Duration time.Duration
}

// ProcessBatchWithSavepoints processes multiple operations with individual savepoints.
// Each operation gets its own savepoint for isolation.
func (e *GormTransactionEngine) ProcessBatchWithSavepoints(ctx context.Context, operations []BatchOperation) *BatchResult {
	result := &BatchResult{
		Total:   len(operations),
		Success: []BatchOperationResult{},
		Failed:  []BatchOperationResult{},
	}

	for _, op := range operations {
		startTime := time.Now()

		// Each operation gets its own savepoint
		err := e.WithTransaction(ctx, func(ctx context.Context) error {
			tx := e.GetDBFromContext(ctx)
			return op.Function(tx)
		})

		duration := time.Since(startTime)

		opResult := BatchOperationResult{
			ID:       op.ID,
			Success:  err == nil,
			Error:    err,
			Duration: duration,
		}

		if err == nil {
			result.Success = append(result.Success, opResult)
		} else {
			result.Failed = append(result.Failed, opResult)
		}
	}

	return result
}
