package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/model"
)

// MockAuditHistoryRepository is a mock for AuditHistoryRepository interface
type MockAuditHistoryRepository struct {
	mock.Mock
}

// Create persists a new audit history entry
func (m *MockAuditHistoryRepository) Create(ctx context.Context, audit *model.AuditHistory) error {
	args := m.Called(ctx, audit)
	return args.Error(0)
}

// Search retrieves audit history entries with pagination and filtering
func (m *MockAuditHistoryRepository) Search(ctx context.Context, query model.AuditHistorySearchQuery) (*model.AuditHistorySearchResult, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AuditHistorySearchResult), args.Error(1)
}

// GetTotalCount returns the total number of audit history entries
func (m *MockAuditHistoryRepository) GetTotalCount(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

// GetFilteredCount returns the number of audit history entries matching the search criteria
func (m *MockAuditHistoryRepository) GetFilteredCount(ctx context.Context, searchValue string) (int, error) {
	args := m.Called(ctx, searchValue)
	return args.Int(0), args.Error(1)
}
