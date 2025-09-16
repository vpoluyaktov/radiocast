package server

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"radiocast/internal/models"
	"radiocast/internal/reports"
)

// FileManager handles file I/O operations only (load/store)
type FileManager struct {
	server *Server
}

// StoreAllFiles handles storing generated files to appropriate storage
// Uses unified path structure: reports/YYYY/MM/DD/PropagationReport-YYYY-MM-DD-HH-MM-SS/
func (fm *FileManager) StoreAllFiles(ctx context.Context, files *reports.GeneratedFiles, data *models.PropagationData) error {
	timestamp := data.Timestamp

	// Create temporary directory for file generation
	tempDir := filepath.Join(os.TempDir(), fmt.Sprintf("radiocast_temp_%d", time.Now().Unix()))
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir) // Clean up temp directory

	// Write all files to temporary directory first
	if err := fm.writeFilesToTemp(tempDir, files); err != nil {
		return fmt.Errorf("failed to write files to temp: %w", err)
	}

	// Copy/upload files to final destination
	if fm.server.DeploymentMode == DeploymentLocal {
		// Copy from temp to local reports directory
		finalDir := filepath.Join(fm.server.ReportsDir, files.FolderPath)
		if err := fm.copyTempToLocal(tempDir, finalDir); err != nil {
			return fmt.Errorf("failed to copy files to local storage: %w", err)
		}
		log.Printf("Report stored locally in: %s", finalDir)
	} else {
		// Upload from temp to remote storage
		if err := fm.uploadTempToRemote(ctx, tempDir, files, timestamp); err != nil {
			return fmt.Errorf("failed to upload files to remote storage: %w", err)
		}
		log.Printf("All files uploaded successfully to remote storage: %s", files.FolderPath)
	}

	return nil
}


// NewFileManager creates a new file manager
func NewFileManager(server *Server) *FileManager {
	return &FileManager{server: server}
}


// writeFilesToTemp writes all generated files to temporary directory
func (fm *FileManager) writeFilesToTemp(tempDir string, files *reports.GeneratedFiles) error {
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
	if err := fm.copyBackgroundImage(tempDir); err != nil {
		log.Printf("Warning: Failed to copy background image: %v", err)
	}
	
	return nil
}

// copyBackgroundImage copies background.png from assets directory to temp directory
func (fm *FileManager) copyBackgroundImage(tempDir string) error {
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
func (fm *FileManager) copyTempToLocal(tempDir, finalDir string) error {
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
func (fm *FileManager) uploadTempToRemote(ctx context.Context, tempDir string, files *reports.GeneratedFiles, timestamp time.Time) error {
	// Upload HTML report as index.html
	err := fm.server.Storage.StoreFile(ctx, []byte(files.HTMLContent), "index.html", timestamp)
	if err != nil {
		return fmt.Errorf("failed to store HTML report: %w", err)
	}
	
	// Upload JSON files
	for filename, data := range files.JSONFiles {
		if err := fm.server.Storage.StoreFile(ctx, data, filename, timestamp); err != nil {
			return fmt.Errorf("failed to store JSON file %s: %w", filename, err)
		}
	}
	
	// Upload asset files
	for filename, data := range files.AssetFiles {
		if err := fm.server.Storage.StoreFile(ctx, data, filename, timestamp); err != nil {
			return fmt.Errorf("failed to store asset file %s: %w", filename, err)
		}
	}
	
	// Upload background image from temp directory
	bgPath := filepath.Join(tempDir, "background.png")
	if bgData, err := os.ReadFile(bgPath); err == nil {
		if err := fm.server.Storage.StoreFile(ctx, bgData, "background.png", timestamp); err != nil {
			return fmt.Errorf("failed to store background image: %w", err)
		}
		log.Printf("Uploaded background.png (%d bytes)", len(bgData))
	} else {
		log.Printf("Warning: Failed to read background image for upload: %v", err)
	}
	
	return nil
}
