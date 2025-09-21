package charts

import (
	"encoding/json"
	"fmt"

	"radiocast/internal/models"
)

// generateSolarFluxGaugeSnippet builds an ECharts gauge chart for Solar Flux (10.7cm)
func (cg *ChartGenerator) generateSolarFluxGaugeSnippet(data *models.PropagationData) (ChartSnippet, error) {
	if data == nil {
		return ChartSnippet{}, fmt.Errorf("data cannot be nil")
	}
	
	id := "chart-solar-flux-gauge"
	solarFlux := data.SolarData.SolarFluxIndex

	// Determine status text based on Solar Flux level
	var statusText string
	switch {
	case solarFlux < 70:
		statusText = "Very Low"
	case solarFlux < 100:
		statusText = "Low"
	case solarFlux < 150:
		statusText = "Moderate"
	case solarFlux < 200:
		statusText = "High"
	default:
		statusText = "Very High"
	}

	option := map[string]interface{}{
		"tooltip": map[string]interface{}{
			"formatter": "{a} <br/>{b} : {c}",
		},
		"series": []interface{}{
			map[string]interface{}{
				"name": "Solar Flux",
				"type": "gauge",
				"min": 50,
				"max": 300,
				"splitNumber": 5,
				"radius": "80%",
				"axisLine": map[string]interface{}{
					"lineStyle": map[string]interface{}{
						"width": 20,
						"color": [][]interface{}{
							{0.2, "#dc3545"},  // 50-100: Red (Very Low/Low)
							{0.4, "#fd7e14"},  // 100-150: Orange (Moderate)
							{0.7, "#ffc107"},  // 150-200: Yellow (High)
							{1.0, "#28a745"},  // 200-300: Green (Very High)
						},
					},
				},
				"pointer": map[string]interface{}{
					"itemStyle": map[string]interface{}{
						"color": "auto",
					},
				},
				"axisTick": map[string]interface{}{
					"distance": -20,
					"length": 8,
					"lineStyle": map[string]interface{}{
						"color": "#fff",
						"width": 2,
					},
				},
				"splitLine": map[string]interface{}{
					"distance": -20,
					"length": 20,
					"lineStyle": map[string]interface{}{
						"color": "#fff",
						"width": 3,
					},
				},
				"axisLabel": map[string]interface{}{
					"color": "inherit",
					"fontSize": 14,
					"distance": 35,
				},
				"detail": map[string]interface{}{
					"valueAnimation": true,
					"formatter": fmt.Sprintf("%.0f\n%s", solarFlux, statusText),
					"color": "inherit",
					"fontSize": 14,
					"fontWeight": "bold",
					"offsetCenter": []interface{}{0, "60%"},
				},
				"data": []interface{}{
					map[string]interface{}{
						"value": solarFlux,
						"name": "Solar Flux",
					},
				},
			},
		},
	}

	optJSON, err := json.Marshal(option)
	if err != nil { 
		return ChartSnippet{}, err 
	}

	div := fmt.Sprintf("<div id=\"%s\" style=\"width:100%%;height:250px;\"></div>", id)
	script := fmt.Sprintf(`<script>(function(){var el=document.getElementById('%s');if(!el)return;var c=echarts.init(el);var option=%s;c.setOption(option);window.addEventListener('resize',function(){c.resize();});})();</script>`, id, string(optJSON))

	// Create complete HTML snippet with div and script
	completeHTML := fmt.Sprintf(`<script src="https://cdn.jsdelivr.net/npm/echarts@5.4.3/dist/echarts.min.js"></script>
<div class="gauge-item">
	<h4>Solar Flux (10.7cm)</h4>
	%s
</div>
%s`, div, script)

	return ChartSnippet{ID: id, Title: "Solar Flux Gauge", Div: div, Script: script, HTML: completeHTML}, nil
}
