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

// generateBandConditionsChart creates a time-based heatmap for band conditions
func (cg *ChartGenerator) generateBandConditionsChart(data *models.PropagationData) (string, error) {
	filename := filepath.Join(cg.outputDir, "band_conditions.png")

	// Note: This chart uses current day/night conditions rather than time-series data

	// Band data ordered from highest to lowest frequency
	bands := []string{"6m", "10m", "12m", "15m", "17m", "20m", "40m", "80m"}
	
	// Create improved heatmap using BarChart for better visual control
	graph := chart.BarChart{
		Title: "HF Band Conditions (24-Hour Overview)",
		TitleStyle: chart.Style{
			FontSize: 18,
			FontColor: drawing.ColorBlack,
		},
		Background: chart.Style{
			Padding: chart.Box{
				Top:    50,
				Left:   80,
				Right:  100,
				Bottom: 60,
			},
			FillColor: drawing.Color{R: 248, G: 249, B: 250, A: 255}, // Light gray background
		},
		Height: 450,
		Width:  900,
		BarWidth: 80,
		XAxis: chart.Style{
			FontSize: 11,
			FontColor: drawing.Color{R: 52, G: 58, B: 64, A: 255},
		},
		YAxis: chart.YAxis{
			Name: "Propagation Quality Score",
			NameStyle: chart.Style{
				FontSize: 13,
				FontColor: drawing.Color{R: 52, G: 58, B: 64, A: 255},
			},
			Style: chart.Style{
				FontSize: 10,
				FontColor: drawing.Color{R: 108, G: 117, B: 125, A: 255},
			},
			Range: &chart.ContinuousRange{
				Min: 0,
				Max: 5,
			},
			Ticks: []chart.Tick{
				{Value: 0, Label: "Closed"},
				{Value: 1, Label: "Poor"},
				{Value: 2, Label: "Fair"},
				{Value: 3, Label: "Good"},
				{Value: 4, Label: "Excellent"},
			},
		},
		Bars: []chart.Value{},
	}

	// Calculate average conditions for each band across day/night
	for _, band := range bands {
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

		// Calculate weighted average (day conditions weighted more heavily during day hours)
		currentHour := time.Now().UTC().Hour()
		var avgCondition float64
		if currentHour >= 6 && currentHour < 18 {
			// Daytime: 70% day conditions, 30% night conditions
			avgCondition = (cg.conditionToValue(dayCondition)*0.7 + cg.conditionToValue(nightCondition)*0.3)
		} else {
			// Nighttime: 30% day conditions, 70% night conditions
			avgCondition = (cg.conditionToValue(dayCondition)*0.3 + cg.conditionToValue(nightCondition)*0.7)
		}

		// Create bar with appropriate color and styling
		barColor := cg.getEnhancedConditionColor(avgCondition)
		graph.Bars = append(graph.Bars, chart.Value{
			Value: avgCondition,
			Label: fmt.Sprintf("%s\n(D:%s/N:%s)", band, 
				cg.abbreviateCondition(dayCondition), 
				cg.abbreviateCondition(nightCondition)),
			Style: chart.Style{
				FillColor:   barColor,
				StrokeColor: cg.darkenColor(barColor, 0.2),
				StrokeWidth: 2,
			},
		})
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

// getEnhancedConditionColor returns enhanced colors with better gradients for band conditions
func (cg *ChartGenerator) getEnhancedConditionColor(value float64) drawing.Color {
	switch {
	case value >= 4:
		return drawing.Color{R: 40, G: 167, B: 69, A: 255}    // Bootstrap success green
	case value >= 3:
		return drawing.Color{R: 255, G: 193, B: 7, A: 255}    // Bootstrap warning yellow
	case value >= 2:
		return drawing.Color{R: 253, G: 126, B: 20, A: 255}   // Bootstrap orange
	case value >= 1:
		return drawing.Color{R: 220, G: 53, B: 69, A: 255}    // Bootstrap danger red
	default:
		return drawing.Color{R: 108, G: 117, B: 125, A: 255}  // Bootstrap secondary gray
	}
}

// abbreviateCondition returns shortened condition labels for compact display
func (cg *ChartGenerator) abbreviateCondition(condition string) string {
	switch condition {
	case "Excellent":
		return "Exc"
	case "Good":
		return "Good"
	case "Fair":
		return "Fair"
	case "Poor":
		return "Poor"
	case "Closed":
		return "Cls"
	default:
		return "N/A"
	}
}

// darkenColor darkens a color by the specified factor (0.0 = no change, 1.0 = black)
func (cg *ChartGenerator) darkenColor(color drawing.Color, factor float64) drawing.Color {
	if factor < 0 {
		factor = 0
	}
	if factor > 1 {
		factor = 1
	}
	
	return drawing.Color{
		R: uint8(float64(color.R) * (1.0 - factor)),
		G: uint8(float64(color.G) * (1.0 - factor)),
		B: uint8(float64(color.B) * (1.0 - factor)),
		A: color.A,
	}
}
