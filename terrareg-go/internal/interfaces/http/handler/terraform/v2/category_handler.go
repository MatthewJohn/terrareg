package v2

import (
	"fmt"
	"net/http"

	providerQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/provider"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terrareg"
)

// TerraformV2CategoryHandler groups all /v2/categories handlers
type TerraformV2CategoryHandler struct {
	listCategoriesQuery *providerQuery.ListUserSelectableProviderCategoriesQuery
}

// NewTerraformV2CategoryHandler creates a new TerraformV2CategoryHandler
func NewTerraformV2CategoryHandler(
	listCategoriesQuery *providerQuery.ListUserSelectableProviderCategoriesQuery,
) *TerraformV2CategoryHandler {
	return &TerraformV2CategoryHandler{
		listCategoriesQuery: listCategoriesQuery,
	}
}

// HandleListCategories handles GET /v2/categories
func (h *TerraformV2CategoryHandler) HandleListCategories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get all categories from domain (including non-user-selectable for API response)
	categories, err := h.listCategoriesQuery.Execute(ctx)
	if err != nil {
		terrareg.RespondError(w, http.StatusInternalServerError, "Failed to get categories")
		return
	}

	// Convert domain models to JSON:API format response matching Python implementation
	data := make([]map[string]interface{}, len(categories))
	for i, category := range categories {
		data[i] = map[string]interface{}{
			"type": "categories",
			"id":   fmt.Sprintf("%d", category.ID()),
			"attributes": map[string]interface{}{
				"name":           category.GetDisplayName(),
				"slug":           category.Slug(),
				"user-selectable": category.UserSelectable(),
			},
			"links": map[string]string{
				"self": fmt.Sprintf("/v2/categories/%d", category.ID()),
			},
		}
	}

	response := map[string]interface{}{
		"data": data,
	}

	terrareg.RespondJSON(w, http.StatusOK, response)
}