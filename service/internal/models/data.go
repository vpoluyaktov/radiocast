package models

import (
	"time"
	"github.com/mmcdole/gofeed"
)

// PropagationData represents normalized radio propagation data from all sources
type PropagationData struct {
	Timestamp    time.Time     `json:"timestamp"`
	SolarData    SolarData     `json:"solar_data"`
	GeomagData   GeomagData    `json:"geomag_data"`
	BandData     BandData      `json:"band_data"`
	Forecast     ForecastData  `json:"forecast"`
	SourceEvents []SourceEvent `json:"source_events"`
	
	// Historical time series data for trend analysis
	HistoricalKIndex []KIndexPoint `json:"historical_k_index"`
	HistoricalSolar  []SolarPoint  `json:"historical_solar"`
}

// SourceData contains raw data from all sources before normalization
type SourceData struct {
	NOAAKIndex []NOAAKIndexResponse `json:"noaa_k_index"`
	NOAASolar  []NOAASolarResponse  `json:"noaa_solar"`
	N0NBH      *N0NBHResponse       `json:"n0nbh"`
	SIDC       []*gofeed.Item       `json:"sidc"`
}

// SolarData contains solar activity information
type SolarData struct {
	SolarFluxIndex       float64 `json:"solar_flux_index"`        // 10.7cm flux (current)
	SolarFluxAdjusted    float64 `json:"solar_flux_adjusted"`     // Adjusted 10.7cm flux
	SolarFluxDataSource  string  `json:"solar_flux_data_source"`  // Source API for flux data
	SunspotNumber        int     `json:"sunspot_number"`          // Daily sunspot number
	SunspotDataSource    string  `json:"sunspot_data_source"`     // Source API for sunspot data
	SolarActivity        string  `json:"solar_activity"`          // LLM-generated classification
	
	// Rich N0NBH solar data (previously lost)
	XRayFlux             string  `json:"xray_flux"`               // X-ray flux level (e.g., "C1.2")
	SolarWindSpeed       float64 `json:"solar_wind_speed"`        // km/s
	SolarWindDataSource  string  `json:"solar_wind_data_source"`  // Source API for solar wind data
	ProtonFlux           float64 `json:"proton_flux"`             // particles/cm²/s
	ProtonFluxDataSource string  `json:"proton_flux_data_source"` // Source API for proton flux data
	ElectronFlux         string  `json:"electron_flux"`           // Electron flux level
	HeliumLine           string  `json:"helium_line"`             // Helium line data
	Aurora               string  `json:"aurora"`                  // Aurora activity level
	
	// Derived/calculated fields
	FlareActivity        string  `json:"flare_activity"`          // Current flare status
	SolarCyclePhase      string  `json:"solar_cycle_phase"`       // Current solar cycle info
	LastMajorFlare       string  `json:"last_major_flare"`        // Most recent significant flare
}

// GeomagData contains geomagnetic activity information
type GeomagData struct {
	KIndex              float64 `json:"k_index"`                // Current planetary K-index
	KIndexDataSource    string  `json:"k_index_data_source"`    // Source API for K-index data
	AIndex              float64 `json:"a_index"`                // Current A-index
	AIndexDataSource    string  `json:"a_index_data_source"`    // Source API for A-index data
	GeomagActivity      string  `json:"geomag_activity"`        // LLM-generated classification
	
	// Rich N0NBH geomagnetic data (previously lost)
	MagneticField       float64 `json:"magnetic_field"`         // nT
	MagneticFieldDataSource string `json:"magnetic_field_data_source"` // Source API for magnetic field data
	LatDegree           string  `json:"lat_degree"`             // Latitude degree from N0NBH
	
	// Derived/calculated fields
	GeomagConditions    string  `json:"geomag_conditions"`      // Current conditions description
}

// BandData contains HF band condition information
type BandData struct {
	Band80m         BandCondition `json:"band_80m"`
	Band40m         BandCondition `json:"band_40m"`
	Band20m         BandCondition `json:"band_20m"`
	Band17m         BandCondition `json:"band_17m"`
	Band15m         BandCondition `json:"band_15m"`
	Band12m         BandCondition `json:"band_12m"`
	Band10m         BandCondition `json:"band_10m"`
	Band6m          BandCondition `json:"band_6m"`
	VHFPlus         BandCondition `json:"vhf_plus"`
	BandDataSource  string        `json:"band_data_source"` // Source API for band data
}

// BandCondition represents propagation conditions for a specific band
type BandCondition struct {
	Day   string `json:"day"`   // Poor/Fair/Good/Excellent
	Night string `json:"night"` // Poor/Fair/Good/Excellent
}

// ForecastData contains propagation forecasts
type ForecastData struct {
	Today     DayForecast `json:"today"`
	Tomorrow  DayForecast `json:"tomorrow"`
	DayAfter  DayForecast `json:"day_after"`
	Outlook   string      `json:"outlook"`   // General 3-day outlook
	Warnings  []string    `json:"warnings"`  // Any propagation warnings
}

// DayForecast represents a single day's forecast
type DayForecast struct {
	Date            time.Time `json:"date"`
	KIndexForecast  string    `json:"k_index_forecast"`  // Expected K-index range
	SolarActivity   string    `json:"solar_activity"`    // Expected solar activity
	HFConditions    string    `json:"hf_conditions"`     // Expected HF conditions
	VHFConditions   string    `json:"vhf_conditions"`    // Expected VHF+ conditions
	BestBands       []string  `json:"best_bands"`        // Recommended bands
	WorstBands      []string  `json:"worst_bands"`       // Bands to avoid
}

// SourceEvent represents notable events from data sources
type SourceEvent struct {
	Source      string    `json:"source"`       // NOAA/N0NBH/SIDC
	EventType   string    `json:"event_type"`   // Flare/CME/Storm/etc
	Severity    string    `json:"severity"`     // Low/Moderate/High/Extreme
	Description string    `json:"description"`  // Event description
	Timestamp   time.Time `json:"timestamp"`    // When event occurred/detected
	Impact      string    `json:"impact"`       // Expected propagation impact
}

// KIndexPoint represents a single K-index measurement with timestamp
type KIndexPoint struct {
	Timestamp   time.Time `json:"timestamp"`
	KIndex      float64   `json:"k_index"`
	EstimatedKp float64   `json:"estimated_kp"`
	Source      string    `json:"source"`
}

// SolarPoint represents a single solar measurement with timestamp  
type SolarPoint struct {
	Timestamp         time.Time `json:"timestamp"`
	SolarFlux         float64   `json:"solar_flux"`
	SolarFluxAdjusted float64   `json:"solar_flux_adjusted"`
	SunspotNumber     float64   `json:"sunspot_number"`
	Source            string    `json:"source"`
}

