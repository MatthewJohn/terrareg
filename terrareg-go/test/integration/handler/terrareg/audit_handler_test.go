package terrareg_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	auditQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/audit"
	auditService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/service"
	auditRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/audit"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terrareg"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestAuditHandler_HandleAuditHistoryGet_Success tests successful audit history retrieval
func TestAuditHandler_HandleAuditHistoryGet_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create audit log entries
	testutils.CreateMultipleAuditLogs(t, db, 5, "test-user", "CREATE", "namespace")

	// Create handler
	auditRepository := auditRepo.NewAuditHistoryRepository(db.DB)
	service := auditService.NewAuditService(auditRepository)
	getAuditHistoryQuery := auditQuery.NewGetAuditHistoryQuery(service)
	handler := terrareg.NewAuditHandler(getAuditHistoryQuery)

	// Create request with DataTables parameters
	params := url.Values{}
	params.Add("draw", "1")
	params.Add("start", "0")
	params.Add("length", "10")
	params.Add("search[value]", "")

	req := httptest.NewRequest("GET", "/v1/terrareg/audit-history?"+params.Encode(), nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleAuditHistoryGet(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)

	// Verify DataTables format
	assert.Contains(t, response, "data")
	assert.Contains(t, response, "draw")
	assert.Contains(t, response, "recordsTotal")
	assert.Contains(t, response, "recordsFiltered")

	// Verify values
	assert.Equal(t, float64(1), response["draw"])
	assert.Equal(t, float64(5), response["recordsTotal"])
	assert.Equal(t, float64(5), response["recordsFiltered"])

	// Verify data is an array
	data := response["data"].([]interface{})
	assert.Len(t, data, 5)
}

// TestAuditHandler_HandleAuditHistoryGet_Empty tests audit history with no data
func TestAuditHandler_HandleAuditHistoryGet_Empty(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler with no data
	auditRepository := auditRepo.NewAuditHistoryRepository(db.DB)
	service := auditService.NewAuditService(auditRepository)
	getAuditHistoryQuery := auditQuery.NewGetAuditHistoryQuery(service)
	handler := terrareg.NewAuditHandler(getAuditHistoryQuery)

	params := url.Values{}
	params.Add("draw", "1")
	params.Add("start", "0")
	params.Add("length", "10")

	req := httptest.NewRequest("GET", "/v1/terrareg/audit-history?"+params.Encode(), nil)
	w := httptest.NewRecorder()

	handler.HandleAuditHistoryGet(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)

	assert.Equal(t, float64(0), response["recordsTotal"])
	assert.Equal(t, float64(0), response["recordsFiltered"])

	data := response["data"].([]interface{})
	assert.Len(t, data, 0)
}

// TestAuditHandler_HandleAuditHistoryGet_Pagination tests pagination functionality
func TestAuditHandler_HandleAuditHistoryGet_Pagination(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create 25 audit log entries
	testutils.CreateMultipleAuditLogs(t, db, 25, "test-user", "CREATE", "namespace")

	auditRepository := auditRepo.NewAuditHistoryRepository(db.DB)
	service := auditService.NewAuditService(auditRepository)
	getAuditHistoryQuery := auditQuery.NewGetAuditHistoryQuery(service)
	handler := terrareg.NewAuditHandler(getAuditHistoryQuery)

	// Request first page (10 records)
	params := url.Values{}
	params.Add("draw", "1")
	params.Add("start", "0")
	params.Add("length", "10")

	req := httptest.NewRequest("GET", "/v1/terrareg/audit-history?"+params.Encode(), nil)
	w := httptest.NewRecorder()

	handler.HandleAuditHistoryGet(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)

	assert.Equal(t, float64(25), response["recordsTotal"])
	assert.Equal(t, float64(25), response["recordsFiltered"])

	data := response["data"].([]interface{})
	assert.Len(t, data, 10)

	// Request second page
	params.Set("start", "10")
	params.Set("draw", "2")

	req = httptest.NewRequest("GET", "/v1/terrareg/audit-history?"+params.Encode(), nil)
	w = httptest.NewRecorder()

	handler.HandleAuditHistoryGet(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response = testutils.GetJSONBody(t, w)
	assert.Equal(t, float64(2), response["draw"])

	data = response["data"].([]interface{})
	assert.Len(t, data, 10)
}

// TestAuditHandler_HandleAuditHistoryGet_Search tests search functionality
func TestAuditHandler_HandleAuditHistoryGet_Search(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create audit logs with different usernames
	testutils.CreateAuditLog(t, db, "alice", "CREATE", "namespace", 1)
	testutils.CreateAuditLog(t, db, "bob", "CREATE", "namespace", 2)
	testutils.CreateAuditLog(t, db, "charlie", "CREATE", "namespace", 3)

	auditRepository := auditRepo.NewAuditHistoryRepository(db.DB)
	service := auditService.NewAuditService(auditRepository)
	getAuditHistoryQuery := auditQuery.NewGetAuditHistoryQuery(service)
	handler := terrareg.NewAuditHandler(getAuditHistoryQuery)

	// Search for "alice"
	params := url.Values{}
	params.Add("draw", "1")
	params.Add("start", "0")
	params.Add("length", "10")
	params.Add("search[value]", "alice")

	req := httptest.NewRequest("GET", "/v1/terrareg/audit-history?"+params.Encode(), nil)
	w := httptest.NewRecorder()

	handler.HandleAuditHistoryGet(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)

	// Should have filtered results
	assert.Equal(t, float64(3), response["recordsTotal"])
	assert.Equal(t, float64(1), response["recordsFiltered"])

	data := response["data"].([]interface{})
	assert.Len(t, data, 1)
}

// TestAuditHandler_HandleAuditHistoryGet_Sorting tests sorting functionality
func TestAuditHandler_HandleAuditHistoryGet_Sorting(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create audit logs - they'll have different timestamps
	testutils.CreateAuditLog(t, db, "user1", "CREATE", "namespace", 1)
	testutils.CreateAuditLog(t, db, "user2", "UPDATE", "namespace", 2)
	testutils.CreateAuditLog(t, db, "user3", "DELETE", "namespace", 3)

	auditRepository := auditRepo.NewAuditHistoryRepository(db.DB)
	service := auditService.NewAuditService(auditRepository)
	getAuditHistoryQuery := auditQuery.NewGetAuditHistoryQuery(service)
	handler := terrareg.NewAuditHandler(getAuditHistoryQuery)

	// Request with ascending order
	params := url.Values{}
	params.Add("draw", "1")
	params.Add("start", "0")
	params.Add("length", "10")
	params.Add("order[0][dir]", "asc")
	params.Add("order[0][column]", "0")

	req := httptest.NewRequest("GET", "/v1/terrareg/audit-history?"+params.Encode(), nil)
	w := httptest.NewRecorder()

	handler.HandleAuditHistoryGet(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)

	data := response["data"].([]interface{})
	assert.Len(t, data, 3)
	// First row should have user1 (earliest timestamp)
	firstRow := data[0].([]interface{})
	assert.Contains(t, firstRow[1], "user1")
}

// TestAuditHandler_HandleAuditHistoryGet_DefaultParameters tests default parameter handling
func TestAuditHandler_HandleAuditHistoryGet_DefaultParameters(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create some audit logs
	testutils.CreateMultipleAuditLogs(t, db, 5, "test-user", "CREATE", "namespace")

	auditRepository := auditRepo.NewAuditHistoryRepository(db.DB)
	service := auditService.NewAuditService(auditRepository)
	getAuditHistoryQuery := auditQuery.NewGetAuditHistoryQuery(service)
	handler := terrareg.NewAuditHandler(getAuditHistoryQuery)

	// Request without length and start parameters (should use defaults)
	params := url.Values{}
	// Intentionally not setting length, start, or draw

	req := httptest.NewRequest("GET", "/v1/terrareg/audit-history?"+params.Encode(), nil)
	w := httptest.NewRecorder()

	handler.HandleAuditHistoryGet(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)

	// Default length should be 10
	data := response["data"].([]interface{})
	assert.Len(t, data, 5) // All 5 records should be returned
}

// TestAuditHandler_HandleAuditHistoryGet_DataFormat tests the response data format
func TestAuditHandler_HandleAuditHistoryGet_DataFormat(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create an audit log
	testutils.CreateAuditLog(t, db, "test-user", "CREATE", "namespace", 123)

	auditRepository := auditRepo.NewAuditHistoryRepository(db.DB)
	service := auditService.NewAuditService(auditRepository)
	getAuditHistoryQuery := auditQuery.NewGetAuditHistoryQuery(service)
	handler := terrareg.NewAuditHandler(getAuditHistoryQuery)

	params := url.Values{}
	params.Add("draw", "1")
	params.Add("start", "0")
	params.Add("length", "10")

	req := httptest.NewRequest("GET", "/v1/terrareg/audit-history?"+params.Encode(), nil)
	w := httptest.NewRecorder()

	handler.HandleAuditHistoryGet(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)

	data := response["data"].([]interface{})
	assert.Len(t, data, 1)

	// Verify row format: [timestamp, username, action, object_id, old_value, new_value]
	row := data[0].([]interface{})
	assert.Len(t, row, 6)
	assert.NotEmpty(t, row[0]) // timestamp
	assert.Equal(t, "test-user", row[1])
	assert.Equal(t, "CREATE", row[2])
	assert.Equal(t, "123", row[3])
}
