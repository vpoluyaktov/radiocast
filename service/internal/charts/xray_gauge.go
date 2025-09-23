package charts

import (
	"encoding/json"
	"fmt"

	"radiocast/internal/models"
)

// generateXRayGaugeSnippet builds an ECharts gauge for X-ray flux
func (cg *ChartGenerator) generateXRayGaugeSnippet(data *models.PropagationData) (ChartSnippet, error) {
	if data == nil {
		return ChartSnippet{}, fmt.Errorf("data cannot be nil")
	}
	
	id := "chart-xray-gauge"
	xrayFlux := data.SolarData.XRayFlux
	
	// Determine status text based on X-ray flux level
	var statusText string
	xrayValue := getXrayValue(xrayFlux)
	switch {
	case xrayValue <= 2:
		statusText = "Quiet"
	case xrayValue <= 4:
		statusText = "Minor"
	case xrayValue <= 6:
		statusText = "Moderate"
	case xrayValue <= 8:
		statusText = "Major"
	default:
		statusText = "Extreme"
	}

	option := map[string]interface{}{
		"tooltip": map[string]interface{}{
			"formatter": "{a} <br/>{b} : {c}",
		},
		"series": []interface{}{
			map[string]interface{}{
				"name": "X-ray Flux",
				"type": "gauge",
				"min": 0,
				"max": 10,
				"splitNumber": 5,
				"radius": "80%",
				"axisLine": map[string]interface{}{
					"lineStyle": map[string]interface{}{
						"width": 20,
						"color": [][]interface{}{
							{0.2, "#28a745"}, // A/B class - Green (Excellent)
							{0.4, "#ffc107"}, // C class - Yellow (Good)  
							{0.6, "#fd7e14"}, // M class - Orange (Fair)
							{0.8, "#dc3545"}, // X class - Red (Poor)
							{1.0, "#6f42c1"}, // X+ class - Purple (Closed)
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
					"formatter": fmt.Sprintf("%s\n%s", xrayFlux, statusText),
					"color": "inherit",
					"fontSize": 14,
					"fontWeight": "bold",
					"offsetCenter": []interface{}{0, "80%"},
				},
				"data": []interface{}{
					map[string]interface{}{
						"value": getXrayValue(xrayFlux),
						"name": "X-ray Flux",
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
	<h4>X-ray Activity</h4>
	%s
</div>
%s`, div, script)

	return ChartSnippet{ID: id, Title: "X-ray Activity", Div: div, Script: script, HTML: completeHTML}, nil
}

