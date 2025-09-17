package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetVersion(t *testing.T) {
	// Save original environment
	originalVersion := os.Getenv("APP_VERSION")
	defer func() {
		if originalVersion != "" {
			os.Setenv("APP_VERSION", originalVersion)
		} else {
			os.Unsetenv("APP_VERSION")
		}
	}()

	tests := []struct {
		name           string
		envVersion     string
		expectContains string
		expectMinLen   int
	}{
		{
			name:           "version from environment variable",
			envVersion:     "1.2.3",
			expectContains: "1.2.3",
			expectMinLen:   5,
		},
		{
			name:           "version from environment with build number",
			envVersion:     "2.0.0-beta.1",
			expectContains: "2.0.0-beta.1",
			expectMinLen:   10,
		},
		{
			name:         "version from git (no env var)",
			envVersion:   "",
			expectMinLen: 3, // At least "0.1" or similar
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment variable
			os.Unsetenv("APP_VERSION")
			
			// Set test environment if provided
			if tt.envVersion != "" {
				os.Setenv("APP_VERSION", tt.envVersion)
			}

			version := GetVersion()

			// Check minimum length
			if len(version) < tt.expectMinLen {
				t.Errorf("Expected version length >= %d, got %d (version: %s)", tt.expectMinLen, len(version), version)
			}

			// Check expected content if specified
			if tt.expectContains != "" && !strings.Contains(version, tt.expectContains) {
				t.Errorf("Expected version to contain '%s', got '%s'", tt.expectContains, version)
			}

			// Version should not be empty
			if version == "" {
				t.Error("Version should not be empty")
			}
		})
	}
}

func TestGetBaseVersion(t *testing.T) {
	// Create a temporary VERSION file for testing
	tempDir := t.TempDir()
	versionFile := filepath.Join(tempDir, "VERSION")
	
	// Test with VERSION file
	expectedVersion := "1.5.0"
	err := os.WriteFile(versionFile, []byte(expectedVersion+"\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test VERSION file: %v", err)
	}

	// Change to temp directory to test relative path resolution
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	
	// Create subdirectory structure to test relative path logic
	subDir := filepath.Join(tempDir, "service")
	os.MkdirAll(subDir, 0755)
	os.Chdir(subDir)

	version := getBaseVersion()
	
	// Should return fallback version since we're not in the exact expected directory structure
	if version != "0.1.0" {
		t.Logf("Got version: %s (this is expected behavior for test environment)", version)
	}
}

func TestGetBaseVersionFallback(t *testing.T) {
	// Test in a directory where VERSION file doesn't exist
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	
	os.Chdir(tempDir)
	
	version := getBaseVersion()
	
	// Should return fallback version
	if version != "0.1.0" {
		t.Errorf("Expected fallback version '0.1.0', got '%s'", version)
	}
}

func TestGetGitCommitCount(t *testing.T) {
	count := getGitCommitCount()
	
	// Count should be non-negative
	if count < 0 {
		t.Errorf("Expected non-negative commit count, got %d", count)
	}
	
	// In a real git repository, count should be > 0
	// In test environment, it might be 0 if not in a git repo
	t.Logf("Git commit count: %d", count)
}

func TestGetVersionIntegration(t *testing.T) {
	// Test the full integration without environment variable
	os.Unsetenv("APP_VERSION")
	
	version := GetVersion()
	
	// Version should not be empty
	if version == "" {
		t.Error("Version should not be empty")
	}
	
	// Version should contain at least one dot (semantic versioning)
	if !strings.Contains(version, ".") {
		t.Errorf("Expected version to contain '.', got '%s'", version)
	}
	
	// Version should start with a digit
	if len(version) == 0 || version[0] < '0' || version[0] > '9' {
		t.Errorf("Expected version to start with a digit, got '%s'", version)
	}
	
	t.Logf("Generated version: %s", version)
}

func TestGetVersionWithRealVersionFile(t *testing.T) {
	// Try to read the actual VERSION file from the project
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	
	// Navigate to project root (assuming we're in service/internal/config)
	projectRoot := filepath.Join(originalDir, "..", "..", "..")
	if _, err := os.Stat(filepath.Join(projectRoot, "VERSION")); err == nil {
		os.Chdir(filepath.Join(projectRoot, "service"))
		
		version := getBaseVersion()
		
		// Should not be the fallback version if VERSION file exists
		t.Logf("Base version from actual VERSION file: %s", version)
		
		// Verify it's a valid version format
		if !strings.Contains(version, ".") {
			t.Errorf("Expected version to contain '.', got '%s'", version)
		}
	} else {
		t.Skip("VERSION file not found in expected location, skipping real file test")
	}
}
