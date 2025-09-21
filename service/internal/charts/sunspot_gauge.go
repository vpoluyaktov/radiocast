package charts

import (
	"encoding/json"
	"fmt"

	"radiocast/internal/models"
)

// generateSunspotGaugeSnippet builds an ECharts gauge chart for Sunspot Number
func (cg *ChartGenerator) generateSunspotGaugeSnippet(data *models.PropagationData) (ChartSnippet, error) {
	if data == nil {
		return ChartSnippet{}, fmt.Errorf("data cannot be nil")
	}
	
	id := "chart-sunspot-gauge"
	sunspotNumber := float64(data.SolarData.SunspotNumber)

	// Determine status text based on Sunspot Number
	var statusText string
	switch {
	case sunspotNumber == 0:
		statusText = "None"
	case sunspotNumber < 20:
		statusText = "Very Low"
	case sunspotNumber < 50:
		statusText = "Low"
	case sunspotNumber < 100:
		statusText = "Moderate"
	case sunspotNumber < 150:
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
				"name": "Sunspot Number",
				"type": "gauge",
				"min": 0,
				"max": 200,
				"splitNumber": 4,
				"radius": "80%",
				"axisLine": map[string]interface{}{
					"lineStyle": map[string]interface{}{
						"width": 20,
						"color": [][]interface{}{
							{0.1, "#6c757d"},  // 0-20: Gray (None/Very Low)
							{0.25, "#dc3545"}, // 20-50: Red (Low)
							{0.5, "#fd7e14"},  // 50-100: Orange (Moderate)
							{0.75, "#ffc107"}, // 100-150: Yellow (High)
							{1.0, "#28a745"},  // 150-200: Green (Very High)
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
					"formatter": fmt.Sprintf("%.0f\n%s", sunspotNumber, statusText),
					"color": "inherit",
					"fontSize": 14,
					"fontWeight": "bold",
					"offsetCenter": []interface{}{0, "60%"},
				},
				"data": []interface{}{
					map[string]interface{}{
						"value": sunspotNumber,
						"name": "Sunspot Number",
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
	<h4>Sunspot Number</h4>
	%s
</div>
%s`, div, script)

	return ChartSnippet{ID: id, Title: "Sunspot Number Gauge", Div: div, Script: script, HTML: completeHTML}, nil
}
