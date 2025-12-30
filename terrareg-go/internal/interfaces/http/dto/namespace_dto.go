package dto

// NamespaceResponse represents a namespace in API responses
type NamespaceResponse struct {
	Name        string  `json:"name"`
	DisplayName *string `json:"display_name,omitempty"`
	Type        string  `json:"type"`
}

// NamespaceListResponse represents a list of namespaces
type NamespaceListResponse struct {
	Namespaces []NamespaceResponse `json:"namespaces"`
}

// NamespaceCreateRequest represents a request to create a namespace
type NamespaceCreateRequest struct {
	Name        string  `json:"name" binding:"required"`
	DisplayName *string `json:"display_name"`
	Type        string  `json:"type"`
}

// NamespaceUpdateRequest represents a request to update a namespace
type NamespaceUpdateRequest struct {
	Name        *string `json:"name,omitempty"`
	DisplayName *string `json:"display_name,omitempty"`
	CsrfToken   string  `json:"csrf_token"`
}
