package fetchers

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestFetchN0NBH(t *testing.T) {
	// Test N0NBH API with working XML endpoint
	fetcher := NewDataFetcher()
	ctx := context.Background()
	
	// The fetcher now uses the working XML endpoint internally
	data, err := fetcher.n0nbhFetcher.Fetch(ctx, "https://www.hamqsl.com/solarapi.php?format=json")
	if err != nil {
		t.Skipf("N0NBH fetch failed (API may be temporarily unavailable): %v", err)
	}
	
	if data == nil {
		t.Fatal("Expected N0NBH data, got nil")
	}
	
	// Validate solar data structure
	if data.SolarData.SolarFlux == "" {
		t.Error("Solar flux should not be empty")
	}
	
	if data.SolarData.SunSpots == "" {
		t.Error("Sunspots should not be empty")
	}
	
	if data.SolarData.KIndex == "" {
		t.Error("K-index should not be empty")
	}
	
	if data.SolarData.AIndex == "" {
		t.Error("A-index should not be empty")
	}
	
	// Validate band conditions
	if len(data.Calculatedconditions.Band) == 0 {
		t.Error("Expected band conditions, got none")
	}
	
	t.Logf("N0NBH API working successfully:")
	t.Logf("Solar Flux: %s, Sunspots: %s, K-index: %s, A-index: %s", 
		data.SolarData.SolarFlux, data.SolarData.SunSpots, data.SolarData.KIndex, data.SolarData.AIndex)
	t.Logf("Band conditions: %d entries", len(data.Calculatedconditions.Band))
}

func TestFetchSIDC(t *testing.T) {
	// Test SIDC API with working CSV endpoint
	fetcher := NewDataFetcher()
	ctx := context.Background()
	
	// The fetcher now uses the working CSV endpoint internally
	data, err := fetcher.sidcFetcher.Fetch(ctx, "https://www.sidc.be/products/meu")
	if err != nil {
		t.Fatalf("SIDC fetch failed: %v", err)
	}
	
	if len(data) == 0 {
		t.Fatal("Expected SIDC data, got none")
	}
	
	// Validate data structure
	for i, item := range data {
		if i >= 5 { // Check first 5 items
			break
		}
		
		if item.Title == "" {
			t.Errorf("Item %d: Title should not be empty", i)
		}
		
		if !strings.Contains(item.Title, "Sunspot Number") {
			t.Errorf("Item %d: Title should contain 'Sunspot Number', got: %s", i, item.Title)
		}
		
		if item.PublishedParsed == nil {
			t.Errorf("Item %d: Published date should be parsed", i)
		}
	}
	
	// Validate recent data (should have entries from recent months)
	latestItem := data[len(data)-1]
	if latestItem.PublishedParsed != nil {
		if time.Since(*latestItem.PublishedParsed) > 365*24*time.Hour {
			t.Errorf("Latest SIDC data is too old: %v", latestItem.PublishedParsed)
		}
	}
	
	t.Logf("SIDC API working successfully: %d items", len(data))
	if len(data) > 0 {
		t.Logf("Latest item: %s", data[len(data)-1].Title)
		if data[len(data)-1].PublishedParsed != nil {
			t.Logf("Latest date: %v", data[len(data)-1].PublishedParsed.Format("2006-01"))
		}
	}
}
