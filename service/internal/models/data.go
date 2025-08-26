package models

import "time"

// PropagationData represents normalized radio propagation data from all sources
type PropagationData struct {
	Timestamp    time.Time     `json:"timestamp"`
	SolarData    SolarData     `json:"solar_data"`
	GeomagData   GeomagData    `json:"geomag_data"`
	BandData     BandData      `json:"band_data"`
	Forecast     ForecastData  `json:"forecast"`
	SourceEvents []SourceEvent `json:"source_events"`
}

// SolarData contains solar activity information
type SolarData struct {
	SolarFluxIndex    float64 `json:"solar_flux_index"`    // 10.7cm flux
	SunspotNumber     int     `json:"sunspot_number"`      // Daily sunspot number
	SolarActivity     string  `json:"solar_activity"`      // Low/Moderate/High
	FlareActivity     string  `json:"flare_activity"`      // Current flare status
	SolarCyclePhase   string  `json:"solar_cycle_phase"`   // Current solar cycle info
	LastMajorFlare    string  `json:"last_major_flare"`    // Most recent significant flare
	SolarWindSpeed    float64 `json:"solar_wind_speed"`    // km/s
	ProtonFlux        float64 `json:"proton_flux"`         // particles/cm²/s
}

// GeomagData contains geomagnetic activity information
type GeomagData struct {
	KIndex           float64 `json:"k_index"`            // Current planetary K-index
	AIndex           float64 `json:"a_index"`            // Current A-index
	GeomagActivity   string  `json:"geomag_activity"`    // Quiet/Unsettled/Active/Storm
	GeomagConditions string  `json:"geomag_conditions"`  // Current conditions description
	MagneticField    float64 `json:"magnetic_field"`     // nT
}

// BandData contains HF band condition information
type BandData struct {
	Band80m  BandCondition `json:"band_80m"`
	Band40m  BandCondition `json:"band_40m"`
	Band20m  BandCondition `json:"band_20m"`
	Band17m  BandCondition `json:"band_17m"`
	Band15m  BandCondition `json:"band_15m"`
	Band12m  BandCondition `json:"band_12m"`
	Band10m  BandCondition `json:"band_10m"`
	Band6m   BandCondition `json:"band_6m"`
	VHFPlus  BandCondition `json:"vhf_plus"`
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

// NOAAKIndexResponse represents NOAA K-index JSON response
type NOAAKIndexResponse struct {
	TimeTag     string  `json:"time_tag"`
	KpIndex     float64 `json:"kp_index"`
	EstimatedKp float64 `json:"estimated_kp"`
}

// NOAASolarResponse represents NOAA solar cycle data
type NOAASolarResponse struct {
	TimeTag           string  `json:"time_tag"`
	SolarFlux         float64 `json:"f10.7"`
	SunspotNumber     float64 `json:"ssn"`
	SolarFluxAdjusted float64 `json:"f10.7_adj"`
}

// N0NBHResponse represents N0NBH solar API response
type N0NBHResponse struct {
	SolarData struct {
		SolarFlux     string `json:"solarflux"`
		AIndex        string `json:"aindex"`
		KIndex        string `json:"kindex"`
		KIndexNT      string `json:"kindexnt"`
		SunSpots      string `json:"sunspots"`
		HeliumLine    string `json:"heliumline"`
		ProtonFlux    string `json:"protonflux"`
		ElectronFlux  string `json:"electonflux"`
		Aurora        string `json:"aurora"`
		NormalizationTime string `json:"normalization"`
		LatestSWPCReport  string `json:"latestswpcreport"`
	} `json:"solardata"`
	Time string `json:"time"`
	
	// Band conditions
	Calculatedconditions struct {
		Band []struct {
			Name string `json:"name"`
			Time string `json:"time"`
			Day  string `json:"day"`
			Night string `json:"night"`
		} `json:"band"`
	} `json:"calculatedconditions"`
	
	CalculatedVHFConditions struct {
		Phenomenon []struct {
			Name     string `json:"name"`
			Location string `json:"location"`
		} `json:"phenomenon"`
	} `json:"calculatedvhfconditions"`
}
