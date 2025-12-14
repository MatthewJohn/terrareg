package repository

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/model"
)

// AuditHistoryRepository defines the interface for audit history persistence
type AuditHistoryRepository interface {
	// Create persists a new audit history entry
	Create(ctx context.Context, audit *model.AuditHistory) error

	// Search retrieves audit history entries with pagination and filtering
	Search(ctx context.Context, query model.AuditHistorySearchQuery) (*model.AuditHistorySearchResult, error)

	// GetTotalCount returns the total number of audit history entries
	GetTotalCount(ctx context.Context) (int, error)

	// GetFilteredCount returns the number of audit history entries matching the search criteria
	GetFilteredCount(ctx context.Context, searchValue string) (int, error)
}