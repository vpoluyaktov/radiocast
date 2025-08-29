package config

import (
	"os"
	"path/filepath"
	"strings"
)

// GetVersion reads the version from the VERSION file
func GetVersion() string {
	// Try to read from VERSION file in project root
	versionPath := filepath.Join("..", "VERSION")
	if content, err := os.ReadFile(versionPath); err == nil {
		return strings.TrimSpace(string(content))
	}
	
	// Fallback: try from service directory
	versionPath = filepath.Join("..", "..", "VERSION")
	if content, err := os.ReadFile(versionPath); err == nil {
		return strings.TrimSpace(string(content))
	}
	
	// Final fallback
	return "unknown"
}
