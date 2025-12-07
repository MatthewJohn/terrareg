package terrareg

import (
	"net/http"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/provider_logo"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto"
	httputils "github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/utils"
)

// ProviderLogosHandler handles provider logos requests
type ProviderLogosHandler struct {
	getAllProviderLogosQuery *provider_logo.GetAllProviderLogosQuery
}

// NewProviderLogosHandler creates a new ProviderLogosHandler
func NewProviderLogosHandler(getAllProviderLogosQuery *provider_logo.GetAllProviderLogosQuery) *ProviderLogosHandler {
	return &ProviderLogosHandler{
		getAllProviderLogosQuery: getAllProviderLogosQuery,
	}
}

// HandleGetProviderLogos handles GET /v1/terrareg/provider_logos
func (h *ProviderLogosHandler) HandleGetProviderLogos(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get all provider logos
	providerLogos := h.getAllProviderLogosQuery.Execute(ctx)

	// Convert to DTO response
	response := make(dto.ProviderLogosResponse)
	for providerName, logo := range providerLogos {
		if logo.Exists() {
			response[providerName] = dto.ProviderLogoResponse{
				Source: logo.Source(),
				Alt:    logo.Alt(),
				Tos:    logo.Tos(),
				Link:   logo.Link(),
			}
		}
	}

	httputils.SendJSONResponse(w, http.StatusOK, response)
}