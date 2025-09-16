package storage

import (
	"context"
	"fmt"

	"radiocast/internal/config"
)

// DeploymentMode represents the deployment environment
type DeploymentMode string

const (
	DeploymentLocal DeploymentMode = "local"
	DeploymentGCS   DeploymentMode = "gcs"
)

// NewStorageClient creates a storage client based on deployment mode and configuration
func NewStorageClient(ctx context.Context, deploymentMode DeploymentMode, cfg *config.Config) (StorageClient, error) {
	switch deploymentMode {
	case DeploymentLocal:
		// Determine reports directory for local storage
		reportsDir := cfg.LocalReportsDir
		if reportsDir == "" {
			reportsDir = "reports" // Default fallback
		}
		
		localClient, err := NewLocalStorageClient(reportsDir)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize local storage client: %w", err)
		}
		return localClient, nil
		
	case DeploymentGCS:
		gcsClient, err := NewGCSClient(ctx, cfg.GCSBucket)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize GCS client: %w", err)
		}
		return gcsClient, nil
		
	default:
		return nil, fmt.Errorf("unsupported deployment mode: %s", deploymentMode)
	}
}
