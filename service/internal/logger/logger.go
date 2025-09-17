package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// LogLevel represents the severity level of a log message
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// LogFormat represents the output format for logs
type LogFormat int

const (
	JSONFormat LogFormat = iota
	TextFormat
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Component string                 `json:"component,omitempty"`
	Function  string                 `json:"function,omitempty"`
	File      string                 `json:"file,omitempty"`
	Line      int                    `json:"line,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Error     string                 `json:"error,omitempty"`
}

// Logger represents a structured logger
type Logger struct {
	mu        sync.RWMutex
	level     LogLevel
	format    LogFormat
	output    io.Writer
	component string
}

// Config holds logger configuration
type Config struct {
	Level     LogLevel
	Format    LogFormat
	Output    io.Writer
	Component string
}

// New creates a new logger with the given configuration
func New(config Config) *Logger {
	if config.Output == nil {
		config.Output = os.Stdout
	}
	
	return &Logger{
		level:     config.Level,
		format:    config.Format,
		output:    config.Output,
		component: config.Component,
	}
}

// NewDefault creates a logger with default configuration
func NewDefault() *Logger {
	return New(Config{
		Level:  INFO,
		Format: JSONFormat,
		Output: os.Stdout,
	})
}

// WithComponent creates a new logger with the specified component name
func (l *Logger) WithComponent(component string) *Logger {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	return &Logger{
		level:     l.level,
		format:    l.format,
		output:    l.output,
		component: component,
	}
}

// SetLevel sets the minimum log level
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// SetFormat sets the log output format
func (l *Logger) SetFormat(format LogFormat) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.format = format
}

// log is the internal logging method
func (l *Logger) log(level LogLevel, message string, fields map[string]interface{}, err error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	// Check if we should log this level
	if level < l.level {
		return
	}
	
	// Get caller information
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		file = "unknown"
		line = 0
	}
	
	// Extract function name from the call stack
	pc, _, _, ok := runtime.Caller(3)
	var funcName string
	if ok {
		fn := runtime.FuncForPC(pc)
		if fn != nil {
			funcName = fn.Name()
			// Extract just the function name without package path
			if lastSlash := strings.LastIndex(funcName, "/"); lastSlash >= 0 {
				funcName = funcName[lastSlash+1:]
			}
		}
	}
	
	// Create log entry
	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     level.String(),
		Message:   message,
		Component: l.component,
		Function:  funcName,
		File:      file,
		Line:      line,
		Fields:    fields,
	}
	
	if err != nil {
		entry.Error = err.Error()
	}
	
	// Format and write the log entry
	var output string
	switch l.format {
	case JSONFormat:
		jsonBytes, _ := json.Marshal(entry)
		output = string(jsonBytes) + "\n"
	case TextFormat:
		output = l.formatText(entry)
	default:
		output = l.formatText(entry)
	}
	
	l.output.Write([]byte(output))
	
	// If this is a fatal log, exit the program
	if level == FATAL {
		os.Exit(1)
	}
}

// formatText formats a log entry as human-readable text
func (l *Logger) formatText(entry LogEntry) string {
	var parts []string
	
	// Timestamp and level
	parts = append(parts, fmt.Sprintf("[%s] %s", entry.Timestamp, entry.Level))
	
	// Component
	if entry.Component != "" {
		parts = append(parts, fmt.Sprintf("[%s]", entry.Component))
	}
	
	// Message
	parts = append(parts, entry.Message)
	
	// Fields
	if len(entry.Fields) > 0 {
		fieldParts := make([]string, 0, len(entry.Fields))
		for k, v := range entry.Fields {
			fieldParts = append(fieldParts, fmt.Sprintf("%s=%v", k, v))
		}
		parts = append(parts, fmt.Sprintf("fields={%s}", strings.Join(fieldParts, ", ")))
	}
	
	// Error
	if entry.Error != "" {
		parts = append(parts, fmt.Sprintf("error=%s", entry.Error))
	}
	
	// File and line (for debug builds)
	if entry.File != "" && entry.Line > 0 {
		parts = append(parts, fmt.Sprintf("(%s:%d)", entry.File, entry.Line))
	}
	
	return strings.Join(parts, " ") + "\n"
}

// Debug logs a debug message
func (l *Logger) Debug(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(DEBUG, message, f, nil)
}

// Info logs an info message
func (l *Logger) Info(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(INFO, message, f, nil)
}

// Warn logs a warning message
func (l *Logger) Warn(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(WARN, message, f, nil)
}

// Error logs an error message
func (l *Logger) Error(message string, err error, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(ERROR, message, f, err)
}

// Fatal logs a fatal message and exits the program
func (l *Logger) Fatal(message string, err error, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(FATAL, message, f, err)
}

// Convenience methods with formatted messages

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.Debug(fmt.Sprintf(format, args...))
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.Info(fmt.Sprintf(format, args...))
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.Warn(fmt.Sprintf(format, args...))
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Error(fmt.Sprintf(format, args...), nil)
}

// Fatalf logs a formatted fatal message and exits
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.Fatal(fmt.Sprintf(format, args...), nil)
}
