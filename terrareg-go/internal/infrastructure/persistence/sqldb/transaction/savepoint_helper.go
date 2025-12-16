package transaction

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// SavepointHelper provides savepoint functionality for nested transactions
type SavepointHelper struct {
	db *gorm.DB
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

	// Create savepoint using raw SQL
	if err := h.db.WithContext(ctx).Exec(fmt.Sprintf("SAVEPOINT %s", savepointName)).Error; err != nil {
		return fmt.Errorf("failed to create savepoint %s: %w", savepointName, err)
	}

	// Execute the function
	err := fn(h.db.WithContext(ctx))

	if err != nil {
		// Rollback to savepoint on error
		if rollbackErr := h.db.WithContext(ctx).Exec(fmt.Sprintf("ROLLBACK TO SAVEPOINT %s", savepointName)).Error; rollbackErr != nil {
			return fmt.Errorf("failed to rollback to savepoint %s: %w (original error: %v)", savepointName, rollbackErr, err)
		}
		return err
	}

	// Release the savepoint on success
	if releaseErr := h.db.WithContext(ctx).Exec(fmt.Sprintf("RELEASE SAVEPOINT %s", savepointName)).Error; releaseErr != nil {
		return fmt.Errorf("failed to release savepoint %s: %w", savepointName, releaseErr)
	}

	return nil
}

// WithSavepointNamed creates a savepoint with a specific name
func (h *SavepointHelper) WithSavepointNamed(ctx context.Context, savepointName string, fn func(*gorm.DB) error) error {
	// Create savepoint using raw SQL
	if err := h.db.WithContext(ctx).Exec(fmt.Sprintf("SAVEPOINT %s", savepointName)).Error; err != nil {
		return fmt.Errorf("failed to create savepoint %s: %w", savepointName, err)
	}

	// Execute the function
	err := fn(h.db.WithContext(ctx))

	if err != nil {
		// Rollback to savepoint on error
		if rollbackErr := h.db.WithContext(ctx).Exec(fmt.Sprintf("ROLLBACK TO SAVEPOINT %s", savepointName)).Error; rollbackErr != nil {
			return fmt.Errorf("failed to rollback to savepoint %s: %w (original error: %v)", savepointName, rollbackErr, err)
		}
		return err
	}

	// Release the savepoint on success
	if releaseErr := h.db.WithContext(ctx).Exec(fmt.Sprintf("RELEASE SAVEPOINT %s", savepointName)).Error; releaseErr != nil {
		return fmt.Errorf("failed to release savepoint %s: %w", savepointName, releaseErr)
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
	ID        string
	Success   bool
	Error     error
	Duration  time.Duration
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