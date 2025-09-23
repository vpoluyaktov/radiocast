package charts

import (
	"fmt"
	"strings"

	"radiocast/internal/models"
)

// generateSpaceWeatherDashboardSnippet builds a combined panel with individual space weather gauges
func (cg *ChartGenerator) generateSpaceWeatherDashboardSnippet(data *models.PropagationData) (ChartSnippet, error) {
	if data == nil {
		return ChartSnippet{}, fmt.Errorf("data cannot be nil")
	}
	
	id := "chart-space-weather-dashboard"

	// Generate individual gauge snippets
	xrayGauge, err := cg.generateXRayGaugeSnippet(data)
	if err != nil {
		return ChartSnippet{}, fmt.Errorf("failed to generate X-ray gauge: %w", err)
	}

	solarWindGauge, err := cg.generateSolarWindGaugeSnippet(data)
	if err != nil {
		return ChartSnippet{}, fmt.Errorf("failed to generate Solar Wind gauge: %w", err)
	}

	auroraGauge, err := cg.generateAuroraGaugeSnippet(data)
	if err != nil {
		return ChartSnippet{}, fmt.Errorf("failed to generate Aurora gauge: %w", err)
	}

	// Combine all scripts (remove duplicate ECharts script tags)
	var allScripts []string
	
	// Extract script content (without <script> tags) from each gauge
	xrayScript := extractScriptContent(xrayGauge.Script)
	solarWindScript := extractScriptContent(solarWindGauge.Script)
	auroraScript := extractScriptContent(auroraGauge.Script)
	
	if xrayScript != "" {
		allScripts = append(allScripts, xrayScript)
	}
	if solarWindScript != "" {
		allScripts = append(allScripts, solarWindScript)
	}
	if auroraScript != "" {
		allScripts = append(allScripts, auroraScript)
	}

	// Combine all scripts into one
	combinedScript := fmt.Sprintf("<script>%s</script>", strings.Join(allScripts, "\n"))

	// Create a responsive grid layout for the gauges
	combinedDiv := fmt.Sprintf(`<div class="space-weather-grid" style="display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; margin: 20px 0;">
		%s
		%s
		%s
	</div>`, xrayGauge.Div, solarWindGauge.Div, auroraGauge.Div)

	// Create complete HTML snippet
	completeHTML := fmt.Sprintf(`<script src="https://cdn.jsdelivr.net/npm/echarts@5.4.3/dist/echarts.min.js"></script>
<div class="chart-container">
	<h3>Space Weather Dashboard</h3>
	%s
</div>
%s`, combinedDiv, combinedScript)

	return ChartSnippet{ID: id, Title: "Space Weather Dashboard", Div: combinedDiv, Script: combinedScript, HTML: completeHTML}, nil
}

