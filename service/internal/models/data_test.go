package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestPropagationDataSerialization(t *testing.T) {
	// Create test data
	testTime := time.Date(2025, 9, 17, 12, 0, 0, 0, time.UTC)
	
	data := PropagationData{
		Timestamp: testTime,
		SolarData: SolarData{
			SolarFluxIndex:      150.5,
			SolarFluxDataSource: "NOAA",
			SunspotNumber:       120,
			SunspotDataSource:   "SIDC",
			SolarActivity:       "Moderate",
			FlareActivity:       "None",
			SolarCyclePhase:     "Rising",
			LastMajorFlare:      "2025-09-15 X1.2",
			SolarWindSpeed:      450.0,
			SolarWindDataSource: "NOAA",
			ProtonFlux:          1.5,
			ProtonFluxDataSource: "NOAA",
		},
		GeomagData: GeomagData{
			KIndex:                  3.5,
			KIndexDataSource:        "NOAA",
			AIndex:                  15.0,
			AIndexDataSource:        "N0NBH",
			GeomagActivity:          "Unsettled",
			GeomagConditions:        "Minor disturbances",
			MagneticField:           45000.0,
			MagneticFieldDataSource: "NOAA",
		},
		BandData: BandData{
			Band80m: BandCondition{Day: "Good", Night: "Excellent"},
			Band40m: BandCondition{Day: "Fair", Night: "Good"},
			Band20m: BandCondition{Day: "Excellent", Night: "Fair"},
			Band17m: BandCondition{Day: "Good", Night: "Poor"},
			Band15m: BandCondition{Day: "Excellent", Night: "Poor"},
			Band12m: BandCondition{Day: "Good", Night: "Poor"},
			Band10m: BandCondition{Day: "Fair", Night: "Poor"},
			Band6m:  BandCondition{Day: "Poor", Night: "Poor"},
			VHFPlus: BandCondition{Day: "Poor", Night: "Poor"},
			BandDataSource: "N0NBH",
		},
		Forecast: ForecastData{
			Today: DayForecast{
				Date:            testTime,
				KIndexForecast:  "3-4",
				SolarActivity:   "Moderate",
				HFConditions:    "Fair to Good",
				VHFConditions:   "Poor",
				BestBands:       []string{"20m", "17m", "15m"},
				WorstBands:      []string{"6m", "VHF+"},
			},
			Tomorrow: DayForecast{
				Date:            testTime.AddDate(0, 0, 1),
				KIndexForecast:  "2-3",
				SolarActivity:   "Low to Moderate",
				HFConditions:    "Good",
				VHFConditions:   "Poor",
				BestBands:       []string{"40m", "20m", "17m"},
				WorstBands:      []string{"10m", "6m"},
			},
			DayAfter: DayForecast{
				Date:            testTime.AddDate(0, 0, 2),
				KIndexForecast:  "1-2",
				SolarActivity:   "Low",
				HFConditions:    "Excellent",
				VHFConditions:   "Fair",
				BestBands:       []string{"80m", "40m", "20m"},
				WorstBands:      []string{},
			},
			Outlook:  "Improving conditions over next 3 days",
			Warnings: []string{"Minor geomagnetic storm possible"},
		},
		SourceEvents: []SourceEvent{
			{
				Source:      "NOAA",
				EventType:   "Solar Flare",
				Severity:    "Moderate",
				Description: "M-class flare detected",
				Timestamp:   testTime.Add(-2 * time.Hour),
				Impact:      "Minor HF radio blackout",
			},
		},
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal PropagationData to JSON: %v", err)
	}

	// Test JSON deserialization
	var unmarshaled PropagationData
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal PropagationData from JSON: %v", err)
	}

	// Verify key fields
	if !unmarshaled.Timestamp.Equal(data.Timestamp) {
		t.Errorf("Timestamp mismatch: expected %v, got %v", data.Timestamp, unmarshaled.Timestamp)
	}
	
	if unmarshaled.SolarData.SolarFluxIndex != data.SolarData.SolarFluxIndex {
		t.Errorf("SolarFluxIndex mismatch: expected %f, got %f", data.SolarData.SolarFluxIndex, unmarshaled.SolarData.SolarFluxIndex)
	}
	
	if unmarshaled.GeomagData.KIndex != data.GeomagData.KIndex {
		t.Errorf("KIndex mismatch: expected %f, got %f", data.GeomagData.KIndex, unmarshaled.GeomagData.KIndex)
	}
	
	if unmarshaled.BandData.Band20m.Day != data.BandData.Band20m.Day {
		t.Errorf("Band20m Day condition mismatch: expected %s, got %s", data.BandData.Band20m.Day, unmarshaled.BandData.Band20m.Day)
	}
}

func TestBandConditionValidation(t *testing.T) {
	validConditions := []string{"Poor", "Fair", "Good", "Excellent"}
	
	for _, condition := range validConditions {
		bc := BandCondition{
			Day:   condition,
			Night: condition,
		}
		
		// Test that valid conditions can be created
		if bc.Day != condition || bc.Night != condition {
			t.Errorf("BandCondition creation failed for %s", condition)
		}
	}
}

func TestSourceEventValidation(t *testing.T) {
	testTime := time.Now()
	
	event := SourceEvent{
		Source:      "NOAA",
		EventType:   "Solar Flare",
		Severity:    "", // LLM-generated field - should be empty from normalizer
		Description: "X-class flare detected at 12:00 UTC",
		Timestamp:   testTime,
		Impact:      "", // LLM-generated field - should be empty from normalizer
	}
	
	// Verify all fields are set correctly
	if event.Source != "NOAA" {
		t.Errorf("Expected Source 'NOAA', got '%s'", event.Source)
	}
	if event.EventType != "Solar Flare" {
		t.Errorf("Expected EventType 'Solar Flare', got '%s'", event.EventType)
	}
	if event.Severity != "" {
		t.Errorf("Expected empty Severity (LLM-generated), got '%s'", event.Severity)
	}
	if event.Description != "X-class flare detected at 12:00 UTC" {
		t.Errorf("Expected Description 'X-class flare detected at 12:00 UTC', got '%s'", event.Description)
	}
	if event.Impact != "" {
		t.Errorf("Expected empty Impact (LLM-generated), got '%s'", event.Impact)
	}
	if !event.Timestamp.Equal(testTime) {
		t.Errorf("Timestamp mismatch: expected %v, got %v", testTime, event.Timestamp)
	}
}

func TestDayForecastValidation(t *testing.T) {
	forecast := DayForecast{
		Date:            time.Time{}, // LLM-generated field - should be empty from normalizer
		KIndexForecast:  "", // LLM-generated field - should be empty from normalizer
		SolarActivity:   "", // LLM-generated field - should be empty from normalizer
		HFConditions:    "", // LLM-generated field - should be empty from normalizer
		VHFConditions:   "", // LLM-generated field - should be empty from normalizer
		BestBands:       []string{"20m", "40m"},
		WorstBands:      []string{"6m"},
	}
	
	// Test LLM-generated fields are empty
	if !forecast.Date.IsZero() {
		t.Errorf("Expected empty Date (LLM-generated), got %v", forecast.Date)
	}
	if forecast.KIndexForecast != "" {
		t.Errorf("Expected empty KIndexForecast (LLM-generated), got '%s'", forecast.KIndexForecast)
	}
	if forecast.SolarActivity != "" {
		t.Errorf("Expected empty SolarActivity (LLM-generated), got '%s'", forecast.SolarActivity)
	}
	if forecast.HFConditions != "" {
		t.Errorf("Expected empty HFConditions (LLM-generated), got '%s'", forecast.HFConditions)
	}
	if forecast.VHFConditions != "" {
		t.Errorf("Expected empty VHFConditions (LLM-generated), got '%s'", forecast.VHFConditions)
	}
	
	// Test that arrays are properly handled
	if len(forecast.BestBands) != 2 {
		t.Errorf("Expected 2 best bands, got %d", len(forecast.BestBands))
	}
	if len(forecast.WorstBands) != 1 {
		t.Errorf("Expected 1 worst band, got %d", len(forecast.WorstBands))
	}
	
	// Test band values
	if forecast.BestBands[0] != "20m" || forecast.BestBands[1] != "40m" {
		t.Errorf("BestBands values incorrect: %v", forecast.BestBands)
	}
	if forecast.WorstBands[0] != "6m" {
		t.Errorf("WorstBands values incorrect: %v", forecast.WorstBands)
	}
}

func TestSourceDataStructure(t *testing.T) {
	// Test that SourceData can hold various data types
	sourceData := SourceData{
		NOAAKIndex: []NOAAKIndexResponse{},
		NOAASolar:  []NOAASolarResponse{},
		N0NBH:      nil, // Can be nil
		SIDC:       nil, // Can be nil
	}
	
	// Test JSON serialization of SourceData
	jsonData, err := json.Marshal(sourceData)
	if err != nil {
		t.Fatalf("Failed to marshal SourceData: %v", err)
	}
	
	var unmarshaled SourceData
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal SourceData: %v", err)
	}
	
	// Verify nil fields remain nil
	if unmarshaled.N0NBH != nil {
		t.Error("Expected N0NBH to remain nil after serialization")
	}
	if unmarshaled.SIDC != nil {
		t.Error("Expected SIDC to remain nil after serialization")
	}
}

func TestForecastDataWithEmptyArrays(t *testing.T) {
	forecast := ForecastData{
		Today:    DayForecast{BestBands: []string{}, WorstBands: []string{}},
		Tomorrow: DayForecast{BestBands: []string{}, WorstBands: []string{}},
		DayAfter: DayForecast{BestBands: []string{}, WorstBands: []string{}},
		Outlook:  "",
		Warnings: []string{},
	}
	
	// Test JSON serialization with empty arrays
	jsonData, err := json.Marshal(forecast)
	if err != nil {
		t.Fatalf("Failed to marshal ForecastData with empty arrays: %v", err)
	}
	
	var unmarshaled ForecastData
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal ForecastData: %v", err)
	}
	
	// Verify empty arrays are preserved
	if unmarshaled.Today.BestBands == nil {
		t.Error("BestBands should be empty array, not nil")
	}
	if len(unmarshaled.Today.BestBands) != 0 {
		t.Error("BestBands should be empty")
	}
	if len(unmarshaled.Warnings) != 0 {
		t.Error("Warnings should be empty")
	}
}

func TestComplexPropagationDataScenario(t *testing.T) {
	// Test a complex scenario with all fields populated
	testTime := time.Date(2025, 9, 17, 15, 30, 0, 0, time.UTC)
	
	data := PropagationData{
		Timestamp: testTime,
		SolarData: SolarData{
			SolarFluxIndex:       200.0,
			SolarFluxDataSource:  "NOAA",
			SunspotNumber:        150,
			SunspotDataSource:    "SIDC",
			SolarActivity:        "High",
			FlareActivity:        "X1.5 in progress",
			SolarCyclePhase:      "Solar Maximum",
			LastMajorFlare:       "2025-09-17 X1.5",
			SolarWindSpeed:       600.0,
			SolarWindDataSource:  "NOAA",
			ProtonFlux:           10.0,
			ProtonFluxDataSource: "NOAA",
		},
		GeomagData: GeomagData{
			KIndex:                  7.0,
			KIndexDataSource:        "NOAA",
			AIndex:                  50.0,
			AIndexDataSource:        "N0NBH",
			GeomagActivity:          "Major Storm",
			GeomagConditions:        "Severe geomagnetic disturbances",
			MagneticField:           55000.0,
			MagneticFieldDataSource: "NOAA",
		},
		BandData: BandData{
			Band80m:        BandCondition{Day: "Excellent", Night: "Excellent"},
			Band40m:        BandCondition{Day: "Good", Night: "Excellent"},
			Band20m:        BandCondition{Day: "Poor", Night: "Fair"},
			Band17m:        BandCondition{Day: "Poor", Night: "Poor"},
			Band15m:        BandCondition{Day: "Poor", Night: "Poor"},
			Band12m:        BandCondition{Day: "Poor", Night: "Poor"},
			Band10m:        BandCondition{Day: "Poor", Night: "Poor"},
			Band6m:         BandCondition{Day: "Poor", Night: "Poor"},
			VHFPlus:        BandCondition{Day: "Poor", Night: "Poor"},
			BandDataSource: "N0NBH",
		},
		Forecast: ForecastData{
			Today: DayForecast{
				Date:           testTime,
				KIndexForecast: "6-8",
				SolarActivity:  "Very High",
				HFConditions:   "Poor",
				VHFConditions:  "Poor",
				BestBands:      []string{"80m", "40m"},
				WorstBands:     []string{"20m", "17m", "15m", "12m", "10m", "6m"},
			},
			Tomorrow: DayForecast{
				Date:           testTime.AddDate(0, 0, 1),
				KIndexForecast: "4-6",
				SolarActivity:  "High",
				HFConditions:   "Poor to Fair",
				VHFConditions:  "Poor",
				BestBands:      []string{"80m", "40m"},
				WorstBands:     []string{"15m", "12m", "10m", "6m"},
			},
			DayAfter: DayForecast{
				Date:           testTime.AddDate(0, 0, 2),
				KIndexForecast: "2-4",
				SolarActivity:  "Moderate",
				HFConditions:   "Fair to Good",
				VHFConditions:  "Poor to Fair",
				BestBands:      []string{"80m", "40m", "20m"},
				WorstBands:     []string{"10m", "6m"},
			},
			Outlook: "Severe geomagnetic storm in progress. Conditions expected to improve over 48-72 hours.",
			Warnings: []string{
				"Strong HF radio blackout in progress",
				"GPS navigation may be affected",
				"Aurora visible at lower latitudes",
			},
		},
		SourceEvents: []SourceEvent{
			{
				Source:      "NOAA",
				EventType:   "Solar Flare",
				Severity:    "Extreme",
				Description: "X1.5 solar flare detected at 15:25 UTC",
				Timestamp:   testTime.Add(-5 * time.Minute),
				Impact:      "Strong HF radio blackout on sunlit side of Earth",
			},
			{
				Source:      "NOAA",
				EventType:   "Geomagnetic Storm",
				Severity:    "High",
				Description: "G3 geomagnetic storm conditions observed",
				Timestamp:   testTime.Add(-2 * time.Hour),
				Impact:      "HF radio propagation severely degraded",
			},
		},
	}
	
	// Serialize and deserialize
	jsonData, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal complex PropagationData: %v", err)
	}
	
	var unmarshaled PropagationData
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal complex PropagationData: %v", err)
	}
	
	// Verify complex data integrity
	if len(unmarshaled.SourceEvents) != 2 {
		t.Errorf("Expected 2 source events, got %d", len(unmarshaled.SourceEvents))
	}
	
	if len(unmarshaled.Forecast.Warnings) != 3 {
		t.Errorf("Expected 3 warnings, got %d", len(unmarshaled.Forecast.Warnings))
	}
	
	if unmarshaled.GeomagData.KIndex != 7.0 {
		t.Errorf("Expected K-index 7.0, got %f", unmarshaled.GeomagData.KIndex)
	}
	
	if unmarshaled.SolarData.SolarActivity != "High" {
		t.Errorf("Expected High solar activity, got %s", unmarshaled.SolarData.SolarActivity)
	}
}
