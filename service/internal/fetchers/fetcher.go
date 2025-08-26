package fetchers

import (
	"context"
	"encoding/json"
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
	
	var data []models.NOAAKIndexResponse
	if err := json.Unmarshal(resp.Body(), &data); err != nil {
		return nil, fmt.Errorf("failed to parse NOAA K-index response: %w", err)
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
	
	var data []models.NOAASolarResponse
	if err := json.Unmarshal(resp.Body(), &data); err != nil {
		return nil, fmt.Errorf("failed to parse NOAA solar response: %w", err)
	}
	
	return data, nil
}

// fetchN0NBH fetches data from N0NBH solar API
func (f *DataFetcher) fetchN0NBH(ctx context.Context, url string) (*models.N0NBHResponse, error) {
	resp, err := f.client.R().
		SetContext(ctx).
		SetHeader("Accept", "application/json").
		Get(url)
	
	if err != nil {
		return nil, fmt.Errorf("failed to fetch N0NBH data: %w", err)
	}
	
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("N0NBH API returned status %d", resp.StatusCode())
	}
	
	var data models.N0NBHResponse
	if err := json.Unmarshal(resp.Body(), &data); err != nil {
		return nil, fmt.Errorf("failed to parse N0NBH response: %w", err)
	}
	
	return &data, nil
}

// fetchSIDC fetches RSS data from SIDC
func (f *DataFetcher) fetchSIDC(ctx context.Context, url string) ([]*gofeed.Item, error) {
	resp, err := f.client.R().
		SetContext(ctx).
		Get(url)
	
	if err != nil {
		return nil, fmt.Errorf("failed to fetch SIDC RSS: %w", err)
	}
	
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("SIDC RSS returned status %d", resp.StatusCode())
	}
	
	feed, err := f.parser.ParseString(string(resp.Body()))
	if err != nil {
		return nil, fmt.Errorf("failed to parse SIDC RSS: %w", err)
	}
	
	return feed.Items, nil
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
