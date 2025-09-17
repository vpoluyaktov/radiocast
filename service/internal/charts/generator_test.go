package charts

import (
	"testing"
	"time"

	"radiocast/internal/models"
)

func TestNewChartGenerator(t *testing.T) {
	outputDir := "/test/output"
	generator := NewChartGenerator(outputDir)
	
	if generator == nil {
		t.Fatal("NewChartGenerator returned nil")
	}
	
	if generator.outputDir != outputDir {
		t.Errorf("Expected outputDir %s, got %s", outputDir, generator.outputDir)
	}
}

func TestGenerateEChartsSnippetsWithSources(t *testing.T) {
	generator := NewChartGenerator("/test")
	
	// Create test data
	testTime := time.Date(2025, 9, 17, 12, 0, 0, 0, time.UTC)
	data := &models.PropagationData{
		Timestamp: testTime,
		SolarData: models.SolarData{
			SolarFluxIndex: 150.5,
			SunspotNumber:  120,
		},
		GeomagData: models.GeomagData{
			KIndex: 3.5,
		},
		BandData: models.BandData{
			Band80m: models.BandCondition{Day: "Good", Night: "Excellent"},
			Band40m: models.BandCondition{Day: "Fair", Night: "Good"},
			Band20m: models.BandCondition{Day: "Excellent", Night: "Fair"},
		},
		Forecast: models.ForecastData{
			Today: models.DayForecast{
				Date:           testTime,
				KIndexForecast: "3-4",
				SolarActivity:  "Moderate",
			},
		},
	}
	
	sourceData := &models.SourceData{
		NOAAKIndex: []models.NOAAKIndexResponse{
			{TimeTag: "2025-09-17T12:00:00", KpIndex: 3, EstimatedKp: 3.33},
			{TimeTag: "2025-09-17T09:00:00", KpIndex: 2, EstimatedKp: 2.67},
		},
		NOAASolar: []models.NOAASolarResponse{
			{TimeTag: "2025-08", SunspotNumber: 133.5, SolarFlux: 150.0},
		},
	}
	
	snippets, err := generator.GenerateEChartsSnippetsWithSources(data, sourceData)
	if err != nil {
		t.Fatalf("GenerateEChartsSnippetsWithSources failed: %v", err)
	}
	
	// Should generate multiple chart snippets
	if len(snippets) == 0 {
		t.Error("Expected at least one chart snippet, got none")
	}
	
	// Verify each snippet has required fields
	for i, snippet := range snippets {
		if snippet.ID == "" {
			t.Errorf("Snippet %d has empty ID", i)
		}
		if snippet.Title == "" {
			t.Errorf("Snippet %d has empty Title", i)
		}
		if snippet.Div == "" {
			t.Errorf("Snippet %d has empty Div", i)
		}
		if snippet.Script == "" {
			t.Errorf("Snippet %d has empty Script", i)
		}
		if snippet.HTML == "" {
			t.Errorf("Snippet %d has empty HTML", i)
		}
		
		t.Logf("Generated snippet %d: ID=%s, Title=%s", i, snippet.ID, snippet.Title)
	}
}

func TestGenerateEChartsSnippetsWithNilData(t *testing.T) {
	generator := NewChartGenerator("/test")
	
	// Test with nil data - function should be resilient and return empty slice
	snippets, err := generator.GenerateEChartsSnippetsWithSources(nil, nil)
	if err != nil {
		t.Errorf("Expected no error with nil data, got: %v", err)
	}
	if len(snippets) != 0 {
		t.Errorf("Expected no snippets with nil data, got %d", len(snippets))
	}
}

func TestGenerateEChartsSnippetsWithEmptyData(t *testing.T) {
	generator := NewChartGenerator("/test")
	
	// Test with empty data
	data := &models.PropagationData{}
	sourceData := &models.SourceData{}
	
	snippets, err := generator.GenerateEChartsSnippetsWithSources(data, sourceData)
	if err != nil {
		t.Fatalf("GenerateEChartsSnippetsWithSources failed with empty data: %v", err)
	}
	
	// Should still generate snippets, even with empty data
	if len(snippets) == 0 {
		t.Error("Expected at least one chart snippet with empty data, got none")
	}
	
	// Verify snippets are valid
	for i, snippet := range snippets {
		if snippet.ID == "" {
			t.Errorf("Snippet %d has empty ID", i)
		}
		if snippet.Div == "" {
			t.Errorf("Snippet %d has empty Div", i)
		}
		if snippet.Script == "" {
			t.Errorf("Snippet %d has empty Script", i)
		}
	}
}

func TestGenerateEChartsSnippetsWithPartialData(t *testing.T) {
	generator := NewChartGenerator("/test")
	
	// Test with partial data
	data := &models.PropagationData{
		SolarData: models.SolarData{
			SolarFluxIndex: 100.0,
			SunspotNumber:  50,
		},
		GeomagData: models.GeomagData{
			KIndex: 2.0,
		},
		// Missing BandData and Forecast
	}
	
	sourceData := &models.SourceData{
		NOAAKIndex: []models.NOAAKIndexResponse{
			{TimeTag: "2025-09-17T12:00:00", KpIndex: 2, EstimatedKp: 2.0},
		},
		// Missing other source data
	}
	
	snippets, err := generator.GenerateEChartsSnippetsWithSources(data, sourceData)
	if err != nil {
		t.Fatalf("GenerateEChartsSnippetsWithSources failed with partial data: %v", err)
	}
	
	// Should generate some snippets
	if len(snippets) == 0 {
		t.Error("Expected at least one chart snippet with partial data, got none")
	}
	
	// Verify basic structure
	for _, snippet := range snippets {
		if snippet.ID == "" {
			t.Error("Generated snippet has empty ID")
		}
		if snippet.Div == "" {
			t.Error("Generated snippet has empty Div")
		}
		if snippet.Script == "" {
			t.Error("Generated snippet has empty Script")
		}
	}
}

func TestChartGeneratorOutputDir(t *testing.T) {
	tests := []string{
		"/test/output",
		"./local/output",
		"",
		"/very/long/path/to/output/directory",
	}
	
	for _, outputDir := range tests {
		generator := NewChartGenerator(outputDir)
		if generator.outputDir != outputDir {
			t.Errorf("Expected outputDir %s, got %s", outputDir, generator.outputDir)
		}
	}
}

func TestGenerateEChartsSnippetsConsistency(t *testing.T) {
	generator := NewChartGenerator("/test")
	
	// Create consistent test data
	testTime := time.Date(2025, 9, 17, 12, 0, 0, 0, time.UTC)
	data := &models.PropagationData{
		Timestamp: testTime,
		SolarData: models.SolarData{
			SolarFluxIndex: 150.0,
			SunspotNumber:  100,
		},
		GeomagData: models.GeomagData{
			KIndex: 3.0,
		},
	}
	
	sourceData := &models.SourceData{
		NOAAKIndex: []models.NOAAKIndexResponse{
			{TimeTag: "2025-09-17T12:00:00", KpIndex: 3, EstimatedKp: 3.0},
		},
	}
	
	// Generate snippets twice
	snippets1, err1 := generator.GenerateEChartsSnippetsWithSources(data, sourceData)
	snippets2, err2 := generator.GenerateEChartsSnippetsWithSources(data, sourceData)
	
	if err1 != nil {
		t.Fatalf("First generation failed: %v", err1)
	}
	if err2 != nil {
		t.Fatalf("Second generation failed: %v", err2)
	}
	
	// Should generate same number of snippets
	if len(snippets1) != len(snippets2) {
		t.Errorf("Inconsistent snippet count: first=%d, second=%d", len(snippets1), len(snippets2))
	}
	
	// Compare snippet IDs and titles
	for i := 0; i < len(snippets1) && i < len(snippets2); i++ {
		if snippets1[i].ID != snippets2[i].ID {
			t.Errorf("Snippet %d ID mismatch: %s != %s", i, snippets1[i].ID, snippets2[i].ID)
		}
		if snippets1[i].Title != snippets2[i].Title {
			t.Errorf("Snippet %d Title mismatch: %s != %s", i, snippets1[i].Title, snippets2[i].Title)
		}
	}
}
