package charts

import (
	"encoding/json"
	"fmt"

	"radiocast/internal/models"
)

// generateSolarWindGaugeSnippet builds an ECharts gauge for solar wind speed
func (cg *ChartGenerator) generateSolarWindGaugeSnippet(data *models.PropagationData) (ChartSnippet, error) {
	if data == nil {
		return ChartSnippet{}, fmt.Errorf("data cannot be nil")
	}
	
	id := "chart-solar-wind-gauge"
	solarWindSpeed := data.SolarData.SolarWindSpeed
	
	// Determine status text based on solar wind speed
	var statusText string
	switch {
	case solarWindSpeed <= 350:
		statusText = "Slow"
	case solarWindSpeed <= 500:
		statusText = "Normal"
	case solarWindSpeed <= 650:
		statusText = "Fast"
	default:
		statusText = "Very Fast"
	}

	option := map[string]interface{}{
		"tooltip": map[string]interface{}{
			"formatter": "{a} <br/>{b} : {c} km/s",
		},
		"series": []interface{}{
			map[string]interface{}{
				"name": "Solar Wind Speed",
				"type": "gauge",
				"min": 200,
				"max": 800,
				"splitNumber": 6,
				"radius": "80%",
				"axisLine": map[string]interface{}{
					"lineStyle": map[string]interface{}{
						"width": 20,
						"color": [][]interface{}{
							{0.25, "#28a745"}, // 200-350 km/s - Green (Excellent)
							{0.5, "#ffc107"},  // 350-500 km/s - Yellow (Good)
							{0.75, "#fd7e14"}, // 500-650 km/s - Orange (Fair)
							{1.0, "#dc3545"},  // 650-800 km/s - Red (Poor)
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
					"formatter": fmt.Sprintf("%.0f km/s\n%s", solarWindSpeed, statusText),
					"color": "inherit",
					"fontSize": 14,
					"fontWeight": "bold",
					"offsetCenter": []interface{}{0, "80%"},
				},
				"data": []interface{}{
					map[string]interface{}{
						"value": solarWindSpeed,
						"name": "Solar Wind Speed",
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
	<h4>Solar Wind Speed</h4>
	%s
</div>
%s`, div, script)

	return ChartSnippet{ID: id, Title: "Solar Wind", Div: div, Script: script, HTML: completeHTML}, nil
}
