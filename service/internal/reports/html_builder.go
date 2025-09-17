package reports

import (
	"bytes"
	"fmt"
	"html/template"
	"time"

	"radiocast/internal/charts"
	"radiocast/internal/config"
	"radiocast/internal/logger"
	"radiocast/internal/models"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

// HTMLBuilder handles HTML generation with goldmark
type HTMLBuilder struct {
	templateLoader *TemplateLoader
	goldmark       goldmark.Markdown
}

// NewHTMLBuilder creates an HTML builder
func NewHTMLBuilder() *HTMLBuilder {
	// Configure goldmark with extensions
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithUnsafe(), // Allow raw HTML in markdown
		),
	)

	return &HTMLBuilder{
		templateLoader: NewTemplateLoader(),
		goldmark:       md,
	}
}

// TemplateData represents the data structure for the HTML template
type TemplateData struct {
	Date                     string
	GeneratedAt              string
	Content                  template.HTML
	Version                  string
	
	// Chart placeholders
	SunGif                   template.HTML
	SolarActivityChart       template.HTML
	BandConditionsChart      template.HTML
	KIndexChart              template.HTML
	ForecastChart            template.HTML
	PropagationTimelineChart template.HTML
}

// ConvertMarkdownToHTML converts markdown to HTML using goldmark
func (h *HTMLBuilder) ConvertMarkdownToHTML(markdownContent string) (string, error) {
	var buf bytes.Buffer
	if err := h.goldmark.Convert([]byte(markdownContent), &buf); err != nil {
		return "", fmt.Errorf("failed to convert markdown: %w", err)
	}
	return buf.String(), nil
}




// GenerateChartData creates chart data using chart generators
func (h *HTMLBuilder) GenerateChartData(data *models.PropagationData, sourceData *models.SourceData, folderPath string) (*TemplateData, error) {
	// Create chart generator
	chartGen := charts.NewChartGenerator(folderPath)
	
	// Generate chart snippets
	snippets, err := chartGen.GenerateEChartsSnippetsWithSources(data, sourceData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate chart snippets: %w", err)
	}

	// Create chart data with empty defaults
	chartData := &TemplateData{
		SolarActivityChart:       template.HTML(""),
		BandConditionsChart:      template.HTML(""),
		KIndexChart:              template.HTML(""),
		ForecastChart:            template.HTML(""),
		PropagationTimelineChart: template.HTML(""),
	}

	// Map snippets by ID to template data
	for _, snippet := range snippets {
		switch snippet.ID {
		case "chart-solar-activity":
			chartData.SolarActivityChart = template.HTML(snippet.HTML)
		case "chart-band-conditions":
			chartData.BandConditionsChart = template.HTML(snippet.HTML)
		case "chart-geomagnetic-conditions":
			chartData.KIndexChart = template.HTML(snippet.HTML)
		case "chart-forecast":
			chartData.ForecastChart = template.HTML(snippet.HTML)
		case "chart-propagation-timeline":
			chartData.PropagationTimelineChart = template.HTML(snippet.HTML)
		}
	}

	return chartData, nil
}

// BuildCompleteHTML creates a complete HTML document with template substitution
func (h *HTMLBuilder) BuildCompleteHTML(
	processedHTMLContent string,
	data *models.PropagationData,
	chartData *TemplateData,
	sunGifHTML template.HTML,
	folderPath string) (string, error) {

	logger.Info("Building complete HTML...")

	// Use the already processed HTML content directly (no markdown conversion needed)
	htmlContent := processedHTMLContent
	
	// Debug: Log the processed HTML content
	logger.Debugf("Processed HTML content length: %d", len(htmlContent))
	logger.Debugf("HTML content preview: %s", htmlContent[:min(200, len(htmlContent))])

	// Prepare template data
	templateData := TemplateData{
		Date:                     time.Now().Format("2006-01-02"),
		GeneratedAt:              time.Now().Format("2006-01-02 15:04:05 UTC"),
		Content:                  template.HTML(htmlContent),
		Version:                  config.GetVersion(),
		SunGif:                   sunGifHTML,
		SolarActivityChart:       chartData.SolarActivityChart,
		BandConditionsChart:      chartData.BandConditionsChart,
		KIndexChart:              chartData.KIndexChart,
		ForecastChart:            chartData.ForecastChart,
		PropagationTimelineChart: chartData.PropagationTimelineChart,
	}

	// Execute template
	finalHTML, err := h.executeTemplate(templateData)
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}


	logger.Debugf("Complete HTML built successfully (%d characters)", len(finalHTML))
	return finalHTML, nil
}


// executeTemplate executes the HTML template with the provided data
func (h *HTMLBuilder) executeTemplate(data TemplateData) (string, error) {
	// Load HTML template
	htmlTemplate, err := h.templateLoader.LoadHTMLTemplate()
	if err != nil {
		return "", fmt.Errorf("failed to load HTML template: %w", err)
	}

	// Parse template with functions
	tmpl, err := template.New("report").Funcs(template.FuncMap{
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
	}).Parse(htmlTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Debug: Log template data before execution
	logger.Debugf("Template data Content length: %d", len(string(data.Content)))

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	result := buf.String()

	return result, nil
}



// ProcessMarkdownWithPlaceholders processes markdown content and substitutes template placeholders
func (h *HTMLBuilder) ProcessMarkdownWithPlaceholders(
	markdownContent string,
	chartData *TemplateData,
	sunGifHTML template.HTML) (string, error) {

	// First convert markdown to HTML
	htmlContent, err := h.ConvertMarkdownToHTML(markdownContent)
	if err != nil {
		return "", err
	}

	// Create a template from the HTML content and execute it with data
	tmpl, err := template.New("content").Parse(htmlContent)
	if err != nil {
		return "", fmt.Errorf("failed to parse content template: %w", err)
	}

	// Prepare data for placeholder substitution
	data := struct {
		SunGif                   template.HTML
		SolarActivityChart       template.HTML
		BandConditionsChart      template.HTML
		KIndexChart              template.HTML
		ForecastChart            template.HTML
		PropagationTimelineChart template.HTML
	}{
		SunGif:                   sunGifHTML,
		SolarActivityChart:       chartData.SolarActivityChart,
		BandConditionsChart:      chartData.BandConditionsChart,
		KIndexChart:              chartData.KIndexChart,
		ForecastChart:            chartData.ForecastChart,
		PropagationTimelineChart: chartData.PropagationTimelineChart,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute content template: %w", err)
	}

	return buf.String(), nil
}
