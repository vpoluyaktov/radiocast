package storage

import (
	"context"
	"time"
)

// StorageClient defines the interface for storage operations
type StorageClient interface {
	// Close closes the storage client
	Close() error
	
	// StoreFile stores any file (JSON, text, HTML, images, etc.) in the report folder
	StoreFile(ctx context.Context, fileData []byte, filename string, timestamp time.Time) error
	
	// GetFile retrieves any file
	GetFile(ctx context.Context, filePath string) ([]byte, error)
	
	// GetLatestReport gets the most recent report
	GetLatestReport() (string, error)
	
	// ListReports lists recent reports, sorted by creation time (newest first)
	ListReports(ctx context.Context, limit int) ([]string, error)
}
