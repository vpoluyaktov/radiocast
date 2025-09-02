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

	// Generate propagation quality timeline
	if timelineChart, err := cg.generatePropagationTimelineChart(data, sourceData); err == nil {
		chartFiles = append(chartFiles, timelineChart)
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

	// Create color-coded data points based on K-index zones
	var coloredSeries []chart.Series
	for i, kValue := range yValues {
		color := cg.getKIndexZoneColor(kValue)
		coloredSeries = append(coloredSeries, chart.TimeSeries{
			Name: fmt.Sprintf("K=%.1f", kValue),
			Style: chart.Style{
				StrokeColor: color,
				StrokeWidth: 3,
				DotColor:    color,
				DotWidth:    6,
			},
			XValues: []time.Time{xValues[i]},
			YValues: []float64{kValue},
		})
	}

	// Add connecting line for trend
	mainSeries := chart.TimeSeries{
		Name: "K-index Trend",
		Style: chart.Style{
			StrokeColor: drawing.Color{R: 51, G: 102, B: 204, A: 255}, // Blue
			StrokeWidth: 2,
			},
		XValues: xValues,
		YValues: yValues,
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
		Series: append([]chart.Series{mainSeries}, coloredSeries...),
	}

	// Add color-coded background zones for K-index levels
	if len(xValues) > 0 {
		minTime := xValues[0]
		maxTime := xValues[len(xValues)-1]
		
		// Quiet zone (0-2): Green background
		graph.Series = append(graph.Series, chart.TimeSeries{
			Name: "Quiet Zone (K≤2)",
			Style: chart.Style{
				StrokeColor:     drawing.Color{R: 0, G: 255, B: 0, A: 50}, // Transparent green
				StrokeWidth:     1,
				StrokeDashArray: []float64{5, 5},
			},
			XValues: []time.Time{minTime, maxTime},
			YValues: []float64{2, 2},
		})

		// Active zone (4+): Red line
		graph.Series = append(graph.Series, chart.TimeSeries{
			Name: "Active Zone (K≥4)",
			Style: chart.Style{
				StrokeColor:     drawing.Color{R: 255, G: 0, B: 0, A: 200}, // Red
				StrokeWidth:     2,
				StrokeDashArray: []float64{5, 5},
			},
			XValues: []time.Time{minTime, maxTime},
			YValues: []float64{4, 4},
		})

		// Unsettled zone (3): Yellow line
		graph.Series = append(graph.Series, chart.TimeSeries{
			Name: "Unsettled (K=3)",
			Style: chart.Style{
				StrokeColor:     drawing.Color{R: 255, G: 255, B: 0, A: 150}, // Yellow
				StrokeWidth:     1,
				StrokeDashArray: []float64{3, 3},
			},
			XValues: []time.Time{minTime, maxTime},
			YValues: []float64{3, 3},
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

// generateBandConditionsChart creates a time-based heatmap for band conditions
func (cg *ChartGenerator) generateBandConditionsChart(data *models.PropagationData) (string, error) {
	filename := filepath.Join(cg.outputDir, "band_conditions.png")

	// Create 24-hour timeline (every 2 hours for visibility)
	now := time.Now().UTC()
	startTime := now.Truncate(24 * time.Hour) // Start of today
	var timePoints []time.Time
	for i := 0; i < 12; i++ { // 12 points = every 2 hours
		timePoints = append(timePoints, startTime.Add(time.Duration(i*2)*time.Hour))
	}

	// Band data with current conditions (simulated progression for demo)
	bands := []string{"80m", "40m", "20m", "17m", "15m", "12m", "10m", "6m"}
	
	// Create heatmap-style visualization using multiple series
	graph := chart.Chart{
		Title: "24-Hour Band Conditions Heatmap",
		TitleStyle: chart.Style{
			FontSize: 16,
			FontColor: drawing.ColorBlack,
		},
		Background: chart.Style{
			Padding: chart.Box{
				Top:    40,
				Left:   60,
				Right:  20,
				Bottom: 80,
			},
		},
		Height: 400,
		Width:  800,
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
			Name: "HF Bands",
			NameStyle: chart.Style{
				FontSize: 12,
			},
			Style: chart.Style{
				FontSize: 10,
			},
			Range: &chart.ContinuousRange{
				Min: 0,
				Max: float64(len(bands)),
			},
			Ticks: func() []chart.Tick {
				var ticks []chart.Tick
				for i, band := range bands {
					ticks = append(ticks, chart.Tick{
						Value: float64(i) + 0.5,
						Label: band,
					})
				}
				return ticks
			}(),
		},
		Series: []chart.Series{},
	}

	// Add series for each band showing conditions over time
	for bandIdx, band := range bands {
		// Get current conditions for this band
		var dayCondition, nightCondition string
		switch band {
		case "80m":
			dayCondition, nightCondition = data.BandData.Band80m.Day, data.BandData.Band80m.Night
		case "40m":
			dayCondition, nightCondition = data.BandData.Band40m.Day, data.BandData.Band40m.Night
		case "20m":
			dayCondition, nightCondition = data.BandData.Band20m.Day, data.BandData.Band20m.Night
		case "17m":
			dayCondition, nightCondition = data.BandData.Band17m.Day, data.BandData.Band17m.Night
		case "15m":
			dayCondition, nightCondition = data.BandData.Band15m.Day, data.BandData.Band15m.Night
		case "12m":
			dayCondition, nightCondition = data.BandData.Band12m.Day, data.BandData.Band12m.Night
		case "10m":
			dayCondition, nightCondition = data.BandData.Band10m.Day, data.BandData.Band10m.Night
		case "6m":
			dayCondition, nightCondition = data.BandData.Band6m.Day, data.BandData.Band6m.Night
		}

		// Create time series for this band with day/night alternating
		var xValues []time.Time
		var yValues []float64
		var colors []drawing.Color
		
		for _, t := range timePoints {
			xValues = append(xValues, t)
			yValues = append(yValues, float64(bandIdx)+0.5)
			
			// Determine if this is day or night time (simple: 06:00-18:00 = day)
			hour := t.Hour()
			var condition string
			if hour >= 6 && hour < 18 {
				condition = dayCondition
			} else {
				condition = nightCondition
			}
			colors = append(colors, cg.getConditionColor(cg.conditionToValue(condition)))
		}

		// Add individual points with colors
		for pointIdx := range xValues {
			graph.Series = append(graph.Series, chart.TimeSeries{
				Name: fmt.Sprintf("%s-%d", band, pointIdx),
				Style: chart.Style{
					StrokeColor: colors[pointIdx],
					StrokeWidth: 8,
					DotColor:    colors[pointIdx],
					DotWidth:    12,
				},
				XValues: []time.Time{xValues[pointIdx]},
				YValues: []float64{yValues[pointIdx]},
			})
		}
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

// GenerateForecastChart creates a forecast chart (exported for testing)
func (cg *ChartGenerator) GenerateForecastChart(data *models.PropagationData) (string, error) {
	return cg.generateForecastChart(data)
}

// generateForecastChart creates a forecast chart
func (cg *ChartGenerator) generateForecastChart(data *models.PropagationData) (string, error) {
	filename := filepath.Join(cg.outputDir, "forecast.png")

	// Use actual forecast dates from the data
	dates := []time.Time{
		data.Forecast.Today.Date,
		data.Forecast.Tomorrow.Date,
		data.Forecast.DayAfter.Date,
	}

	// Extract K-index forecasts with better parsing and realistic values
	kIndexValues := []float64{
		cg.parseKIndexForecast(data.Forecast.Today.KIndexForecast),
		cg.parseKIndexForecast(data.Forecast.Tomorrow.KIndexForecast),
		cg.parseKIndexForecast(data.Forecast.DayAfter.KIndexForecast),
	}

	// Debug output for K-index values
	fmt.Printf("DEBUG Forecast K-index values: %v\n", kIndexValues)
	fmt.Printf("DEBUG Forecast dates: %v\n", dates)
	
	// Ensure we have valid values (not all zeros)
	for i, val := range kIndexValues {
		if val == 0 {
			kIndexValues[i] = 2.0 // Default fallback
		}
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
				Bottom: 60,
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
			Ticks: []chart.Tick{
				{Value: float64(dates[0].Unix()), Label: dates[0].Format("Sep 02")},
				{Value: float64(dates[1].Unix()), Label: dates[1].Format("Sep 03")},
				{Value: float64(dates[2].Unix()), Label: dates[2].Format("Sep 04")},
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
			Ticks: []chart.Tick{
				{Value: 0, Label: "0"},
				{Value: 1, Label: "1"},
				{Value: 2, Label: "2"},
				{Value: 3, Label: "3"},
				{Value: 4, Label: "4"},
				{Value: 5, Label: "5"},
			},
		},
		Series: []chart.Series{
			chart.TimeSeries{
				Name: "K-index Forecast",
				Style: chart.Style{
					StrokeColor: drawing.Color{R: 255, G: 165, B: 0, A: 255}, // Orange
					StrokeWidth: 3,
					DotColor:    drawing.Color{R: 255, G: 165, B: 0, A: 255},
					DotWidth:    6,
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

// getKIndexZoneColor returns color based on K-index value zones
func (cg *ChartGenerator) getKIndexZoneColor(kValue float64) drawing.Color {
	switch {
	case kValue >= 5:
		return drawing.Color{R: 128, G: 0, B: 128, A: 255} // Purple for storm
	case kValue >= 4:
		return drawing.Color{R: 255, G: 0, B: 0, A: 255}   // Red for active
	case kValue >= 3:
		return drawing.Color{R: 255, G: 255, B: 0, A: 255} // Yellow for unsettled
	default:
		return drawing.Color{R: 0, G: 255, B: 0, A: 255}   // Green for quiet
	}
}

// generatePropagationTimelineChart creates a dual Y-axis timeline showing solar flux + K-index + band openings
func (cg *ChartGenerator) generatePropagationTimelineChart(data *models.PropagationData, sourceData *models.SourceData) (string, error) {
	filename := filepath.Join(cg.outputDir, "propagation_timeline.png")

	// Use real K-index data if available
	var xValues []time.Time
	var kIndexValues []float64
	var solarFluxValues []float64
	
	if sourceData != nil && len(sourceData.NOAAKIndex) > 0 {
		for _, kData := range sourceData.NOAAKIndex {
			if parsedTime, err := time.Parse("2006-01-02T15:04:05", kData.TimeTag); err == nil {
				xValues = append(xValues, parsedTime)
				kIndexValues = append(kIndexValues, kData.EstimatedKp)
				// Use current solar flux for all time points (could be enhanced with historical solar data)
				solarFluxValues = append(solarFluxValues, data.SolarData.SolarFluxIndex)
			}
		}
	}

	// Fallback data if no real data available
	if len(xValues) == 0 {
		now := time.Now()
		for i := 0; i < 6; i++ {
			xValues = append(xValues, now.Add(time.Duration(-5+i)*time.Hour))
			kIndexValues = append(kIndexValues, 2.0+float64(i)*0.5)
			solarFluxValues = append(solarFluxValues, data.SolarData.SolarFluxIndex)
		}
	}

	graph := chart.Chart{
		Title: "Propagation Quality Timeline (24 Hours)",
		TitleStyle: chart.Style{
			FontSize: 16,
			FontColor: drawing.ColorBlack,
		},
		Background: chart.Style{
			Padding: chart.Box{
				Top:    40,
				Left:   70,
				Right:  70,
				Bottom: 60,
			},
		},
		Height: 400,
		Width:  800,
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
				Min: 0,
				Max: 6,
			},
		},
		YAxisSecondary: chart.YAxis{
			Name: "Solar Flux (SFU)",
			NameStyle: chart.Style{
				FontSize: 12,
			},
			Style: chart.Style{
				FontSize: 10,
			},
			Range: &chart.ContinuousRange{
				Min: 50,
				Max: 300,
			},
		},
		Series: []chart.Series{
			chart.TimeSeries{
				Name: "K-index",
				YAxis: chart.YAxisPrimary,
				Style: chart.Style{
					StrokeColor: drawing.Color{R: 255, G: 0, B: 0, A: 255}, // Red
					StrokeWidth: 3,
					DotColor:    drawing.Color{R: 255, G: 0, B: 0, A: 255},
					DotWidth:    5,
				},
				XValues: xValues,
				YValues: kIndexValues,
			},
			chart.TimeSeries{
				Name: "Solar Flux",
				YAxis: chart.YAxisSecondary,
				Style: chart.Style{
					StrokeColor: drawing.Color{R: 255, G: 165, B: 0, A: 255}, // Orange
					StrokeWidth: 2,
					DotColor:    drawing.Color{R: 255, G: 165, B: 0, A: 255},
					DotWidth:    4,
				},
				XValues: xValues,
				YValues: solarFluxValues,
			},
		},
	}

	f, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("failed to create propagation timeline chart file: %w", err)
	}
	defer f.Close()

	err = graph.Render(chart.PNG, f)
	if err != nil {
		return "", fmt.Errorf("failed to render propagation timeline chart: %w", err)
	}

	return filename, nil
}

// parseKIndexForecast extracts average K-index from forecast string
func (cg *ChartGenerator) parseKIndexForecast(forecast string) float64 {
	fmt.Printf("DEBUG: Parsing K-index forecast: '%s'\n", forecast)
	
	if forecast == "" {
		fmt.Printf("DEBUG: Empty forecast, returning 2.0\n")
		return 2.0
	}
	
	// Handle single values like "3.0" or "3.7"
	if len(forecast) == 3 && forecast[1] == '.' {
		if forecast[0] >= '0' && forecast[0] <= '9' && forecast[2] >= '0' && forecast[2] <= '9' {
			whole := float64(forecast[0] - '0')
			decimal := float64(forecast[2] - '0') / 10.0
			result := whole + decimal
			fmt.Printf("DEBUG: Single value '%s' parsed as %f\n", forecast, result)
			return result
		}
	}
	
	// Handle ranges like "3.2-4.2" or "2.7-4.7"
	if len(forecast) == 7 && forecast[1] == '.' && forecast[3] == '-' && forecast[5] == '.' {
		if forecast[0] >= '0' && forecast[0] <= '9' && forecast[2] >= '0' && forecast[2] <= '9' &&
		   forecast[4] >= '0' && forecast[4] <= '9' && forecast[6] >= '0' && forecast[6] <= '9' {
			min := float64(forecast[0] - '0') + float64(forecast[2] - '0')/10.0
			max := float64(forecast[4] - '0') + float64(forecast[6] - '0')/10.0
			result := (min + max) / 2
			fmt.Printf("DEBUG: Range '%s' parsed as %f (min=%f, max=%f)\n", forecast, result, min, max)
			return result
		}
	}
	
	// Handle simple ranges like "2-4"
	if len(forecast) == 3 && forecast[1] == '-' {
		if forecast[0] >= '0' && forecast[0] <= '9' && forecast[2] >= '0' && forecast[2] <= '9' {
			min := float64(forecast[0] - '0')
			max := float64(forecast[2] - '0')
			result := (min + max) / 2
			fmt.Printf("DEBUG: Simple range '%s' parsed as %f (min=%f, max=%f)\n", forecast, result, min, max)
			return result
		}
	}
	
	fmt.Printf("DEBUG: Could not parse '%s', returning default 2.0\n", forecast)
	return 2.0
}
