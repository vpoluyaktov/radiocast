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
	"radiocast/internal/reports"
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
	
	// Create timestamped directory for this report
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	reportDir := filepath.Join(s.ReportsDir, timestamp)
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		log.Printf("Failed to create report directory: %v", err)
		s.serveMainPage(w)
		return
	}
	
	// Fetch data from all sources
	log.Println("Starting data fetch from all sources...")
	data, err := s.Fetcher.FetchAllData(
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
	
	// Save API data as JSON
	apiDataJSON, _ := json.MarshalIndent(data, "", "  ")
	apiDataPath := filepath.Join(reportDir, "01_api_data.json")
	if err := os.WriteFile(apiDataPath, apiDataJSON, 0644); err != nil {
		log.Printf("Failed to save API data: %v", err)
	}
	
	// Save system prompt
	systemPrompt := s.LLMClient.GetSystemPrompt()
	log.Printf("DEBUG: System prompt length: %d", len(systemPrompt))
	systemPromptPath := filepath.Join(reportDir, "llm_system_prompt.txt")
	log.Printf("DEBUG: Writing system prompt to: %s", systemPromptPath)
	if err := os.WriteFile(systemPromptPath, []byte(systemPrompt), 0644); err != nil {
		log.Printf("Failed to save system prompt: %v", err)
	} else {
		log.Printf("System prompt saved successfully to: %s", systemPromptPath)
	}
	
	// Generate LLM prompt and save it
	llmPrompt := s.LLMClient.BuildPrompt(data)
	promptPath := filepath.Join(reportDir, "02_llm_prompt.txt")
	if err := os.WriteFile(promptPath, []byte(llmPrompt), 0644); err != nil {
		log.Printf("Failed to save LLM prompt: %v", err)
	}

	log.Printf("Generating report for %s", time.Now().Format("2006-01-02"))
	markdown, err := s.LLMClient.GenerateReport(data)
	if err != nil {
		log.Printf("Failed to generate report: %v", err)
		http.Error(w, "Failed to generate report", http.StatusInternalServerError)
		return
	}
	log.Printf("Generated report with %d characters", len(markdown))

	markdownPath := filepath.Join(reportDir, "03_llm_response.md")
	if err := os.WriteFile(markdownPath, []byte(markdown), 0644); err != nil {
		log.Printf("Failed to save markdown report: %v", err)
	}

	log.Printf("Converting markdown to HTML and generating charts...")
	
	var html string
	var reportPath string
	
	// For production/staging, store in GCS and upload charts
	log.Printf("DEBUG: Storage client status: %v", s.Storage != nil)
	if s.Storage != nil {
		log.Printf("DEBUG: Entering GCS deployment mode with PNG chart generation")
		// Generate charts first
		chartGen := reports.NewChartGenerator(reportDir)
		log.Printf("Generating PNG charts in directory: %s", reportDir)
		chartFiles, err := chartGen.GenerateCharts(data)
		if err != nil {
			log.Printf("Warning: Failed to generate charts: %v", err)
			chartFiles = []string{}
		} else {
			log.Printf("Successfully generated %d chart files: %v", len(chartFiles), chartFiles)
		}
		
		// Upload chart images to GCS
		timestamp := time.Now()
		
		// Generate folder path using the same logic as StoreChartImage
		folderPath := fmt.Sprintf("%04d/%02d/%02d/PropagationReport-%04d-%02d-%02d-%02d-%02d-%02d",
			timestamp.Year(), timestamp.Month(), timestamp.Day(),
			timestamp.Year(), timestamp.Month(), timestamp.Day(),
			timestamp.Hour(), timestamp.Minute(), timestamp.Second())
		
		log.Printf("Using folder path for charts: %s", folderPath)
		
		uploadedCharts := []string{}
		log.Printf("Starting chart upload process for %d files", len(chartFiles))
		for _, chartFile := range chartFiles {
			log.Printf("Attempting to read chart file: %s", chartFile)
			imageData, err := os.ReadFile(chartFile)
			if err != nil {
				log.Printf("Failed to read chart file %s: %v", chartFile, err)
				continue
			}
			
			filename := filepath.Base(chartFile)
			log.Printf("Uploading chart image %s (%d bytes) to GCS", filename, len(imageData))
			publicURL, err := s.Storage.StoreChartImage(ctx, imageData, filename, timestamp)
			if err != nil {
				log.Printf("Failed to store chart image %s: %v", filename, err)
				continue
			}
			
			// Keep track of successfully uploaded charts
			uploadedCharts = append(uploadedCharts, filename)
			log.Printf("Chart image uploaded successfully: %s", publicURL)
		}
		log.Printf("Chart upload completed. Successfully uploaded %d out of %d charts", len(uploadedCharts), len(chartFiles))
		
		// Generate HTML with proper folder path for chart proxy URLs
		html, err = s.generateHTMLWithCharts(markdown, data, uploadedCharts, folderPath)
		if err != nil {
			log.Printf("Failed to generate HTML: %v", err)
			http.Error(w, "Failed to generate HTML", http.StatusInternalServerError)
			return
		}
		
		// Store HTML report in GCS
		reportPath, err = s.Storage.StoreReport(ctx, html, timestamp)
		if err != nil {
			log.Printf("Failed to store report: %v", err)
			http.Error(w, "Failed to store report", http.StatusInternalServerError)
			return
		}
		
		log.Printf("Report stored in GCS at: %s", reportPath)
	} else {
		// For local mode, use existing logic
		html, err = s.Generator.GenerateHTML(markdown, data)
		if err != nil {
			log.Printf("Failed to generate HTML: %v", err)
			http.Error(w, "Failed to generate HTML", http.StatusInternalServerError)
			return
		}
	}
	
	log.Printf("Generated complete HTML report with %d characters", len(html))

	htmlPath := filepath.Join(reportDir, "04_final_report.html")
	if err := os.WriteFile(htmlPath, []byte(html), 0644); err != nil {
		log.Printf("Failed to save HTML report: %v", err)
	}
	
	log.Printf("Report saved to directory: %s", reportDir)
	
	// Serve the generated report
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
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
	startTime := time.Now()
	
	log.Println("Starting propagation report generation...")
	
	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	
	// Fetch data from all sources
	log.Println("Fetching data from external sources...")
	data, err := s.Fetcher.FetchAllData(
		ctx,
		s.Config.NOAAKIndexURL,
		s.Config.NOAASolarURL,
		s.Config.N0NBHSolarURL,
		s.Config.SIDCRSSURL,
	)
	if err != nil {
		log.Printf("Data fetch failed: %v", err)
		http.Error(w, fmt.Sprintf("Data fetch failed: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Generate report using LLM
	log.Println("Generating report with OpenAI...")
	markdownReport, err := s.LLMClient.GenerateReport(data)
	if err != nil {
		log.Printf("Report generation failed: %v", err)
		http.Error(w, fmt.Sprintf("Report generation failed: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Generate PNG charts for GCS deployment
	log.Printf("DEBUG: Storage client status: %v", s.Storage != nil)
	var uploadedCharts []string
	if s.Storage != nil {
		log.Printf("DEBUG: Entering GCS deployment mode with PNG chart generation")
		
		// Create temporary directory for chart generation
		reportDir := filepath.Join(os.TempDir(), fmt.Sprintf("report_%d", time.Now().Unix()))
		if err := os.MkdirAll(reportDir, 0755); err != nil {
			log.Printf("Failed to create report directory: %v", err)
		} else {
			defer os.RemoveAll(reportDir)
			
			// Generate PNG charts
			chartGen := reports.NewChartGenerator(reportDir)
			log.Printf("Generating PNG charts in directory: %s", reportDir)
			chartFiles, err := chartGen.GenerateCharts(data)
			if err != nil {
				log.Printf("Warning: Failed to generate charts: %v", err)
				chartFiles = []string{}
			} else {
				log.Printf("Successfully generated %d chart files: %v", len(chartFiles), chartFiles)
			}
			
			// Upload chart images to GCS
			timestamp := time.Now()
			folderPath := fmt.Sprintf("%04d/%02d/%02d/PropagationReport-%04d-%02d-%02d-%02d-%02d-%02d",
				timestamp.Year(), timestamp.Month(), timestamp.Day(),
				timestamp.Year(), timestamp.Month(), timestamp.Day(),
				timestamp.Hour(), timestamp.Minute(), timestamp.Second())
			
			log.Printf("Using folder path for charts: %s", folderPath)
			
			for _, chartFile := range chartFiles {
				log.Printf("Reading chart file: %s", chartFile)
				imageData, err := os.ReadFile(chartFile)
				if err != nil {
					log.Printf("Failed to read chart file %s: %v", chartFile, err)
					continue
				}
				
				filename := filepath.Base(chartFile)
				log.Printf("Uploading chart image: %s (size: %d bytes)", filename, len(imageData))
				publicURL, err := s.Storage.StoreChartImage(ctx, imageData, filename, timestamp)
				if err != nil {
					log.Printf("Failed to store chart image %s: %v", filename, err)
					continue
				}
				
				log.Printf("Chart image uploaded successfully: %s", publicURL)
				uploadedCharts = append(uploadedCharts, publicURL)
			}
			
			log.Printf("Total charts uploaded: %d", len(uploadedCharts))
		}
	}
	
	// Convert to HTML with charts
	log.Println("Converting to HTML and generating charts...")
	var htmlReport string
	if s.Storage != nil && len(uploadedCharts) > 0 {
		// Use GCS-uploaded chart URLs for HTML generation
		log.Printf("Using %d uploaded chart URLs for HTML generation", len(uploadedCharts))
		htmlReport, err = s.Generator.GenerateHTMLWithChartURLs(markdownReport, data, uploadedCharts)
	} else {
		// Fallback to local chart generation
		htmlReport, err = s.Generator.GenerateHTML(markdownReport, data)
	}
	if err != nil {
		log.Printf("HTML generation failed: %v", err)
		http.Error(w, fmt.Sprintf("HTML generation failed: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Store in GCS
	log.Println("Storing report in GCS...")
	reportURL, err := s.Storage.StoreReport(ctx, htmlReport, data.Timestamp)
	if err != nil {
		log.Printf("Storage failed: %v", err)
		http.Error(w, fmt.Sprintf("Storage failed: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Clean up old reports (keep last 30 days)
	go func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		
		if err := s.Storage.DeleteOldReports(cleanupCtx, 30*24*time.Hour); err != nil {
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

// HandleFileProxy serves any file from report folders through the service
func (s *Server) HandleFileProxy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Extract the file path from URL: /files/{folder}/{filename}
	path := strings.TrimPrefix(r.URL.Path, "/files/")
	if path == "" {
		http.Error(w, "File path required", http.StatusBadRequest)
		return
	}
	
	ctx := r.Context()
	
	// For local mode, serve from local filesystem
	if s.DeploymentMode == DeploymentLocal {
		// Serve from local reports directory
		localPath := filepath.Join(s.ReportsDir, path)
		
		// Security check - ensure path doesn't escape reports directory
		absReportsDir, _ := filepath.Abs(s.ReportsDir)
		absPath, _ := filepath.Abs(localPath)
		if !strings.HasPrefix(absPath, absReportsDir) {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}
		
		// Check if file exists and serve it
		if _, err := os.Stat(localPath); os.IsNotExist(err) {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		
		// Set content type based on file extension
		contentType := s.getContentType(filepath.Ext(localPath))
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Cache-Control", "public, max-age=3600")
		http.ServeFile(w, r, localPath)
		return
	}
	
	// For GCS mode, proxy from GCS
	if s.Storage == nil {
		http.Error(w, "Storage not configured", http.StatusInternalServerError)
		return
	}
	
	// Get file from GCS
	fileData, err := s.Storage.GetFile(ctx, path)
	if err != nil {
		log.Printf("Failed to get file %s: %v", path, err)
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	
	// Set content type based on file extension
	contentType := s.getContentType(filepath.Ext(path))
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Write(fileData)
}

// generateHTMLWithCharts generates HTML with charts using the specified folder path
func (s *Server) generateHTMLWithCharts(markdown string, data *models.PropagationData, chartFiles []string, folderPath string) (string, error) {
	// Convert markdown to HTML
	htmlContent := s.Generator.MarkdownToHTML(markdown)
	
	// Build chart HTML references with folder path
	chartsHTML := s.Generator.BuildChartsHTML(chartFiles, folderPath)
	
	// Combine everything into a complete HTML document
	fullHTML, err := s.Generator.BuildCompleteHTML(htmlContent, chartsHTML, data)
	if err != nil {
		return "", fmt.Errorf("failed to build complete HTML: %w", err)
	}
	
	return fullHTML, nil
}

// getContentType returns the appropriate content type for a file extension
func (s *Server) getContentType(ext string) string {
	switch strings.ToLower(ext) {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".html", ".htm":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".json":
		return "application/json"
	case ".txt", ".md":
		return "text/plain"
	case ".pdf":
		return "application/pdf"
	default:
		return "application/octet-stream"
	}
}
