package server

import (
	"fmt"
	"os"
	"path/filepath"
)

// FileManager handles basic file I/O operations only
type FileManager struct {
	baseDir string
}

// NewFileManager creates a new file manager
func NewFileManager(baseDir string) *FileManager {
	return &FileManager{baseDir: baseDir}
}

// WriteFile writes data to a file at the specified path
func (fm *FileManager) WriteFile(filePath string, data []byte) error {
	fullPath := filepath.Join(fm.baseDir, filePath)
	
	// Create directory if it doesn't exist
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}
	
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", fullPath, err)
	}
	
	return nil
}

// ReadFile reads data from a file at the specified path
func (fm *FileManager) ReadFile(filePath string) ([]byte, error) {
	fullPath := filepath.Join(fm.baseDir, filePath)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", fullPath, err)
	}
	return data, nil
}

// CopyFile copies a file from source to destination
func (fm *FileManager) CopyFile(srcPath, dstPath string) error {
	data, err := fm.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}
	
	if err := fm.WriteFile(dstPath, data); err != nil {
		return fmt.Errorf("failed to write destination file: %w", err)
	}
	
	return nil
}

// CreateDirectory creates a directory at the specified path
func (fm *FileManager) CreateDirectory(dirPath string) error {
	fullPath := filepath.Join(fm.baseDir, dirPath)
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", fullPath, err)
	}
	return nil
}

// DeleteDirectory removes a directory and all its contents
func (fm *FileManager) DeleteDirectory(dirPath string) error {
	fullPath := filepath.Join(fm.baseDir, dirPath)
	if err := os.RemoveAll(fullPath); err != nil {
		return fmt.Errorf("failed to delete directory %s: %w", fullPath, err)
	}
	return nil
}

// ListFiles returns a list of files in the specified directory
func (fm *FileManager) ListFiles(dirPath string) ([]string, error) {
	fullPath := filepath.Join(fm.baseDir, dirPath)
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", fullPath, err)
	}
	
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}
	
	return files, nil
}

