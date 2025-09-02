package reports

import (
	"bytes"
	"fmt"
	"html/template"
	"time"

	"github.com/russross/blackfriday/v2"

	"radiocast/internal/config"
	"radiocast/internal/models"
)

// HTMLBuilder handles HTML generation and template processing
type HTMLBuilder struct {
	templateLoader *TemplateLoader
}

// NewHTMLBuilder creates a new HTML builder
func NewHTMLBuilder() *HTMLBuilder {
	return &HTMLBuilder{
		templateLoader: NewTemplateLoader(),
	}
}

// MarkdownToHTML converts markdown to HTML using blackfriday
func (h *HTMLBuilder) MarkdownToHTML(markdownText string) string {
	htmlBytes := blackfriday.Run([]byte(markdownText))
	return string(htmlBytes)
}

// ConvertMarkdownToHTML converts markdown content to a complete HTML document using configurable templates
func (h *HTMLBuilder) ConvertMarkdownToHTML(markdownContent string, date string) (string, error) {
	// Convert markdown to HTML using blackfriday
	htmlBytes := blackfriday.Run([]byte(markdownContent))
	htmlContent := string(htmlBytes)
	
	// Load HTML template
	htmlTemplate, err := h.templateLoader.LoadHTMLTemplate()
	if err != nil {
		return "", fmt.Errorf("failed to load HTML template: %w", err)
	}
	
	// Load CSS styles
	cssStyles, err := h.templateLoader.LoadCSSStyles()
	if err != nil {
		return "", fmt.Errorf("failed to load CSS styles: %w", err)
	}
	
	// Parse the HTML template with proper functions for unescaped content
	tmpl, err := template.New("report").Funcs(template.FuncMap{
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"safeCSS": func(s string) template.CSS {
			return template.CSS(s)
		},
	}).Parse(htmlTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML template: %w", err)
	}
	
	// Prepare template data
	templateData := struct {
		Date        string
		GeneratedAt string
		Content     template.HTML
		CSSStyles   template.CSS
		Charts      template.HTML
		Version     string
	}{
		Date:        date,
		GeneratedAt: time.Now().Format("2006-01-02 15:04:05 UTC"),
		Content:     template.HTML(htmlContent),
		CSSStyles:   template.CSS(cssStyles),
		Charts:      template.HTML(""), // Charts will be embedded in content
		Version:     config.GetVersion(),
	}
	
	// Execute the template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}
	
	return buf.String(), nil
}

// ConvertMarkdownToHTMLWithCharts converts markdown content to HTML with charts
func (h *HTMLBuilder) ConvertMarkdownToHTMLWithCharts(markdownContent string, charts string, date string) (string, error) {
	// Convert markdown to HTML using blackfriday
	htmlBytes := blackfriday.Run([]byte(markdownContent))
	htmlContent := string(htmlBytes)
	
	// Load HTML template
	htmlTemplate, err := h.templateLoader.LoadHTMLTemplate()
	if err != nil {
		return "", fmt.Errorf("failed to load HTML template: %w", err)
	}
	
	// Load CSS styles
	cssStyles, err := h.templateLoader.LoadCSSStyles()
	if err != nil {
		return "", fmt.Errorf("failed to load CSS styles: %w", err)
	}
	
	// Parse the HTML template with proper functions for unescaped content
	tmpl, err := template.New("report").Funcs(template.FuncMap{
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"safeCSS": func(s string) template.CSS {
			return template.CSS(s)
		},
	}).Parse(htmlTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML template: %w", err)
	}
	
	// Prepare template data with charts
	templateData := struct {
		Date        string
		GeneratedAt string
		Content     template.HTML
		CSSStyles   template.CSS
		Charts      template.HTML
		Version     string
	}{
		Date:        date,
		GeneratedAt: time.Now().Format("2006-01-02 15:04:05 UTC"),
		Content:     template.HTML(htmlContent),
		CSSStyles:   template.CSS(cssStyles),
		Charts:      template.HTML(charts), // Now properly populated with charts
		Version:     config.GetVersion(),
	}
	
	// Execute the template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}
	
	return buf.String(), nil
}

// BuildCompleteHTML creates a complete HTML document
func (h *HTMLBuilder) BuildCompleteHTML(content, charts string, data *models.PropagationData) (string, error) {
	// Use the new template-based conversion with charts
	result, err := h.ConvertMarkdownToHTMLWithCharts(content, charts, time.Now().Format("2006-01-02"))
	if err != nil {
		return "", err
	}
	return result, nil
}
