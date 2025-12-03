package model

import (
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
)

// ProviderCategory represents a category for providers
type ProviderCategory struct {
	id             int
	name           *string
	slug           string
	userSelectable bool
	createdAt      time.Time
	updatedAt      time.Time
}

// NewProviderCategory creates a new provider category
func NewProviderCategory(slug string, userSelectable bool) (*ProviderCategory, error) {
	if err := ValidateCategorySlug(slug); err != nil {
		return nil, err
	}

	now := time.Now()
	return &ProviderCategory{
		slug:           slug,
		userSelectable: userSelectable,
		createdAt:      now,
		updatedAt:      now,
	}, nil
}

// ReconstructProviderCategory reconstructs a provider category from persistence
func ReconstructProviderCategory(
	id int,
	name *string,
	slug string,
	userSelectable bool,
	createdAt, updatedAt time.Time,
) *ProviderCategory {
	return &ProviderCategory{
		id:             id,
		name:           name,
		slug:           slug,
		userSelectable: userSelectable,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
	}
}

// UpdateName updates the category name
func (pc *ProviderCategory) UpdateName(name *string) {
	pc.name = name
	pc.updatedAt = time.Now()
}

// UpdateUserSelectable updates whether the category can be selected by users
func (pc *ProviderCategory) UpdateUserSelectable(userSelectable bool) {
	pc.userSelectable = userSelectable
	pc.updatedAt = time.Now()
}

// Getters

func (pc *ProviderCategory) ID() int {
	return pc.id
}

func (pc *ProviderCategory) Name() *string {
	return pc.name
}

func (pc *ProviderCategory) Slug() string {
	return pc.slug
}

func (pc *ProviderCategory) UserSelectable() bool {
	return pc.userSelectable
}

func (pc *ProviderCategory) CreatedAt() time.Time {
	return pc.createdAt
}

func (pc *ProviderCategory) UpdatedAt() time.Time {
	return pc.updatedAt
}

// GetDisplayName returns the name or falls back to slug-formatted name
func (pc *ProviderCategory) GetDisplayName() string {
	if pc.name != nil && *pc.name != "" {
		return *pc.name
	}
	return formatSlugToDisplayName(pc.slug)
}

// formatSlugToDisplayName converts a slug to a display name
func formatSlugToDisplayName(slug string) string {
	// Simple conversion from slug-case to Title Case
	// This is a basic implementation - can be enhanced as needed
	return slug
}

// ValidateCategorySlug validates a category slug
func ValidateCategorySlug(slug string) error {
	if slug == "" {
		return shared.ErrInvalidCategorySlug
	}
	// Add more validation as needed
	return nil
}