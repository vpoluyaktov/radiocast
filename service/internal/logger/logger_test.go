package logger

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestLoggerLevels(t *testing.T) {
	var buf bytes.Buffer
	
	// Create logger with DEBUG level
	logger := New(Config{
		Level:     DEBUG,
		Format:    JSONFormat,
		Output:    &buf,
		Component: "test",
	})
	
	// Test all log levels
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message", nil)
	
	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	
	if len(lines) != 4 {
		t.Errorf("Expected 4 log lines, got %d", len(lines))
	}
	
	// Verify each line is valid JSON
	for i, line := range lines {
		var entry LogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Errorf("Line %d is not valid JSON: %v", i+1, err)
		}
	}
}

func TestLoggerLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	
	// Create logger with WARN level (should filter out DEBUG and INFO)
	logger := New(Config{
		Level:     WARN,
		Format:    JSONFormat,
		Output:    &buf,
		Component: "test",
	})
	
	logger.Debug("debug message")  // Should be filtered
	logger.Info("info message")    // Should be filtered
	logger.Warn("warn message")    // Should appear
	logger.Error("error message", nil) // Should appear
	
	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	
	if len(lines) != 2 {
		t.Errorf("Expected 2 log lines with WARN level, got %d", len(lines))
	}
}

func TestJSONFormat(t *testing.T) {
	var buf bytes.Buffer
	
	logger := New(Config{
		Level:     INFO,
		Format:    JSONFormat,
		Output:    &buf,
		Component: "test-component",
	})
	
	logger.Info("test message", map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	})
	
	var entry LogEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}
	
	if entry.Level != "INFO" {
		t.Errorf("Expected level INFO, got %s", entry.Level)
	}
	
	if entry.Message != "test message" {
		t.Errorf("Expected message 'test message', got %s", entry.Message)
	}
	
	if entry.Component != "test-component" {
		t.Errorf("Expected component 'test-component', got %s", entry.Component)
	}
	
	if entry.Fields["key1"] != "value1" {
		t.Errorf("Expected field key1='value1', got %v", entry.Fields["key1"])
	}
	
	if entry.Fields["key2"] != float64(42) { // JSON numbers are float64
		t.Errorf("Expected field key2=42, got %v", entry.Fields["key2"])
	}
}

func TestTextFormat(t *testing.T) {
	var buf bytes.Buffer
	
	logger := New(Config{
		Level:     INFO,
		Format:    TextFormat,
		Output:    &buf,
		Component: "test-component",
	})
	
	logger.Info("test message", map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	})
	
	output := buf.String()
	
	if !strings.Contains(output, "INFO") {
		t.Error("Expected output to contain 'INFO'")
	}
	
	if !strings.Contains(output, "[test-component]") {
		t.Error("Expected output to contain '[test-component]'")
	}
	
	if !strings.Contains(output, "test message") {
		t.Error("Expected output to contain 'test message'")
	}
	
	if !strings.Contains(output, "key1=value1") {
		t.Error("Expected output to contain 'key1=value1'")
	}
}

func TestWithComponent(t *testing.T) {
	var buf bytes.Buffer
	
	baseLogger := New(Config{
		Level:     INFO,
		Format:    JSONFormat,
		Output:    &buf,
		Component: "base",
	})
	
	componentLogger := baseLogger.WithComponent("specific-component")
	componentLogger.Info("test message")
	
	var entry LogEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}
	
	if entry.Component != "specific-component" {
		t.Errorf("Expected component 'specific-component', got %s", entry.Component)
	}
}

func TestErrorLogging(t *testing.T) {
	var buf bytes.Buffer
	
	logger := New(Config{
		Level:  ERROR,
		Format: JSONFormat,
		Output: &buf,
	})
	
	testErr := &testError{msg: "test error"}
	logger.Error("operation failed", testErr, map[string]interface{}{
		"operation": "test_op",
	})
	
	var entry LogEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}
	
	if entry.Error != "test error" {
		t.Errorf("Expected error 'test error', got %s", entry.Error)
	}
	
	if entry.Fields["operation"] != "test_op" {
		t.Errorf("Expected operation field 'test_op', got %v", entry.Fields["operation"])
	}
}

func TestGlobalLogger(t *testing.T) {
	var buf bytes.Buffer
	
	// Save original global logger
	originalLogger := GetGlobalLogger()
	defer SetGlobalLogger(originalLogger)
	
	// Set test logger
	testLogger := New(Config{
		Level:     INFO,
		Format:    JSONFormat,
		Output:    &buf,
		Component: "global-test",
	})
	SetGlobalLogger(testLogger)
	
	// Test global convenience functions
	Info("global info message")
	Warn("global warn message")
	
	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	
	if len(lines) != 2 {
		t.Errorf("Expected 2 log lines, got %d", len(lines))
	}
	
	// Check first line
	var entry1 LogEntry
	if err := json.Unmarshal([]byte(lines[0]), &entry1); err != nil {
		t.Fatalf("Failed to parse first JSON line: %v", err)
	}
	
	if entry1.Level != "INFO" || entry1.Message != "global info message" {
		t.Errorf("First line incorrect: level=%s, message=%s", entry1.Level, entry1.Message)
	}
	
	// Check second line
	var entry2 LogEntry
	if err := json.Unmarshal([]byte(lines[1]), &entry2); err != nil {
		t.Fatalf("Failed to parse second JSON line: %v", err)
	}
	
	if entry2.Level != "WARN" || entry2.Message != "global warn message" {
		t.Errorf("Second line incorrect: level=%s, message=%s", entry2.Level, entry2.Message)
	}
}

func TestFormattedLogging(t *testing.T) {
	var buf bytes.Buffer
	
	logger := New(Config{
		Level:  INFO,
		Format: JSONFormat,
		Output: &buf,
	})
	
	logger.Infof("User %s logged in with ID %d", "john", 123)
	
	var entry LogEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}
	
	expected := "User john logged in with ID 123"
	if entry.Message != expected {
		t.Errorf("Expected message '%s', got '%s'", expected, entry.Message)
	}
}

func TestEnvironmentConfiguration(t *testing.T) {
	// Save original environment
	originalLevel := os.Getenv("LOG_LEVEL")
	originalFormat := os.Getenv("LOG_FORMAT")
	defer func() {
		os.Setenv("LOG_LEVEL", originalLevel)
		os.Setenv("LOG_FORMAT", originalFormat)
	}()
	
	// Test DEBUG level parsing
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_FORMAT", "text")
	
	level := parseLogLevel("DEBUG")
	if level != DEBUG {
		t.Errorf("Expected DEBUG level, got %v", level)
	}
	
	format := parseLogFormat("text")
	if format != TextFormat {
		t.Errorf("Expected TextFormat, got %v", format)
	}
	
	// Test case insensitivity
	level = parseLogLevel("debug")
	if level != DEBUG {
		t.Errorf("Expected DEBUG level for lowercase 'debug', got %v", level)
	}
	
	format = parseLogFormat("JSON")
	if format != JSONFormat {
		t.Errorf("Expected JSONFormat for uppercase 'JSON', got %v", format)
	}
}

func TestLogLevelString(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{DEBUG, "DEBUG"},
		{INFO, "INFO"},
		{WARN, "WARN"},
		{ERROR, "ERROR"},
		{FATAL, "FATAL"},
	}
	
	for _, test := range tests {
		if test.level.String() != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, test.level.String())
		}
	}
}

// Helper type for testing error logging
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

// Benchmark tests
func BenchmarkJSONLogging(b *testing.B) {
	var buf bytes.Buffer
	logger := New(Config{
		Level:  INFO,
		Format: JSONFormat,
		Output: &buf,
	})
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message", map[string]interface{}{
			"iteration": i,
			"benchmark": true,
		})
	}
}

func BenchmarkTextLogging(b *testing.B) {
	var buf bytes.Buffer
	logger := New(Config{
		Level:  INFO,
		Format: TextFormat,
		Output: &buf,
	})
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message", map[string]interface{}{
			"iteration": i,
			"benchmark": true,
		})
	}
}

func BenchmarkLevelFiltering(b *testing.B) {
	var buf bytes.Buffer
	logger := New(Config{
		Level:  WARN, // DEBUG messages should be filtered
		Format: JSONFormat,
		Output: &buf,
	})
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Debug("debug message that should be filtered")
	}
}
