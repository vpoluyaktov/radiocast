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
	storage        storage.StorageClient
	deploymentMode string
}

// NewStorageOrchestrator creates a new storage orchestrator
func NewStorageOrchestrator(storage storage.StorageClient, deploymentMode string) *StorageOrchestrator {
	return &StorageOrchestrator{
		storage:        storage,
		deploymentMode: deploymentMode,
	}
}

// StoreAllFiles handles storing generated files using StorageClient
func (so *StorageOrchestrator) StoreAllFiles(ctx context.Context, files *GeneratedFiles, data *models.PropagationData) error {
	timestamp := data.Timestamp

	// Store files directly using StorageClient (no temp directory needed)
	if err := so.storeFilesViaStorage(ctx, files, timestamp); err != nil {
		return fmt.Errorf("failed to store files: %w", err)
	}
	log.Printf("All files stored successfully via storage client")

	return nil
}

// storeFilesViaStorage stores files using the StorageClient interface
func (so *StorageOrchestrator) storeFilesViaStorage(ctx context.Context, files *GeneratedFiles, timestamp time.Time) error {
	// Build report folder path
	reportFolderPath := "reports/" + storage.GenerateReportFolderPath(timestamp)
	
	// Store HTML file
	htmlPath := reportFolderPath + "/index.html"
	if err := so.storage.StoreFile(ctx, htmlPath, []byte(files.HTMLContent)); err != nil {
		return fmt.Errorf("failed to store HTML file: %w", err)
	}
	
	// Store JSON files
	for filename, data := range files.JSONFiles {
		jsonPath := reportFolderPath + "/" + filename
		if err := so.storage.StoreFile(ctx, jsonPath, data); err != nil {
			return fmt.Errorf("failed to store JSON file %s: %w", filename, err)
		}
	}
	
	// Store asset files
	for filename, data := range files.AssetFiles {
		assetPath := reportFolderPath + "/" + filename
		if err := so.storage.StoreFile(ctx, assetPath, data); err != nil {
			return fmt.Errorf("failed to store asset file %s: %w", filename, err)
		}
	}
	
	// Store background image
	if err := so.storeBackgroundImage(ctx, reportFolderPath); err != nil {
		log.Printf("Warning: Failed to store background image: %v", err)
	}
	
	return nil
}

// storeBackgroundImage stores background.png using StorageClient
func (so *StorageOrchestrator) storeBackgroundImage(ctx context.Context, reportFolderPath string) error {
	// Try to find background image in various locations
	candidates := []string{
		filepath.Join("internal", "assets", "background.png"),
		filepath.Join("service", "internal", "assets", "background.png"),
		filepath.Join("..", "service", "internal", "assets", "background.png"),
	}
	
	for _, path := range candidates {
		if data, err := os.ReadFile(path); err == nil {
			backgroundPath := reportFolderPath + "/background.png"
			if err := so.storage.StoreFile(ctx, backgroundPath, data); err != nil {
				return fmt.Errorf("failed to store background image: %w", err)
			}
			log.Printf("Stored background.png (%d bytes)", len(data))
			return nil
		}
	}
	
	return fmt.Errorf("background.png not found in any candidate location")
}
