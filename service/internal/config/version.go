package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// GetVersion returns version from environment variable or calculates from git
func GetVersion() string {
	// First try to get version from environment variable (set by CI/CD)
	if envVersion := os.Getenv("APP_VERSION"); envVersion != "" {
		return envVersion
	}
	
	// Fallback to calculating from git (for local development)
	baseVersion := getBaseVersion()
	commitCount := getGitCommitCount()
	
	if commitCount > 0 {
		// Cut the last digit from base version and replace with commit count
		// e.g., "0.1.0" -> "0.1" + "." + "231" = "0.1.231"
		parts := strings.Split(baseVersion, ".")
		if len(parts) >= 3 {
			// Take first two parts (major.minor) and add commit count as patch
			return parts[0] + "." + parts[1] + "." + strconv.Itoa(commitCount)
		}
		// Fallback if version format is unexpected
		return baseVersion + "." + strconv.Itoa(commitCount)
	}
	
	return baseVersion
}

// getBaseVersion reads the base version from VERSION file
func getBaseVersion() string {
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
	return "0.1.0"
}

// getGitCommitCount gets the total commit count from git
func getGitCommitCount() int {
	cmd := exec.Command("git", "rev-list", "--all", "--count", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}
	
	countStr := strings.TrimSpace(string(output))
	count, err := strconv.Atoi(countStr)
	if err != nil {
		return 0
	}
	
	return count
}
