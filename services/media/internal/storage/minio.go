package storage

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/google/uuid"
)

// MinIOClient wraps MinIO client with bucket operations
type MinIOClient struct {
	client     *minio.Client
	bucketName string
}

// MediaStorageConfig holds MinIO configuration
type MediaStorageConfig struct {
	Endpoint   string
	AccessKey  string
	SecretKey  string
	UseSSL     bool
	BucketName string
}

// NewMinIOClient creates a new MinIO client instance
func NewMinIOClient(config MediaStorageConfig) (*MinIOClient, error) {
	// Initialize MinIO client
	minioClient, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKey, config.SecretKey, ""),
		Secure: config.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	client := &MinIOClient{
		client:     minioClient,
		bucketName: config.BucketName,
	}

	// Ensure bucket exists
	if err := client.ensureBucket(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ensure bucket exists: %w", err)
	}

	return client, nil
}

// ensureBucket creates the bucket if it doesn't exist
func (m *MinIOClient) ensureBucket(ctx context.Context) error {
	exists, err := m.client.BucketExists(ctx, m.bucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = m.client.MakeBucket(ctx, m.bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return nil
}

// UploadFile uploads a file to MinIO and returns the file URL
func (m *MinIOClient) UploadFile(ctx context.Context, reader io.Reader, fileSize int64, contentType, category string) (string, error) {
	// Generate unique filename
	fileID := uuid.New().String()
	extension := getExtensionFromContentType(contentType)
	objectName := fmt.Sprintf("%s/%s%s", category, fileID, extension)

	// Upload file
	_, err := m.client.PutObject(ctx, m.bucketName, objectName, reader, fileSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	// Return the file URL
	return fmt.Sprintf("/%s/%s", m.bucketName, objectName), nil
}

// DeleteFile deletes a file from MinIO
func (m *MinIOClient) DeleteFile(ctx context.Context, fileURL string) error {
	objectName := extractObjectNameFromURL(fileURL, m.bucketName)
	if objectName == "" {
		return fmt.Errorf("invalid file URL: %s", fileURL)
	}

	err := m.client.RemoveObject(ctx, m.bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// GetFileURL generates a presigned URL for file access
func (m *MinIOClient) GetFileURL(ctx context.Context, fileURL string, expiry time.Duration) (string, error) {
	objectName := extractObjectNameFromURL(fileURL, m.bucketName)
	if objectName == "" {
		return "", fmt.Errorf("invalid file URL: %s", fileURL)
	}

	presignedURL, err := m.client.PresignedGetObject(ctx, m.bucketName, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignedURL.String(), nil
}

// getExtensionFromContentType returns file extension based on content type
func getExtensionFromContentType(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "video/mp4":
		return ".mp4"
	case "video/webm":
		return ".webm"
	case "video/quicktime":
		return ".mov"
	default:
		return ".bin"
	}
}

// extractObjectNameFromURL extracts object name from file URL
func extractObjectNameFromURL(fileURL, bucketName string) string {
	// Expected format: /bucket-name/category/filename.ext
	prefix := fmt.Sprintf("/%s/", bucketName)
	if strings.HasPrefix(fileURL, prefix) {
		return strings.TrimPrefix(fileURL, prefix)
	}
	return ""
}
