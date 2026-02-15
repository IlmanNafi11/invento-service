package middleware

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"invento-service/internal/logger"

	"github.com/gofiber/fiber/v2"
)

const (
	// RequestBodyLoggingKey enables/disables request body logging
	RequestBodyLoggingKey = "log_body"
	// MaxBodySize is the maximum body size to log (in bytes)
	MaxBodySize = 1024
)

// LoggingConfig holds configuration for the logging middleware
type LoggingConfig struct {
	Logger          *logger.Logger
	LogRequestBody  bool
	SanitizePaths   []string
	SensitiveFields []string
	MaxBodySize     int
	SkipPaths       []string
}

// DefaultLoggingConfig returns default logging configuration
func DefaultLoggingConfig(log *logger.Logger) *LoggingConfig {
	return &LoggingConfig{
		Logger:          log,
		LogRequestBody:  false,
		SanitizePaths:   []string{"/auth/login", "/auth/refresh", "/auth/register"},
		SensitiveFields: []string{"password", "token", "otp", "secret"},
		MaxBodySize:     MaxBodySize,
		SkipPaths:       []string{"/health", "/uploads"},
	}
}

// RequestLogger creates a middleware that logs all HTTP requests with duration
// and relevant information including request ID, user ID, endpoint, method, and status
func RequestLogger(log *logger.Logger) fiber.Handler {
	config := DefaultLoggingConfig(log)
	return RequestLoggerWithConfig(config)
}

// RequestLoggerWithConfig creates a logging middleware with custom configuration
func RequestLoggerWithConfig(config *LoggingConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip logging for specified paths
		if shouldSkipPath(c.Path(), config.SkipPaths) {
			return c.Next()
		}

		start := time.Now()
		requestID := GetRequestID(c)
		userID := logger.GetUserIDFromContext(c)

		// Create a logger with request ID context
		reqLogger := config.Logger.WithRequestID(requestID)

		// Capture request body if enabled
		var requestBody string
		if config.LogRequestBody && c.Method() != "GET" && shouldLogBody(c.Path(), config.SanitizePaths) {
			requestBody = captureRequestBody(c, config.MaxBodySize, config.SensitiveFields)
		}

		// Continue with the request
		err := c.Next()

		// Calculate request duration
		duration := time.Since(start)

		// Log the request completion
		statusCode := c.Response().StatusCode()
		fields := map[string]interface{}{
			"method":     c.Method(),
			"path":       c.Path(),
			"status":     statusCode,
			"duration":   fmt.Sprintf("%dms", duration.Milliseconds()),
			"ip":         c.IP(),
			"user_agent": c.Get("User-Agent"),
		}

		if requestID != "" {
			fields["request_id"] = requestID
		}

		if userID != "" {
			fields["user_id"] = userID
		}

		if requestBody != "" {
			fields["request_body"] = requestBody
		}

		// Log based on status code
		if statusCode >= 500 {
			reqLogger.Error(fmt.Sprintf("%s %s completed with server error", c.Method(), c.Path()), fields)
		} else if statusCode >= 400 {
			reqLogger.Warn(fmt.Sprintf("%s %s completed with client error", c.Method(), c.Path()), fields)
		} else {
			reqLogger.Info(fmt.Sprintf("%s %s completed", c.Method(), c.Path()), fields)
		}

		// Handle any errors from next middleware
		if err != nil {
			reqLogger.Error("Request handling error", map[string]interface{}{
				"error": err.Error(),
				"path":  c.Path(),
			})
			return err
		}

		return nil
	}
}

// captureRequestBody captures and sanitizes the request body
func captureRequestBody(c *fiber.Ctx, maxSize int, sensitiveFields []string) string {
	body := c.Body()

	if len(body) == 0 {
		return ""
	}

	// Truncate if too large
	if len(body) > maxSize {
		body = body[:maxSize]
	}

	// Sanitize sensitive fields
	bodyStr := string(body)
	for _, field := range sensitiveFields {
		bodyStr = sanitizeField(bodyStr, field)
	}

	return bodyStr
}

// sanitizeField replaces sensitive field values with *** in JSON body
func sanitizeField(body, field string) string {
	// Simple sanitization - for production, use a JSON parser
	search := fmt.Sprintf("\"%s\":", field)
	idx := 0
	result := ""

	for {
		i := indexOf(body, search, idx)
		if i == -1 {
			result += body[idx:]
			break
		}
		result += body[idx : i+len(search)]

		// Find the value (skip whitespace and quotes)
		j := i + len(search)
		for j < len(body) && (body[j] == ' ' || body[j] == '\t' || body[j] == '\n') {
			j++
		}

		if j < len(body) && body[j] == '"' {
			// String value - skip it
			j++
			for j < len(body) && body[j] != '"' {
				j++
			}
			result += "***\""
			idx = j
		} else if j < len(body) && (body[j] == 't' || body[j] == 'f' || body[j] == 'n') {
			// Boolean or null value
			result += "***"
			idx = j + 4
		} else {
			// Number or other value
			result += "***"
			for j < len(body) && body[j] != ',' && body[j] != '}' && body[j] != ']' {
				j++
			}
			idx = j
		}
	}

	return result
}

// indexOf finds the index of a substring starting from a given position
func indexOf(s, substr string, start int) int {
	if start >= len(s) {
		return -1
	}
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// shouldSkipPath checks if a path should be skipped from logging
func shouldSkipPath(path string, skipPaths []string) bool {
	for _, skipPath := range skipPaths {
		if path == skipPath {
			return true
		}
	}
	return false
}

// shouldLogBody checks if request body should be logged for a given path
func shouldLogBody(path string, sanitizePaths []string) bool {
	// Log body for all paths except those in sanitize list
	for _, sp := range sanitizePaths {
		if path == sp {
			return false
		}
	}
	return true
}

// captureRequestBodyFromReader captures request body from io.Reader
func captureRequestBodyFromReader(reader io.Reader, maxSize int) ([]byte, error) {
	buf := &bytes.Buffer{}
	limitedReader := io.LimitReader(reader, int64(maxSize))
	_, err := buf.ReadFrom(limitedReader)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
