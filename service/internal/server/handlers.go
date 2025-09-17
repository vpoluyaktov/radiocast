package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"radiocast/internal/logger"
	"radiocast/internal/reports"
)

// HandleRoot serves the main page with redirect to latest report
func (s *Server) HandleRoot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	ctx := r.Context()
	
	// Find latest report and redirect to it
	latestReportURL, err := s.findLatestReportURL(ctx)
	if err != nil {
		logger.Warn("No reports available", map[string]interface{}{"error": err.Error()})
		// show initial page
		s.serveInitialPage(w)
		return
	}
	
	// Redirect to the latest report with 302 status
	logger.Info("Redirecting to latest report", map[string]interface{}{"url": latestReportURL})
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
		logger.Error("Failed to load initial page template", err)
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
		logger.Warn("Report generation already in progress, rejecting new request")
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
	
	logger.Info("Starting report generation...")
	
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
		logger.Error("Report generation failed", err)
		http.Error(w, "Report generation failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	logger.Info("Report generation completed successfully")
	
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
	// Both local and GCS store files with "reports/" prefix in the unified structure
	actualFilePath := "reports/" + filePath
	
	fileData, err := s.Storage.GetFile(ctx, actualFilePath)
	if err != nil {
		logger.Error("Failed to get file from storage", err, map[string]interface{}{"path": actualFilePath})
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
	
	// List all files in reports directory recursively
	allFiles, err := s.Storage.ListDir(ctx, "reports", true)
	if err != nil {
		logger.Error("Failed to list reports", err)
		http.Error(w, "Failed to list reports: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Filter for index.html files and extract report paths
	var reports []string
	for _, file := range allFiles {
		if strings.HasSuffix(file, "/index.html") {
			reports = append(reports, file)
		}
	}
	
	// Sort and limit results (newest first - reverse alphabetical)
	sort.Strings(reports)
	for i, j := 0, len(reports)-1; i < j; i, j = i+1, j-1 {
		reports[i], reports[j] = reports[j], reports[i]
	}
	if limit > 0 && limit < len(reports) {
		reports = reports[:limit]
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
	// List all files in reports directory recursively
	allFiles, err := s.Storage.ListDir(ctx, "reports", true)
	if err != nil {
		return "", fmt.Errorf("failed to list reports: %w", err)
	}
	
	// Filter for index.html files
	var reports []string
	for _, file := range allFiles {
		if strings.HasSuffix(file, "/index.html") {
			reports = append(reports, file)
		}
	}
	
	if len(reports) == 0 {
		return "", fmt.Errorf("no reports available")
	}
	
	// Sort and get the latest (newest first - reverse alphabetical)
	sort.Strings(reports)
	for i, j := 0, len(reports)-1; i < j; i, j = i+1, j-1 {
		reports[i], reports[j] = reports[j], reports[i]
	}
	
	reportPath := reports[0]
	// Add leading slash (reportPath already includes "reports/" prefix)
	return "/" + reportPath, nil
}

// HandleHistory serves the history page
func (s *Server) HandleHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	w.Header().Set("Content-Type", "text/html")
	
	ctx := r.Context()
	
	// Load history page from storage
	historyContent, err := s.Storage.GetFile(ctx, "history/index.html")
	if err != nil {
		logger.Error("Failed to load history page", err)
		http.Error(w, "History page not found", http.StatusInternalServerError)
		return
	}
	
	w.Write(historyContent)
}

// HandleTheory serves the theory page
func (s *Server) HandleTheory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	w.Header().Set("Content-Type", "text/html")
	
	ctx := r.Context()
	
	// Load theory page from storage
	theoryContent, err := s.Storage.GetFile(ctx, "theory/index.html")
	if err != nil {
		logger.Error("Failed to load theory page", err)
		http.Error(w, "Theory page not found", http.StatusInternalServerError)
		return
	}
	
	w.Write(theoryContent)
}

// HandleStaticCSS serves the static CSS file for History and Theory pages
func (s *Server) HandleStaticCSS(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	w.Header().Set("Content-Type", "text/css")
	
	ctx := r.Context()
	
	// Try to get CSS from static assets storage path
	cssPath := "static/styles.css"
	cssContent, err := s.Storage.GetFile(ctx, cssPath)
	if err != nil {
		logger.Error("Failed to load CSS from storage", err, map[string]interface{}{"path": cssPath})
		http.Error(w, "CSS not found", http.StatusInternalServerError)
		return
	}
	
	w.Write(cssContent)
}

// HandleStaticBackground serves the background image for History and Theory pages
func (s *Server) HandleStaticBackground(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	w.Header().Set("Content-Type", "image/png")
	
	ctx := r.Context()
	
	// Try to get background image from static assets storage path
	imagePath := "static/background.png"
	imageContent, err := s.Storage.GetFile(ctx, imagePath)
	if err != nil {
		logger.Error("Failed to load background image from storage", err, map[string]interface{}{"path": imagePath})
		http.Error(w, "Background image not found", http.StatusInternalServerError)
		return
	}
	
	w.Write(imageContent)
}

// requireAPIKey is a middleware that validates API key for protected endpoints
func (s *Server) requireAPIKey(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// If no API key is configured, allow access (backward compatibility)
		if s.Config.RadiocastAPIKey == "" {
			logger.Debug("No API key configured, allowing access")
			next(w, r)
			return
		}
		
		// Extract API key from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			logger.Warn("Missing Authorization header for protected endpoint", map[string]interface{}{
				"endpoint": r.URL.Path,
				"method":   r.Method,
				"remote_addr": r.RemoteAddr,
			})
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			response := map[string]interface{}{
				"error":   "Missing Authorization header",
				"message": "This endpoint requires an API key. Please provide it in the Authorization header as 'Bearer <your-api-key>'",
				"status":  "unauthorized",
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		
		// Check if header starts with "Bearer "
		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			logger.Warn("Invalid Authorization header format", map[string]interface{}{
				"endpoint": r.URL.Path,
				"method":   r.Method,
				"remote_addr": r.RemoteAddr,
			})
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			response := map[string]interface{}{
				"error":   "Invalid Authorization header format",
				"message": "Authorization header must be in format 'Bearer <your-api-key>'",
				"status":  "unauthorized",
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		
		// Extract the API key
		providedKey := strings.TrimSpace(authHeader[len(bearerPrefix):])
		if providedKey == "" {
			logger.Warn("Empty API key provided", map[string]interface{}{
				"endpoint": r.URL.Path,
				"method":   r.Method,
				"remote_addr": r.RemoteAddr,
			})
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			response := map[string]interface{}{
				"error":   "Empty API key",
				"message": "API key cannot be empty",
				"status":  "unauthorized",
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		
		// Validate API key
		if providedKey != s.Config.RadiocastAPIKey {
			logger.Warn("Invalid API key provided", map[string]interface{}{
				"endpoint": r.URL.Path,
				"method":   r.Method,
				"remote_addr": r.RemoteAddr,
				"provided_key_length": len(providedKey),
			})
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			response := map[string]interface{}{
				"error":   "Invalid API key",
				"message": "The provided API key is not valid",
				"status":  "unauthorized",
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		
		// API key is valid, proceed to the next handler
		logger.Debug("Valid API key provided, allowing access", map[string]interface{}{
			"endpoint": r.URL.Path,
			"method":   r.Method,
		})
		next(w, r)
	}
}


