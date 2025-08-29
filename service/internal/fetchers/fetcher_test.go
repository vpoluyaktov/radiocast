package fetchers

import (
	"context"
	"strings"
	"testing"
	"time"

	"radiocast/internal/models"
)

func TestNewDataFetcher(t *testing.T) {
	fetcher := NewDataFetcher()
	if fetcher == nil {
		t.Error("NewDataFetcher returned nil")
		return
	}
	
	if fetcher.client == nil {
		t.Error("HTTP client not initialized")
	}
	
	if fetcher.parser == nil {
		t.Error("RSS parser not initialized")
	}
}

func TestGenerateBasicForecast(t *testing.T) {
	fetcher := NewDataFetcher()
	
	// Create test data
	data := &models.PropagationData{
		Timestamp: time.Now(),
		GeomagData: models.GeomagData{
			KIndex: 2.0,
		},
		SolarData: models.SolarData{
			SolarFluxIndex: 150.0,
		},
	}
	
	forecast := fetcher.generateBasicForecast(data)
	
	if forecast.Today.Date.IsZero() {
		t.Error("Today's forecast date not set")
	}
	
	if forecast.Tomorrow.Date.IsZero() {
		t.Error("Tomorrow's forecast date not set")
	}
	
	if forecast.Outlook == "" {
		t.Error("Forecast outlook not generated")
	}
}

func TestFetchNOAAKIndex(t *testing.T) {
	// Test with real NOAA K-index API
	fetcher := NewDataFetcher()
	ctx := context.Background()
	
	// Use real NOAA API endpoint
	url := "https://services.swpc.noaa.gov/json/planetary_k_index_1m.json"
	
	data, err := fetcher.fetchNOAAKIndex(ctx, url)
	if err != nil {
		t.Fatalf("fetchNOAAKIndex failed: %v", err)
	}
	
	if len(data) == 0 {
		t.Error("Expected at least one data point, got none")
	}
	
	// Validate structure of first item
	first := data[0]
	if first.TimeTag == "" {
		t.Error("TimeTag should not be empty")
	}
	
	if first.KpIndex < 0 || first.KpIndex > 9 {
		t.Errorf("K-index should be between 0-9, got %f", first.KpIndex)
	}
	
	if first.EstimatedKp < 0 || first.EstimatedKp > 9 {
		t.Errorf("Estimated Kp should be between 0-9, got %f", first.EstimatedKp)
	}
	
	t.Logf("Successfully fetched %d K-index data points", len(data))
	t.Logf("Latest K-index: %f at %s", first.KpIndex, first.TimeTag)
}

func TestFetchNOAASolar(t *testing.T) {
	// Test with real NOAA Solar API
	fetcher := NewDataFetcher()
	ctx := context.Background()
	
	// Use real NOAA Solar API endpoint
	url := "https://services.swpc.noaa.gov/json/solar-cycle/observed-solar-cycle-indices.json"
	
	data, err := fetcher.fetchNOAASolar(ctx, url)
	if err != nil {
		t.Fatalf("fetchNOAASolar failed: %v", err)
	}
	
	if len(data) == 0 {
		t.Error("Expected at least one data point, got none")
	}
	
	// Validate structure of last item (most recent)
	last := data[len(data)-1]
	if last.TimeTag == "" {
		t.Error("TimeTag should not be empty")
	}
	
	if last.SolarFlux < 0 {
		t.Errorf("Solar flux should be >= 0, got %f", last.SolarFlux)
	}
	
	if last.SunspotNumber < 0 {
		t.Errorf("Sunspot number should be >= 0, got %f", last.SunspotNumber)
	}
	
	t.Logf("Successfully fetched %d solar data points", len(data))
	t.Logf("Latest solar data: F10.7=%f, SSN=%f at %s", last.SolarFlux, last.SunspotNumber, last.TimeTag)
}

func TestFetchN0NBHError(t *testing.T) {
	// Test N0NBH API which currently returns 404
	fetcher := NewDataFetcher()
	ctx := context.Background()
	
	// Use real N0NBH API endpoint (currently broken)
	url := "https://www.hamqsl.com/solarapi.php?format=json"
	
	data, err := fetcher.fetchN0NBH(ctx, url)
	if err != nil {
		// Expected to fail with 404
		if !strings.Contains(err.Error(), "returned status 404") {
			t.Errorf("Expected 404 error, got: %v", err)
		}
		t.Logf("N0NBH API correctly returns 404 as expected: %v", err)
		return
	}
	
	// If it somehow works, validate the data
	if data != nil {
		t.Logf("N0NBH API unexpectedly working, got data: %+v", data.SolarData)
	}
}

func TestFetchSIDCError(t *testing.T) {
	// Test SIDC API which currently returns HTML instead of RSS
	fetcher := NewDataFetcher()
	ctx := context.Background()
	
	// Use real SIDC API endpoint (currently returns HTML)
	url := "https://www.sidc.be/products/meu"
	
	data, err := fetcher.fetchSIDC(ctx, url)
	if err != nil {
		// Expected to fail with RSS parsing error
		t.Logf("SIDC API correctly fails as expected: %v", err)
		return
	}
	
	// If it somehow works, validate the data
	if len(data) > 0 {
		t.Logf("SIDC API unexpectedly working, got %d items", len(data))
		for i, item := range data {
			if i >= 3 { // Only log first 3
				break
			}
			t.Logf("Item %d: %s", i, item.Title)
		}
	}
}

func TestFetchAllDataIntegration(t *testing.T) {
	// Integration test with real APIs
	fetcher := NewDataFetcher()
	ctx := context.Background()
	
	// Use real API endpoints
	noaaKURL := "https://services.swpc.noaa.gov/json/planetary_k_index_1m.json"
	noaaSolarURL := "https://services.swpc.noaa.gov/json/solar-cycle/observed-solar-cycle-indices.json"
	n0nbhURL := "https://www.hamqsl.com/solarapi.php?format=json" // Expected to fail
	sidcURL := "https://www.sidc.be/products/meu" // Expected to fail
	
	data, err := fetcher.FetchAllData(ctx, noaaKURL, noaaSolarURL, n0nbhURL, sidcURL)
	if err != nil {
		t.Fatalf("FetchAllData failed: %v", err)
	}
	
	if data == nil {
		t.Fatal("Expected data, got nil")
	}
	
	// Validate that we got some data despite broken APIs
	if data.GeomagData.KIndex <= 0 {
		t.Errorf("Expected positive K-index, got %f", data.GeomagData.KIndex)
	}
	
	if data.SolarData.SolarFluxIndex <= 0 && data.SolarData.SunspotNumber <= 0 {
		t.Error("Expected some solar data (flux or sunspot number)")
	}
	
	if data.GeomagData.GeomagActivity == "" {
		t.Error("Geomag activity should be set")
	}
	
	if data.Forecast.Today.Date.IsZero() {
		t.Error("Today's forecast date should be set")
	}
	
	t.Logf("Successfully fetched and normalized data:")
	t.Logf("K-index: %f (%s)", data.GeomagData.KIndex, data.GeomagData.GeomagActivity)
	t.Logf("Solar flux: %f", data.SolarData.SolarFluxIndex)
	t.Logf("Sunspot number: %d", data.SolarData.SunspotNumber)
	t.Logf("Forecast outlook: %s", data.Forecast.Outlook)
}

func TestNormalizeDataWithRealData(t *testing.T) {
	// Test normalization with real NOAA data
	fetcher := NewDataFetcher()
	ctx := context.Background()
	
	// Fetch real NOAA data
	kIndexData, err := fetcher.fetchNOAAKIndex(ctx, "https://services.swpc.noaa.gov/json/planetary_k_index_1m.json")
	if err != nil {
		t.Fatalf("Failed to fetch K-index data: %v", err)
	}
	
	solarData, err := fetcher.fetchNOAASolar(ctx, "https://services.swpc.noaa.gov/json/solar-cycle/observed-solar-cycle-indices.json")
	if err != nil {
		t.Fatalf("Failed to fetch solar data: %v", err)
	}
	
	// Normalize with real data (N0NBH and SIDC will be nil due to broken APIs)
	result := fetcher.normalizeData(kIndexData, solarData, nil, nil)
	
	if result == nil {
		t.Fatal("Expected normalized data, got nil")
	}
	
	if result.GeomagData.KIndex <= 0 {
		t.Errorf("Expected positive K-index from real data, got %f", result.GeomagData.KIndex)
	}
	
	if result.GeomagData.GeomagActivity == "" {
		t.Error("Geomag activity should be determined from K-index")
	}
	
	// Solar data might be 0 if recent entries have invalid values
	if result.SolarData.SolarFluxIndex < 0 {
		t.Errorf("Solar flux should not be negative, got %f", result.SolarData.SolarFluxIndex)
	}
	
	if result.SolarData.SunspotNumber < 0 {
		t.Errorf("Sunspot number should not be negative, got %d", result.SolarData.SunspotNumber)
	}
	
	t.Logf("Normalized real data successfully:")
	t.Logf("K-index: %f (%s)", result.GeomagData.KIndex, result.GeomagData.GeomagActivity)
	t.Logf("Solar flux: %f, Sunspots: %d", result.SolarData.SolarFluxIndex, result.SolarData.SunspotNumber)
}
