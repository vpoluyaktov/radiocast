package reports

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ChartHTMLBuilder handles chart HTML generation
type ChartHTMLBuilder struct{}

// NewChartHTMLBuilder creates a new chart HTML builder
func NewChartHTMLBuilder() *ChartHTMLBuilder {
	return &ChartHTMLBuilder{}
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
