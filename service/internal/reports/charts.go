package reports

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"

	"radiocast/internal/models"
)

// ChartGenerator handles creation of static chart images
type ChartGenerator struct {
	outputDir string
}

// NewChartGenerator creates a new chart generator
func NewChartGenerator(outputDir string) *ChartGenerator {
	return &ChartGenerator{
		outputDir: outputDir,
	}
}

// GenerateCharts creates all chart images for the report
func (cg *ChartGenerator) GenerateCharts(data *models.PropagationData) ([]string, error) {
	return cg.GenerateChartsWithSources(data, nil)
}

// GenerateChartsWithSources creates all chart images for the report with access to source data
func (cg *ChartGenerator) GenerateChartsWithSources(data *models.PropagationData, sourceData *models.SourceData) ([]string, error) {
	var chartFiles []string

	// Generate solar activity chart
	if solarChart, err := cg.generateSolarActivityChart(data); err == nil {
		chartFiles = append(chartFiles, solarChart)
	}

	// Generate K-index trend chart with real data
	if kIndexChart, err := cg.generateKIndexChartWithSources(data, sourceData); err == nil {
		chartFiles = append(chartFiles, kIndexChart)
	}

	// Generate band conditions chart
	if bandChart, err := cg.generateBandConditionsChart(data); err == nil {
		chartFiles = append(chartFiles, bandChart)
	}

	// Generate forecast chart
	if forecastChart, err := cg.generateForecastChart(data); err == nil {
		chartFiles = append(chartFiles, forecastChart)
	}

	return chartFiles, nil
}

// generateSolarActivityChart creates a chart showing current solar conditions
func (cg *ChartGenerator) generateSolarActivityChart(data *models.PropagationData) (string, error) {
	filename := filepath.Join(cg.outputDir, "solar_activity.png")

	// Create bar chart for solar metrics
	graph := chart.BarChart{
		Title: "Current Solar Activity",
		TitleStyle: chart.Style{
			FontSize: 16,
			FontColor: drawing.ColorBlack,
		},
		Background: chart.Style{
			Padding: chart.Box{
				Top:    40,
				Left:   20,
				Right:  20,
				Bottom: 20,
			},
		},
		Height: 400,
		Width:  600,
		Bars: []chart.Value{
			{Value: data.SolarData.SolarFluxIndex, Label: "Solar Flux"},
			{Value: float64(data.SolarData.SunspotNumber), Label: "Sunspots"},
			{Value: data.GeomagData.KIndex * 50, Label: "K-index (x50)"}, // Scale for visibility
		},
		BarWidth: 80,
		XAxis: chart.Style{
			FontSize: 12,
		},
		YAxis: chart.YAxis{
			Name: "Values",
			NameStyle: chart.Style{
				FontSize: 12,
			},
			Style: chart.Style{
				FontSize: 10,
			},
		},
	}

	// Save chart to file
	f, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("failed to create solar activity chart file: %w", err)
	}
	defer f.Close()

	err = graph.Render(chart.PNG, f)
	if err != nil {
		return "", fmt.Errorf("failed to render solar activity chart: %w", err)
	}

	return filename, nil
}

// generateKIndexChart creates a time series chart for K-index (backward compatibility)
func (cg *ChartGenerator) generateKIndexChart(data *models.PropagationData) (string, error) {
	return cg.generateKIndexChartWithSources(data, nil)
}

// generateKIndexChartWithSources creates a time series chart for K-index using real historical data
func (cg *ChartGenerator) generateKIndexChartWithSources(data *models.PropagationData, sourceData *models.SourceData) (string, error) {
	filename := filepath.Join(cg.outputDir, "k_index_trend.png")

	// Use real K-index historical data from source data
	var xValues []time.Time
	var yValues []float64
	
	// Extract real K-index data from source data
	if sourceData != nil && len(sourceData.NOAAKIndex) > 0 {
		for _, kData := range sourceData.NOAAKIndex {
			if parsedTime, err := time.Parse("2006-01-02T15:04:05", kData.TimeTag); err == nil {
				xValues = append(xValues, parsedTime)
				yValues = append(yValues, kData.EstimatedKp)
			}
		}
	}
	
	// Fallback to sample data if no real data available
	if len(xValues) == 0 {
		now := time.Now()
		xValues = []time.Time{
			now.Add(-6 * time.Hour),
			now.Add(-3 * time.Hour),
			now,
		}
		yValues = []float64{2.0, 3.0, data.GeomagData.KIndex}
	}

	graph := chart.Chart{
		Title: "K-index Trend (24 Hours)",
		TitleStyle: chart.Style{
			FontSize: 16,
			FontColor: drawing.ColorBlack,
		},
		Background: chart.Style{
			Padding: chart.Box{
				Top:    40,
				Left:   70,
				Right:  20,
				Bottom: 60,
			},
		},
		Height: 350,
		Width:  700,
		XAxis: chart.XAxis{
			Name: "Time (UTC)",
			NameStyle: chart.Style{
				FontSize: 12,
			},
			Style: chart.Style{
				FontSize: 9,
			},
			ValueFormatter: func(v interface{}) string {
				if t, ok := v.(time.Time); ok {
					return t.Format("15:04")
				}
				return ""
			},
		},
		YAxis: chart.YAxis{
			Name: "K-index",
			NameStyle: chart.Style{
				FontSize: 12,
			},
			Style: chart.Style{
				FontSize: 10,
			},
			Range: &chart.ContinuousRange{
				Min: 0.0,
				Max: 6.0,
			},
			Ticks: []chart.Tick{
				{Value: 0, Label: "0"},
				{Value: 1, Label: "1"},
				{Value: 2, Label: "2 (Quiet)"},
				{Value: 3, Label: "3"},
				{Value: 4, Label: "4 (Active)"},
				{Value: 5, Label: "5 (Storm)"},
			},
		},
		Series: []chart.Series{
			chart.TimeSeries{
				Name: "K-index",
				Style: chart.Style{
					StrokeColor: drawing.Color{R: 51, G: 102, B: 204, A: 255}, // Blue
					StrokeWidth: 3,
					DotColor:    drawing.Color{R: 51, G: 102, B: 204, A: 255},
					DotWidth:    4,
				},
				XValues: xValues,
				YValues: yValues,
			},
		},
	}

	// Add reference lines for K-index levels (only if we have data points)
	if len(xValues) > 0 {
		minTime := xValues[0].Unix()
		maxTime := xValues[len(xValues)-1].Unix()
		
		graph.Series = append(graph.Series, chart.ContinuousSeries{
			Name: "Quiet (K≤2)",
			Style: chart.Style{
				StrokeColor:     drawing.ColorGreen,
				StrokeWidth:     1,
				StrokeDashArray: []float64{5, 5},
			},
			XValues: []float64{float64(minTime), float64(maxTime)},
			YValues: []float64{2, 2},
		})

		graph.Series = append(graph.Series, chart.ContinuousSeries{
			Name: "Active (K≥4)",
			Style: chart.Style{
				StrokeColor:     drawing.ColorRed,
				StrokeWidth:     1,
				StrokeDashArray: []float64{5, 5},
			},
			XValues: []float64{float64(minTime), float64(maxTime)},
			YValues: []float64{4, 4},
		})
	}

	f, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("failed to create K-index chart file: %w", err)
	}
	defer f.Close()

	err = graph.Render(chart.PNG, f)
	if err != nil {
		return "", fmt.Errorf("failed to render K-index chart: %w", err)
	}

	return filename, nil
}

// generateBandConditionsChart creates a heatmap-style chart for band conditions
func (cg *ChartGenerator) generateBandConditionsChart(data *models.PropagationData) (string, error) {
	filename := filepath.Join(cg.outputDir, "band_conditions.png")

	// Convert band conditions to numeric values for visualization
	bands := []string{"80m", "40m", "20m", "17m", "15m", "12m", "10m", "6m", "VHF+"}
	dayValues := []float64{
		cg.conditionToValue(data.BandData.Band80m.Day),
		cg.conditionToValue(data.BandData.Band40m.Day),
		cg.conditionToValue(data.BandData.Band20m.Day),
		cg.conditionToValue(data.BandData.Band17m.Day),
		cg.conditionToValue(data.BandData.Band15m.Day),
		cg.conditionToValue(data.BandData.Band12m.Day),
		cg.conditionToValue(data.BandData.Band10m.Day),
		cg.conditionToValue(data.BandData.Band6m.Day),
		cg.conditionToValue(data.BandData.VHFPlus.Day),
	}
	nightValues := []float64{
		cg.conditionToValue(data.BandData.Band80m.Night),
		cg.conditionToValue(data.BandData.Band40m.Night),
		cg.conditionToValue(data.BandData.Band20m.Night),
		cg.conditionToValue(data.BandData.Band17m.Night),
		cg.conditionToValue(data.BandData.Band15m.Night),
		cg.conditionToValue(data.BandData.Band12m.Night),
		cg.conditionToValue(data.BandData.Band10m.Night),
		cg.conditionToValue(data.BandData.Band6m.Night),
		cg.conditionToValue(data.BandData.VHFPlus.Night),
	}

	graph := chart.BarChart{
		Title: "HF Band Conditions",
		TitleStyle: chart.Style{
			FontSize: 16,
			FontColor: drawing.ColorBlack,
		},
		Background: chart.Style{
			Padding: chart.Box{
				Top:    40,
				Left:   60,
				Right:  20,
				Bottom: 60,
			},
		},
		Height:   400,
		Width:    700,
		BarWidth: 35,
		Bars:     []chart.Value{},
	}

	// Add day and night bars for each band
	for i, band := range bands {
		graph.Bars = append(graph.Bars, 
			chart.Value{
				Value: dayValues[i],
				Label: fmt.Sprintf("%s Day", band),
				Style: chart.Style{
					FillColor: cg.getConditionColor(dayValues[i]),
				},
			},
			chart.Value{
				Value: nightValues[i],
				Label: fmt.Sprintf("%s Night", band),
				Style: chart.Style{
					FillColor: cg.getConditionColor(nightValues[i]),
				},
			},
		)
	}

	graph.YAxis = chart.YAxis{
		Name: "Condition Quality",
		NameStyle: chart.Style{
			FontSize: 12,
		},
		Style: chart.Style{
			FontSize: 10,
		},
		Range: &chart.ContinuousRange{
			Min: 0,
			Max: 4,
		},
		Ticks: []chart.Tick{
			{Value: 0, Label: "Closed"},
			{Value: 1, Label: "Poor"},
			{Value: 2, Label: "Fair"},
			{Value: 3, Label: "Good"},
			{Value: 4, Label: "Excellent"},
		},
	}

	f, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("failed to create band conditions chart file: %w", err)
	}
	defer f.Close()

	err = graph.Render(chart.PNG, f)
	if err != nil {
		return "", fmt.Errorf("failed to render band conditions chart: %w", err)
	}

	return filename, nil
}

// generateForecastChart creates a forecast chart
func (cg *ChartGenerator) generateForecastChart(data *models.PropagationData) (string, error) {
	filename := filepath.Join(cg.outputDir, "forecast.png")

	// Create forecast data (using the forecast data from models)
	dates := []time.Time{
		data.Forecast.Today.Date,
		data.Forecast.Tomorrow.Date,
		data.Forecast.DayAfter.Date,
	}

	// Extract K-index forecasts (parse from string ranges like "2-3")
	kIndexValues := []float64{
		cg.parseKIndexForecast(data.Forecast.Today.KIndexForecast),
		cg.parseKIndexForecast(data.Forecast.Tomorrow.KIndexForecast),
		cg.parseKIndexForecast(data.Forecast.DayAfter.KIndexForecast),
	}

	graph := chart.Chart{
		Title: "3-Day K-index Forecast",
		TitleStyle: chart.Style{
			FontSize: 16,
			FontColor: drawing.ColorBlack,
		},
		Background: chart.Style{
			Padding: chart.Box{
				Top:    40,
				Left:   60,
				Right:  20,
				Bottom: 40,
			},
		},
		Height: 300,
		Width:  500,
		XAxis: chart.XAxis{
			Name: "Date",
			NameStyle: chart.Style{
				FontSize: 12,
			},
			Style: chart.Style{
				FontSize: 10,
			},
		},
		YAxis: chart.YAxis{
			Name: "K-index",
			NameStyle: chart.Style{
				FontSize: 12,
			},
			Style: chart.Style{
				FontSize: 10,
			},
			Range: &chart.ContinuousRange{
				Min: 0,
				Max: 6,
			},
		},
		Series: []chart.Series{
			chart.TimeSeries{
				Name: "K-index Forecast",
				Style: chart.Style{
					StrokeColor: drawing.Color{R: 255, G: 165, B: 0, A: 255}, // Orange
					StrokeWidth: 3,
					DotColor:    drawing.ColorRed,
					DotWidth:    5,
				},
				XValues: dates,
				YValues: kIndexValues,
			},
		},
	}

	f, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("failed to create forecast chart file: %w", err)
	}
	defer f.Close()

	err = graph.Render(chart.PNG, f)
	if err != nil {
		return "", fmt.Errorf("failed to render forecast chart: %w", err)
	}

	return filename, nil
}

// Helper functions

// conditionToValue converts condition strings to numeric values
func (cg *ChartGenerator) conditionToValue(condition string) float64 {
	switch condition {
	case "Closed":
		return 0
	case "Poor":
		return 1
	case "Fair":
		return 2
	case "Good":
		return 3
	case "Excellent":
		return 4
	default:
		return 1 // Default to Poor if unknown
	}
}

// getConditionColor returns color based on condition value
func (cg *ChartGenerator) getConditionColor(value float64) drawing.Color {
	switch {
	case value >= 4:
		return drawing.ColorGreen
	case value >= 3:
		return drawing.Color{R: 255, G: 255, B: 0, A: 255} // Yellow
	case value >= 2:
		return drawing.Color{R: 255, G: 165, B: 0, A: 255} // Orange
	case value >= 1:
		return drawing.ColorRed
	default:
		return drawing.Color{R: 128, G: 128, B: 128, A: 255} // Gray
	}
}

// parseKIndexForecast extracts average K-index from forecast string
func (cg *ChartGenerator) parseKIndexForecast(forecast string) float64 {
	// Simple parsing for ranges like "2-3", "1-2", etc.
	// In a real implementation, you'd want more robust parsing
	if len(forecast) >= 3 && forecast[1] == '-' {
		// Take the average of the range
		if forecast[0] >= '0' && forecast[0] <= '9' && forecast[2] >= '0' && forecast[2] <= '9' {
			min := float64(forecast[0] - '0')
			max := float64(forecast[2] - '0')
			return (min + max) / 2
		}
	}
	// Default fallback
	return 2.0
}
