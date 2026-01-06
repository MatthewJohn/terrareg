package dto

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	Limit      int  `json:"limit"`
	Offset     int  `json:"offset"`
	TotalCount int  `json:"total_count,omitempty"`
}
