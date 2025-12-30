package model

// ProviderLogoInfo contains the static logo information for a provider
type ProviderLogoInfo struct {
	Source string `json:"source"`
	Alt    string `json:"alt"`
	Tos    string `json:"tos"`
	Link   string `json:"link"`
}
