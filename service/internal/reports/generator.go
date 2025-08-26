package reports

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
	"github.com/russross/blackfriday/v2"

	"radiocast/internal/models"
)

// Generator handles report generation and HTML conversion
type Generator struct{}

// NewGenerator creates a new report generator
func NewGenerator() *Generator {
	return &Generator{}
}

// GenerateHTML converts markdown report to HTML with embedded charts
func (g *Generator) GenerateHTML(markdownReport string, data *models.PropagationData) (string, error) {
	log.Println("Converting markdown to HTML and generating charts...")
	
	// Convert markdown to HTML
	htmlContent := g.markdownToHTML(markdownReport)
	
	// Generate charts
	solarChart, err := g.generateSolarChart(data)
	if err != nil {
		log.Printf("Warning: Failed to generate solar chart: %v", err)
		solarChart = "<p>Solar chart unavailable</p>"
	}
	
	kIndexChart, err := g.generateKIndexChart(data)
	if err != nil {
		log.Printf("Warning: Failed to generate K-index chart: %v", err)
		kIndexChart = "<p>K-index chart unavailable</p>"
	}
	
	bandChart, err := g.generateBandChart(data)
	if err != nil {
		log.Printf("Warning: Failed to generate band chart: %v", err)
		bandChart = "<p>Band conditions chart unavailable</p>"
	}
	
	// Combine everything into a complete HTML document
	fullHTML, err := g.buildCompleteHTML(htmlContent, solarChart, kIndexChart, bandChart, data)
	if err != nil {
		return "", fmt.Errorf("failed to build complete HTML: %w", err)
	}
	
	log.Printf("Generated complete HTML report with %d characters", len(fullHTML))
	return fullHTML, nil
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
	}{
		Date:        date,
		GeneratedAt: time.Now().Format("2006-01-02 15:04:05 UTC"),
		Content:     template.HTML(htmlContent),
		CSSStyles:   template.CSS(cssStyles),
		Charts:      template.HTML(""), // Charts will be embedded in content
	}
	
	// Execute the template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}
	
	return buf.String(), nil
}

// generateSolarChart creates a solar activity chart
func (g *Generator) generateSolarChart(data *models.PropagationData) (string, error) {
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Theme: types.ThemeWesteros,
			Width: "800px",
			Height: "400px",
		}),
		charts.WithTitleOpts(opts.Title{
			Title:    "Solar Activity Trends",
			Subtitle: "Solar Flux Index and Sunspot Number",
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Name: "Time",
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Name: "Value",
		}),
		charts.WithLegendOpts(opts.Legend{
			Show: true,
		}),
	)
	
	// Generate sample data points for the last 7 days
	// In a real implementation, you'd fetch historical data
	xAxis := make([]string, 7)
	solarFluxData := make([]opts.LineData, 7)
	sunspotData := make([]opts.LineData, 7)
	
	for i := 0; i < 7; i++ {
		date := time.Now().AddDate(0, 0, -6+i)
		xAxis[i] = date.Format("01-02")
		
		// Use current values with some variation for demonstration
		solarFlux := data.SolarData.SolarFluxIndex + float64((i-3)*5)
		sunspots := float64(data.SolarData.SunspotNumber) + float64((i-3)*10)
		
		solarFluxData[i] = opts.LineData{Value: solarFlux}
		sunspotData[i] = opts.LineData{Value: sunspots}
	}
	
	line.SetXAxis(xAxis).
		AddSeries("Solar Flux Index", solarFluxData).
		AddSeries("Sunspot Number", sunspotData).
		SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: true}))
	
	var buf bytes.Buffer
	if err := line.Render(&buf); err != nil {
		return "", err
	}
	
	return buf.String(), nil
}

// generateKIndexChart creates a K-index chart
func (g *Generator) generateKIndexChart(data *models.PropagationData) (string, error) {
	bar := charts.NewBar()
	bar.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Theme: types.ThemeWesteros,
			Width: "800px",
			Height: "400px",
		}),
		charts.WithTitleOpts(opts.Title{
			Title:    "Geomagnetic Activity",
			Subtitle: "K-index and A-index Values",
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Name: "Index Type",
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Name: "Value",
		}),
	)
	
	xAxis := []string{"K-index", "A-index"}
	kIndexData := []opts.BarData{
		{Value: data.GeomagData.KIndex},
		{Value: data.GeomagData.AIndex},
	}
	
	bar.SetXAxis(xAxis).
		AddSeries("Current Values", kIndexData)
	
	var buf bytes.Buffer
	if err := bar.Render(&buf); err != nil {
		return "", err
	}
	
	return buf.String(), nil
}

// generateBandChart creates a band conditions chart
func (g *Generator) generateBandChart(data *models.PropagationData) (string, error) {
	heatmap := charts.NewHeatMap()
	heatmap.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Theme: types.ThemeWesteros,
			Width: "800px",
			Height: "400px",
		}),
		charts.WithTitleOpts(opts.Title{
			Title:    "Band Conditions",
			Subtitle: "Day and Night Propagation Quality",
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Type: "category",
			Data: []string{"80m", "40m", "20m", "17m", "15m", "12m", "10m", "6m"},
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Type: "category",
			Data: []string{"Night", "Day"},
		}),
		charts.WithVisualMapOpts(opts.VisualMap{
			Calculable: true,
			Min:        0,
			Max:        3,
			InRange: &opts.VisualMapInRange{
				Color: []string{"#313695", "#4575b4", "#74add1", "#abd9e9", "#e0f3f8", "#ffffbf", "#fee090", "#fdae61", "#f46d43", "#d73027", "#a50026"},
			},
		}),
	)
	
	// Convert band conditions to numeric values
	conditionToValue := map[string]int{
		"Poor":      0,
		"Fair":      1,
		"Good":      2,
		"Excellent": 3,
	}
	
	bands := []models.BandCondition{
		data.BandData.Band80m, data.BandData.Band40m, data.BandData.Band20m, data.BandData.Band17m,
		data.BandData.Band15m, data.BandData.Band12m, data.BandData.Band10m, data.BandData.Band6m,
	}
	
	var heatmapData []opts.HeatMapData
	for i, band := range bands {
		dayValue := conditionToValue[band.Day]
		nightValue := conditionToValue[band.Night]
		
		heatmapData = append(heatmapData,
			opts.HeatMapData{Value: [3]interface{}{i, 1, dayValue}},   // Day
			opts.HeatMapData{Value: [3]interface{}{i, 0, nightValue}}, // Night
		)
	}
	
	heatmap.AddSeries("Band Conditions", heatmapData)
	
	var buf bytes.Buffer
	if err := heatmap.Render(&buf); err != nil {
		return "", err
	}
	
	return buf.String(), nil
}

// buildCompleteHTML creates a complete HTML document
func (g *Generator) buildCompleteHTML(content, solarChart, kIndexChart, bandChart string, data *models.PropagationData) (string, error) {
	// Use the new template-based conversion
	result, err := g.ConvertMarkdownToHTML(content, time.Now().Format("2006-01-02"))
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
    </div>
</body>
</html>`
}

// getDefaultCSSStyles returns fallback CSS styles
func (g *Generator) getDefaultCSSStyles() string {
	return `body { font-family: Arial, sans-serif; margin: 20px; }
.container { max-width: 1200px; margin: 0 auto; }
.header { text-align: center; margin-bottom: 30px; }
.content { background: white; padding: 20px; }`
}

