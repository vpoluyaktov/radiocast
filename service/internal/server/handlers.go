package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"radiocast/internal/reports"
)

// HandleRoot serves the main page with redirect to latest report
func (s *Server) HandleRoot(w http.ResponseWriter, r *http.Request) {
	log.Printf("DEBUG: handleRoot called - method: %s, URL: %s", r.Method, r.URL.Path)
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	ctx := r.Context()
	
	// Find latest report and redirect to it
	latestReportURL, err := s.findLatestReportURL(ctx)
	if err != nil {
		log.Printf("No reports available: %v", err)
		// show initial page
		s.serveInitialPage(w)
		return
	}
	
	// Redirect to the latest report with 302 status
	log.Printf("Redirecting to latest report: %s", latestReportURL)
	w.Header().Set("Location", latestReportURL)
	w.WriteHeader(http.StatusFound) // 302 redirect
}

// serveInitialPage shows a loading page while report is being generated
func (s *Server) serveInitialPage(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>Radio Propagation Service</title>
	<meta http-equiv="refresh" content="60">
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background: #f5f5f5; text-align: center; }
        .container { max-width: 800px; margin: 0 auto; background: white; padding: 40px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #333; }
        .spinner { border: 4px solid #f3f3f3; border-top: 4px solid #007bff; border-radius: 50%%; width: 50px; height: 50px; animation: spin 1s linear infinite; margin: 20px auto; }
        @keyframes spin { 0%% { transform: rotate(0deg); } 100%% { transform: rotate(360deg); } }
        .status { background: #e3f2fd; padding: 20px; border-radius: 5px; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Radio Propagation Service</h1>
        <div class="spinner"></div>
        <div class="status">
            <h3>No reports available yet...</h3>
            <p>Please come back later.</p>
        </div>
        <p style="color: #666; margin-top: 30px;">
            For amateur radio operators worldwide | 73!
        </p>
    </div>
</body>
</html>`)
}

// HandleHealth provides health check endpoint
func (s *Server) HandleHealth(w http.ResponseWriter, r *http.Request) {
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

// HandleGenerate generates a new propagation report (HTTP handler)
func (s *Server) HandleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Try to acquire the mutex - if already locked, return error immediately
	if !s.generateMutex.TryLock() {
		log.Printf("Report generation already in progress, rejecting new request")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		response := map[string]interface{}{
			"error":   "Report generation already in progress",
			"message": "Another report generation is currently running. Please wait for it to complete before starting a new one.",
			"status":  "conflict",
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	
	// Ensure mutex is released when function exits
	defer s.generateMutex.Unlock()
	
	ctx := r.Context()
	
	log.Printf("Starting report generation...")
	
	// Generate new report
	storageOrchestrator := reports.NewStorageOrchestrator(s.Storage, string(s.DeploymentMode))
	deploymentModeStr := string(s.DeploymentMode)
	result, err := s.ReportGenerator.GenerateCompleteReport(
		ctx,
		s.Config,
		s.Fetcher,
		s.LLMClient,
		s.MockService,
		s.Storage,
		deploymentModeStr,
		storageOrchestrator,
	)
	if err != nil {
		log.Printf("Report generation failed: %v", err)
		http.Error(w, "Report generation failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	log.Printf("Report generation completed successfully")
	
	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

// HandleFileProxy serves files from local storage or GCS
func (s *Server) HandleFileProxy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Extract file path from URL (e.g., /files/2025-09-05_23-40-58/index.html)
	filePath := strings.TrimPrefix(r.URL.Path, "/files/")
	if filePath == "" {
		http.Error(w, "File path required", http.StatusBadRequest)
		return
	}
	
	ctx := r.Context()
	
	// In local mode, serve from local storage directly
	if s.DeploymentMode == "local" {
		// Security check: prevent directory traversal
		if strings.Contains(filePath, "..") {
			http.Error(w, "Invalid file path", http.StatusBadRequest)
			return
		}
		
		// Serve from local reports directory
		localPath := filepath.Join(s.ReportsDir, filePath)
		if _, err := os.Stat(localPath); os.IsNotExist(err) {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		
		http.ServeFile(w, r, localPath)
		return
	}
	
	// GCS mode - serve from GCS storage
	if s.DeploymentMode == "gcs" && s.Storage != nil {
		fileData, err := s.Storage.GetFile(ctx, filePath)
		if err != nil {
			log.Printf("Failed to get file from GCS: %v", err)
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		
		// Set appropriate content type based on file extension
		contentType := GetContentType(filePath)
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Cache-Control", "public, max-age=3600")
		w.Write(fileData)
		return
	}
	
	// If no storage is configured, return error
	http.Error(w, "Storage not configured", http.StatusInternalServerError)
}

// HandleListReports lists recent reports
func (s *Server) HandleListReports(w http.ResponseWriter, r *http.Request) {
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
	
	reports, err := s.Storage.ListReports(ctx, limit)
	if err != nil {
		log.Printf("Failed to list reports: %v", err)
		http.Error(w, "Failed to list reports: "+err.Error(), http.StatusInternalServerError)
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

// findLatestReportURL finds the URL of the latest report
func (s *Server) findLatestReportURL(ctx context.Context) (string, error) {
	if s.Storage != nil {
		// GCS mode - get latest report from storage
		reports, err := s.Storage.ListReports(ctx, 1)
		if err != nil || len(reports) == 0 {
			return "", fmt.Errorf("no reports available")
		}
		return fmt.Sprintf("/files/%s", reports[0]), nil
	}
	
	// Local mode - find latest timestamped directory
	entries, err := os.ReadDir(s.ReportsDir)
	if err != nil {
		return "", fmt.Errorf("failed to read reports directory: %w", err)
	}
	
	// Find the most recent directory (sorted by name which includes timestamp)
	var latestDir string
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() > latestDir {
			latestDir = entry.Name()
		}
	}
	
	if latestDir == "" {
		return "", fmt.Errorf("no report directories found")
	}
	
	return fmt.Sprintf("/%s/index.html", latestDir), nil
}

