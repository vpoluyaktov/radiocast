package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"radiocast/internal/fetchers"
	"radiocast/internal/models"
	"radiocast/internal/reports"
)

// FileManager handles unified file operations for both local and GCS modes
type FileManager struct {
	server *Server
}

// NewFileManager creates a new file manager
func NewFileManager(server *Server) *FileManager {
	return &FileManager{server: server}
}

// ReportFiles contains all files generated for a report
type ReportFiles struct {
	HTMLContent    string
	ChartFiles     []string
	JSONFiles      map[string][]byte
	ReportDir      string
	FolderPath     string // GCS folder path
}

// GenerateAllFiles creates all report files (HTML, charts, JSON) in a unified way
func (fm *FileManager) GenerateAllFiles(ctx context.Context, data *models.PropagationData, sourceData *fetchers.SourceData, markdown string) (*ReportFiles, error) {
	timestamp := data.Timestamp
	
	// Create report directory (local or temp for GCS)
	var reportDir string
	if fm.server.Storage != nil {
		// GCS mode - use temp directory
		reportDir = filepath.Join(os.TempDir(), fmt.Sprintf("report_%d", time.Now().Unix()))
	} else {
		// Local mode - use timestamped directory
		reportDir = filepath.Join(fm.server.ReportsDir, timestamp.Format("2006-01-02_15-04-05"))
	}
	
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create report directory: %w", err)
	}
	
	files := &ReportFiles{
		ReportDir: reportDir,
		JSONFiles: make(map[string][]byte),
	}
	
	// Generate GCS folder path for consistency
	files.FolderPath = fmt.Sprintf("%04d/%02d/%02d/PropagationReport-%04d-%02d-%02d-%02d-%02d-%02d",
		timestamp.Year(), timestamp.Month(), timestamp.Day(),
		timestamp.Year(), timestamp.Month(), timestamp.Day(),
		timestamp.Hour(), timestamp.Minute(), timestamp.Second())
	
	// 1. Save separate JSON files for each data source
	if err := fm.saveSourceJSONFiles(reportDir, sourceData, files); err != nil {
		log.Printf("Warning: Failed to save source JSON files: %v", err)
	}
	
	// 2. Save normalized data JSON
	if err := fm.saveNormalizedData(reportDir, data, files); err != nil {
		log.Printf("Warning: Failed to save normalized data: %v", err)
	}
	
	// 3. Save LLM-related files
	if err := fm.saveLLMFiles(reportDir, data, sourceData, markdown, files); err != nil {
		log.Printf("Warning: Failed to save LLM files: %v", err)
	}
	
	// 4. Generate PNG charts
	chartFiles, err := fm.generateCharts(reportDir, data)
	if err != nil {
		log.Printf("Warning: Failed to generate charts: %v", err)
		chartFiles = []string{}
	}
	files.ChartFiles = chartFiles
	
	// 5. Generate HTML report
	html, err := fm.generateHTML(markdown, data, chartFiles, files.FolderPath)
	if err != nil {
		return nil, fmt.Errorf("failed to generate HTML: %w", err)
	}
	files.HTMLContent = html
	
	// 6. Save HTML file locally (index.html for consistency)
	htmlPath := filepath.Join(reportDir, "index.html")
	if err := os.WriteFile(htmlPath, []byte(html), 0644); err != nil {
		log.Printf("Failed to save HTML report: %v", err)
	}
	
	return files, nil
}

// saveSourceJSONFiles saves separate JSON files for each data source
func (fm *FileManager) saveSourceJSONFiles(reportDir string, sourceData *fetchers.SourceData, files *ReportFiles) error {
	if sourceData.NOAAKIndex != nil {
		data, _ := json.MarshalIndent(sourceData.NOAAKIndex, "", "  ")
		path := filepath.Join(reportDir, "noaa_k_index.json")
		if err := os.WriteFile(path, data, 0644); err != nil {
			return err
		}
		files.JSONFiles["noaa_k_index.json"] = data
		log.Printf("Saved NOAA K-Index data to: %s", path)
	}
	
	if sourceData.NOAASolar != nil {
		data, _ := json.MarshalIndent(sourceData.NOAASolar, "", "  ")
		path := filepath.Join(reportDir, "noaa_solar.json")
		if err := os.WriteFile(path, data, 0644); err != nil {
			return err
		}
		files.JSONFiles["noaa_solar.json"] = data
		log.Printf("Saved NOAA Solar data to: %s", path)
	}
	
	if sourceData.N0NBH != nil {
		data, _ := json.MarshalIndent(sourceData.N0NBH, "", "  ")
		path := filepath.Join(reportDir, "n0nbh_data.json")
		if err := os.WriteFile(path, data, 0644); err != nil {
			return err
		}
		files.JSONFiles["n0nbh_data.json"] = data
		log.Printf("Saved N0NBH data to: %s", path)
	}
	
	if sourceData.SIDC != nil {
		data, _ := json.MarshalIndent(sourceData.SIDC, "", "  ")
		path := filepath.Join(reportDir, "sidc_data.json")
		if err := os.WriteFile(path, data, 0644); err != nil {
			return err
		}
		files.JSONFiles["sidc_data.json"] = data
		log.Printf("Saved SIDC data to: %s", path)
	}
	
	return nil
}

// saveNormalizedData saves the normalized/combined data
func (fm *FileManager) saveNormalizedData(reportDir string, data *models.PropagationData, files *ReportFiles) error {
	normalizedData, _ := json.MarshalIndent(data, "", "  ")
	path := filepath.Join(reportDir, "normalized_data.json")
	if err := os.WriteFile(path, normalizedData, 0644); err != nil {
		return err
	}
	files.JSONFiles["normalized_data.json"] = normalizedData
	return nil
}

// saveLLMFiles saves LLM-related files (prompts, responses)
func (fm *FileManager) saveLLMFiles(reportDir string, data *models.PropagationData, sourceData *fetchers.SourceData, markdown string, files *ReportFiles) error {
	// Save system prompt
	systemPrompt := fm.server.LLMClient.GetSystemPrompt()
	systemPromptPath := filepath.Join(reportDir, "llm_system_prompt.txt")
	if err := os.WriteFile(systemPromptPath, []byte(systemPrompt), 0644); err != nil {
		return err
	}
	files.JSONFiles["llm_system_prompt.txt"] = []byte(systemPrompt)
	
	// Save user prompt (using raw source data)
	llmPrompt := fm.server.LLMClient.BuildPromptWithRawData(sourceData, data)
	promptPath := filepath.Join(reportDir, "llm_prompt.txt")
	if err := os.WriteFile(promptPath, []byte(llmPrompt), 0644); err != nil {
		return err
	}
	files.JSONFiles["llm_prompt.txt"] = []byte(llmPrompt)
	
	// Save markdown response
	markdownPath := filepath.Join(reportDir, "llm_response.md")
	if err := os.WriteFile(markdownPath, []byte(markdown), 0644); err != nil {
		return err
	}
	files.JSONFiles["llm_response.md"] = []byte(markdown)
	
	return nil
}

// generateCharts creates PNG chart files
func (fm *FileManager) generateCharts(reportDir string, data *models.PropagationData) ([]string, error) {
	chartGen := reports.NewChartGenerator(reportDir)
	log.Printf("Generating PNG charts in directory: %s", reportDir)
	chartFiles, err := chartGen.GenerateCharts(data)
	if err != nil {
		return nil, err
	}
	log.Printf("Successfully generated %d chart files: %v", len(chartFiles), chartFiles)
	return chartFiles, nil
}

// generateHTML creates the HTML report
func (fm *FileManager) generateHTML(markdown string, data *models.PropagationData, chartFiles []string, folderPath string) (string, error) {
	if fm.server.Storage != nil {
		// GCS mode - charts will be uploaded separately, use folder path for URLs
		return fm.server.Generator.GenerateHTMLWithFolderPath(markdown, data, chartFiles, folderPath)
	} else {
		// Local mode - charts are in same directory
		return fm.server.Generator.GenerateHTMLWithLocalCharts(markdown, data, chartFiles)
	}
}

// UploadToGCS uploads all files to GCS storage
func (fm *FileManager) UploadToGCS(ctx context.Context, files *ReportFiles, timestamp time.Time) (string, error) {
	if fm.server.Storage == nil {
		return "", fmt.Errorf("GCS storage not configured")
	}
	
	log.Printf("Uploading files to GCS folder: %s", files.FolderPath)
	
	// 1. Upload chart images
	for _, chartFile := range files.ChartFiles {
		imageData, err := os.ReadFile(chartFile)
		if err != nil {
			log.Printf("Failed to read chart file %s: %v", chartFile, err)
			continue
		}
		
		filename := filepath.Base(chartFile)
		log.Printf("Uploading chart image %s (%d bytes) to GCS", filename, len(imageData))
		publicURL, err := fm.server.Storage.StoreChartImage(ctx, imageData, filename, timestamp)
		if err != nil {
			log.Printf("Failed to store chart image %s: %v", filename, err)
			continue
		}
		
		log.Printf("Chart image uploaded successfully: %s", publicURL)
	}
	
	// 2. Upload JSON files
	for filename, data := range files.JSONFiles {
		log.Printf("Uploading JSON file %s (%d bytes) to GCS", filename, len(data))
		if err := fm.server.Storage.StoreFile(ctx, data, filename, timestamp); err != nil {
			log.Printf("Failed to store JSON file %s: %v", filename, err)
		} else {
			log.Printf("JSON file uploaded successfully: %s", filename)
		}
	}
	
	// 3. Upload HTML report
	log.Printf("Uploading HTML report (%d bytes) to GCS", len(files.HTMLContent))
	reportURL, err := fm.server.Storage.StoreReport(ctx, files.HTMLContent, timestamp)
	if err != nil {
		return "", fmt.Errorf("failed to store HTML report: %w", err)
	}
	
	log.Printf("All files uploaded successfully. Report URL: %s", reportURL)
	return reportURL, nil
}

// Cleanup removes temporary files (for GCS mode)
func (fm *FileManager) Cleanup(files *ReportFiles) {
	if fm.server.Storage != nil && files.ReportDir != "" {
		log.Printf("Cleaning up temporary directory: %s", files.ReportDir)
		os.RemoveAll(files.ReportDir)
	}
}
