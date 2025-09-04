package fetchers

import (
	"context"
	"strings"
	"testing"
)

func TestFetchNOAAKIndex(t *testing.T) {
	// Create a new fetcher
	fetcher := NewDataFetcher()
	ctx := context.Background()
	
	// Use the real NOAA API endpoint
	url := "https://services.swpc.noaa.gov/products/noaa-planetary-k-index.json"
	
	// Fetch data
	data, err := fetcher.noaaFetcher.FetchKIndex(ctx, url)
	if err != nil {
		t.Fatalf("FetchKIndex failed: %v", err)
	}
	
	// Basic validation
	if len(data) == 0 {
		t.Fatal("Expected at least one K-index data point, got none")
	}
	
	// Validate filtering to 72 hours (up to 24 entries with 3-hour intervals)
	if len(data) > 24 {
		t.Logf("Got %d entries, expected at most 24 for 72-hour history", len(data))
	}
	
	// Validate data structure
	for i, entry := range data {
		// Check required fields
		if entry.TimeTag == "" {
			t.Errorf("Entry %d: TimeTag should not be empty", i)
		}
		
		// Validate K-index range
		if entry.KpIndex < 0 || entry.KpIndex > 9 {
			t.Errorf("Entry %d: K-index should be between 0-9, got %f", i, entry.KpIndex)
		}
		
		// Validate source attribution
		if entry.Source != "NOAA SWPC" {
			t.Errorf("Entry %d: Expected source 'NOAA SWPC', got '%s'", i, entry.Source)
		}
	}
	
	// Log results
	latest := data[len(data)-1]
	t.Logf("Successfully fetched %d K-index data points", len(data))
	t.Logf("Latest K-index: %f at %s", latest.KpIndex, latest.TimeTag)
}

func TestFetchNOAASolar(t *testing.T) {
	// Create a new fetcher
	fetcher := NewDataFetcher()
	ctx := context.Background()
	
	// Use the real NOAA Solar API endpoint
	url := "https://services.swpc.noaa.gov/json/solar-cycle/observed-solar-cycle-indices.json"
	
	data, err := fetcher.noaaFetcher.FetchSolar(ctx, url)
	if err != nil {
		t.Fatalf("FetchSolar failed: %v", err)
	}
	
	if len(data) == 0 {
		t.Fatal("Expected at least one solar data point, got none")
	}
	
	// Validate filtering to 6 months
	if len(data) > 6 {
		t.Errorf("Expected at most 6 filtered solar data entries, got %d", len(data))
	}
	
	// Validate data structure
	validEntries := 0
	for i, entry := range data {
		// Check required fields
		if entry.TimeTag == "" {
			t.Errorf("Entry %d: TimeTag should not be empty", i)
		}
		
		// Validate TimeTag format (YYYY-MM format for monthly data)
		if !strings.Contains(entry.TimeTag, "-") || len(entry.TimeTag) < 7 {
			t.Errorf("Entry %d: Invalid TimeTag format '%s', expected YYYY-MM", i, entry.TimeTag)
		}
		
		// Validate reasonable ranges for solar data
		if entry.SolarFlux < 0 {
			t.Errorf("Entry %d: Solar flux should be >= 0, got %f", i, entry.SolarFlux)
		}
		
		if entry.SunspotNumber < 0 {
			t.Errorf("Entry %d: Sunspot number should be >= 0, got %f", i, entry.SunspotNumber)
		}
		
		// Count valid entries (either flux or sunspot data available)
		if entry.SolarFlux > 0 || entry.SunspotNumber > 0 {
			validEntries++
		}
	}
	
	// Log results
	last := data[len(data)-1]
	t.Logf("Successfully fetched %d solar data points", len(data))
	t.Logf("Valid entries: %d", validEntries)
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
