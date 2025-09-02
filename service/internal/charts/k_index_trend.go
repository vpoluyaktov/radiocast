package charts

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"

	"radiocast/internal/models"
)

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
			Ticks: cg.generateTimeTicks(xValues),
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

// generateTimeTicks creates appropriate time ticks for the X-axis
func (cg *ChartGenerator) generateTimeTicks(xValues []time.Time) []chart.Tick {
	var ticks []chart.Tick
	
	if len(xValues) == 0 {
		return ticks
	}
	
	// Create ticks for all data points
	for _, t := range xValues {
		ticks = append(ticks, chart.Tick{
			Value: chart.TimeToFloat64(t),
			Label: t.Format("15:04"),
		})
	}
	
	return ticks
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
