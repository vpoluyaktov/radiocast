package fetchers

import (
	"context"
	"encoding/xml"
	"fmt"

	"radiocast/internal/logger"
	"radiocast/internal/models"

	"github.com/go-resty/resty/v2"
)

// Configuration constants for N0NBH data (currently real-time only)
const (
	// N0NBH provides real-time data only, no historical filtering needed
	// These constants are here for consistency and potential future use
	N0NBHDataRetentionHours = 1 // Real-time data, no retention needed
)

// N0NBHFetcher handles fetching data from N0NBH XML API
type N0NBHFetcher struct {
	client *resty.Client
}

// NewN0NBHFetcher creates a new N0NBH fetcher instance
func NewN0NBHFetcher(client *resty.Client) *N0NBHFetcher {
	return &N0NBHFetcher{
		client: client,
	}
}

// Fetch fetches data from N0NBH solar API (XML format)
func (f *N0NBHFetcher) Fetch(ctx context.Context, url string) (*models.N0NBHResponse, error) {
	// Use the working XML endpoint instead of the broken JSON endpoint
	workingURL := "https://www.hamqsl.com/solarxml.php"
	
	resp, err := f.client.R().
		SetContext(ctx).
		SetHeader("Accept", "application/xml").
		Get(workingURL)
	
	if err != nil {
		return nil, fmt.Errorf("failed to fetch N0NBH data: %w", err)
	}
	
	if resp.StatusCode() != 200 {
		bodyLen := len(resp.Body())
		if bodyLen > 200 {
			bodyLen = 200
		}
		logger.Warnf("N0NBH API returned status %d, response: %s", resp.StatusCode(), string(resp.Body()[:bodyLen]))
		return nil, fmt.Errorf("N0NBH API returned status %d", resp.StatusCode())
	}
	
	// Parse XML response
	var xmlData models.N0NBHXMLResponse
	if err := xml.Unmarshal(resp.Body(), &xmlData); err != nil {
		return nil, fmt.Errorf("failed to parse N0NBH XML response: %w", err)
	}
	
	// Convert XML structure to expected JSON structure - PRESERVE ALL RICH FIELDS
	data := &models.N0NBHResponse{
		Time:   xmlData.Time,
		Source: "N0NBH",
	}
	
	// Map all fields including rich XML fields
	data.SolarData.SolarFlux = xmlData.SolarData.SolarFlux
	data.SolarData.AIndex = xmlData.SolarData.AIndex
	data.SolarData.KIndex = xmlData.SolarData.KIndex
	data.SolarData.KIndexNT = xmlData.SolarData.KIndexNT
	data.SolarData.SunSpots = xmlData.SolarData.SunSpots
	data.SolarData.HeliumLine = xmlData.SolarData.HeliumLine
	data.SolarData.ProtonFlux = xmlData.SolarData.ProtonFlux
	data.SolarData.ElectronFlux = xmlData.SolarData.ElectronFlux
	data.SolarData.Aurora = xmlData.SolarData.Aurora
	data.SolarData.NormalizationTime = xmlData.SolarData.Normalization
	data.SolarData.LatestSWPCReport = "" // Not in XML
	// Rich fields from XML (previously lost)
	data.SolarData.XRay = xmlData.SolarData.XRay
	data.SolarData.SolarWind = xmlData.SolarData.SolarWind
	data.SolarData.MagneticField = xmlData.SolarData.MagneticField
	data.SolarData.LatDegree = xmlData.SolarData.LatDegree
	
	// Convert band conditions - XML has separate entries for day/night
	bandConditions := make(map[string]models.N0NBHBandCondition)
	
	for _, band := range xmlData.SolarData.CalculatedConditions.Band {
		key := band.Name
		if existing, ok := bandConditions[key]; ok {
			// Update existing entry
			switch band.Time {
			case "day":
				existing.Day = band.Condition
			case "night":
				existing.Night = band.Condition
			}
			bandConditions[key] = existing
		} else {
			// Create new entry
			newBand := models.N0NBHBandCondition{
				Name: band.Name,
				Time: band.Time,
			}
			switch band.Time {
			case "day":
				newBand.Day = band.Condition
			case "night":
				newBand.Night = band.Condition
			}
			bandConditions[key] = newBand
		}
	}

	// Convert map to slice
	for _, bandCond := range bandConditions {
		data.Calculatedconditions.Band = append(data.Calculatedconditions.Band, models.N0NBHBandCondition{
			Name:   bandCond.Name,
			Time:   bandCond.Time,
			Day:    bandCond.Day,
			Night:  bandCond.Night,
			Source: "N0NBH",
		})
	}
	
	return data, nil
}
