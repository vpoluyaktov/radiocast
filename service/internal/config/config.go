package config

import (
	"context"
	"fmt"

	"github.com/sethvargo/go-envconfig"
)

// Config holds all configuration for the radio propagation service
type Config struct {
	// Server configuration
	Port string `env:"PORT,default=8981"`
	
	// OpenAI configuration
	OpenAIAPIKey string `env:"OPENAI_API_KEY,required"`
	OpenAIModel  string `env:"OPENAI_MODEL,default=gpt-4.1"`
	
	// GCP configuration (optional for local testing)
	GCPProjectID string `env:"GCP_PROJECT_ID"`
	GCSBucket    string `env:"GCS_BUCKET"`
	
	// Local testing configuration
	LocalReportsDir string `env:"LOCAL_REPORTS_DIR,default=./reports"`
	MockupMode      bool   `env:"MOCKUP_MODE,default=false"`
	
	// Data source URLs
	NOAAKIndexURL string `env:"NOAA_K_INDEX_URL,default=https://services.swpc.noaa.gov/products/noaa-planetary-k-index.json"`
	NOAASolarURL  string `env:"NOAA_SOLAR_URL,default=https://services.swpc.noaa.gov/json/solar-cycle/observed-solar-cycle-indices.json"`
	N0NBHSolarURL string `env:"N0NBH_SOLAR_URL,default=https://www.hamqsl.com/solarapi.php?format=json"`
	SIDCRSSURL    string `env:"SIDC_RSS_URL,default=https://www.sidc.be/products/meu"`
	
	// Service configuration
	Environment string `env:"ENVIRONMENT,default=development"`
	LogLevel    string `env:"LOG_LEVEL,default=info"`
}

// Load loads configuration from environment variables
func Load(ctx context.Context) (*Config, error) {
	var cfg Config
	if err := envconfig.Process(ctx, &cfg); err != nil {
		return nil, fmt.Errorf("failed to process config: %w", err)
	}
	return &cfg, nil
}
