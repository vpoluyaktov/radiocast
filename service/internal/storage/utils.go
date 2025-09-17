package storage

import (
	"fmt"
	"strings"
	"time"
)

// GenerateReportFolderPath generates a consistent folder path for reports
// Format: YYYY/MM/DD/PropagationReport-YYYY-MM-DD-HH-MM-SS
func GenerateReportFolderPath(timestamp time.Time) string {
	return fmt.Sprintf("%04d/%02d/%02d/PropagationReport-%04d-%02d-%02d-%02d-%02d-%02d",
		timestamp.Year(), timestamp.Month(), timestamp.Day(),
		timestamp.Year(), timestamp.Month(), timestamp.Day(),
		timestamp.Hour(), timestamp.Minute(), timestamp.Second())
}


// GetContentType determines the MIME content type based on file extension
func GetContentType(filename string) string {
	if strings.HasSuffix(filename, ".json") {
		return "application/json"
	} else if strings.HasSuffix(filename, ".txt") {
		return "text/plain"
	} else if strings.HasSuffix(filename, ".html") {
		return "text/html"
	} else if strings.HasSuffix(filename, ".css") {
		return "text/css"
	} else if strings.HasSuffix(filename, ".md") {
		return "text/markdown"
	} else if strings.HasSuffix(filename, ".png") {
		return "image/png"
	} else if strings.HasSuffix(filename, ".jpg") || strings.HasSuffix(filename, ".jpeg") {
		return "image/jpeg"
	} else if strings.HasSuffix(filename, ".gif") {
		return "image/gif"
	} else {
		return "application/octet-stream"
	}
}
