package server

import "strings"

// GetContentType returns the appropriate content type for a file based on its extension
func GetContentType(filePath string) string {
	if strings.HasSuffix(filePath, ".html") {
		return "text/html"
	} else if strings.HasSuffix(filePath, ".png") {
		return "image/png"
	} else if strings.HasSuffix(filePath, ".gif") {
		return "image/gif"
	} else if strings.HasSuffix(filePath, ".json") {
		return "application/json"
	} else if strings.HasSuffix(filePath, ".txt") {
		return "text/plain"
	} else if strings.HasSuffix(filePath, ".md") {
		return "text/markdown"
	} else if strings.HasSuffix(filePath, ".css") {
		return "text/css"
	} else if strings.HasSuffix(filePath, ".js") {
		return "application/javascript"
	}
	return "application/octet-stream"
}
