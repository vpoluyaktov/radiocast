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

	option := map[string]interface{}{
		"title": map[string]interface{}{
			"text": "X-ray Activity",
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
				"name": "X-ray Flux",
				"type": "gauge",
				"min": 0,
				"max": 10,
				"splitNumber": 5,
				"axisLine": map[string]interface{}{
					"lineStyle": map[string]interface{}{
						"width": 10,
						"color": [][]interface{}{
							{0.2, "#67e0e3"}, // A/B class - blue
							{0.4, "#37a2da"}, // C class - light blue  
							{0.6, "#ffdb5c"}, // M class - yellow
							{0.8, "#ff9f7f"}, // X class - orange
							{1.0, "#fb7293"}, // X+ class - red
						},
					},
				},
				"pointer": map[string]interface{}{
					"width": 6,
				},
				"detail": map[string]interface{}{
					"formatter": xrayFlux,
					"fontSize": 16,
					"fontWeight": "bold",
					"offsetCenter": []string{"0%", "40%"},
				},
				"data": []interface{}{
					map[string]interface{}{"value": getXrayValue(xrayFlux), "name": ""},
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
	<h3>X-ray Activity</h3>
	%s
</div>
%s`, div, script)

	return ChartSnippet{ID: id, Title: "X-ray Activity", Div: div, Script: script, HTML: completeHTML}, nil
}

