package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// LocalStorageClient handles local file system storage operations
type LocalStorageClient struct {
	baseDir string
}

// NewLocalStorageClient creates a new local storage client
func NewLocalStorageClient(baseDir string) (*LocalStorageClient, error) {
	// Ensure base directory exists
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory %s: %w", baseDir, err)
	}
	
	return &LocalStorageClient{
		baseDir: baseDir,
	}, nil
}

// Close is a no-op for local storage (implements same interface as GCSClient)
func (l *LocalStorageClient) Close() error {
	return nil
}


// StoreFile stores any file (JSON, text, etc.) locally in the same folder as the report
func (l *LocalStorageClient) StoreFile(ctx context.Context, fileData []byte, filename string, timestamp time.Time) error {
	// Generate the file path for this file
	filePath := filepath.Join(l.baseDir, GenerateReportFolderPath(timestamp), filename)
	
	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}
	
	// Write the file
	if err := os.WriteFile(filePath, fileData, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}
	
	return nil
}


// GetFile retrieves any file from local storage
func (l *LocalStorageClient) GetFile(ctx context.Context, filePath string) ([]byte, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	return data, nil
}

// GetLatestReport gets the most recent report from local storage
func (l *LocalStorageClient) GetLatestReport() (string, error) {
	ctx := context.Background()
	reports, err := l.ListReports(ctx, 1)
	if err != nil {
		return "", err
	}
	
	if len(reports) == 0 {
		return "", fmt.Errorf("no reports found")
	}
	
	return reports[0], nil
}


// ListReports lists recent reports from local storage, sorted by creation time (newest first)
func (l *LocalStorageClient) ListReports(ctx context.Context, limit int) ([]string, error) {
	reportsPath := filepath.Join(l.baseDir, "reports")
	
	var reportPaths []string
	
	err := filepath.Walk(reportsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors and continue
		}
		
		// Look for index.html files
		if info.Name() == "index.html" {
			// Get relative path from baseDir
			relPath, _ := filepath.Rel(l.baseDir, path)
			reportPaths = append(reportPaths, relPath)
		}
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to walk reports directory: %w", err)
	}
	
	// Sort alphabetically, then reverse for newest first
	sort.Strings(reportPaths)
	for i, j := 0, len(reportPaths)-1; i < j; i, j = i+1, j-1 {
		reportPaths[i], reportPaths[j] = reportPaths[j], reportPaths[i]
	}
	
	// Apply limit
	if limit > 0 && limit < len(reportPaths) {
		reportPaths = reportPaths[:limit]
	}
	
	return reportPaths, nil
}


