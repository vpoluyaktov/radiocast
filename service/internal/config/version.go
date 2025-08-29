package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// GetVersion calculates version using base version + git commit count
func GetVersion() string {
	baseVersion := getBaseVersion()
	commitCount := getGitCommitCount()
	
	if commitCount > 0 {
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
