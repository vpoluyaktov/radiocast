package fetchers

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"radiocast/internal/models"

	"github.com/go-resty/resty/v2"
	"github.com/mmcdole/gofeed"
)

// DataFetcher handles fetching data from all external sources
type DataFetcher struct {
	client *resty.Client
	parser *gofeed.Parser
}

// NewDataFetcher creates a new data fetcher instance
func NewDataFetcher() *DataFetcher {
	client := resty.New()
	client.SetTimeout(30 * time.Second)
	client.SetRetryCount(3)
	client.SetRetryWaitTime(2 * time.Second)
	
	return &DataFetcher{
		client: client,
		parser: gofeed.NewParser(),
	}
}

// FetchAllData fetches and normalizes data from all sources
func (f *DataFetcher) FetchAllData(ctx context.Context, noaaKURL, noaaSolarURL, n0nbhURL, sidcURL string) (*models.PropagationData, error) {
	log.Println("Starting data fetch from all sources...")
	
	// Fetch data from all sources concurrently
	kIndexChan := make(chan []models.NOAAKIndexResponse, 1)
	solarChan := make(chan []models.NOAASolarResponse, 1)
	n0nbhChan := make(chan *models.N0NBHResponse, 1)
	sidcChan := make(chan []*gofeed.Item, 1)
	
	errChan := make(chan error, 4)
	
	// NOAA K-index data
	go func() {
		data, err := f.fetchNOAAKIndex(ctx, noaaKURL)
		if err != nil {
			errChan <- fmt.Errorf("NOAA K-index fetch failed: %w", err)
			return
		}
		kIndexChan <- data
	}()
	
	// NOAA Solar data
	go func() {
		data, err := f.fetchNOAASolar(ctx, noaaSolarURL)
		if err != nil {
			errChan <- fmt.Errorf("NOAA Solar fetch failed: %w", err)
			return
		}
		solarChan <- data
	}()
	
	// N0NBH data
	go func() {
		data, err := f.fetchN0NBH(ctx, n0nbhURL)
		if err != nil {
			errChan <- fmt.Errorf("N0NBH fetch failed: %w", err)
			return
		}
		n0nbhChan <- data
	}()
	
	// SIDC RSS data
	go func() {
		data, err := f.fetchSIDC(ctx, sidcURL)
		if err != nil {
			errChan <- fmt.Errorf("SIDC fetch failed: %w", err)
			return
		}
		sidcChan <- data
	}()
	
	// Collect results
	var kIndexData []models.NOAAKIndexResponse
	var solarData []models.NOAASolarResponse
	var n0nbhData *models.N0NBHResponse
	var sidcData []*gofeed.Item
	
	completed := 0
	for completed < 4 {
		select {
		case data := <-kIndexChan:
			kIndexData = data
			completed++
		case data := <-solarChan:
			solarData = data
			completed++
		case data := <-n0nbhChan:
			n0nbhData = data
			completed++
		case data := <-sidcChan:
			sidcData = data
			completed++
		case err := <-errChan:
			log.Printf("Data fetch error: %v", err)
			completed++
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	
	// Normalize and combine all data
	propagationData := f.normalizeData(kIndexData, solarData, n0nbhData, sidcData)
	
	log.Println("Data fetch and normalization completed successfully")
	return propagationData, nil
}

// fetchNOAAKIndex fetches K-index data from NOAA
func (f *DataFetcher) fetchNOAAKIndex(ctx context.Context, url string) ([]models.NOAAKIndexResponse, error) {
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
		})
	}
	
	return data, nil
}

// fetchNOAASolar fetches solar data from NOAA
func (f *DataFetcher) fetchNOAASolar(ctx context.Context, url string) ([]models.NOAASolarResponse, error) {
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
		})
	}
	
	return data, nil
}

// fetchN0NBH fetches data from N0NBH solar API (XML format)
func (f *DataFetcher) fetchN0NBH(ctx context.Context, url string) (*models.N0NBHResponse, error) {
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

// fetchSIDC fetches sunspot data from SIDC (CSV format)
func (f *DataFetcher) fetchSIDC(ctx context.Context, url string) ([]*gofeed.Item, error) {
	// Use the working CSV endpoint instead of the broken RSS endpoint
	workingURL := "https://www.sidc.be/SILSO/INFO/snmtotcsv.php"
	
	resp, err := f.client.R().
		SetContext(ctx).
		Get(workingURL)
	
	if err != nil {
		return nil, fmt.Errorf("failed to fetch SIDC data: %w", err)
	}
	
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("SIDC API returned status %d", resp.StatusCode())
	}
	
	bodyStr := string(resp.Body())
	
	// Parse as CSV data (SIDC format: Year;Month;Date_fraction;SSN_value;SSN_error;Nb_observations;Definitive)
	return f.parseSIDCCSV(bodyStr)
}

// parseSIDCCSV parses SIDC CSV data and converts to RSS-like items
// Format: Year;Month;Date_fraction;SSN_value;SSN_error;Nb_observations;Definitive
func (f *DataFetcher) parseSIDCCSV(csvData string) ([]*gofeed.Item, error) {
	lines := strings.Split(csvData, "\n")
	var items []*gofeed.Item
	
	// Get recent entries (last 100 lines for recent months)
	startIdx := len(lines) - 100
	if startIdx < 0 {
		startIdx = 0
	}
	
	for i := startIdx; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		// Parse semicolon-separated values
		fields := strings.Split(line, ";")
		if len(fields) < 4 {
			continue
		}
		
		year := strings.TrimSpace(fields[0])
		month := strings.TrimSpace(fields[1])
		ssnValue := strings.TrimSpace(fields[3])
		
		if year == "" || month == "" || ssnValue == "" {
			continue
		}
		
		// Create a feed item from CSV data
		item := &gofeed.Item{
			Title:       fmt.Sprintf("Monthly Sunspot Number: %s", ssnValue),
			Description: fmt.Sprintf("Date: %s-%s, SSN: %s", year, month, ssnValue),
		}
		
		// Parse date (year and month only for monthly data)
		if yearInt, err := strconv.Atoi(year); err == nil {
			if monthInt, err := strconv.Atoi(month); err == nil {
				date := time.Date(yearInt, time.Month(monthInt), 1, 0, 0, 0, 0, time.UTC)
				item.PublishedParsed = &date
			}
		}
		
		items = append(items, item)
	}
	
	return items, nil
}

// normalizeData combines and normalizes data from all sources
func (f *DataFetcher) normalizeData(kIndex []models.NOAAKIndexResponse, solar []models.NOAASolarResponse, n0nbh *models.N0NBHResponse, sidc []*gofeed.Item) *models.PropagationData {
	now := time.Now()
	
	data := &models.PropagationData{
		Timestamp: now,
	}
	
	// Process NOAA K-index data (get latest)
	if len(kIndex) > 0 {
		latest := kIndex[len(kIndex)-1]
		data.GeomagData.KIndex = latest.KpIndex
		if latest.EstimatedKp > 0 {
			data.GeomagData.KIndex = latest.EstimatedKp
		}
		
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
	
	// Process NOAA Solar data (get latest)
	if len(solar) > 0 {
		latest := solar[len(solar)-1]
		data.SolarData.SolarFluxIndex = latest.SolarFlux
		data.SolarData.SunspotNumber = int(latest.SunspotNumber)
		
		// Determine solar activity level
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
		}
		
		if aIndex, err := strconv.ParseFloat(n0nbh.SolarData.AIndex, 64); err == nil {
			data.GeomagData.AIndex = aIndex
		}
		
		if protonFlux, err := strconv.ParseFloat(n0nbh.SolarData.ProtonFlux, 64); err == nil {
			data.SolarData.ProtonFlux = protonFlux
		}
		
		// Process band conditions
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
	data.Forecast = f.generateBasicForecast(data)
	
	return data
}

// generateBasicForecast creates a basic forecast based on current conditions
func (f *DataFetcher) generateBasicForecast(data *models.PropagationData) models.ForecastData {
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
