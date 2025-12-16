package storage

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/storage/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/storage/service"
)

// DefaultDataDir is used when no data directory is configured
const DefaultDataDir = "./data"

// StorageFactoryImpl implements StorageFactory interface
// Replicates the Python FileStorageFactory logic
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
func (f *StorageFactoryImpl) CreateStorageService(
	config *model.StorageConfig,
) (service.StorageService, error) {
	if config == nil {
		return nil, model.ErrConfigInvalid
	}

	switch config.Type {
	case model.StorageTypeLocal:
		return NewLocalStorageService(config.DataDirectory, f.pathBuilder)
	case model.StorageTypeS3:
		if config.S3Config == nil || config.S3Config.Bucket == "" {
			return nil, fmt.Errorf("S3 configuration with bucket is required for S3 storage")
		}
		return NewS3StorageService(config.S3Config, f.pathBuilder)
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", config.Type.String())
	}
}

// GetDefaultStorageService creates a default storage service
func (f *StorageFactoryImpl) GetDefaultStorageService() (service.StorageService, error) {
	dataDirectory := DefaultDataDir

	config := &model.StorageConfig{
		Type:            f.DetectStorageType(dataDirectory),
		DataDirectory:   dataDirectory,
		UploadDirectory: path.Join(dataDirectory, "upload"),
	}

	return f.CreateStorageService(config)
}

// DetectStorageType detects storage type based on the directory string
func (f *StorageFactoryImpl) DetectStorageType(dataDirectory string) model.StorageType {
	dataDirectory = strings.TrimSpace(dataDirectory)
	if strings.HasPrefix(dataDirectory, "s3://") {
		return model.StorageTypeS3
	}
	return model.StorageTypeLocal
}

// CreateStorageConfigFromDirectory creates a storage config from a directory string
func (f *StorageFactoryImpl) CreateStorageConfigFromDirectory(
	dataDirectory string,
) (*model.StorageConfig, error) {
	storageType := f.DetectStorageType(dataDirectory)

	config := &model.StorageConfig{
		Type:          storageType,
		DataDirectory: dataDirectory,
	}

	if storageType == model.StorageTypeS3 {
		s3Config, err := parseS3URL(dataDirectory)
		if err != nil {
			return nil, fmt.Errorf("failed to parse S3 URL: %w", err)
		}
		config.S3Config = s3Config
		// Upload directory is local for S3 storage
		config.UploadDirectory = path.Join(DefaultDataDir, "upload")
	} else {
		// Local storage
		config.UploadDirectory = path.Join(dataDirectory, "upload")
	}

	return config, nil
}

// parseS3URL parses S3 URL into S3Config
// Format: s3://bucket-name/path/prefix
func parseS3URL(s3URL string) (*model.S3Config, error) {
	if !strings.HasPrefix(s3URL, "s3://") {
		return nil, fmt.Errorf("invalid S3 URL format: %s", s3URL)
	}

	url := strings.TrimPrefix(s3URL, "s3://")
	parts := strings.SplitN(url, "/", 2)
	bucket := parts[0]

	keyPrefix := ""
	if len(parts) == 2 {
		keyPrefix = strings.TrimSuffix(parts[1], "/")
	}

	return &model.S3Config{
		Bucket:    bucket,
		KeyPrefix: keyPrefix,
		// AWS SDK will pick up region and credentials from environment
	}, nil
}

// ValidateStorageConfig validates a storage configuration
func (f *StorageFactoryImpl) ValidateStorageConfig(config *model.StorageConfig) error {
	if config == nil {
		return fmt.Errorf("storage config is required")
	}

	if config.DataDirectory == "" {
		return fmt.Errorf("data directory is required")
	}

	switch config.Type {
	case model.StorageTypeLocal:
		if !strings.HasPrefix(config.DataDirectory, "/") && !strings.HasPrefix(config.DataDirectory, "./") {
			return fmt.Errorf("local data directory must be absolute or relative path")
		}
		info, err := os.Stat(config.UploadDirectory)
		if os.IsNotExist(err) {
			return fmt.Errorf("upload directory does not exist: %s", config.UploadDirectory)
		}
		if err != nil || !info.IsDir() {
			return fmt.Errorf("upload directory is invalid: %s", config.UploadDirectory)
		}
	case model.StorageTypeS3:
		if config.S3Config == nil || config.S3Config.Bucket == "" {
			return fmt.Errorf("S3 configuration with bucket is required for S3 storage")
		}
		if strings.HasPrefix(config.UploadDirectory, "s3://") {
			return fmt.Errorf("upload directory must be local when using S3 storage")
		}
	default:
		return fmt.Errorf("unsupported storage type: %s", config.Type.String())
	}

	return nil
}
