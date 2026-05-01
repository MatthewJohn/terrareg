package dto

// DownloadSummaryResponse represents download statistics in API responses
type DownloadSummaryResponse struct {
	Data DownloadData `json:"data"`
}

// DownloadData represents the download data structure
type DownloadData struct {
	Type       string             `json:"type"`
	ID         string             `json:"id"`
	Attributes DownloadAttributes `json:"attributes"`
}

// DownloadAttributes represents download statistics attributes
// Matches Python: AnalyticsEngine.get_module_provider_download_stats()
type DownloadAttributes struct {
	Week  int `json:"week"`
	Month int `json:"month"`
	Year  int `json:"year"`
	Total int `json:"total"`
}
