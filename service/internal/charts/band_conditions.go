package charts

import (
    "fmt"
    "math"
    "os"
    "path/filepath"

    "github.com/wcharczuk/go-chart/v2"
    "github.com/wcharczuk/go-chart/v2/drawing"

    "radiocast/internal/models"
)

// heatmapCell represents one matrix cell
type heatmapCell struct {
    x0, x1 float64
    y0, y1 float64
    color  drawing.Color
    label  string
}

// heatmapSeries renders a grid of rectangles with labels
type heatmapSeries struct {
    Cells []heatmapCell
}

// legendOnlySeries is a dummy series used only to populate the chart legend
type legendOnlySeries struct {
    name  string
    color drawing.Color
}

func (ls legendOnlySeries) GetName() string           { return ls.name }
func (ls legendOnlySeries) GetStyle() chart.Style     { return chart.Style{FillColor: ls.color, StrokeColor: ls.color} }
func (ls legendOnlySeries) GetYAxis() chart.YAxisType { return chart.YAxisPrimary }
func (ls legendOnlySeries) Len() int                  { return 0 }
func (ls legendOnlySeries) Validate() error           { return nil }
func (ls legendOnlySeries) Render(r chart.Renderer, canvasBox chart.Box, xrange, yrange chart.Range, defaults chart.Style) {}

func (hs heatmapSeries) GetName() string                 { return "heatmap" }
func (hs heatmapSeries) GetStyle() chart.Style           { return chart.Style{} }
func (hs heatmapSeries) GetYAxis() chart.YAxisType       { return chart.YAxisPrimary }
func (hs heatmapSeries) Len() int                        { return len(hs.Cells) }
func (hs heatmapSeries) Validate() error                 { return nil }
func (hs heatmapSeries) Render(r chart.Renderer, canvasBox chart.Box, xrange, yrange chart.Range, defaults chart.Style) {
    for _, c := range hs.Cells {
        x0 := canvasBox.Left + xrange.Translate(c.x0)
        x1 := canvasBox.Left + xrange.Translate(c.x1)
        y0 := canvasBox.Bottom - yrange.Translate(c.y0)
        y1 := canvasBox.Bottom - yrange.Translate(c.y1)
        if x1 < x0 {
            x0, x1 = x1, x0
        }
        if y1 < y0 {
            y0, y1 = y1, y0
        }
        // add small gutters so cells don't touch each other or axes
        gutter := 10 // pixels
        if (x1 - x0) > 2*gutter {
            x0 += gutter
            x1 -= gutter
        }
        if (y1 - y0) > 2*gutter {
            y0 += gutter
            y1 -= gutter
        }
        // draw filled circle (big dot) at cell center with subtle white border
        cx := (x0 + x1) / 2
        cy := (y0 + y1) / 2
        // radius fits inside cell with a small margin
        w := x1 - x0
        h := y1 - y0
        rpx := w
        if h < rpx {
            rpx = h
        }
        radius := rpx / 2
        if radius > 12 { // cap for aesthetics
            radius = 12
        }
        // approximate circle with polygon path
        const steps = 32
        r.SetFillColor(c.color)
        for i := 0; i <= steps; i++ {
            angle := 2 * math.Pi * float64(i) / float64(steps)
            px := cx + int(float64(radius)*math.Cos(angle))
            py := cy + int(float64(radius)*math.Sin(angle))
            if i == 0 {
                r.MoveTo(px, py)
            } else {
                r.LineTo(px, py)
            }
        }
        r.Close()
        r.Fill()
        // stroke border
        r.SetStrokeColor(drawing.Color{R: 255, G: 255, B: 255, A: 120})
        r.SetStrokeWidth(1)
        r.Stroke()
    }
}

// generateBandConditionsChart creates a time-based heatmap for band conditions
func (cg *ChartGenerator) generateBandConditionsChart(data *models.PropagationData) (string, error) {
    filename := filepath.Join(cg.outputDir, "band_conditions.png")

    // Bands from high to low frequency (top to bottom)
    bands := []string{"6m", "10m", "12m", "15m", "17m", "20m", "40m", "80m"}

    // Build hourly condition grid using simple day/night split
    // Day: 06–17 UTC -> use Day condition; Night: otherwise -> Night condition
    type bandCond struct{ day, night string }
    bc := map[string]bandCond{
        "80m": {data.BandData.Band80m.Day, data.BandData.Band80m.Night},
        "40m": {data.BandData.Band40m.Day, data.BandData.Band40m.Night},
        "20m": {data.BandData.Band20m.Day, data.BandData.Band20m.Night},
        "17m": {data.BandData.Band17m.Day, data.BandData.Band17m.Night},
        "15m": {data.BandData.Band15m.Day, data.BandData.Band15m.Night},
        "12m": {data.BandData.Band12m.Day, data.BandData.Band12m.Night},
        "10m": {data.BandData.Band10m.Day, data.BandData.Band10m.Night},
        "6m":  {data.BandData.Band6m.Day, data.BandData.Band6m.Night},
    }

    // Prepare axes ranges
    xTicks := make([]chart.Tick, 0, 24)
    for h := 0; h < 24; h++ {
        xTicks = append(xTicks, chart.Tick{Value: float64(h) + 0.5, Label: fmt.Sprintf("%02d", h)})
    }

    yTicks := make([]chart.Tick, 0, len(bands))
    for i, b := range bands {
        yTicks = append(yTicks, chart.Tick{Value: float64(i), Label: b})
    }

    graph := chart.Chart{
        Title: "HF Band Conditions (24h Matrix)",
        TitleStyle: chart.Style{FontSize: 18, FontColor: drawing.ColorBlack},
        Background: chart.Style{Padding: chart.Box{Top: 60, Left: 110, Right: 80, Bottom: 130}},
        Width:  1200,
        Height: 600,
        XAxis: chart.XAxis{
            Name: "UTC Hour",
            NameStyle: chart.Style{FontSize: 12},
            Style: chart.Style{FontSize: 11},
            Ticks: xTicks,
            Range: &chart.ContinuousRange{Min: 0, Max: 24},
            GridMajorStyle: chart.Style{StrokeColor: drawing.Color{R: 220, G: 220, B: 220, A: 255}, StrokeWidth: 1, StrokeDashArray: []float64{2, 3}},
        },
        YAxis: chart.YAxis{
            Name: "Band",
            NameStyle: chart.Style{FontSize: 12},
            Style: chart.Style{FontSize: 11},
            Ticks: yTicks,
            Range: &chart.ContinuousRange{Min: -0.5, Max: float64(len(bands)) - 0.5},
            GridMajorStyle: chart.Style{StrokeColor: drawing.Color{R: 230, G: 230, B: 230, A: 255}, StrokeWidth: 1},
        },
    }

    hs := heatmapSeries{}
    // Fill cells
    for row, band := range bands {
        b := bc[band]
        for h := 0; h < 24; h++ {
            cond := b.night
            if h >= 6 && h < 18 {
                cond = b.day
            }
            label := cg.abbreviateCondition(cond)
            color := cg.getEnhancedConditionColor(cg.conditionToValue(cond))
            hs.Cells = append(hs.Cells, heatmapCell{
                x0: float64(h), x1: float64(h+1), y0: float64(row)-0.5, y1: float64(row)+0.5,
                color: color, label: label,
            })
        }
    }
    graph.Series = append(graph.Series, hs)

    // add legend entries using dummy series so the legend shows required mapping
    legendItems := []legendOnlySeries{
        {name: "Excellent", color: drawing.Color{R: 40, G: 167, B: 69, A: 255}},  // 🟢
        {name: "Good",      color: drawing.Color{R: 51, G: 102, B: 204, A: 255}}, // 🔵
        {name: "Fair",      color: drawing.Color{R: 255, G: 193, B: 7, A: 255}},  // 🟡
        {name: "Poor",      color: drawing.Color{R: 220, G: 53, B: 69, A: 255}},  // 🔴
        {name: "Closed",    color: drawing.Color{R: 30, G: 30, B: 30, A: 255}},   // ⚫
    }
    for _, li := range legendItems {
        graph.Series = append(graph.Series, li)
    }
    // render built-in legend (potentially on right; bottom placement depends on library support)
    graph.Elements = []chart.Renderable{chart.Legend(&graph)}

    f, err := os.Create(filename)
    if err != nil {
        return "", fmt.Errorf("failed to create band conditions chart file: %w", err)
    }
    defer f.Close()

    if err := graph.Render(chart.PNG, f); err != nil {
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
        // 🟢 Excellent -> green
        return drawing.Color{R: 40, G: 167, B: 69, A: 255}
    case value >= 3:
        // 🔵 Good -> blue
        return drawing.Color{R: 51, G: 102, B: 204, A: 255}
    case value >= 2:
        // 🟡 Fair -> yellow
        return drawing.Color{R: 255, G: 193, B: 7, A: 255}
    case value >= 1:
        // 🔴 Poor -> red
        return drawing.Color{R: 220, G: 53, B: 69, A: 255}
    default:
        // ⚫ Closed -> black/dark
        return drawing.Color{R: 30, G: 30, B: 30, A: 255}
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
        // Default to Poor instead of N/A to avoid confusing labels when data is missing/unknown
        return "Poor"
    }
}
