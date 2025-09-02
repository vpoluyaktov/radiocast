package charts

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"

	"radiocast/internal/models"
)

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
