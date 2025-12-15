package provider

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// ProviderModel represents the database model for providers
type ProviderModel struct {
	ID                    int       `gorm:"primaryKey;autoIncrement"`
	NamespaceID           int       `gorm:"not null;index"`
	Name                  string    `gorm:"not null;size:255;index:idx_provider_namespace_name"`
	Description           *string   `gorm:"type:text"`
	Tier                  string    `gorm:"size:50"`
	CategoryID            *int      `gorm:"index"`
	RepositoryID          *int      `gorm:"index"`
	LatestVersionID       *int      `gorm:"index"`
	UseProviderSourceAuth bool      `gorm:"default:false"`
	CreatedAt             time.Time `gorm:"autoCreateTime"`
	UpdatedAt             time.Time `gorm:"autoUpdateTime"`

	// Associations
	Namespace     NamespaceModel         `gorm:"foreignKey:NamespaceID"`
	Category      *ProviderCategoryModel `gorm:"foreignKey:CategoryID"`
	Repository    *RepositoryModel       `gorm:"foreignKey:RepositoryID"`
	LatestVersion *ProviderVersionModel  `gorm:"foreignKey:LatestVersionID"`

	// Collections (loaded with eager loading or separately)
	Versions []ProviderVersionModel `gorm:"foreignKey:ProviderID"`
	GPGKeys  []GPGKeyModel          `gorm:"foreignKey:ProviderID"`
}

// TableName specifies the table name for the ProviderModel
func (ProviderModel) TableName() string {
	return "providers"
}

// ProviderVersionModel represents the database model for provider versions
type ProviderVersionModel struct {
	ID               int     `gorm:"primaryKey;autoIncrement"`
	ProviderID       int     `gorm:"not null;index"`
	Version          string  `gorm:"not null;size:50;index:idx_version_provider"`
	GitTag           *string `gorm:"size:255"`
	Beta             bool    `gorm:"default:false"`
	PublishedAt      *time.Time
	GPGKeyID         *int      `gorm:"index"`
	ProtocolVersions string    `gorm:"type:text"` // JSON array of protocol versions
	CreatedAt        time.Time `gorm:"autoCreateTime"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime"`

	// Associations
	Provider ProviderModel `gorm:"foreignKey:ProviderID"`
	GPGKey   *GPGKeyModel  `gorm:"foreignKey:GPGKeyID"`

	// Collections
	Binaries []ProviderBinaryModel `gorm:"foreignKey:VersionID"`
}

// TableName specifies the table name for the ProviderVersionModel
func (ProviderVersionModel) TableName() string {
	return "provider_versions"
}

// ProviderBinaryModel represents the database model for provider binaries
type ProviderBinaryModel struct {
	ID              int       `gorm:"primaryKey;autoIncrement"`
	VersionID       int       `gorm:"not null;index"`
	OperatingSystem string    `gorm:"not null;size:50"`
	Architecture    string    `gorm:"not null;size:50"`
	FileName        string    `gorm:"not null;size:255"`
	FileSize        int64     `gorm:"not null"`
	FileHash        string    `gorm:"not null;size:64;index"`
	DownloadURL     string    `gorm:"not null;size:500"`
	CreatedAt       time.Time `gorm:"autoCreateTime"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime"`

	// Associations
	Version ProviderVersionModel `gorm:"foreignKey:VersionID"`
}

// TableName specifies the table name for the ProviderBinaryModel
func (ProviderBinaryModel) TableName() string {
	return "provider_binaries"
}

// GPGKeyModel represents the database model for GPG keys
type GPGKeyModel struct {
	ID             int       `gorm:"primaryKey;autoIncrement"`
	ProviderID     int       `gorm:"not null;index"`
	KeyText        string    `gorm:"type:longtext"`
	AsciiArmor     string    `gorm:"type:longtext"`
	KeyID          string    `gorm:"not null;size:64;index"`
	TrustSignature *string   `gorm:"type:text"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime"`

	// Associations
	Provider ProviderModel `gorm:"foreignKey:ProviderID"`
}

// TableName specifies the table name for the GPGKeyModel
func (GPGKeyModel) TableName() string {
	return "provider_gpg_keys"
}

// ProviderCategoryModel represents the database model for provider categories
type ProviderCategoryModel struct {
	ID             int       `gorm:"primaryKey;autoIncrement"`
	Name           *string   `gorm:"size:255"`
	Slug           string    `gorm:"not null;size:100;uniqueIndex"`
	UserSelectable bool      `gorm:"default:true"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime"`

	// Collections
	Providers []ProviderModel `gorm:"foreignKey:CategoryID"`
}

// TableName specifies the table name for the ProviderCategoryModel
func (ProviderCategoryModel) TableName() string {
	return "provider_categories"
}

// NamespaceModel represents a minimal namespace model for foreign key reference
type NamespaceModel struct {
	ID   int    `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"not null;size:255;uniqueIndex"`
}

// TableName specifies the table name for the NamespaceModel
func (NamespaceModel) TableName() string {
	return "namespaces"
}

// RepositoryModel represents a minimal repository model for foreign key reference
type RepositoryModel struct {
	ID         int    `gorm:"primaryKey;autoIncrement"`
	Name       string `gorm:"not null;size:255"`
	CloneURL   string `gorm:"not null;size:500"`
	ModulePath string `gorm:"size:255"`
}

// TableName specifies the table name for the RepositoryModel
func (RepositoryModel) TableName() string {
	return "repositories"
}

// BeforeCreate hook for ProviderModel
func (p *ProviderModel) BeforeCreate(tx *gorm.DB) error {
	// Set default tier if not provided
	if p.Tier == "" {
		p.Tier = "community"
	}
	return nil
}

// BeforeCreate hook for ProviderVersionModel
func (pv *ProviderVersionModel) BeforeCreate(tx *gorm.DB) error {
	// Set default protocol versions if not provided
	if pv.ProtocolVersions == "" {
		pv.ProtocolVersions = `["5.0"]`
	}
	return nil
}

// BeforeCreate hook for ProviderBinaryModel
func (pb *ProviderBinaryModel) BeforeCreate(tx *gorm.DB) error {
	// Validate required fields
	if pb.VersionID == 0 {
		return fmt.Errorf("version ID is required")
	}
	if pb.OperatingSystem == "" {
		return fmt.Errorf("operating system is required")
	}
	if pb.Architecture == "" {
		return fmt.Errorf("architecture is required")
	}
	if pb.FileName == "" {
		return fmt.Errorf("filename is required")
	}
	if pb.DownloadURL == "" {
		return fmt.Errorf("download URL is required")
	}
	return nil
}

// BeforeCreate hook for GPGKeyModel
func (gk *GPGKeyModel) BeforeCreate(tx *gorm.DB) error {
	// Validate required fields
	if gk.ProviderID == 0 {
		return fmt.Errorf("provider ID is required")
	}
	if gk.KeyID == "" {
		return fmt.Errorf("key ID is required")
	}
	return nil
}

// BeforeCreate hook for ProviderCategoryModel
func (pc *ProviderCategoryModel) BeforeCreate(tx *gorm.DB) error {
	// Validate required fields
	if pc.Slug == "" {
		return fmt.Errorf("slug is required")
	}
	return nil
}

// AfterFind hook for ProviderVersionModel to deserialize protocol versions
func (pv *ProviderVersionModel) AfterFind(tx *gorm.DB) error {
	// This would be handled by a custom type or GORM custom scanner
	// For now, the protocol versions remain as JSON string
	return nil
}

// BeforeSave hook for ProviderVersionModel to serialize protocol versions
func (pv *ProviderVersionModel) BeforeSave(tx *gorm.DB) error {
	// This would be handled by a custom type or GORM custom marshaller
	// For now, the protocol versions remain as JSON string
	return nil
}
