package charts

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"radiocast/internal/models"
)

func TestGenerateSolarActivitySnippet(t *testing.T) {
	generator := NewChartGenerator("/test")
	
	// Create test data
	data := &models.PropagationData{
		Timestamp: time.Date(2025, 9, 17, 12, 0, 0, 0, time.UTC),
		SolarData: models.SolarData{
			SolarFluxIndex: 150.5,
			SunspotNumber:  120,
		},
		GeomagData: models.GeomagData{
			KIndex: 3.5,
		},
	}
	
	snippet, err := generator.generateSolarActivitySnippet(data)
	if err != nil {
		t.Fatalf("generateSolarActivitySnippet failed: %v", err)
	}
	
	// Verify basic structure
	if snippet.ID == "" {
		t.Error("Expected non-empty ID")
	}
	if snippet.Title == "" {
		t.Error("Expected non-empty Title")
	}
	if snippet.Div == "" {
		t.Error("Expected non-empty Div")
	}
	if snippet.Script == "" {
		t.Error("Expected non-empty Script")
	}
	if snippet.HTML == "" {
		t.Error("Expected non-empty HTML")
	}
	
	// Verify ID format
	expectedID := "chart-solar-activity"
	if snippet.ID != expectedID {
		t.Errorf("Expected ID '%s', got '%s'", expectedID, snippet.ID)
	}
	
	// Verify title
	expectedTitle := "Current Solar Activity"
	if snippet.Title != expectedTitle {
		t.Errorf("Expected Title '%s', got '%s'", expectedTitle, snippet.Title)
	}
}

func TestGenerateSolarActivitySnippetDivStructure(t *testing.T) {
	generator := NewChartGenerator("/test")
	
	data := &models.PropagationData{
		SolarData: models.SolarData{
			SolarFluxIndex: 100.0,
			SunspotNumber:  50,
		},
		GeomagData: models.GeomagData{
			KIndex: 2.0,
		},
	}
	
	snippet, err := generator.generateSolarActivitySnippet(data)
	if err != nil {
		t.Fatalf("generateSolarActivitySnippet failed: %v", err)
	}
	
	// Verify div contains expected ID
	if !strings.Contains(snippet.Div, "chart-solar-activity") {
		t.Error("Div should contain chart ID")
	}
	
	// Verify div has style attributes
	if !strings.Contains(snippet.Div, "style=") {
		t.Error("Div should contain style attribute")
	}
	
	// Verify div dimensions
	if !strings.Contains(snippet.Div, "width:100%") {
		t.Error("Div should contain width:100%")
	}
	if !strings.Contains(snippet.Div, "height:360px") {
		t.Error("Div should contain height:360px")
	}
}

func TestGenerateSolarActivitySnippetScriptStructure(t *testing.T) {
	generator := NewChartGenerator("/test")
	
	data := &models.PropagationData{
		SolarData: models.SolarData{
			SolarFluxIndex: 200.0,
			SunspotNumber:  80,
		},
		GeomagData: models.GeomagData{
			KIndex: 4.0,
		},
	}
	
	snippet, err := generator.generateSolarActivitySnippet(data)
	if err != nil {
		t.Fatalf("generateSolarActivitySnippet failed: %v", err)
	}
	
	// Verify script structure
	if !strings.HasPrefix(snippet.Script, "<script>") {
		t.Error("Script should start with <script>")
	}
	if !strings.HasSuffix(snippet.Script, "</script>") {
		t.Error("Script should end with </script>")
	}
	
	// Verify script contains echarts initialization
	if !strings.Contains(snippet.Script, "echarts.init") {
		t.Error("Script should contain echarts.init")
	}
	
	// Verify script contains chart ID
	if !strings.Contains(snippet.Script, "chart-solar-activity") {
		t.Error("Script should contain chart ID")
	}
	
	// Verify script contains setOption
	if !strings.Contains(snippet.Script, "setOption") {
		t.Error("Script should contain setOption")
	}
}

func TestGenerateSolarActivitySnippetHTMLStructure(t *testing.T) {
	generator := NewChartGenerator("/test")
	
	data := &models.PropagationData{
		SolarData: models.SolarData{
			SolarFluxIndex: 175.0,
			SunspotNumber:  90,
		},
		GeomagData: models.GeomagData{
			KIndex: 2.5,
		},
	}
	
	snippet, err := generator.generateSolarActivitySnippet(data)
	if err != nil {
		t.Fatalf("generateSolarActivitySnippet failed: %v", err)
	}
	
	// Verify HTML contains ECharts CDN
	if !strings.Contains(snippet.HTML, "echarts") {
		t.Error("HTML should contain echarts reference")
	}
	
	// Verify HTML contains chart container
	if !strings.Contains(snippet.HTML, "chart-container") {
		t.Error("HTML should contain chart-container class")
	}
	
	// Verify HTML contains title
	if !strings.Contains(snippet.HTML, "Current Solar Activity") {
		t.Error("HTML should contain chart title")
	}
	
	// Verify HTML contains both div and script
	if !strings.Contains(snippet.HTML, snippet.Div) {
		t.Error("HTML should contain the div content")
	}
	if !strings.Contains(snippet.HTML, snippet.Script) {
		t.Error("HTML should contain the script content")
	}
}

func TestGenerateSolarActivitySnippetDataValues(t *testing.T) {
	generator := NewChartGenerator("/test")
	
	testCases := []struct {
		name       string
		solarFlux  float64
		sunspots   int
		kIndex     float64
	}{
		{"Normal values", 150.0, 100, 3.0},
		{"High values", 300.0, 200, 8.0},
		{"Low values", 70.0, 0, 0.0},
		{"Zero values", 0.0, 0, 0.0},
		{"Decimal values", 125.7, 75, 2.33},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data := &models.PropagationData{
				SolarData: models.SolarData{
					SolarFluxIndex: tc.solarFlux,
					SunspotNumber:  tc.sunspots,
				},
				GeomagData: models.GeomagData{
					KIndex: tc.kIndex,
				},
			}
			
			snippet, err := generator.generateSolarActivitySnippet(data)
			if err != nil {
				t.Fatalf("generateSolarActivitySnippet failed for %s: %v", tc.name, err)
			}
			
			// Extract and verify the JSON option from script
			scriptContent := snippet.Script
			startIdx := strings.Index(scriptContent, "var option=") + len("var option=")
			endIdx := strings.Index(scriptContent[startIdx:], ";")
			if endIdx == -1 {
				t.Fatalf("Could not find option JSON in script for %s", tc.name)
			}
			
			optionJSON := scriptContent[startIdx : startIdx+endIdx]
			
			var option map[string]interface{}
			err = json.Unmarshal([]byte(optionJSON), &option)
			if err != nil {
				t.Fatalf("Failed to parse option JSON for %s: %v", tc.name, err)
			}
			
			// Verify chart structure
			if option["xAxis"] == nil {
				t.Errorf("Missing xAxis in option for %s", tc.name)
			}
			if option["yAxis"] == nil {
				t.Errorf("Missing yAxis in option for %s", tc.name)
			}
			if option["series"] == nil {
				t.Errorf("Missing series in option for %s", tc.name)
			}
			
			// Verify xAxis data contains expected labels
			xAxis := option["xAxis"].(map[string]interface{})
			xData := xAxis["data"].([]interface{})
			expectedLabels := []string{"Solar Flux", "Sunspots", "K-index"}
			
			if len(xData) != len(expectedLabels) {
				t.Errorf("Expected %d labels, got %d for %s", len(expectedLabels), len(xData), tc.name)
			}
			
			for i, label := range expectedLabels {
				if i < len(xData) && xData[i].(string) != label {
					t.Errorf("Expected label '%s', got '%s' for %s", label, xData[i], tc.name)
				}
			}
		})
	}
}

func TestGenerateSolarActivitySnippetWithNilData(t *testing.T) {
	generator := NewChartGenerator("/test")
	
	_, err := generator.generateSolarActivitySnippet(nil)
	if err == nil {
		t.Error("Expected error with nil data, got none")
	}
}

func TestGenerateSolarActivitySnippetWithEmptyData(t *testing.T) {
	generator := NewChartGenerator("/test")
	
	data := &models.PropagationData{}
	
	snippet, err := generator.generateSolarActivitySnippet(data)
	if err != nil {
		t.Fatalf("generateSolarActivitySnippet failed with empty data: %v", err)
	}
	
	// Should still generate valid snippet with zero values
	if snippet.ID == "" {
		t.Error("Expected non-empty ID with empty data")
	}
	if snippet.Div == "" {
		t.Error("Expected non-empty Div with empty data")
	}
	if snippet.Script == "" {
		t.Error("Expected non-empty Script with empty data")
	}
}

func TestGenerateSolarActivitySnippetConsistency(t *testing.T) {
	generator := NewChartGenerator("/test")
	
	data := &models.PropagationData{
		SolarData: models.SolarData{
			SolarFluxIndex: 150.0,
			SunspotNumber:  100,
		},
		GeomagData: models.GeomagData{
			KIndex: 3.0,
		},
	}
	
	// Generate snippet twice
	snippet1, err1 := generator.generateSolarActivitySnippet(data)
	snippet2, err2 := generator.generateSolarActivitySnippet(data)
	
	if err1 != nil {
		t.Fatalf("First generation failed: %v", err1)
	}
	if err2 != nil {
		t.Fatalf("Second generation failed: %v", err2)
	}
	
	// Should generate identical snippets
	if snippet1.ID != snippet2.ID {
		t.Errorf("Inconsistent ID: %s != %s", snippet1.ID, snippet2.ID)
	}
	if snippet1.Title != snippet2.Title {
		t.Errorf("Inconsistent Title: %s != %s", snippet1.Title, snippet2.Title)
	}
	if snippet1.Div != snippet2.Div {
		t.Errorf("Inconsistent Div")
	}
	if snippet1.Script != snippet2.Script {
		t.Errorf("Inconsistent Script")
	}
}

func TestGenerateSolarActivitySnippetExtremeValues(t *testing.T) {
	generator := NewChartGenerator("/test")
	
	extremeCases := []struct {
		name      string
		solarFlux float64
		sunspots  int
		kIndex    float64
	}{
		{"Maximum realistic", 500.0, 300, 9.0},
		{"Very low", 60.0, 0, 0.0},
		{"Negative values", -10.0, -5, -1.0},
		{"Very large", 1000.0, 1000, 20.0},
	}
	
	for _, tc := range extremeCases {
		t.Run(tc.name, func(t *testing.T) {
			data := &models.PropagationData{
				SolarData: models.SolarData{
					SolarFluxIndex: tc.solarFlux,
					SunspotNumber:  tc.sunspots,
				},
				GeomagData: models.GeomagData{
					KIndex: tc.kIndex,
				},
			}
			
			snippet, err := generator.generateSolarActivitySnippet(data)
			if err != nil {
				t.Fatalf("generateSolarActivitySnippet failed for %s: %v", tc.name, err)
			}
			
			// Should still generate valid snippet
			if snippet.ID == "" {
				t.Errorf("Expected non-empty ID for %s", tc.name)
			}
			if snippet.Script == "" {
				t.Errorf("Expected non-empty Script for %s", tc.name)
			}
			
			// Verify script is valid JavaScript (basic check)
			if !strings.Contains(snippet.Script, "echarts") {
				t.Errorf("Script should contain echarts for %s", tc.name)
			}
		})
	}
}
