package dto

// ProviderLogoResponse represents the provider logo information in API responses
type ProviderLogoResponse struct {
	Source string `json:"source"`
	Alt    string `json:"alt"`
	Tos    string `json:"tos"`
	Link   string `json:"link"`
}

// ProviderLogosResponse represents the response for the provider logos endpoint
type ProviderLogosResponse map[string]ProviderLogoResponse
