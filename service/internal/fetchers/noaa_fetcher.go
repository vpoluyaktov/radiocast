package fetchers

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"radiocast/internal/models"

	"github.com/go-resty/resty/v2"
)

// Configuration constants for data filtering
const (
	// KIndexHistoryHours is the number of hours of K-index history to include
	KIndexHistoryHours = 72 // 72 hours = up to 24 entries (3-hour intervals)
	// SolarDataHistoryMonths defines how many months of solar data to keep
	SolarDataHistoryMonths = 6
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

// FetchKIndex fetches K-index data from NOAA for the last 72 hours
func (f *NOAAFetcher) FetchKIndex(ctx context.Context, url string) ([]models.NOAAKIndexResponse, error) {
	// Always use the provided URL - this is critical for tests
	kIndexURL := url
	// Only use default if URL is empty
	if kIndexURL == "" {
		kIndexURL = "https://services.swpc.noaa.gov/products/noaa-planetary-k-index.json"
	}
	
	// Fetch data from the endpoint
	resp, err := f.client.R().
		SetContext(ctx).
		SetHeader("Accept", "application/json").
		Get(kIndexURL)
	
	if err != nil {
		return nil, fmt.Errorf("failed to fetch NOAA K-index: %w", err)
	}
	
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("NOAA K-index API returned status %d", resp.StatusCode())
	}
	
	// NOAA endpoint returns array with header row: [["time_tag","Kp","a_running","station_count"], ["2025-08-27 00:00:00.000","1.33","5","8"], ...]
	var rawData [][]string
	if err := json.Unmarshal(resp.Body(), &rawData); err != nil {
		return nil, fmt.Errorf("failed to parse NOAA K-index response: %w", err)
	}
	
	// Skip the header row (index 0)
	var kIndexData []models.NOAAKIndexResponse
	for i := 1; i < len(rawData); i++ {
		row := rawData[i]
		if len(row) >= 2 { // Need at least time_tag and Kp
			// Parse Kp value
			var kpValue float64
			if kpStr := row[1]; kpStr != "" {
				if kpParsed, err := parseFloat(kpStr); err == nil {
					kpValue = kpParsed
				}
			}
			
			// Convert time format from "2025-08-27 00:00:00.000" to "2025-08-27T00:00:00"
			timeTag := row[0]
			if strings.Contains(timeTag, " ") {
				timeParts := strings.Split(timeTag, " ")
				if len(timeParts) >= 2 {
					// Remove milliseconds if present
					timePart := timeParts[1]
					if idx := strings.Index(timePart, "."); idx != -1 {
						timePart = timePart[:idx]
					}
					timeTag = timeParts[0] + "T" + timePart
				}
			}
			
			// Create response object
			kIndexData = append(kIndexData, models.NOAAKIndexResponse{
				TimeTag:     timeTag,
				KpIndex:     kpValue,
				EstimatedKp: kpValue,
				Source:      "NOAA SWPC",
			})
		}
	}
	
	// Filter to recent data only (last 72 hours - up to 24 entries)
	return f.filterKIndexRecent(kIndexData), nil
}

// FetchSolar fetches solar data from NOAA for the last 6 months
func (f *NOAAFetcher) FetchSolar(ctx context.Context, url string) ([]models.NOAASolarResponse, error) {
	// Use the provided URL or fall back to standard endpoint for solar data
	solarURL := url
	if solarURL == "" {
		solarURL = "https://services.swpc.noaa.gov/json/solar-cycle/observed-solar-cycle-indices.json"
	}
	
	resp, err := f.client.R().
		SetContext(ctx).
		SetHeader("Accept", "application/json").
		Get(solarURL)
	
	if err != nil {
		return nil, fmt.Errorf("failed to fetch NOAA solar data: %w", err)
	}
	
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("NOAA solar API returned status %d", resp.StatusCode())
	}
	
	// Define the expected API response structure
	type NOAASolarAPIResponse struct {
		TimeTag string  `json:"time-tag"`
		SSN     float64 `json:"ssn"`
		F107    float64 `json:"f10.7"`
	}
	
	// NOAA returns array of objects format: [{"time-tag":"1749-01","ssn":96.7,"f10.7":-1.0}]
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
	
	// Filter to recent data only (last 6 months)
	return f.filterSolarRecent(data), nil
}

// filterKIndexRecent filters K-index data to last 72 hours
func (f *NOAAFetcher) filterKIndexRecent(kIndexData []models.NOAAKIndexResponse) []models.NOAAKIndexResponse {
	if len(kIndexData) == 0 {
		return kIndexData
	}
	
	// First, parse all timestamps and sort data by time
	type timeEntry struct {
		time  time.Time
		entry models.NOAAKIndexResponse
	}
	
	var timeEntries []timeEntry
	for _, entry := range kIndexData {
		// Handle multiple timestamp formats
		var entryTime time.Time
		var err error
		
		// Try multiple formats in sequence
		formats := []string{
			"2006-01-02T15:04:05",
			"2006-01-02 15:04",
			time.RFC3339,
			"2006-01-02 15:04:05.000",
			"2006-01-02T15:04:05Z",
		}
		
		for _, format := range formats {
			entryTime, err = time.Parse(format, entry.TimeTag)
			if err == nil {
				break
			}
		}
		
		if err == nil {
			timeEntries = append(timeEntries, timeEntry{time: entryTime, entry: entry})
		}
	}
	
	// Sort by time (oldest first)
	sort.Slice(timeEntries, func(i, j int) bool {
		return timeEntries[i].time.Before(timeEntries[j].time)
	})
	
	// Get the most recent time
	var latestTime time.Time
	if len(timeEntries) > 0 {
		latestTime = timeEntries[len(timeEntries)-1].time
	} else {
		return kIndexData // No valid entries, return original data
	}
	
	// Calculate cutoff time (72 hours before latest)
	cutoffTime := latestTime.Add(-time.Duration(KIndexHistoryHours) * time.Hour)
	
	// Filter to entries within the last 72 hours
	var filtered []models.NOAAKIndexResponse
	for _, entry := range timeEntries {
		if entry.time.After(cutoffTime) || entry.time.Equal(cutoffTime) {
			filtered = append(filtered, entry.entry)
		}
	}
	
	return filtered
}

// filterSolarRecent filters solar data to keep only the last 6 months
func (f *NOAAFetcher) filterSolarRecent(solarData []models.NOAASolarResponse) []models.NOAASolarResponse {
	if len(solarData) == 0 {
		return solarData
	}
	
	// Take last SolarDataHistoryMonths entries (6 months)
	if len(solarData) > SolarDataHistoryMonths {
		return solarData[len(solarData)-SolarDataHistoryMonths:]
	}
	
	return solarData
}

// parseFloat safely parses a string to float64
func parseFloat(s string) (float64, error) {
	if s == "" {
		return 0, fmt.Errorf("empty string")
	}
	return strconv.ParseFloat(s, 64)
}
