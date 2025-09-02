package charts

import (
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

// GenerateForecastChart creates a forecast chart (exported for testing)
func (cg *ChartGenerator) GenerateForecastChart(data *models.PropagationData) (string, error) {
	return cg.generateForecastChart(data)
}
