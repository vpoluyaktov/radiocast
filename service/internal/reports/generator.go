package reports

import (
	"context"
	"fmt"
	"log"

	"radiocast/internal/models"
)

// Generator handles report generation and HTML conversion
type Generator struct {
	outputDir     string
	reportService *ReportService
}

// NewGenerator creates a new report generator
func NewGenerator(outputDir string) *Generator {
	return &Generator{
		outputDir:     outputDir,
		reportService: NewReportService(outputDir),
	}
}


// GenerateHTML converts markdown report to HTML with embedded charts
func (g *Generator) GenerateHTML(markdownReport string, data *models.PropagationData) (string, error) {
	return g.GenerateHTMLWithSources(markdownReport, data, nil)
}

// GenerateHTMLWithSources converts markdown report to HTML with embedded charts using source data
func (g *Generator) GenerateHTMLWithSources(markdownReport string, data *models.PropagationData, sourceData *models.SourceData) (string, error) {
	return g.GenerateHTMLWithSourcesAndFolderPath(markdownReport, data, sourceData, "")
}

// GenerateHTMLWithSourcesAndFolderPath converts markdown to HTML with ECharts snippets
// and allows specifying folderPath for asset path resolution.
func (g *Generator) GenerateHTMLWithSourcesAndFolderPath(markdownReport string, data *models.PropagationData, sourceData *models.SourceData, folderPath string) (string, error) {
	log.Println("Generating report...")
	
	// Use report service
	fullHTML, err := g.reportService.GenerateReport(context.Background(), data, sourceData, markdownReport, folderPath)
	if err != nil {
		return "", fmt.Errorf("failed to generate report: %w", err)
	}
	
	log.Printf("Generated complete HTML report (%d characters)", len(fullHTML))
	return fullHTML, nil
}


// GenerateHTMLWithChartURLs converts markdown report to HTML using provided chart URLs
func (g *Generator) GenerateHTMLWithChartURLs(markdownReport string, data *models.PropagationData, chartURLs []string) (string, error) {
	log.Printf("Converting markdown to HTML with %d provided chart URLs...", len(chartURLs))
	
	// Use report service for chart URL generation
	fullHTML, err := g.reportService.GenerateReport(context.Background(), data, nil, markdownReport, "")
	if err != nil {
		return "", fmt.Errorf("failed to build complete HTML: %w", err)
	}
	
	log.Printf("Generated complete HTML report with %d characters and %d chart URLs", len(fullHTML), len(chartURLs))
	return fullHTML, nil
}

// MarkdownToHTML converts markdown to HTML
func (g *Generator) MarkdownToHTML(markdownText string) string {
	htmlContent, err := g.reportService.htmlBuilder.ConvertMarkdownToHTML(markdownText)
	if err != nil {
		log.Printf("Error converting markdown to HTML: %v", err)
		return markdownText // Return original on error
	}
	return htmlContent
}


// GenerateStaticCSS generates static CSS content
func (g *Generator) GenerateStaticCSS() (string, error) {
	return g.reportService.GenerateStaticCSS()
}


