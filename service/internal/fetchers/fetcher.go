package fetchers

import (
	"context"
	"fmt"
	"log"
	"time"

	"radiocast/internal/models"

	"github.com/go-resty/resty/v2"
	"github.com/mmcdole/gofeed"
)

// DataFetcher handles fetching data from all external sources
type DataFetcher struct {
	client      *resty.Client
	noaaFetcher *NOAAFetcher
	n0nbhFetcher *N0NBHFetcher
	sidcFetcher *SIDCFetcher
	normalizer  *DataNormalizer
}

// NewDataFetcher creates a new data fetcher instance
func NewDataFetcher() *DataFetcher {
	client := resty.New()
	client.SetTimeout(30 * time.Second)
	client.SetRetryCount(3)
	client.SetRetryWaitTime(2 * time.Second)
	
	return &DataFetcher{
		client:       client,
		noaaFetcher:  NewNOAAFetcher(client),
		n0nbhFetcher: NewN0NBHFetcher(client),
		sidcFetcher:  NewSIDCFetcher(client),
		normalizer:   NewDataNormalizer(),
	}
}

// FetchAllDataWithSources fetches raw data from all sources and returns both raw and normalized data
func (f *DataFetcher) FetchAllDataWithSources(ctx context.Context, noaaKURL, noaaSolarURL, n0nbhURL, sidcURL string) (*models.PropagationData, *models.SourceData, error) {
	log.Println("Starting data fetch from all sources...")
	
	// Fetch data from all sources concurrently
	kIndexChan := make(chan []models.NOAAKIndexResponse, 1)
	solarChan := make(chan []models.NOAASolarResponse, 1)
	n0nbhChan := make(chan *models.N0NBHResponse, 1)
	sidcChan := make(chan []*gofeed.Item, 1)
	
	errChan := make(chan error, 4)
	
	// NOAA K-index data
	go func() {
		log.Println("Fetching NOAA K-index data...")
		data, err := f.noaaFetcher.FetchKIndex(ctx, noaaKURL)
		if err != nil {
			log.Printf("NOAA K-index fetch failed: %v", err)
			errChan <- fmt.Errorf("NOAA K-index fetch failed: %w", err)
			return
		}
		log.Printf("NOAA K-index fetch successful: %d data points", len(data))
		kIndexChan <- data
	}()
	
	// NOAA Solar data
	go func() {
		log.Println("Fetching NOAA Solar data...")
		data, err := f.noaaFetcher.FetchSolar(ctx, noaaSolarURL)
		if err != nil {
			log.Printf("NOAA Solar fetch failed: %v", err)
			errChan <- fmt.Errorf("NOAA Solar fetch failed: %w", err)
			return
		}
		log.Printf("NOAA Solar fetch successful: %d data points", len(data))
		solarChan <- data
	}()
	
	// N0NBH data
	go func() {
		log.Println("Fetching N0NBH solar data...")
		data, err := f.n0nbhFetcher.Fetch(ctx, n0nbhURL)
		if err != nil {
			log.Printf("N0NBH fetch failed: %v", err)
			errChan <- fmt.Errorf("N0NBH fetch failed: %w", err)
			return
		}
		log.Printf("N0NBH fetch successful: Solar flux=%s, K-index=%s, band conditions=%d", 
			data.SolarData.SolarFlux, data.SolarData.KIndex, len(data.Calculatedconditions.Band))
		n0nbhChan <- data
	}()
	
	// SIDC RSS data
	go func() {
		log.Println("Fetching SIDC sunspot data...")
		data, err := f.sidcFetcher.Fetch(ctx, sidcURL)
		if err != nil {
			log.Printf("SIDC fetch failed: %v", err)
			errChan <- fmt.Errorf("SIDC fetch failed: %w", err)
			return
		}
		log.Printf("SIDC fetch successful: %d data points", len(data))
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
			return nil, nil, ctx.Err()
		}
	}
	
	// Create source data structure
	sourceData := &models.SourceData{
		NOAAKIndex: kIndexData,
		NOAASolar:  solarData,
		N0NBH:      n0nbhData,
		SIDC:       sidcData,
	}
	
	// Normalize and combine all data
	propagationData := f.normalizer.NormalizeData(kIndexData, solarData, n0nbhData, sidcData)
	
	log.Printf("Data fetch and normalization completed successfully - NOAA K-index: %d points, NOAA Solar: %d points, N0NBH: %v, SIDC: %d points", 
		len(kIndexData), len(solarData), n0nbhData != nil, len(sidcData))
	return propagationData, sourceData, nil
}

// FetchAllData provides backward compatibility - fetches and normalizes data from all sources
func (f *DataFetcher) FetchAllData(ctx context.Context, noaaKURL, noaaSolarURL, n0nbhURL, sidcURL string) (*models.PropagationData, error) {
	data, _, err := f.FetchAllDataWithSources(ctx, noaaKURL, noaaSolarURL, n0nbhURL, sidcURL)
	return data, err
}
