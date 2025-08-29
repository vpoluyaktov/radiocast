package reports

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
	"github.com/russross/blackfriday/v2"

	"radiocast/internal/config"
	"radiocast/internal/models"
)

// Generator handles report generation and HTML conversion
type Generator struct{}

// NewGenerator creates a new report generator
func NewGenerator() *Generator {
	return &Generator{}
}

// ChartData represents the JSON structure from LLM for chart generation
type ChartData struct {
	BandConditions []BandCondition `json:"bandConditions"`
	SolarActivity  SolarActivity   `json:"solarActivity"`
	Forecast       []ForecastDay   `json:"forecast"`
}

type BandCondition struct {
	Band       string `json:"band"`
	Day        int    `json:"day"`
	Night      int    `json:"night"`
	DayLabel   string `json:"dayLabel"`
	NightLabel string `json:"nightLabel"`
}

type SolarActivity struct {
	SolarFlux      float64 `json:"solarFlux"`
	SunspotNumber  int     `json:"sunspotNumber"`
	KIndex         float64 `json:"kIndex"`
	AIndex         float64 `json:"aIndex"`
	Trend          string  `json:"trend"`
}

type ForecastDay struct {
	Day       string  `json:"day"`
	KIndex    float64 `json:"kIndex"`
	Condition string  `json:"condition"`
}

// GenerateHTML converts markdown report to HTML with embedded charts
func (g *Generator) GenerateHTML(markdownReport string, data *models.PropagationData) (string, error) {
	log.Println("Converting markdown to HTML and generating charts...")
	
	// Extract chart data from markdown
	chartData, err := g.extractChartData(markdownReport)
	if err != nil {
		log.Printf("Warning: Failed to extract chart data: %v", err)
		chartData = nil
	}
	
	// Convert markdown to HTML (remove chart data section)
	htmlContent := g.markdownToHTML(g.removeChartDataSection(markdownReport))
	
	// Generate charts using LLM data if available, fallback to original data
	var charts string
	if chartData != nil {
		charts = g.generateChartsFromLLMData(chartData)
	} else {
		charts = g.generateFallbackCharts(data)
	}
	
	// Combine everything into a complete HTML document
	fullHTML, err := g.buildCompleteHTML(htmlContent, charts, data)
	if err != nil {
		return "", fmt.Errorf("failed to build complete HTML: %w", err)
	}
	
	log.Printf("Generated complete HTML report with %d characters", len(fullHTML))
	return fullHTML, nil
}

// extractChartData extracts JSON chart data from markdown report
func (g *Generator) extractChartData(markdownReport string) (*ChartData, error) {
	// Find the Chart Data section with regex
	re := regexp.MustCompile(`## Chart Data\s*\x60\x60\x60json\s*([\s\S]*?)\s*\x60\x60\x60`)
	matches := re.FindStringSubmatch(markdownReport)
	
	if len(matches) < 2 {
		return nil, fmt.Errorf("no chart data section found")
	}
	
	var chartData ChartData
	if err := json.Unmarshal([]byte(matches[1]), &chartData); err != nil {
		return nil, fmt.Errorf("failed to parse chart data JSON: %w", err)
	}
	
	return &chartData, nil
}

// removeChartDataSection removes the Chart Data section from markdown
func (g *Generator) removeChartDataSection(markdownReport string) string {
	re := regexp.MustCompile(`## Chart Data\s*\x60\x60\x60json[\s\S]*?\x60\x60\x60`)
	return re.ReplaceAllString(markdownReport, "")
}

// generateChartsFromLLMData creates charts using LLM-provided data
func (g *Generator) generateChartsFromLLMData(chartData *ChartData) string {
	var charts []string
	
	// Generate band conditions heatmap
	if bandChart := g.generateBandHeatmap(chartData.BandConditions); bandChart != "" {
		charts = append(charts, bandChart)
	}
	
	// Generate solar activity gauge
	if solarChart := g.generateSolarGauge(chartData.SolarActivity); solarChart != "" {
		charts = append(charts, solarChart)
	}
	
	// Generate forecast line chart
	if forecastChart := g.generateForecastChart(chartData.Forecast); forecastChart != "" {
		charts = append(charts, forecastChart)
	}
	
	return strings.Join(charts, "\n")
}

// generateFallbackCharts creates basic charts from original data
func (g *Generator) generateFallbackCharts(data *models.PropagationData) string {
	var charts []string
	
	// Simple solar activity chart
	if chart := g.generateSimpleSolarChart(data); chart != "" {
		charts = append(charts, chart)
	}
	
	return strings.Join(charts, "\n")
}

// generateBandHeatmap creates a heatmap chart for band conditions
func (g *Generator) generateBandHeatmap(bandConditions []BandCondition) string {
	heatmap := charts.NewHeatMap()
	heatmap.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Theme: types.ThemeWesteros,
			Width: "800px",
			Height: "400px",
		}),
		charts.WithTitleOpts(opts.Title{
			Title:    "📻 Band Conditions",
			Subtitle: "Day and Night Propagation Quality",
		}),
	)
	
	// Prepare data for heatmap
	var heatmapData []opts.HeatMapData
	for i, band := range bandConditions {
		heatmapData = append(heatmapData, 
			opts.HeatMapData{Value: [3]interface{}{i, 0, band.Day}},    // Day
			opts.HeatMapData{Value: [3]interface{}{i, 1, band.Night}}, // Night
		)
	}
	
	heatmap.AddSeries("Band Conditions", heatmapData)
	
	var buf bytes.Buffer
	if err := heatmap.Render(&buf); err != nil {
		return ""
	}
	return fmt.Sprintf("<div class='chart-container'>%s</div>", buf.String())
}

// generateSolarGauge creates a gauge chart for solar activity
func (g *Generator) generateSolarGauge(solarActivity SolarActivity) string {
	gauge := charts.NewGauge()
	gauge.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Theme: types.ThemeWesteros,
			Width: "400px",
			Height: "400px",
		}),
		charts.WithTitleOpts(opts.Title{
			Title: "☀️ Solar Activity",
		}),
	)
	
	gauge.AddSeries("Solar Flux", []opts.GaugeData{
		{Name: "Solar Flux", Value: solarActivity.SolarFlux},
	})
	
	var buf bytes.Buffer
	if err := gauge.Render(&buf); err != nil {
		return ""
	}
	return fmt.Sprintf("<div class='chart-container'>%s</div>", buf.String())
}

// generateForecastChart creates a line chart for K-index forecast
func (g *Generator) generateForecastChart(forecast []ForecastDay) string {
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Theme: types.ThemeWesteros,
			Width: "600px",
			Height: "300px",
		}),
		charts.WithTitleOpts(opts.Title{
			Title: "📈 3-Day K-index Forecast",
		}),
	)
	
	var xAxis []string
	var yData []opts.LineData
	for _, day := range forecast {
		xAxis = append(xAxis, day.Day)
		yData = append(yData, opts.LineData{Value: day.KIndex})
	}
	
	line.SetXAxis(xAxis).AddSeries("K-index", yData)
	
	var buf bytes.Buffer
	if err := line.Render(&buf); err != nil {
		return ""
	}
	return fmt.Sprintf("<div class='chart-container'>%s</div>", buf.String())
}

// generateSimpleSolarChart creates a simple solar activity chart
func (g *Generator) generateSimpleSolarChart(data *models.PropagationData) string {
	bar := charts.NewBar()
	bar.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Theme: types.ThemeWesteros,
			Width: "600px",
			Height: "300px",
		}),
		charts.WithTitleOpts(opts.Title{
			Title: "Current Solar Conditions",
		}),
	)
	
	xAxis := []string{"Solar Flux", "Sunspot Number", "K-index"}
	barData := []opts.BarData{
		{Value: data.SolarData.SolarFluxIndex},
		{Value: data.SolarData.SunspotNumber},
		{Value: data.GeomagData.KIndex * 100}, // Scale K-index for visibility
	}
	
	bar.SetXAxis(xAxis).AddSeries("Values", barData)
	
	var buf bytes.Buffer
	if err := bar.Render(&buf); err != nil {
		return ""
	}
	return fmt.Sprintf("<div class='chart-container'>%s</div>", buf.String())
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

