package models

import (
	"encoding/json"
	"encoding/xml"
	"testing"
)

func TestNOAAKIndexAPIResponseSerialization(t *testing.T) {
	response := NOAAKIndexAPIResponse{
		TimeTag:     "2025-09-17T12:00:00",
		KpIndex:     3,
		EstimatedKp: 3.33,
		Kp:          "3o",
	}
	
	// Test JSON serialization
	jsonData, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal NOAAKIndexAPIResponse: %v", err)
	}
	
	var unmarshaled NOAAKIndexAPIResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal NOAAKIndexAPIResponse: %v", err)
	}
	
	// Verify fields
	if unmarshaled.TimeTag != response.TimeTag {
		t.Errorf("TimeTag mismatch: expected %s, got %s", response.TimeTag, unmarshaled.TimeTag)
	}
	if unmarshaled.KpIndex != response.KpIndex {
		t.Errorf("KpIndex mismatch: expected %d, got %d", response.KpIndex, unmarshaled.KpIndex)
	}
	if unmarshaled.EstimatedKp != response.EstimatedKp {
		t.Errorf("EstimatedKp mismatch: expected %f, got %f", response.EstimatedKp, unmarshaled.EstimatedKp)
	}
	if unmarshaled.Kp != response.Kp {
		t.Errorf("Kp mismatch: expected %s, got %s", response.Kp, unmarshaled.Kp)
	}
}

func TestNOAASolarAPIResponseSerialization(t *testing.T) {
	response := NOAASolarAPIResponse{
		TimeTag:         "2025-08",
		SSN:             133.5,
		SmoothedSSN:     125.0,
		ObservedSWPCSSN: 130.0,
		SmoothedSWPCSSN: 128.0,
		F107:            150.0,
		SmoothedF107:    145.0,
	}
	
	// Test JSON serialization
	jsonData, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal NOAASolarAPIResponse: %v", err)
	}
	
	var unmarshaled NOAASolarAPIResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal NOAASolarAPIResponse: %v", err)
	}
	
	// Verify fields
	if unmarshaled.TimeTag != response.TimeTag {
		t.Errorf("TimeTag mismatch: expected %s, got %s", response.TimeTag, unmarshaled.TimeTag)
	}
	if unmarshaled.SSN != response.SSN {
		t.Errorf("SSN mismatch: expected %f, got %f", response.SSN, unmarshaled.SSN)
	}
	if unmarshaled.F107 != response.F107 {
		t.Errorf("F107 mismatch: expected %f, got %f", response.F107, unmarshaled.F107)
	}
}

func TestN0NBHBandConditionSerialization(t *testing.T) {
	condition := N0NBHBandCondition{
		Name:   "20m",
		Time:   "Day",
		Day:    "Good",
		Night:  "Fair",
		Source: "N0NBH",
	}
	
	// Test JSON serialization
	jsonData, err := json.Marshal(condition)
	if err != nil {
		t.Fatalf("Failed to marshal N0NBHBandCondition: %v", err)
	}
	
	var unmarshaled N0NBHBandCondition
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal N0NBHBandCondition: %v", err)
	}
	
	// Verify all fields
	if unmarshaled.Name != condition.Name {
		t.Errorf("Name mismatch: expected %s, got %s", condition.Name, unmarshaled.Name)
	}
	if unmarshaled.Time != condition.Time {
		t.Errorf("Time mismatch: expected %s, got %s", condition.Time, unmarshaled.Time)
	}
	if unmarshaled.Day != condition.Day {
		t.Errorf("Day mismatch: expected %s, got %s", condition.Day, unmarshaled.Day)
	}
	if unmarshaled.Night != condition.Night {
		t.Errorf("Night mismatch: expected %s, got %s", condition.Night, unmarshaled.Night)
	}
	if unmarshaled.Source != condition.Source {
		t.Errorf("Source mismatch: expected %s, got %s", condition.Source, unmarshaled.Source)
	}
}

func TestN0NBHResponseSerialization(t *testing.T) {
	response := N0NBHResponse{
		Time:   "Sep 17 2025 12:00:00 GMT",
		Source: "N0NBH",
	}
	
	// Set solar data
	response.SolarData.SolarFlux = "150"
	response.SolarData.AIndex = "15"
	response.SolarData.KIndex = "3"
	response.SolarData.KIndexNT = "3.33"
	response.SolarData.SunSpots = "120"
	response.SolarData.HeliumLine = "1083"
	response.SolarData.ProtonFlux = "1.5"
	response.SolarData.ElectronFlux = "Normal"
	response.SolarData.Aurora = "No"
	response.SolarData.NormalizationTime = "12:00"
	response.SolarData.LatestSWPCReport = "All quiet"
	
	// Set band conditions
	response.Calculatedconditions.Band = []struct {
		Name   string `json:"name"`
		Time   string `json:"time"`
		Day    string `json:"day"`
		Night  string `json:"night"`
		Source string `json:"source"`
	}{
		{Name: "80m", Time: "Day", Day: "Good", Night: "Excellent", Source: "N0NBH"},
		{Name: "40m", Time: "Day", Day: "Fair", Night: "Good", Source: "N0NBH"},
		{Name: "20m", Time: "Day", Day: "Excellent", Night: "Fair", Source: "N0NBH"},
	}
	
	// Set VHF conditions
	response.CalculatedVHFConditions.Phenomenon = []struct {
		Name     string `json:"name"`
		Location string `json:"location"`
	}{
		{Name: "Aurora", Location: "High Latitudes"},
		{Name: "Sporadic E", Location: "Mid Latitudes"},
	}
	
	// Test JSON serialization
	jsonData, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal N0NBHResponse: %v", err)
	}
	
	var unmarshaled N0NBHResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal N0NBHResponse: %v", err)
	}
	
	// Verify basic fields
	if unmarshaled.Time != response.Time {
		t.Errorf("Time mismatch: expected %s, got %s", response.Time, unmarshaled.Time)
	}
	if unmarshaled.Source != response.Source {
		t.Errorf("Source mismatch: expected %s, got %s", response.Source, unmarshaled.Source)
	}
	
	// Verify solar data
	if unmarshaled.SolarData.SolarFlux != response.SolarData.SolarFlux {
		t.Errorf("SolarFlux mismatch: expected %s, got %s", response.SolarData.SolarFlux, unmarshaled.SolarData.SolarFlux)
	}
	if unmarshaled.SolarData.KIndex != response.SolarData.KIndex {
		t.Errorf("KIndex mismatch: expected %s, got %s", response.SolarData.KIndex, unmarshaled.SolarData.KIndex)
	}
	if unmarshaled.SolarData.SunSpots != response.SolarData.SunSpots {
		t.Errorf("SunSpots mismatch: expected %s, got %s", response.SolarData.SunSpots, unmarshaled.SolarData.SunSpots)
	}
	
	// Verify band conditions
	if len(unmarshaled.Calculatedconditions.Band) != 3 {
		t.Errorf("Expected 3 band conditions, got %d", len(unmarshaled.Calculatedconditions.Band))
	}
	
	if len(unmarshaled.Calculatedconditions.Band) > 0 {
		firstBand := unmarshaled.Calculatedconditions.Band[0]
		if firstBand.Name != "80m" {
			t.Errorf("First band name mismatch: expected 80m, got %s", firstBand.Name)
		}
		if firstBand.Day != "Good" {
			t.Errorf("First band day condition mismatch: expected Good, got %s", firstBand.Day)
		}
	}
	
	// Verify VHF conditions
	if len(unmarshaled.CalculatedVHFConditions.Phenomenon) != 2 {
		t.Errorf("Expected 2 VHF phenomena, got %d", len(unmarshaled.CalculatedVHFConditions.Phenomenon))
	}
}

func TestN0NBHXMLResponseSerialization(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<solar>
    <solardata>
        <source>N0NBH</source>
        <updated>Sep 17 2025 12:00:00 GMT</updated>
        <solarflux>150</solarflux>
        <aindex>15</aindex>
        <kindex>3</kindex>
        <kindexnt>3.33</kindexnt>
        <xray>A1.0</xray>
        <sunspots>120</sunspots>
        <heliumline>1083</heliumline>
        <protonflux>1.5</protonflux>
        <electonflux>Normal</electonflux>
        <aurora>No</aurora>
        <normalization>12:00</normalization>
        <latdegree>45</latdegree>
        <solarwind>450</solarwind>
        <magneticfield>45000</magneticfield>
        <calculatedconditions>
            <band name="80m" time="Day">Good</band>
            <band name="40m" time="Day">Fair</band>
            <band name="20m" time="Day">Excellent</band>
        </calculatedconditions>
    </solardata>
    <time>Sep 17 2025 12:00:00 GMT</time>
</solar>`
	
	var response N0NBHXMLResponse
	err := xml.Unmarshal([]byte(xmlData), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal N0NBHXMLResponse: %v", err)
	}
	
	// Verify basic structure
	if response.Time != "Sep 17 2025 12:00:00 GMT" {
		t.Errorf("Time mismatch: expected 'Sep 17 2025 12:00:00 GMT', got %s", response.Time)
	}
	
	// Verify solar data
	if response.SolarData.Source != "N0NBH" {
		t.Errorf("Source mismatch: expected N0NBH, got %s", response.SolarData.Source)
	}
	if response.SolarData.SolarFlux != "150" {
		t.Errorf("SolarFlux mismatch: expected 150, got %s", response.SolarData.SolarFlux)
	}
	if response.SolarData.KIndex != "3" {
		t.Errorf("KIndex mismatch: expected 3, got %s", response.SolarData.KIndex)
	}
	if response.SolarData.SunSpots != "120" {
		t.Errorf("SunSpots mismatch: expected 120, got %s", response.SolarData.SunSpots)
	}
	
	// Verify calculated conditions
	if len(response.SolarData.CalculatedConditions.Band) != 3 {
		t.Errorf("Expected 3 band conditions, got %d", len(response.SolarData.CalculatedConditions.Band))
	}
	
	if len(response.SolarData.CalculatedConditions.Band) > 0 {
		firstBand := response.SolarData.CalculatedConditions.Band[0]
		if firstBand.Name != "80m" {
			t.Errorf("First band name mismatch: expected 80m, got %s", firstBand.Name)
		}
		if firstBand.Time != "Day" {
			t.Errorf("First band time mismatch: expected Day, got %s", firstBand.Time)
		}
		if firstBand.Condition != "Good" {
			t.Errorf("First band condition mismatch: expected Good, got %s", firstBand.Condition)
		}
	}
}

func TestN0NBHResponseEmptyConditions(t *testing.T) {
	// Test with empty band conditions
	response := N0NBHResponse{
		Time:   "Sep 17 2025 12:00:00 GMT",
		Source: "N0NBH",
	}
	
	// Empty band conditions
	response.Calculatedconditions.Band = []struct {
		Name   string `json:"name"`
		Time   string `json:"time"`
		Day    string `json:"day"`
		Night  string `json:"night"`
		Source string `json:"source"`
	}{}
	
	// Empty VHF conditions
	response.CalculatedVHFConditions.Phenomenon = []struct {
		Name     string `json:"name"`
		Location string `json:"location"`
	}{}
	
	// Test JSON serialization
	jsonData, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal N0NBHResponse with empty conditions: %v", err)
	}
	
	var unmarshaled N0NBHResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal N0NBHResponse: %v", err)
	}
	
	// Verify empty arrays are preserved
	if len(unmarshaled.Calculatedconditions.Band) != 0 {
		t.Errorf("Expected empty band conditions, got %d", len(unmarshaled.Calculatedconditions.Band))
	}
	if len(unmarshaled.CalculatedVHFConditions.Phenomenon) != 0 {
		t.Errorf("Expected empty VHF conditions, got %d", len(unmarshaled.CalculatedVHFConditions.Phenomenon))
	}
}

func TestNOAAResponsesWithSpecialCharacters(t *testing.T) {
	// Test NOAA responses with special characters and edge cases
	kResponse := NOAAKIndexAPIResponse{
		TimeTag:     "2025-09-17T12:00:00.000Z",
		KpIndex:     9,
		EstimatedKp: 8.67,
		Kp:          "8+",
	}
	
	solarResponse := NOAASolarAPIResponse{
		TimeTag:         "2025-08",
		SSN:             0.0,
		SmoothedSSN:     0.0,
		ObservedSWPCSSN: 0.0,
		SmoothedSWPCSSN: 0.0,
		F107:            67.0, // Minimum expected value
		SmoothedF107:    67.0,
	}
	
	// Test serialization
	kJsonData, err := json.Marshal(kResponse)
	if err != nil {
		t.Fatalf("Failed to marshal K-index response with special characters: %v", err)
	}
	
	sJsonData, err := json.Marshal(solarResponse)
	if err != nil {
		t.Fatalf("Failed to marshal solar response with zero values: %v", err)
	}
	
	// Test deserialization
	var kUnmarshaled NOAAKIndexAPIResponse
	err = json.Unmarshal(kJsonData, &kUnmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal K-index response: %v", err)
	}
	
	var sUnmarshaled NOAASolarAPIResponse
	err = json.Unmarshal(sJsonData, &sUnmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal solar response: %v", err)
	}
	
	// Verify special characters and edge values
	if kUnmarshaled.Kp != "8+" {
		t.Errorf("Kp special character mismatch: expected '8+', got '%s'", kUnmarshaled.Kp)
	}
	if sUnmarshaled.SSN != 0.0 {
		t.Errorf("Zero SSN mismatch: expected 0.0, got %f", sUnmarshaled.SSN)
	}
	if sUnmarshaled.F107 != 67.0 {
		t.Errorf("Minimum F10.7 mismatch: expected 67.0, got %f", sUnmarshaled.F107)
	}
}
