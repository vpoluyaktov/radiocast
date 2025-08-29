package llm

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/sashabaranov/go-openai"

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

	// Build prompt with all data
	prompt := c.buildPrompt(data)

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
			MaxTokens:   4000,
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
	promptPath := filepath.Join("internal", "templates", "system_prompt.txt")
	content, err := os.ReadFile(promptPath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// getDefaultSystemPrompt returns a fallback system prompt
func (c *OpenAIClient) getDefaultSystemPrompt() string {
	return "You are an expert radio propagation analyst and amateur radio operator. Generate a comprehensive daily radio propagation report in markdown format based on the provided solar and space weather data. Focus on practical advice for amateur radio operators."
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
