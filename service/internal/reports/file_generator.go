package reports

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"radiocast/internal/imagery"
	"radiocast/internal/logger"
	"radiocast/internal/models"
	"radiocast/internal/storage"
)

// FileGenerator handles generation of all report files
type FileGenerator struct {
	reportGenerator *ReportGenerator
	mockService     MockService
}

// MockService interface for dependency injection
type MockService interface {
	LoadMockSunGif() ([]byte, error)
}

// GeneratedFiles contains all files generated for a report
type GeneratedFiles struct {
	HTMLContent    string
	ChartFiles     []string
	JSONFiles      map[string][]byte
	AssetFiles     map[string][]byte // CSS, GIFs, images
	FolderPath     string            // GCS folder path for consistency
}

// NewFileGenerator creates a new file generator
func NewFileGenerator(reportGenerator *ReportGenerator, mockService MockService) *FileGenerator {
	return &FileGenerator{
		reportGenerator: reportGenerator,
		mockService:     mockService,
	}
}

// GenerateAllFiles creates all report files (HTML, charts, JSON, assets)
func (fg *FileGenerator) GenerateAllFiles(ctx context.Context, data *models.PropagationData, sourceData *models.SourceData, markdown string, mockupMode bool) (*GeneratedFiles, error) {
	timestamp := data.Timestamp
	
	files := &GeneratedFiles{
		JSONFiles:  make(map[string][]byte),
		AssetFiles: make(map[string][]byte),
	}
	
	// Generate unified folder path using storage utility
	files.FolderPath = storage.GenerateReportFolderPath(timestamp)
	
	// 1. Generate JSON files for each data source
	if err := fg.generateSourceJSONFiles(sourceData, files); err != nil {
		logger.Warn("Failed to generate source JSON files", map[string]interface{}{"error": err.Error()})
	}
	
	// 2. Generate normalized data JSON
	if err := fg.generateNormalizedDataJSON(data, files); err != nil {
		logger.Warn("Failed to generate normalized data", map[string]interface{}{"error": err.Error()})
	}
	
	// 3. Generate LLM-related files
	if err := fg.generateLLMFiles(markdown, files); err != nil {
		logger.Warn("Failed to generate LLM files", map[string]interface{}{"error": err.Error()})
	}
	
	// 4. Generate Sun GIF (last 72h)
	gifRelName := "sun_72h.gif"
	if err := fg.generateSunGIF(ctx, mockupMode, timestamp, gifRelName, files); err != nil {
		logger.Warn("Failed to generate Sun GIF", map[string]interface{}{"error": err.Error()})
	}

	// 5. Generate HTML report (CSS generation removed - all pages use /static/styles.css)
	if err := fg.generateHTML(markdown, data, sourceData, gifRelName, files); err != nil {
		return nil, fmt.Errorf("failed to generate HTML: %w", err)
	}
	
	
	return files, nil
}

// generateSourceJSONFiles generates separate JSON files for each data source
func (fg *FileGenerator) generateSourceJSONFiles(sourceData *models.SourceData, files *GeneratedFiles) error {
	if sourceData.NOAAKIndex != nil {
		data, _ := json.MarshalIndent(sourceData.NOAAKIndex, "", "  ")
		files.JSONFiles["noaa_k_index.json"] = data
		logger.Debug("Generated NOAA K-Index JSON", map[string]interface{}{"bytes": len(data)})
	}
	
	if sourceData.NOAASolar != nil {
		data, _ := json.MarshalIndent(sourceData.NOAASolar, "", "  ")
		files.JSONFiles["noaa_solar.json"] = data
		logger.Debug("Generated NOAA Solar JSON", map[string]interface{}{"bytes": len(data)})
	}
	
	if sourceData.N0NBH != nil {
		data, _ := json.MarshalIndent(sourceData.N0NBH, "", "  ")
		files.JSONFiles["n0nbh_data.json"] = data
		logger.Debug("Generated N0NBH JSON", map[string]interface{}{"bytes": len(data)})
	}
	
	if sourceData.SIDC != nil {
		data, _ := json.MarshalIndent(sourceData.SIDC, "", "  ")
		files.JSONFiles["sidc_data.json"] = data
		logger.Debug("Generated SIDC JSON", map[string]interface{}{"bytes": len(data)})
	}
	
	return nil
}

// generateNormalizedDataJSON generates the normalized/combined data JSON
func (fg *FileGenerator) generateNormalizedDataJSON(data *models.PropagationData, files *GeneratedFiles) error {
	normalizedData, _ := json.MarshalIndent(data, "", "  ")
	files.JSONFiles["normalized_data.json"] = normalizedData
	logger.Debug("Generated normalized data JSON", map[string]interface{}{"bytes": len(normalizedData)})
	return nil
}

// generateLLMFiles generates LLM-related files (prompts, responses)
func (fg *FileGenerator) generateLLMFiles(markdown string, files *GeneratedFiles) error {
	// Note: This requires access to LLMClient which should be passed in
	// For now, we'll store the markdown response
	files.JSONFiles["llm_response.md"] = []byte(markdown)
	logger.Debug("Generated LLM response file", map[string]interface{}{"bytes": len(markdown)})
	return nil
}

// generateSunGIF generates Sun GIF using Helioviewer or mock data
func (fg *FileGenerator) generateSunGIF(ctx context.Context, mockupMode bool, timestamp time.Time, gifRelName string, files *GeneratedFiles) error {
	if mockupMode && fg.mockService != nil {
		// Use mock Sun GIF
		logger.Info("Generating mock Sun GIF data...")
		mockGifData, err := fg.mockService.LoadMockSunGif()
		if err != nil {
			return fmt.Errorf("failed to load mock Sun GIF: %w", err)
		}
		files.AssetFiles[gifRelName] = mockGifData
		logger.Debug("Generated mock Sun GIF", map[string]interface{}{"bytes": len(mockGifData)})
	} else {
		// Generate Sun GIF using Helioviewer and ffmpeg
		// Create temporary file for generation
		tempDir := os.TempDir()
		gifPath := filepath.Join(tempDir, fmt.Sprintf("sun_gif_%d.gif", time.Now().Unix()))
		
		if err := imagery.GenerateSunGIF(ctx, tempDir, timestamp, gifPath); err != nil {
			return fmt.Errorf("failed to generate Sun GIF: %w", err)
		}
		
		// Read the generated GIF
		gifData, err := os.ReadFile(gifPath)
		if err != nil {
			return fmt.Errorf("failed to read generated GIF: %w", err)
		}
		
		files.AssetFiles[gifRelName] = gifData
		logger.Debug("Generated Sun GIF", map[string]interface{}{"bytes": len(gifData)})
		
		// Clean up temporary file
		os.Remove(gifPath)
	}
	
	return nil
}


// generateHTML generates HTML report
func (fg *FileGenerator) generateHTML(markdown string, data *models.PropagationData, sourceData *models.SourceData, gifRelName string, files *GeneratedFiles) error {
	// Generate HTML with folder path for GCS compatibility
	html, err := fg.reportGenerator.GenerateHTML(markdown, data, sourceData, files.FolderPath)
	if err != nil {
		return fmt.Errorf("failed to generate HTML: %w", err)
	}
	
	// Inject Sun GIF section into HTML
	files.HTMLContent = fg.injectSunGIFIntoHTML(html, gifRelName, files.FolderPath)
	logger.Debug("Generated HTML report", map[string]interface{}{"bytes": len(files.HTMLContent)})
	return nil
}


// prepareSunGIFHTML generates the HTML section for the Sun GIF with the correct path
func (fg *FileGenerator) prepareSunGIFHTML(gifRelName, folderPath string) string {
	var imgSrc string
	if folderPath != "" {
		// GCS mode - use the full folder path
		if !strings.HasSuffix(folderPath, "/") {
			folderPath += "/"
		}
		imgSrc = "/reports/" + folderPath + gifRelName
	} else {
		// Local mode - use relative path
		imgSrc = gifRelName
	}
	
	return fmt.Sprintf(`<div class="chart-section"><div class="chart-container-integrated"><h3>Sun Images for Past 72 Hours</h3><img src="%s" alt="Sun last 72h" style="max-width:100%%;height:auto;border-radius:8px;" /><br/><i>Images copyrighted by the SDO/NASA and Helioviewer project</i></div></div>`, imgSrc)
}

// injectSunGIFIntoHTML replaces the {{.SunGif}} placeholder with the actual Sun GIF HTML
func (fg *FileGenerator) injectSunGIFIntoHTML(html, gifRelName, folderPath string) string {
	sunGifHTML := fg.prepareSunGIFHTML(gifRelName, folderPath)
	const placeholder = "{{.SunGif}}"
	return strings.Replace(html, placeholder, sunGifHTML, 1)
}
