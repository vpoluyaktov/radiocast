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

	"radiocast/internal/models"
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
		// Auto-generate first report if none exists
		s.autoGenerateReport(w, r)
		return
	}
	
	// Redirect to the latest report with 302 status
	log.Printf("Redirecting to latest report: %s", latestReportURL)
	w.Header().Set("Location", latestReportURL)
	w.WriteHeader(http.StatusFound) // 302 redirect
}

// autoGenerateReport automatically generates a report when none exists
func (s *Server) autoGenerateReport(w http.ResponseWriter, _ *http.Request) {
	log.Println("No reports found - automatically generating first report...")
	
	// Show loading page while generating
	s.serveLoadingPage(w)
	
	// Generate report in background
	go func() {
		ctx := context.Background()
		
		var data *models.PropagationData
		var sourceData *models.SourceData
		var markdownReport string
		var err error
		
		if s.Config.MockupMode && s.MockService != nil {
			// Use mock data
			log.Println("Auto-generation: Loading mock data...")
			data, sourceData, err = s.MockService.LoadMockData()
			if err != nil {
				log.Printf("Auto-generation: Mock data loading failed: %v", err)
				return
			}
			
			log.Println("Auto-generation: Loading mock LLM response...")
			markdownReport, err = s.MockService.LoadMockLLMResponse()
			if err != nil {
				log.Printf("Auto-generation: Mock LLM response loading failed: %v", err)
				return
			}
		} else {
			// Fetch data from all sources
			log.Println("Auto-generation: Fetching data from all sources...")
			data, sourceData, err = s.Fetcher.FetchAllDataWithSources(ctx, s.Config.NOAAKIndexURL, s.Config.NOAASolarURL, s.Config.N0NBHSolarURL, s.Config.SIDCRSSURL)
			if err != nil {
				log.Printf("Auto-generation: Data fetching failed: %v", err)
				return
			}
			
			// Generate LLM report
			log.Println("Auto-generation: Generating LLM report...")
			markdownReport, err = s.LLMClient.GenerateReportWithSources(data, sourceData)
			if err != nil {
				log.Printf("Auto-generation: LLM report generation failed: %v", err)
				return
			}
		}
		
		// Generate all files
		fileManager := NewFileManager(s)
		files, err := fileManager.GenerateAllFiles(ctx, data, sourceData, markdownReport)
		if err != nil {
			log.Printf("Auto-generation: File generation failed: %v", err)
			return
		}
		
		// Handle storage
		if s.DeploymentMode == DeploymentGCS && s.Storage != nil {
			_, err = fileManager.UploadToGCS(ctx, files, data.Timestamp)
			if err != nil {
				log.Printf("Auto-generation: GCS upload failed: %v", err)
				fileManager.Cleanup(files)
				return
			}
			fileManager.Cleanup(files)
		}
		
		log.Println("Auto-generation: Report generated successfully")
	}()
}

// serveLoadingPage shows a loading page while report is being generated
func (s *Server) serveLoadingPage(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>Radio Propagation Service - Generating Report</title>
    <meta http-equiv="refresh" content="10">
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
        <h1>📡 Radio Propagation Service</h1>
        <div class="spinner"></div>
        <div class="status">
            <h3>Generating Your First Report...</h3>
            <p>Please wait while we fetch the latest propagation data and generate your report.</p>
            <p>This page will automatically refresh in 10 seconds.</p>
            <p><strong>Status:</strong> Fetching data from NOAA, N0NBH, and SIDC...</p>
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

// HandleGenerate generates a new propagation report
func (s *Server) HandleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	ctx := r.Context()
	log.Println("Starting report generation...")

	var data *models.PropagationData
	var sourceData *models.SourceData
	var markdownReport string
	var err error

	if s.Config.MockupMode && s.MockService != nil {
		// Use mock data
		log.Println("Using mock data for report generation...")
		data, sourceData, err = s.MockService.LoadMockData()
		if err != nil {
			log.Printf("Mock data loading failed: %v", err)
			http.Error(w, "Mock data loading failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		log.Println("Loading mock LLM response...")
		markdownReport, err = s.MockService.LoadMockLLMResponse()
		if err != nil {
			log.Printf("Mock LLM response loading failed: %v", err)
			http.Error(w, "Mock LLM response loading failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		
		log.Printf("Mock data loaded successfully for timestamp: %s", data.Timestamp.Format(time.RFC3339))
		log.Printf("Mock LLM report loaded successfully (length: %d characters)", len(markdownReport))
	} else {
		// Fetch data from all sources
		log.Println("Fetching data from all sources...")
		data, sourceData, err = s.Fetcher.FetchAllDataWithSources(ctx, s.Config.NOAAKIndexURL, s.Config.NOAASolarURL, s.Config.N0NBHSolarURL, s.Config.SIDCRSSURL)
		if err != nil {
			log.Printf("Data fetching failed: %v", err)
			http.Error(w, "Data fetching failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("Data fetched successfully for timestamp: %s", data.Timestamp.Format(time.RFC3339))

		// Generate LLM report with raw source data
		log.Println("Generating LLM report with raw source data...")
		markdownReport, err = s.LLMClient.GenerateReportWithSources(data, sourceData)
		if err != nil {
			log.Printf("LLM report generation failed: %v", err)
			http.Error(w, "LLM report generation failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("LLM report generated successfully (length: %d characters)", len(markdownReport))
	}

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
	
	// Extract file path from URL (e.g., /2025-09-05_23-40-58/index.html)
	filePath := strings.TrimPrefix(r.URL.Path, "/")
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
			contentType := s.getContentType(filePath)
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
	
	// Serve from local reports directory
	localPath := filepath.Join(s.ReportsDir, filePath)
	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
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

// findLatestReportURL finds the URL of the latest report
func (s *Server) findLatestReportURL(ctx context.Context) (string, error) {
	if s.Storage != nil {
		// GCS mode - get latest report from storage
		reports, err := s.Storage.ListReports(ctx, 1)
		if err != nil || len(reports) == 0 {
			return "", fmt.Errorf("no reports available")
		}
		return fmt.Sprintf("/files/%s/index.html", reports[0]), nil
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

// getContentType returns the appropriate content type for a file
func (s *Server) getContentType(filePath string) string {
	if strings.HasSuffix(filePath, ".html") {
		return "text/html"
	} else if strings.HasSuffix(filePath, ".png") {
		return "image/png"
	} else if strings.HasSuffix(filePath, ".gif") {
		return "image/gif"
	} else if strings.HasSuffix(filePath, ".json") {
		return "application/json"
	} else if strings.HasSuffix(filePath, ".txt") {
		return "text/plain"
	} else if strings.HasSuffix(filePath, ".md") {
		return "text/markdown"
	} else if strings.HasSuffix(filePath, ".css") {
		return "text/css"
	} else if strings.HasSuffix(filePath, ".js") {
		return "application/javascript"
	}
	return "application/octet-stream"
}
