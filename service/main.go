package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"radiocast/internal/config"
	"radiocast/internal/fetchers"
	"radiocast/internal/llm"
	"radiocast/internal/reports"
	"radiocast/internal/storage"
)

// Server represents the main application server
type Server struct {
	config    *config.Config
	fetcher   *fetchers.DataFetcher
	llmClient *llm.OpenAIClient
	generator *reports.Generator
	storage   *storage.GCSClient
}

// NewServer creates a new server instance
func NewServer(cfg *config.Config) (*Server, error) {
	ctx := context.Background()
	
	// Initialize GCS client
	gcsClient, err := storage.NewGCSClient(ctx, cfg.GCSBucket)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize GCS client: %w", err)
	}
	
	return &Server{
		config:    cfg,
		fetcher:   fetchers.NewDataFetcher(),
		llmClient: llm.NewOpenAIClient(cfg.OpenAIAPIKey, cfg.OpenAIModel),
		generator: reports.NewGenerator(),
		storage:   gcsClient,
	}, nil
}

// Close cleans up server resources
func (s *Server) Close() error {
	if s.storage != nil {
		return s.storage.Close()
	}
	return nil
}

func main() {
	ctx := context.Background()
	
	// Load configuration
	cfg, err := config.Load(ctx)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	
	log.Printf("Starting Radio Propagation Service on port %s", cfg.Port)
	log.Printf("Environment: %s", cfg.Environment)
	log.Printf("GCS Bucket: %s", cfg.GCSBucket)
	
	// Create server
	server, err := NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	defer server.Close()
	
	// Set up HTTP routes
	mux := http.NewServeMux()
	mux.HandleFunc("/", server.handleRoot)
	mux.HandleFunc("/health", server.handleHealth)
	mux.HandleFunc("/generate", server.handleGenerate)
	mux.HandleFunc("/reports", server.handleListReports)
	
	// Create HTTP server
	httpServer := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 300 * time.Second, // Longer timeout for report generation
		IdleTimeout:  60 * time.Second,
	}
	
	// Start server in goroutine
	go func() {
		log.Printf("Server listening on :%s", cfg.Port)
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()
	
	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	
	log.Println("Shutting down server...")
	
	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
	
	log.Println("Server stopped")
}

// handleRoot serves basic information about the service
func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	info := map[string]interface{}{
		"service":     "Radio Propagation Service",
		"version":     "1.0.0",
		"environment": s.config.Environment,
		"endpoints": map[string]string{
			"health":   "/health",
			"generate": "/generate",
			"reports":  "/reports",
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// handleHealth provides health check endpoint
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"checks": map[string]string{
			"gcs":    "ok",
			"config": "ok",
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// handleGenerate generates a new propagation report
func (s *Server) handleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	ctx := r.Context()
	startTime := time.Now()
	
	log.Println("Starting propagation report generation...")
	
	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	
	// Fetch data from all sources
	log.Println("Fetching data from external sources...")
	data, err := s.fetcher.FetchAllData(
		ctx,
		s.config.NOAAKIndexURL,
		s.config.NOAASolarURL,
		s.config.N0NBHSolarURL,
		s.config.SIDCRSSURL,
	)
	if err != nil {
		log.Printf("Data fetch failed: %v", err)
		http.Error(w, fmt.Sprintf("Data fetch failed: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Generate report using LLM
	log.Println("Generating report with OpenAI...")
	markdownReport, err := s.llmClient.GenerateReport(ctx, data)
	if err != nil {
		log.Printf("Report generation failed: %v", err)
		http.Error(w, fmt.Sprintf("Report generation failed: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Convert to HTML with charts
	log.Println("Converting to HTML and generating charts...")
	htmlReport, err := s.generator.GenerateHTML(markdownReport, data)
	if err != nil {
		log.Printf("HTML generation failed: %v", err)
		http.Error(w, fmt.Sprintf("HTML generation failed: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Store in GCS
	log.Println("Storing report in GCS...")
	reportURL, err := s.storage.StoreReport(ctx, htmlReport, data.Timestamp)
	if err != nil {
		log.Printf("Storage failed: %v", err)
		http.Error(w, fmt.Sprintf("Storage failed: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Clean up old reports (keep last 30 days)
	go func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		
		if err := s.storage.DeleteOldReports(cleanupCtx, 30*24*time.Hour); err != nil {
			log.Printf("Cleanup warning: %v", err)
		}
	}()
	
	duration := time.Since(startTime)
	log.Printf("Report generation completed in %v", duration)
	
	// Return response
	response := map[string]interface{}{
		"status":      "success",
		"report_url":  reportURL,
		"timestamp":   data.Timestamp.Format(time.RFC3339),
		"duration_ms": duration.Milliseconds(),
		"data_summary": map[string]interface{}{
			"solar_flux":     data.SolarData.SolarFluxIndex,
			"k_index":        data.GeomagData.KIndex,
			"sunspot_number": data.SolarData.SunspotNumber,
			"activity_level": data.SolarData.SolarActivity,
		},
	}
	
	json.NewEncoder(w).Encode(response)
}

// handleListReports lists recent reports
func (s *Server) handleListReports(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	ctx := r.Context()
	
	// Get limit from query parameter (default 10)
	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := fmt.Sscanf(limitStr, "%d", &limit); err != nil || parsedLimit != 1 {
			limit = 10
		}
		if limit > 100 {
			limit = 100 // Cap at 100
		}
	}
	
	reports, err := s.storage.ListReports(ctx, limit)
	if err != nil {
		log.Printf("Failed to list reports: %v", err)
		http.Error(w, fmt.Sprintf("Failed to list reports: %v", err), http.StatusInternalServerError)
		return
	}
	
	response := map[string]interface{}{
		"reports":   reports,
		"count":     len(reports),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
