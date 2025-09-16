package reports

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"time"

	"radiocast/internal/charts"
	"radiocast/internal/config"
	"radiocast/internal/fetchers"
	"radiocast/internal/llm"
	"radiocast/internal/mocks"
	"radiocast/internal/models"
	"radiocast/internal/storage"
)

// StorageInterface defines the interface for storage operations
type StorageInterface interface {
	StoreAllFiles(ctx context.Context, files *GeneratedFiles, data *models.PropagationData) error
}

// ReportGenerator handles report generation and HTML conversion
type ReportGenerator struct {
	chartGen    *charts.ChartGenerator
	htmlBuilder *HTMLBuilder
}

// NewReportGenerator creates a new report generator
func NewReportGenerator() *ReportGenerator {
	return &ReportGenerator{
		chartGen:    charts.NewChartGenerator(""), // Empty outputDir since charts don't need it
		htmlBuilder: NewHTMLBuilder(),
	}
}

// GenerateReport generates a complete HTML report
func (rg *ReportGenerator) GenerateReport(ctx context.Context,
	propagationData *models.PropagationData,
	sourceData *models.SourceData,
	markdownContent string,
	folderPath string) (string, error) {

	log.Println("Starting report generation...")

	// Generate charts
	log.Println("Generating charts...")
	chartData, err := rg.generateCharts(propagationData, sourceData, folderPath)
	if err != nil {
		return "", fmt.Errorf("failed to generate charts: %w", err)
	}

	// Use placeholder for Sun GIF (will be replaced by file manager)
	sunGifHTML := template.HTML("{{.SunGif}}")

	// Process markdown with template placeholders
	log.Println("Processing markdown with placeholders...")
	processedContent, err := rg.htmlBuilder.ProcessMarkdownWithPlaceholders(
		markdownContent, chartData, sunGifHTML)
	if err != nil {
		return "", fmt.Errorf("failed to process markdown: %w", err)
	}

	// Build complete HTML document
	log.Println("Building complete HTML document...")
	log.Printf("Processed content length: %d", len(processedContent))
	log.Printf("Processed content preview: %s", processedContent[:min(300, len(processedContent))])
	finalHTML, err := rg.htmlBuilder.BuildCompleteHTML(
		processedContent, propagationData, chartData, sunGifHTML, folderPath)
	if err != nil {
		return "", fmt.Errorf("failed to build complete HTML: %w", err)
	}

	log.Printf("Report generation completed successfully (%d characters)", len(finalHTML))
	return finalHTML, nil
}

// GenerateHTML converts markdown report to HTML with embedded charts
func (rg *ReportGenerator) GenerateHTML(markdownReport string, data *models.PropagationData) (string, error) {
	return rg.GenerateHTMLWithSources(markdownReport, data, nil)
}

// GenerateHTMLWithSources converts markdown report to HTML with embedded charts using source data
func (rg *ReportGenerator) GenerateHTMLWithSources(markdownReport string, data *models.PropagationData, sourceData *models.SourceData) (string, error) {
	return rg.GenerateHTMLWithSourcesAndFolderPath(markdownReport, data, sourceData, "")
}

// GenerateHTMLWithSourcesAndFolderPath converts markdown to HTML with ECharts snippets
// and allows specifying folderPath for asset path resolution.
func (rg *ReportGenerator) GenerateHTMLWithSourcesAndFolderPath(markdownReport string, data *models.PropagationData, sourceData *models.SourceData, folderPath string) (string, error) {
	log.Println("Generating report...")
	
	// Use main report generation method
	fullHTML, err := rg.GenerateReport(context.Background(), data, sourceData, markdownReport, folderPath)
	if err != nil {
		return "", fmt.Errorf("failed to generate report: %w", err)
	}
	
	log.Printf("Generated complete HTML report (%d characters)", len(fullHTML))
	return fullHTML, nil
}

// GenerateHTMLWithChartURLs converts markdown report to HTML using provided chart URLs
func (rg *ReportGenerator) GenerateHTMLWithChartURLs(markdownReport string, data *models.PropagationData, chartURLs []string) (string, error) {
	log.Printf("Converting markdown to HTML with %d provided chart URLs...", len(chartURLs))
	
	// Use main report generation method
	fullHTML, err := rg.GenerateReport(context.Background(), data, nil, markdownReport, "")
	if err != nil {
		return "", fmt.Errorf("failed to build complete HTML: %w", err)
	}
	
	log.Printf("Generated complete HTML report with %d characters and %d chart URLs", len(fullHTML), len(chartURLs))
	return fullHTML, nil
}

// MarkdownToHTML converts markdown to HTML
func (rg *ReportGenerator) MarkdownToHTML(markdownText string) string {
	htmlContent, err := rg.htmlBuilder.ConvertMarkdownToHTML(markdownText)
	if err != nil {
		log.Printf("Error converting markdown to HTML: %v", err)
		return markdownText // Return original on error
	}
	return htmlContent
}

// GenerateStaticCSS generates static CSS content for saving to the report folder
func (rg *ReportGenerator) GenerateStaticCSS() (string, error) {
	return rg.htmlBuilder.GenerateStaticCSS()
}



// GenerateCompleteReport handles the complete report generation pipeline
func (rg *ReportGenerator) GenerateCompleteReport(ctx context.Context,
	cfg *config.Config,
	fetcher *fetchers.DataFetcher,
	llmClient *llm.OpenAIClient,
	mockService *mocks.MockService,
	storage storage.StorageClient,
	deploymentMode string,
	storageOrchestrator StorageInterface) (map[string]interface{}, error) {

	log.Println("Starting report generation...")

	// Step 1: Get data and generate markdown report
	data, sourceData, markdownReport, err := rg.fetchDataAndGenerateReport(ctx, cfg, fetcher, llmClient, mockService)
	if err != nil {
		return nil, err
	}

	// Step 2: Generate files using FileGenerator
	fileGenerator := NewFileGenerator(rg, mockService)
	files, err := fileGenerator.GenerateAllFiles(ctx, data, sourceData, markdownReport, cfg.MockupMode)
	if err != nil {
		return nil, fmt.Errorf("failed to generate files: %w", err)
	}

	// Step 3: Store files using StorageOrchestrator
	if err := storageOrchestrator.StoreAllFiles(ctx, files, data); err != nil {
		return nil, fmt.Errorf("failed to store files: %w", err)
	}

	// Step 4: Determine report URL based on deployment mode
	var reportURL string
	if deploymentMode == "gcs" && storage != nil {
		reportURL = "/files/" + data.Timestamp.Format("2006-01-02_15-04-05") + "/index.html"
	} else {
		reportURL = "/files/index.html"
	}

	return map[string]interface{}{
		"status":     "success",
		"message":    "Report generated successfully",
		"reportURL":  reportURL,
		"timestamp":  data.Timestamp.Format(time.RFC3339),
		"dataPoints": len(data.SourceEvents),
		"folderPath": data.Timestamp.Format("2006-01-02_15-04-05"),
	}, nil
}

// fetchDataAndGenerateReport handles data fetching and LLM report generation
func (rg *ReportGenerator) fetchDataAndGenerateReport(ctx context.Context,
	cfg *config.Config,
	fetcher *fetchers.DataFetcher,
	llmClient *llm.OpenAIClient,
	mockService *mocks.MockService) (*models.PropagationData, *models.SourceData, string, error) {

	var data *models.PropagationData
	var sourceData *models.SourceData
	var markdownReport string
	var err error

	if cfg.MockupMode && mockService != nil {
		// Use mock data
		log.Println("Using mock data for report generation...")
		data, sourceData, err = mockService.LoadMockData()
		if err != nil {
			return nil, nil, "", fmt.Errorf("mock data loading failed: %w", err)
		}

		log.Println("Loading mock LLM response...")
		markdownReport, err = mockService.LoadMockLLMResponse()
		if err != nil {
			return nil, nil, "", fmt.Errorf("mock LLM response loading failed: %w", err)
		}
		
		log.Printf("Mock data loaded successfully for timestamp: %s", data.Timestamp.Format(time.RFC3339))
		log.Printf("Mock LLM report loaded successfully (length: %d characters)", len(markdownReport))
	} else {
		// Fetch data from all sources
		log.Println("Fetching data from all sources...")
		data, sourceData, err = fetcher.FetchAllDataWithSources(ctx, cfg.NOAAKIndexURL, cfg.NOAASolarURL, cfg.N0NBHSolarURL, cfg.SIDCRSSURL)
		if err != nil {
			return nil, nil, "", fmt.Errorf("data fetching failed: %w", err)
		}

		log.Printf("Data fetched successfully for timestamp: %s", data.Timestamp.Format(time.RFC3339))

		// Generate LLM report with raw source data
		log.Println("Generating LLM report with raw source data...")
		markdownReport, err = llmClient.GenerateReportWithSources(data, sourceData)
		if err != nil {
			return nil, nil, "", fmt.Errorf("LLM report generation failed: %w", err)
		}

		log.Printf("LLM report generated successfully (length: %d characters)", len(markdownReport))
	}

	return data, sourceData, markdownReport, nil
}

// generateCharts creates charts and returns template data
func (rg *ReportGenerator) generateCharts(data *models.PropagationData, sourceData *models.SourceData, folderPath string) (*ChartTemplateData, error) {
	return rg.htmlBuilder.GenerateChartData(data, sourceData, folderPath)
}
