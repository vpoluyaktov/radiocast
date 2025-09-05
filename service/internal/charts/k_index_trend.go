package charts

import (
	"encoding/json"
	"fmt"
	"time"

	"radiocast/internal/models"
)

// generateKIndexTrendSnippet builds an ECharts line chart for K-index over last ~72h with EMA(5) and guide lines
func (cg *ChartGenerator) generateKIndexTrendSnippet(data *models.PropagationData, sourceData *models.SourceData) (ChartSnippet, error) {
	id := "chart-k-index-trend"

	// extract points from source data (prefer NOAA K-index)
	var times []time.Time
	var values []float64
	if sourceData != nil && len(sourceData.NOAAKIndex) > 0 {
		for _, d := range sourceData.NOAAKIndex {
			// try multiple layouts
			if t, err := parseTimeMulti(d.TimeTag); err == nil {
				v := d.EstimatedKp
				if v == 0 { v = d.KpIndex }
				times = append(times, t.UTC())
				values = append(values, v)
			}
		}
	}
	// fallback minimal data
	if len(times) == 0 {
		now := time.Now().UTC()
		times = []time.Time{now.Add(-72 * time.Hour), now.Add(-48 * time.Hour), now.Add(-24 * time.Hour), now}
		values = []float64{0, 0, 0, data.GeomagData.KIndex}
	}

	// compute EMA(5)
	ema := emaSeries(values, 5)

	// format for echarts: xAxis data as ISO times, series data arrays
	xdata := make([]string, len(times))
	for i, t := range times { xdata[i] = t.Format(time.RFC3339) }

	option := map[string]interface{}{
		// "title": map[string]interface{}{"text": "K-index Trend (72 Hours)", "left": "center"},
		"tooltip": map[string]interface{}{"trigger": "axis"},
		"grid": map[string]interface{}{"left": "8%", "right": "4%", "bottom": "12%", "containLabel": true},
		"xAxis": map[string]interface{}{"type": "category", "data": xdata, "axisLabel": map[string]interface{}{"rotate": 0}},
		"yAxis": map[string]interface{}{"type": "value", "min": 0, "max": 9},
		"series": []interface{}{
			map[string]interface{}{"name": "K-index", "type": "line", "showSymbol": true, "symbolSize": 6, "data": values},
			map[string]interface{}{"name": "EMA(5)", "type": "line", "showSymbol": false, "lineStyle": map[string]interface{}{"width": 2}, "data": ema},
		},
		"legend": map[string]interface{}{"data": []string{"K-index", "EMA(5)"}, "bottom": 0},
		"visualMap": []interface{}{},
		"markLine": map[string]interface{}{"silent": true, "data": []interface{}{
			map[string]interface{}{"yAxis": 2, "name": "Quiet (2)"},
			map[string]interface{}{"yAxis": 3, "name": "Unsettled (3)"},
			map[string]interface{}{"yAxis": 4, "name": "Active (4)"},
		}},
	}

	optJSON, err := json.Marshal(option)
	if err != nil { return ChartSnippet{}, err }

	div := fmt.Sprintf("<div id=\"%s\" style=\"width:100%%;height:420px;\"></div>", id)
	script := fmt.Sprintf(`<script>(function(){var el=document.getElementById('%s');if(!el)return;var c=echarts.init(el);var option=%s;c.setOption(option);window.addEventListener('resize',function(){c.resize();});})();</script>`, id, string(optJSON))

	return ChartSnippet{ID: id, Title: "K-index Trend (72 Hours)", Div: div, Script: script}, nil
}

func emaSeries(vals []float64, period int) []float64 {
	if period <= 1 || len(vals) == 0 { return vals }
	k := 2.0 / (float64(period) + 1.0)
	out := make([]float64, len(vals))
	out[0] = vals[0]
	for i := 1; i < len(vals); i++ {
		out[i] = vals[i]*k + out[i-1]*(1.0-k)
	}
	return out
}

func parseTimeMulti(s string) (time.Time, error) {
	layouts := []string{
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05.000",
		"2006-01-02 15:04:05",
	}
	var last error
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil { return t, nil } else { last = err }
	}
	return time.Time{}, last
}
