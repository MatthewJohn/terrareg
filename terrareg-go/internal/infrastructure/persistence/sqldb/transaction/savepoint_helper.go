package transaction

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"gorm.io/gorm"
)

// Context key for storing transaction state
type contextKey string

const (
	TransactionDBKey contextKey = "transaction_db"
)

// SavepointHelper provides savepoint functionality for nested transactions
type SavepointHelper struct {
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

// NewSavepointHelper creates a new savepoint helper
func NewSavepointHelper(db *gorm.DB) *SavepointHelper {
	return &SavepointHelper{
		db: db,
	}
}

// Savepoint represents a database savepoint
type Savepoint struct {
	Name      string
	CreatedAt time.Time
	Active    bool
}

// WithTransaction executes a function within a database transaction or savepoint
// This is the unified method that handles all transaction scenarios:
// - Creates new transaction if none exists
// - Creates savepoint if transaction already exists
// - Ensures consistent context propagation with transaction
// - Provides both context and database instance to callback
func (h *SavepointHelper) WithTransaction(ctx context.Context, fn func(context.Context, *gorm.DB) error) error {
	return h.WithNamedTransaction(ctx, "", fn)
}

// WithNamedTransaction executes a function within a transaction or savepoint with a specific name
// If savepointName is empty, generates a unique name
func (h *SavepointHelper) WithNamedTransaction(ctx context.Context, savepointName string, fn func(context.Context, *gorm.DB) error) error {
	if savepointName == "" {
		savepointName = fmt.Sprintf("tx_%d", time.Now().UnixNano())
	}

	// Check if transaction already exists in context
	if tx, exists := h.getTransactionFromContext(ctx); exists {
		// Use existing transaction and create savepoint
		ctxWithTx := context.WithValue(ctx, TransactionDBKey, tx)
		return h.withSavepointAndTx(tx, savepointName, func(db *gorm.DB) error {
			return fn(ctxWithTx, db)
		})
	}

	// Create new transaction
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ctxWithTx := context.WithValue(ctx, TransactionDBKey, tx)
		return h.withSavepointAndTx(tx, savepointName, func(db *gorm.DB) error {
			return fn(ctxWithTx, db)
		})
	})
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

// ProcessBatchWithSavepoints processes multiple operations with individual savepoints
// Each operation gets its own savepoint for isolation
func (h *SavepointHelper) ProcessBatchWithSavepoints(ctx context.Context, operations []BatchOperation) *BatchResult {
	result := &BatchResult{
		Total:   len(operations),
		Success: []BatchOperationResult{},
		Failed:  []BatchOperationResult{},
	}

	for _, op := range operations {
		startTime := time.Now()

		// Each operation gets its own savepoint
		err := h.WithTransaction(ctx, func(ctx context.Context, tx *gorm.DB) error {
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

// WithContext returns the database instance for raw SQL execution
func (h *SavepointHelper) WithContext(ctx context.Context) *gorm.DB {
	return h.db.WithContext(ctx)
}

// WithTransactionContext executes a function within a transaction context
// If a transaction is already active in the context, it uses that transaction
// Otherwise, it creates a new transaction
func (h *SavepointHelper) WithTransactionContext(ctx context.Context, fn func(context.Context, *gorm.DB) error) error {
	// Check if transaction already exists in context
	if tx, exists := h.getTransactionFromContext(ctx); exists {
		// Use existing transaction
		return fn(ctx, tx)
	}

	// Create new transaction
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Store transaction in context
		ctxWithTx := context.WithValue(ctx, TransactionDBKey, tx)
		return fn(ctxWithTx, tx)
	})
}

// WithSavepointOrTransaction executes a function within a savepoint or transaction
// If a transaction exists in context, creates a savepoint
// Otherwise, creates a new transaction
func (h *SavepointHelper) WithSavepointOrTransaction(ctx context.Context, savepointName string, fn func(context.Context, *gorm.DB) error) error {
	// Check if transaction already exists in context
	if tx, exists := h.getTransactionFromContext(ctx); exists {
		// Use existing transaction and create savepoint
	// IMPORTANT: Ensure the context has the transaction stored for consistency
		// Even though tx exists in ctx, we double-verify for nested calls
		ctxWithTx := context.WithValue(ctx, TransactionDBKey, tx)
		return h.withSavepointAndTx(tx, savepointName, func(txDB *gorm.DB) error {
			return fn(ctxWithTx, txDB)
		})
	}

	// Create new transaction with savepoint
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ctxWithTx := context.WithValue(ctx, TransactionDBKey, tx)
		return h.withSavepointAndTx(tx, savepointName, func(txDB *gorm.DB) error {
			return fn(ctxWithTx, txDB)
		})
	})
}

// getTransactionFromContext retrieves the transaction from context
func (h *SavepointHelper) getTransactionFromContext(ctx context.Context) (*gorm.DB, bool) {
	if tx, ok := ctx.Value(TransactionDBKey).(*gorm.DB); ok {
		return tx, true
	}
	return nil, false
}

// IsTransactionActive checks if a transaction is active in the context
func (h *SavepointHelper) IsTransactionActive(ctx context.Context) bool {
	_, exists := h.getTransactionFromContext(ctx)
	return exists
}

// withSavepointAndTx creates a savepoint using a specific transaction instance
// This bypasses the normal WithSavepointNamed to ensure we use the correct transaction
func (h *SavepointHelper) withSavepointAndTx(tx *gorm.DB, savepointName string, fn func(*gorm.DB) error) error {
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

// WithLegacyTransaction provides backward compatibility for services using old callback signature
// DEPRECATED: Use WithTransaction instead which provides context in callback
func (h *SavepointHelper) WithLegacyTransaction(ctx context.Context, savepointName string, fn func(*gorm.DB) error) error {
	return h.WithNamedTransaction(ctx, savepointName, func(ctx context.Context, tx *gorm.DB) error {
		return fn(tx)
	})
}

// WithSmartSavepointOrTransaction is an alias for WithLegacyTransaction for backward compatibility
// DEPRECATED: Use WithTransaction or WithLegacyTransaction instead
func (h *SavepointHelper) WithSmartSavepointOrTransaction(ctx context.Context, savepointName string, fn func(*gorm.DB) error) error {
	return h.WithLegacyTransaction(ctx, savepointName, fn)
}
