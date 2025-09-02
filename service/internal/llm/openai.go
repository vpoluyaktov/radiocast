package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/sashabaranov/go-openai"

	"radiocast/internal/fetchers"
	"radiocast/internal/models"
)

// OpenAIClient handles OpenAI API interactions
type OpenAIClient struct {
	client *openai.Client
	model  string
}

// NewOpenAIClient creates a new OpenAI client
func NewOpenAIClient(apiKey, model string) *OpenAIClient {
	return &OpenAIClient{
		client: openai.NewClient(apiKey),
		model:  model,
	}
}

// GenerateReport generates a propagation report using OpenAI
func (c *OpenAIClient) GenerateReport(data *models.PropagationData) (string, error) {
	return c.GenerateReportWithSources(data, nil)
}

// GenerateReportWithSources generates a propagation report using OpenAI with raw source data
func (c *OpenAIClient) GenerateReportWithSources(data *models.PropagationData, sourceData *fetchers.SourceData) (string, error) {
	if c.client == nil {
		return "", fmt.Errorf("OpenAI client not initialized")
	}

	log.Printf("Generating report for %s", data.Timestamp.Format("2006-01-02"))

	// Load system prompt from file
	systemPrompt, err := c.loadSystemPrompt()
	if err != nil {
		log.Printf("Failed to load system prompt: %v", err)
		systemPrompt = c.getDefaultSystemPrompt()
	}

	// Build prompt with raw data if available, otherwise use normalized data
	var prompt string
	if sourceData != nil {
		prompt = c.buildPromptWithRawData(sourceData, data)
	} else {
		prompt = c.buildPrompt(data)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	resp, err := c.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: c.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			MaxTokens:   16000,
			Temperature: 0.3,
		},
	)

	if err != nil {
		log.Printf("OpenAI API error: %v", err)
		return "", fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	report := resp.Choices[0].Message.Content
	log.Printf("Generated report with %d characters", len(report))
	
	return report, nil
}

// loadSystemPrompt loads the system prompt from file
func (c *OpenAIClient) loadSystemPrompt() (string, error) {
	// Try multiple possible paths
	possiblePaths := []string{
		filepath.Join("internal", "templates", "system_prompt.txt"),
		filepath.Join("service", "internal", "templates", "system_prompt.txt"),
		"internal/templates/system_prompt.txt",
		"service/internal/templates/system_prompt.txt",
	}
	
	for _, promptPath := range possiblePaths {
		content, err := os.ReadFile(promptPath)
		if err == nil {
			return string(content), nil
		}
	}
	
	return "", fmt.Errorf("system prompt file not found in any expected location")
}

// getDefaultSystemPrompt returns a fallback system prompt
func (c *OpenAIClient) getDefaultSystemPrompt() string {
	return "You are an expert radio propagation analyst and amateur radio operator. Generate a comprehensive daily radio propagation report in markdown format based on the provided solar and space weather data. Focus on practical advice for amateur radio operators."
}

// BuildPrompt constructs data for the LLM (instructions are in system prompt) - public method
func (c *OpenAIClient) BuildPrompt(data *models.PropagationData) string {
	return c.buildPrompt(data)
}

// BuildPromptWithRawData constructs prompt with raw JSON data - public method
func (c *OpenAIClient) BuildPromptWithRawData(sourceData *fetchers.SourceData, data *models.PropagationData) string {
	return c.buildPromptWithRawData(sourceData, data)
}

// GetSystemPrompt returns the system prompt used for LLM - public method
func (c *OpenAIClient) GetSystemPrompt() string {
	systemPrompt, err := c.loadSystemPrompt()
	if err != nil {
		log.Printf("Failed to load system prompt: %v", err)
		return c.getDefaultSystemPrompt()
	}
	return systemPrompt
}

// buildPrompt constructs data for the LLM (instructions are in system prompt)
func (c *OpenAIClient) buildPrompt(data *models.PropagationData) string {
	prompt := fmt.Sprintf(`## Current Solar and Geomagnetic Data (as of %s)

### Solar Activity:
- Solar Flux Index (10.7cm): %.1f sfu
- Sunspot Number: %d
- Solar Activity Level: %s
- Proton Flux: %.2e particles/cm²/s

### Geomagnetic Activity:
- Planetary K-index: %.1f
- A-index: %.1f
- Geomagnetic Activity Level: %s
- Current Conditions: %s

### HF Band Conditions:
- 80m: Day=%s, Night=%s
- 40m: Day=%s, Night=%s
- 20m: Day=%s, Night=%s
- 17m: Day=%s, Night=%s
- 15m: Day=%s, Night=%s
- 12m: Day=%s, Night=%s
- 10m: Day=%s, Night=%s
- 6m: Day=%s, Night=%s

### 3-Day Forecast:
- Today: %s (K-index: %s)
- Tomorrow: %s (K-index: %s)
- Day After: %s (K-index: %s)
- General Outlook: %s`,
		data.Timestamp.Format("2006-01-02 15:04 UTC"),
		data.SolarData.SolarFluxIndex,
		data.SolarData.SunspotNumber,
		data.SolarData.SolarActivity,
		data.SolarData.ProtonFlux,
		data.GeomagData.KIndex,
		data.GeomagData.AIndex,
		data.GeomagData.GeomagActivity,
		data.GeomagData.GeomagConditions,
		data.BandData.Band80m.Day, data.BandData.Band80m.Night,
		data.BandData.Band40m.Day, data.BandData.Band40m.Night,
		data.BandData.Band20m.Day, data.BandData.Band20m.Night,
		data.BandData.Band17m.Day, data.BandData.Band17m.Night,
		data.BandData.Band15m.Day, data.BandData.Band15m.Night,
		data.BandData.Band12m.Day, data.BandData.Band12m.Night,
		data.BandData.Band10m.Day, data.BandData.Band10m.Night,
		data.BandData.Band6m.Day, data.BandData.Band6m.Night,
		data.Forecast.Today.HFConditions, data.Forecast.Today.KIndexForecast,
		data.Forecast.Tomorrow.HFConditions, data.Forecast.Tomorrow.KIndexForecast,
		data.Forecast.DayAfter.HFConditions, data.Forecast.DayAfter.KIndexForecast,
		data.Forecast.Outlook,
	)
	
	// Add recent events if any
	if len(data.SourceEvents) > 0 {
		prompt += "\n\n### Recent Solar/Space Weather Events:\n"
		for _, event := range data.SourceEvents {
			prompt += fmt.Sprintf("- %s (%s): %s [%s severity]\n",
				event.EventType, event.Source, event.Description, event.Severity)
		}
	}
	
	// Add warnings if any
	if len(data.Forecast.Warnings) > 0 {
		prompt += "\n\n### Current Warnings:\n"
		for _, warning := range data.Forecast.Warnings {
			prompt += fmt.Sprintf("- %s\n", warning)
		}
	}

	return prompt
}

// buildPromptWithRawData constructs prompt using raw JSON data from all sources
func (c *OpenAIClient) buildPromptWithRawData(sourceData *fetchers.SourceData, data *models.PropagationData) string {
	prompt := fmt.Sprintf(`## Raw Solar and Space Weather Data (as of %s)

Please analyze the following comprehensive space weather data and generate a detailed radio propagation report. The data includes:
- NOAA K-index data (geomagnetic activity - last 24 hours, 3-hour intervals)
- NOAA Solar data (solar flux, sunspot numbers - last 7 days)
- N0NBH real-time conditions (band conditions, solar metrics)
- SIDC solar event alerts (last 30 days)

`, data.Timestamp.Format("2006-01-02 15:04 UTC"))

	// Filter data to last 30 days
	thirtyDaysAgo := data.Timestamp.AddDate(0, 0, -30)

	// Add NOAA K-index data (last 24 hours only, every 3 hours)
	if len(sourceData.NOAAKIndex) > 0 {
		prompt += "### NOAA K-Index Data (Last 24 Hours, 3-Hour Intervals):\n```json\n"
		recentKIndex := c.filterKIndexRecent(sourceData.NOAAKIndex)
		if jsonData, err := json.MarshalIndent(recentKIndex, "", "  "); err == nil {
			prompt += string(jsonData)
		} else {
			prompt += "Error marshaling NOAA K-index data"
		}
		prompt += "\n```\n\n"
	}

	// Add NOAA Solar data (last 7 days only)
	if len(sourceData.NOAASolar) > 0 {
		prompt += "### NOAA Solar Data (Last 7 Days):\n```json\n"
		// Take last 7 entries to avoid token limits
		recentSolar := sourceData.NOAASolar
		if len(recentSolar) > 7 {
			recentSolar = recentSolar[len(recentSolar)-7:]
		}
		if jsonData, err := json.MarshalIndent(recentSolar, "", "  "); err == nil {
			prompt += string(jsonData)
		} else {
			prompt += "Error marshaling NOAA Solar data"
		}
		prompt += "\n```\n\n"
	}

	// Add N0NBH data
	if sourceData.N0NBH != nil {
		prompt += "### N0NBH Real-time Data (Current Conditions):\n```json\n"
		if jsonData, err := json.MarshalIndent(sourceData.N0NBH, "", "  "); err == nil {
			prompt += string(jsonData)
		} else {
			prompt += "Error marshaling N0NBH data"
		}
		prompt += "\n```\n\n"
	}

	// Add SIDC data (last 30 days only)
	if len(sourceData.SIDC) > 0 {
		prompt += "### SIDC Solar Event Alerts (Last 30 Days):\n```json\n"
		recentSIDC := c.filterSIDCByDate(sourceData.SIDC, thirtyDaysAgo)
		if jsonData, err := json.MarshalIndent(recentSIDC, "", "  "); err == nil {
			prompt += string(jsonData)
		} else {
			prompt += "Error marshaling SIDC data"
		}
		prompt += "\n```\n\n"
	}

	prompt += `### Instructions:
Analyze all the above data and provide:
1. Current solar activity summary (solar flux, sunspots, flares)
2. Geomagnetic conditions (K-index trends, magnetic field)
3. HF band conditions for each amateur band (80m-10m)
4. VHF/UHF propagation outlook
5. 3-day forecast with specific recommendations
6. Best/worst bands for current conditions
7. Any alerts or warnings for amateur radio operators

Focus on practical advice for amateur radio operators based on the comprehensive data provided.`

	return prompt
}

// filterKIndexByDate filters K-index data to only include entries within the specified time range
func (c *OpenAIClient) filterKIndexByDate(kIndexData []models.NOAAKIndexResponse, cutoffDate time.Time) []models.NOAAKIndexResponse {
	var filtered []models.NOAAKIndexResponse
	
	for _, entry := range kIndexData {
		if entryTime, err := time.Parse("2006-01-02T15:04:05", entry.TimeTag); err == nil {
			if entryTime.After(cutoffDate) {
				filtered = append(filtered, entry)
			}
		}
	}
	
	return filtered
}

// filterKIndexRecent filters K-index data to last 24 hours with 3-hour intervals
func (c *OpenAIClient) filterKIndexRecent(kIndexData []models.NOAAKIndexResponse) []models.NOAAKIndexResponse {
	if len(kIndexData) == 0 {
		return kIndexData
	}
	
	var filtered []models.NOAAKIndexResponse
	now := time.Now()
	twentyFourHoursAgo := now.Add(-24 * time.Hour)
	
	// Sample every 3 hours (180 minutes)
	lastSampleTime := time.Time{}
	
	for _, entry := range kIndexData {
		if entryTime, err := time.Parse("2006-01-02T15:04:05", entry.TimeTag); err == nil {
			if entryTime.After(twentyFourHoursAgo) {
				// Include if it's the first entry or 3+ hours since last sample
				if lastSampleTime.IsZero() || entryTime.Sub(lastSampleTime) >= 3*time.Hour {
					filtered = append(filtered, entry)
					lastSampleTime = entryTime
				}
			}
		}
	}
	
	// If no entries found, take the last 8 entries (roughly last day)
	if len(filtered) == 0 && len(kIndexData) > 0 {
		start := len(kIndexData) - 8
		if start < 0 {
			start = 0
		}
		filtered = kIndexData[start:]
	}
	
	return filtered
}

// filterSIDCByDate filters SIDC data to only include entries within the specified time range
func (c *OpenAIClient) filterSIDCByDate(sidcData []*gofeed.Item, cutoffDate time.Time) []*gofeed.Item {
	var filtered []*gofeed.Item
	
	for _, entry := range sidcData {
		if entry.PublishedParsed != nil && entry.PublishedParsed.After(cutoffDate) {
			filtered = append(filtered, entry)
		}
	}
	
	return filtered
}
