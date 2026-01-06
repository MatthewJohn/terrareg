package terrareg

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	providerQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/provider"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/module"
	httputils "github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/utils"
)

type SearchFiltersHandler struct {
	searchFiltersQuery         *module.SearchFiltersQuery
	providerSearchFiltersQuery *providerQuery.SearchFiltersQuery
}

func NewSearchFiltersHandler(
	searchFiltersQuery *module.SearchFiltersQuery,
	providerSearchFiltersQuery *providerQuery.SearchFiltersQuery,
) *SearchFiltersHandler {
	return &SearchFiltersHandler{
		searchFiltersQuery:         searchFiltersQuery,
		providerSearchFiltersQuery: providerSearchFiltersQuery,
	}
}

func (h *SearchFiltersHandler) RegisterRoutes(r chi.Router) {
	r.Get("/modules/search/filters", h.HandleModuleSearchFilters)
	r.Get("/search_filters", h.HandleModuleSearchFiltersLegacy)
}

func (h *SearchFiltersHandler) HandleModuleSearchFilters(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")

	counts, err := h.searchFiltersQuery.Execute(r.Context(), query)
	if err != nil {
		httputils.SendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	httputils.SendJSONResponse(w, http.StatusOK, counts)
}

func (h *SearchFiltersHandler) HandleModuleSearchFiltersLegacy(w http.ResponseWriter, r *http.Request) {
	h.HandleModuleSearchFilters(w, r)
}

func (h *SearchFiltersHandler) HandleProviderSearchFilters(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")

	counts, err := h.providerSearchFiltersQuery.Execute(r.Context(), query)
	if err != nil {
		httputils.SendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	httputils.SendJSONResponse(w, http.StatusOK, counts)
}
