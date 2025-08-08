package logger

import (
	"go.uber.org/zap"
)

// NewLogger creates a new Zap logger.
func NewLogger() (*zap.Logger, error) {
	// For development, a simple development logger is fine.
	// For production, you would configure a more robust logger (e.g., with JSON output).
	return zap.NewDevelopment()
}
