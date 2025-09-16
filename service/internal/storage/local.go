package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// LocalStorageClient handles local file system storage operations
// Uses local_gcs as the root directory to mirror GCS bucket structure
type LocalStorageClient struct {
	rootDir string // Always "local_gcs"
}

// NewLocalStorageClient creates a new local storage client
func NewLocalStorageClient(baseDir string) (*LocalStorageClient, error) {
	// Always use local_gcs as the root directory, ignore baseDir parameter
	rootDir := "local_gcs"
	
	// Ensure root directory exists
	if err := os.MkdirAll(rootDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create root directory %s: %w", rootDir, err)
	}
	
	return &LocalStorageClient{
		rootDir: rootDir,
	}, nil
}

// Close is a no-op for local storage
func (l *LocalStorageClient) Close() error {
	return nil
}

// CreateDir creates a directory (and any necessary parent directories)
func (l *LocalStorageClient) CreateDir(ctx context.Context, dirPath string) error {
	fullPath := filepath.Join(l.rootDir, dirPath)
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", fullPath, err)
	}
	return nil
}

// StoreFile stores a file at the specified path
func (l *LocalStorageClient) StoreFile(ctx context.Context, filePath string, fileData []byte) error {
	fullPath := filepath.Join(l.rootDir, filePath)
	
	// Ensure directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}
	
	// Write the file
	if err := os.WriteFile(fullPath, fileData, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", fullPath, err)
	}
	
	return nil
}

// GetFile retrieves a file from the specified path
func (l *LocalStorageClient) GetFile(ctx context.Context, filePath string) ([]byte, error) {
	fullPath := filepath.Join(l.rootDir, filePath)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", fullPath, err)
	}
	return data, nil
}

// ListDir lists contents of a directory
func (l *LocalStorageClient) ListDir(ctx context.Context, dirPath string, recursive bool) ([]string, error) {
	fullPath := filepath.Join(l.rootDir, dirPath)
	
	var files []string
	
	if recursive {
		err := filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip errors and continue
			}
			
			if !info.IsDir() {
				// Get relative path from rootDir
				relPath, _ := filepath.Rel(l.rootDir, path)
				files = append(files, relPath)
			}
			return nil
		})
		
		if err != nil {
			return nil, fmt.Errorf("failed to walk directory %s: %w", fullPath, err)
		}
	} else {
		entries, err := os.ReadDir(fullPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read directory %s: %w", fullPath, err)
		}
		
		for _, entry := range entries {
			entryPath := filepath.Join(dirPath, entry.Name())
			files = append(files, entryPath)
		}
	}
	
	return files, nil
}

// FileExists checks if a file exists at the specified path
func (l *LocalStorageClient) FileExists(ctx context.Context, filePath string) (bool, error) {
	fullPath := filepath.Join(l.rootDir, filePath)
	_, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check file existence %s: %w", fullPath, err)
	}
	return true, nil
}


