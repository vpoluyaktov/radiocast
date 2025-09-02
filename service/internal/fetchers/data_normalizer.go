package fetchers

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"radiocast/internal/models"

	"github.com/mmcdole/gofeed"
)

// DataNormalizer handles data normalization and forecast generation
type DataNormalizer struct{}

// NewDataNormalizer creates a new data normalizer instance
func NewDataNormalizer() *DataNormalizer {
	return &DataNormalizer{}
}

// NormalizeData combines and normalizes data from all sources
func (n *DataNormalizer) NormalizeData(kIndex []models.NOAAKIndexResponse, solar []models.NOAASolarResponse, n0nbh *models.N0NBHResponse, sidc []*gofeed.Item) *models.PropagationData {
	now := time.Now()
	
	data := &models.PropagationData{
		Timestamp: now,
	}
	
	// Process K-index data from NOAA
	if len(kIndex) > 0 {
		latest := kIndex[len(kIndex)-1]
		data.GeomagData.KIndex = latest.KpIndex
		data.GeomagData.KIndexDataSource = latest.Source
		if latest.EstimatedKp > 0 {
			data.GeomagData.KIndex = latest.EstimatedKp
		}
		log.Printf("DEBUG: Set K-Index source to: %s", data.GeomagData.KIndexDataSource)
		
		// Determine geomagnetic activity level
		if data.GeomagData.KIndex <= 2 {
			data.GeomagData.GeomagActivity = "Quiet"
		} else if data.GeomagData.KIndex <= 3 {
			data.GeomagData.GeomagActivity = "Unsettled"
		} else if data.GeomagData.KIndex <= 4 {
			data.GeomagData.GeomagActivity = "Active"
		} else {
			data.GeomagData.GeomagActivity = "Storm"
		}
	}
	
	// Process solar data from NOAA
	if len(solar) > 0 {
		latest := solar[len(solar)-1]
		data.SolarData.SolarFluxIndex = latest.SolarFlux
		data.SolarData.SolarFluxDataSource = latest.Source
		data.SolarData.SunspotNumber = int(latest.SunspotNumber)
		data.SolarData.SunspotDataSource = latest.Source
		log.Printf("DEBUG: Set Solar Flux source to: %s", data.SolarData.SolarFluxDataSource)
		log.Printf("DEBUG: Set Sunspot source to: %s", data.SolarData.SunspotDataSource)
	}
	
	// Classify solar activity based on solar flux index
	if data.SolarData.SolarFluxIndex > 0 {
		if data.SolarData.SolarFluxIndex < 100 {
			data.SolarData.SolarActivity = "Low"
		} else if data.SolarData.SolarFluxIndex < 150 {
			data.SolarData.SolarActivity = "Moderate"
		} else {
			data.SolarData.SolarActivity = "High"
		}
	}
	
	// Process N0NBH data
	if n0nbh != nil {
		// Parse additional solar data
		if flux, err := strconv.ParseFloat(n0nbh.SolarData.SolarFlux, 64); err == nil {
			if data.SolarData.SolarFluxIndex == 0 {
				data.SolarData.SolarFluxIndex = flux
			}
			// Always set N0NBH as source for solar flux when available
			data.SolarData.SolarFluxDataSource = n0nbh.Source
			log.Printf("DEBUG: Set Solar Flux source to N0NBH: %s", data.SolarData.SolarFluxDataSource)
			
			// Re-classify solar activity after N0NBH data update
			if data.SolarData.SolarFluxIndex < 100 {
				data.SolarData.SolarActivity = "Low"
			} else if data.SolarData.SolarFluxIndex < 150 {
				data.SolarData.SolarActivity = "Moderate"
			} else {
				data.SolarData.SolarActivity = "High"
			}
		}
		
		// Parse A-index
		if aIndex, err := strconv.ParseFloat(n0nbh.SolarData.AIndex, 64); err == nil {
			data.GeomagData.AIndex = aIndex
			data.GeomagData.AIndexDataSource = n0nbh.Source
			log.Printf("DEBUG: Set A-Index source to: %s", data.GeomagData.AIndexDataSource)
		}
		
		// Parse proton flux
		if protonFlux, err := strconv.ParseFloat(n0nbh.SolarData.ProtonFlux, 64); err == nil {
			data.SolarData.ProtonFlux = protonFlux
			data.SolarData.ProtonFluxDataSource = n0nbh.Source
			log.Printf("DEBUG: Set Proton Flux source to: %s", data.SolarData.ProtonFluxDataSource)
		}
		
		// Process band conditions
		data.BandData.BandDataSource = n0nbh.Source
		for _, band := range n0nbh.Calculatedconditions.Band {
			condition := models.BandCondition{
				Day:   band.Day,
				Night: band.Night,
			}
			
			switch strings.ToLower(band.Name) {
			case "80m-40m":
				data.BandData.Band80m = condition
				data.BandData.Band40m = condition
			case "30m-20m":
				data.BandData.Band20m = condition
			case "17m-15m":
				data.BandData.Band17m = condition
				data.BandData.Band15m = condition
			case "12m-10m":
				data.BandData.Band12m = condition
				data.BandData.Band10m = condition
			case "6m":
				data.BandData.Band6m = condition
			}
		}
	}
	
	// Process SIDC events
	for _, item := range sidc {
		if item.PublishedParsed != nil && item.PublishedParsed.After(now.Add(-24*time.Hour)) {
			event := models.SourceEvent{
				Source:      "SIDC",
				EventType:   "Solar Event",
				Description: item.Title,
				Timestamp:   *item.PublishedParsed,
				Impact:      "Variable", // Would need more parsing to determine
			}
			
			// Simple severity classification based on keywords
			title := strings.ToLower(item.Title)
			if strings.Contains(title, "x-class") || strings.Contains(title, "extreme") {
				event.Severity = "Extreme"
			} else if strings.Contains(title, "m-class") || strings.Contains(title, "major") {
				event.Severity = "High"
			} else if strings.Contains(title, "c-class") || strings.Contains(title, "moderate") {
				event.Severity = "Moderate"
			} else {
				event.Severity = "Low"
			}
			
			data.SourceEvents = append(data.SourceEvents, event)
		}
	}
	
	// Generate basic forecast
	data.Forecast = n.GenerateBasicForecast(data)
	
	// Debug: Log all source attributions before returning
	log.Printf("DEBUG: Final source attribution - Solar Flux: '%s', Sunspot: '%s', K-Index: '%s', A-Index: '%s', Band Data: '%s'", 
		data.SolarData.SolarFluxDataSource, 
		data.SolarData.SunspotDataSource,
		data.GeomagData.KIndexDataSource,
		data.GeomagData.AIndexDataSource,
		data.BandData.BandDataSource)
	
	return data
}

// GenerateBasicForecast creates a basic forecast based on current conditions
func (n *DataNormalizer) GenerateBasicForecast(data *models.PropagationData) models.ForecastData {
	forecast := models.ForecastData{
		Today: models.DayForecast{
			Date: time.Now(),
		},
		Tomorrow: models.DayForecast{
			Date: time.Now().Add(24 * time.Hour),
		},
		DayAfter: models.DayForecast{
			Date: time.Now().Add(48 * time.Hour),
		},
	}
	
	// Basic forecast logic based on current conditions
	kIndex := data.GeomagData.KIndex
	solarFlux := data.SolarData.SolarFluxIndex
	
	// Determine HF conditions
	var hfConditions string
	if kIndex <= 2 && solarFlux > 120 {
		hfConditions = "Good to Excellent"
		forecast.Today.BestBands = []string{"20m", "17m", "15m", "12m", "10m"}
	} else if kIndex <= 3 && solarFlux > 100 {
		hfConditions = "Fair to Good"
		forecast.Today.BestBands = []string{"40m", "20m", "17m"}
	} else {
		hfConditions = "Poor to Fair"
		forecast.Today.BestBands = []string{"80m", "40m"}
		forecast.Today.WorstBands = []string{"15m", "12m", "10m"}
	}
	
	forecast.Today.HFConditions = hfConditions
	forecast.Tomorrow.HFConditions = hfConditions // Simplified
	forecast.DayAfter.HFConditions = hfConditions
	
	// K-index forecast
	forecast.Today.KIndexForecast = fmt.Sprintf("%.1f", kIndex)
	forecast.Tomorrow.KIndexForecast = fmt.Sprintf("%.1f-%.1f", kIndex-0.5, kIndex+0.5)
	forecast.DayAfter.KIndexForecast = fmt.Sprintf("%.1f-%.1f", kIndex-1, kIndex+1)
	
	// General outlook
	if kIndex <= 2 {
		forecast.Outlook = "Stable geomagnetic conditions expected. Good propagation likely."
	} else if kIndex <= 4 {
		forecast.Outlook = "Unsettled to active conditions. Variable propagation expected."
	} else {
		forecast.Outlook = "Geomagnetic storm conditions. Poor HF propagation likely."
		forecast.Warnings = append(forecast.Warnings, "Geomagnetic storm in progress - expect poor HF conditions")
	}
	
	return forecast
}
