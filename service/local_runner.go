package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
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
		generator: reports.NewGenerator(),
	}
}

func (lt *LocalTester) GenerateTestReport() error {
	ctx := context.Background()
	startTime := time.Now()

	log.Println("🚀 Starting local report generation test...")

	// Use default config URLs
	cfg := &config.Config{
		NOAAKIndexURL: "https://services.swpc.noaa.gov/products/noaa-planetary-k-index.json",
		NOAASolarURL:  "https://services.swpc.noaa.gov/products/solar-wind/plasma-7-day.json",
		N0NBHSolarURL: "https://www.hamqsl.com/solarxml.php",
		SIDCRSSURL:    "https://www.sidc.be/silso/INFO/snmtotcsv.php",
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

	// Save to local file
	filename := fmt.Sprintf("test_report_%s.html", time.Now().Format("2006-01-02_15-04-05"))
	if err := os.WriteFile(filename, []byte(htmlReport), 0644); err != nil {
		return fmt.Errorf("failed to save report: %w", err)
	}

	duration := time.Since(startTime)
	log.Printf("🎉 Report generation completed in %v", duration)
	log.Printf("📄 Report saved to: %s", filename)
	log.Printf("🌐 Open in browser: file://%s/%s", mustGetWD(), filename)

	// Print summary
	summary := map[string]interface{}{
		"status":         "success",
		"filename":       filename,
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

// Run local test if called directly
func runLocalTest() {
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
