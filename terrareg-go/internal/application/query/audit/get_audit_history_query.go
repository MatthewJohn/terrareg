package audit

import (
	"context"
	"strconv"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/service"
)

// GetAuditHistoryRequest represents a request to get audit history
type GetAuditHistoryRequest struct {
	SearchValue string
	Length      int
	Start       int
	Draw        int
	OrderDir    string
	OrderColumn int
}

// GetAuditHistoryResponse represents the response from get audit history query
type GetAuditHistoryResponse struct {
	Data           [][]interface{} `json:"data"`
	Draw           int              `json:"draw"`
	RecordsTotal   int              `json:"recordsTotal"`
	RecordsFiltered int             `json:"recordsFiltered"`
}

// GetAuditHistoryQuery retrieves audit history with pagination
type GetAuditHistoryQuery struct {
	auditService *service.AuditService
}

// NewGetAuditHistoryQuery creates a new GetAuditHistoryQuery
func NewGetAuditHistoryQuery(auditService *service.AuditService) *GetAuditHistoryQuery {
	return &GetAuditHistoryQuery{
		auditService: auditService,
	}
}

// Execute retrieves audit history
func (q *GetAuditHistoryQuery) Execute(ctx context.Context, req GetAuditHistoryRequest) (*GetAuditHistoryResponse, error) {
	// Build search query
	searchQuery := model.AuditHistorySearchQuery{
		SearchValue: req.SearchValue,
		Length:      req.Length,
		Start:       req.Start,
		Draw:        req.Draw,
		OrderDir:    req.OrderDir,
		OrderColumn: req.OrderColumn,
	}

	// Get search results
	result, err := q.auditService.SearchHistory(ctx, searchQuery)
	if err != nil {
		return nil, err
	}

	// Convert to DataTables format
	data := make([][]interface{}, len(result.Records))
	for i, record := range result.Records {
		data[i] = []interface{}{
			record.Timestamp().Format("2006-01-02 15:04:05"),
			record.Username(),
			string(record.Action()),
			record.ObjectID(),
			record.OldValue(),
			record.NewValue(),
		}
	}

	return &GetAuditHistoryResponse{
		Data:            data,
		Draw:            req.Draw,
		RecordsTotal:    result.TotalCount,
		RecordsFiltered: result.FilteredCount,
	}, nil
}

// ParseQueryParams parses HTTP query parameters into GetAuditHistoryRequest
func ParseQueryParams(searchValue, length, start, draw, orderDir, orderColumn string) (GetAuditHistoryRequest, error) {
	req := GetAuditHistoryRequest{
		SearchValue: searchValue,
		OrderDir:    orderDir,
	}

	if length != "" {
		if l, err := strconv.Atoi(length); err == nil {
			req.Length = l
		}
	}
	if req.Length == 0 {
		req.Length = 10 // Default page size
	}

	if start != "" {
		if s, err := strconv.Atoi(start); err == nil {
			req.Start = s
		}
	}

	if draw != "" {
		if d, err := strconv.Atoi(draw); err == nil {
			req.Draw = d
		}
	}

	if orderColumn != "" {
		if c, err := strconv.Atoi(orderColumn); err == nil {
			req.OrderColumn = c
		}
	}

	return req, nil
}