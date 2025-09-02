package models

// NOAAKIndexResponse represents NOAA K-index JSON response
type NOAAKIndexResponse struct {
	TimeTag     string  `json:"time_tag"`
	KpIndex     float64 `json:"kp_index"`
	EstimatedKp float64 `json:"estimated_kp"`
	Source      string  `json:"source"`
}

// NOAASolarResponse represents NOAA solar cycle data
type NOAASolarResponse struct {
	TimeTag           string  `json:"time_tag"`
	SolarFlux         float64 `json:"f10.7"`
	SunspotNumber     float64 `json:"ssn"`
	SolarFluxAdjusted float64 `json:"f10.7_adj"`
	Source            string  `json:"source"`
}
