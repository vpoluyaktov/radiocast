package fetchers

import (
	"context"
	"encoding/json"
	"fmt"

	"radiocast/internal/models"

	"github.com/go-resty/resty/v2"
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
	type NOAAKIndexAPIResponse struct {
		TimeTag     string  `json:"time_tag"`
		KpIndex     int     `json:"kp_index"`
		EstimatedKp float64 `json:"estimated_kp"`
		Kp          string  `json:"kp"`
	}

	var rawData []NOAAKIndexAPIResponse
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
	
	return data, nil
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
	type NOAASolarAPIResponse struct {
		TimeTag           string  `json:"time-tag"`
		SSN               float64 `json:"ssn"`
		SmoothedSSN       float64 `json:"smoothed_ssn"`
		ObservedSWPCSSN   float64 `json:"observed_swpc_ssn"`
		SmoothedSWPCSSN   float64 `json:"smoothed_swpc_ssn"`
		F107              float64 `json:"f10.7"`
		SmoothedF107      float64 `json:"smoothed_f10.7"`
	}
	
	var rawData []NOAASolarAPIResponse
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
	
	return data, nil
}
