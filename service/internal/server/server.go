package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
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
	
	// Initialize static assets
	if err := server.initializeStaticAssets(ctx); err != nil {
		log.Printf("ERROR: Failed to initialize static assets: %v", err)
		log.Printf("Static pages (/history, /theory, /static/*) may not work correctly")
	} else {
		log.Printf("Static assets initialized successfully")
	}
	
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
	
	// Handle static pages
	mux.HandleFunc("/history", s.HandleHistory)
	mux.HandleFunc("/theory", s.HandleTheory)
	mux.HandleFunc("/static/styles.css", s.HandleStaticCSS)
	mux.HandleFunc("/static/background.png", s.HandleStaticBackground)
	
	// Handle root path last (catch-all)
	mux.HandleFunc("/", s.HandleRoot)
	
	return mux
}

// initializeStaticAssets uploads static assets and HTML pages to storage
func (s *Server) initializeStaticAssets(ctx context.Context) error {
	staticDir := filepath.Join("internal", "static")
	templatesDir := filepath.Join("internal", "templates")
	
	// Store CSS file
	cssPath := filepath.Join(staticDir, "styles.css")
	cssData, err := os.ReadFile(cssPath)
	if err != nil {
		return fmt.Errorf("failed to read CSS file: %w", err)
	}
	if err := s.Storage.StoreFile(ctx, "static/styles.css", cssData); err != nil {
		return fmt.Errorf("failed to store CSS file: %w", err)
	}
	log.Printf("Static CSS file uploaded successfully")
	
	// Store background image
	bgPath := filepath.Join(staticDir, "background.png")
	bgData, err := os.ReadFile(bgPath)
	if err != nil {
		return fmt.Errorf("failed to read background image: %w", err)
	}
	if err := s.Storage.StoreFile(ctx, "static/background.png", bgData); err != nil {
		return fmt.Errorf("failed to store background image: %w", err)
	}
	log.Printf("Static background image uploaded successfully")
	
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
	log.Printf("History page uploaded successfully")
	
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
	log.Printf("Theory page uploaded successfully")
	
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
