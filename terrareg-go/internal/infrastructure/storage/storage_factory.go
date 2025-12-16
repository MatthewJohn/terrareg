package storage

import (
	"fmt"
	"strings"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/storage/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/storage/service"
)

// StorageFactoryImpl implements StorageFactory interface
// This replicates the Python FileStorageFactory class
type StorageFactoryImpl struct {
	pathBuilder service.PathBuilder
}

// NewStorageFactory creates a new storage factory
func NewStorageFactory(pathBuilder service.PathBuilder) *StorageFactoryImpl {
	return &StorageFactoryImpl{
		pathBuilder: pathBuilder,
	}
}

// CreateStorageService creates a storage service based on configuration
func (f *StorageFactoryImpl) CreateStorageService(config *model.StorageConfig) (service.StorageService, error) {
	if config == nil {
		return nil, model.ErrConfigInvalid
	}

	switch config.Type {
	case model.StorageTypeLocal:
		return NewLocalStorageService(config.DataDirectory, f.pathBuilder)
	case model.StorageTypeS3:
		if config.S3Config == nil {
			return nil, fmt.Errorf("S3 configuration is required for S3 storage")
		}
		return NewS3StorageService(config.S3Config, f.pathBuilder)
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", config.Type.String())
	}
}

// GetDefaultStorageService creates storage service based on data directory format
func (f *StorageFactoryImpl) GetDefaultStorageService() (service.StorageService, error) {
	// This would read from configuration in a real implementation
	// For now, return local storage with default path
	dataDirectory := "./data" // This should come from config

	config := &model.StorageConfig{
		Type:           f.DetectStorageType(dataDirectory),
		DataDirectory:  dataDirectory,
		UploadDirectory: dataDirectory + "/upload",
	}

	return f.CreateStorageService(config)
}

// DetectStorageType detects storage type from data directory format
// This replicates the Python FileStorageFactory.get_file_storage logic
func (f *StorageFactoryImpl) DetectStorageType(dataDirectory string) model.StorageType {
	if strings.HasPrefix(dataDirectory, "s3://") {
		return model.StorageTypeS3
	}
	return model.StorageTypeLocal
}

// CreateStorageConfigFromDirectory creates storage config from data directory string
func (f *StorageFactoryImpl) CreateStorageConfigFromDirectory(dataDirectory string) *model.StorageConfig {
	storageType := f.DetectStorageType(dataDirectory)

	config := &model.StorageConfig{
		Type:          storageType,
		DataDirectory: dataDirectory,
	}

	// Set upload directory - if S3, must be local (Python behavior)
	if storageType == model.StorageTypeS3 {
		config.UploadDirectory = "./data/upload" // Default local upload directory
	} else {
		config.UploadDirectory = dataDirectory + "/upload"
	}

	// Parse S3 configuration if needed
	if storageType == model.StorageTypeS3 {
		s3Config, err := parseS3URL(dataDirectory)
		if err != nil {
			// Return error config that will be caught during storage creation
			config.S3Config = &model.S3Config{
				Bucket: "",
				Region: "",
			}
		} else {
			config.S3Config = s3Config
		}
	}

	return config
}

// parseS3URL parses S3 URL into S3Config
// This handles the S3 URL format: s3://bucket-name/path/prefix
func parseS3URL(s3URL string) (*model.S3Config, error) {
	if !strings.HasPrefix(s3URL, "s3://") {
		return nil, fmt.Errorf("invalid S3 URL format: %s", s3URL)
	}

	// Remove s3:// prefix
	url := strings.TrimPrefix(s3URL, "s3://")

	// Split bucket and key
	parts := strings.SplitN(url, "/", 2)
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid S3 URL format: %s", s3URL)
	}

	bucket := parts[0]
	keyPrefix := ""
	if len(parts) > 1 {
		keyPrefix = parts[1]
	}

	// Remove trailing slash from key prefix
	keyPrefix = strings.TrimSuffix(keyPrefix, "/")

	return &model.S3Config{
		Bucket:    bucket,
		Region:    "us-east-1", // Default region, should be configurable
		KeyPrefix: keyPrefix,
		// AccessKey and SecretKey should come from environment variables or config
	}, nil
}

// ValidateStorageConfig validates storage configuration
func (f *StorageFactoryImpl) ValidateStorageConfig(config *model.StorageConfig) error {
	if config == nil {
		return fmt.Errorf("storage config is required")
	}

	if config.DataDirectory == "" {
		return fmt.Errorf("data directory is required")
	}

	switch config.Type {
	case model.StorageTypeLocal:
		// Local storage validation
		if !strings.HasPrefix(config.DataDirectory, "/") && !strings.HasPrefix(config.DataDirectory, "./") {
			return fmt.Errorf("local data directory must be absolute or relative path")
		}
	case model.StorageTypeS3:
		// S3 validation
		if config.S3Config == nil {
			return fmt.Errorf("S3 configuration is required for S3 storage")
		}
		if config.S3Config.Bucket == "" {
			return fmt.Errorf("S3 bucket is required")
		}
		// Upload directory must be local for S3 storage (Python behavior)
		if strings.HasPrefix(config.UploadDirectory, "s3://") {
			return fmt.Errorf("upload directory must be local when using S3 storage")
		}
	default:
		return fmt.Errorf("unsupported storage type: %s", config.Type.String())
	}

	return nil
}