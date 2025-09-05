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

// integrateChartsInContent replaces chart placeholders with actual chart HTML
func (h *HTMLBuilder) integrateChartsInContent(content, charts string) string {
	// Parse chart HTML to extract individual chart elements
	chartMap := h.parseChartsHTML(charts)
	
	// If no charts parsed, try direct filename-based mapping
	if len(chartMap) == 0 && charts != "" {
		chartMap = h.createDirectChartMapping(charts)
	}
	
	// Extract the ECharts script and initialization scripts from the charts HTML
	var echartsScript, initScripts string
	if charts != "" {
		// Extract the ECharts script tag - try multiple patterns
		echartsScriptMatch := regexp.MustCompile(`<script[^>]*src="[^"]*echarts\.min\.js[^"]*"[^>]*></script>`).FindString(charts)
		if echartsScriptMatch == "" {
			// Try alternative pattern
			echartsScriptMatch = regexp.MustCompile(`<script[^>]*src="[^"]*files/echarts\.min\.js[^"]*"[^>]*></script>`).FindString(charts)
		}
		if echartsScriptMatch == "" {
			// Hardcode the script tag as fallback
			echartsScript = `<script src="echarts.min.js"></script>`
			log.Println("Using hardcoded ECharts script tag")
		} else {
			// Fix the path to use a relative path instead of /files/
			echartsScriptMatch = strings.Replace(echartsScriptMatch, `src="/files/echarts.min.js"`, `src="echarts.min.js"`, -1)
			echartsScript = echartsScriptMatch
			log.Println("Found and fixed ECharts script tag in HTML")
		}
		
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
			log.Println("No chart initialization scripts found")
		}
	}
	
	// Debug logging to see what's in the chartMap
	for placeholder, _ := range chartMap {
		log.Printf("Chart placeholder found: %s", placeholder)
	}
	
	// Add hardcoded placeholders if they're missing
	if _, ok := chartMap["{{SOLAR_ACTIVITY_CHART}}"]; !ok && strings.Contains(charts, "chart-solar-activity") {
		// Find the chart container for solar activity
		solarActivityMatch := regexp.MustCompile(`<div[^>]*chart-container[^>]*>[\s\S]*?<h3>Current Solar Activity</h3>[\s\S]*?</div>[\s]*</div>`).FindString(charts)
		if solarActivityMatch != "" {
			chartMap["{{SOLAR_ACTIVITY_CHART}}"] = solarActivityMatch
			log.Println("Added solar activity chart from regex match")
		}
	}
	
	if _, ok := chartMap["{{K_INDEX_CHART}}"]; !ok && (strings.Contains(charts, "chart-k-index") || strings.Contains(charts, "chart-forecast")) {
		// Find the chart container for K-index
		kIndexMatch := regexp.MustCompile(`<div[^>]*chart-container[^>]*>[\s\S]*?<h3>([^<]*K-[iI]ndex[^<]*)</h3>[\s\S]*?</div>[\s]*</div>`).FindString(charts)
		if kIndexMatch != "" {
			chartMap["{{K_INDEX_CHART}}"] = kIndexMatch
			log.Println("Added K-index chart from regex match")
		} else {
			// Try with forecast chart as fallback
			kIndexMatch = regexp.MustCompile(`<div[^>]*chart-container[^>]*>[\s\S]*?<h3>3-Day K-index Forecast</h3>[\s\S]*?</div>[\s]*</div>`).FindString(charts)
			if kIndexMatch != "" {
				chartMap["{{K_INDEX_CHART}}"] = kIndexMatch
				log.Println("Added K-index forecast chart from regex match")
			}
		}
	}
	
	if _, ok := chartMap["{{BAND_CONDITIONS_CHART}}"]; !ok && strings.Contains(charts, "chart-band-conditions") {
		// Find the chart container for band conditions
		bandConditionsMatch := regexp.MustCompile(`<div[^>]*chart-container[^>]*>[\s\S]*?<h3>HF Band Conditions[\s\S]*?</h3>[\s\S]*?</div>[\s]*</div>`).FindString(charts)
		if bandConditionsMatch != "" {
			chartMap["{{BAND_CONDITIONS_CHART}}"] = bandConditionsMatch
			log.Println("Added band conditions chart from regex match")
		}
	}
	
	if _, ok := chartMap["{{FORECAST_CHART}}"]; !ok && strings.Contains(charts, "chart-forecast") {
		// Find the chart container for forecast
		forecastMatch := regexp.MustCompile(`<div[^>]*chart-container[^>]*>[\s\S]*?<h3>3-Day K-index Forecast</h3>[\s\S]*?</div>[\s]*</div>`).FindString(charts)
		if forecastMatch != "" {
			chartMap["{{FORECAST_CHART}}"] = forecastMatch
			log.Println("Added forecast chart from regex match")
		}
	}
	
	// Replace placeholders with actual chart HTML
	integratedContent := content
	
	// Replace chart placeholders with professional chart sections
	for placeholder, chartHTML := range chartMap {
		// Create a properly escaped chart section
		chartSection := fmt.Sprintf(`
<div class="chart-section">
	<div class="chart-container-integrated">
		%s
	</div>
</div>`, chartHTML)
		
		// Make sure the placeholder is on its own line to avoid partial replacements
		// This helps prevent issues with markdown parsing and HTML tags
		if strings.Contains(integratedContent, "<p>"+placeholder+"</p>") {
			// If the placeholder is wrapped in <p> tags, replace the whole thing
			integratedContent = strings.Replace(integratedContent, "<p>"+placeholder+"</p>", chartSection, -1)
			log.Printf("Replaced placeholder in p tags: %s", placeholder)
		} else {
			// Otherwise do a direct replacement
			integratedContent = strings.Replace(integratedContent, placeholder, chartSection, -1)
			log.Printf("Replaced placeholder directly: %s", placeholder)
		}
	}
	
	// Clean up any potential literal </div> tags that might have been incorrectly parsed
	integratedContent = strings.Replace(integratedContent, "&lt;/div&gt;", "", -1)
	
	// Add the ECharts script and initialization scripts before the closing body tag
	if echartsScript != "" || initScripts != "" {
		scriptSection := "\n<!-- Chart Scripts -->\n" + echartsScript + "\n" + initScripts
		
		// Make sure we have a body tag to replace
		if strings.Contains(integratedContent, "</body>") {
			integratedContent = strings.Replace(integratedContent, "</body>", scriptSection+"\n</body>", 1)
			log.Println("Added chart scripts before closing body tag")
		} else {
			// If no body tag, add at the end
			integratedContent = integratedContent + "\n" + scriptSection
			log.Println("Added chart scripts at the end of HTML")
		}
		
		// Verify scripts were added
		if !strings.Contains(integratedContent, echartsScript) {
			log.Println("WARNING: ECharts script was not added to the HTML")
		}
		
		log.Println("Added chart scripts to HTML")
	} else {
		log.Println("No chart scripts to add")
	}
	
	return integratedContent
}

// createDirectChartMapping creates chart mapping based on filenames when parsing fails
func (h *HTMLBuilder) createDirectChartMapping(charts string) map[string]string {
	chartMap := make(map[string]string)
	
	// Look for image tags and map them directly
	if strings.Contains(charts, "solar_activity.png") {
		start := strings.Index(charts, "<div class=\"chart-container\">")
		if start != -1 {
			end := strings.Index(charts[start:], "</div>")
			if end != -1 {
				chartHTML := charts[start : start+end+6]
				if strings.Contains(chartHTML, "solar_activity.png") {
					chartMap["{{SOLAR_ACTIVITY_CHART}}"] = chartHTML
				}
			}
		}
	}
	
	if strings.Contains(charts, "k_index_trend.png") {
		// Find the chart container for K-index
		lines := strings.Split(charts, "\n")
		var currentChart strings.Builder
		inKIndexChart := false
		
		for _, line := range lines {
			if strings.Contains(line, "chart-container") && !inKIndexChart {
				currentChart.Reset()
				inKIndexChart = true
			}
			
			if inKIndexChart {
				currentChart.WriteString(line + "\n")
				if strings.Contains(line, "k_index_trend.png") {
					// Continue until we find the closing div
					continue
				}
				if strings.Contains(line, "</div>") && strings.Contains(currentChart.String(), "k_index_trend.png") {
					chartMap["{{K_INDEX_CHART}}"] = currentChart.String()
					inKIndexChart = false
					break
				}
			}
		}
	}
	
	if strings.Contains(charts, "band_conditions.png") {
		// Similar logic for band conditions
		lines := strings.Split(charts, "\n")
		var currentChart strings.Builder
		inBandChart := false
		
		for _, line := range lines {
			if strings.Contains(line, "chart-container") && !inBandChart {
				currentChart.Reset()
				inBandChart = true
			}
			
			if inBandChart {
				currentChart.WriteString(line + "\n")
				if strings.Contains(line, "</div>") && strings.Contains(currentChart.String(), "band_conditions.png") {
					chartMap["{{BAND_CONDITIONS_CHART}}"] = currentChart.String()
					inBandChart = false
					break
				}
			}
		}
	}
	
	if strings.Contains(charts, "forecast.png") {
		lines := strings.Split(charts, "\n")
		var currentChart strings.Builder
		inForecastChart := false
		
		for _, line := range lines {
			if strings.Contains(line, "chart-container") && !inForecastChart {
				currentChart.Reset()
				inForecastChart = true
			}
			
			if inForecastChart {
				currentChart.WriteString(line + "\n")
				if strings.Contains(line, "</div>") && strings.Contains(currentChart.String(), "forecast.png") {
					chartMap["{{FORECAST_CHART}}"] = currentChart.String()
					inForecastChart = false
					break
				}
			}
		}
	}
	
	if strings.Contains(charts, "propagation_timeline.png") {
		lines := strings.Split(charts, "\n")
		var currentChart strings.Builder
		inTimelineChart := false
		
		for _, line := range lines {
			if strings.Contains(line, "chart-container") && !inTimelineChart {
				currentChart.Reset()
				inTimelineChart = true
			}
			
			if inTimelineChart {
				currentChart.WriteString(line + "\n")
				if strings.Contains(line, "</div>") && strings.Contains(currentChart.String(), "propagation_timeline.png") {
					chartMap["{{PROPAGATION_TIMELINE_CHART}}"] = currentChart.String()
					inTimelineChart = false
					break
				}
			}
		}
	}
	
	return chartMap
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
