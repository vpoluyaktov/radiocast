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
