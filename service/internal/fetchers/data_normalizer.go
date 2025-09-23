package fetchers

import (
	"strconv"
	"strings"
	"time"

	"radiocast/internal/logger"
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
	
	// Process K-index data from NOAA - PRESERVE ALL HISTORICAL DATA
	if len(kIndex) > 0 {
		// Convert all K-index data to historical points
		for _, kPoint := range kIndex {
			if timestamp, err := parseTimeMulti(kPoint.TimeTag); err == nil {
				data.HistoricalKIndex = append(data.HistoricalKIndex, models.KIndexPoint{
					Timestamp:   timestamp,
					KIndex:      kPoint.KpIndex,
					EstimatedKp: kPoint.EstimatedKp,
					Source:      kPoint.Source,
				})
			}
		}
		
		// Set current values from latest data point
		latest := kIndex[len(kIndex)-1]
		data.GeomagData.KIndex = latest.KpIndex
		data.GeomagData.KIndexDataSource = latest.Source
		if latest.EstimatedKp > 0 {
			data.GeomagData.KIndex = latest.EstimatedKp
		}
		logger.Debugf("DEBUG: Set K-Index source to: %s, preserved %d historical points", data.GeomagData.KIndexDataSource, len(data.HistoricalKIndex))
		
		// Let LLM determine geomagnetic activity level - no hardcoded classification
	}
	
	// Process solar data from NOAA - PRESERVE ALL HISTORICAL DATA
	if len(solar) > 0 {
		// Convert all solar data to historical points
		for _, sPoint := range solar {
			if timestamp, err := parseTimeMulti(sPoint.TimeTag); err == nil {
				data.HistoricalSolar = append(data.HistoricalSolar, models.SolarPoint{
					Timestamp:         timestamp,
					SolarFlux:         sPoint.SolarFlux,
					SolarFluxAdjusted: sPoint.SolarFluxAdjusted,
					SunspotNumber:     sPoint.SunspotNumber,
					Source:            sPoint.Source,
				})
			}
		}
		
		// Set current values from latest data point
		latest := solar[len(solar)-1]
		data.SolarData.SolarFluxIndex = latest.SolarFlux
		data.SolarData.SolarFluxAdjusted = latest.SolarFluxAdjusted  // PRESERVE adjusted value
		data.SolarData.SolarFluxDataSource = latest.Source
		data.SolarData.SunspotNumber = int(latest.SunspotNumber)
		data.SolarData.SunspotDataSource = latest.Source
		logger.Debugf("DEBUG: Set Solar Flux source to: %s, preserved %d historical points", data.SolarData.SolarFluxDataSource, len(data.HistoricalSolar))
		logger.Debugf("DEBUG: Set Sunspot source to: %s", data.SolarData.SunspotDataSource)
	}
	
	// Let LLM classify solar activity based on solar flux index - no hardcoded classification
	
	// Process N0NBH data - EXTRACT ALL RICH FIELDS
	if n0nbh != nil {
		// Parse solar flux data
		if flux, err := strconv.ParseFloat(n0nbh.SolarData.SolarFlux, 64); err == nil {
			if data.SolarData.SolarFluxIndex == 0 {
				data.SolarData.SolarFluxIndex = flux
			}
			// Always set N0NBH as source for solar flux when available
			data.SolarData.SolarFluxDataSource = n0nbh.Source
			logger.Debugf("DEBUG: Set Solar Flux source to N0NBH: %s", data.SolarData.SolarFluxDataSource)
			
			// Let LLM classify solar activity - no hardcoded re-classification
		}
		
		// Parse A-index
		if aIndex, err := strconv.ParseFloat(n0nbh.SolarData.AIndex, 64); err == nil {
			data.GeomagData.AIndex = aIndex
			data.GeomagData.AIndexDataSource = n0nbh.Source
			logger.Debugf("DEBUG: Set A-Index source to: %s", data.GeomagData.AIndexDataSource)
		}
		
		// Parse proton flux
		if protonFlux, err := strconv.ParseFloat(n0nbh.SolarData.ProtonFlux, 64); err == nil {
			data.SolarData.ProtonFlux = protonFlux
			data.SolarData.ProtonFluxDataSource = n0nbh.Source
			logger.Debugf("DEBUG: Set Proton Flux source to: %s", data.SolarData.ProtonFluxDataSource)
		}
		
		// EXTRACT RICH N0NBH FIELDS (previously lost)
		data.SolarData.XRayFlux = n0nbh.SolarData.XRay
		data.SolarData.ElectronFlux = n0nbh.SolarData.ElectronFlux
		data.SolarData.HeliumLine = n0nbh.SolarData.HeliumLine
		data.SolarData.Aurora = n0nbh.SolarData.Aurora
		
		// Parse solar wind speed (previously lost)
		if n0nbh.SolarData.SolarWind != "" {
			if solarWind, err := strconv.ParseFloat(n0nbh.SolarData.SolarWind, 64); err == nil {
				data.SolarData.SolarWindSpeed = solarWind
				data.SolarData.SolarWindDataSource = n0nbh.Source
				logger.Debugf("DEBUG: Set Solar Wind source to: %s", data.SolarData.SolarWindDataSource)
			}
		}
		
		// Parse magnetic field (previously lost)
		if n0nbh.SolarData.MagneticField != "" {
			if magField, err := strconv.ParseFloat(n0nbh.SolarData.MagneticField, 64); err == nil {
				data.GeomagData.MagneticField = magField
				data.GeomagData.MagneticFieldDataSource = n0nbh.Source
				logger.Debugf("DEBUG: Set Magnetic Field source to: %s", data.GeomagData.MagneticFieldDataSource)
			}
		}
		
		// Extract latitude degree (previously lost)
		data.GeomagData.LatDegree = n0nbh.SolarData.LatDegree
		
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
	
	// Process SIDC events - pass through without classification
	for _, item := range sidc {
		if item.PublishedParsed != nil && item.PublishedParsed.After(now.Add(-24*time.Hour)) {
			event := models.SourceEvent{
				Source:      "SIDC",
				EventType:   "Solar Event",
				Description: item.Title,
				Timestamp:   *item.PublishedParsed,
				Impact:      "", // Let LLM determine impact
				Severity:    "", // Let LLM determine severity
			}
			
			// Let LLM classify severity and impact based on event description
			
			data.SourceEvents = append(data.SourceEvents, event)
		}
	}
	
	// Let LLM generate forecast - no hardcoded forecast logic
	
	// Debug: Log all source attributions before returning
	logger.Debugf("DEBUG: Final source attribution - Solar Flux: '%s', Sunspot: '%s', K-Index: '%s', A-Index: '%s', Band Data: '%s'", 
		data.SolarData.SolarFluxDataSource, 
		data.SolarData.SunspotDataSource,
		data.GeomagData.KIndexDataSource,
		data.GeomagData.AIndexDataSource,
		data.BandData.BandDataSource)
	
	return data
}

// Removed GenerateBasicForecast - let LLM handle all forecasting and analysis

// parseTimeMulti attempts to parse time strings with multiple possible layouts
func parseTimeMulti(s string) (time.Time, error) {
	layouts := []string{
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05.000",
		"2006-01-02 15:04:05",
		"2006-01", // For NOAA solar data format like "2025-03"
		time.RFC3339,
	}
	var last error
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil { 
			return t, nil 
		} else { 
			last = err 
		}
	}
	return time.Time{}, last
}
