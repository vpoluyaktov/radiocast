package charts

import (
	"encoding/json"
	"fmt"

	"radiocast/internal/models"
)

// generateAuroraGaugeSnippet builds an ECharts gauge for aurora activity
func (cg *ChartGenerator) generateAuroraGaugeSnippet(data *models.PropagationData) (ChartSnippet, error) {
	if data == nil {
		return ChartSnippet{}, fmt.Errorf("data cannot be nil")
	}
	
	id := "chart-aurora-gauge"
	auroraLevel := parseFluxValue(data.SolarData.Aurora)
	
	// Determine status text based on aurora activity level
	var statusText string
	switch {
	case auroraLevel <= 2:
		statusText = "Quiet"
	case auroraLevel <= 4:
		statusText = "Minor"
	case auroraLevel <= 6:
		statusText = "Moderate"
	case auroraLevel <= 8:
		statusText = "Strong"
	default:
		statusText = "Extreme"
	}

	option := map[string]interface{}{
		"tooltip": map[string]interface{}{
			"formatter": "{a} <br/>{b} : {c}",
		},
		"series": []interface{}{
			map[string]interface{}{
				"name": "Aurora Activity",
				"type": "gauge",
				"min": 0,
				"max": 9,
				"splitNumber": 9,
				"radius": "80%",
				"axisLine": map[string]interface{}{
					"lineStyle": map[string]interface{}{
						"width": 20,
						"color": [][]interface{}{
							{0.22, "#28a745"}, // 0-2: Green (Excellent)
							{0.44, "#ffc107"}, // 2-4: Yellow (Good)
							{0.67, "#fd7e14"}, // 4-6: Orange (Fair)
							{0.89, "#dc3545"}, // 6-8: Red (Poor)
							{1.0, "#6f42c1"},  // 8-9: Purple (Closed)
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
					"formatter": fmt.Sprintf("%.0f\n%s", auroraLevel, statusText),
					"color": "inherit",
					"fontSize": 14,
					"fontWeight": "bold",
					"offsetCenter": []interface{}{0, "80%"},
				},
				"data": []interface{}{
					map[string]interface{}{
						"value": auroraLevel,
						"name": "Aurora Activity",
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
	<h4>Aurora Activity</h4>
	%s
</div>
%s`, div, script)

	return ChartSnippet{ID: id, Title: "Aurora Activity", Div: div, Script: script, HTML: completeHTML}, nil
}
