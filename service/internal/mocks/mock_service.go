package mocks

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"radiocast/internal/models"
)

// MockService handles loading mock data for testing
type MockService struct {
	mocksDir string
}

// NewMockService creates a new mock service
func NewMockService(mocksDir string) *MockService {
	return &MockService{
		mocksDir: filepath.Join(mocksDir, "data"),
	}
}

// LoadMockData loads all mock data from files
func (m *MockService) LoadMockData() (*models.PropagationData, *models.SourceData, error) {
	// Load normalized data
	propagationData, err := m.loadNormalizedData()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load normalized data: %w", err)
	}

	// Load source data
	sourceData, err := m.loadSourceData()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load source data: %w", err)
	}

	return propagationData, sourceData, nil
}

// LoadMockLLMResponse loads the mock LLM response
func (m *MockService) LoadMockLLMResponse() (string, error) {
	filePath := filepath.Join(m.mocksDir, "llm_response.md")
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read mock LLM response: %w", err)
	}
	return string(content), nil
}

// LoadMockSunGif loads the mock Sun GIF
func (m *MockService) LoadMockSunGif() ([]byte, error) {
	filePath := filepath.Join(m.mocksDir, "sun_72h.gif")
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read mock Sun GIF: %w", err)
	}
	return content, nil
}

// loadNormalizedData loads the normalized propagation data
func (m *MockService) loadNormalizedData() (*models.PropagationData, error) {
	filePath := filepath.Join(m.mocksDir, "normalized_data.json")
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open normalized data file: %w", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read normalized data file: %w", err)
	}

	var data models.PropagationData
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal normalized data: %w", err)
	}

	// Update timestamp to current time for fresh reports
	data.Timestamp = time.Now()

	return &data, nil
}

// loadSourceData loads the raw source data
func (m *MockService) loadSourceData() (*models.SourceData, error) {
	sourceData := &models.SourceData{}

	// Load NOAA K-Index data
	var noaaKIndexData []models.NOAAKIndexResponse
	_, err := m.loadTypedJSONFile("noaa_k_index.json", &noaaKIndexData)
	if err != nil {
		return nil, fmt.Errorf("failed to load NOAA K-Index data: %w", err)
	}
	sourceData.NOAAKIndex = noaaKIndexData

	// Load NOAA Solar data
	var noaaSolarData []models.NOAASolarResponse
	_, err = m.loadTypedJSONFile("noaa_solar.json", &noaaSolarData)
	if err != nil {
		return nil, fmt.Errorf("failed to load NOAA Solar data: %w", err)
	}
	sourceData.NOAASolar = noaaSolarData

	// Load N0NBH data
	var n0nbhData models.N0NBHResponse
	_, err = m.loadTypedJSONFile("n0nbh_data.json", &n0nbhData)
	if err != nil {
		return nil, fmt.Errorf("failed to load N0NBH data: %w", err)
	}
	sourceData.N0NBH = &n0nbhData

	// SIDC data is complex RSS feed data, so we'll set it to nil for mock purposes
	// since the report generation doesn't strictly depend on the exact SIDC structure
	sourceData.SIDC = nil

	return sourceData, nil
}

// loadJSONFile loads a JSON file and returns the raw data
func (m *MockService) loadJSONFile(filename string) (interface{}, error) {
	filePath := filepath.Join(m.mocksDir, filename)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filename, err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	var data interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal file %s: %w", filename, err)
	}

	return data, nil
}

// loadTypedJSONFile loads a JSON file and unmarshals it into the provided type
func (m *MockService) loadTypedJSONFile(filename string, target interface{}) (interface{}, error) {
	filePath := filepath.Join(m.mocksDir, filename)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filename, err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	if err := json.Unmarshal(content, target); err != nil {
		return nil, fmt.Errorf("failed to unmarshal file %s: %w", filename, err)
	}

	return target, nil
}
