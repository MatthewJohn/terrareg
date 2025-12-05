package dto

// IsAuthenticatedResponse represents the response from the is_authenticated endpoint
type IsAuthenticatedResponse struct {
	Authenticated        bool            `json:"authenticated"`
	ReadAccess          bool            `json:"read_access"`
	SiteAdmin           bool            `json:"site_admin"`
	NamespacePermissions map[string]string `json:"namespace_permissions"`
}