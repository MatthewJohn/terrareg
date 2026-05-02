package transaction

import "context"

// TransactionManager provides an abstraction for transaction management
// that can be implemented by infrastructure layer while keeping the domain
// layer independent of specific database technologies.
//
// This interface allows domain services to manage transactions without
// depending on infrastructure concerns like GORM, SQL, or specific database
// implementations.
type TransactionManager interface {
	// WithTransaction executes the given function within a transaction.
	// If a transaction already exists in the context, it will participate
	// in that transaction (supports nested transactions via savepoints).
	//
	// The function receives a context with the transaction active.
	// If the function returns an error, the transaction (or savepoint) is
	// rolled back. Otherwise, it is committed.
	//
	// This enables domain services to be tested with mock transaction
	// managers without requiring a real database connection.
	WithTransaction(ctx context.Context, fn func(context.Context) error) error

	// WithNamedTransaction is like WithTransaction but allows specifying
	// a name for the transaction/savepoint. This is useful for debugging
	// and for creating named savepoints in nested transactions.
	//
	// The name is typically used for savepoint identifiers in databases
	// that support nested transactions.
	WithNamedTransaction(ctx context.Context, name string, fn func(context.Context) error) error

	// IsTransactionActive checks if a transaction is currently active
	// in the given context. This can be used by domain services to
	// determine if they're participating in an existing transaction or
	// if a new one would be created.
	IsTransactionActive(ctx context.Context) bool
}
