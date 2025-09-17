package reports

import (
	"context"
	"fmt"
	"log"
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
	
	// Note: Asset files (CSS, background image) are now served from /static/ folder
	// No need to store them in each report folder
	
	return nil
}

