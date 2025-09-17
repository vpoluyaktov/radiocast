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
		return "", err
	}
	return string(content), nil
}

