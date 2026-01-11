package audit

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// auditHistoryRepositoryImpl implements the audit history repository interface
type auditHistoryRepositoryImpl struct {
	db *gorm.DB
}

// NewAuditHistoryRepository creates a new audit history repository
func NewAuditHistoryRepository(db *gorm.DB) repository.AuditHistoryRepository {
	return &auditHistoryRepositoryImpl{
		db: db,
	}
}

// Create persists a new audit history entry
func (r *auditHistoryRepositoryImpl) Create(ctx context.Context, audit *model.AuditHistory) error {
	dbModel := audit.ToDBModel()

	if err := r.db.WithContext(ctx).Create(dbModel).Error; err != nil {
		return fmt.Errorf("failed to create audit history entry: %w", err)
	}

	// Set the ID back to the domain model
	audit.SetID(dbModel.ID)
	return nil
}

// Search retrieves audit history entries with pagination and filtering
func (r *auditHistoryRepositoryImpl) Search(ctx context.Context, query model.AuditHistorySearchQuery) (*model.AuditHistorySearchResult, error) {
	var dbModels []*sqldb.AuditHistoryDB
	var totalCount int64
	var filteredCount int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&sqldb.AuditHistoryDB{}).Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Build base query
	dbQuery := r.db.WithContext(ctx).Model(&sqldb.AuditHistoryDB{})

	// Apply search filter if provided
	if query.SearchValue != "" {
		searchPattern := "%" + query.SearchValue + "%"
		dbQuery = dbQuery.Where(
			"username LIKE ? OR action LIKE ? OR object_id LIKE ? OR old_value LIKE ? OR new_value LIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern, searchPattern,
		)
	}

	// Get filtered count
	if err := dbQuery.Count(&filteredCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get filtered count: %w", err)
	}

	// Apply ordering
	orderColumn := "timestamp"
	switch query.OrderColumn {
	case 0:
		orderColumn = "timestamp"
	case 1:
		orderColumn = "username"
	case 2:
		orderColumn = "action"
	case 3:
		orderColumn = "object_id"
	case 4:
		orderColumn = "old_value"
	case 5:
		orderColumn = "new_value"
	}

	orderDir := "DESC"
	if query.OrderDir == "asc" {
		orderDir = "ASC"
	}

	// Apply ordering and pagination
	if err := dbQuery.
		Order(fmt.Sprintf("%s %s", orderColumn, orderDir)).
		Offset(query.Start).
		Limit(query.Length).
		Find(&dbModels).Error; err != nil {
		return nil, fmt.Errorf("failed to search audit history: %w", err)
	}

	// Convert to domain models
	records := make([]*model.AuditHistory, len(dbModels))
	for i, dbModel := range dbModels {
		records[i] = r.dbModelToDomain(dbModel)
	}

	return &model.AuditHistorySearchResult{
		Records:       records,
		TotalCount:    int(totalCount),
		FilteredCount: int(filteredCount),
		Draw:          query.Draw,
	}, nil
}

// GetTotalCount returns the total number of audit history entries
func (r *auditHistoryRepositoryImpl) GetTotalCount(ctx context.Context) (int, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&sqldb.AuditHistoryDB{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to get total count: %w", err)
	}
	return int(count), nil
}

// GetFilteredCount returns the number of audit history entries matching the search criteria
func (r *auditHistoryRepositoryImpl) GetFilteredCount(ctx context.Context, searchValue string) (int, error) {
	dbQuery := r.db.WithContext(ctx).Model(&sqldb.AuditHistoryDB{})

	if searchValue != "" {
		searchPattern := "%" + searchValue + "%"
		dbQuery = dbQuery.Where(
			"username LIKE ? OR action LIKE ? OR object_id LIKE ? OR old_value LIKE ? OR new_value LIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern, searchPattern,
		)
	}

	var count int64
	if err := dbQuery.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to get filtered count: %w", err)
	}
	return int(count), nil
}

// dbModelToDomain converts a database model to a domain model
func (r *auditHistoryRepositoryImpl) dbModelToDomain(dbModel *sqldb.AuditHistoryDB) *model.AuditHistory {
	username := ""
	objectType := ""
	objectID := ""

	if dbModel.Username != nil {
		username = *dbModel.Username
	}
	if dbModel.ObjectType != nil {
		objectType = *dbModel.ObjectType
	}
	if dbModel.ObjectID != nil {
		objectID = *dbModel.ObjectID
	}

	// Convert database action type to domain action type
	action := model.AuditAction(dbModel.Action)

	// Create new audit history entry
	audit := model.NewAuditHistory(
		username,
		action,
		objectType,
		objectID,
		dbModel.OldValue,
		dbModel.NewValue,
	)

	// Set ID
	audit.SetID(dbModel.ID)

	// Set timestamp if available
	if dbModel.Timestamp != nil {
		audit.SetTimestamp(*dbModel.Timestamp)
	}

	return audit
}
