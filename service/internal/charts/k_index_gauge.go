package charts

import (
	"encoding/json"
	"fmt"

	"radiocast/internal/models"
)

// generateKIndexGaugeSnippet builds an ECharts gauge chart for current K-index
func (cg *ChartGenerator) generateKIndexGaugeSnippet(data *models.PropagationData) (ChartSnippet, error) {
	if data == nil {
		return ChartSnippet{}, fmt.Errorf("data cannot be nil")
	}
	
	id := "chart-k-index-gauge"
	kIndex := data.GeomagData.KIndex

	// Determine status text based on K-index level
	var statusText string
	switch {
	case kIndex <= 2:
		statusText = "Quiet"
	case kIndex <= 4:
		statusText = "Unsettled"
	case kIndex <= 6:
		statusText = "Active"
	case kIndex <= 8:
		statusText = "Storm"
	default:
		statusText = "Severe Storm"
	}

	option := map[string]interface{}{
		"tooltip": map[string]interface{}{
			"formatter": "{a} <br/>{b} : {c}",
		},
		"series": []interface{}{
			map[string]interface{}{
				"name": "K-index",
				"type": "gauge",
				"min": 0,
				"max": 9,
				"splitNumber": 9,
				"radius": "80%",
				"axisLine": map[string]interface{}{
					"lineStyle": map[string]interface{}{
						"width": 20, // Reduced thickness for better proportion
						"color": [][]interface{}{
							{0.22, "#28a745"}, // 0-2: Green (Quiet)
							{0.44, "#ffc107"}, // 2-4: Yellow (Unsettled)
							{0.67, "#fd7e14"}, // 4-6: Orange (Active)
							{0.89, "#dc3545"}, // 6-8: Red (Storm)
							{1.0, "#6f42c1"},  // 8-9: Purple (Severe Storm)
						},
					},
				},
				"pointer": map[string]interface{}{
					"itemStyle": map[string]interface{}{
						"color": "auto", // Auto color like demo
					},
				},
				"axisTick": map[string]interface{}{
					"distance": -20, // Adjusted for thinner gauge
					"length": 8,
					"lineStyle": map[string]interface{}{
						"color": "#fff",
						"width": 2,
					},
				},
				"splitLine": map[string]interface{}{
					"distance": -20, // Adjusted for thinner gauge
					"length": 20,    // Reduced length for better proportion
					"lineStyle": map[string]interface{}{
						"color": "#fff",
						"width": 3, // Slightly thinner
					},
				},
				"axisLabel": map[string]interface{}{
					"color": "inherit",
					"fontSize": 14,     // Match other gauges
					"distance": 35,     // Adjusted spacing
				},
				"detail": map[string]interface{}{
					"valueAnimation": true,
					"formatter": fmt.Sprintf("%.1f\n%s", kIndex, statusText),
					"color": "inherit",
					"fontSize": 14,     // Match other gauges
					"fontWeight": "bold",
					"offsetCenter": []interface{}{0, "60%"}, // Moved down to avoid overlap
				},
				"data": []interface{}{
					map[string]interface{}{
						"value": kIndex,
						"name": "K-index",
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
	<h4>K-index</h4>
	%s
</div>
%s`, div, script)

	return ChartSnippet{ID: id, Title: "Solar K-index Gauge", Div: div, Script: script, HTML: completeHTML}, nil
}
