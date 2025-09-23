package charts

import (
	"radiocast/internal/models"
)

// ChartGenerator handles creation of ECharts snippets for interactive charts
type ChartGenerator struct {
	outputDir string
}

// NewChartGenerator creates a new chart generator
func NewChartGenerator(outputDir string) *ChartGenerator {
	return &ChartGenerator{
		outputDir: outputDir,
	}
}

// GenerateEChartsSnippetsWithSources builds embeddable go-echarts charts
func (cg *ChartGenerator) GenerateEChartsSnippetsWithSources(data *models.PropagationData, sourceData *models.SourceData) ([]ChartSnippet, error) {
    var snippets []ChartSnippet
    // Gauge Panel (K-index, Solar Flux, Sunspot combined)
    if sn, err := cg.generateGaugePanelSnippet(data); err == nil {
        snippets = append(snippets, sn)
    }
    // K-index Trend (Line + EMA(5) + guide lines)
    if sn, err := cg.generateKIndexTrendSnippet(data, sourceData); err == nil {
        snippets = append(snippets, sn)
    }
    // Forecast (Bar)
    if sn, err := cg.generateForecastSnippet(data); err == nil {
        snippets = append(snippets, sn)
    }
    // Propagation Timeline (Dual-axis Line)
    if sn, err := cg.generatePropagationTimelineSnippet(data, sourceData); err == nil {
        snippets = append(snippets, sn)
    }
    return snippets, nil
}
