package models

import "encoding/xml"

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
