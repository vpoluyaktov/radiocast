package fetchers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"radiocast/internal/models"

	"github.com/go-resty/resty/v2"
)

// Configuration constants for data filtering
const (
	// KIndexHistoryHours defines how many hours of K-index data to keep
	KIndexHistoryHours = 24
	// KIndexSampleIntervalHours defines the sampling interval for K-index data
	KIndexSampleIntervalHours = 1
	// SolarDataHistoryDays defines how many days of solar data to keep
	SolarDataHistoryDays = 7
)

// NOAAFetcher handles fetching data from NOAA APIs
type NOAAFetcher struct {
	client *resty.Client
}

// NewNOAAFetcher creates a new NOAA fetcher instance
func NewNOAAFetcher(client *resty.Client) *NOAAFetcher {
	return &NOAAFetcher{
		client: client,
	}
}

// FetchKIndex fetches K-index data from NOAA
func (f *NOAAFetcher) FetchKIndex(ctx context.Context, url string) ([]models.NOAAKIndexResponse, error) {
	resp, err := f.client.R().
		SetContext(ctx).
		SetHeader("Accept", "application/json").
		Get(url)
	
	if err != nil {
		return nil, fmt.Errorf("failed to fetch NOAA K-index: %w", err)
	}
	
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("NOAA K-index API returned status %d", resp.StatusCode())
	}
	
	// NOAA returns array of objects format: [{"time_tag":"...","kp_index":0,"estimated_kp":0.33,"kp":"0P"}]
	var rawData []models.NOAAKIndexAPIResponse
	if err := json.Unmarshal(resp.Body(), &rawData); err != nil {
		return nil, fmt.Errorf("failed to parse NOAA K-index response: %w", err)
	}

	if len(rawData) == 0 {
		return nil, fmt.Errorf("NOAA K-index response has no data")
	}

	// Convert to our internal format
	var data []models.NOAAKIndexResponse
	for _, item := range rawData {
		data = append(data, models.NOAAKIndexResponse{
			TimeTag:     item.TimeTag,
			KpIndex:     item.EstimatedKp, // Use estimated_kp instead of kp_index
			EstimatedKp: item.EstimatedKp,
			Source:      "NOAA SWPC",
		})
	}
	
	// Filter to recent data only
	return f.filterKIndexRecent(data), nil
}

// FetchSolar fetches solar data from NOAA
func (f *NOAAFetcher) FetchSolar(ctx context.Context, url string) ([]models.NOAASolarResponse, error) {
	resp, err := f.client.R().
		SetContext(ctx).
		SetHeader("Accept", "application/json").
		Get(url)
	
	if err != nil {
		return nil, fmt.Errorf("failed to fetch NOAA solar data: %w", err)
	}
	
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("NOAA solar API returned status %d", resp.StatusCode())
	}
	
	// NOAA now returns array of objects format: [{"time-tag":"1749-01","ssn":96.7,"f10.7":-1.0}]
	var rawData []models.NOAASolarAPIResponse
	if err := json.Unmarshal(resp.Body(), &rawData); err != nil {
		return nil, fmt.Errorf("failed to parse NOAA solar response: %w", err)
	}
	
	if len(rawData) == 0 {
		return nil, fmt.Errorf("NOAA solar response has no data")
	}
	
	// Convert to our internal format, prioritize entries with valid data
	var data []models.NOAASolarResponse
	for _, item := range rawData {
		// Skip entries where both critical values are invalid
		if item.F107 < 0 && item.SSN < 0 {
			continue
		}
		
		// Only include entries with at least one valid value
		solarFlux := item.F107
		sunspotNumber := item.SSN
		
		// Skip entries with invalid solar flux but keep if SSN is valid
		if solarFlux < 0 && sunspotNumber >= 0 {
			// Use a reasonable default for solar flux when missing
			solarFlux = 100.0 // Typical quiet sun value
		} else if solarFlux < 0 {
			continue // Skip if solar flux is invalid and no valid SSN
		}
		
		if sunspotNumber < 0 {
			sunspotNumber = 0 // Use 0 for invalid sunspot numbers
		}
		
		data = append(data, models.NOAASolarResponse{
			TimeTag:           item.TimeTag,
			SolarFlux:         solarFlux,
			SunspotNumber:     sunspotNumber,
			SolarFluxAdjusted: solarFlux, // Use same value
			Source:            "NOAA SWPC",
		})
	}
	
	// Filter to recent data only
	return f.filterSolarRecent(data), nil
}

// filterKIndexRecent filters K-index data to last 24 hours with 3-hour intervals
func (f *NOAAFetcher) filterKIndexRecent(kIndexData []models.NOAAKIndexResponse) []models.NOAAKIndexResponse {
	if len(kIndexData) == 0 {
		return kIndexData
	}
	
	var filtered []models.NOAAKIndexResponse
	now := time.Now()
	cutoffTime := now.Add(-time.Duration(KIndexHistoryHours) * time.Hour)
	
	// Sample every KIndexSampleIntervalHours hours
	lastSampleTime := time.Time{}
	
	for _, entry := range kIndexData {
		if entryTime, err := time.Parse("2006-01-02T15:04:05", entry.TimeTag); err == nil {
			if entryTime.After(cutoffTime) {
				// Include if it's the first entry or interval hours since last sample
				if lastSampleTime.IsZero() || entryTime.Sub(lastSampleTime) >= time.Duration(KIndexSampleIntervalHours)*time.Hour {
					filtered = append(filtered, entry)
					lastSampleTime = entryTime
				}
			}
		}
	}
	
	// If no entries found, take the last 8 entries (roughly last day)
	if len(filtered) == 0 && len(kIndexData) > 0 {
		start := len(kIndexData) - 8
		if start < 0 {
			start = 0
		}
		filtered = kIndexData[start:]
	}
	
	return filtered
}

// filterSolarRecent filters solar data to keep only recent entries
func (f *NOAAFetcher) filterSolarRecent(solarData []models.NOAASolarResponse) []models.NOAASolarResponse {
	if len(solarData) == 0 {
		return solarData
	}
	
	// Take last SolarDataHistoryDays entries to avoid token limits
	if len(solarData) > SolarDataHistoryDays {
		return solarData[len(solarData)-SolarDataHistoryDays:]
	}
	
	return solarData
}
