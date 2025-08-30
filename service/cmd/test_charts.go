package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"radiocast/internal/models"
	"radiocast/internal/reports"
)

func main() {
	// Create test data
	testData := &models.PropagationData{
		Timestamp: time.Now(),
		SolarData: models.SolarData{
			SolarFluxIndex: 150.5,
			SunspotNumber:  85,
			SolarActivity:  "Moderate",
		},
		GeomagData: models.GeomagData{
			KIndex:         2.3,
			AIndex:         15.0,
			GeomagActivity: "Quiet",
		},
		BandData: models.BandData{
			Band80m: models.BandCondition{Day: "Fair", Night: "Good"},
			Band40m: models.BandCondition{Day: "Good", Night: "Excellent"},
			Band20m: models.BandCondition{Day: "Excellent", Night: "Good"},
			Band17m: models.BandCondition{Day: "Good", Night: "Fair"},
			Band15m: models.BandCondition{Day: "Excellent", Night: "Poor"},
			Band12m: models.BandCondition{Day: "Good", Night: "Poor"},
			Band10m: models.BandCondition{Day: "Fair", Night: "Closed"},
			Band6m:  models.BandCondition{Day: "Poor", Night: "Closed"},
			VHFPlus: models.BandCondition{Day: "Fair", Night: "Poor"},
		},
		Forecast: models.ForecastData{
			Today: models.DayForecast{
				Date:           time.Now(),
				KIndexForecast: "2-3",
				SolarActivity:  "Moderate",
				HFConditions:   "Good",
			},
			Tomorrow: models.DayForecast{
				Date:           time.Now().Add(24 * time.Hour),
				KIndexForecast: "1-2",
				SolarActivity:  "Low",
				HFConditions:   "Excellent",
			},
			DayAfter: models.DayForecast{
				Date:           time.Now().Add(48 * time.Hour),
				KIndexForecast: "2-4",
				SolarActivity:  "Moderate",
				HFConditions:   "Fair",
			},
		},
	}

	// Create test output directory
	testDir := "test_charts_output"
	if err := os.MkdirAll(testDir, 0755); err != nil {
		log.Fatalf("Failed to create test directory: %v", err)
	}

	log.Printf("🧪 Testing chart generation...")
	log.Printf("📁 Output directory: %s", testDir)

	// Create chart generator
	chartGen := reports.NewChartGenerator(testDir)

	// Generate charts
	chartFiles, err := chartGen.GenerateCharts(testData)
	if err != nil {
		log.Fatalf("❌ Chart generation failed: %v", err)
	}

	log.Printf("✅ Successfully generated %d charts:", len(chartFiles))
	for _, file := range chartFiles {
		fullPath := filepath.Join(testDir, file)
		if stat, err := os.Stat(fullPath); err == nil {
			log.Printf("   📊 %s (%d bytes)", file, stat.Size())
		} else {
			log.Printf("   ❌ %s (file not found)", file)
		}
	}

	// Test HTML generation
	generator := reports.NewGenerator(testDir)
	testMarkdown := `# Test Radio Propagation Report

## Current Conditions
- Solar Flux: 150.5
- K-index: 2.3
- Sunspot Number: 85

## Band Conditions
Good conditions on 20m and 40m bands.

## Forecast
Stable conditions expected for the next 3 days.
`

	log.Printf("🎨 Testing HTML generation with charts...")
	html, err := generator.GenerateHTML(testMarkdown, testData)
	if err != nil {
		log.Fatalf("❌ HTML generation failed: %v", err)
	}

	// Save test HTML
	htmlPath := filepath.Join(testDir, "test_report.html")
	if err := os.WriteFile(htmlPath, []byte(html), 0644); err != nil {
		log.Printf("Failed to save test HTML: %v", err)
	} else {
		log.Printf("✅ Test HTML report saved: %s", htmlPath)
	}

	log.Printf("🎉 Chart generation test completed successfully!")
	log.Printf("📂 Check the '%s' directory for generated files", testDir)

	// Print file listing
	files, _ := os.ReadDir(testDir)
	fmt.Println("\n📋 Generated files:")
	for _, file := range files {
		if info, err := file.Info(); err == nil {
			fmt.Printf("  - %s (%d bytes)\n", file.Name(), info.Size())
		}
	}
}
