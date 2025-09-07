package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"radiocast/internal/imagery"
	"radiocast/internal/models"
)

// FileManager handles unified file operations for both local and GCS modes
type FileManager struct {
	server *Server
}

// prepareSunGIFHTML generates the HTML section for the Sun GIF with the correct path.
// If folderPath is non-empty (GCS mode), it will prefix the src with that path.
func (fm *FileManager) prepareSunGIFHTML(gifRelName, folderPath string) string {
	var imgSrc string
	if fm.server.Storage != nil {
		// GCS mode - use the full folder path
		// Ensure folderPath ends with '/'
		if !strings.HasSuffix(folderPath, "/") { folderPath += "/" }
		imgSrc = "/files/" + folderPath + gifRelName
	} else {
		// Local mode - use the timestamp directory name from reportDir
		// For local mode, we need to extract the timestamp directory from the ReportDir
		timestampDir := filepath.Base(filepath.Dir(filepath.Join(fm.server.ReportsDir, gifRelName)))
		
		// In GenerateAllFiles, we can get the timestamp directory directly from the ReportFiles.ReportDir
		if reportDir := filepath.Base(filepath.Dir(gifRelName)); reportDir != "" && reportDir != "." {
			timestampDir = reportDir
		}
		
		// Use the timestamp directory in the URL path
		imgSrc = "/files/" + timestampDir + "/" + gifRelName
	}
	
	return fmt.Sprintf(`<div class="chart-section"><div class="chart-container-integrated"><h3>Sun Images for Past 72 Hours</h3><img src="%s" alt="Sun last 72h" style="max-width:100%%;height:auto;border-radius:8px;" /><br/><i>Images copyrighted by the SDO/NASA and Helioviewer project</i></div></div>`, imgSrc)
}

// injectSunGIFIntoHTML replaces the {{SUN_GIF}} placeholder with the actual Sun GIF HTML.
// If the placeholder is not found, it inserts after the Current Solar Activity header.
func (fm *FileManager) injectSunGIFIntoHTML(html, gifRelName, folderPath string) string {
	// Generate the HTML for the Sun GIF
	sunGifHTML := fm.prepareSunGIFHTML(gifRelName, folderPath)
	
	// First try to replace the placeholder if it exists
	const placeholder = "{{SUN_GIF}}"
	if strings.Contains(html, placeholder) {
		return strings.Replace(html, placeholder, sunGifHTML, 1)
	}
	
	// If no placeholder, insert after the Current Solar Activity header
	if strings.Contains(html, "<h2>📊 Current Solar Activity</h2>") {
		return strings.Replace(html, "<h2>📊 Current Solar Activity</h2>", "<h2>📊 Current Solar Activity</h2>\n"+sunGifHTML, 1)
	} else if strings.Contains(html, "Current Solar Activity") {
		// Try a more generic match if the exact header isn't found
		regex := regexp.MustCompile(`(<h2[^>]*>.*Current Solar Activity.*</h2>)`)
		return regex.ReplaceAllString(html, "${1}\n"+sunGifHTML)
	} else {
		// Fallback: Insert right after opening <body> if section header not found
		if strings.Contains(html, "<body>") {
			return strings.Replace(html, "<body>", "<body>\n"+sunGifHTML, 1)
		}
		return html + sunGifHTML
	}
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
	AssetFiles     map[string][]byte // additional assets like GIFs
}

// GenerateAllFiles creates all report files (HTML, charts, JSON) in a unified way
func (fm *FileManager) GenerateAllFiles(ctx context.Context, data *models.PropagationData, sourceData *models.SourceData, markdown string) (*ReportFiles, error) {
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
		AssetFiles: make(map[string][]byte),
	}
	
	// Generate GCS folder path for consistency
	files.FolderPath = fmt.Sprintf("%04d/%02d/%02d/PropagationReport-%04d-%02d-%02d-%02d-%02d-%02d",
		timestamp.Year(), timestamp.Month(), timestamp.Day(),
		timestamp.Year(), timestamp.Month(), timestamp.Day(),
		timestamp.Hour(), timestamp.Minute(), timestamp.Second())
	
	// Copy local chart assets (echarts.min.js) into report directory if available
	if err := fm.copyLocalChartAssets(reportDir); err != nil {
		log.Printf("Warning: Failed to copy local chart assets: %v", err)
	}
	
	// Add background image to AssetFiles for GCS upload
	backgroundFile := filepath.Join(reportDir, "background.png")
	if _, err := os.Stat(backgroundFile); err == nil {
		backgroundData, err := os.ReadFile(backgroundFile)
		if err == nil {
			files.AssetFiles["background.png"] = backgroundData
			log.Printf("Added background.png (%d bytes) to GCS upload queue", len(backgroundData))
		} else {
			log.Printf("Warning: Could not read background image for GCS upload: %v", err)
		}
	}
	
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
	
	// 3.5. Generate Sun GIF (last 72h) using Helioviewer and ffmpeg
	gifRelName := "sun_72h.gif"
	gifPath := filepath.Join(reportDir, gifRelName)
	if err := imagery.GenerateSunGIF(ctx, reportDir, data.Timestamp, gifPath); err != nil {
		log.Printf("Warning: Failed to generate Sun GIF: %v", err)
	} else {
		// Read the GIF into memory for potential GCS upload
		if b, rerr := os.ReadFile(gifPath); rerr == nil {
			files.AssetFiles[gifRelName] = b
			log.Printf("Generated Sun GIF: %s (%d bytes)", gifPath, len(b))
		} else {
			log.Printf("Warning: Could not read generated GIF %s: %v", gifPath, rerr)
		}
	}

	// 4. Generate CSS file with processed templates
	var cssContent string
	var cssErr error
	// Generate static CSS content (no longer needs folderPath since background is in HTML)
	cssContent, cssErr = fm.server.Generator.GenerateStaticCSS()
	if cssErr != nil {
		log.Printf("Warning: Failed to generate CSS: %v", cssErr)
	} else {
		// Save CSS file
		cssPath := filepath.Join(reportDir, "styles.css")
		if writeErr := os.WriteFile(cssPath, []byte(cssContent), 0644); writeErr != nil {
			log.Printf("Warning: Failed to save CSS file: %v", writeErr)
		} else {
			files.AssetFiles["styles.css"] = []byte(cssContent)
			log.Printf("Generated CSS file: %s (%d bytes)", cssPath, len(cssContent))
		}
	}

	// 5. Generate HTML report using ECharts snippets only
	var html string
	var htmlErr error
	if fm.server.Storage != nil {
		// GCS mode - provide folderPath so /files route can resolve local echarts.min.js
		html, htmlErr = fm.server.Generator.GenerateHTMLWithSourcesAndFolderPath(markdown, data, sourceData, files.FolderPath)
	} else {
		// Local mode
		html, htmlErr = fm.server.Generator.GenerateHTMLWithSources(markdown, data, sourceData)
	}
	if htmlErr != nil {
		return nil, fmt.Errorf("failed to generate HTML: %w", htmlErr)
	}
	// Inject Sun GIF section into HTML if generated
	files.HTMLContent = fm.injectSunGIFIntoHTML(html, gifRelName, files.FolderPath)
	
	// 5. Save HTML file locally (index.html for consistency)
	htmlPath := filepath.Join(reportDir, "index.html")
	if err := os.WriteFile(htmlPath, []byte(files.HTMLContent), 0644); err != nil {
		log.Printf("Failed to save HTML report: %v", err)
	}
	
	return files, nil
}

// saveSourceJSONFiles saves separate JSON files for each data source
func (fm *FileManager) saveSourceJSONFiles(reportDir string, sourceData *models.SourceData, files *ReportFiles) error {
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
func (fm *FileManager) saveLLMFiles(reportDir string, data *models.PropagationData, sourceData *models.SourceData, markdown string, files *ReportFiles) error {
	// Save system prompt
	systemPrompt := fm.server.LLMClient.GetSystemPrompt()
	systemPromptPath := filepath.Join(reportDir, "llm_system_prompt.txt")
	if err := os.WriteFile(systemPromptPath, []byte(systemPrompt), 0644); err != nil {
		return err
	}
	files.JSONFiles["llm_system_prompt.txt"] = []byte(systemPrompt)
	
	// Save user prompt (using raw source data)
	llmPrompt := fm.server.LLMClient.BuildPrompt(sourceData, data)
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

// UploadToGCS uploads all files to GCS storage
func (fm *FileManager) UploadToGCS(ctx context.Context, files *ReportFiles, timestamp time.Time) (string, error) {
	if fm.server.Storage == nil {
		return "", fmt.Errorf("GCS storage not configured")
	}
	
	log.Printf("Uploading files to GCS folder: %s", files.FolderPath)
	
	// 2. Upload JSON files
	for filename, data := range files.JSONFiles {
		log.Printf("Uploading JSON file %s (%d bytes) to GCS", filename, len(data))
		if err := fm.server.Storage.StoreFile(ctx, data, filename, timestamp); err != nil {
			log.Printf("Failed to store JSON file %s: %v", filename, err)
		} else {
			log.Printf("JSON file uploaded successfully: %s", filename)
		}
	}

	// 3. Upload asset files (images, GIFs)
	for filename, data := range files.AssetFiles {
		log.Printf("Uploading asset file %s (%d bytes) to GCS", filename, len(data))
		if err := fm.server.Storage.StoreFile(ctx, data, filename, timestamp); err != nil {
			log.Printf("Failed to store asset file %s: %v", filename, err)
		} else {
			log.Printf("Asset file uploaded successfully: %s", filename)
		}
	}
	
	// Special handling for background image if not already in AssetFiles
	if _, exists := files.AssetFiles["background.png"]; !exists {
		// Try to find background image in assets directory
		candidates := []string{
			filepath.Join("internal", "assets", "background.png"),
			filepath.Join("service", "internal", "assets", "background.png"),
			filepath.Join("..", "service", "internal", "assets", "background.png"),
		}
		
		for _, path := range candidates {
			if data, err := os.ReadFile(path); err == nil {
				log.Printf("Found background.png at %s, uploading to GCS (%d bytes)", path, len(data))
				if err := fm.server.Storage.StoreFile(ctx, data, "background.png", timestamp); err != nil {
					log.Printf("Failed to store background.png: %v", err)
				} else {
					log.Printf("Background image uploaded successfully")
				}
				break
			}
		}
	}

	// 4. Upload HTML report
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

// copyLocalChartAssets copies vendored chart assets (no CDN) into the report directory if available.
// Currently copies: echarts.min.js and background.png from service/internal/assets/
func (fm *FileManager) copyLocalChartAssets(reportDir string) error {
    // List of assets to copy
    assets := []string{"echarts.min.js", "background.png"}
    
    for _, asset := range assets {
        // Try different possible paths for the asset
        candidates := []string{
            filepath.Join("internal", "assets", asset),
            filepath.Join("service", "internal", "assets", asset),
            filepath.Join("..", "service", "internal", "assets", asset),
        }
        
        var src string
        for _, c := range candidates {
            if _, err := os.Stat(c); err == nil {
                src = c
                break
            }
        }
        
        if src == "" {
            // Asset not present; non-fatal per requirements
            log.Printf("%s not found in assets; skipping copy", asset)
            continue
        }
        
        // Copy the asset to the report directory
        dst := filepath.Join(reportDir, asset)
        in, err := os.Open(src)
        if err != nil {
            return fmt.Errorf("open asset %s: %w", src, err)
        }
        defer in.Close()
        
        out, err := os.Create(dst)
        if err != nil {
            return fmt.Errorf("create asset %s: %w", dst, err)
        }
        defer out.Close()
        
        if _, err = io.Copy(out, in); err != nil {
            return fmt.Errorf("copy asset %s to %s: %w", src, dst, err)
        }
        
        log.Printf("Copied asset %s to %s", src, dst)
    }
    
    return nil
}
