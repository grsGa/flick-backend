package repository

import (
	"context"
	"io"
	"time"

	"github.com/flick/backend/services/media/internal/storage"
)

// MinIOStorageRepository implements MediaStorageRepository using MinIO
type MinIOStorageRepository struct {
	client *storage.MinIOClient
}

// NewMinIOStorageRepository creates a new MinIO storage repository
func NewMinIOStorageRepository(client *storage.MinIOClient) *MinIOStorageRepository {
	return &MinIOStorageRepository{
		client: client,
	}
}

// UploadFile uploads a file to MinIO storage with user ID for hierarchical structure
func (r *MinIOStorageRepository) UploadFile(ctx context.Context, reader io.Reader, fileSize int64, contentType, category, userID string) (string, error) {
	return r.client.UploadFile(ctx, reader, fileSize, contentType, category, userID)
}

// DeleteFile deletes a file from MinIO storage
func (r *MinIOStorageRepository) DeleteFile(ctx context.Context, fileURL string) error {
	return r.client.DeleteFile(ctx, fileURL)
}

// GetFileURL generates a presigned URL for file access
func (r *MinIOStorageRepository) GetFileURL(ctx context.Context, fileURL string, expiry time.Duration) (string, error) {
	return r.client.GetFileURL(ctx, fileURL, expiry)
}
