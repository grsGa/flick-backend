package storage

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOClient wraps MinIO client with bucket operations
type MinIOClient struct {
	client     *minio.Client
	bucketName string
	endpoint   string
	publicURL  string
}

// MediaStorageConfig holds MinIO configuration
type MediaStorageConfig struct {
	Endpoint   string
	AccessKey  string
	SecretKey  string
	UseSSL     bool
	BucketName string
	PublicURL  string // Public URL for browser access
}

// NewMinIOClient creates a new MinIO client instance
func NewMinIOClient(endpoint, accessKey, secretKey, bucketName string) (*MinIOClient, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	// Use 127.0.0.1 for external access instead of container hostname
	publicURL := "http://127.0.0.1:9000"
	if endpoint != "minio:9000" {
		publicURL = fmt.Sprintf("http://%s", endpoint)
	}

	m := &MinIOClient{
		client:     client,
		bucketName: bucketName,
		publicURL:  publicURL,
	}

	// Ensure bucket exists
	if err := m.ensureBucket(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ensure bucket exists: %w", err)
	}

	return m, nil
}

// ensureBucket creates the bucket if it doesn't exist and sets public read policy
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

	// Set public read policy for the bucket
	policy := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": "*",
				"Action": ["s3:GetObject"],
				"Resource": ["arn:aws:s3:::%s/*"]
			}
		]
	}`, m.bucketName)

	err = m.client.SetBucketPolicy(ctx, m.bucketName, policy)
	if err != nil {
		fmt.Printf("[MINIO] Warning: failed to set bucket policy: %v\n", err)
		// Don't fail initialization if policy setting fails
	} else {
		fmt.Printf("[MINIO] Bucket policy set successfully for public read access\n")
	}

	return nil
}

// UploadFile uploads a file to MinIO using hierarchical path structure
func (m *MinIOClient) UploadFile(ctx context.Context, reader io.Reader, fileSize int64, contentType, category, userID string) (string, error) {
	// Generate unique filename
	fileID := uuid.New().String()
	extension := getExtensionFromContentType(contentType)

	// Create hierarchical path: category/userId/filename
	var objectName string
	switch category {
	case "avatars":
		objectName = fmt.Sprintf("avatars/%s/avatar_%s%s", userID, fileID, extension)
	case "banners":
		objectName = fmt.Sprintf("banners/%s/banner_%s%s", userID, fileID, extension)
	case "posts":
		// For posts, we'll need postID as well, but for now use fileID as placeholder
		objectName = fmt.Sprintf("posts/%s/%s/img_%s%s", userID, fileID, fileID, extension)
	default:
		// Fallback to simple structure
		objectName = fmt.Sprintf("%s/%s/%s%s", category, userID, fileID, extension)
	}

	// Upload file
	_, err := m.client.PutObject(ctx, m.bucketName, objectName, reader, fileSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	// Debug logging for MinIO upload
	fmt.Printf("[MINIO] File uploaded successfully to: %s\n", objectName)
	fmt.Printf("[MINIO] Bucket: %s, Size: %d bytes\n", m.bucketName, fileSize)

	// Return the file URL - use public URL for browser access
	publicURL := m.publicURL
	if publicURL == "" {
		// Fallback to endpoint if no public URL configured
		protocol := "http"
		if strings.Contains(m.endpoint, "https") {
			protocol = "https"
		}
		publicURL = fmt.Sprintf("%s://%s", protocol, m.endpoint)
	}
	
	finalURL := fmt.Sprintf("%s/%s/%s", publicURL, m.bucketName, objectName)
	fmt.Printf("[MINIO] Generated public URL: %s\n", finalURL)
	fmt.Printf("[MINIO] Public URL config: %s\n", publicURL)
	
	return finalURL, nil
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
	// Support both formats:
	// 1. Full URL: http://minio:9000/social-media/avatars/user123/avatar_abc.jpg
	// 2. Relative path: /social-media/avatars/user123/avatar_abc.jpg

	// Remove protocol and host if present
	if strings.Contains(fileURL, "://") {
		parts := strings.SplitN(fileURL, "://", 2)
		if len(parts) == 2 {
			// Remove host part, keep path
			hostAndPath := parts[1]
			slashIndex := strings.Index(hostAndPath, "/")
			if slashIndex != -1 {
				fileURL = hostAndPath[slashIndex:]
			}
		}
	}

	// Expected format: /bucket-name/category/filename.ext
	prefix := fmt.Sprintf("/%s/", bucketName)
	if strings.HasPrefix(fileURL, prefix) {
		return strings.TrimPrefix(fileURL, prefix)
	}
	return ""
}
