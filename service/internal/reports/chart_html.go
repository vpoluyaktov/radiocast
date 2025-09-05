package reports

import (
	"fmt"
	"path/filepath"
	"strings"

	"radiocast/internal/charts"
)

// ChartHTMLBuilder handles chart HTML generation
type ChartHTMLBuilder struct{}

// NewChartHTMLBuilder creates a new chart HTML builder
func NewChartHTMLBuilder() *ChartHTMLBuilder {
	return &ChartHTMLBuilder{}
}

// BuildEChartsHTML creates HTML for go-echarts charts using local asset path.
// folderPath is empty for local mode and non-empty for GCS mode (used to prefix the /files route).
// This method ensures echarts.min.js is loaded once from the proxied /files path (no CDN).
func (c *ChartHTMLBuilder) BuildEChartsHTML(snippets []charts.ChartSnippet, folderPath string) string {
	if len(snippets) == 0 {
		return "<p>No charts available</p>"
	}

	var html strings.Builder
	html.WriteString("<div class=\"charts-section\">\n")
	html.WriteString("<h2>Charts and Analysis</h2>\n")
	html.WriteString("<div class=\"charts-grid\">\n")

	for _, sn := range snippets {
		// Title above each chart div for consistency with previous layout
		html.WriteString("\t<div class=\"chart-container\">\n")
		if sn.Title != "" {
			html.WriteString(fmt.Sprintf("\t\t<h3>%s</h3>\n", sn.Title))
		}
		// Insert the chart container div
		html.WriteString("\t\t" + sn.Div + "\n")
		html.WriteString("\t</div>\n")
	}

	html.WriteString("</div>\n")

	// Load ECharts from public CDN instead of local files
	const cdnPath = "https://cdn.jsdelivr.net/npm/echarts@5.4.3/dist/echarts.min.js"
	html.WriteString(fmt.Sprintf("<script src=\"%s\"></script>\n", cdnPath))

	// Append all chart init scripts
	for _, sn := range snippets {
		html.WriteString(sn.Script)
		if !strings.HasSuffix(sn.Script, "\n") {
			html.WriteString("\n")
		}
	}

	html.WriteString("</div>\n")
	return html.String()
}

// BuildChartsHTML creates HTML for chart images using proxy URLs
func (c *ChartHTMLBuilder) BuildChartsHTML(chartFiles []string, folderPath string) string {
	if len(chartFiles) == 0 {
		return "<p>No charts available</p>"
	}
	
	var html strings.Builder
	html.WriteString("<div class=\"charts-section\">\n")
	html.WriteString("<h2>Charts and Analysis</h2>\n")
	html.WriteString("<div class=\"charts-grid\">\n")
	
	for _, chartFile := range chartFiles {
		// Extract filename from path for display
		filename := filepath.Base(chartFile)
		// Remove file extension for title
		title := strings.TrimSuffix(filename, filepath.Ext(filename))
		// Convert underscores to spaces and title case
		title = strings.ReplaceAll(title, "_", " ")
		title = ToTitleCase(title)
		
		// Build proxy URL path
		var imageSrc string
		if folderPath != "" {
			// For GCS deployment, use proxy URL with folder path
			imageSrc = fmt.Sprintf("/files/%s/%s", folderPath, filename)
		} else {
			// For local deployment, use proxy URL with just filename
			imageSrc = fmt.Sprintf("/files/%s", filename)
		}
		
		html.WriteString(fmt.Sprintf(`
		<div class="chart-container">
			<h3>%s</h3>
			<img src="%s" alt="%s" class="chart-image">
		</div>
		`, title, imageSrc, title))
	}
	
	html.WriteString("</div>\n")
	html.WriteString("</div>\n")
	
	return html.String()
}

// BuildChartsHTMLFromURLs creates HTML for chart images using provided URLs
func (c *ChartHTMLBuilder) BuildChartsHTMLFromURLs(chartURLs []string) string {
	if len(chartURLs) == 0 {
		return "<p>No charts available</p>"
	}
	
	var html strings.Builder
	html.WriteString("<div class=\"charts-section\">\n")
	html.WriteString("<h2>Charts and Analysis</h2>\n")
	html.WriteString("<div class=\"charts-grid\">\n")
	
	// Define chart titles in expected order
	chartTitles := map[string]string{
		"solar_activity.png":   "Solar Activity",
		"k_index_trend.png":    "K Index Trend", 
		"band_conditions.png":  "Band Conditions",
		"forecast.png":         "Forecast",
	}
	
	for _, chartURL := range chartURLs {
		// Extract filename from URL for title lookup
		filename := filepath.Base(chartURL)
		title, exists := chartTitles[filename]
		if !exists {
			// Fallback: convert filename to title
			title = strings.TrimSuffix(filename, filepath.Ext(filename))
			title = strings.ReplaceAll(title, "_", " ")
			title = ToTitleCase(title)
		}
		
		html.WriteString(fmt.Sprintf(`
		<div class="chart-container">
			<h3>%s</h3>
			<img src="%s" alt="%s" class="chart-image">
		</div>
		`, title, chartURL, title))
	}
	
	html.WriteString("</div>\n")
	html.WriteString("</div>\n")
	
	return html.String()
}
