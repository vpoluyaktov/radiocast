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

// generateForecastChart creates a forecast chart with visible bars
func (cg *ChartGenerator) generateForecastChart(data *models.PropagationData) (string, error) {
	filename := filepath.Join(cg.outputDir, "forecast.png")

	// Extract K-index forecasts
	kIndexValues := []float64{
		cg.parseKIndexForecast(data.Forecast.Today.KIndexForecast),
		cg.parseKIndexForecast(data.Forecast.Tomorrow.KIndexForecast),
		cg.parseKIndexForecast(data.Forecast.DayAfter.KIndexForecast),
	}

	// Debug output
	fmt.Printf("DEBUG Forecast K-index values: %v\n", kIndexValues)
	
	// Ensure we have valid values
	for i, val := range kIndexValues {
		if val == 0 {
			kIndexValues[i] = 2.0
		}
	}

	// Create date labels
	now := time.Now().UTC()
	today := now.Format("Jan 02")
	tomorrow := now.Add(24 * time.Hour).Format("Jan 02")
	dayAfter := now.Add(48 * time.Hour).Format("Jan 02")

	// Use BarChart with explicit styling - this should work
	graph := chart.BarChart{
		Title: "3-Day K-index Forecast",
		TitleStyle: chart.Style{
			FontSize: 18,
			FontColor: drawing.ColorBlack,
		},
		Background: chart.Style{
			Padding: chart.Box{
				Top:    50,
				Left:   80,
				Right:  50,
				Bottom: 80,
			},
			FillColor: drawing.Color{R: 248, G: 249, B: 250, A: 255},
		},
		Height: 400,
		Width:  600,
		BarWidth: 120,
		XAxis: chart.Style{
			FontSize: 12,
			FontColor: drawing.Color{R: 52, G: 58, B: 64, A: 255},
		},
		YAxis: chart.YAxis{
			Name: "K-index Level",
			NameStyle: chart.Style{
				FontSize: 14,
				FontColor: drawing.Color{R: 52, G: 58, B: 64, A: 255},
			},
			Style: chart.Style{
				FontSize: 11,
				FontColor: drawing.Color{R: 108, G: 117, B: 125, A: 255},
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
		Bars: []chart.Value{
			{
				Value: kIndexValues[0],
				Label: fmt.Sprintf("Today\n%s\nK=%.1f", today, kIndexValues[0]),
				Style: chart.Style{
					FillColor:   cg.getKIndexColor(kIndexValues[0]),
					StrokeColor: drawing.Color{R: 52, G: 58, B: 64, A: 255},
					StrokeWidth: 2,
				},
			},
			{
				Value: kIndexValues[1],
				Label: fmt.Sprintf("Tomorrow\n%s\nK=%.1f", tomorrow, kIndexValues[1]),
				Style: chart.Style{
					FillColor:   cg.getKIndexColor(kIndexValues[1]),
					StrokeColor: drawing.Color{R: 52, G: 58, B: 64, A: 255},
					StrokeWidth: 2,
				},
			},
			{
				Value: kIndexValues[2],
				Label: fmt.Sprintf("Day After\n%s\nK=%.1f", dayAfter, kIndexValues[2]),
				Style: chart.Style{
					FillColor:   cg.getKIndexColor(kIndexValues[2]),
					StrokeColor: drawing.Color{R: 52, G: 58, B: 64, A: 255},
					StrokeWidth: 2,
				},
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

// parseKIndexForecast extracts average K-index from forecast string
func (cg *ChartGenerator) parseKIndexForecast(forecast string) float64 {
	fmt.Printf("DEBUG: Parsing K-index forecast: '%s'\n", forecast)
	
	if forecast == "" {
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

// getKIndexColor returns appropriate colors for K-index forecast values
func (cg *ChartGenerator) getKIndexColor(kValue float64) drawing.Color {
	switch {
	case kValue >= 5:
		return drawing.Color{R: 128, G: 0, B: 128, A: 255} // Purple for storm
	case kValue >= 4:
		return drawing.Color{R: 220, G: 53, B: 69, A: 255}  // Red for active
	case kValue >= 3:
		return drawing.Color{R: 253, G: 126, B: 20, A: 255} // Orange for unsettled
	case kValue >= 2:
		return drawing.Color{R: 255, G: 193, B: 7, A: 255}  // Yellow for fair
	default:
		return drawing.Color{R: 40, G: 167, B: 69, A: 255}  // Green for quiet
	}
}

