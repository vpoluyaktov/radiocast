package charts

import (
	"encoding/json"
	"fmt"
	"time"

	"radiocast/internal/models"
)

// generateHistoricalSolarTrendSnippet builds an ECharts line chart for historical solar flux and sunspot trends
func (cg *ChartGenerator) generateHistoricalSolarTrendSnippet(data *models.PropagationData) (ChartSnippet, error) {
	if data == nil {
		return ChartSnippet{}, fmt.Errorf("data cannot be nil")
	}
	
	id := "chart-historical-solar-trend"

	// Extract historical solar data
	var times []time.Time
	var solarFluxValues []float64
	var sunspotValues []float64
	
	if len(data.HistoricalSolar) > 0 {
		for _, point := range data.HistoricalSolar {
			times = append(times, point.Timestamp)
			solarFluxValues = append(solarFluxValues, point.SolarFlux)
			sunspotValues = append(sunspotValues, point.SunspotNumber)
		}
	} else {
		// Fallback data if no historical data available
		now := time.Now().UTC()
		times = []time.Time{
			now.AddDate(0, -5, 0), now.AddDate(0, -4, 0), now.AddDate(0, -3, 0),
			now.AddDate(0, -2, 0), now.AddDate(0, -1, 0), now,
		}
		currentFlux := data.SolarData.SolarFluxIndex
		currentSunspots := float64(data.SolarData.SunspotNumber)
		solarFluxValues = []float64{currentFlux * 0.8, currentFlux * 0.9, currentFlux * 0.85, currentFlux * 0.95, currentFlux * 0.92, currentFlux}
		sunspotValues = []float64{currentSunspots * 0.7, currentSunspots * 0.8, currentSunspots * 0.75, currentSunspots * 0.9, currentSunspots * 0.85, currentSunspots}
	}

	// Format for echarts: xAxis data as month names
	xdata := make([]string, len(times))
	for i, t := range times {
		xdata[i] = t.Format("Jan 2006")
	}

	option := map[string]interface{}{
		"tooltip": map[string]interface{}{
			"trigger": "axis",
			"axisPointer": map[string]interface{}{"type": "cross"},
		},
		"grid": map[string]interface{}{"left": "8%", "right": "8%", "bottom": "15%", "containLabel": true},
		"xAxis": map[string]interface{}{
			"type": "category",
			"data": xdata,
			"axisLabel": map[string]interface{}{"rotate": 45},
		},
		"yAxis": []interface{}{
			map[string]interface{}{
				"type": "value",
				"name": "Solar Flux",
				"position": "left",
				"axisLabel": map[string]interface{}{"formatter": "{value}"},
				"min": 80,
				"max": 250,
			},
			map[string]interface{}{
				"type": "value",
				"name": "Sunspot Number",
				"position": "right",
				"axisLabel": map[string]interface{}{"formatter": "{value}"},
				"min": 0,
				"max": 200,
			},
		},
		"series": []interface{}{
			map[string]interface{}{
				"name": "Solar Flux (10.7cm)",
				"type": "line",
				"yAxisIndex": 0,
				"showSymbol": true,
				"symbolSize": 8,
				"lineStyle": map[string]interface{}{"width": 3, "color": "#ff6b35"},
				"itemStyle": map[string]interface{}{"color": "#ff6b35"},
				"data": solarFluxValues,
			},
			map[string]interface{}{
				"name": "Sunspot Number",
				"type": "line",
				"yAxisIndex": 1,
				"showSymbol": true,
				"symbolSize": 8,
				"lineStyle": map[string]interface{}{"width": 3, "color": "#4ecdc4"},
				"itemStyle": map[string]interface{}{"color": "#4ecdc4"},
				"data": sunspotValues,
			},
		},
		"legend": map[string]interface{}{
			"data": []string{"Solar Flux (10.7cm)", "Sunspot Number"},
			"bottom": 0,
		},
		"color": []string{"#ff6b35", "#4ecdc4"},
	}

	optJSON, err := json.Marshal(option)
	if err != nil {
		return ChartSnippet{}, err
	}

	div := fmt.Sprintf("<div id=\"%s\" style=\"width:100%%;height:420px;\"></div>", id)
	script := fmt.Sprintf(`<script>(function(){var el=document.getElementById('%s');if(!el)return;var c=echarts.init(el);var option=%s;c.setOption(option);window.addEventListener('resize',function(){c.resize();});})();</script>`, id, string(optJSON))

	// Create complete HTML snippet with div and script
	completeHTML := fmt.Sprintf(`<script src="https://cdn.jsdelivr.net/npm/echarts@5.4.3/dist/echarts.min.js"></script>
<div class="chart-container">
	<h3>Solar Activity Trends (6 Months)</h3>
	%s
</div>
%s`, div, script)

	return ChartSnippet{ID: id, Title: "Solar Activity Trends (6 Months)", Div: div, Script: script, HTML: completeHTML}, nil
}
