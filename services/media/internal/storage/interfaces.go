package storage

import (
	"context"
	"io"
	"time"
)

// MediaStorage defines the interface for media storage operations
type MediaStorage interface {
	// UploadFile uploads a file and returns the file URL
	UploadFile(ctx context.Context, reader io.Reader, fileSize int64, contentType, category, userID string) (string, error)
	
	// UploadFileWithPath uploads a file to a specific path and returns the file URL
	UploadFileWithPath(ctx context.Context, bucketName, objectPath string, reader io.Reader, fileSize int64, contentType string) (string, error)
	
	// GetFile downloads a file and returns a reader
	GetFile(ctx context.Context, bucketName, objectPath string) (io.ReadCloser, error)
	
	// DeleteFile deletes a file by URL
	DeleteFile(ctx context.Context, fileURL string) error
	
	// GetFileURL generates a presigned URL for file access
	GetFileURL(ctx context.Context, fileURL string, expiry time.Duration) (string, error)
}

// MediaCategory defines the categories for media files
type MediaCategory string

const (
	CategoryAvatar MediaCategory = "avatars"
	CategoryBanner MediaCategory = "banners" 
	CategoryPost   MediaCategory = "posts"
)

// String returns the string representation of MediaCategory
func (c MediaCategory) String() string {
	return string(c)
}
