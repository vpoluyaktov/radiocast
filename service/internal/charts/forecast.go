package charts

import (
	"encoding/json"
	"fmt"
	"time"

	"radiocast/internal/models"
)

// generateForecastSnippet builds an ECharts bar chart snippet for the 3-day K-index forecast
func (cg *ChartGenerator) generateForecastSnippet(data *models.PropagationData) (ChartSnippet, error) {
	if data == nil {
		return ChartSnippet{}, fmt.Errorf("data cannot be nil")
	}
	
	id := "chart-forecast"

	// Extract K-index forecasts using existing parser
	kIndexValues := []float64{
		cg.parseKIndexForecast(data.Forecast.Today.KIndexForecast),
		cg.parseKIndexForecast(data.Forecast.Tomorrow.KIndexForecast),
		cg.parseKIndexForecast(data.Forecast.DayAfter.KIndexForecast),
	}
	for i, v := range kIndexValues {
		if v == 0 {
			kIndexValues[i] = 2.0
		}
	}

	now := time.Now().UTC()
	labels := []string{
		fmt.Sprintf("Today\n%s", now.Format("Jan 02")),
		fmt.Sprintf("Tomorrow\n%s", now.Add(24*time.Hour).Format("Jan 02")),
		fmt.Sprintf("Day After\n%s", now.Add(48*time.Hour).Format("Jan 02")),
	}

	// Per-bar colors matching existing aesthetic
	colors := []string{
		kIndexColorHex(kIndexValues[0]),
		kIndexColorHex(kIndexValues[1]),
		kIndexColorHex(kIndexValues[2]),
	}

	// Build data with itemStyle for individual colors
	seriesData := make([]map[string]interface{}, 0, len(kIndexValues))
	for i, val := range kIndexValues {
		seriesData = append(seriesData, map[string]interface{}{
			"value": val,
			"itemStyle": map[string]interface{}{
				"color": colors[i],
			},
		})
	}

	option := map[string]interface{}{
		// "title": map[string]interface{}{
		// 	"text": "3-Day K-index Forecast",
		// 	"left": "center",
		// },
		"tooltip": map[string]interface{}{
			"trigger": "axis",
			"axisPointer": map[string]interface{}{"type": "shadow"},
		},
		"grid": map[string]interface{}{
			"left": "8%",
			"right": "4%",
			"bottom": "8%",
			"containLabel": true,
		},
		"xAxis": map[string]interface{}{
			"type": "category",
			"data": labels,
			"axisLabel": map[string]interface{}{"color": "#343a40"},
			"axisLine": map[string]interface{}{"lineStyle": map[string]interface{}{"color": "#ced4da"}},
		},
		"yAxis": map[string]interface{}{
			"type": "value",
			"min": 0,
			"max": 6,
			"axisLabel": map[string]interface{}{"color": "#6c757d"},
			"splitLine": map[string]interface{}{"lineStyle": map[string]interface{}{"color": "#e9ecef"}},
		},
		"series": []interface{}{
			map[string]interface{}{
				"type": "bar",
				"data": seriesData,
				"barWidth": "40%",
				"label": map[string]interface{}{"show": false},
			},
		},
	}

	optJSON, err := json.Marshal(option)
	if err != nil {
		return ChartSnippet{}, fmt.Errorf("marshal echarts option: %w", err)
	}

	div := fmt.Sprintf("<div id=\"%s\" style=\"width:100%%;height:360px;\"></div>", id)
	script := fmt.Sprintf(`<script>(function(){
var el=document.getElementById('%s');
if(!el){return;}
var c=echarts.init(el);
var option=%s;
c.setOption(option);
window.addEventListener('resize', function(){ c.resize(); });
})();</script>
`, id, string(optJSON))

	// Create complete HTML snippet with div and script
	completeHTML := fmt.Sprintf(`<script src="https://cdn.jsdelivr.net/npm/echarts@5.4.3/dist/echarts.min.js"></script>
<div class="chart-container">
	<h3>3-Day K-index Forecast</h3>
	%s
</div>
%s`, div, script)

	return ChartSnippet{
		ID:     id,
		Title:  "3-Day K-index Forecast",
		Div:    div,
		Script: script,
		HTML:   completeHTML,
	}, nil
}

// kIndexColorHex maps K-index value to a hex color aligned with existing palette
func kIndexColorHex(k float64) string {
	switch {
	case k >= 5:
		return "#800080" // Purple for storm
	case k >= 4:
		return "#dc3545" // Red for active
	case k >= 3:
		return "#fd7e14" // Orange for unsettled
	case k >= 2:
		return "#ffc107" // Yellow for fair
	default:
		return "#28a745" // Green for quiet
	}
}
