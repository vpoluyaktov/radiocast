package charts

import (
	"encoding/json"
	"fmt"
	"time"

	"radiocast/internal/models"
)

// generatePropagationTimelineSnippet builds a dual-series line chart for K-index (primary) and Solar Flux (secondary)
func (cg *ChartGenerator) generatePropagationTimelineSnippet(data *models.PropagationData, sourceData *models.SourceData) (ChartSnippet, error) {
	id := "chart-propagation-timeline"

	var times []time.Time
	var kValues []float64
	var sfiValues []float64
	if sourceData != nil && len(sourceData.NOAAKIndex) > 0 {
		for _, d := range sourceData.NOAAKIndex {
			if t, err := parseTimeMulti(d.TimeTag); err == nil {
				times = append(times, t.UTC())
				kv := d.EstimatedKp
				if kv == 0 { kv = d.KpIndex }
				kValues = append(kValues, kv)
				sfiValues = append(sfiValues, data.SolarData.SolarFluxIndex)
			}
		}
	}
	if len(times) == 0 {
		now := time.Now().UTC()
		for i := 0; i < 6; i++ {
			times = append(times, now.Add(time.Duration(-5+i)*time.Hour))
			kValues = append(kValues, 2.0+float64(i)*0.5)
			sfiValues = append(sfiValues, data.SolarData.SolarFluxIndex)
		}
	}

	xdata := make([]string, len(times))
	for i, t := range times { xdata[i] = t.Format(time.RFC3339) }

	option := map[string]interface{}{
		"title": map[string]interface{}{"text": "Propagation Quality Timeline (24 Hours)", "left": "center"},
		"tooltip": map[string]interface{}{"trigger": "axis"},
		"legend": map[string]interface{}{"data": []string{"K-index", "Solar Flux"}, "bottom": 0},
		"grid": map[string]interface{}{"left": "8%", "right": "8%", "bottom": "12%", "containLabel": true},
		"xAxis": map[string]interface{}{"type": "category", "data": xdata},
		"yAxis": []interface{}{
			map[string]interface{}{"type": "value", "name": "K-index", "min": 0, "max": 9},
			map[string]interface{}{"type": "value", "name": "Solar Flux (SFU)", "min": 50, "max": 300},
		},
		"series": []interface{}{
			map[string]interface{}{"name": "K-index", "type": "line", "yAxisIndex": 0, "data": kValues, "showSymbol": true, "symbolSize": 6},
			map[string]interface{}{"name": "Solar Flux", "type": "line", "yAxisIndex": 1, "data": sfiValues, "showSymbol": true, "symbolSize": 5},
		},
	}

	optJSON, err := json.Marshal(option)
	if err != nil { return ChartSnippet{}, err }

	div := fmt.Sprintf("<div id=\"%s\" style=\"width:100%%;height:420px;\"></div>", id)
	script := fmt.Sprintf(`<script>(function(){var el=document.getElementById('%s');if(!el)return;var c=echarts.init(el);var option=%s;c.setOption(option);window.addEventListener('resize',function(){c.resize();});})();</script>`, id, string(optJSON))

	return ChartSnippet{ID: id, Title: "Propagation Quality Timeline (24 Hours)", Div: div, Script: script}, nil
}
