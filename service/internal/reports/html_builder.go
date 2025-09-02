package reports

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/russross/blackfriday/v2"

	"radiocast/internal/config"
	"radiocast/internal/models"
)

// HTMLBuilder handles HTML generation and template processing
type HTMLBuilder struct {
	templateLoader *TemplateLoader
}

// NewHTMLBuilder creates a new HTML builder
func NewHTMLBuilder() *HTMLBuilder {
	return &HTMLBuilder{
		templateLoader: NewTemplateLoader(),
	}
}

// MarkdownToHTML converts markdown to HTML using blackfriday
func (h *HTMLBuilder) MarkdownToHTML(markdownText string) string {
	htmlBytes := blackfriday.Run([]byte(markdownText))
	return string(htmlBytes)
}

// ConvertMarkdownToHTML converts markdown content to a complete HTML document using configurable templates
func (h *HTMLBuilder) ConvertMarkdownToHTML(markdownContent string, date string) (string, error) {
	// Convert markdown to HTML using blackfriday
	htmlBytes := blackfriday.Run([]byte(markdownContent))
	htmlContent := string(htmlBytes)
	
	// Load HTML template
	htmlTemplate, err := h.templateLoader.LoadHTMLTemplate()
	if err != nil {
		return "", fmt.Errorf("failed to load HTML template: %w", err)
	}
	
	// Load CSS styles
	cssStyles, err := h.templateLoader.LoadCSSStyles()
	if err != nil {
		return "", fmt.Errorf("failed to load CSS styles: %w", err)
	}
	
	// Parse the HTML template with proper functions for unescaped content
	tmpl, err := template.New("report").Funcs(template.FuncMap{
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"safeCSS": func(s string) template.CSS {
			return template.CSS(s)
		},
	}).Parse(htmlTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML template: %w", err)
	}
	
	// Prepare template data
	templateData := struct {
		Date        string
		GeneratedAt string
		Content     template.HTML
		CSSStyles   template.CSS
		Charts      template.HTML
		Version     string
	}{
		Date:        date,
		GeneratedAt: time.Now().Format("2006-01-02 15:04:05 UTC"),
		Content:     template.HTML(htmlContent),
		CSSStyles:   template.CSS(cssStyles),
		Charts:      template.HTML(""), // Charts will be embedded in content
		Version:     config.GetVersion(),
	}
	
	// Execute the template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}
	
	return buf.String(), nil
}

// ConvertMarkdownToHTMLWithCharts converts markdown content to HTML with charts
func (h *HTMLBuilder) ConvertMarkdownToHTMLWithCharts(markdownContent string, charts string, date string) (string, error) {
	// Convert markdown to HTML using blackfriday
	htmlBytes := blackfriday.Run([]byte(markdownContent))
	htmlContent := string(htmlBytes)
	
	// Load HTML template
	htmlTemplate, err := h.templateLoader.LoadHTMLTemplate()
	if err != nil {
		return "", fmt.Errorf("failed to load HTML template: %w", err)
	}
	
	// Load CSS styles
	cssStyles, err := h.templateLoader.LoadCSSStyles()
	if err != nil {
		return "", fmt.Errorf("failed to load CSS styles: %w", err)
	}
	
	// Parse the HTML template with proper functions for unescaped content
	tmpl, err := template.New("report").Funcs(template.FuncMap{
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"safeCSS": func(s string) template.CSS {
			return template.CSS(s)
		},
	}).Parse(htmlTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML template: %w", err)
	}
	
	// Prepare template data with charts
	templateData := struct {
		Date        string
		GeneratedAt string
		Content     template.HTML
		CSSStyles   template.CSS
		Charts      template.HTML
		Version     string
	}{
		Date:        date,
		GeneratedAt: time.Now().Format("2006-01-02 15:04:05 UTC"),
		Content:     template.HTML(htmlContent),
		CSSStyles:   template.CSS(cssStyles),
		Charts:      template.HTML(charts), // Now properly populated with charts
		Version:     config.GetVersion(),
	}
	
	// Execute the template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}
	
	return buf.String(), nil
}

// BuildCompleteHTML creates a complete HTML document
func (h *HTMLBuilder) BuildCompleteHTML(content, charts string, data *models.PropagationData) (string, error) {
	// Integrate charts throughout the content instead of at the end
	integratedContent := h.integrateChartsInContent(content, charts)
	
	// Use the new template-based conversion without separate charts section
	result, err := h.ConvertMarkdownToHTMLWithCharts(integratedContent, "", time.Now().Format("2006-01-02"))
	if err != nil {
		return "", err
	}
	return result, nil
}

// integrateChartsInContent replaces chart placeholders with actual chart HTML
func (h *HTMLBuilder) integrateChartsInContent(content, charts string) string {
	// Parse chart HTML to extract individual chart elements
	chartMap := h.parseChartsHTML(charts)
	
	// Replace placeholders with actual chart HTML
	integratedContent := content
	
	// Replace chart placeholders with professional chart sections
	for placeholder, chartHTML := range chartMap {
		chartSection := fmt.Sprintf(`
<div class="chart-section">
	<div class="chart-container-integrated">
		%s
	</div>
</div>`, chartHTML)
		integratedContent = strings.Replace(integratedContent, placeholder, chartSection, -1)
	}
	
	return integratedContent
}

// parseChartsHTML extracts individual charts from the charts HTML string
func (h *HTMLBuilder) parseChartsHTML(charts string) map[string]string {
	chartMap := make(map[string]string)
	
	if charts == "" {
		return chartMap
	}
	
	// Parse chart containers more robustly
	lines := strings.Split(charts, "\n")
	var currentChart strings.Builder
	var chartTitle string
	inChartContainer := false
	
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		
		// Start of a chart container
		if strings.Contains(trimmedLine, "chart-container") && strings.Contains(trimmedLine, "<div") {
			// Save previous chart if exists
			if currentChart.Len() > 0 && chartTitle != "" {
				h.mapChartToPlaceholder(chartTitle, currentChart.String(), chartMap)
			}
			// Reset for new chart
			currentChart.Reset()
			chartTitle = ""
			inChartContainer = true
		}
		
		// Extract chart title from h3 tags
		if inChartContainer && strings.Contains(trimmedLine, "<h3>") {
			start := strings.Index(trimmedLine, ">") + 1
			end := strings.LastIndex(trimmedLine, "<")
			if start > 0 && end > start {
				chartTitle = strings.TrimSpace(trimmedLine[start:end])
			}
		}
		
		// Add line to current chart if we're inside a container
		if inChartContainer {
			currentChart.WriteString(line + "\n")
		}
		
		// End of chart container
		if inChartContainer && strings.Contains(trimmedLine, "</div>") && 
		   (strings.Contains(currentChart.String(), "chart-container") || strings.Contains(currentChart.String(), "img")) {
			// This might be the end of the chart container
			if chartTitle != "" {
				h.mapChartToPlaceholder(chartTitle, currentChart.String(), chartMap)
			}
			inChartContainer = false
		}
	}
	
	// Handle last chart if still processing
	if currentChart.Len() > 0 && chartTitle != "" {
		h.mapChartToPlaceholder(chartTitle, currentChart.String(), chartMap)
	}
	
	return chartMap
}

// mapChartToPlaceholder maps chart titles to their placeholders
func (h *HTMLBuilder) mapChartToPlaceholder(title, chartHTML string, chartMap map[string]string) {
	switch {
	case strings.Contains(title, "Solar Activity"):
		chartMap["{{SOLAR_ACTIVITY_CHART}}"] = chartHTML
	case strings.Contains(title, "K Index") || strings.Contains(title, "K-Index"):
		chartMap["{{K_INDEX_CHART}}"] = chartHTML
	case strings.Contains(title, "Band Conditions"):
		chartMap["{{BAND_CONDITIONS_CHART}}"] = chartHTML
	case strings.Contains(title, "Forecast"):
		chartMap["{{FORECAST_CHART}}"] = chartHTML
	case strings.Contains(title, "Propagation Timeline"):
		chartMap["{{PROPAGATION_TIMELINE_CHART}}"] = chartHTML
	}
}
