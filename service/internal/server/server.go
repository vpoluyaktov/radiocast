package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"radiocast/internal/config"
	"radiocast/internal/fetchers"
	"radiocast/internal/llm"
	"radiocast/internal/mocks"
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
	MockService    *mocks.MockService
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
	
	// Initialize mock service if mockup mode is enabled
	if cfg.MockupMode {
		mocksDir := filepath.Join("internal", "mocks")
		server.MockService = mocks.NewMockService(mocksDir)
		log.Printf("Mockup mode enabled - using mock data from %s", mocksDir)
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

// SetupRoutes configures HTTP routes for the server
func (s *Server) SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	
	// Serve static report files from reports directory
	reportsDir := "./reports/"
	if s.DeploymentMode == DeploymentGCS {
		reportsDir = "/tmp/reports/"
	}
	fileServer := http.FileServer(http.Dir(reportsDir))
	
	// Handle specific API routes first
	mux.HandleFunc("/health", s.HandleHealth)
	mux.HandleFunc("/generate", s.HandleGenerate)
	mux.HandleFunc("/reports", s.HandleListReports)
	mux.HandleFunc("/files/", s.HandleFileProxy)
	
	// Handle report directory paths (dates like 2025-09-07_15-11-33)
	mux.Handle("/2", http.StripPrefix("/", fileServer))
	
	// Handle root path last (catch-all)
	mux.HandleFunc("/", s.HandleRoot)
	
	return mux
}

// Close cleans up server resources
func (s *Server) Close() error {
	if s.Storage != nil {
		return s.Storage.Close()
	}
	return nil
}
