package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"radiocast/service/internal/config"
	"radiocast/service/internal/fetchers"
	"radiocast/service/internal/llm"
)

func main() {
	openaiKey := os.Getenv("OPENAI_API_KEY")
	if openaiKey == "" {
		log.Fatal("❌ OPENAI_API_KEY environment variable is required")
	}

	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = "gpt-4o-mini"
	}

	log.Println("🚀 Testing LLM report generation...")

	// Use default config URLs
	cfg := &config.Config{
		NOAAKIndexURL: "https://services.swpc.noaa.gov/products/noaa-planetary-k-index.json",
		NOAASolarURL:  "https://services.swpc.noaa.gov/products/solar-wind/plasma-7-day.json",
		N0NBHSolarURL: "https://www.hamqsl.com/solarxml.php",
		SIDCRSSURL:    "https://www.sidc.be/silso/INFO/snmtotcsv.php",
	}

	// Fetch data
	fetcher := fetchers.NewDataFetcher()
	data, err := fetcher.FetchAllData(
		context.Background(),
		cfg.NOAAKIndexURL,
		cfg.NOAASolarURL,
		cfg.N0NBHSolarURL,
		cfg.SIDCRSSURL,
	)
	if err != nil {
		log.Fatalf("Data fetch failed: %v", err)
	}

	log.Printf("✅ Data fetched - Solar Flux: %.1f, K-Index: %.1f", 
		data.SolarData.SolarFluxIndex, data.GeomagData.KIndex)

	// Generate report with LLM
	llmClient := llm.NewOpenAIClient(openaiKey, model)
	markdownReport, err := llmClient.GenerateReport(data)
	if err != nil {
		log.Fatalf("Report generation failed: %v", err)
	}

	log.Printf("✅ Report generated (%d characters)", len(markdownReport))

	// Check for Chart Data section
	if strings.Contains(markdownReport, "## Chart Data") {
		log.Println("✅ Chart Data section found!")
		
		// Extract the Chart Data section
		parts := strings.Split(markdownReport, "## Chart Data")
		if len(parts) > 1 {
			chartSection := parts[1]
			if strings.Contains(chartSection, "```json") {
				log.Println("✅ JSON block found in Chart Data section!")
				
				// Extract JSON content
				jsonStart := strings.Index(chartSection, "```json") + 7
				jsonEnd := strings.Index(chartSection[jsonStart:], "```")
				if jsonEnd > 0 {
					jsonContent := chartSection[jsonStart : jsonStart+jsonEnd]
					log.Printf("📊 Chart JSON content:\n%s", strings.TrimSpace(jsonContent))
				}
			} else {
				log.Println("❌ No JSON block found in Chart Data section")
			}
		}
	} else {
		log.Println("❌ Chart Data section NOT found!")
		log.Println("📄 Report content preview:")
		lines := strings.Split(markdownReport, "\n")
		for i, line := range lines {
			if i < 20 { // Show first 20 lines
				log.Printf("  %2d: %s", i+1, line)
			}
		}
		if len(lines) > 20 {
			log.Printf("  ... (%d more lines)", len(lines)-20)
		}
	}

	// Check for Band-by-Band Analysis table
	if strings.Contains(markdownReport, "| Band |") {
		log.Println("✅ Band-by-Band Analysis table found!")
	} else {
		log.Println("❌ Band-by-Band Analysis table NOT found!")
	}

	// Save full report for inspection
	filename := "debug_report.md"
	if err := os.WriteFile(filename, []byte(markdownReport), 0644); err != nil {
		log.Printf("Warning: Could not save report: %v", err)
	} else {
		log.Printf("📄 Full report saved to: %s", filename)
	}
}
