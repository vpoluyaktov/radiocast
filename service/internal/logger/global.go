package logger

import (
	"os"
	"strings"
)

var (
	// Global logger instance
	globalLogger *Logger
)

func init() {
	// Initialize with default configuration
	globalLogger = NewDefault()
	
	// Configure from environment variables
	configureFromEnv()
}

// configureFromEnv configures the global logger from environment variables
func configureFromEnv() {
	// Set log level from environment
	if levelStr := os.Getenv("LOG_LEVEL"); levelStr != "" {
		if level := parseLogLevel(levelStr); level != -1 {
			globalLogger.SetLevel(level)
		}
	}
	
	// Set log format from environment
	if formatStr := os.Getenv("LOG_FORMAT"); formatStr != "" {
		if format := parseLogFormat(formatStr); format != -1 {
			globalLogger.SetFormat(format)
		}
	}
}

// parseLogLevel parses a log level string
func parseLogLevel(level string) LogLevel {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN", "WARNING":
		return WARN
	case "ERROR":
		return ERROR
	case "FATAL":
		return FATAL
	default:
		return -1
	}
}

// parseLogFormat parses a log format string
func parseLogFormat(format string) LogFormat {
	switch strings.ToLower(format) {
	case "json":
		return JSONFormat
	case "text":
		return TextFormat
	default:
		return -1
	}
}

// GetGlobalLogger returns the global logger instance
func GetGlobalLogger() *Logger {
	return globalLogger
}

// SetGlobalLogger sets the global logger instance
func SetGlobalLogger(logger *Logger) {
	globalLogger = logger
}

// Global convenience functions that use the global logger

// Debug logs a debug message using the global logger
func Debug(message string, fields ...map[string]interface{}) {
	globalLogger.Debug(message, fields...)
}

// Info logs an info message using the global logger
func Info(message string, fields ...map[string]interface{}) {
	globalLogger.Info(message, fields...)
}

// Warn logs a warning message using the global logger
func Warn(message string, fields ...map[string]interface{}) {
	globalLogger.Warn(message, fields...)
}

// Error logs an error message using the global logger
func Error(message string, err error, fields ...map[string]interface{}) {
	globalLogger.Error(message, err, fields...)
}

// Fatal logs a fatal message using the global logger and exits
func Fatal(message string, err error, fields ...map[string]interface{}) {
	globalLogger.Fatal(message, err, fields...)
}

// Debugf logs a formatted debug message using the global logger
func Debugf(format string, args ...interface{}) {
	globalLogger.Debugf(format, args...)
}

// Infof logs a formatted info message using the global logger
func Infof(format string, args ...interface{}) {
	globalLogger.Infof(format, args...)
}

// Warnf logs a formatted warning message using the global logger
func Warnf(format string, args ...interface{}) {
	globalLogger.Warnf(format, args...)
}

// Errorf logs a formatted error message using the global logger
func Errorf(format string, args ...interface{}) {
	globalLogger.Errorf(format, args...)
}

// Fatalf logs a formatted fatal message using the global logger and exits
func Fatalf(format string, args ...interface{}) {
	globalLogger.Fatalf(format, args...)
}
