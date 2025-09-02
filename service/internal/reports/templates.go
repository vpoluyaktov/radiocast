package reports

import (
	"os"
	"path/filepath"
)

// TemplateLoader handles loading HTML templates and CSS styles
type TemplateLoader struct{}

// NewTemplateLoader creates a new template loader
func NewTemplateLoader() *TemplateLoader {
	return &TemplateLoader{}
}

// LoadHTMLTemplate loads the HTML template from file
func (t *TemplateLoader) LoadHTMLTemplate() (string, error) {
	templatePath := filepath.Join("internal", "templates", "report_template.html")
	content, err := os.ReadFile(templatePath)
	if err != nil {
		// Return default template if file not found
		return t.GetDefaultHTMLTemplate(), nil
	}
	return string(content), nil
}

// LoadCSSStyles loads the CSS styles from file
func (t *TemplateLoader) LoadCSSStyles() (string, error) {
	cssPath := filepath.Join("internal", "templates", "report_styles.css")
	content, err := os.ReadFile(cssPath)
	if err != nil {
		// Return default styles if file not found
		return t.GetDefaultCSSStyles(), nil
	}
	return string(content), nil
}

// GetDefaultHTMLTemplate returns a fallback HTML template
func (t *TemplateLoader) GetDefaultHTMLTemplate() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Radio Propagation Report - {{.Date}}</title>
    <style>{{.CSSStyles}}</style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Radio Propagation Report</h1>
            <h2>{{.Date}}</h2>
        </div>
        <div class="content">
            {{.Content}}
        </div>
        <div class="footer">
            <hr>
            <p class="version-info">Generated on {{.GeneratedAt}} | Radio Propagation Service v{{.Version}}</p>
        </div>
    </div>
</body>
</html>`
}

// GetDefaultCSSStyles returns fallback CSS styles
func (t *TemplateLoader) GetDefaultCSSStyles() string {
	return `body { font-family: Arial, sans-serif; margin: 20px; }
.container { max-width: 1200px; margin: 0 auto; }
.header { text-align: center; margin-bottom: 30px; }
.content { background: white; padding: 20px; }
.footer { margin-top: 30px; text-align: center; }
.version-info { color: #666; font-size: 0.9em; margin: 10px 0; }`
}
