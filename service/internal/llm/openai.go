package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"radiocast/internal/models"

	"github.com/sashabaranov/go-openai"
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
func (c *OpenAIClient) GenerateReportWithSources(data *models.PropagationData, sourceData *models.SourceData) (string, error) {
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

	// Build prompt with raw data (sourceData should always be available now)
	if sourceData == nil {
		return "", fmt.Errorf("sourceData is required for report generation")
	}
	prompt := c.buildPrompt(sourceData, data)

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

// BuildPrompt constructs prompt with raw JSON data - public method
func (c *OpenAIClient) BuildPrompt(sourceData *models.SourceData, data *models.PropagationData) string {
	return c.buildPrompt(sourceData, data)
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


// buildPrompt constructs prompt using raw JSON data from all sources
func (c *OpenAIClient) buildPrompt(sourceData *models.SourceData, data *models.PropagationData) string {
	prompt := fmt.Sprintf(`## Raw Solar and Space Weather Data (as of %s)

Please analyze the following comprehensive space weather data and generate a detailed radio propagation report. The data includes:
- NOAA K-index data (geomagnetic activity - last 24 hours, 1-hour intervals)
- NOAA Solar data (solar flux, sunspot numbers - last 7 months)
- N0NBH real-time conditions (band conditions, solar metrics)
- SIDC monthly sunspot data (last 12 months)

`, data.Timestamp.Format("2006-01-02 15:04 UTC"))

	// Add NOAA K-index data (pre-filtered by fetcher)
	if len(sourceData.NOAAKIndex) > 0 {
		prompt += "### NOAA K-Index Data (Last 24 Hours, 1-Hour Intervals):\n```json\n"
		if jsonData, err := json.MarshalIndent(sourceData.NOAAKIndex, "", "  "); err == nil {
			prompt += string(jsonData)
		} else {
			prompt += "Error marshaling NOAA K-index data"
		}
		prompt += "\n```\n\n"
	}

	// Add NOAA Solar data (pre-filtered by fetcher)
	if len(sourceData.NOAASolar) > 0 {
		prompt += "### NOAA Solar Data (Last 7 Months):\n```json\n"
		if jsonData, err := json.MarshalIndent(sourceData.NOAASolar, "", "  "); err == nil {
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

	// Add SIDC data (pre-filtered by fetcher)
	if len(sourceData.SIDC) > 0 {
		prompt += "### SIDC Monthly Sunspot Data (Last 12 Months):\n```json\n"
		if jsonData, err := json.MarshalIndent(sourceData.SIDC, "", "  "); err == nil {
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

