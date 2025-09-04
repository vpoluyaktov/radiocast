package charts

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"

	"radiocast/internal/models"
)

// generateKIndexChartWithSources creates a time series chart for K-index using real historical data
func (cg *ChartGenerator) generateKIndexChartWithSources(data *models.PropagationData, sourceData *models.SourceData) (string, error) {
	filename := filepath.Join(cg.outputDir, "k_index_trend.png")
	// Extract and parse K-index data
	xValues, yValues := extractKIndexPoints(sourceData)
	// If still nothing, create a minimal 72h window so chart renders
	if len(xValues) == 0 {
		now := time.Now().UTC()
		xValues = []time.Time{now.Add(-72 * time.Hour), now.Add(-48 * time.Hour), now.Add(-24 * time.Hour), now}
		yValues = []float64{0, 0, 0, data.GeomagData.KIndex}
	}

	// Build main line series with dots (order from extractKIndexPoints is chronological)
	mainSeries := chart.TimeSeries{
		Name: "K-index Trend",
		Style: chart.Style{
			// Make the line white (invisible on white background) and keep only dots visible
			StrokeColor: drawing.Color{R: 255, G: 255, B: 255, A: 255},
			StrokeWidth: 0,
			DotColor:   drawing.Color{R: 51, G: 102, B: 204, A: 255},
			DotWidth:   3,
		},
		XValues: xValues,
		YValues: yValues,
	}

	// Determine earliest/latest from data; ensure display window covers up to 72 hours starting at earliest point
	earliestTime := xValues[0]
	latestTime := xValues[len(xValues)-1]
	if latestTime.Sub(earliestTime) < 72*time.Hour {
		// Extend chart window to a full 72 hours, anchoring at the first data point
		latestTime = earliestTime.Add(72 * time.Hour)
	}

	// Compute dynamic Y range (cap to 9.0, never below 5.0 max)
	maxY := 0.0
	for _, v := range yValues {
		if v > maxY {
			maxY = v
		}
	}
	yMax := maxY
	if yMax < 5.0 {
		yMax = 5.0
	}
	if yMax > 9.0 {
		yMax = 9.0
	}
	// Round up to nearest 0.5 for nicer headroom
	yMax = float64(int(yMax*2+0.999)) / 2.0

	// Build ticks every 12 hours
	timeTickInterval := 12
	timeTicks := make([]chart.Tick, 0, 72/timeTickInterval+1)
	for i := 0; i <= 72; i += timeTickInterval {
		tickTime := earliestTime.Add(time.Duration(i) * time.Hour)
		timeTicks = append(timeTicks, chart.Tick{
			Value: chart.TimeToFloat64(tickTime),
			Label: tickTime.UTC().Format("Jan 02 15:04"),
		})
	}

	graph := chart.Chart{
		Title: "K-index Trend (72 Hours)",
		TitleStyle: chart.Style{
			FontSize: 16,
			FontColor: drawing.ColorBlack,
		},
		Background: chart.Style{
			Padding: chart.Box{
				Top:    40,
				Left:   70,
				Right:  20,
				Bottom: 70,
			},
		},
		Height: 520,
		Width:  1100,
		XAxis: chart.XAxis{
			Name: "Time (UTC)",
			NameStyle: chart.Style{
				FontSize: 12,
			},
			Style: chart.Style{
				FontSize: 11,
			},
			// Vertical grid lines at major ticks
			GridMajorStyle: chart.Style{
				StrokeColor:     drawing.Color{R: 200, G: 200, B: 200, A: 255},
				StrokeWidth:     1,
				StrokeDashArray: []float64{2, 3},
			},
			ValueFormatter: nil,
			Ticks: timeTicks,
			Range: &chart.ContinuousRange{
				Min: chart.TimeToFloat64(earliestTime),
				Max: chart.TimeToFloat64(latestTime),
			},
		},
		YAxis: chart.YAxis{
			Name: "K-index",
			NameStyle: chart.Style{
				FontSize: 12,
			},
			Style: chart.Style{
				FontSize: 11,
			},
			Range: &chart.ContinuousRange{
				Min: 0.0,
				Max: yMax,
			},
			Ticks: buildKpTicks(yMax),
		},
		Series: []chart.Series{},
	}

	// Add vertical grid lines at every tick as individual series (drawn beneath data)
	for _, tk := range timeTicks {
		t := chart.TimeFromFloat64(tk.Value)
		graph.Series = append(graph.Series, chart.TimeSeries{
			Style: chart.Style{
				StrokeColor:     drawing.Color{R: 200, G: 200, B: 200, A: 120},
				StrokeWidth:     1,
				StrokeDashArray: []float64{2, 3},
			},
			XValues: []time.Time{t, t},
			YValues: []float64{0, yMax},
		})
	}

	// Add color-coded background zones for K-index levels
	if len(xValues) > 0 {
		minTime := earliestTime
		maxTime := latestTime

		// Quiet zone (0-2): Green line (more visible)
		graph.Series = append(graph.Series, chart.TimeSeries{
			Name: "Quiet Zone (K≤2)",
			Style: chart.Style{
				StrokeColor:     drawing.Color{R: 0, G: 200, B: 0, A: 180},
				StrokeWidth:     2,
				StrokeDashArray: []float64{4, 4},
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

		// Unsettled zone (3): Yellow line (more visible)
		graph.Series = append(graph.Series, chart.TimeSeries{
			Name: "Unsettled (K=3)",
			Style: chart.Style{
				StrokeColor:     drawing.Color{R: 255, G: 215, B: 0, A: 220}, // Golden
				StrokeWidth:     2,
				StrokeDashArray: []float64{4, 4},
			},
			XValues: []time.Time{minTime, maxTime},
			YValues: []float64{3, 3},
		})
	}

	// Add EMA smoothing line as a guide
	ema := &chart.EMASeries{
		Name:        "EMA(5)",
		Period:      5,
		InnerSeries: mainSeries,
		Style: chart.Style{
			StrokeColor: drawing.Color{R: 30, G: 80, B: 200, A: 220},
			StrokeWidth: 5,
		},
	}
	graph.Series = append(graph.Series, ema)

	// Finally add the main data dots on top so they are not obscured by lines
	graph.Series = append(graph.Series, mainSeries)

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

// filterKIndexRecent filters K-index data to only include entries from the last 72 hours
// and sorts them chronologically
func filterKIndexRecent(data []models.NOAAKIndexResponse) []models.NOAAKIndexResponse {
	if len(data) == 0 {
		return data
	}

	now := time.Now().UTC()
	// Use 72 hours for K-index history to show 3 days of data
	historyStart := now.Add(-72 * time.Hour)

	// Filter for recent data
	var filtered []models.NOAAKIndexResponse
	for _, d := range data {
		t, err := parseNOAATime(d.TimeTag)
		if err != nil {
			continue
		}
		if t.After(historyStart) {
			filtered = append(filtered, d)
		}
	}

	// Sort by time to ensure proper ordering
	sort.Slice(filtered, func(i, j int) bool {
		ti, _ := parseNOAATime(filtered[i].TimeTag)
		tj, _ := parseNOAATime(filtered[j].TimeTag)
		return ti.Before(tj)
	})

	return filtered
}

// extractKIndexPoints converts sourceData NOAAKIndex into time/value slices with robust parsing
func extractKIndexPoints(sourceData *models.SourceData) ([]time.Time, []float64) {
	var points []struct{ t time.Time; v float64 }
	if sourceData == nil || len(sourceData.NOAAKIndex) == 0 {
		return nil, nil
	}
	for _, d := range sourceData.NOAAKIndex {
		t, err := parseNOAATime(d.TimeTag)
		if err != nil {
			continue
		}
		val := d.EstimatedKp
		if val == 0 {
			val = d.KpIndex
		}
		points = append(points, struct{ t time.Time; v float64 }{t.UTC(), val})
	}
	if len(points) == 0 {
		return nil, nil
	}
	// Sort chronologically
	sort.Slice(points, func(i, j int) bool { return points[i].t.Before(points[j].t) })
	// Anchor 72h window to latest data timestamp
	windowEnd := points[len(points)-1].t
	windowStart := windowEnd.Add(-72 * time.Hour)
	var x []time.Time
	var y []float64
	for _, p := range points {
		if p.t.Before(windowStart) {
			continue
		}
		x = append(x, p.t)
		y = append(y, p.v)
	}
	return x, y
}

// buildKpTicks generates ticks for Kp index values
func buildKpTicks(yMax float64) []chart.Tick {
    maxInt := int(yMax + 0.0001)
    if maxInt > 9 { maxInt = 9 }
    ticks := make([]chart.Tick, 0, maxInt+1)
    for i := 0; i <= maxInt; i++ {
        label := strconv.Itoa(i)
        switch i {
        case 2:
            label = "2 (Quiet)"
        case 4:
            label = "4 (Active)"
        case 5:
            label = "5 (Storm)"
        }
        ticks = append(ticks, chart.Tick{Value: float64(i), Label: label})
    }
    return ticks
}

// parseNOAATime tries multiple layouts used by NOAA responses
func parseNOAATime(s string) (time.Time, error) {
	layouts := []string{
		"2006-01-02T15:04:05",       // ISO without Z
		"2006-01-02T15:04:05Z07:00", // ISO with zone
		"2006-01-02 15:04:05.000",   // Space with milliseconds
		"2006-01-02 15:04:05",       // Space without milliseconds
	}
	var lastErr error
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		} else {
			lastErr = err
		}
	}
	return time.Time{}, lastErr
}
