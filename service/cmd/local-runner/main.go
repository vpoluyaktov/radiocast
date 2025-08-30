package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"radiocast/internal/config"
	"radiocast/internal/fetchers"
	"radiocast/internal/llm"
	"radiocast/internal/reports"
)

// LocalTester runs report generation without GCS
type LocalTester struct {
	fetcher   *fetchers.DataFetcher
	llmClient *llm.OpenAIClient
	generator *reports.Generator
}

func NewLocalTester(openaiKey, model string) *LocalTester {
	return &LocalTester{
		fetcher:   fetchers.NewDataFetcher(),
		llmClient: llm.NewOpenAIClient(openaiKey, model),
		generator: reports.NewGenerator("reports"), // Use reports directory
	}
}

func (lt *LocalTester) GenerateTestReport() error {
	ctx := context.Background()
	startTime := time.Now()

	log.Println("🚀 Starting local report generation test...")

	// Use default config URLs
	cfg := &config.Config{
		NOAAKIndexURL: "https://services.swpc.noaa.gov/json/planetary_k_index_1m.json",
		NOAASolarURL:  "https://services.swpc.noaa.gov/json/solar-cycle/observed-solar-cycle-indices.json",
		N0NBHSolarURL: "https://www.hamqsl.com/solarapi.php?format=json",
		SIDCRSSURL:    "https://www.sidc.be/products/meu",
	}

	// Fetch data
	log.Println("📡 Fetching data from external sources...")
	data, err := lt.fetcher.FetchAllData(
		ctx,
		cfg.NOAAKIndexURL,
		cfg.NOAASolarURL,
		cfg.N0NBHSolarURL,
		cfg.SIDCRSSURL,
	)
	if err != nil {
		return fmt.Errorf("data fetch failed: %w", err)
	}

	log.Printf("✅ Data fetched successfully:")
	log.Printf("   Solar Flux: %.1f", data.SolarData.SolarFluxIndex)
	log.Printf("   K-Index: %.1f", data.GeomagData.KIndex)
	log.Printf("   Sunspot Number: %d", data.SolarData.SunspotNumber)
	log.Printf("   Activity: %s", data.SolarData.SolarActivity)

	// Generate report with LLM
	log.Println("🤖 Generating report with OpenAI...")
	markdownReport, err := lt.llmClient.GenerateReport(data)
	if err != nil {
		return fmt.Errorf("report generation failed: %w", err)
	}

	log.Printf("✅ Markdown report generated (%d characters)", len(markdownReport))

	// Convert to HTML
	log.Println("🎨 Converting to HTML with charts...")
	htmlReport, err := lt.generator.GenerateHTML(markdownReport, data)
	if err != nil {
		return fmt.Errorf("HTML generation failed: %w", err)
	}

	// Create timestamped directory for this report
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	reportDir := filepath.Join("reports", timestamp)
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		return fmt.Errorf("failed to create report directory: %w", err)
	}

	// Save API data as JSON
	apiDataJSON, _ := json.MarshalIndent(data, "", "  ")
	apiDataPath := filepath.Join(reportDir, "01_api_data.json")
	if err := os.WriteFile(apiDataPath, apiDataJSON, 0644); err != nil {
		log.Printf("Failed to save API data: %v", err)
	}

	// Save system prompt
	systemPrompt := lt.llmClient.GetSystemPrompt()
	log.Printf("DEBUG: System prompt length: %d", len(systemPrompt))
	if len(systemPrompt) > 100 {
		log.Printf("DEBUG: System prompt preview: %.100s...", systemPrompt)
	} else {
		log.Printf("DEBUG: Full system prompt: %s", systemPrompt)
	}
	systemPromptPath := filepath.Join(reportDir, "llm_system_prompt.txt")
	log.Printf("DEBUG: Writing system prompt to: %s", systemPromptPath)
	if err := os.WriteFile(systemPromptPath, []byte(systemPrompt), 0644); err != nil {
		log.Printf("Failed to save system prompt: %v", err)
	} else {
		log.Printf("System prompt saved successfully to: %s", systemPromptPath)
	}

	// Generate LLM prompt and save it
	llmPrompt := lt.llmClient.BuildPrompt(data)
	promptPath := filepath.Join(reportDir, "02_llm_prompt.txt")
	if err := os.WriteFile(promptPath, []byte(llmPrompt), 0644); err != nil {
		log.Printf("Failed to save LLM prompt: %v", err)
	}

	// Save LLM response as markdown
	markdownPath := filepath.Join(reportDir, "03_llm_response.md")
	if err := os.WriteFile(markdownPath, []byte(markdownReport), 0644); err != nil {
		log.Printf("Failed to save markdown report: %v", err)
	}

	// Save final HTML report
	htmlPath := filepath.Join(reportDir, "04_final_report.html")
	if err := os.WriteFile(htmlPath, []byte(htmlReport), 0644); err != nil {
		log.Printf("Failed to save HTML report: %v", err)
	}

	duration := time.Since(startTime)
	log.Printf("🎉 Report generation completed in %v", duration)
	log.Printf("📁 Report directory: %s", reportDir)
	log.Printf("📄 Files saved:")
	log.Printf("   - API Data: %s", apiDataPath)
	log.Printf("   - System Prompt: %s", systemPromptPath)
	log.Printf("   - LLM Prompt: %s", promptPath)
	log.Printf("   - LLM Response: %s", markdownPath)
	log.Printf("   - Final Report: %s", htmlPath)
	log.Printf("🌐 Open in browser: file://%s/%s", mustGetWD(), htmlPath)

	// Print summary
	summary := map[string]interface{}{
		"status":         "success",
		"report_dir":     reportDir,
		"duration_ms":    duration.Milliseconds(),
		"report_size":    len(htmlReport),
		"markdown_size":  len(markdownReport),
		"timestamp":      data.Timestamp.Format(time.RFC3339),
		"data_summary": map[string]interface{}{
			"solar_flux":     data.SolarData.SolarFluxIndex,
			"k_index":        data.GeomagData.KIndex,
			"sunspot_number": data.SolarData.SunspotNumber,
			"activity_level": data.SolarData.SolarActivity,
		},
	}

	summaryJSON, _ := json.MarshalIndent(summary, "", "  ")
	log.Printf("📊 Generation Summary:\n%s", summaryJSON)

	return nil
}

func mustGetWD() string {
	wd, err := os.Getwd()
	if err != nil {
		return "/tmp"
	}
	return wd
}

func main() {
	openaiKey := os.Getenv("OPENAI_API_KEY")
	if openaiKey == "" {
		log.Fatal("❌ OPENAI_API_KEY environment variable is required")
	}

	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = "gpt-4o-mini" // Default model
	}

	tester := NewLocalTester(openaiKey, model)
	if err := tester.GenerateTestReport(); err != nil {
		log.Fatalf("❌ Test failed: %v", err)
	}
}
