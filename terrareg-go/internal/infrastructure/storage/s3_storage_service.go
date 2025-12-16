package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/storage/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/storage/service"
)

// S3StorageService implements StorageService for S3
// This provides real S3 functionality using AWS SDK v2
type S3StorageService struct {
	s3Config    *model.S3Config
	pathBuilder service.PathBuilder
	s3Client    *s3.Client
	awsConfig   *aws.Config
}

// NewS3StorageService creates a new S3 storage service
func NewS3StorageService(s3Config *model.S3Config, pathBuilder service.PathBuilder) (*S3StorageService, error) {
	if s3Config.Bucket == "" {
		return nil, fmt.Errorf("S3 bucket is required")
	}

	// Load AWS configuration
	awsConfig, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(s3Config.Region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client
	s3Client := s3.NewFromConfig(awsConfig)

	return &S3StorageService{
		s3Config:    s3Config,
		pathBuilder: pathBuilder,
		s3Client:    s3Client,
		awsConfig:   &awsConfig,
	}, nil
}

// generateKey generates S3 key from path
// This replicates Python's _generate_key method
func (s *S3StorageService) generateKey(path string) string {
	key := s.s3Config.KeyPrefix
	if key != "" && !strings.HasSuffix(key, "/") {
		key += "/"
	}
	if strings.HasPrefix(path, "/") {
		key += path[1:]
	} else {
		key += path
	}
	return key
}

// UploadFile uploads a file to S3
// This replicates Python's upload_file method
func (s *S3StorageService) UploadFile(ctx context.Context, sourcePath string, destDirectory string, destFilename string) error {
	// Open source file
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	// Generate S3 key
	key := s.generateKey(s.pathBuilder.SafeJoinPaths(destDirectory, destFilename))

	// Get file info for content length
	fileInfo, err := sourceFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// Upload to S3
	_, err = s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.s3Config.Bucket),
		Key:           aws.String(key),
		Body:          sourceFile,
		ContentLength: aws.Int64(fileInfo.Size()),
		ContentType:   aws.String("application/octet-stream"),
	})
	if err != nil {
		return fmt.Errorf("failed to upload file to S3: %w", err)
	}

	return nil
}

// ReadFile reads file content from S3
// This replicates Python's read_file method
func (s *S3StorageService) ReadFile(ctx context.Context, path string, bytesMode bool) ([]byte, error) {
	key := s.generateKey(path)

	// Get object from S3
	result, err := s.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.s3Config.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		// Handle not found error
		if strings.Contains(err.Error(), "NoSuchKey") {
			return nil, fmt.Errorf("file not found: %s", path)
		}
		return nil, fmt.Errorf("failed to read file from S3: %w", err)
	}
	defer result.Body.Close()

	// Read the content
	content, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read S3 object content: %w", err)
	}

	return content, nil
}

// WriteFile writes content to S3
func (s *S3StorageService) WriteFile(ctx context.Context, path string, content any, binary bool) error {
	var body io.ReadSeeker

	switch v := content.(type) {
	case []byte:
		body = bytes.NewReader(v)
	case string:
		body = strings.NewReader(v)
	case io.ReadSeeker:
		body = v
	default:
		return fmt.Errorf("unsupported content type: %T", content)
	}

	key := s.generateKey(path)

	// Upload to S3
	_, err := s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.s3Config.Bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String("application/octet-stream"),
	})
	if err != nil {
		return fmt.Errorf("failed to write file to S3: %w", err)
	}

	return nil
}

// FileExists checks if a file exists in S3
func (s *S3StorageService) FileExists(ctx context.Context, path string) (bool, error) {
	key := s.generateKey(path)

	// Use HeadObject to check if file exists
	_, err := s.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.s3Config.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		// Handle not found error
		if strings.Contains(err.Error(), "NotFound") || strings.Contains(err.Error(), "NoSuchKey") {
			return false, nil
		}
		return false, fmt.Errorf("failed to check file existence in S3: %w", err)
	}

	return true, nil
}

// DirectoryExists is a no-op for S3 (S3 doesn't have directories)
func (s *S3StorageService) DirectoryExists(ctx context.Context, path string) (bool, error) {
	// In S3, directories don't exist - only files do
	// However, we can check if there are any objects with this prefix
	key := s.generateKey(path)
	if !strings.HasSuffix(key, "/") {
		key += "/"
	}

	// Use ListObjects to check for any objects with this prefix
	result, err := s.s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:  aws.String(s.s3Config.Bucket),
		Prefix:  aws.String(key),
		MaxKeys: aws.Int32(1), // We only need to know if at least one exists
	})
	if err != nil {
		return false, fmt.Errorf("failed to check directory existence in S3: %w", err)
	}

	return len(result.Contents) > 0, nil
}

// MakeDirectory is a no-op for S3
func (s *S3StorageService) MakeDirectory(ctx context.Context, directory string) error {
	// Directories are implicit in S3, so this is a no-op
	// We just return success to match Python behavior
	return nil
}

// DeleteFile deletes a file from S3
// This replicates Python's delete_file method
func (s *S3StorageService) DeleteFile(ctx context.Context, path string) error {
	key := s.generateKey(path)

	// Delete object from S3
	_, err := s.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.s3Config.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}

	return nil
}

// DeleteDirectory is a no-op for S3
// This replicates Python's delete_directory method
func (s *S3StorageService) DeleteDirectory(ctx context.Context, path string) error {
	// Directories don't exist in S3, so this is a no-op
	// In Python, this is used to clean up directory structures but S3 handles this implicitly
	return nil
}
