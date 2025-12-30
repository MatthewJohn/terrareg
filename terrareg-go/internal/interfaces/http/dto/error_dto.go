package dto

// Error represents an API error response
type Error struct {
	Message string `json:"message"`
}

// NewError creates a new error DTO
func NewError(message string) Error {
	return Error{Message: message}
}
