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

	option := map[string]interface{}{
		"title": map[string]interface{}{
			"text": "Solar Wind Speed",
			"left": "center",
			"top": "2%",
			"textStyle": map[string]interface{}{
				"fontSize": 16,
				"fontWeight": "bold",
			},
		},
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
				"axisLine": map[string]interface{}{
					"lineStyle": map[string]interface{}{
						"width": 10,
						"color": [][]interface{}{
							{0.3, "#67e0e3"}, // Slow - blue
							{0.6, "#37a2da"}, // Normal - light blue
							{0.8, "#ffdb5c"}, // Fast - yellow
							{1.0, "#ff9f7f"}, // Very fast - orange
						},
					},
				},
				"pointer": map[string]interface{}{
					"width": 6,
				},
				"detail": map[string]interface{}{
					"formatter": fmt.Sprintf("%.0f km/s", solarWindSpeed),
					"fontSize": 16,
					"fontWeight": "bold",
					"offsetCenter": []string{"0%", "40%"},
				},
				"data": []interface{}{
					map[string]interface{}{"value": solarWindSpeed, "name": ""},
				},
			},
		},
	}

	optJSON, err := json.Marshal(option)
	if err != nil {
		return ChartSnippet{}, err
	}

	div := fmt.Sprintf("<div id=\"%s\" style=\"width:100%%;height:300px;\"></div>", id)
	script := fmt.Sprintf(`<script>(function(){var el=document.getElementById('%s');if(!el)return;var c=echarts.init(el);var option=%s;c.setOption(option);window.addEventListener('resize',function(){c.resize();});})();</script>`, id, string(optJSON))

	// Create complete HTML snippet with div and script
	completeHTML := fmt.Sprintf(`<script src="https://cdn.jsdelivr.net/npm/echarts@5.4.3/dist/echarts.min.js"></script>
<div class="chart-container">
	<h3>Solar Wind</h3>
	%s
</div>
%s`, div, script)

	return ChartSnippet{ID: id, Title: "Solar Wind", Div: div, Script: script, HTML: completeHTML}, nil
}
