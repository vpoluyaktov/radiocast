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

// LoadCSSStyles loads the CSS styles from file
func (t *TemplateLoader) LoadCSSStyles() (string, error) {
	cssPath := filepath.Join("internal", "templates", "report_styles.css")
	content, err := os.ReadFile(cssPath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
