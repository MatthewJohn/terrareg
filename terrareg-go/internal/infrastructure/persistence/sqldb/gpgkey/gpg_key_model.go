package gpgkey

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// GPGKeyDBModel represents the database model for GPG keys
type GPGKeyDBModel struct {
	ID             int       `gorm:"primaryKey;autoIncrement"`
	NamespaceID    int       `gorm:"not null;index"`
	ASCIILArmor    string    `gorm:"type:longtext;not null;column:ascii_armor"`
	KeyID          string    `gorm:"not null;size:1024;index"`
	Fingerprint    string    `gorm:"not null;size:1024;uniqueIndex"`
	Source         string    `gorm:"size:1024;default:''"`
	SourceURL      *string   `gorm:"size:1024;column:source_url"`
	TrustSignature *string   `gorm:"type:text;column:trust_signature"`
	CreatedAt      time.Time `gorm:"autoCreateTime;column:created_at"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime;column:updated_at"`

	// Associations
	Namespace NamespaceDBModel `gorm:"foreignKey:NamespaceID"`
}

// TableName specifies the table name for the GPGKeyDBModel
func (GPGKeyDBModel) TableName() string {
	return "gpg_key"
}

// NamespaceDBModel represents a minimal namespace model for foreign key reference
type NamespaceDBModel struct {
	ID   int    `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"not null;size:255;uniqueIndex"`
}

// TableName specifies the table name for the NamespaceDBModel
func (NamespaceDBModel) TableName() string {
	return "namespaces"
}

// BeforeCreate hook for GPGKeyDBModel
func (g *GPGKeyDBModel) BeforeCreate(tx *gorm.DB) error {
	// Validate required fields
	if g.NamespaceID == 0 {
		return fmt.Errorf("namespace ID is required")
	}
	if g.ASCIILArmor == "" {
		return fmt.Errorf("ASCII armor is required")
	}
	if g.KeyID == "" {
		return fmt.Errorf("key ID is required")
	}
	if g.Fingerprint == "" {
		return fmt.Errorf("fingerprint is required")
	}
	return nil
}

// BeforeUpdate hook for GPGKeyDBModel
func (g *GPGKeyDBModel) BeforeUpdate(tx *gorm.DB) error {
	// Validate fields that can be updated
	if g.ASCIILArmor == "" {
		return fmt.Errorf("ASCII armor cannot be empty")
	}
	if g.KeyID == "" {
		return fmt.Errorf("key ID cannot be empty")
	}
	if g.Fingerprint == "" {
		return fmt.Errorf("fingerprint cannot be empty")
	}
	return nil
}
