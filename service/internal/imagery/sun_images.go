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

// GenerateSunImagesUnavailableHTML creates the HTML snippet shown when the Helioviewer API is unavailable
func (sig *SunImageGenerator) GenerateSunImagesUnavailableHTML() template.HTML {
	html := `<div class="chart-section"><div class="chart-container"><h3>Sun Images for Past 72 Hours</h3>` +
		`<div style="padding:40px 20px;text-align:center;color:#666;background:rgba(255,255,255,0.05);border-radius:8px;border:1px dashed #444;">` +
		`<p style="font-size:1.1em;margin-bottom:10px;">Sun images are temporarily unavailable because the ` +
		`<a href="https://helioviewer.org" target="_blank" rel="noopener noreferrer" style="color:#007bff;">Helioviewer API</a> is currently not accessible.</p>` +
		`<p style="font-size:0.9em;">For more information, see the ` +
		`<a href="https://helioviewer-project.github.io/2026/03/05/building-move" target="_blank" rel="noopener noreferrer" style="color:#007bff;">Helioviewer maintenance page</a>.</p>` +
		`</div></div></div>`
	return template.HTML(html)
}
