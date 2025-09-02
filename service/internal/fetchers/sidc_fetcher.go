package fetchers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/mmcdole/gofeed"
)

// SIDCFetcher handles fetching data from SIDC CSV API
type SIDCFetcher struct {
	client *resty.Client
}

// NewSIDCFetcher creates a new SIDC fetcher instance
func NewSIDCFetcher(client *resty.Client) *SIDCFetcher {
	return &SIDCFetcher{
		client: client,
	}
}

// Fetch fetches sunspot data from SIDC (CSV format)
func (f *SIDCFetcher) Fetch(ctx context.Context, url string) ([]*gofeed.Item, error) {
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
	return f.parseCSV(bodyStr)
}

// parseCSV parses SIDC CSV data and converts to RSS-like items
// Format: Year;Month;Date_fraction;SSN_value;SSN_error;Nb_observations;Definitive
func (f *SIDCFetcher) parseCSV(csvData string) ([]*gofeed.Item, error) {
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
