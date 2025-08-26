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
