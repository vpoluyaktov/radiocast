package reports

import (
	"context"
	"fmt"
	"html/template"
	"log"

	"radiocast/internal/charts"
	"radiocast/internal/models"
)

// ReportService orchestrates report generation
type ReportService struct {
	outputDir   string
	chartGen    *charts.ChartGenerator
	htmlBuilder *HTMLBuilder
}

// NewReportService creates a new report service
func NewReportService(outputDir string) *ReportService {
	return &ReportService{
		outputDir:   outputDir,
		chartGen:    charts.NewChartGenerator(outputDir),
		htmlBuilder: NewHTMLBuilder(),
	}
}

// GenerateReport generates a complete HTML report
func (rs *ReportService) GenerateReport(ctx context.Context,
	propagationData *models.PropagationData,
	sourceData *models.SourceData,
	markdownContent string,
	folderPath string) (string, error) {

	log.Println("Starting report generation...")

	// Generate charts
	log.Println("Generating charts...")
	chartData, err := rs.generateCharts(propagationData, sourceData, folderPath)
	if err != nil {
		return "", fmt.Errorf("failed to generate charts: %w", err)
	}

	// Generate Sun GIF
	log.Println("Generating Sun GIF...")
	sunGifHTML, err := rs.generateSunGif(ctx, folderPath)
	if err != nil {
		log.Printf("Warning: Failed to generate Sun GIF: %v", err)
		sunGifHTML = template.HTML(`<div class="sun-placeholder">Sun imagery unavailable</div>`)
	}

	// Process markdown with template placeholders
	log.Println("Processing markdown with placeholders...")
	bandTableHTML := rs.htmlBuilder.generateBandTable(propagationData)
	processedContent, err := rs.htmlBuilder.ProcessMarkdownWithPlaceholders(
		markdownContent, chartData, sunGifHTML, bandTableHTML)
	if err != nil {
		return "", fmt.Errorf("failed to process markdown: %w", err)
	}

	// Build complete HTML document
	log.Println("Building complete HTML document...")
	log.Printf("Processed content length: %d", len(processedContent))
	log.Printf("Processed content preview: %s", processedContent[:min(300, len(processedContent))])
	finalHTML, err := rs.htmlBuilder.BuildCompleteHTML(
		processedContent, propagationData, chartData, sunGifHTML, folderPath)
	if err != nil {
		return "", fmt.Errorf("failed to build complete HTML: %w", err)
	}

	log.Printf("Report generation completed successfully (%d characters)", len(finalHTML))
	return finalHTML, nil
}

// GenerateStaticCSS generates static CSS content for saving to the report folder
func (rs *ReportService) GenerateStaticCSS() (string, error) {
	return rs.htmlBuilder.GenerateStaticCSS()
}

// generateCharts creates charts and returns template data
func (rs *ReportService) generateCharts(data *models.PropagationData, sourceData *models.SourceData, folderPath string) (*ChartTemplateData, error) {
	return rs.htmlBuilder.GenerateChartData(data, sourceData, folderPath)
}


// generateSunGif creates the Sun GIF HTML
func (rs *ReportService) generateSunGif(_ context.Context, folderPath string) (template.HTML, error) {
	// Generate Sun GIF path
	gifPath := "sun_72h.gif"
	if folderPath != "" {
		gifPath = fmt.Sprintf("/files/%s/sun_72h.gif", folderPath)
	}

	// Create HTML for Sun GIF
	sunHTML := fmt.Sprintf(`
<div class="sun-gif-container">
	<img src="%s" alt="72-hour Sun Animation" class="sun-gif" />
	<p class="sun-caption">72-hour solar activity animation from Helioviewer</p>
</div>`, gifPath)

	return template.HTML(sunHTML), nil
}

// GenerateHTMLWithSources generates HTML with source data
func (rs *ReportService) GenerateHTMLWithSources(markdownReport string, data *models.PropagationData, sourceData *models.SourceData) (string, error) {
	return rs.GenerateReport(context.Background(), data, sourceData, markdownReport, "")
}

// GenerateHTMLWithSourcesAndFolderPath generates HTML with folder path
func (rs *ReportService) GenerateHTMLWithSourcesAndFolderPath(markdownReport string, data *models.PropagationData, sourceData *models.SourceData, folderPath string) (string, error) {
	return rs.GenerateReport(context.Background(), data, sourceData, markdownReport, folderPath)
}
