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
type DownloadAttributes struct {
	Month string `json:"month,omitempty"`
	Week  string `json:"week,omitempty"`
	Year  string `json:"year,omitempty"`
	Total int    `json:"total"`
}
