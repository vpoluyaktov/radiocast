package config

import (
	"context"
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		expectError bool
		validate    func(*Config) error
	}{
		{
			name: "valid config with required fields",
			envVars: map[string]string{
				"OPENAI_API_KEY": "test-key",
			},
			expectError: false,
			validate: func(cfg *Config) error {
				if cfg.OpenAIAPIKey != "test-key" {
					t.Errorf("Expected OpenAIAPIKey to be 'test-key', got '%s'", cfg.OpenAIAPIKey)
				}
				if cfg.Port != "8981" {
					t.Errorf("Expected default Port to be '8981', got '%s'", cfg.Port)
				}
				if cfg.OpenAIModel != "gpt-4.1" {
					t.Errorf("Expected default OpenAIModel to be 'gpt-4.1', got '%s'", cfg.OpenAIModel)
				}
				if cfg.LocalReportsDir != "./reports" {
					t.Errorf("Expected default LocalReportsDir to be './reports', got '%s'", cfg.LocalReportsDir)
				}
				if cfg.MockupMode != false {
					t.Errorf("Expected default MockupMode to be false, got %v", cfg.MockupMode)
				}
				if cfg.Environment != "development" {
					t.Errorf("Expected default Environment to be 'development', got '%s'", cfg.Environment)
				}
				if cfg.LogLevel != "info" {
					t.Errorf("Expected default LogLevel to be 'info', got '%s'", cfg.LogLevel)
				}
				if cfg.LogFormat != "auto" {
					t.Errorf("Expected default LogFormat to be 'auto', got '%s'", cfg.LogFormat)
				}
				return nil
			},
		},
		{
			name: "custom configuration values",
			envVars: map[string]string{
				"OPENAI_API_KEY":     "custom-key",
				"PORT":               "9000",
				"OPENAI_MODEL":       "gpt-3.5-turbo",
				"GCP_PROJECT_ID":     "test-project",
				"GCS_BUCKET":         "test-bucket",
				"LOCAL_REPORTS_DIR":  "/custom/reports",
				"MOCKUP_MODE":        "true",
				"ENVIRONMENT":        "production",
				"LOG_LEVEL":          "debug",
				"LOG_FORMAT":         "json",
			},
			expectError: false,
			validate: func(cfg *Config) error {
				if cfg.OpenAIAPIKey != "custom-key" {
					t.Errorf("Expected OpenAIAPIKey to be 'custom-key', got '%s'", cfg.OpenAIAPIKey)
				}
				if cfg.Port != "9000" {
					t.Errorf("Expected Port to be '9000', got '%s'", cfg.Port)
				}
				if cfg.OpenAIModel != "gpt-3.5-turbo" {
					t.Errorf("Expected OpenAIModel to be 'gpt-3.5-turbo', got '%s'", cfg.OpenAIModel)
				}
				if cfg.GCPProjectID != "test-project" {
					t.Errorf("Expected GCPProjectID to be 'test-project', got '%s'", cfg.GCPProjectID)
				}
				if cfg.GCSBucket != "test-bucket" {
					t.Errorf("Expected GCSBucket to be 'test-bucket', got '%s'", cfg.GCSBucket)
				}
				if cfg.LocalReportsDir != "/custom/reports" {
					t.Errorf("Expected LocalReportsDir to be '/custom/reports', got '%s'", cfg.LocalReportsDir)
				}
				if cfg.MockupMode != true {
					t.Errorf("Expected MockupMode to be true, got %v", cfg.MockupMode)
				}
				if cfg.Environment != "production" {
					t.Errorf("Expected Environment to be 'production', got '%s'", cfg.Environment)
				}
				if cfg.LogLevel != "debug" {
					t.Errorf("Expected LogLevel to be 'debug', got '%s'", cfg.LogLevel)
				}
				if cfg.LogFormat != "json" {
					t.Errorf("Expected LogFormat to be 'json', got '%s'", cfg.LogFormat)
				}
				return nil
			},
		},
		{
			name: "custom data source URLs",
			envVars: map[string]string{
				"OPENAI_API_KEY":   "test-key",
				"NOAA_K_INDEX_URL": "https://custom.noaa.gov/k-index",
				"NOAA_SOLAR_URL":   "https://custom.noaa.gov/solar",
				"N0NBH_SOLAR_URL":  "https://custom.hamqsl.com/api",
				"SIDC_RSS_URL":     "https://custom.sidc.be/rss",
			},
			expectError: false,
			validate: func(cfg *Config) error {
				if cfg.NOAAKIndexURL != "https://custom.noaa.gov/k-index" {
					t.Errorf("Expected custom NOAA K-index URL, got '%s'", cfg.NOAAKIndexURL)
				}
				if cfg.NOAASolarURL != "https://custom.noaa.gov/solar" {
					t.Errorf("Expected custom NOAA Solar URL, got '%s'", cfg.NOAASolarURL)
				}
				if cfg.N0NBHSolarURL != "https://custom.hamqsl.com/api" {
					t.Errorf("Expected custom N0NBH URL, got '%s'", cfg.N0NBHSolarURL)
				}
				if cfg.SIDCRSSURL != "https://custom.sidc.be/rss" {
					t.Errorf("Expected custom SIDC RSS URL, got '%s'", cfg.SIDCRSSURL)
				}
				return nil
			},
		},
		{
			name:        "missing required OpenAI API key",
			envVars:     map[string]string{},
			expectError: true,
			validate:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			clearEnv()
			
			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Load configuration
			cfg, err := Load(context.Background())

			// Check error expectation
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
				return
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
				return
			}

			// Validate configuration if no error expected
			if !tt.expectError && tt.validate != nil {
				if err := tt.validate(cfg); err != nil {
					t.Errorf("Configuration validation failed: %v", err)
				}
			}

			// Clean up
			clearEnv()
		})
	}
}

func TestLoadDefaultURLs(t *testing.T) {
	// Clear environment and set only required field
	clearEnv()
	os.Setenv("OPENAI_API_KEY", "test-key")

	cfg, err := Load(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Test default URLs
	expectedURLs := map[string]string{
		"NOAAKIndexURL": "https://services.swpc.noaa.gov/products/noaa-planetary-k-index.json",
		"NOAASolarURL":  "https://services.swpc.noaa.gov/json/solar-cycle/observed-solar-cycle-indices.json",
		"N0NBHSolarURL": "https://www.hamqsl.com/solarapi.php?format=json",
		"SIDCRSSURL":    "https://www.sidc.be/products/meu",
	}

	if cfg.NOAAKIndexURL != expectedURLs["NOAAKIndexURL"] {
		t.Errorf("Expected NOAAKIndexURL to be '%s', got '%s'", expectedURLs["NOAAKIndexURL"], cfg.NOAAKIndexURL)
	}
	if cfg.NOAASolarURL != expectedURLs["NOAASolarURL"] {
		t.Errorf("Expected NOAASolarURL to be '%s', got '%s'", expectedURLs["NOAASolarURL"], cfg.NOAASolarURL)
	}
	if cfg.N0NBHSolarURL != expectedURLs["N0NBHSolarURL"] {
		t.Errorf("Expected N0NBHSolarURL to be '%s', got '%s'", expectedURLs["N0NBHSolarURL"], cfg.N0NBHSolarURL)
	}
	if cfg.SIDCRSSURL != expectedURLs["SIDCRSSURL"] {
		t.Errorf("Expected SIDCRSSURL to be '%s', got '%s'", expectedURLs["SIDCRSSURL"], cfg.SIDCRSSURL)
	}

	clearEnv()
}

func TestLoadWithContext(t *testing.T) {
	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	os.Setenv("OPENAI_API_KEY", "test-key")
	
	// Should still work as envconfig doesn't use context for cancellation
	cfg, err := Load(ctx)
	if err != nil {
		t.Errorf("Expected no error with cancelled context, got: %v", err)
	}
	if cfg == nil {
		t.Error("Expected config to be loaded even with cancelled context")
	}

	clearEnv()
}

// Helper function to clear relevant environment variables
func clearEnv() {
	envVars := []string{
		"PORT", "OPENAI_API_KEY", "OPENAI_MODEL", "GCP_PROJECT_ID", "GCS_BUCKET",
		"LOCAL_REPORTS_DIR", "MOCKUP_MODE", "NOAA_K_INDEX_URL", "NOAA_SOLAR_URL",
		"N0NBH_SOLAR_URL", "SIDC_RSS_URL", "ENVIRONMENT", "LOG_LEVEL", "LOG_FORMAT",
	}
	for _, env := range envVars {
		os.Unsetenv(env)
	}
}
