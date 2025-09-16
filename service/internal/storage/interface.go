package storage

import (
	"context"
)

// StorageClient defines the interface for basic storage operations
type StorageClient interface {
	// Close closes the storage client
	Close() error
	
	// CreateDir creates a directory (and any necessary parent directories)
	CreateDir(ctx context.Context, dirPath string) error
	
	// StoreFile stores a file at the specified path
	StoreFile(ctx context.Context, filePath string, fileData []byte) error
	
	// GetFile retrieves a file from the specified path
	GetFile(ctx context.Context, filePath string) ([]byte, error)
	
	// ListDir lists contents of a directory
	ListDir(ctx context.Context, dirPath string, recursive bool) ([]string, error)
	
	// FileExists checks if a file exists at the specified path
	FileExists(ctx context.Context, filePath string) (bool, error)
}
