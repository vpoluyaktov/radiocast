package server

import (
	"context"
	"fmt"
	"log"

	"radiocast/internal/config"
	"radiocast/internal/fetchers"
	"radiocast/internal/llm"
	"radiocast/internal/reports"
	"radiocast/internal/storage"
)

// DeploymentMode represents the deployment mode
type DeploymentMode string

const (
	DeploymentLocal DeploymentMode = "local"
	DeploymentGCS   DeploymentMode = "gcs"
)

// Server represents the main application server
type Server struct {
	Config         *config.Config
	Fetcher        *fetchers.DataFetcher
	LLMClient      *llm.OpenAIClient
	Generator      *reports.Generator
	Storage        *storage.GCSClient
	DeploymentMode DeploymentMode
	ReportsDir     string
}

// NewServer creates a new server instance
func NewServer(cfg *config.Config, deploymentMode DeploymentMode) (*Server, error) {
	ctx := context.Background()
	
	// Determine reports directory
	reportsDir := "reports" // Default to service/reports
	if deploymentMode == DeploymentLocal {
		reportsDir = cfg.LocalReportsDir
		if reportsDir == "" {
			reportsDir = "reports" // Fallback to service/reports
		}
	}
	
	server := &Server{
		Config:         cfg,
		Fetcher:        fetchers.NewDataFetcher(),
		LLMClient:      llm.NewOpenAIClient(cfg.OpenAIAPIKey, cfg.OpenAIModel),
		DeploymentMode: deploymentMode,
		ReportsDir:     reportsDir,
	}
	
	// Initialize components based on deployment mode
	if deploymentMode == DeploymentLocal {
		log.Printf("Local deployment mode - reports will be saved to: %s", reportsDir)
		server.Generator = reports.NewGenerator(reportsDir)
		server.Storage = nil
	} else {
		log.Printf("GCS deployment mode - reports will be saved to GCS bucket: %s", cfg.GCSBucket)
		gcsClient, err := storage.NewGCSClient(ctx, cfg.GCSBucket)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize GCS client: %w", err)
		}
		server.Generator = reports.NewGenerator("") // Empty for GCS mode
		server.Storage = gcsClient
	}
	
	return server, nil
}

// Close cleans up server resources
func (s *Server) Close() error {
	if s.Storage != nil {
		return s.Storage.Close()
	}
	return nil
}
