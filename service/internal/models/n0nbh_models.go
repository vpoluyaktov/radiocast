package models

import "encoding/xml"

// NOAA API Response DTOs
type NOAAKIndexAPIResponse struct {
	TimeTag     string  `json:"time_tag"`
	KpIndex     int     `json:"kp_index"`
	EstimatedKp float64 `json:"estimated_kp"`
	Kp          string  `json:"kp"`
}

type NOAASolarAPIResponse struct {
	TimeTag           string  `json:"time-tag"`
	SSN               float64 `json:"ssn"`
	SmoothedSSN       float64 `json:"smoothed_ssn"`
	ObservedSWPCSSN   float64 `json:"observed_swpc_ssn"`
	SmoothedSWPCSSN   float64 `json:"smoothed_swpc_ssn"`
	F107              float64 `json:"f10.7"`
	SmoothedF107      float64 `json:"smoothed_f10.7"`
}

// N0NBH Band Condition DTO
type N0NBHBandCondition struct {
	Name   string `json:"name"`
	Time   string `json:"time"`
	Day    string `json:"day"`
	Night  string `json:"night"`
	Source string `json:"source"`
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
	Time   string `json:"time"`
	Source string `json:"source"`
	
	// Band conditions
	Calculatedconditions struct {
		Band []struct {
			Name   string `json:"name"`
			Time   string `json:"time"`
			Day    string `json:"day"`
			Night  string `json:"night"`
			Source string `json:"source"`
		} `json:"band"`
	} `json:"calculatedconditions"`
	
	CalculatedVHFConditions struct {
		Phenomenon []struct {
			Name     string `json:"name"`
			Location string `json:"location"`
		} `json:"phenomenon"`
	} `json:"calculatedvhfconditions"`
}

// N0NBHXMLResponse represents N0NBH XML API response structure
type N0NBHXMLResponse struct {
	XMLName    xml.Name `xml:"solar"`
	SolarData  struct {
		Source        string `xml:"source"`
		Updated       string `xml:"updated"`
		SolarFlux     string `xml:"solarflux"`
		AIndex        string `xml:"aindex"`
		KIndex        string `xml:"kindex"`
		KIndexNT      string `xml:"kindexnt"`
		XRay          string `xml:"xray"`
		SunSpots      string `xml:"sunspots"`
		HeliumLine    string `xml:"heliumline"`
		ProtonFlux    string `xml:"protonflux"`
		ElectronFlux  string `xml:"electonflux"`
		Aurora        string `xml:"aurora"`
		Normalization string `xml:"normalization"`
		LatDegree     string `xml:"latdegree"`
		SolarWind     string `xml:"solarwind"`
		MagneticField string `xml:"magneticfield"`
		
		CalculatedConditions struct {
			Band []struct {
				Name      string `xml:"name,attr"`
				Time      string `xml:"time,attr"`
				Condition string `xml:",chardata"`
			} `xml:"band"`
		} `xml:"calculatedconditions"`
	} `xml:"solardata"`
	Time string `xml:"time"`
}
