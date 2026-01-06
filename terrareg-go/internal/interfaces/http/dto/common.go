package dto

// PaginationMeta represents pagination metadata (matching Python ResultData.meta)
type PaginationMeta struct {
	Limit         int  `json:"limit"`
	CurrentOffset int  `json:"current_offset"`
	PrevOffset    *int `json:"prev_offset,omitempty"`
	NextOffset    *int `json:"next_offset,omitempty"`
}
