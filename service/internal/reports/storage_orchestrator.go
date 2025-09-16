package reports

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"radiocast/internal/models"
	"radiocast/internal/storage"
)

// StorageOrchestrator handles the business logic of storing generated files
type StorageOrchestrator struct {
	reportsDir     string
	storage        storage.StorageClient
	deploymentMode string
}

// NewStorageOrchestrator creates a new storage orchestrator
func NewStorageOrchestrator(reportsDir string, storage storage.StorageClient, deploymentMode string) *StorageOrchestrator {
	return &StorageOrchestrator{
		reportsDir:     reportsDir,
		storage:        storage,
		deploymentMode: deploymentMode,
	}
}

// StoreAllFiles handles storing generated files to appropriate storage
func (so *StorageOrchestrator) StoreAllFiles(ctx context.Context, files *GeneratedFiles, data *models.PropagationData) error {
	timestamp := data.Timestamp

	// Create temporary directory for file generation
	tempDir := filepath.Join(os.TempDir(), fmt.Sprintf("radiocast_temp_%d", time.Now().Unix()))
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir) // Clean up temp directory

	// Write all files to temporary directory first
	if err := so.writeFilesToTemp(tempDir, files); err != nil {
		return fmt.Errorf("failed to write files to temp: %w", err)
	}

	// Copy/upload files to final destination
	if so.deploymentMode == "local" {
		// Copy from temp to local reports directory
		finalDir := filepath.Join(so.reportsDir, files.FolderPath)
		if err := so.copyTempToLocal(tempDir, finalDir); err != nil {
			return fmt.Errorf("failed to copy files to local storage: %w", err)
		}
		log.Printf("Report stored locally in: %s", finalDir)
	} else {
		// Upload from temp to remote storage
		if err := so.uploadTempToRemote(ctx, tempDir, files, timestamp); err != nil {
			return fmt.Errorf("failed to upload files to remote storage: %w", err)
		}
		log.Printf("All files uploaded successfully to remote storage: %s", files.FolderPath)
	}

	return nil
}

// writeFilesToTemp writes all generated files to temporary directory
func (so *StorageOrchestrator) writeFilesToTemp(tempDir string, files *GeneratedFiles) error {
	// Write HTML file
	htmlPath := filepath.Join(tempDir, "index.html")
	if err := os.WriteFile(htmlPath, []byte(files.HTMLContent), 0644); err != nil {
		return fmt.Errorf("failed to write HTML file: %w", err)
	}
	
	// Write JSON files
	for filename, data := range files.JSONFiles {
		path := filepath.Join(tempDir, filename)
		if err := os.WriteFile(path, data, 0644); err != nil {
			return fmt.Errorf("failed to write JSON file %s: %w", filename, err)
		}
	}
	
	// Write asset files
	for filename, data := range files.AssetFiles {
		path := filepath.Join(tempDir, filename)
		if err := os.WriteFile(path, data, 0644); err != nil {
			return fmt.Errorf("failed to write asset file %s: %w", filename, err)
		}
	}
	
	// Copy background image from assets directory
	if err := so.copyBackgroundImage(tempDir); err != nil {
		log.Printf("Warning: Failed to copy background image: %v", err)
	}
	
	return nil
}

// copyBackgroundImage copies background.png from assets directory to temp directory
func (so *StorageOrchestrator) copyBackgroundImage(tempDir string) error {
	// Try to find background image in various locations
	candidates := []string{
		filepath.Join("internal", "assets", "background.png"),
		filepath.Join("service", "internal", "assets", "background.png"),
		filepath.Join("..", "service", "internal", "assets", "background.png"),
	}
	
	for _, path := range candidates {
		if data, err := os.ReadFile(path); err == nil {
			bgPath := filepath.Join(tempDir, "background.png")
			if err := os.WriteFile(bgPath, data, 0644); err != nil {
				return fmt.Errorf("failed to write background image: %w", err)
			}
			log.Printf("Copied background.png (%d bytes)", len(data))
			return nil
		}
	}
	
	return fmt.Errorf("background.png not found in any candidate location")
}

// copyTempToLocal copies files from temp directory to local reports directory
func (so *StorageOrchestrator) copyTempToLocal(tempDir, finalDir string) error {
	// Create final directory
	if err := os.MkdirAll(finalDir, 0755); err != nil {
		return fmt.Errorf("failed to create final directory: %w", err)
	}
	
	// Copy all files from temp to final directory
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		return fmt.Errorf("failed to read temp directory: %w", err)
	}
	
	for _, entry := range entries {
		if entry.IsDir() {
			continue // Skip subdirectories for now
		}
		
		srcPath := filepath.Join(tempDir, entry.Name())
		dstPath := filepath.Join(finalDir, entry.Name())
		
		data, err := os.ReadFile(srcPath)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", srcPath, err)
		}
		
		if err := os.WriteFile(dstPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", dstPath, err)
		}
	}
	
	return nil
}

// uploadTempToRemote uploads files from temp directory to remote storage
func (so *StorageOrchestrator) uploadTempToRemote(ctx context.Context, tempDir string, files *GeneratedFiles, timestamp time.Time) error {
	// Upload HTML report as index.html
	err := so.storage.StoreFile(ctx, []byte(files.HTMLContent), "index.html", timestamp)
	if err != nil {
		return fmt.Errorf("failed to store HTML report: %w", err)
	}
	
	// Upload JSON files
	for filename, data := range files.JSONFiles {
		if err := so.storage.StoreFile(ctx, data, filename, timestamp); err != nil {
			return fmt.Errorf("failed to store JSON file %s: %w", filename, err)
		}
	}
	
	// Upload asset files
	for filename, data := range files.AssetFiles {
		if err := so.storage.StoreFile(ctx, data, filename, timestamp); err != nil {
			return fmt.Errorf("failed to store asset file %s: %w", filename, err)
		}
	}
	
	// Upload background image from temp directory
	bgPath := filepath.Join(tempDir, "background.png")
	if bgData, err := os.ReadFile(bgPath); err == nil {
		if err := so.storage.StoreFile(ctx, bgData, "background.png", timestamp); err != nil {
			return fmt.Errorf("failed to store background image: %w", err)
		}
		log.Printf("Uploaded background.png (%d bytes)", len(bgData))
	} else {
		log.Printf("Warning: Failed to read background image for upload: %v", err)
	}
	
	return nil
}
