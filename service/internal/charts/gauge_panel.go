package charts

import (
	"fmt"
	"strings"

	"radiocast/internal/models"
)

// generateGaugePanelSnippet builds a combined panel with K-index, Solar Flux, and Sunspot gauges
func (cg *ChartGenerator) generateGaugePanelSnippet(data *models.PropagationData) (ChartSnippet, error) {
	if data == nil {
		return ChartSnippet{}, fmt.Errorf("data cannot be nil")
	}
	
	id := "chart-gauge-panel"

	// Generate individual gauge snippets
	kIndexGauge, err := cg.generateKIndexGaugeSnippet(data)
	if err != nil {
		return ChartSnippet{}, fmt.Errorf("failed to generate K-index gauge: %w", err)
	}

	solarFluxGauge, err := cg.generateSolarFluxGaugeSnippet(data)
	if err != nil {
		return ChartSnippet{}, fmt.Errorf("failed to generate Solar Flux gauge: %w", err)
	}

	sunspotGauge, err := cg.generateSunspotGaugeSnippet(data)
	if err != nil {
		return ChartSnippet{}, fmt.Errorf("failed to generate Sunspot gauge: %w", err)
	}

	// Combine all scripts (remove duplicate ECharts script tags)
	var allScripts []string
	
	// Extract script content (without <script> tags) from each gauge
	kIndexScript := extractScriptContent(kIndexGauge.Script)
	solarFluxScript := extractScriptContent(solarFluxGauge.Script)
	sunspotScript := extractScriptContent(sunspotGauge.Script)
	
	if kIndexScript != "" {
		allScripts = append(allScripts, kIndexScript)
	}
	if solarFluxScript != "" {
		allScripts = append(allScripts, solarFluxScript)
	}
	if sunspotScript != "" {
		allScripts = append(allScripts, sunspotScript)
	}

	// Create combined HTML with responsive layout
	completeHTML := fmt.Sprintf(`<script src="https://cdn.jsdelivr.net/npm/echarts@5.4.3/dist/echarts.min.js"></script>
<div class="gauge-panel">
	<h3>Current Space Weather Conditions</h3>
	<div class="gauge-container">
		%s
		%s
		%s
	</div>
</div>
<script>
%s
</script>`, 
		extractGaugeItemContent(kIndexGauge.HTML),
		extractGaugeItemContent(solarFluxGauge.HTML),
		extractGaugeItemContent(sunspotGauge.HTML),
		strings.Join(allScripts, "\n"))

	// Combine all divs for the Div field
	combinedDiv := fmt.Sprintf(`<div class="gauge-panel">
	<h3>Current Space Weather Conditions</h3>
	<div class="gauge-container">
		%s
		%s
		%s
	</div>
</div>`, 
		extractGaugeItemContent(kIndexGauge.HTML),
		extractGaugeItemContent(solarFluxGauge.HTML),
		extractGaugeItemContent(sunspotGauge.HTML))

	// Combine all scripts
	combinedScript := fmt.Sprintf("<script>\n%s\n</script>", strings.Join(allScripts, "\n"))

	return ChartSnippet{
		ID:     id, 
		Title:  "Space Weather Gauge Panel", 
		Div:    combinedDiv, 
		Script: combinedScript, 
		HTML:   completeHTML,
	}, nil
}

// extractScriptContent extracts the JavaScript content from a script tag
func extractScriptContent(script string) string {
	// Remove <script> and </script> tags
	content := strings.TrimSpace(script)
	content = strings.TrimPrefix(content, "<script>")
	content = strings.TrimSuffix(content, "</script>")
	return strings.TrimSpace(content)
}

// extractGaugeItemContent extracts the gauge-item div content from HTML
func extractGaugeItemContent(html string) string {
	// Find the gauge-item div
	startTag := `<div class="gauge-item">`
	endTag := `</div>`
	
	startIdx := strings.Index(html, startTag)
	if startIdx == -1 {
		return ""
	}
	
	// Find the matching closing div (simple approach for this case)
	content := html[startIdx:]
	endIdx := strings.LastIndex(content, endTag)
	if endIdx == -1 {
		return ""
	}
	
	return content[:endIdx+len(endTag)]
}
