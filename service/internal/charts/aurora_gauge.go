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

	option := map[string]interface{}{
		"title": map[string]interface{}{
			"text": "Aurora Activity",
			"left": "center",
			"top": "2%",
			"textStyle": map[string]interface{}{
				"fontSize": 16,
				"fontWeight": "bold",
			},
		},
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
				"axisLine": map[string]interface{}{
					"lineStyle": map[string]interface{}{
						"width": 10,
						"color": [][]interface{}{
							{0.3, "#67e0e3"}, // Quiet - blue
							{0.6, "#37a2da"}, // Minor - light blue
							{0.8, "#ffdb5c"}, // Moderate - yellow
							{1.0, "#ff9f7f"}, // Strong - orange
						},
					},
				},
				"pointer": map[string]interface{}{
					"width": 6,
				},
				"detail": map[string]interface{}{
					"formatter": fmt.Sprintf("%.0f", auroraLevel),
					"fontSize": 16,
					"fontWeight": "bold",
					"offsetCenter": []string{"0%", "40%"},
				},
				"data": []interface{}{
					map[string]interface{}{"value": auroraLevel, "name": ""},
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
	<h3>Aurora Activity</h3>
	%s
</div>
%s`, div, script)

	return ChartSnippet{ID: id, Title: "Aurora Activity", Div: div, Script: script, HTML: completeHTML}, nil
}
