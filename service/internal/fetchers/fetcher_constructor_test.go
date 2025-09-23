package fetchers

import (
	"testing"
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

// Removed TestGenerateBasicForecast - forecast generation moved to LLM
