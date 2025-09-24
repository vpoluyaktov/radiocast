package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"radiocast/internal/config"
	"radiocast/internal/fetchers"
	"radiocast/internal/llm"
	"radiocast/internal/logger"
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
		logger.Infof("Mockup mode enabled - using mock data from %s", mocksDir)
	}
	
	// Initialize storage client using factory
	storageClient, err := storage.NewStorageClient(ctx, deploymentMode, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage client: %w", err)
	}
	server.Storage = storageClient
	
	// Initialize report generator
	server.ReportGenerator = reports.NewReportGenerator()
	
	// Initialize static assets
	if err := server.initializeStaticAssets(ctx); err != nil {
		logger.Infof("ERROR: Failed to initialize static assets: %v", err)
		logger.Infof("Static pages (/history, /theory, /static/*) may not work correctly")
	} else {
		logger.Debugf("Static assets initialized successfully")
	}
	
	// Log deployment mode
	if deploymentMode == storage.DeploymentLocal {
		logger.Debugf("Local deployment mode - reports directory determined by storage client")
	} else {
		logger.Debugf("GCS deployment mode - reports will be saved to GCS bucket: %s", cfg.GCSBucket)
	}
	
	
	return server, nil
}

// SetupRoutes configures HTTP routes for the server
func (s *Server) SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	
	
	// Handle specific API routes first
	mux.HandleFunc("/health", s.HandleHealth)
	mux.HandleFunc("/generate", s.requireAPIKey(s.HandleGenerate))
	mux.HandleFunc("/reports", s.HandleListReports)
	mux.HandleFunc("/reports/", s.HandleFileProxy)
	
	// Handle static pages
	mux.HandleFunc("/history", s.HandleHistory)
	mux.HandleFunc("/theory", s.HandleTheory)
	mux.HandleFunc("/about", s.HandleAbout)
	mux.HandleFunc("/static/", s.HandleStaticFiles)
	
	// Handle root path last (catch-all)
	mux.HandleFunc("/", s.HandleRoot)
	
	return mux
}

// initializeStaticAssets uploads static assets and HTML pages to storage
func (s *Server) initializeStaticAssets(ctx context.Context) error {
	staticDir := filepath.Join("internal", "static")
	templatesDir := filepath.Join("internal", "templates")
	
	// Store all files from static directory
	staticFiles, err := os.ReadDir(staticDir)
	if err != nil {
		return fmt.Errorf("failed to read static directory: %w", err)
	}
	
	for _, file := range staticFiles {
		if file.IsDir() {
			continue // Skip directories
		}
		
		filePath := filepath.Join(staticDir, file.Name())
		fileData, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read static file %s: %w", file.Name(), err)
		}
		
		if err := s.Storage.StoreFile(ctx, "static/"+file.Name(), fileData); err != nil {
			return fmt.Errorf("failed to store static file %s: %w", file.Name(), err)
		}
		
		logger.Debugf("Static file %s uploaded successfully", file.Name())
	}
	
	// Store history page
	historyPath := filepath.Join(templatesDir, "history_template.html")
	historyData, err := os.ReadFile(historyPath)
	if err != nil {
		return fmt.Errorf("failed to read history template: %w", err)
	}
	// Store as a regular file in history folder
	if err := s.storeHTMLPage(ctx, historyData, "history/index.html"); err != nil {
		return fmt.Errorf("failed to store history page: %w", err)
	}
	logger.Debugf("History page uploaded successfully")
	
	// Store theory page
	theoryPath := filepath.Join(templatesDir, "theory_template.html")
	theoryData, err := os.ReadFile(theoryPath)
	if err != nil {
		return fmt.Errorf("failed to read theory template: %w", err)
	}
	// Store as a regular file in theory folder
	if err := s.storeHTMLPage(ctx, theoryData, "theory/index.html"); err != nil {
		return fmt.Errorf("failed to store theory page: %w", err)
	}
	logger.Debugf("Theory page uploaded successfully")
	
	// Store about page
	aboutPath := filepath.Join(templatesDir, "about_template.html")
	aboutData, err := os.ReadFile(aboutPath)
	if err != nil {
		return fmt.Errorf("failed to read about template: %w", err)
	}
	// Store as a regular file in about folder
	if err := s.storeHTMLPage(ctx, aboutData, "about/index.html"); err != nil {
		return fmt.Errorf("failed to store about page: %w", err)
	}
	logger.Debugf("About page uploaded successfully")
	
	return nil
}

// storeHTMLPage stores an HTML page directly to storage
func (s *Server) storeHTMLPage(ctx context.Context, htmlData []byte, filePath string) error {
	// Use the unified storage interface for both local and GCS
	if err := s.Storage.StoreFile(ctx, filePath, htmlData); err != nil {
		return fmt.Errorf("failed to store HTML file %s: %w", filePath, err)
	}
	return nil
}

// Close cleans up server resources
func (s *Server) Close() error {
	if s.Storage != nil {
		return s.Storage.Close()
	}
	return nil
}
