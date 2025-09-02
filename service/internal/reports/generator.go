package reports

import (
	"fmt"
	"log"

	"radiocast/internal/models"
)

// Generator handles report generation and HTML conversion
type Generator struct {
	outputDir        string
	htmlBuilder      *HTMLBuilder
	chartHTMLBuilder *ChartHTMLBuilder
}

// NewGenerator creates a new report generator
func NewGenerator(outputDir string) *Generator {
	return &Generator{
		outputDir:        outputDir,
		htmlBuilder:      NewHTMLBuilder(),
		chartHTMLBuilder: NewChartHTMLBuilder(),
	}
}


// GenerateHTML converts markdown report to HTML with embedded charts (backward compatibility)
func (g *Generator) GenerateHTML(markdownReport string, data *models.PropagationData) (string, error) {
	return g.GenerateHTMLWithSources(markdownReport, data, nil)
}

// GenerateHTMLWithSources converts markdown report to HTML with embedded charts using source data
func (g *Generator) GenerateHTMLWithSources(markdownReport string, data *models.PropagationData, sourceData *models.SourceData) (string, error) {
	log.Println("Converting markdown to HTML and generating charts...")
	
	// Generate chart images using the new chart generator with source data
	chartGen := NewChartGenerator(g.outputDir)
	chartFiles, err := chartGen.GenerateChartsWithSources(data, sourceData)
	if err != nil {
		log.Printf("Warning: Failed to generate charts: %v", err)
		chartFiles = []string{}
	}
	
	// Convert markdown to HTML
	htmlContent := g.htmlBuilder.MarkdownToHTML(markdownReport)
	
	// Build chart HTML references (empty folderPath for now - will be updated for GCS)
	chartsHTML := g.chartHTMLBuilder.BuildChartsHTML(chartFiles, "")
	
	// Combine everything into a complete HTML document
	fullHTML, err := g.htmlBuilder.BuildCompleteHTML(htmlContent, chartsHTML, data)
	if err != nil {
		return "", fmt.Errorf("failed to build complete HTML: %w", err)
	}
	
	log.Printf("Generated complete HTML report with %d characters and %d charts", len(fullHTML), len(chartFiles))
	return fullHTML, nil
}

// GenerateHTMLWithLocalCharts converts markdown report to HTML with pre-generated local chart files
func (g *Generator) GenerateHTMLWithLocalCharts(markdownReport string, data *models.PropagationData, chartFiles []string) (string, error) {
	log.Println("Converting markdown to HTML with pre-generated charts...")
	
	// Convert markdown to HTML
	htmlContent := g.htmlBuilder.MarkdownToHTML(markdownReport)
	
	// Build chart HTML references (empty folderPath for local mode)
	chartsHTML := g.chartHTMLBuilder.BuildChartsHTML(chartFiles, "")
	
	// Combine everything into a complete HTML document
	fullHTML, err := g.htmlBuilder.BuildCompleteHTML(htmlContent, chartsHTML, data)
	if err != nil {
		return "", fmt.Errorf("failed to build complete HTML: %w", err)
	}
	
	log.Printf("Generated complete HTML report with %d characters and %d charts", len(fullHTML), len(chartFiles))
	return fullHTML, nil
}

// GenerateHTMLWithFolderPath converts markdown report to HTML with pre-generated chart files using folder path
func (g *Generator) GenerateHTMLWithFolderPath(markdownReport string, data *models.PropagationData, chartFiles []string, folderPath string) (string, error) {
	log.Printf("Converting markdown to HTML with %d charts using folder path: %s", len(chartFiles), folderPath)
	
	// Convert markdown to HTML
	htmlContent := g.htmlBuilder.MarkdownToHTML(markdownReport)
	
	// Build chart HTML references with folder path for GCS URLs
	chartsHTML := g.chartHTMLBuilder.BuildChartsHTML(chartFiles, folderPath)
	
	// Combine everything into a complete HTML document
	fullHTML, err := g.htmlBuilder.BuildCompleteHTML(htmlContent, chartsHTML, data)
	if err != nil {
		return "", fmt.Errorf("failed to build complete HTML: %w", err)
	}
	
	log.Printf("Generated complete HTML report with %d characters and %d charts with folder path", len(fullHTML), len(chartFiles))
	return fullHTML, nil
}

// GenerateHTMLWithChartURLs converts markdown report to HTML using provided chart URLs
func (g *Generator) GenerateHTMLWithChartURLs(markdownReport string, data *models.PropagationData, chartURLs []string) (string, error) {
	log.Printf("Converting markdown to HTML with %d provided chart URLs...", len(chartURLs))
	
	// Convert markdown to HTML
	htmlContent := g.htmlBuilder.MarkdownToHTML(markdownReport)
	
	// Build chart HTML references using provided URLs
	chartsHTML := g.chartHTMLBuilder.BuildChartsHTMLFromURLs(chartURLs)
	
	// Combine everything into a complete HTML document
	fullHTML, err := g.htmlBuilder.BuildCompleteHTML(htmlContent, chartsHTML, data)
	if err != nil {
		return "", fmt.Errorf("failed to build complete HTML: %w", err)
	}
	
	log.Printf("Generated complete HTML report with %d characters and %d chart URLs", len(fullHTML), len(chartURLs))
	return fullHTML, nil
}

// MarkdownToHTML converts markdown to HTML (delegated method for backward compatibility)
func (g *Generator) MarkdownToHTML(markdownText string) string {
	return g.htmlBuilder.MarkdownToHTML(markdownText)
}

// BuildChartsHTML creates HTML for chart images (delegated method for backward compatibility)
func (g *Generator) BuildChartsHTML(chartFiles []string, folderPath string) string {
	return g.chartHTMLBuilder.BuildChartsHTML(chartFiles, folderPath)
}

// BuildCompleteHTML creates a complete HTML document (delegated method for backward compatibility)
func (g *Generator) BuildCompleteHTML(content, charts string, data *models.PropagationData) (string, error) {
	return g.htmlBuilder.BuildCompleteHTML(content, charts, data)
}


