package fetchers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	// Mock NOAA K-index response
	mockResponse := [][]interface{}{
		{"time_tag", "Kp", "a_running", "station_count"},
		{"2025-08-29 00:00:00.000", "2.33", "9", "8"},
		{"2025-08-29 03:00:00.000", "1.67", "6", "8"},
		{"2025-08-29 06:00:00.000", "2.00", "7", "8"},
	}
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()
	
	fetcher := NewDataFetcher()
	ctx := context.Background()
	
	data, err := fetcher.fetchNOAAKIndex(ctx, server.URL)
	if err != nil {
		t.Fatalf("fetchNOAAKIndex failed: %v", err)
	}
	
	if len(data) != 3 {
		t.Errorf("Expected 3 data points, got %d", len(data))
	}
	
	if data[0].KpIndex != 2.33 {
		t.Errorf("Expected K-index 2.33, got %f", data[0].KpIndex)
	}
	
	if data[0].TimeTag != "2025-08-29 00:00:00.000" {
		t.Errorf("Expected time tag '2025-08-29 00:00:00.000', got '%s'", data[0].TimeTag)
	}
}

func TestFetchNOAAKIndexError(t *testing.T) {
	// Test server returning error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	
	fetcher := NewDataFetcher()
	ctx := context.Background()
	
	_, err := fetcher.fetchNOAAKIndex(ctx, server.URL)
	if err == nil {
		t.Error("Expected error for 500 status, got nil")
	}
	
	if !strings.Contains(err.Error(), "returned status 500") {
		t.Errorf("Expected status error message, got: %v", err)
	}
}

func TestFetchNOAAKIndexInvalidJSON(t *testing.T) {
	// Test invalid JSON response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()
	
	fetcher := NewDataFetcher()
	ctx := context.Background()
	
	_, err := fetcher.fetchNOAAKIndex(ctx, server.URL)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
	
	if !strings.Contains(err.Error(), "failed to parse") {
		t.Errorf("Expected parse error message, got: %v", err)
	}
}

func TestFetchN0NBH(t *testing.T) {
	// Mock N0NBH XML response
	mockXML := `<?xml version="1.0" encoding="UTF-8" ?>
<solar>
	<solardata>
		<source url="http://www.hamqsl.com/solar.html">N0NBH</source>
		<updated> 29 Aug 2025 1509 GMT</updated>
		<solarflux>232</solarflux>
		<aindex> 7</aindex>
		<kindex> 1</kindex>
		<sunspots>213</sunspots>
		<heliumline>152.4</heliumline>
		<protonflux>712</protonflux>
		<electonflux>78600</electonflux>
		<aurora> 1</aurora>
		<normalization>1.99</normalization>
		<latdegree>67.5</latdegree>
		<solarwind>378.5</solarwind>
		<magneticfield>  0.2</magneticfield>
	</solardata>
	<calculatedconditions>
		<band name="80m-40m" time="Day" day="Good" night="Good"></band>
		<band name="30m-20m" time="Day" day="Fair" night="Poor"></band>
	</calculatedconditions>
</solar>`
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(mockXML))
	}))
	defer server.Close()
	
	fetcher := NewDataFetcher()
	ctx := context.Background()
	
	data, err := fetcher.fetchN0NBH(ctx, server.URL)
	if err != nil {
		t.Fatalf("fetchN0NBH failed: %v", err)
	}
	
	if data.SolarData.SolarFlux != "232" {
		t.Errorf("Expected solar flux '232', got '%s'", data.SolarData.SolarFlux)
	}
	
	if data.SolarData.SunSpots != "213" {
		t.Errorf("Expected sunspot number '213', got '%s'", data.SolarData.SunSpots)
	}
	
	if strings.TrimSpace(data.SolarData.KIndex) != "1" {
		t.Errorf("Expected K-index '1', got '%s'", strings.TrimSpace(data.SolarData.KIndex))
	}
	
	if len(data.Calculatedconditions.Band) != 2 {
		t.Errorf("Expected 2 band conditions, got %d", len(data.Calculatedconditions.Band))
	}
}

func TestNormalizeData(t *testing.T) {
	fetcher := NewDataFetcher()
	
	// Create mock data
	kIndexData := []models.NOAAKIndexResponse{
		{TimeTag: "2025-08-29 00:00:00.000", KpIndex: 2.5, EstimatedKp: 2.5},
		{TimeTag: "2025-08-29 03:00:00.000", KpIndex: 1.8, EstimatedKp: 1.8},
	}
	
	// NOAA solar data provides sunspot numbers but no solar flux
	solarData := []models.NOAASolarResponse{
		{TimeTag: "2025-08-29 06:00:00.000", SolarFlux: 0.0, SunspotNumber: 45.0},
	}
	
	n0nbhData := &models.N0NBHResponse{
		SolarData: struct {
			SolarFlux     string `json:"solarflux"`
			AIndex        string `json:"aindex"`
			KIndex        string `json:"kindex"`
			KIndexNT      string `json:"kindexnt"`
			SunSpots      string `json:"sunspots"`
			HeliumLine    string `json:"heliumline"`
			ProtonFlux    string `json:"protonflux"`
			ElectronFlux  string `json:"electonflux"`
			Aurora        string `json:"aurora"`
			NormalizationTime string `json:"normalization"`
			LatestSWPCReport  string `json:"latestswpcreport"`
		}{
			SolarFlux:     "180",
			AIndex:        "8",
			KIndex:        "2",
			SunSpots:      "45",
			ProtonFlux:    "500",
			Aurora:        "1",
			NormalizationTime: "1.5",
			LatestSWPCReport: "Test report",
		},
	}
	
	result := fetcher.normalizeData(kIndexData, solarData, n0nbhData, nil)
	
	if result.GeomagData.KIndex != 1.8 {
		t.Errorf("Expected latest K-index 1.8, got %f", result.GeomagData.KIndex)
	}
	
	// N0NBH solar flux should override NOAA since it's more recent/reliable
	if result.SolarData.SolarFluxIndex != 180 {
		t.Errorf("Expected solar flux 180, got %f", result.SolarData.SolarFluxIndex)
	}
	
	// Sunspot number comes from NOAA solar data
	if result.SolarData.SunspotNumber != 45 {
		t.Errorf("Expected sunspot number 45, got %d", result.SolarData.SunspotNumber)
	}
	
	if result.GeomagData.GeomagActivity == "" {
		t.Error("Geomag activity not set")
	}
	
	// Check A-index from N0NBH
	if result.GeomagData.AIndex != 8 {
		t.Errorf("Expected A-index 8, got %f", result.GeomagData.AIndex)
	}
}

func TestParseSIDCCSV(t *testing.T) {
	fetcher := NewDataFetcher()
	
	csvData := `# This is a comment
2025 08 28 2025.6575 45.2 3.1 25 1
2025 08 29 2025.6603 47.8 3.2 26 1`
	
	items, err := fetcher.parseSIDCCSV(csvData)
	if err != nil {
		t.Fatalf("parseSIDCCSV failed: %v", err)
	}
	
	if len(items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(items))
	}
	
	if !strings.Contains(items[0].Title, "45.2") {
		t.Errorf("Expected sunspot number 45.2 in title, got: %s", items[0].Title)
	}
	
	if items[0].PublishedParsed == nil {
		t.Error("Published date not parsed")
	}
}
