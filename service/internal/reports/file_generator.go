package reports

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"radiocast/internal/imagery"
	"radiocast/internal/models"
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
	
	// Generate unified folder path for both local and GCS modes
	files.FolderPath = fmt.Sprintf("reports/%04d/%02d/%02d/PropagationReport-%04d-%02d-%02d-%02d-%02d-%02d",
		timestamp.Year(), timestamp.Month(), timestamp.Day(),
		timestamp.Year(), timestamp.Month(), timestamp.Day(),
		timestamp.Hour(), timestamp.Minute(), timestamp.Second())
	
	// 1. Generate JSON files for each data source
	if err := fg.generateSourceJSONFiles(sourceData, files); err != nil {
		log.Printf("Warning: Failed to generate source JSON files: %v", err)
	}
	
	// 2. Generate normalized data JSON
	if err := fg.generateNormalizedDataJSON(data, files); err != nil {
		log.Printf("Warning: Failed to generate normalized data: %v", err)
	}
	
	// 3. Generate LLM-related files
	if err := fg.generateLLMFiles(data, sourceData, markdown, files); err != nil {
		log.Printf("Warning: Failed to generate LLM files: %v", err)
	}
	
	// 4. Generate Sun GIF (last 72h)
	gifRelName := "sun_72h.gif"
	if err := fg.generateSunGIF(ctx, mockupMode, timestamp, gifRelName, files); err != nil {
		log.Printf("Warning: Failed to generate Sun GIF: %v", err)
	}

	// 5. Generate CSS file
	if err := fg.generateCSS(files); err != nil {
		log.Printf("Warning: Failed to generate CSS: %v", err)
	}

	// 6. Generate HTML report
	if err := fg.generateHTML(markdown, data, sourceData, gifRelName, files); err != nil {
		return nil, fmt.Errorf("failed to generate HTML: %w", err)
	}
	
	// 7. Generate static assets (background image, echarts.js)
	if err := fg.generateStaticAssets(files); err != nil {
		log.Printf("Warning: Failed to generate static assets: %v", err)
	}
	
	return files, nil
}

// generateSourceJSONFiles generates separate JSON files for each data source
func (fg *FileGenerator) generateSourceJSONFiles(sourceData *models.SourceData, files *GeneratedFiles) error {
	if sourceData.NOAAKIndex != nil {
		data, _ := json.MarshalIndent(sourceData.NOAAKIndex, "", "  ")
		files.JSONFiles["noaa_k_index.json"] = data
		log.Printf("Generated NOAA K-Index JSON (%d bytes)", len(data))
	}
	
	if sourceData.NOAASolar != nil {
		data, _ := json.MarshalIndent(sourceData.NOAASolar, "", "  ")
		files.JSONFiles["noaa_solar.json"] = data
		log.Printf("Generated NOAA Solar JSON (%d bytes)", len(data))
	}
	
	if sourceData.N0NBH != nil {
		data, _ := json.MarshalIndent(sourceData.N0NBH, "", "  ")
		files.JSONFiles["n0nbh_data.json"] = data
		log.Printf("Generated N0NBH JSON (%d bytes)", len(data))
	}
	
	if sourceData.SIDC != nil {
		data, _ := json.MarshalIndent(sourceData.SIDC, "", "  ")
		files.JSONFiles["sidc_data.json"] = data
		log.Printf("Generated SIDC JSON (%d bytes)", len(data))
	}
	
	return nil
}

// generateNormalizedDataJSON generates the normalized/combined data JSON
func (fg *FileGenerator) generateNormalizedDataJSON(data *models.PropagationData, files *GeneratedFiles) error {
	normalizedData, _ := json.MarshalIndent(data, "", "  ")
	files.JSONFiles["normalized_data.json"] = normalizedData
	log.Printf("Generated normalized data JSON (%d bytes)", len(normalizedData))
	return nil
}

// generateLLMFiles generates LLM-related files (prompts, responses)
func (fg *FileGenerator) generateLLMFiles(data *models.PropagationData, sourceData *models.SourceData, markdown string, files *GeneratedFiles) error {
	// Note: This requires access to LLMClient which should be passed in
	// For now, we'll store the markdown response
	files.JSONFiles["llm_response.md"] = []byte(markdown)
	log.Printf("Generated LLM response file (%d bytes)", len(markdown))
	return nil
}

// generateSunGIF generates Sun GIF using Helioviewer or mock data
func (fg *FileGenerator) generateSunGIF(ctx context.Context, mockupMode bool, timestamp time.Time, gifRelName string, files *GeneratedFiles) error {
	if mockupMode && fg.mockService != nil {
		// Use mock Sun GIF
		log.Println("Generating mock Sun GIF data...")
		mockGifData, err := fg.mockService.LoadMockSunGif()
		if err != nil {
			return fmt.Errorf("failed to load mock Sun GIF: %w", err)
		}
		files.AssetFiles[gifRelName] = mockGifData
		log.Printf("Generated mock Sun GIF (%d bytes)", len(mockGifData))
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
		log.Printf("Generated Sun GIF (%d bytes)", len(gifData))
		
		// Clean up temporary file
		os.Remove(gifPath)
	}
	
	return nil
}

// generateCSS generates CSS content
func (fg *FileGenerator) generateCSS(files *GeneratedFiles) error {
	cssContent, err := fg.reportGenerator.GenerateStaticCSS()
	if err != nil {
		return fmt.Errorf("failed to generate CSS: %w", err)
	}
	
	files.AssetFiles["styles.css"] = []byte(cssContent)
	log.Printf("Generated CSS file (%d bytes)", len(cssContent))
	return nil
}

// generateHTML generates HTML report
func (fg *FileGenerator) generateHTML(markdown string, data *models.PropagationData, sourceData *models.SourceData, gifRelName string, files *GeneratedFiles) error {
	// Generate HTML with folder path for GCS compatibility
	html, err := fg.reportGenerator.GenerateHTMLWithSourcesAndFolderPath(markdown, data, sourceData, files.FolderPath)
	if err != nil {
		return fmt.Errorf("failed to generate HTML: %w", err)
	}
	
	// Inject Sun GIF section into HTML
	files.HTMLContent = fg.injectSunGIFIntoHTML(html, gifRelName, files.FolderPath)
	log.Printf("Generated HTML report (%d bytes)", len(files.HTMLContent))
	return nil
}

// generateStaticAssets generates static assets like background image and echarts.js
func (fg *FileGenerator) generateStaticAssets(files *GeneratedFiles) error {
	// Try to find and include background image
	candidates := []string{
		filepath.Join("internal", "assets", "background.png"),
		filepath.Join("service", "internal", "assets", "background.png"),
		filepath.Join("..", "service", "internal", "assets", "background.png"),
	}
	
	for _, path := range candidates {
		if data, err := os.ReadFile(path); err == nil {
			files.AssetFiles["background.png"] = data
			log.Printf("Generated background.png asset (%d bytes)", len(data))
			break
		}
	}
	
	// Try to find and include echarts.min.js
	for _, path := range candidates {
		jsPath := strings.Replace(path, "background.png", "echarts.min.js", 1)
		if data, err := os.ReadFile(jsPath); err == nil {
			files.AssetFiles["echarts.min.js"] = data
			log.Printf("Generated echarts.min.js asset (%d bytes)", len(data))
			break
		}
	}
	
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
		imgSrc = "/files/" + folderPath + gifRelName
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
