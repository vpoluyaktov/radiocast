package reports

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"radiocast/internal/models"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
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
	fullHTML := g.buildCompleteHTML(htmlContent, solarChart, kIndexChart, bandChart, data)
	
	log.Printf("Generated complete HTML report with %d characters", len(fullHTML))
	return fullHTML, nil
}

// markdownToHTML converts markdown to HTML
func (g *Generator) markdownToHTML(markdownText string) string {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse([]byte(markdownText))
	
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)
	
	return string(markdown.Render(doc, renderer))
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
func (g *Generator) buildCompleteHTML(content, solarChart, kIndexChart, bandChart string, data *models.PropagationData) string {
	timestamp := data.Timestamp.Format("2006-01-02 15:04:05 UTC")
	
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Radio Propagation Report - %s</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f8f9fa;
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            color: white;
            padding: 30px;
            border-radius: 10px;
            margin-bottom: 30px;
            text-align: center;
        }
        .header h1 {
            margin: 0;
            font-size: 2.5em;
        }
        .header .timestamp {
            opacity: 0.9;
            margin-top: 10px;
        }
        .summary-cards {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        .card {
            background: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            border-left: 4px solid #667eea;
        }
        .card h3 {
            margin-top: 0;
            color: #667eea;
        }
        .metric {
            font-size: 1.5em;
            font-weight: bold;
            color: #333;
        }
        .content {
            background: white;
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            margin-bottom: 30px;
        }
        .charts-section {
            background: white;
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            margin-bottom: 30px;
        }
        .chart-container {
            margin-bottom: 40px;
        }
        .footer {
            text-align: center;
            color: #666;
            font-size: 0.9em;
            margin-top: 30px;
        }
        h1, h2, h3 { color: #333; }
        h2 { border-bottom: 2px solid #667eea; padding-bottom: 5px; }
        code { background: #f4f4f4; padding: 2px 4px; border-radius: 3px; }
        pre { background: #f4f4f4; padding: 15px; border-radius: 5px; overflow-x: auto; }
        blockquote { border-left: 4px solid #667eea; margin: 0; padding-left: 20px; color: #666; }
        table { border-collapse: collapse; width: 100%%; margin: 20px 0; }
        th, td { border: 1px solid #ddd; padding: 12px; text-align: left; }
        th { background-color: #f8f9fa; font-weight: bold; }
        .status-good { color: #28a745; font-weight: bold; }
        .status-fair { color: #ffc107; font-weight: bold; }
        .status-poor { color: #dc3545; font-weight: bold; }
    </style>
</head>
<body>
    <div class="header">
        <h1>📡 Radio Propagation Report</h1>
        <div class="timestamp">Generated: %s</div>
    </div>
    
    <div class="summary-cards">
        <div class="card">
            <h3>Solar Activity</h3>
            <div class="metric">%.1f SFU</div>
            <div>Solar Flux Index</div>
            <div style="margin-top: 10px;">%s Activity</div>
        </div>
        <div class="card">
            <h3>Geomagnetic</h3>
            <div class="metric">K=%.1f</div>
            <div>Planetary K-index</div>
            <div style="margin-top: 10px;">%s Conditions</div>
        </div>
        <div class="card">
            <h3>Sunspots</h3>
            <div class="metric">%d</div>
            <div>Daily Count</div>
        </div>
        <div class="card">
            <h3>Best Bands</h3>
            <div style="margin-top: 10px;">%s</div>
        </div>
    </div>
    
    <div class="content">
        %s
    </div>
    
    <div class="charts-section">
        <h2>📊 Propagation Data Visualization</h2>
        
        <div class="chart-container">
            %s
        </div>
        
        <div class="chart-container">
            %s
        </div>
        
        <div class="chart-container">
            %s
        </div>
    </div>
    
    <div class="footer">
        <p>Report generated by Radio Propagation Service | Data sources: NOAA SWPC, N0NBH, SIDC</p>
        <p>For amateur radio operators worldwide | 73!</p>
    </div>
</body>
</html>`, 
		data.Timestamp.Format("2006-01-02"),
		timestamp,
		data.SolarData.SolarFluxIndex,
		data.SolarData.SolarActivity,
		data.GeomagData.KIndex,
		data.GeomagData.GeomagActivity,
		data.SolarData.SunspotNumber,
		formatBestBands(data.Forecast.Today.BestBands),
		content,
		solarChart,
		kIndexChart,
		bandChart,
	)
	
	return html
}

// formatBestBands formats the best bands list for display
func formatBestBands(bands []string) string {
	if len(bands) == 0 {
		return "Check forecast"
	}
	
	result := ""
	for i, band := range bands {
		if i > 0 {
			result += ", "
		}
		result += band
		if i >= 2 { // Limit to first 3 bands
			break
		}
	}
	
	return result
}
