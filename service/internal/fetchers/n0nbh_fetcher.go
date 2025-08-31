package fetchers

import (
	"context"
	"encoding/xml"
	"fmt"
	"log"

	"radiocast/internal/models"

	"github.com/go-resty/resty/v2"
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
		log.Printf("N0NBH API returned status %d, response: %s", resp.StatusCode(), string(resp.Body()[:bodyLen]))
		return nil, fmt.Errorf("N0NBH API returned status %d", resp.StatusCode())
	}
	
	// Parse XML response
	var xmlData models.N0NBHXMLResponse
	if err := xml.Unmarshal(resp.Body(), &xmlData); err != nil {
		return nil, fmt.Errorf("failed to parse N0NBH XML response: %w", err)
	}
	
	// Convert XML structure to expected JSON structure
	data := &models.N0NBHResponse{
		SolarData: struct {
			SolarFlux     string `json:"solarflux"`
			AIndex        string `json:"aindex"`
			KIndex        string `json:"kindex"`
			KIndexNT      string `json:"kindexnt"`
			SunSpots      string `json:"sunspots"`
			HeliumLine    string `json:"heliumline"`
			ProtonFlux    string `json:"protonflux"`
			ElectronFlux  string `json:"electonflux"`
			Aurora        string `json:"aurora"`
			NormalizationTime string `json:"normalization"`
			LatestSWPCReport  string `json:"latestswpcreport"`
		}{
			SolarFlux:     xmlData.SolarData.SolarFlux,
			AIndex:        xmlData.SolarData.AIndex,
			KIndex:        xmlData.SolarData.KIndex,
			KIndexNT:      xmlData.SolarData.KIndexNT,
			SunSpots:      xmlData.SolarData.SunSpots,
			HeliumLine:    xmlData.SolarData.HeliumLine,
			ProtonFlux:    xmlData.SolarData.ProtonFlux,
			ElectronFlux:  xmlData.SolarData.ElectronFlux,
			Aurora:        xmlData.SolarData.Aurora,
			NormalizationTime: xmlData.SolarData.Normalization,
			LatestSWPCReport:  "", // Not in XML
		},
		Time: xmlData.Time,
	}
	
	// Convert band conditions - XML has separate entries for day/night
	bandConditions := make(map[string]struct {
		Name  string `json:"name"`
		Time  string `json:"time"`
		Day   string `json:"day"`
		Night string `json:"night"`
	})
	
	for _, band := range xmlData.SolarData.CalculatedConditions.Band {
		key := band.Name
		if existing, ok := bandConditions[key]; ok {
			// Update existing entry
			if band.Time == "day" {
				existing.Day = band.Condition
			} else if band.Time == "night" {
				existing.Night = band.Condition
			}
			bandConditions[key] = existing
		} else {
			// Create new entry
			newBand := struct {
				Name  string `json:"name"`
				Time  string `json:"time"`
				Day   string `json:"day"`
				Night string `json:"night"`
			}{
				Name: band.Name,
				Time: band.Time,
			}
			if band.Time == "day" {
				newBand.Day = band.Condition
			} else if band.Time == "night" {
				newBand.Night = band.Condition
			}
			bandConditions[key] = newBand
		}
	}

	// Convert map to slice
	for _, bandCond := range bandConditions {
		data.Calculatedconditions.Band = append(data.Calculatedconditions.Band, struct {
			Name  string `json:"name"`
			Time  string `json:"time"`
			Day   string `json:"day"`
			Night string `json:"night"`
		}{
			Name:  bandCond.Name,
			Time:  bandCond.Time,
			Day:   bandCond.Day,
			Night: bandCond.Night,
		})
	}
	
	return data, nil
}
