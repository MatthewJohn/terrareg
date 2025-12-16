package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/storage/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/storage/service"
	"github.com/rs/zerolog"
)

// S3StorageService implements StorageService for S3
// This replicates the Python S3FileStorage class
type S3StorageService struct {
	s3Config    *model.S3Config
	pathBuilder service.PathBuilder
	s3Client    *s3.Client
	logger      *zerolog.Logger
}

// NewS3StorageService creates a new S3 storage service
func NewS3StorageService(
	s3Config *model.S3Config,
	pathBuilder service.PathBuilder,
	logger *zerolog.Logger,
) (*S3StorageService, error) {
	if s3Config == nil {
		return nil, fmt.Errorf("S3 config must not be nil")
	}
	if s3Config.Bucket == "" {
		return nil, fmt.Errorf("S3 bucket is required")
	}

	return &S3StorageService{
		s3Config:    s3Config,
		pathBuilder: pathBuilder,
		logger:      logger,
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
	file, err := os.Open(sourcePath)
	if err != nil {
		log.Printf("Couldn't open file %v to upload. Here's why: %v\n", fileName, err)
		return err
	}
	defer file.Close()

	objectKey := fmt.Sprintf("%s/%s/%s", s.s3Config.KeyPrefix, destDirectory, destFilename)
	_, err = s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.s3Config.Bucket),
		Key:    aws.String(objectKey),
		Body:   file,
	})
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) && apiErr.ErrorCode() == "EntityTooLarge" {
			log.Printf("Error while uploading object to %s. The object is too large.\n"+
				"To upload objects larger than 5GB, use the S3 console (160GB max)\n"+
				"or the multipart upload API (5TB max).", s.s3Config.Bucket)
		} else {
			log.Printf("Couldn't upload file %v to %v:%v. Here's why: %v\n",
				sourcePath, s.s3Config.Bucket, objectKey, err)
		}
		return err
	}
	err = s3.NewObjectExistsWaiter(s.s3Client).Wait(
		ctx, &s3.HeadObjectInput{Bucket: aws.String(s.s3Config.Bucket), Key: aws.String(objectKey)}, time.Minute)
	if err != nil {
		log.Printf("Failed attempt to wait for object %s to exist.\n", objectKey)
	}
	return err
}

// ReadFile reads file content from S3
// This replicates Python's read_file method
func (s *S3StorageService) ReadFile(ctx context.Context, path string, bytesMode bool) ([]byte, error) {
	key := s.generateKey(path)

	resp, err := s.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.s3Config.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// bytesMode has no behavioral difference in Go since []byte is native
	return data, nil
}

// WriteFile writes content to S3
func (s *S3StorageService) WriteFile(ctx context.Context, path string, content any, binary bool) error {
	key := s.generateKey(path)

	var body io.Reader

	switch v := content.(type) {
	case []byte:
		body = bytes.NewReader(v)
	case string:
		body = strings.NewReader(v)
	case io.Reader:
		body = v
	default:
		return fmt.Errorf("unsupported content type %T", content)
	}

	_, err := s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.s3Config.Bucket),
		Key:    aws.String(key),
		Body:   body,
	})
	return err
}

// FileExists checks if a file exists in S3
func (s *S3StorageService) FileExists(ctx context.Context, path string) (bool, error) {
	key := s.generateKey(path)

	_, err := s.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.s3Config.Bucket),
		Key:    aws.String(key),
	})

	if err == nil {
		return true, nil
	}

	var apiErr smithy.APIError
	if errors.As(err, &apiErr) && apiErr.ErrorCode() == "NotFound" {
		return false, nil
	}

	return false, err
}

// DirectoryExists is a no-op for S3 (S3 doesn't have directories)
func (s *S3StorageService) DirectoryExists(ctx context.Context, path string) (bool, error) {
	return true, nil // Directories don't exist in S3
}

// MakeDirectory is a no-op for S3
func (s *S3StorageService) MakeDirectory(ctx context.Context, directory string) error {
	return nil // Directories are implicit in S3
}

// DeleteFile deletes a file from S3
// This replicates Python's delete_file method
func (s *S3StorageService) DeleteFile(ctx context.Context, path string) error {
	key := s.generateKey(path)

	_, err := s.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.s3Config.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}

	// Wait until the object is actually deleted
	return s3.NewObjectNotExistsWaiter(s.s3Client).Wait(
		ctx,
		&s3.HeadObjectInput{
			Bucket: aws.String(s.s3Config.Bucket),
			Key:    aws.String(key),
		},
		time.Minute,
	)
}

// DeleteDirectory is a no-op for S3
// This replicates Python's delete_directory method
func (s *S3StorageService) DeleteDirectory(ctx context.Context, path string) error {
	return nil // Directories don't exist in S3
}
