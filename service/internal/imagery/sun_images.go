package imagery

import (
	"fmt"
	"html/template"
	"strings"
)

// SunImageGenerator handles generation of Sun image HTML snippets
type SunImageGenerator struct{}

// NewSunImageGenerator creates a new SunImageGenerator
func NewSunImageGenerator() *SunImageGenerator {
	return &SunImageGenerator{}
}

// GenerateSunImagesHTML creates the HTML snippet for the Sun Images section
func (sig *SunImageGenerator) GenerateSunImagesHTML(gifRelName, folderPath string) template.HTML {
	var imgSrc string
	if folderPath != "" {
		// GCS mode - use the full folder path
		if !strings.HasSuffix(folderPath, "/") {
			folderPath += "/"
		}
		imgSrc = "/reports/" + folderPath + gifRelName
	} else {
		// Local mode - use relative path
		imgSrc = gifRelName
	}
	
	html := fmt.Sprintf(`<div class="chart-section"><div class="chart-container"><h3>Sun Images for Past 72 Hours</h3><img src="%s" alt="Sun last 72h" style="max-width:100%%;height:auto;border-radius:8px;" /><br/><i>Images copyrighted by the SDO/NASA and Helioviewer project</i></div></div>`, imgSrc)
	
	return template.HTML(html)
}
