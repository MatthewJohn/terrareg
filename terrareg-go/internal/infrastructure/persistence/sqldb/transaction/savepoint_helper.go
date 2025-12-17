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

// WithTransaction executes a function within a database transaction
func (h *SavepointHelper) WithTransaction(ctx context.Context, fn func(*gorm.DB) error) error {
	return h.db.WithContext(ctx).Transaction(fn)
}

// WithSavepoint creates a nested transaction (savepoint) and executes the function
// If the function returns an error, the savepoint is rolled back
// If successful, the savepoint is committed
func (h *SavepointHelper) WithSavepoint(ctx context.Context, fn func(*gorm.DB) error) error {
	savepointName := fmt.Sprintf("sp_%d", time.Now().UnixNano())

	// Create savepoint using raw SQL with proper quoting for cross-database compatibility
	if err := h.db.WithContext(ctx).Exec(fmt.Sprintf("SAVEPOINT `%s`", savepointName)).Error; err != nil {
		return fmt.Errorf("failed to create savepoint %s: %w", savepointName, err)
	}

	// Execute the function
	err := fn(h.db.WithContext(ctx))

	if err != nil {
		// Rollback to savepoint on error
		if rollbackErr := h.db.WithContext(ctx).Exec(fmt.Sprintf("ROLLBACK TO SAVEPOINT `%s`", savepointName)).Error; rollbackErr != nil {
			return fmt.Errorf("failed to rollback to savepoint %s: %w (original error: %v)", savepointName, rollbackErr, err)
		}
		return err
	}

	// Release the savepoint on success
	if releaseErr := h.db.WithContext(ctx).Exec(fmt.Sprintf("RELEASE SAVEPOINT `%s`", savepointName)).Error; releaseErr != nil {
		return fmt.Errorf("failed to release savepoint %s: %w", savepointName, releaseErr)
	}

	return nil
}

// WithSavepointNamed creates a savepoint with a specific name
func (h *SavepointHelper) WithSavepointNamed(ctx context.Context, savepointName string, fn func(*gorm.DB) error) error {
	// Sanitize savepoint name to be SQL-safe across all databases (SQLite, MySQL, PostgreSQL)
	safeName := sanitizeSavepointName(savepointName)

	// Create savepoint using raw SQL with proper quoting for cross-database compatibility
	// Different databases have different quoting rules:
	// - PostgreSQL and MySQL: `identifier` or "identifier"
	// - SQLite: `identifier` or "identifier" or [identifier]
	// We'll use backticks for MySQL/SQLite and PostgreSQL compatibility
	if err := h.db.WithContext(ctx).Exec(fmt.Sprintf("SAVEPOINT `%s`", safeName)).Error; err != nil {
		return fmt.Errorf("failed to create savepoint %s: %w", safeName, err)
	}

	// Execute the function
	err := fn(h.db.WithContext(ctx))

	if err != nil {
		// Rollback to savepoint on error
		if rollbackErr := h.db.WithContext(ctx).Exec(fmt.Sprintf("ROLLBACK TO SAVEPOINT `%s`", safeName)).Error; rollbackErr != nil {
			return fmt.Errorf("failed to rollback to savepoint %s: %w (original error: %v)", safeName, rollbackErr, err)
		}
		return err
	}

	// Release the savepoint on success
	if releaseErr := h.db.WithContext(ctx).Exec(fmt.Sprintf("RELEASE SAVEPOINT `%s`", safeName)).Error; releaseErr != nil {
		return fmt.Errorf("failed to release savepoint %s: %w", safeName, releaseErr)
	}

	return nil
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
		err := h.WithSavepoint(ctx, func(tx *gorm.DB) error {
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
	if _, exists := h.getTransactionFromContext(ctx); exists {
		// Use existing transaction and create savepoint
		return h.WithSavepointNamed(ctx, savepointName, func(txDB *gorm.DB) error {
			return fn(ctx, txDB)
		})
	}

	// Create new transaction with savepoint
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ctxWithTx := context.WithValue(ctx, TransactionDBKey, tx)
		return h.WithSavepointNamed(ctxWithTx, savepointName, func(txDB *gorm.DB) error {
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

// WithSmartSavepointOrTransaction is a helper for services to use
// It automatically detects if a transaction is active and uses the appropriate wrapper
// This is a simplified wrapper that doesn't modify the function signature for existing services
func (h *SavepointHelper) WithSmartSavepointOrTransaction(ctx context.Context, savepointName string, fn func(*gorm.DB) error) error {
	if h.IsTransactionActive(ctx) {
		// Use existing transaction with savepoint
		return h.WithSavepointNamed(ctx, savepointName, fn)
	}

	// Create new transaction with savepoint
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ctxWithTx := context.WithValue(ctx, TransactionDBKey, tx)
		return h.WithSavepointNamed(ctxWithTx, savepointName, fn)
	})
}
