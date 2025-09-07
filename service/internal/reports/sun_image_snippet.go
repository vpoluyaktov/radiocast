package reports

import (
	"context"
	"fmt"
	"html/template"
)

// generateSunImageSnippet creates the Sun image HTML snippet
func (rs *ReportService) generateSunImageSnippet(_ context.Context, folderPath string) (template.HTML, error) {
	// Generate Sun GIF path
	gifPath := "sun_72h.gif"
	if folderPath != "" {
		gifPath = fmt.Sprintf("/files/%s/sun_72h.gif", folderPath)
	}

	// Create HTML for Sun GIF
	sunHTML := fmt.Sprintf(`
<div class="sun-gif-container">
	<img src="%s" alt="72-hour Sun Animation" class="sun-gif" />
	<p class="sun-caption">72-hour solar activity animation from Helioviewer</p>
</div>`, gifPath)

	return template.HTML(sunHTML), nil
}
