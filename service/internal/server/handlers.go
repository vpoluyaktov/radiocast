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

// serveInitialPage shows an initial page if no reports are available
func (s *Server) serveInitialPage(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html")
	
	// Load template from file
	templatePath := filepath.Join("internal", "templates", "initial_page.html")
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		log.Printf("Failed to load initial page template: %v", err)
		// Fallback to simple error message
		fmt.Fprintf(w, "<html><body><h1>Service Unavailable</h1><p>Please try again later.</p></body></html>")
		return
	}
	
	w.Write(templateContent)
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
	
	// Extract file path from URL (e.g., /reports/2025-09-05_23-40-58/index.html)
	filePath := strings.TrimPrefix(r.URL.Path, "/reports/")
	if filePath == "" {
		http.Error(w, "File path required", http.StatusBadRequest)
		return
	}
	
	ctx := r.Context()
	
	// Security check: prevent directory traversal
	if strings.Contains(filePath, "..") {
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		return
	}
	
	// Use storage client to get file (works for both local and remote storage)
	fileData, err := s.Storage.GetFile(ctx, filePath)
	if err != nil {
		log.Printf("Failed to get file from storage: %v", err)
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	
	// Set appropriate content type
	contentType := GetContentType(filePath)
	w.Header().Set("Content-Type", contentType)
	
	// Write file data to response
	w.Write(fileData)
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
	// Use storage client to get latest report (works for both local and remote)
	reports, err := s.Storage.ListReports(ctx, 1)
	if err != nil || len(reports) == 0 {
		return "", fmt.Errorf("no reports available")
	}
	return fmt.Sprintf("/reports/%s", reports[0]), nil
}

