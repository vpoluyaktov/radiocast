package storage

import (
	"testing"
	"time"
)

func TestGenerateReportFolderPath(t *testing.T) {
	tests := []struct {
		name      string
		timestamp time.Time
		expected  string
	}{
		{
			name:      "standard date and time",
			timestamp: time.Date(2025, 9, 17, 14, 30, 45, 0, time.UTC),
			expected:  "2025/09/17/PropagationReport-2025-09-17-14-30-45",
		},
		{
			name:      "new year date",
			timestamp: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			expected:  "2025/01/01/PropagationReport-2025-01-01-00-00-00",
		},
		{
			name:      "end of year date",
			timestamp: time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
			expected:  "2024/12/31/PropagationReport-2024-12-31-23-59-59",
		},
		{
			name:      "leap year date",
			timestamp: time.Date(2024, 2, 29, 12, 15, 30, 0, time.UTC),
			expected:  "2024/02/29/PropagationReport-2024-02-29-12-15-30",
		},
		{
			name:      "single digit month and day",
			timestamp: time.Date(2025, 3, 5, 8, 7, 6, 0, time.UTC),
			expected:  "2025/03/05/PropagationReport-2025-03-05-08-07-06",
		},
		{
			name:      "future date",
			timestamp: time.Date(2030, 11, 22, 16, 45, 12, 0, time.UTC),
			expected:  "2030/11/22/PropagationReport-2030-11-22-16-45-12",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateReportFolderPath(tt.timestamp)
			if result != tt.expected {
				t.Errorf("GenerateReportFolderPath() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGenerateReportFolderPathConsistency(t *testing.T) {
	// Test that the same timestamp always generates the same path
	timestamp := time.Date(2025, 9, 17, 14, 30, 45, 0, time.UTC)
	
	path1 := GenerateReportFolderPath(timestamp)
	path2 := GenerateReportFolderPath(timestamp)
	
	if path1 != path2 {
		t.Errorf("GenerateReportFolderPath() should be consistent: %s != %s", path1, path2)
	}
}

func TestGenerateReportFolderPathUniqueness(t *testing.T) {
	// Test that different timestamps generate different paths
	timestamp1 := time.Date(2025, 9, 17, 14, 30, 45, 0, time.UTC)
	timestamp2 := time.Date(2025, 9, 17, 14, 30, 46, 0, time.UTC) // 1 second later
	
	path1 := GenerateReportFolderPath(timestamp1)
	path2 := GenerateReportFolderPath(timestamp2)
	
	if path1 == path2 {
		t.Errorf("Different timestamps should generate different paths: %s == %s", path1, path2)
	}
}

func TestGetContentType(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{
			name:     "JSON file",
			filename: "data.json",
			expected: "application/json",
		},
		{
			name:     "HTML file",
			filename: "index.html",
			expected: "text/html",
		},
		{
			name:     "CSS file",
			filename: "styles.css",
			expected: "text/css",
		},
		{
			name:     "Text file",
			filename: "readme.txt",
			expected: "text/plain",
		},
		{
			name:     "Markdown file",
			filename: "README.md",
			expected: "text/markdown",
		},
		{
			name:     "PNG image",
			filename: "chart.png",
			expected: "image/png",
		},
		{
			name:     "JPEG image",
			filename: "photo.jpg",
			expected: "image/jpeg",
		},
		{
			name:     "JPEG image with jpeg extension",
			filename: "photo.jpeg",
			expected: "image/jpeg",
		},
		{
			name:     "GIF image",
			filename: "animation.gif",
			expected: "image/gif",
		},
		{
			name:     "Unknown file type",
			filename: "data.xyz",
			expected: "application/octet-stream",
		},
		{
			name:     "File without extension",
			filename: "Dockerfile",
			expected: "application/octet-stream",
		},
		{
			name:     "Empty filename",
			filename: "",
			expected: "application/octet-stream",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetContentType(tt.filename)
			if result != tt.expected {
				t.Errorf("GetContentType(%s) = %v, want %v", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestGetContentTypeWithPaths(t *testing.T) {
	// Test that function works with full paths, not just filenames
	tests := []struct {
		name     string
		filepath string
		expected string
	}{
		{
			name:     "nested JSON file",
			filepath: "reports/2025/09/17/data.json",
			expected: "application/json",
		},
		{
			name:     "nested HTML file",
			filepath: "reports/2025/09/17/PropagationReport-2025-09-17-14-30-45/index.html",
			expected: "text/html",
		},
		{
			name:     "nested CSS file",
			filepath: "static/css/styles.css",
			expected: "text/css",
		},
		{
			name:     "nested image file",
			filepath: "images/charts/solar-activity.png",
			expected: "image/png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetContentType(tt.filepath)
			if result != tt.expected {
				t.Errorf("GetContentType(%s) = %v, want %v", tt.filepath, result, tt.expected)
			}
		})
	}
}

func TestGetContentTypeCaseSensitivity(t *testing.T) {
	// Test that the function is case sensitive (as expected)
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{
			name:     "uppercase JSON",
			filename: "data.JSON",
			expected: "application/octet-stream", // Should not match
		},
		{
			name:     "mixed case HTML",
			filename: "index.Html",
			expected: "application/octet-stream", // Should not match
		},
		{
			name:     "uppercase PNG",
			filename: "image.PNG",
			expected: "application/octet-stream", // Should not match
		},
		{
			name:     "lowercase extensions work",
			filename: "file.json",
			expected: "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetContentType(tt.filename)
			if result != tt.expected {
				t.Errorf("GetContentType(%s) = %v, want %v", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestGetContentTypeEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{
			name:     "multiple dots in filename",
			filename: "backup.data.json",
			expected: "application/json",
		},
		{
			name:     "filename starting with dot",
			filename: ".gitignore",
			expected: "application/octet-stream",
		},
		{
			name:     "filename ending with dot",
			filename: "file.",
			expected: "application/octet-stream",
		},
		{
			name:     "extension at beginning",
			filename: "json.data",
			expected: "application/octet-stream",
		},
		{
			name:     "very long filename",
			filename: "very-long-filename-with-lots-of-characters-and-numbers-12345.html",
			expected: "text/html",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetContentType(tt.filename)
			if result != tt.expected {
				t.Errorf("GetContentType(%s) = %v, want %v", tt.filename, result, tt.expected)
			}
		})
	}
}

func BenchmarkGenerateReportFolderPath(b *testing.B) {
	timestamp := time.Date(2025, 9, 17, 14, 30, 45, 0, time.UTC)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateReportFolderPath(timestamp)
	}
}

func BenchmarkGetContentType(b *testing.B) {
	filenames := []string{
		"data.json",
		"index.html",
		"styles.css",
		"image.png",
		"document.txt",
		"unknown.xyz",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, filename := range filenames {
			GetContentType(filename)
		}
	}
}
