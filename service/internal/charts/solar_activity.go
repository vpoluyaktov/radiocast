package charts

import (
	"encoding/json"
	"fmt"

	"radiocast/internal/models"
)

// generateSolarActivitySnippet builds an ECharts bar chart for current solar activity
func (cg *ChartGenerator) generateSolarActivitySnippet(data *models.PropagationData) (ChartSnippet, error) {
	id := "chart-solar-activity"

	labels := []string{"Solar Flux", "Sunspots", "K-index"}
	values := []float64{data.SolarData.SolarFluxIndex, float64(data.SolarData.SunspotNumber), data.GeomagData.KIndex}

	seriesData := make([]map[string]interface{}, 0, len(values))
	for _, v := range values {
		seriesData = append(seriesData, map[string]interface{}{"value": v})
	}

	option := map[string]interface{}{
		// "title": map[string]interface{}{"text": "Current Solar Activity", "left": "center"},
		"tooltip": map[string]interface{}{"trigger": "axis", "axisPointer": map[string]interface{}{"type": "shadow"}},
		"grid": map[string]interface{}{"left": "8%", "right": "4%", "bottom": "8%", "containLabel": true},
		"xAxis": map[string]interface{}{"type": "category", "data": labels},
		"yAxis": map[string]interface{}{"type": "value"},
		"series": []interface{}{map[string]interface{}{"type": "bar", "data": seriesData, "barWidth": "40%"}},
	}

	optJSON, err := json.Marshal(option)
	if err != nil { return ChartSnippet{}, err }

	div := fmt.Sprintf("<div id=\"%s\" style=\"width:100%%;height:360px;\"></div>", id)
	script := fmt.Sprintf(`<script>(function(){var el=document.getElementById('%s');if(!el)return;var c=echarts.init(el);var option=%s;c.setOption(option);window.addEventListener('resize',function(){c.resize();});})();</script>`, id, string(optJSON))

	// Create complete HTML snippet with div and script
	completeHTML := fmt.Sprintf(`<script src="https://cdn.jsdelivr.net/npm/echarts@5.4.3/dist/echarts.min.js"></script>
<div class="chart-container">
	<h3>Current Solar Activity</h3>
	%s
</div>
%s`, div, script)

	return ChartSnippet{ID: id, Title: "Current Solar Activity", Div: div, Script: script, HTML: completeHTML}, nil
}
