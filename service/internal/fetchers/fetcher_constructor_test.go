package fetchers

import (
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
	
	if fetcher.noaaFetcher == nil {
		t.Error("NOAA fetcher not initialized")
	}
	
	if fetcher.n0nbhFetcher == nil {
		t.Error("N0NBH fetcher not initialized")
	}
	
	if fetcher.sidcFetcher == nil {
		t.Error("SIDC fetcher not initialized")
	}
	
	if fetcher.normalizer == nil {
		t.Error("Data normalizer not initialized")
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
	
	forecast := fetcher.normalizer.GenerateBasicForecast(data)
	
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
