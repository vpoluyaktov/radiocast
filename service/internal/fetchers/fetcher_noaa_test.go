package fetchers

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestFetchNOAAKIndex(t *testing.T) {
	// Test with real NOAA K-index API
	fetcher := NewDataFetcher()
	ctx := context.Background()
	
	// Use real NOAA API endpoint
	url := "https://services.swpc.noaa.gov/json/planetary_k_index_1m.json"
	
	data, err := fetcher.noaaFetcher.FetchKIndex(ctx, url)
	if err != nil {
		t.Fatalf("fetchNOAAKIndex failed: %v", err)
	}
	
	if len(data) == 0 {
		t.Fatal("Expected at least one data point, got none")
	}
	
	// Validate we have recent data (filtered to last 24 hours with 3-hour intervals)
	// Should have at most 8 entries (24 hours / 3 hour intervals)
	if len(data) == 0 {
		t.Errorf("Expected at least 1 recent K-index entry, got %d", len(data))
	}
	if len(data) > 8 {
		t.Errorf("Expected at most 8 filtered K-index entries, got %d", len(data))
	}
	
	// Validate structure and content of multiple items
	validEntries := 0
	for i, entry := range data {
		if i >= 10 { // Check first 10 entries
			break
		}
		
		if entry.TimeTag == "" {
			t.Errorf("Entry %d: TimeTag should not be empty", i)
			continue
		}
		
		// Validate TimeTag format (should be ISO 8601)
		if _, err := time.Parse("2006-01-02T15:04:05", entry.TimeTag); err != nil {
			t.Errorf("Entry %d: Invalid TimeTag format '%s': %v", i, entry.TimeTag, err)
		}
		
		if entry.KpIndex < 0 || entry.KpIndex > 9 {
			t.Errorf("Entry %d: K-index should be between 0-9, got %f", i, entry.KpIndex)
		}
		
		if entry.EstimatedKp < 0 || entry.EstimatedKp > 9 {
			t.Errorf("Entry %d: Estimated Kp should be between 0-9, got %f", i, entry.EstimatedKp)
		}
		
		// Verify EstimatedKp is being used as primary value (should be > 0 for recent data)
		if entry.EstimatedKp > 0 {
			validEntries++
		}
	}
	
	// Log data quality - during very quiet geomagnetic conditions, EstimatedKp can be 0
	if validEntries == 0 {
		t.Logf("All K-index entries have EstimatedKp = 0 (very quiet geomagnetic conditions)")
	}
	
	// Validate latest entry has reasonable timestamp (within last 24 hours)
	latest := data[len(data)-1]
	if latestTime, err := time.Parse("2006-01-02T15:04:05", latest.TimeTag); err == nil {
		if time.Since(latestTime) > 24*time.Hour {
			t.Errorf("Latest K-index data is too old: %s (over 24 hours ago)", latest.TimeTag)
		}
	}
	
	t.Logf("Successfully fetched %d K-index data points", len(data))
	t.Logf("Valid entries with EstimatedKp > 0: %d", validEntries)
	t.Logf("Latest K-index: %f (EstimatedKp: %f) at %s", latest.KpIndex, latest.EstimatedKp, latest.TimeTag)
}

func TestFetchNOAASolar(t *testing.T) {
	// Test with real NOAA Solar API
	fetcher := NewDataFetcher()
	ctx := context.Background()
	
	// Use real NOAA Solar API endpoint
	url := "https://services.swpc.noaa.gov/json/solar-cycle/observed-solar-cycle-indices.json"
	
	data, err := fetcher.noaaFetcher.FetchSolar(ctx, url)
	if err != nil {
		t.Fatalf("fetchNOAASolar failed: %v", err)
	}
	
	if len(data) == 0 {
		t.Fatal("Expected at least one data point, got none")
	}
	
	// Should have filtered solar data (at most 7 entries as per SolarDataHistoryDays)
	if len(data) > 7 {
		t.Errorf("Expected at most 7 filtered solar data entries, got %d entries", len(data))
	}
	
	// Validate recent entries (all available since we now have filtered data)
	validRecentEntries := 0
	startIdx := 0
	
	for i := startIdx; i < len(data); i++ {
		entry := data[i]
		if entry.TimeTag == "" {
			t.Errorf("Entry %d: TimeTag should not be empty", i)
			continue
		}
		
		// Validate TimeTag format (YYYY-MM format for monthly data)
		if !strings.Contains(entry.TimeTag, "-") || len(entry.TimeTag) < 7 {
			t.Errorf("Entry %d: Invalid TimeTag format '%s', expected YYYY-MM", i, entry.TimeTag)
		}
		
		if entry.SolarFlux < 0 {
			t.Errorf("Entry %d: Solar flux should be >= 0, got %f", i, entry.SolarFlux)
		}
		
		if entry.SunspotNumber < 0 {
			t.Errorf("Entry %d: Sunspot number should be >= 0, got %f", i, entry.SunspotNumber)
		}
		
		// Validate reasonable ranges for solar data
		if entry.SolarFlux > 0 && entry.SolarFlux < 50 {
			t.Errorf("Entry %d: Solar flux %f seems too low (< 50)", i, entry.SolarFlux)
		}
		if entry.SolarFlux > 500 {
			t.Errorf("Entry %d: Solar flux %f seems too high (> 500)", i, entry.SolarFlux)
		}
		
		if entry.SunspotNumber > 500 {
			t.Errorf("Entry %d: Sunspot number %f seems too high (> 500)", i, entry.SunspotNumber)
		}
		
		// Count valid entries (either flux or sunspot data available)
		if entry.SolarFlux > 0 || entry.SunspotNumber > 0 {
			validRecentEntries++
		}
	}
	
	// Ensure we have meaningful recent data
	if validRecentEntries == 0 {
		t.Error("No valid recent solar data found (all zeros)")
	}
	
	// Validate latest entry
	last := data[len(data)-1]
	if last.SolarFlux == 0 && last.SunspotNumber == 0 {
		t.Error("Latest solar entry has no valid data (both flux and sunspot are 0)")
	}
	
	// Check data processing logic - entries with F10.7 = -1 should be handled
	processedEntries := 0
	// Check all available entries since we now have filtered data (max 7)
	for _, entry := range data {
		if entry.SolarFlux == 100.0 { // Default value used for invalid F10.7
			processedEntries++
		}
	}
	
	t.Logf("Successfully fetched %d solar data points", len(data))
	t.Logf("Valid recent entries: %d/10", validRecentEntries)
	t.Logf("Entries with processed default flux: %d", processedEntries)
	t.Logf("Latest solar data: F10.7=%f, SSN=%f at %s", last.SolarFlux, last.SunspotNumber, last.TimeTag)
}

func TestFetchNOAAKIndexInvalidURL(t *testing.T) {
	// Test with invalid URL to ensure proper error handling
	fetcher := NewDataFetcher()
	ctx := context.Background()
	
	_, err := fetcher.noaaFetcher.FetchKIndex(ctx, "https://invalid-url-that-does-not-exist.com/api")
	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
	}
	
	// Accept either network error or parsing error (both are valid failure modes)
	if !strings.Contains(err.Error(), "failed to fetch NOAA K-index") && 
	   !strings.Contains(err.Error(), "failed to parse NOAA K-index response") {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

func TestFetchNOAAKIndexMalformedJSON(t *testing.T) {
	// Test with URL that returns malformed JSON
	fetcher := NewDataFetcher()
	ctx := context.Background()
	
	// Use a URL that returns HTML instead of JSON
	_, err := fetcher.noaaFetcher.FetchKIndex(ctx, "https://www.google.com")
	if err == nil {
		t.Error("Expected error for malformed JSON, got nil")
	}
	
	if !strings.Contains(err.Error(), "failed to parse NOAA K-index response") {
		t.Errorf("Expected JSON parsing error, got: %v", err)
	}
}

func TestFetchNOAASolarInvalidURL(t *testing.T) {
	// Test with invalid URL to ensure proper error handling
	fetcher := NewDataFetcher()
	ctx := context.Background()
	
	_, err := fetcher.noaaFetcher.FetchSolar(ctx, "https://invalid-url-that-does-not-exist.com/api")
	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
	}
	
	// Accept either network error or parsing error (both are valid failure modes)
	if !strings.Contains(err.Error(), "failed to fetch NOAA solar data") && 
	   !strings.Contains(err.Error(), "failed to parse NOAA solar response") {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}
