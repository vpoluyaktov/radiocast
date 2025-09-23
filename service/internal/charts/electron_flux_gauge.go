package charts

import (
	"encoding/json"
	"fmt"

	"radiocast/internal/models"
)

// generateElectronFluxGaugeSnippet builds an ECharts gauge for electron flux
func (cg *ChartGenerator) generateElectronFluxGaugeSnippet(data *models.PropagationData) (ChartSnippet, error) {
	if data == nil {
		return ChartSnippet{}, fmt.Errorf("data cannot be nil")
	}
	
	id := "chart-electron-flux-gauge"
	electronFlux := parseFluxValue(data.SolarData.ElectronFlux)

	option := map[string]interface{}{
		"title": map[string]interface{}{
			"text": "Electron Flux",
			"left": "center",
			"top": "5%",
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
				"name": "Electron Flux",
				"type": "gauge",
				"min": 0,
				"max": 5000,
				"splitNumber": 5,
				"axisLine": map[string]interface{}{
					"lineStyle": map[string]interface{}{
						"width": 10,
						"color": [][]interface{}{
							{0.4, "#67e0e3"}, // Low - blue
							{0.7, "#37a2da"}, // Moderate - light blue
							{0.9, "#ffdb5c"}, // High - yellow
							{1.0, "#ff9f7f"}, // Very high - orange
						},
					},
				},
				"pointer": map[string]interface{}{
					"width": 6,
				},
				"detail": map[string]interface{}{
					"formatter": fmt.Sprintf("%.0f", electronFlux),
					"fontSize": 16,
					"fontWeight": "bold",
					"offsetCenter": []string{"0%", "40%"},
				},
				"data": []interface{}{
					map[string]interface{}{"value": electronFlux, "name": ""},
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
	<h3>Electron Flux</h3>
	%s
</div>
%s`, div, script)

	return ChartSnippet{ID: id, Title: "Electron Flux", Div: div, Script: script, HTML: completeHTML}, nil
}

