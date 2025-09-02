package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// HandleRoot serves the main page with a generated report
func (s *Server) HandleRoot(w http.ResponseWriter, r *http.Request) {
	log.Printf("DEBUG: handleRoot called - method: %s, URL: %s", r.Method, r.URL.Path)
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Try to get the latest report from storage
	ctx := r.Context()
	if s.Storage != nil {
		reports, err := s.Storage.ListReports(ctx, 1)
		if err != nil || len(reports) == 0 {
			log.Printf("No reports available: %v", err)
			// Fall back to main page if no report available
			s.serveMainPage(w)
			return
		}
		
		// Serve the latest report
		reportContent, err := s.Storage.GetReport(ctx, reports[0])
		if err != nil {
			log.Printf("Failed to get report: %v", err)
			s.serveMainPage(w)
			return
		}
		
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(reportContent))
		return
	}
	
	// Local mode - generate report on demand
	log.Println("Local mode: generating report on demand...")
	
	// Fetch data from all sources
	log.Println("Starting data fetch from all sources...")
	data, sourceData, err := s.Fetcher.FetchAllDataWithSources(
		ctx,
		s.Config.NOAAKIndexURL,
		s.Config.NOAASolarURL,
		s.Config.N0NBHSolarURL,
		s.Config.SIDCRSSURL,
	)
	if err != nil {
		log.Printf("Data fetch failed: %v", err)
		s.serveMainPage(w)
		return
	}
	log.Printf("Data fetch and normalization completed successfully")
	
	// Generate LLM report with raw source data
	log.Println("Generating LLM report with raw source data...")
	markdownReport, err := s.LLMClient.GenerateReportWithSources(data, sourceData)
	if err != nil {
		log.Printf("LLM report generation failed: %v", err)
		s.serveMainPage(w)
		return
	}
	
	// Use unified file manager to generate all files
	fileManager := NewFileManager(s)
	files, err := fileManager.GenerateAllFiles(ctx, data, sourceData, markdownReport)
	if err != nil {
		log.Printf("File generation failed: %v", err)
		s.serveMainPage(w)
		return
	}
	
	// Serve the generated HTML report
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(files.HTMLContent))
}

// serveMainPage serves the main service information page
func (s *Server) serveMainPage(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>Radio Propagation Service</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background: #f5f5f5; }
        .container { max-width: 800px; margin: 0 auto; background: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #333; text-align: center; }
        .status { background: #e8f5e8; padding: 15px; border-radius: 5px; margin: 20px 0; }
        .endpoints { background: #f8f9fa; padding: 20px; border-radius: 5px; margin: 20px 0; }
        .endpoint { margin: 10px 0; }
        a { color: #007bff; text-decoration: none; }
        a:hover { text-decoration: underline; }
        .note { background: #fff3cd; padding: 15px; border-radius: 5px; margin: 20px 0; border-left: 4px solid #ffc107; }
    </style>
</head>
<body>
    <div class="container">
        <h1>📡 Radio Propagation Service</h1>
        <div class="note">
            <strong>Note:</strong> No propagation reports have been generated yet. Generate your first report using the /generate endpoint.
        </div>
        <div class="status">
            <strong>Status:</strong> Service is running and ready to generate propagation reports.
        </div>
        <div class="endpoints">
            <h3>Available Endpoints:</h3>
            <div class="endpoint"><strong>GET /health</strong> - Service health check</div>
            <div class="endpoint"><strong>POST /generate</strong> - Generate new propagation report</div>
            <div class="endpoint"><strong>GET /reports</strong> - List recent reports</div>
        </div>
        <p style="text-align: center; color: #666; margin-top: 30px;">
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

// HandleGenerate generates a new propagation report
func (s *Server) HandleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	ctx := r.Context()
	log.Println("Starting report generation...")

	// Fetch data from all sources
	log.Println("Fetching data from all sources...")
	data, sourceData, err := s.Fetcher.FetchAllDataWithSources(ctx, s.Config.NOAAKIndexURL, s.Config.NOAASolarURL, s.Config.N0NBHSolarURL, s.Config.SIDCRSSURL)
	if err != nil {
		log.Printf("Data fetching failed: %v", err)
		http.Error(w, "Data fetching failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Data fetched successfully for timestamp: %s", data.Timestamp.Format(time.RFC3339))

	// Generate LLM report with raw source data
	log.Println("Generating LLM report with raw source data...")
	markdownReport, err := s.LLMClient.GenerateReportWithSources(data, sourceData)
	if err != nil {
		log.Printf("LLM report generation failed: %v", err)
		http.Error(w, "LLM report generation failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("LLM report generated successfully (length: %d characters)", len(markdownReport))

	// Use unified file manager to generate all files
	fileManager := NewFileManager(s)
	files, err := fileManager.GenerateAllFiles(ctx, data, sourceData, markdownReport)
	if err != nil {
		log.Printf("File generation failed: %v", err)
		http.Error(w, "File generation failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Handle storage based on deployment mode
	var reportURL string
	if s.DeploymentMode == DeploymentGCS && s.Storage != nil {
		// Upload all files to GCS
		log.Println("Uploading all files to GCS...")
		reportURL, err = fileManager.UploadToGCS(ctx, files, data.Timestamp)
		if err != nil {
			log.Printf("Storage failed: %v", err)
			http.Error(w, "Failed to upload files to GCS: "+err.Error(), http.StatusInternalServerError)
			fileManager.Cleanup(files)
			return
		}
		log.Printf("All files uploaded to GCS successfully: %s", reportURL)
		// Clean up temporary files
		fileManager.Cleanup(files)
	} else {
		// Local mode - files are already saved
		reportURL = "/files/index.html"
		log.Printf("Report stored locally in: %s", files.ReportDir)
	}
	
	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "success",
		"message":    "Report generated successfully",
		"reportURL":  reportURL,
		"timestamp":  data.Timestamp.Format(time.RFC3339),
		"dataPoints": len(data.SourceEvents),
		"folderPath": files.FolderPath,
	})
}

// HandleFileProxy serves files from local storage or GCS
func (s *Server) HandleFileProxy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Extract file path from URL
	filePath := strings.TrimPrefix(r.URL.Path, "/files/")
	if filePath == "" {
		http.Error(w, "File path required", http.StatusBadRequest)
		return
	}
	
	ctx := r.Context()
	
	// Try to serve from GCS first if available
	if s.Storage != nil {
		fileData, err := s.Storage.GetFile(ctx, filePath)
		if err == nil {
			// Set appropriate content type based on file extension
			contentType := "application/octet-stream"
			if strings.HasSuffix(filePath, ".html") {
				contentType = "text/html"
			} else if strings.HasSuffix(filePath, ".png") {
				contentType = "image/png"
			} else if strings.HasSuffix(filePath, ".json") {
				contentType = "application/json"
			} else if strings.HasSuffix(filePath, ".txt") {
				contentType = "text/plain"
			} else if strings.HasSuffix(filePath, ".md") {
				contentType = "text/markdown"
			}
			
			w.Header().Set("Content-Type", contentType)
			w.Header().Set("Cache-Control", "public, max-age=3600")
			w.Write(fileData)
			return
		}
		log.Printf("Failed to get file from GCS: %v", err)
	}
	
	// Fall back to local file serving
	// Security check: prevent directory traversal
	if strings.Contains(filePath, "..") {
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		return
	}
	
	// Try to find the file in the most recent report directory
	localPath := filepath.Join(s.ReportsDir, filePath)
	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		// If not found directly, try to find in the latest report directory
		entries, err := os.ReadDir(s.ReportsDir)
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		
		// Find the most recent directory
		var latestDir string
		for _, entry := range entries {
			if entry.IsDir() && entry.Name() > latestDir {
				latestDir = entry.Name()
			}
		}
		
		if latestDir != "" {
			localPath = filepath.Join(s.ReportsDir, latestDir, filepath.Base(filePath))
		}
	}
	
	http.ServeFile(w, r, localPath)
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
