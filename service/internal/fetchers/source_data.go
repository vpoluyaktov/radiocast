package fetchers

import (
	"radiocast/internal/models"
	"github.com/mmcdole/gofeed"
)

// SourceData contains raw data from all sources before normalization
type SourceData struct {
	NOAAKIndex []models.NOAAKIndexResponse `json:"noaa_k_index"`
	NOAASolar  []models.NOAASolarResponse  `json:"noaa_solar"`
	N0NBH      *models.N0NBHResponse       `json:"n0nbh"`
	SIDC       []*gofeed.Item              `json:"sidc"`
}
