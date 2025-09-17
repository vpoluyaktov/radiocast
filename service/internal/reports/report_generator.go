package reports

import (
	"context"
	"fmt"
	"html/template"
	"time"

	"radiocast/internal/charts"
	"radiocast/internal/config"
	"radiocast/internal/fetchers"
	"radiocast/internal/llm"
	"radiocast/internal/logger"
	"radiocast/internal/models"
	"radiocast/internal/mocks"
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

	logger.Info("Starting report generation...")

	// Generate charts
	logger.Info("Generating charts...")
	chartData, err := rg.htmlBuilder.GenerateChartData(propagationData, sourceData, folderPath)
	if err != nil {
		return "", fmt.Errorf("failed to generate charts: %w", err)
	}

	// Use placeholder for Sun GIF (will be replaced by file manager)
	sunGifHTML := template.HTML("{{.SunGif}}")

	// Process markdown with template placeholders
	logger.Info("Processing markdown with placeholders...")
	processedContent, err := rg.htmlBuilder.ProcessMarkdownWithPlaceholders(
		markdownContent, chartData, sunGifHTML)
	if err != nil {
		return "", fmt.Errorf("failed to process markdown: %w", err)
	}

	// Build complete HTML document
	logger.Info("Building complete HTML document...")
	logger.Debug("Processed content length", map[string]interface{}{"length": len(processedContent)})
	logger.Debug("Processed content preview", map[string]interface{}{"preview": processedContent[:min(300, len(processedContent))]})
	finalHTML, err := rg.htmlBuilder.BuildCompleteHTML(
		processedContent, propagationData, chartData, sunGifHTML, folderPath)
	if err != nil {
		return "", fmt.Errorf("failed to build complete HTML: %w", err)
	}

	logger.Info("Report generation completed successfully", map[string]interface{}{"characters": len(finalHTML)})
	return finalHTML, nil
}

// GenerateHTML converts markdown report to HTML with embedded charts
// This is the main public method - all other HTML generation methods are deprecated
func (rg *ReportGenerator) GenerateHTML(markdownReport string, data *models.PropagationData, sourceData *models.SourceData, folderPath string) (string, error) {
	return rg.GenerateReport(context.Background(), data, sourceData, markdownReport, folderPath)
}

// MarkdownToHTML converts markdown to HTML
func (rg *ReportGenerator) MarkdownToHTML(markdownText string) string {
	htmlContent, err := rg.htmlBuilder.ConvertMarkdownToHTML(markdownText)
	if err != nil {
		logger.Error("Error converting markdown to HTML", err)
		return markdownText // Return original on error
	}
	return htmlContent
}

// GenerateStaticCSS generates static CSS content for saving to the report folder
func (rg *ReportGenerator) GenerateStaticCSS() (string, error) {
	return rg.htmlBuilder.LoadStaticCSS()
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

	logger.Info("Starting complete report generation...")

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

	return map[string]interface{}{
		"status":     "success",
		"message":    "Report generated successfully",
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
		logger.Info("Using mock data for report generation...")
		data, sourceData, err = mockService.LoadMockData()
		if err != nil {
			return nil, nil, "", fmt.Errorf("mock data loading failed: %w", err)
		}

		logger.Info("Loading mock LLM response...")
		markdownReport, err = mockService.LoadMockLLMResponse()
		if err != nil {
			return nil, nil, "", fmt.Errorf("mock LLM response loading failed: %w", err)
		}
		
		logger.Info("Mock data loaded successfully", map[string]interface{}{"timestamp": data.Timestamp.Format(time.RFC3339)})
		logger.Info("Mock LLM report loaded successfully", map[string]interface{}{"length": len(markdownReport)})
	} else {
		// Fetch data from all sources
		logger.Info("Fetching data from all sources...")
		data, sourceData, err = fetcher.FetchAllDataWithSources(ctx, cfg.NOAAKIndexURL, cfg.NOAASolarURL, cfg.N0NBHSolarURL, cfg.SIDCRSSURL)
		if err != nil {
			return nil, nil, "", fmt.Errorf("data fetching failed: %w", err)
		}

		logger.Info("Data fetched successfully", map[string]interface{}{"timestamp": data.Timestamp.Format(time.RFC3339)})

		// Generate LLM report with raw source data
		logger.Info("Generating LLM report with raw source data...")
		markdownReport, err = llmClient.GenerateReportWithSources(data, sourceData)
		if err != nil {
			return nil, nil, "", fmt.Errorf("LLM report generation failed: %w", err)
		}

		logger.Info("LLM report generated successfully", map[string]interface{}{"length": len(markdownReport)})
	}

	return data, sourceData, markdownReport, nil
}

