package logger

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
)

// LogLevel represents the severity level of a log entry
type LogLevel string

const (
	DEBUG LogLevel = "DEBUG"
	INFO  LogLevel = "INFO"
	WARN  LogLevel = "WARN"
	ERROR LogLevel = "ERROR"
	FATAL LogLevel = "FATAL"
)

// LogFormat represents the output format for logs
type LogFormat string

const (
	JSONFormat LogFormat = "json"
	TextFormat LogFormat = "text"
)

// Logger is a structured logger that supports both JSON and text formats
type Logger struct {
	minLevel  LogLevel
	format    LogFormat
	output    io.Writer
	requestID string
}

// NewLogger creates a new logger instance with the specified configuration
func NewLogger(level LogLevel, format LogFormat) *Logger {
	return &Logger{
		minLevel: level,
		format:   format,
		output:   os.Stdout,
	}
}

// NewDefaultLogger creates a logger with default settings (INFO level, text format in development)
func NewDefaultLogger(isDevelopment bool) *Logger {
	format := TextFormat
	if !isDevelopment {
		format = JSONFormat
	}
	return NewLogger(INFO, format)
}

// SetOutput sets the output destination for the logger
func (l *Logger) SetOutput(w io.Writer) {
	l.output = w
}

// SetRequestID sets the request ID for the logger context
func (l *Logger) SetRequestID(requestID string) {
	l.requestID = requestID
}

// shouldLog determines whether a message with the given level should be logged
func (l *Logger) shouldLog(level LogLevel) bool {
	levels := map[LogLevel]int{
		DEBUG: 0,
		INFO:  1,
		WARN:  2,
		ERROR: 3,
		FATAL: 4,
	}

	currentLevel, ok := levels[l.minLevel]
	if !ok {
		return true
	}

	msgLevel, ok := levels[level]
	if !ok {
		return true
	}

	return msgLevel >= currentLevel
}

// log is the internal logging method that formats and writes log entries
func (l *Logger) log(level LogLevel, message string, fields map[string]interface{}) {
	if !l.shouldLog(level) {
		return
	}

	timestamp := time.Now().Format(time.RFC3339)
	logData := map[string]interface{}{
		"timestamp":  timestamp,
		"level":      string(level),
		"message":    message,
		"request_id": l.requestID,
	}

	// Add additional fields
	for k, v := range fields {
		logData[k] = v
	}

	if l.format == JSONFormat {
		l.logJSON(logData)
	} else {
		l.logText(logData)
	}
}

// logJSON outputs logs in JSON format
func (l *Logger) logJSON(data map[string]interface{}) {
	// Simple JSON formatting (for production use with log aggregators)
	fmt.Fprintf(l.output, "{\"timestamp\":\"%s\",\"level\":\"%s\",\"message\":\"%s\",\"request_id\":\"%s\"",
		data["timestamp"], data["level"], data["message"], data["request_id"])

	for k, v := range data {
		if k != "timestamp" && k != "level" && k != "message" && k != "request_id" {
			fmt.Fprintf(l.output, ",\"%s\":\"%v\"", k, v)
		}
	}
	fmt.Fprintln(l.output, "}")
}

// logText outputs logs in human-readable text format
func (l *Logger) logText(data map[string]interface{}) {
	timestamp := data["timestamp"]
	level := data["level"]
	message := data["message"]
	requestID := data["request_id"]

	prefix := fmt.Sprintf("[%s] [%s]", timestamp, level)
	if requestID != nil && requestID != "" {
		prefix += fmt.Sprintf(" [req:%s]", requestID)
	}

	fmt.Fprintf(l.output, "%s %s", prefix, message)

	// Append additional fields
	for k, v := range data {
		if k != "timestamp" && k != "level" && k != "message" && k != "request_id" {
			fmt.Fprintf(l.output, " %s=%v", k, v)
		}
	}
	fmt.Fprintln(l.output)
}

// Debug logs a debug message
func (l *Logger) Debug(message string, fields ...map[string]interface{}) {
	extraFields := map[string]interface{}{}
	if len(fields) > 0 {
		extraFields = fields[0]
	}
	l.log(DEBUG, message, extraFields)
}

// Debugf logs a debug message with formatting
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.Debug(fmt.Sprintf(format, args...))
}

// Info logs an info message
func (l *Logger) Info(message string, fields ...map[string]interface{}) {
	extraFields := map[string]interface{}{}
	if len(fields) > 0 {
		extraFields = fields[0]
	}
	l.log(INFO, message, extraFields)
}

// Infof logs an info message with formatting
func (l *Logger) Infof(format string, args ...interface{}) {
	l.Info(fmt.Sprintf(format, args...))
}

// Warn logs a warning message
func (l *Logger) Warn(message string, fields ...map[string]interface{}) {
	extraFields := map[string]interface{}{}
	if len(fields) > 0 {
		extraFields = fields[0]
	}
	l.log(WARN, message, extraFields)
}

// Warnf logs a warning message with formatting
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.Warn(fmt.Sprintf(format, args...))
}

// Error logs an error message
func (l *Logger) Error(message string, fields ...map[string]interface{}) {
	extraFields := map[string]interface{}{}
	if len(fields) > 0 {
		extraFields = fields[0]
	}
	l.log(ERROR, message, extraFields)
}

// Errorf logs an error message with formatting
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Error(fmt.Sprintf(format, args...))
}

// Fatal logs a fatal message and exits the application
func (l *Logger) Fatal(message string, fields ...map[string]interface{}) {
	extraFields := map[string]interface{}{}
	if len(fields) > 0 {
		extraFields = fields[0]
	}
	l.log(FATAL, message, extraFields)
	os.Exit(1)
}

// Fatalf logs a fatal message with formatting and exits
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.Fatal(fmt.Sprintf(format, args...))
}

// WithRequestID returns a new logger instance with the specified request ID
func (l *Logger) WithRequestID(requestID string) *Logger {
	newLogger := &Logger{
		minLevel:  l.minLevel,
		format:    l.format,
		output:    l.output,
		requestID: requestID,
	}
	return newLogger
}

// GetRequestIDFromContext retrieves the request ID from the Fiber context
func GetRequestIDFromContext(c *fiber.Ctx) string {
	if requestID := c.Locals("request_id"); requestID != nil {
		if rid, ok := requestID.(string); ok {
			return rid
		}
	}
	return ""
}

// GetUserIDFromContext retrieves the user ID from the Fiber context
func GetUserIDFromContext(c *fiber.Ctx) string {
	if userID := c.Locals("user_id"); userID != nil {
		if uid, ok := userID.(string); ok {
			return uid
		}
		if uid, ok := userID.(uint); ok {
			return fmt.Sprintf("%d", uid)
		}
	}
	return ""
}

// ParseLogLevel parses a string to LogLevel, defaulting to INFO if invalid
func ParseLogLevel(level string) LogLevel {
	switch level {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	case "FATAL":
		return FATAL
	default:
		return INFO
	}
}

// ParseLogFormat parses a string to LogFormat, defaulting to text if invalid
func ParseLogFormat(format string) LogFormat {
	switch format {
	case "json":
		return JSONFormat
	case "text":
		return TextFormat
	default:
		return TextFormat
	}
}
