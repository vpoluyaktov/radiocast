package reports

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"

	"radiocast/internal/config"
	"radiocast/internal/models"
	"radiocast/internal/charts"
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
	CSSFilePath              string
	Version                  string
	BackgroundImage          string
	
	// Chart placeholders
	SunGif                   template.HTML
	SolarActivityChart       template.HTML
	BandConditionsChart      template.HTML
	KIndexChart              template.HTML
	ForecastChart            template.HTML
	PropagationTimelineChart template.HTML
}

// ChartTemplateData represents chart data for template substitution
type ChartTemplateData struct {
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

// LoadStaticCSS loads the static CSS content without template processing
func (h *HTMLBuilder) LoadStaticCSS() (string, error) {
	// Load raw CSS content directly
	cssContent, err := h.templateLoader.LoadCSSStyles()
	if err != nil {
		return "", fmt.Errorf("failed to load CSS: %w", err)
	}
	return cssContent, nil
}

// getBackgroundImagePath returns the path for the background image based on deployment mode
func (h *HTMLBuilder) getBackgroundImagePath(folderPath string) string {
	if folderPath == "" {
		// Local mode - relative path
		return "background.png"
	}
	// GCS mode - use the folder path prefix
	return folderPath + "/background.png"
}

// getCSSFilePath returns the path for the CSS file based on deployment mode
func (h *HTMLBuilder) getCSSFilePath(folderPath string) string {
	if folderPath == "" {
		// Local mode - relative path
		return "styles.css"
	}
	// GCS mode - use the folder path prefix
	return folderPath + "/styles.css"
}

// GenerateStaticCSS returns the static CSS content
func (h *HTMLBuilder) GenerateStaticCSS() (string, error) {
	return h.LoadStaticCSS()
}

// GenerateChartData creates chart data using chart generators
func (h *HTMLBuilder) GenerateChartData(data *models.PropagationData, sourceData *models.SourceData, folderPath string) (*ChartTemplateData, error) {
	// Create chart generator
	chartGen := charts.NewChartGenerator(folderPath)
	
	// Generate chart snippets
	snippets, err := chartGen.GenerateEChartsSnippetsWithSources(data, sourceData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate chart snippets: %w", err)
	}

	// Create chart data with empty defaults
	chartData := &ChartTemplateData{
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
	chartData *ChartTemplateData,
	sunGifHTML template.HTML,
	folderPath string) (string, error) {

	log.Println("Building complete HTML...")

	// Use the already processed HTML content directly (no markdown conversion needed)
	htmlContent := processedHTMLContent
	
	// Debug: Log the processed HTML content
	log.Printf("Processed HTML content length: %d", len(htmlContent))
	log.Printf("HTML content preview: %s", htmlContent[:min(200, len(htmlContent))])

	// Get CSS file path (CSS will be saved separately as static file)
	cssFilePath := h.getCSSFilePath(folderPath)

	// Prepare template data
	templateData := TemplateData{
		Date:                     time.Now().Format("2006-01-02"),
		GeneratedAt:              time.Now().Format("2006-01-02 15:04:05 UTC"),
		Content:                  template.HTML(htmlContent),
		CSSFilePath:              cssFilePath,
		Version:                  config.GetVersion(),
		BackgroundImage:          h.getBackgroundImagePath(folderPath),
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

	// Update asset URLs for deployment mode (excluding background image which is already handled in CSS)
	if folderPath != "" {
		finalHTML = h.updateAssetURLs(finalHTML, folderPath)
	}

	log.Printf("Complete HTML built successfully (%d characters)", len(finalHTML))
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
		"safeCSS": func(s string) template.CSS {
			return template.CSS(s)
		},
	}).Parse(htmlTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Debug: Log template data before execution
	log.Printf("Template data Content length: %d", len(string(data.Content)))
	log.Printf("Template data Content preview: %s", string(data.Content)[:min(200, len(string(data.Content)))])

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	result := buf.String()
	log.Printf("Template execution result length: %d", len(result))
	log.Printf("Template result preview: %s", result[:min(300, len(result))])

	return result, nil
}

// generateBandTable creates the band analysis table HTML
func (h *HTMLBuilder) generateBandTable(data *models.PropagationData) template.HTML {
	if data == nil {
		return template.HTML("")
	}

	var buf strings.Builder
	buf.WriteString(`<table class="band-analysis-table">`)
	buf.WriteString(`<thead><tr><th>Band</th><th>Day Condition</th><th>Night Condition</th><th>Best Times</th><th>Notes</th></tr></thead>`)
	buf.WriteString(`<tbody>`)

	bands := []struct {
		name      string
		condition models.BandCondition
		bestTimes string
		notes     string
	}{
		{"80m", data.BandData.Band80m, "2200-0600 UTC", "Low noise after sunset"},
		{"40m", data.BandData.Band40m, "2100-0700 UTC", "Best DX band at night"},
		{"20m", data.BandData.Band20m, "1000-2200 UTC", "Reliable all day"},
		{"17m", data.BandData.Band17m, "1200-2000 UTC", "Good for EU/NA"},
		{"15m", data.BandData.Band15m, "1400-1800 UTC", "Solar dependent"},
		{"12m", data.BandData.Band12m, "1500-1700 UTC", "Sporadic openings"},
		{"10m", data.BandData.Band10m, "1600-1700 UTC", "Solar cycle dependent"},
	}

	for _, band := range bands {
		buf.WriteString(fmt.Sprintf(`<tr>
			<td>%s</td>
			<td>%s</td>
			<td>%s</td>
			<td>%s</td>
			<td>%s</td>
		</tr>`, 
			band.name,
			h.formatCondition(band.condition.Day),
			h.formatCondition(band.condition.Night),
			template.HTMLEscapeString(band.bestTimes),
			template.HTMLEscapeString(band.notes),
		))
	}

	buf.WriteString(`</tbody></table>`)
	return template.HTML(buf.String())
}

// formatCondition formats a condition string with emoji
func (h *HTMLBuilder) formatCondition(condition string) string {
	switch strings.ToLower(strings.TrimSpace(condition)) {
	case "excellent":
		return "🟢 Excellent"
	case "good":
		return "🟡 Good"
	case "fair":
		return "🟠 Fair"
	case "poor":
		return "🔴 Poor"
	case "closed":
		return "⚫ Closed"
	default:
		return template.HTMLEscapeString(condition)
	}
}

// updateAssetURLs updates asset URLs based on deployment mode
func (h *HTMLBuilder) updateAssetURLs(html, folderPath string) string {
	// Update background image URLs in inline styles for GCS deployment
	html = strings.ReplaceAll(html, 
		"url('background.png')",
		fmt.Sprintf("url('/files/%s/background.png')", folderPath))
	
	html = strings.ReplaceAll(html,
		`url("background.png")`,
		fmt.Sprintf(`url("/files/%s/background.png")`, folderPath))

	return html
}

// ProcessMarkdownWithPlaceholders processes markdown content and substitutes template placeholders
func (h *HTMLBuilder) ProcessMarkdownWithPlaceholders(
	markdownContent string,
	chartData *ChartTemplateData,
	sunGifHTML template.HTML,
	bandTableHTML template.HTML) (string, error) {

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
