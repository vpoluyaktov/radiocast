package charts

import (
	"encoding/json"
	"fmt"

	"radiocast/internal/models"
)

// generateBandConditionsSnippet builds an ECharts heatmap-like dot matrix for 24h x bands
func (cg *ChartGenerator) generateBandConditionsSnippet(data *models.PropagationData) (ChartSnippet, error) {
	// No debug needed
	id := "chart-band-conditions"

	bands := []string{"10m", "12m", "15m", "17m", "20m", "40m", "80m"}
	// map band to day/night strings
	type bn struct{ day, night string }
	bc := map[string]bn{
		"80m": {data.BandData.Band80m.Day, data.BandData.Band80m.Night},
		"40m": {data.BandData.Band40m.Day, data.BandData.Band40m.Night},
		"20m": {data.BandData.Band20m.Day, data.BandData.Band20m.Night},
		"17m": {data.BandData.Band17m.Day, data.BandData.Band17m.Night},
		"15m": {data.BandData.Band15m.Day, data.BandData.Band15m.Night},
		"12m": {data.BandData.Band12m.Day, data.BandData.Band12m.Night},
		"10m": {data.BandData.Band10m.Day, data.BandData.Band10m.Night},
	}

	// build 24h x bands points; encode condition as numeric for visual encoding
	points := make([][]interface{}, 0, 24*len(bands))
	for row, band := range bands {
		b := bc[band]
		for h := 0; h < 24; h++ {
			cond := b.night
			if h >= 6 && h < 18 { cond = b.day }
			// Ensure we're correctly mapping the condition to its numeric value
			val := cg.conditionToValue(cond)
			// Only use 3 elements per data point to match the working example
			points = append(points, []interface{}{h, row, val})
		}
	}

	// visualMap for categories mapped to colors
	// 0 Closed (black), 1 Poor (red), 2 Fair (orange), 3 Good (yellow), 4 Excellent (green)
	option := map[string]interface{}{
		"title": map[string]interface{}{"text": "HF Band Conditions (24h Matrix)", "left": "center"},
		"tooltip": map[string]interface{}{
			"position": "top",
			"formatter": `function(params) {
				var bands = ['10m', '12m', '15m', '17m', '20m', '40m', '80m'];
				var value = params.data[2];
				// Ensure we're correctly mapping all five condition levels
				var label = value === 0 ? 'Closed' : 
						   value === 1 ? 'Poor' : 
						   value === 2 ? 'Fair' : 
						   value === 3 ? 'Good' : 
						   value === 4 ? 'Excellent' : 'Unknown';
				return label + ' | ' + bands[params.data[1]] + ' @ ' + params.data[0] + ':00';
			}`,
		},
		"grid": map[string]interface{}{"left": 110, "right": 40, "bottom": 80, "top": 60},
		"xAxis": map[string]interface{}{
			"type": "category", 
			"data": hours24(), 
			"name": "UTC Hour",
			"splitArea": map[string]interface{}{"show": true},
		},
		"yAxis": map[string]interface{}{
			"type": "category", 
			"data": bands, 
			"name": "Band",
			"splitArea": map[string]interface{}{"show": true},
		},
		"visualMap": map[string]interface{}{
			"type": "piecewise",
			"orient": "horizontal", 
			"left": "center", 
			"bottom": 30,
			"showLabel": true,
			"pieces": []interface{}{
				map[string]interface{}{"value": 0, "label": "Closed", "color": "#000000"},
				map[string]interface{}{"value": 1, "label": "Poor", "color": "#dc3545"},
				map[string]interface{}{"value": 2, "label": "Fair", "color": "#fd7e14"},
				map[string]interface{}{"value": 3, "label": "Good", "color": "#ffc107"},
				map[string]interface{}{"value": 4, "label": "Excellent", "color": "#28a745"},
			},
		},
		"series": []interface{}{
			map[string]interface{}{
				"type": "heatmap",
				"data": points,
				"label": map[string]interface{}{"show": false},
				"emphasis": map[string]interface{}{"itemStyle": map[string]interface{}{"shadowBlur": 10, "shadowColor": "rgba(0,0,0,0.3)"}},
				"itemStyle": map[string]interface{}{
					"borderWidth": 1,
					"borderColor": "#f5f5f5",
				},
			},
		},
	}

	// Extract the formatter function string before JSON serialization
	formatterFunc := option["tooltip"].(map[string]interface{})["formatter"].(string)
	
	// Remove the formatter from the option map for proper JSON serialization
	delete(option["tooltip"].(map[string]interface{}), "formatter")
	
	optJSON, err := json.Marshal(option)
	if err != nil { return ChartSnippet{}, err }
	
	div := fmt.Sprintf("<div id=\"%s\" style=\"width:100%%;height:500px;\"></div>", id)
	
	// Create the script with the formatter as a proper JavaScript function, not a string
	script := fmt.Sprintf(`<script>(function(){
	var el=document.getElementById('%s');
	if(!el)return;
	var c=echarts.init(el);
	var option=%s;
	// Set the formatter as a JavaScript function, not a string
	option.tooltip = option.tooltip || {};
	option.tooltip.position = 'top';
	option.tooltip.formatter = %s;
	c.setOption(option);
	window.addEventListener('resize',function(){c.resize();});
})();</script>`, id, string(optJSON), formatterFunc)

	return ChartSnippet{ID: id, Title: "HF Band Conditions (24h)", Div: div, Script: script}, nil
}

func hours24() []string {
	arr := make([]string, 24)
	for i := 0; i < 24; i++ { if i < 10 { arr[i] = fmt.Sprintf("0%d", i) } else { arr[i] = fmt.Sprintf("%d", i) } }
	return arr
}
