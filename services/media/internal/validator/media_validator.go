package validator

import (
	"fmt"
	"mime"
	"path/filepath"
	"strings"

	"github.com/flick/backend/services/media/internal/storage"
)

// MediaLimits defines upload limits for different media types
type MediaLimits struct {
	MaxFileSize    int64    // in bytes
	AllowedTypes   []string // MIME types
	AllowedFormats []string // file extensions
}

// MediaValidator handles media file validation
type MediaValidator struct {
	limits map[storage.MediaCategory]MediaLimits
}

// NewMediaValidator creates a new media validator with predefined limits
func NewMediaValidator() *MediaValidator {
	return &MediaValidator{
		limits: map[storage.MediaCategory]MediaLimits{
			// Avatar limits: 5MB, common image formats
			storage.CategoryAvatar: {
				MaxFileSize:    5 * 1024 * 1024, // 5MB
				AllowedTypes:   []string{"image/jpeg", "image/png", "image/webp", "image/gif"},
				AllowedFormats: []string{".jpg", ".jpeg", ".png", ".webp", ".gif"},
			},
			// Banner limits: 10MB, common image formats
			storage.CategoryBanner: {
				MaxFileSize:    10 * 1024 * 1024, // 10MB
				AllowedTypes:   []string{"image/jpeg", "image/png", "image/webp", "image/gif"},
				AllowedFormats: []string{".jpg", ".jpeg", ".png", ".webp", ".gif"},
			},
			// Post media limits: 100MB, images and videos
			storage.CategoryPost: {
				MaxFileSize: 100 * 1024 * 1024, // 100MB
				AllowedTypes: []string{
					// Images
					"image/jpeg", "image/png", "image/webp", "image/gif",
					// Videos
					"video/mp4", "video/webm", "video/quicktime", "video/x-msvideo",
				},
				AllowedFormats: []string{
					// Images
					".jpg", ".jpeg", ".png", ".webp", ".gif",
					// Videos
					".mp4", ".webm", ".mov", ".avi",
				},
			},
		},
	}
}

// ValidateUpload validates a media file upload request
func (v *MediaValidator) ValidateUpload(category storage.MediaCategory, filename string, fileSize int64, contentType string) error {
	limits, exists := v.limits[category]
	if !exists {
		return fmt.Errorf("unsupported media category: %s", category)
	}

	// Validate file size
	if fileSize > limits.MaxFileSize {
		return fmt.Errorf("file size %d bytes exceeds maximum allowed size %d bytes for category %s", 
			fileSize, limits.MaxFileSize, category)
	}

	// Validate content type
	if !v.isContentTypeAllowed(contentType, limits.AllowedTypes) {
		return fmt.Errorf("content type %s is not allowed for category %s. Allowed types: %v", 
			contentType, category, limits.AllowedTypes)
	}

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(filename))
	if !v.isExtensionAllowed(ext, limits.AllowedFormats) {
		return fmt.Errorf("file extension %s is not allowed for category %s. Allowed extensions: %v", 
			ext, category, limits.AllowedFormats)
	}

	// Cross-validate content type and extension
	if !v.isContentTypeConsistentWithExtension(contentType, ext) {
		return fmt.Errorf("content type %s is inconsistent with file extension %s", contentType, ext)
	}

	return nil
}

// isContentTypeAllowed checks if content type is in allowed list
func (v *MediaValidator) isContentTypeAllowed(contentType string, allowedTypes []string) bool {
	for _, allowed := range allowedTypes {
		if contentType == allowed {
			return true
		}
	}
	return false
}

// isExtensionAllowed checks if file extension is in allowed list
func (v *MediaValidator) isExtensionAllowed(ext string, allowedFormats []string) bool {
	for _, allowed := range allowedFormats {
		if ext == allowed {
			return true
		}
	}
	return false
}

// isContentTypeConsistentWithExtension validates content type matches file extension
func (v *MediaValidator) isContentTypeConsistentWithExtension(contentType, ext string) bool {
	// Handle common variations explicitly
	switch {
	case contentType == "image/jpeg" && (ext == ".jpg" || ext == ".jpeg"):
		return true
	case contentType == "image/png" && ext == ".png":
		return true
	case contentType == "image/gif" && ext == ".gif":
		return true
	case contentType == "image/webp" && ext == ".webp":
		return true
	case contentType == "video/mp4" && ext == ".mp4":
		return true
	case contentType == "video/webm" && ext == ".webm":
		return true
	case contentType == "video/quicktime" && ext == ".mov":
		return true
	case contentType == "video/x-msvideo" && ext == ".avi":
		return true
	default:
		// Get expected MIME type from extension as fallback
		expectedType := mime.TypeByExtension(ext)
		if expectedType == "" {
			return false
		}
		// Check if content type matches expected type
		return contentType == expectedType || strings.HasPrefix(expectedType, contentType)
	}
}

// GetLimitsForCategory returns upload limits for a specific category
func (v *MediaValidator) GetLimitsForCategory(category storage.MediaCategory) (MediaLimits, error) {
	limits, exists := v.limits[category]
	if !exists {
		return MediaLimits{}, fmt.Errorf("unsupported media category: %s", category)
	}
	return limits, nil
}
