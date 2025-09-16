package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"sync"

	"radiocast/internal/config"
	"radiocast/internal/fetchers"
	"radiocast/internal/llm"
	"radiocast/internal/mocks"
	"radiocast/internal/reports"
	"radiocast/internal/storage"
)


// Server represents the HTTP server with all its dependencies
type Server struct {
	Config          *config.Config
	Fetcher         *fetchers.DataFetcher
	LLMClient       *llm.OpenAIClient
	MockService     *mocks.MockService
	ReportGenerator *reports.ReportGenerator
	Storage         storage.StorageClient
	DeploymentMode  storage.DeploymentMode
	
	// Mutex to prevent concurrent report generation
	generateMutex   sync.Mutex
}

// NewServer creates a new server instance
func NewServer(cfg *config.Config, deploymentMode storage.DeploymentMode) (*Server, error) {
	ctx := context.Background()
	
	server := &Server{
		Config:         cfg,
		Fetcher:        fetchers.NewDataFetcher(),
		LLMClient:      llm.NewOpenAIClient(cfg.OpenAIAPIKey, cfg.OpenAIModel),
		DeploymentMode: deploymentMode,
	}
	
	// Initialize mock service if mockup mode is enabled
	if cfg.MockupMode {
		mocksDir := filepath.Join("internal", "mocks")
		server.MockService = mocks.NewMockService(mocksDir)
		log.Printf("Mockup mode enabled - using mock data from %s", mocksDir)
	}
	
	// Initialize storage client using factory
	storageClient, err := storage.NewStorageClient(ctx, deploymentMode, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage client: %w", err)
	}
	server.Storage = storageClient
	
	// Initialize report generator
	server.ReportGenerator = reports.NewReportGenerator()
	
	// Log deployment mode
	if deploymentMode == storage.DeploymentLocal {
		log.Printf("Local deployment mode - reports directory determined by storage client")
	} else {
		log.Printf("GCS deployment mode - reports will be saved to GCS bucket: %s", cfg.GCSBucket)
	}
	
	
	return server, nil
}

// SetupRoutes configures HTTP routes for the server
func (s *Server) SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	
	
	// Handle specific API routes first
	mux.HandleFunc("/health", s.HandleHealth)
	mux.HandleFunc("/generate", s.HandleGenerate)
	mux.HandleFunc("/reports", s.HandleListReports)
	mux.HandleFunc("/reports/", s.HandleFileProxy)
	
	
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
