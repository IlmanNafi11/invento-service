package middleware

import (
	"bytes"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"invento-service/internal/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

// TestRequestLogger_SkipHealthPath tests that health check path is not logged
func TestRequestLogger_SkipHealthPath(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer
	log := logger.NewLogger(logger.INFO, logger.TextFormat)
	log.SetOutput(&buf)

	app := fiber.New()
	app.Use(RequestID())
	app.Use(RequestLogger(log))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(200).SendString("OK")
	})

	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	// Buffer should be empty since /health is in skip paths
	require.Empty(t, buf.String(), "Health path should not be logged")
}

// TestRequestLogger_LogSuccessfulRequest tests that successful 2xx requests are logged
func TestRequestLogger_LogSuccessfulRequest(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLogger(logger.INFO, logger.TextFormat)
	log.SetOutput(&buf)

	app := fiber.New()
	app.Use(RequestID())
	app.Use(RequestLogger(log))

	app.Get("/api/test", func(c *fiber.Ctx) error {
		return c.Status(200).SendString("OK")
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	logOutput := buf.String()
	require.Contains(t, logOutput, "GET /api/test completed", "Should log successful request")
	require.Contains(t, logOutput, "status=200", "Should include status code")
	require.Contains(t, logOutput, "[INFO]", "Should be INFO level")
}

// TestRequestLogger_LogClientError tests that 4xx client errors are logged with WARN level
func TestRequestLogger_LogClientError(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLogger(logger.INFO, logger.TextFormat)
	log.SetOutput(&buf)

	app := fiber.New()
	app.Use(RequestID())
	app.Use(RequestLogger(log))

	app.Get("/api/notfound", func(c *fiber.Ctx) error {
		return c.Status(404).SendString("Not Found")
	})

	req := httptest.NewRequest("GET", "/api/notfound", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	require.Equal(t, 404, resp.StatusCode)

	logOutput := buf.String()
	require.Contains(t, logOutput, "GET /api/notfound completed with client error", "Should log client error")
	require.Contains(t, logOutput, "status=404", "Should include status code")
	require.Contains(t, logOutput, "[WARN]", "Should be WARN level")
}

// TestRequestLogger_LogServerError tests that 5xx server errors are logged with ERROR level
func TestRequestLogger_LogServerError(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLogger(logger.INFO, logger.TextFormat)
	log.SetOutput(&buf)

	app := fiber.New()
	app.Use(RequestID())
	app.Use(RequestLogger(log))

	app.Get("/api/error", func(c *fiber.Ctx) error {
		return c.Status(500).SendString("Internal Server Error")
	})

	req := httptest.NewRequest("GET", "/api/error", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	require.Equal(t, 500, resp.StatusCode)

	logOutput := buf.String()
	require.Contains(t, logOutput, "GET /api/error completed with server error", "Should log server error")
	require.Contains(t, logOutput, "status=500", "Should include status code")
	require.Contains(t, logOutput, "[ERROR]", "Should be ERROR level")
}

// TestRequestLogger_IncludesRequestID tests that request_id is included in logs
func TestRequestLogger_IncludesRequestID(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLogger(logger.INFO, logger.TextFormat)
	log.SetOutput(&buf)

	app := fiber.New()
	app.Use(RequestID())
	app.Use(RequestLogger(log))

	var capturedRequestID string
	app.Get("/api/test", func(c *fiber.Ctx) error {
		capturedRequestID = GetRequestID(c)
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	require.NotEmpty(t, capturedRequestID)

	logOutput := buf.String()
	require.Contains(t, logOutput, "req:"+capturedRequestID, "Should include request_id in logs")
}

// TestRequestLogger_IncludesUserID tests that user_id is included in logs when present
func TestRequestLogger_IncludesUserID(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLogger(logger.INFO, logger.TextFormat)
	log.SetOutput(&buf)

	app := fiber.New()
	app.Use(RequestID())
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", "12345")
		return c.Next()
	})
	app.Use(RequestLogger(log))

	app.Get("/api/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	logOutput := buf.String()
	require.Contains(t, logOutput, "user_id=12345", "Should include user_id in logs")
}

// TestRequestLogger_SanitizesPassword tests that password field is sanitized in logs
func TestRequestLogger_SanitizesPassword(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLogger(logger.INFO, logger.TextFormat)
	log.SetOutput(&buf)

	config := &LoggingConfig{
		Logger:          log,
		LogRequestBody:  true,
		SanitizePaths:   []string{},
		SensitiveFields: []string{"password"},
		MaxBodySize:     MaxBodySize,
		SkipPaths:       []string{},
	}

	app := fiber.New()
	app.Use(RequestID())
	app.Use(RequestLoggerWithConfig(config))

	app.Post("/api/login", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

 requestBody := []byte(`{"email":"test@example.com","password":"secret123"}`)
	req := httptest.NewRequest("POST", "/api/login", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	logOutput := buf.String()
	require.Contains(t, logOutput, "request_body=", "Should include request body")
	require.NotContains(t, logOutput, "secret123", "Should not contain actual password")
	require.Contains(t, logOutput, "***", "Should contain sanitized placeholder")
}

// TestRequestLogger_SanitizesMultipleSensitiveFields tests sanitization of multiple fields
func TestRequestLogger_SanitizesMultipleSensitiveFields(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLogger(logger.INFO, logger.TextFormat)
	log.SetOutput(&buf)

	config := &LoggingConfig{
		Logger:          log,
		LogRequestBody:  true,
		SanitizePaths:   []string{},
		SensitiveFields: []string{"password", "token", "otp", "secret"},
		MaxBodySize:     MaxBodySize,
		SkipPaths:       []string{},
	}

	app := fiber.New()
	app.Use(RequestID())
	app.Use(RequestLoggerWithConfig(config))

	app.Post("/api/auth", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	requestBody := []byte(`{
		"email":"test@example.com",
		"password":"secret123",
		"token":"abc123xyz",
		"otp":"123456",
		"secret":"mysecret"
	}`)
	req := httptest.NewRequest("POST", "/api/auth", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	logOutput := buf.String()
	require.NotContains(t, logOutput, "secret123", "Should not contain password")
	require.NotContains(t, logOutput, "abc123xyz", "Should not contain token")
	require.NotContains(t, logOutput, "123456", "Should not contain otp")
	require.NotContains(t, logOutput, "mysecret", "Should not contain secret")
}

// TestRequestLogger_TruncatesLargeBody tests that large request bodies are truncated
func TestRequestLogger_TruncatesLargeBody(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLogger(logger.INFO, logger.TextFormat)
	log.SetOutput(&buf)

	config := &LoggingConfig{
		Logger:          log,
		LogRequestBody:  true,
		SanitizePaths:   []string{},
		SensitiveFields: []string{},
		MaxBodySize:     100,
		SkipPaths:       []string{},
	}

	app := fiber.New()
	app.Use(RequestID())
	app.Use(RequestLoggerWithConfig(config))

	app.Post("/api/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Create a body larger than MaxBodySize
	largeBody := strings.Repeat("a", 200)
	requestBody := []byte(`{"data":"` + largeBody + `"}`)

	req := httptest.NewRequest("POST", "/api/test", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	logOutput := buf.String()
	// The logged body should be truncated to around MaxBodySize
	require.Contains(t, logOutput, "request_body=", "Should include truncated request body")
}

// TestRequestLogger_HandlesEmptyBody tests that empty bodies are handled gracefully
func TestRequestLogger_HandlesEmptyBody(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLogger(logger.INFO, logger.TextFormat)
	log.SetOutput(&buf)

	config := &LoggingConfig{
		Logger:          log,
		LogRequestBody:  true,
		SanitizePaths:   []string{},
		SensitiveFields: []string{},
		MaxBodySize:     MaxBodySize,
		SkipPaths:       []string{},
	}

	app := fiber.New()
	app.Use(RequestID())
	app.Use(RequestLoggerWithConfig(config))

	app.Post("/api/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("POST", "/api/test", bytes.NewReader([]byte{}))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	// Should not crash with empty body
	logOutput := buf.String()
	require.Contains(t, logOutput, "POST /api/test completed", "Should log request even with empty body")
}

// TestRequestLogger_SkipsUploadsPath tests that uploads path is skipped
func TestRequestLogger_SkipsUploadsPath(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLogger(logger.INFO, logger.TextFormat)
	log.SetOutput(&buf)

	app := fiber.New()
	app.Use(RequestID())
	app.Use(RequestLogger(log))

	app.Post("/uploads", func(c *fiber.Ctx) error {
		return c.SendString("Uploaded")
	})

	req := httptest.NewRequest("POST", "/uploads", bytes.NewReader([]byte("test")))
	resp, err := app.Test(req)

	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	// Buffer should be empty since /uploads is in skip paths
	require.Empty(t, buf.String(), "Uploads path should not be logged")
}

// TestRequestLogger_IncludesDuration tests that request duration is logged
func TestRequestLogger_IncludesDuration(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLogger(logger.INFO, logger.TextFormat)
	log.SetOutput(&buf)

	app := fiber.New()
	app.Use(RequestID())
	app.Use(RequestLogger(log))

	app.Get("/api/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	logOutput := buf.String()
	require.Contains(t, logOutput, "duration=", "Should include request duration")
	require.Contains(t, logOutput, "ms", "Duration should be in milliseconds")
}

// TestRequestLogger_IncludesMethodAndPath tests that HTTP method and path are logged
func TestRequestLogger_IncludesMethodAndPath(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLogger(logger.INFO, logger.TextFormat)
	log.SetOutput(&buf)

	app := fiber.New()
	app.Use(RequestID())
	app.Use(RequestLogger(log))

	app.Post("/api/users", func(c *fiber.Ctx) error {
		return c.SendString("Created")
	})

	req := httptest.NewRequest("POST", "/api/users", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	logOutput := buf.String()
	require.Contains(t, logOutput, "POST", "Should include HTTP method")
	require.Contains(t, logOutput, "/api/users", "Should include request path")
}

// TestDefaultLoggingConfig tests that default logging config is created correctly
func TestDefaultLoggingConfig(t *testing.T) {
	log := logger.NewLogger(logger.INFO, logger.TextFormat)
	config := DefaultLoggingConfig(log)

	require.NotNil(t, config)
	require.Equal(t, log, config.Logger)
	require.False(t, config.LogRequestBody)
	require.Equal(t, MaxBodySize, config.MaxBodySize)
	require.Contains(t, config.SkipPaths, "/health")
	require.Contains(t, config.SkipPaths, "/uploads")
	require.Contains(t, config.SanitizePaths, "/auth/login")
	require.Contains(t, config.SensitiveFields, "password")
	require.Contains(t, config.SensitiveFields, "token")
}

// TestCaptureRequestBody tests capturing and sanitizing request body
func TestCaptureRequestBody(t *testing.T) {
	app := fiber.New()
	app.Post("/test", func(c *fiber.Ctx) error {
		body := captureRequestBody(c, 1024, []string{"password", "secret"})
		return c.SendString(body)
	})

	requestBody := []byte(`{"email":"test@example.com","password":"secret123","secret":"mysecret"}`)
	req := httptest.NewRequest("POST", "/test", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	require.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	bodyStr := string(body)
	require.NotContains(t, bodyStr, "secret123", "Should sanitize password")
	require.NotContains(t, bodyStr, "mysecret", "Should sanitize secret")
	require.Contains(t, bodyStr, "test@example.com", "Should keep non-sensitive fields")
}

// TestSanitizeField tests sanitization of individual fields
func TestSanitizeField(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		field    string
		contains []string
		notContains []string
	}{
		{
			name:  "sanitize password field",
			body:  `{"password":"secret123","username":"test"}`,
			field: "password",
			contains: []string{"***\""},
			notContains: []string{"secret123"},
		},
		{
			name:  "sanitize token field",
			body:  `{"token":"abc123xyz","user_id":"123"}`,
			field: "token",
			contains: []string{"***\""},
			notContains: []string{"abc123xyz"},
		},
		{
			name:  "sanitize multiple occurrences",
			body:  `{"password":"pass1","confirm_password":"pass2"}`,
			field: "password",
			contains: []string{"***\""},
			notContains: []string{"pass1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeField(tt.body, tt.field)

			for _, substr := range tt.contains {
				require.Contains(t, result, substr)
			}
			for _, substr := range tt.notContains {
				require.NotContains(t, result, substr)
			}
		})
	}
}

// TestIndexOf tests finding substring index
func TestIndexOf(t *testing.T) {
	tests := []struct {
		name      string
		s         string
		substr    string
		start     int
		expected  int
	}{
		{
			name:     "find at beginning",
			s:        "hello world",
			substr:   "hello",
			start:    0,
			expected: 0,
		},
		{
			name:     "find in middle",
			s:        "hello world",
			substr:   "world",
			start:    0,
			expected: 6,
		},
		{
			name:     "find with start offset",
			s:        "hello hello world",
			substr:   "hello",
			start:    2,
			expected: 6,
		},
		{
			name:     "not found",
			s:        "hello world",
			substr:   "foo",
			start:    0,
			expected: -1,
		},
		{
			name:     "start beyond length",
			s:        "hello",
			substr:   "hello",
			start:    10,
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := indexOf(tt.s, tt.substr, tt.start)
			require.Equal(t, tt.expected, result)
		})
	}
}

// TestShouldSkipPath tests path skipping logic
func TestShouldSkipPath(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		skipPaths []string
		expected  bool
	}{
		{
			name:      "skip health path",
			path:      "/health",
			skipPaths: []string{"/health", "/uploads"},
			expected:  true,
		},
		{
			name:      "skip uploads path",
			path:      "/uploads",
			skipPaths: []string{"/health", "/uploads"},
			expected:  true,
		},
		{
			name:      "don't skip api path",
			path:      "/api/users",
			skipPaths: []string{"/health", "/uploads"},
			expected:  false,
		},
		{
			name:      "empty skip list",
			path:      "/health",
			skipPaths: []string{},
			expected:  false,
		},
		{
			name:      "exact match only",
			path:      "/health/extra",
			skipPaths: []string{"/health"},
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldSkipPath(tt.path, tt.skipPaths)
			require.Equal(t, tt.expected, result)
		})
	}
}

// TestShouldLogBody tests body logging path filtering
func TestShouldLogBody(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		sanitizePaths []string
		expected      bool
	}{
		{
			name:          "log regular path",
			path:          "/api/users",
			sanitizePaths: []string{"/auth/login", "/auth/refresh"},
			expected:      true,
		},
		{
			name:          "don't log login path",
			path:          "/auth/login",
			sanitizePaths: []string{"/auth/login", "/auth/refresh"},
			expected:      false,
		},
		{
			name:          "don't log refresh path",
			path:          "/auth/refresh",
			sanitizePaths: []string{"/auth/login", "/auth/refresh"},
			expected:      false,
		},
		{
			name:          "empty sanitize list",
			path:          "/api/users",
			sanitizePaths: []string{},
			expected:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldLogBody(tt.path, tt.sanitizePaths)
			require.Equal(t, tt.expected, result)
		})
	}
}

// TestCaptureRequestBodyFromReader tests capturing body from io.Reader
func TestCaptureRequestBodyFromReader(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		maxSize     int
		expectedLen int
		expectError bool
	}{
		{
			name:        "capture small body",
			input:       "hello world",
			maxSize:     100,
			expectedLen: 11,
			expectError: false,
		},
		{
			name:        "truncate large body",
			input:       strings.Repeat("a", 200),
			maxSize:     100,
			expectedLen: 100,
			expectError: false,
		},
		{
			name:        "empty body",
			input:       "",
			maxSize:     100,
			expectedLen: 0,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			result, err := captureRequestBodyFromReader(reader, tt.maxSize)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Len(t, result, tt.expectedLen)
			}
		})
	}
}

// TestRequestLogger_WithConfig tests custom logging configuration
func TestRequestLogger_WithConfig(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLogger(logger.INFO, logger.TextFormat)
	log.SetOutput(&buf)

	config := &LoggingConfig{
		Logger:          log,
		LogRequestBody:  true,
		SanitizePaths:   []string{"/custom/sensitive"},
		SensitiveFields: []string{"api_key"},
		MaxBodySize:     500,
		SkipPaths:       []string{"/custom/skip"},
	}

	app := fiber.New()
	app.Use(RequestID())
	app.Use(RequestLoggerWithConfig(config))

	app.Get("/custom/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/custom/test", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	logOutput := buf.String()
	require.Contains(t, logOutput, "GET /custom/test completed", "Should use custom config")
}

// TestRequestLogger_DoesNotLogGetRequestBody tests that GET requests don't log body
func TestRequestLogger_DoesNotLogGetRequestBody(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLogger(logger.INFO, logger.TextFormat)
	log.SetOutput(&buf)

	config := &LoggingConfig{
		Logger:          log,
		LogRequestBody:  true,
		SanitizePaths:   []string{},
		SensitiveFields: []string{},
		MaxBodySize:     MaxBodySize,
		SkipPaths:       []string{},
	}

	app := fiber.New()
	app.Use(RequestID())
	app.Use(RequestLoggerWithConfig(config))

	app.Get("/api/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	logOutput := buf.String()
	require.NotContains(t, logOutput, "request_body=", "GET requests should not log body")
}
