package service

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/repository"
)

// AuditService handles audit-related business logic
type AuditService struct {
	auditRepo repository.AuditHistoryRepository
}

// NewAuditService creates a new audit service
func NewAuditService(auditRepo repository.AuditHistoryRepository) *AuditService {
	return &AuditService{
		auditRepo: auditRepo,
	}
}

// LogEvent logs an audit event
func (s *AuditService) LogEvent(ctx context.Context, audit *model.AuditHistory) error {
	return s.auditRepo.Create(ctx, audit)
}

// SearchHistory retrieves audit history with pagination and search
func (s *AuditService) SearchHistory(ctx context.Context, query model.AuditHistorySearchQuery) (*model.AuditHistorySearchResult, error) {
	return s.auditRepo.Search(ctx, query)
}

// GetTotalCount returns the total number of audit entries
func (s *AuditService) GetTotalCount(ctx context.Context) (int, error) {
	return s.auditRepo.GetTotalCount(ctx)
}

// GetFilteredCount returns the number of entries matching the search criteria
func (s *AuditService) GetFilteredCount(ctx context.Context, searchValue string) (int, error) {
	return s.auditRepo.GetFilteredCount(ctx, searchValue)
}