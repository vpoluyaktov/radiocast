package reports

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/russross/blackfriday/v2"

	"radiocast/internal/config"
	"radiocast/internal/models"
)

// Generator handles report generation and HTML conversion
type Generator struct{
	outputDir string
}

// NewGenerator creates a new report generator
func NewGenerator(outputDir string) *Generator {
	return &Generator{
		outputDir: outputDir,
	}
}


// GenerateHTML converts markdown report to HTML with embedded charts
func (g *Generator) GenerateHTML(markdownReport string, data *models.PropagationData) (string, error) {
	log.Println("Converting markdown to HTML and generating charts...")
	
	// Generate chart images using the new chart generator
	chartGen := NewChartGenerator(g.outputDir)
	chartFiles, err := chartGen.GenerateCharts(data)
	if err != nil {
		log.Printf("Warning: Failed to generate charts: %v", err)
		chartFiles = []string{}
	}
	
	// Convert markdown to HTML
	htmlContent := g.markdownToHTML(markdownReport)
	
	// Build chart HTML references
	chartsHTML := g.buildChartsHTML(chartFiles)
	
	// Combine everything into a complete HTML document
	fullHTML, err := g.buildCompleteHTML(htmlContent, chartsHTML, data)
	if err != nil {
		return "", fmt.Errorf("failed to build complete HTML: %w", err)
	}
	
	log.Printf("Generated complete HTML report with %d characters and %d charts", len(fullHTML), len(chartFiles))
	return fullHTML, nil
}

// buildChartsHTML creates HTML references to chart image files
func (g *Generator) buildChartsHTML(chartFiles []string) string {
	if len(chartFiles) == 0 {
		return ""
	}
	
	var html strings.Builder
	html.WriteString("<div class='charts-section'>\n")
	html.WriteString("<h2>📊 Propagation Charts</h2>\n")
	
	for _, chartFile := range chartFiles {
		// Create a descriptive title based on filename
		title := g.getChartTitle(chartFile)
		html.WriteString(fmt.Sprintf("<div class='chart-container'>\n"))
		html.WriteString(fmt.Sprintf("<h3>%s</h3>\n", title))
		html.WriteString(fmt.Sprintf("<img src='%s' alt='%s' class='chart-image'/>\n", chartFile, title))
		html.WriteString("</div>\n")
	}
	
	html.WriteString("</div>\n")
	return html.String()
}

// getChartTitle returns a human-readable title for a chart file
func (g *Generator) getChartTitle(filename string) string {
	switch filename {
	case "solar_activity.png":
		return "Current Solar Activity"
	case "k_index_trend.png":
		return "K-index Trend (24 Hours)"
	case "band_conditions.png":
		return "HF Band Conditions"
	case "forecast.png":
		return "3-Day K-index Forecast"
	default:
		return strings.TrimSuffix(filename, ".png")
	}
}


// markdownToHTML converts markdown to HTML using blackfriday
func (g *Generator) markdownToHTML(markdownText string) string {
	htmlBytes := blackfriday.Run([]byte(markdownText))
	return string(htmlBytes)
}


// ConvertMarkdownToHTML converts markdown content to a complete HTML document using configurable templates
func (g *Generator) ConvertMarkdownToHTML(markdownContent string, date string) (string, error) {
	// Convert markdown to HTML using blackfriday
	htmlBytes := blackfriday.Run([]byte(markdownContent))
	htmlContent := string(htmlBytes)
	
	// Load HTML template
	htmlTemplate, err := g.loadHTMLTemplate()
	if err != nil {
		return "", fmt.Errorf("failed to load HTML template: %w", err)
	}
	
	// Load CSS styles
	cssStyles, err := g.loadCSSStyles()
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
func (g *Generator) ConvertMarkdownToHTMLWithCharts(markdownContent string, charts string, date string) (string, error) {
	// Convert markdown to HTML using blackfriday
	htmlBytes := blackfriday.Run([]byte(markdownContent))
	htmlContent := string(htmlBytes)
	
	// Load HTML template
	htmlTemplate, err := g.loadHTMLTemplate()
	if err != nil {
		return "", fmt.Errorf("failed to load HTML template: %w", err)
	}
	
	// Load CSS styles
	cssStyles, err := g.loadCSSStyles()
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

// buildCompleteHTML creates a complete HTML document
func (g *Generator) buildCompleteHTML(content, charts string, data *models.PropagationData) (string, error) {
	// Use the new template-based conversion with charts
	result, err := g.ConvertMarkdownToHTMLWithCharts(content, charts, time.Now().Format("2006-01-02"))
	if err != nil {
		return "", err
	}
	return result, nil
}

// loadHTMLTemplate loads the HTML template from file
func (g *Generator) loadHTMLTemplate() (string, error) {
	templatePath := filepath.Join("internal", "templates", "report_template.html")
	content, err := os.ReadFile(templatePath)
	if err != nil {
		// Return default template if file not found
		return g.getDefaultHTMLTemplate(), nil
	}
	return string(content), nil
}

// loadCSSStyles loads the CSS styles from file
func (g *Generator) loadCSSStyles() (string, error) {
	cssPath := filepath.Join("internal", "templates", "report_styles.css")
	content, err := os.ReadFile(cssPath)
	if err != nil {
		// Return default styles if file not found
		return g.getDefaultCSSStyles(), nil
	}
	return string(content), nil
}

// getDefaultHTMLTemplate returns a fallback HTML template
func (g *Generator) getDefaultHTMLTemplate() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Radio Propagation Report - {{.Date}}</title>
    <style>{{.CSSStyles}}</style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Radio Propagation Report</h1>
            <h2>{{.Date}}</h2>
        </div>
        <div class="content">
            {{.Content}}
        </div>
        <div class="footer">
            <hr>
            <p class="version-info">Generated on {{.GeneratedAt}} | Radio Propagation Service v{{.Version}}</p>
        </div>
    </div>
</body>
</html>`
}

// getDefaultCSSStyles returns fallback CSS styles
func (g *Generator) getDefaultCSSStyles() string {
	return `body { font-family: Arial, sans-serif; margin: 20px; }
.container { max-width: 1200px; margin: 0 auto; }
.header { text-align: center; margin-bottom: 30px; }
.content { background: white; padding: 20px; }
.footer { margin-top: 30px; text-align: center; }
.version-info { color: #666; font-size: 0.9em; margin: 10px 0; }`
}

