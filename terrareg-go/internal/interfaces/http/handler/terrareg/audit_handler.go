package terrareg

import (
	"net/http"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/audit"
)

// AuditHandler handles audit-related requests
type AuditHandler struct {
	getAuditHistoryQuery *audit.GetAuditHistoryQuery
}

// NewAuditHandler creates a new audit handler
func NewAuditHandler(
	getAuditHistoryQuery *audit.GetAuditHistoryQuery,
) *AuditHandler {
	return &AuditHandler{
		getAuditHistoryQuery: getAuditHistoryQuery,
	}
}

// HandleAuditHistoryGet handles GET /v1/terrareg/audit-history
func (h *AuditHandler) HandleAuditHistoryGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	searchValue := r.URL.Query().Get("search[value]")
	length := r.URL.Query().Get("length")
	start := r.URL.Query().Get("start")
	draw := r.URL.Query().Get("draw")
	orderDir := r.URL.Query().Get("order[0][dir]")
	orderColumn := r.URL.Query().Get("order[0][column]")

	// Parse parameters
	req, err := audit.ParseQueryParams(searchValue, length, start, draw, orderDir, orderColumn)
	if err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Execute query
	response, err := h.getAuditHistoryQuery.Execute(ctx, req)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Send response
	RespondJSON(w, http.StatusOK, response)
}