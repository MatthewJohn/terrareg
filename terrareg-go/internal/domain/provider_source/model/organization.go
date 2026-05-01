package model

// Organization represents a GitHub organization or user that can be mapped to a namespace
type Organization struct {
	// Name is the organization/user name (also used as namespace name)
	Name string `json:"name"`
	// Type indicates whether this is a "user" or "organization"
	Type string `json:"type"`
	// CanPublishProviders indicates if the user can publish providers for this organization
	CanPublishProviders bool `json:"can_publish_providers"`
}

// NewOrganization creates a new organization model
func NewOrganization(name, orgType string, canPublish bool) *Organization {
	return &Organization{
		Name:                name,
		Type:                orgType,
		CanPublishProviders: canPublish,
	}
}

// GetName returns the organization name
func (o *Organization) GetName() string {
	return o.Name
}

// GetType returns the organization type
func (o *Organization) GetType() string {
	return o.Type
}

// CanPublish returns whether providers can be published for this organization
func (o *Organization) CanPublish() bool {
	return o.CanPublishProviders
}
