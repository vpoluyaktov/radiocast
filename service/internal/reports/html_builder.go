package reports

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/russross/blackfriday/v2"

	"radiocast/internal/models"
	"radiocast/internal/config"
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
	htmlContent := string(htmlBytes)
	
	// Find the band analysis section
	if idx := strings.Index(htmlContent, "<h2>📻 Band-by-Band Analysis</h2>"); idx != -1 {
		// Look for the first table after the band analysis heading
		afterHeading := htmlContent[idx:]
		
		// The blackfriday library sometimes wraps tables in <p> tags
		// First, check if we have a <p><table pattern
		tableStartIdx := strings.Index(afterHeading, "<p><table")
		if tableStartIdx != -1 {
			// Found a table wrapped in <p> tags
			tableStartIdx += idx + 3 // Add 3 to skip the <p> tag
			
			// Get the part before and after the table tag
			partBeforeTable := htmlContent[:tableStartIdx]
			partAfterTable := htmlContent[tableStartIdx:]
			
			// Replace the first occurrence of <table, but check if it already has the class
			if !strings.Contains(partAfterTable[:50], "band-analysis-table") {
				// Replace just the first occurrence of <table
				partAfterTable = strings.Replace(partAfterTable, "<table", "<table class=\"band-analysis-table\"", 1)
				htmlContent = partBeforeTable + partAfterTable
			} else {
				// If it already has the class, make sure it doesn't have duplicates
				partAfterTable = strings.Replace(partAfterTable, "class=\"band-analysis-table\" class=\"band-analysis-table\"", "class=\"band-analysis-table\"", 1)
				htmlContent = partBeforeTable + partAfterTable
			}
		} else {
			// Try to find a regular <table> tag
			tableStartIdx = strings.Index(afterHeading, "<table")
			if tableStartIdx != -1 {
				// Calculate absolute position
				tableStartIdx += idx
				
				// Get the part before and after the table tag
				partBeforeTable := htmlContent[:tableStartIdx]
				partAfterTable := htmlContent[tableStartIdx:]
				
				// Check if the table already has our class
				if !strings.Contains(partAfterTable[:50], "band-analysis-table") {
					// Replace just the first occurrence of <table
					partAfterTable = strings.Replace(partAfterTable, "<table", "<table class=\"band-analysis-table\"", 1)
					htmlContent = partBeforeTable + partAfterTable
				} else {
					// If it already has the class, make sure it doesn't have duplicates
					partAfterTable = strings.Replace(partAfterTable, "class=\"band-analysis-table\" class=\"band-analysis-table\"", "class=\"band-analysis-table\"", 1)
					htmlContent = partBeforeTable + partAfterTable
				}
			}
		}
	}
	
	return htmlContent
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
func (h *HTMLBuilder) ConvertMarkdownToHTMLWithCharts(content string, charts string, date string) (string, error) {
	// Note: content may already be HTML if we're calling from BuildCompleteHTML
	// We'll skip the markdown conversion in that case
	htmlContent := content
	
	// If the content doesn't look like HTML, convert it from markdown
	if !strings.Contains(content, "<p>") && !strings.Contains(content, "<div>") {
		htmlBytes := blackfriday.Run([]byte(content))
		htmlContent = string(htmlBytes)
	}
	
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
	
	// Clean up any potential literal HTML tags that might have been incorrectly parsed
	htmlContent = strings.Replace(htmlContent, "&lt;/div&gt;", "", -1)
	htmlContent = strings.Replace(htmlContent, "&lt;div&gt;", "", -1)
	
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
		Charts:      template.HTML(charts),
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
	// First convert markdown to HTML to ensure proper HTML structure
	htmlContent := h.MarkdownToHTML(content)
	
	// Then integrate charts throughout the content
	integratedContent := h.integrateChartsInContent(htmlContent, charts)
	
	// Use the template-based conversion without separate charts section
	result, err := h.ConvertMarkdownToHTMLWithCharts(integratedContent, "", time.Now().Format("2006-01-02"))
	if err != nil {
		return "", err
	}
	return result, nil
}

// Chart represents a chart with its ID, title, HTML content, and placeholder
type Chart struct {
	ID         string
	Title      string
	HTML       string
	Placeholder string
}

// integrateChartsInContent replaces chart placeholders with actual chart HTML
func (h *HTMLBuilder) integrateChartsInContent(content, charts string) string {
	// Extract all charts from the HTML
	extractedCharts := h.extractCharts(charts)
	
	// Map charts to their placeholders
	chartMap := h.mapChartsToPlaceholders(extractedCharts)
	
	// Extract the ECharts script and initialization scripts from the charts HTML
	echartsScript, initScripts := h.extractScripts(charts)
	
	// Replace placeholders with chart HTML
	integratedContent := h.replacePlaceholders(content, chartMap)
	
	// Add scripts to the HTML
	integratedContent = h.addScriptsToHTML(integratedContent, echartsScript, initScripts)
	
	return integratedContent
}

// extractCharts extracts all charts from the HTML
func (h *HTMLBuilder) extractCharts(charts string) []Chart {
	if charts == "" {
		return nil
	}
	
	var extractedCharts []Chart
	
	// Use regex to find all chart containers
	chartContainerRegex := regexp.MustCompile(`<div[^>]*chart-container[^>]*>[\s\S]*?<h3>([^<]+)</h3>[\s\S]*?<div[^>]*id="(chart-[^"]+)"[^>]*>[\s\S]*?</div>[\s\S]*?</div>`)
	matches := chartContainerRegex.FindAllStringSubmatch(charts, -1)
	
	processedIDs := make(map[string]bool)
	
	for _, match := range matches {
		if len(match) >= 3 {
			title := strings.TrimSpace(match[1])
			id := match[2]
			html := match[0]
			
			// Skip if we've already processed this chart ID
			if processedIDs[id] {
				log.Printf("DEBUG: Skipping duplicate chart with ID: %s and title: %s", id, title)
				continue
			}
			
			processedIDs[id] = true
			
			// Determine placeholder based on title and ID
			placeholder := h.determinePlaceholder(title, id)
			
			if placeholder != "" {
				extractedCharts = append(extractedCharts, Chart{
					ID:         id,
					Title:      title,
					HTML:       html,
					Placeholder: placeholder,
				})
				log.Printf("DEBUG: Extracted chart - ID: %s, Title: %s, Placeholder: %s", id, title, placeholder)
			}
		}
	}
	
	// If no charts were found with the regex, try a more lenient approach
	if len(extractedCharts) == 0 {
		extractedCharts = h.extractChartsLenient(charts)
	}
	
	return extractedCharts
}

// extractScripts extracts the ECharts script and initialization scripts from the charts HTML
func (h *HTMLBuilder) extractScripts(charts string) (string, string) {
	var initScripts string
	
	if charts == "" {
		return "", ""
	}
	
	// Always use the CDN version of ECharts
	echartsScript := `<script src="https://cdn.jsdelivr.net/npm/echarts@5.4.3/dist/echarts.min.js"></script>`
	log.Println("Using CDN version of ECharts")
	
	// Extract the initialization scripts - try multiple patterns
	initScriptsMatches := regexp.MustCompile(`<script[^>]*>\s*\(function\([^)]*\)\s*\{[\s\S]*?\}\)[\s\S]*?;\s*</script>`).FindAllString(charts, -1)
	if len(initScriptsMatches) == 0 {
		// Try alternative pattern
		initScriptsMatches = regexp.MustCompile(`<script[^>]*>[\s\S]*?function[\s\S]*?\{[\s\S]*?\}[\s\S]*?</script>`).FindAllString(charts, -1)
	}
	
	if len(initScriptsMatches) > 0 {
		for _, script := range initScriptsMatches {
			initScripts += script + "\n"
			log.Println("Found chart initialization script")
		}
	} else {
		// If no initialization scripts found, generate default ones for common chart IDs
		log.Println("No chart initialization scripts found, generating default ones")
		
		// Generate default initialization scripts for common chart IDs
		commonChartIDs := []string{"chart-solar-activity", "chart-band-conditions", "chart-propagation-timeline", "chart-k-index-trend", "chart-forecast"}
		
		for _, id := range commonChartIDs {
			if strings.Contains(charts, id) || strings.Contains(charts, "id=\""+id+"\"") {
				initScripts += fmt.Sprintf(`
<script>
  document.addEventListener('DOMContentLoaded', function() {
    var chartDom = document.getElementById('%s');
    if (chartDom) {
      var myChart = echarts.init(chartDom);
      var option = {
        title: { text: '' },
        tooltip: { trigger: 'axis' },
        legend: { data: ['Data'] },
        grid: { left: '3%%', right: '4%%', bottom: '3%%', containLabel: true },
        xAxis: { type: 'category', data: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'] },
        yAxis: { type: 'value' },
        series: [{ name: 'Data', type: 'line', data: [120, 132, 101, 134, 90, 230, 210] }]
      };
      myChart.setOption(option);
      window.addEventListener('resize', function() { myChart.resize(); });
    }
  });
</script>`, id)
				log.Printf("Generated default initialization script for chart ID: %s", id)
			}
		}
	}
	
	return echartsScript, initScripts
}

// extractChartsLenient extracts charts using a more lenient approach
func (h *HTMLBuilder) extractChartsLenient(charts string) []Chart {
	var extractedCharts []Chart
	processedIDs := make(map[string]bool)
	
	// Check for specific chart patterns
	chartPatterns := []struct {
		IDPattern   string
		TitlePattern string
		Placeholder string
	}{
		{"chart-solar-activity", "Solar Activity", "{{CHART_SOLAR_ACTIVITY}}"},
		{"chart-geomagnetic-conditions", "Geomagnetic Conditions", "{{CHART_GEOMAGNETIC_CONDITIONS}}"},
		{"chart-propagation-timeline", "Propagation Quality Timeline", "{{CHART_PROPAGATION_TIMELINE}}"},
		{"chart-band-conditions", "Band Conditions", "{{CHART_BAND_CONDITIONS}}"},
	}
	
	for _, pattern := range chartPatterns {
		// Try to find the chart div by ID
		idRegex := regexp.MustCompile(fmt.Sprintf(`<div[^>]*chart-container[^>]*>[\s\S]*?<div[^>]*id="%s"[^>]*>[\s\S]*?</div>[\s\S]*?</div>`, pattern.IDPattern))
		match := idRegex.FindString(charts)
		
		if match != "" && !processedIDs[pattern.IDPattern] {
			processedIDs[pattern.IDPattern] = true
			extractedCharts = append(extractedCharts, Chart{
				ID:         pattern.IDPattern,
				Title:      pattern.TitlePattern,
				HTML:       match,
				Placeholder: pattern.Placeholder,
			})
			log.Printf("DEBUG: Extracted chart (lenient) - ID: %s, Placeholder: %s", pattern.IDPattern, pattern.Placeholder)
		}
	}
	
	return extractedCharts
}

// determinePlaceholder determines the placeholder for a chart based on its title and ID
func (h *HTMLBuilder) determinePlaceholder(title, id string) string {
	// Map chart titles to placeholders
	titleToPlaceholder := map[string]string{
		"Solar Activity":              "{{SOLAR_ACTIVITY_CHART}}",
		"Geomagnetic Conditions":      "{{K_INDEX_CHART}}",
		"K-Index":                     "{{K_INDEX_CHART}}",
		"Propagation Quality Timeline": "{{PROPAGATION_TIMELINE_CHART}}",
		"Band Conditions":             "{{BAND_CONDITIONS_CHART}}",
		"Forecast":                    "{{FORECAST_CHART}}",
		"Propagation Forecast":        "{{FORECAST_CHART}}",
	}
	
	// Map chart IDs to placeholders as a fallback
	idToPlaceholder := map[string]string{
		"chart-solar-activity":         "{{SOLAR_ACTIVITY_CHART}}",
		"chart-geomagnetic-conditions": "{{K_INDEX_CHART}}",
		"chart-k-index":               "{{K_INDEX_CHART}}",
		"chart-propagation-timeline":   "{{PROPAGATION_TIMELINE_CHART}}",
		"chart-band-conditions":        "{{BAND_CONDITIONS_CHART}}",
		"chart-forecast":              "{{FORECAST_CHART}}",
		"chart-propagation-forecast":   "{{FORECAST_CHART}}",
	}
	
	// First try to match by title
	if placeholder, ok := titleToPlaceholder[title]; ok {
		log.Printf("DEBUG: Matched chart by title: %s -> %s", title, placeholder)
		return placeholder
	}
	
	// Then try to match by ID
	if placeholder, ok := idToPlaceholder[id]; ok {
		log.Printf("DEBUG: Matched chart by ID: %s -> %s", id, placeholder)
		return placeholder
	}
	
	// Try partial title matches
	for knownTitle, placeholder := range titleToPlaceholder {
		if strings.Contains(title, knownTitle) || strings.Contains(knownTitle, title) {
			log.Printf("DEBUG: Matched chart by partial title: %s ~ %s -> %s", title, knownTitle, placeholder)
			return placeholder
		}
	}
	
	// Try partial ID matches
	for knownID, placeholder := range idToPlaceholder {
		if strings.Contains(id, knownID) || strings.Contains(knownID, id) {
			log.Printf("DEBUG: Matched chart by partial ID: %s ~ %s -> %s", id, knownID, placeholder)
			return placeholder
		}
	}
	
	log.Printf("WARNING: Could not determine placeholder for chart - Title: %s, ID: %s", title, id)
	return ""
}

// mapChartsToPlaceholders maps charts to their placeholders, avoiding duplicates
func (h *HTMLBuilder) mapChartsToPlaceholders(charts []Chart) map[string]string {
	chartMap := make(map[string]string)
	processedPlaceholders := make(map[string]bool)
	
	// First pass: map charts to placeholders
	for _, chart := range charts {
		// Skip if we've already processed this placeholder
		if processedPlaceholders[chart.Placeholder] {
			log.Printf("DEBUG: Skipping duplicate placeholder: %s for chart ID: %s", chart.Placeholder, chart.ID)
			continue
		}
		
		processedPlaceholders[chart.Placeholder] = true
		chartMap[chart.Placeholder] = chart.HTML
		log.Printf("DEBUG: Mapped chart ID: %s to placeholder: %s", chart.ID, chart.Placeholder)
	}
	
	return chartMap
}

// replacePlaceholders replaces placeholders in the content with chart HTML
func (h *HTMLBuilder) replacePlaceholders(content string, chartMap map[string]string) string {
	result := content
	
	// Replace placeholders with chart HTML
	for placeholder, chartHTML := range chartMap {
		// Create a properly formatted chart section with appropriate div structure
		chartSection := fmt.Sprintf(`
<div class="chart-section">
	<div class="chart-container-integrated">
		%s
	</div>
</div>`, chartHTML)
		
		// Check if the placeholder exists in the content
		if strings.Contains(result, placeholder) {
			log.Printf("DEBUG: Replacing placeholder: %s with chart HTML", placeholder)
			result = strings.ReplaceAll(result, placeholder, chartSection)
		} else {
			// Check if the placeholder is wrapped in paragraph tags
			paragraphWrappedPlaceholder := fmt.Sprintf("<p>%s</p>", placeholder)
			if strings.Contains(result, paragraphWrappedPlaceholder) {
				log.Printf("DEBUG: Replacing paragraph-wrapped placeholder: %s with chart HTML", paragraphWrappedPlaceholder)
				result = strings.ReplaceAll(result, paragraphWrappedPlaceholder, chartSection)
			} else {
				log.Printf("WARNING: Placeholder not found in content: %s", placeholder)
			}
		}
	}
	
	// Remove any paragraph tags that might be wrapping our chart sections
	result = strings.ReplaceAll(result, "<p>\n<div class=\"chart-section\">", "\n<div class=\"chart-section\">")
	result = strings.ReplaceAll(result, "</div></p>", "</div>")
	
	return result
}

// addScriptsToHTML adds ECharts scripts and initialization scripts to the HTML
func (h *HTMLBuilder) addScriptsToHTML(content, echartsScript, initScripts string) string {
	result := content
	
	// Add ECharts script and initialization scripts before the closing body tag
	if echartsScript != "" || initScripts != "" {
		// Ensure we have the ECharts script
		if echartsScript == "" {
			echartsScript = `<script src="https://cdn.jsdelivr.net/npm/echarts@5.4.3/dist/echarts.min.js"></script>`
			log.Println("DEBUG: Added default ECharts CDN script")
		}
		
		// Combine scripts
		scripts := echartsScript + "\n" + initScripts
		
		// Check if body tag exists
		if strings.Contains(result, "</body>") {
			result = strings.Replace(result, "</body>", scripts+"\n</body>", 1)
		} else {
			// If no body tag, append to the end
			result = result + "\n" + scripts
		}
		
		log.Println("DEBUG: Added ECharts script and initialization scripts to HTML")
	}
	
	return result
}

