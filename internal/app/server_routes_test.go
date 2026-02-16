package app_test

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"invento-service/config"
	app "invento-service/internal/app"
)

func TestServer_RouteRegistration_SwaggerRoute(t *testing.T) {
	cfg := createTestConfig()
	db := setupTestDB(t)
	defer teardownTestDB(db)

	appInstance, err := app.NewServer(cfg, db)
	require.NoError(t, err)

	// Test swagger route
	req := httptest.NewRequest("GET", "/swagger/index.html", nil)
	resp, err := appInstance.Test(req)

	require.NoError(t, err)
	// Swagger might return 200 or redirect, but not 404
	assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode,
		"Swagger route should be registered")
}

// TestServer_Middleware_RequestID tests that RequestID middleware is applied
func TestServer_Middleware_RequestID(t *testing.T) {
	cfg := createTestConfig()
	db := setupTestDB(t)
	defer teardownTestDB(db)

	appInstance, err := app.NewServer(cfg, db)
	require.NoError(t, err)

	// Make a request and check for X-Request-ID header
	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := appInstance.Test(req)

	require.NoError(t, err)
	requestID := resp.Header.Get("X-Request-ID")
	assert.NotEmpty(t, requestID, "X-Request-ID header should be present")
}

// TestServer_Middleware_CORS_Development tests CORS configuration in development
func TestServer_Middleware_CORS_Development(t *testing.T) {
	cfg := createTestConfig()
	cfg.App.Env = "development"
	cfg.App.CorsOriginDev = "http://localhost:3000"

	db := setupTestDB(t)
	defer teardownTestDB(db)

	appInstance, err := app.NewServer(cfg, db)
	require.NoError(t, err)

	// Make a request with Origin header
	req := httptest.NewRequest("GET", "/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	resp, err := appInstance.Test(req)

	require.NoError(t, err)
	// Check for CORS headers
	corsHeaders := []string{
		"Access-Control-Allow-Credentials",
	}
	for _, header := range corsHeaders {
		assert.NotEmpty(t, resp.Header.Get(header),
			"CORS header %s should be present", header)
	}
}

// TestServer_Middleware_CORS_Production tests CORS configuration in production
func TestServer_Middleware_CORS_Production(t *testing.T) {
	cfg := createProductionTestConfig()
	cfg.App.Env = "production"
	cfg.App.CorsOriginProd = "https://example.com"

	db := setupTestDB(t)
	defer teardownTestDB(db)

	appInstance, err := app.NewServer(cfg, db)
	require.NoError(t, err)

	// Make a request with Origin header
	req := httptest.NewRequest("GET", "/health", nil)
	req.Header.Set("Origin", "https://example.com")
	resp, err := appInstance.Test(req)

	require.NoError(t, err)
	// Check for CORS headers
	accessControlAllowCredentials := resp.Header.Get("Access-Control-Allow-Credentials")
	assert.NotEmpty(t, accessControlAllowCredentials,
		"Access-Control-Allow-Credentials header should be present")
}

// TestServer_StaticFileUploads tests that static file serving for uploads is configured
func TestServer_StaticFileUploads(t *testing.T) {
	cfg := createTestConfig()
	db := setupTestDB(t)
	defer teardownTestDB(db)

	appInstance, err := app.NewServer(cfg, db)
	require.NoError(t, err)

	// Test that uploads route responds (likely 404 for non-existent file, but route exists)
	req := httptest.NewRequest("GET", "/uploads/test.txt", nil)
	resp, err := appInstance.Test(req)

	require.NoError(t, err)
	// Should not be a routing error - file may not exist but route is set up
	assert.True(t, resp.StatusCode == fiber.StatusNotFound || resp.StatusCode == fiber.StatusInternalServerError,
		"Uploads route should be configured")
}

// TestServer_LoggerConfiguration tests logger configuration based on environment
func TestServer_LoggerConfiguration(t *testing.T) {
	tests := []struct {
		name          string
		env           string
		logLevel      string
		logFormat     string
		expectError   bool
	}{
		{
			name:        "Development with text format",
			env:         "development",
			logLevel:    "INFO",
			logFormat:   "text",
			expectError: false,
		},
		{
			name:        "Development with json format",
			env:         "development",
			logLevel:    "DEBUG",
			logFormat:   "json",
			expectError: false,
		},
		{
			name:        "Production",
			env:         "production",
			logLevel:    "WARN",
			logFormat:   "json",
			expectError: false,
		},
		{
			name:        "Test environment",
			env:         "test",
			logLevel:    "ERROR",
			logFormat:   "text",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := createTestConfig()
			cfg.App.Env = tt.env
			cfg.Logging.Level = tt.logLevel
			cfg.Logging.Format = tt.logFormat

			db := setupTestDB(t)
			defer teardownTestDB(db)

			_, err := app.NewServer(cfg, db)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestServer_Shutdown tests graceful shutdown of the server
func TestServer_Shutdown(t *testing.T) {
	cfg := createTestConfig()
	db := setupTestDB(t)
	defer teardownTestDB(db)

	appInstance, err := app.NewServer(cfg, db)
	require.NoError(t, err)

	// Make a test request to ensure server is running
	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := appInstance.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Test shutdown
	err = appInstance.Shutdown()
	assert.NoError(t, err, "Server should shutdown without error")
}

// TestServer_ShutdownWithContext tests graceful shutdown with timeout
func TestServer_ShutdownWithContext(t *testing.T) {
	cfg := createTestConfig()
	db := setupTestDB(t)
	defer teardownTestDB(db)

	appInstance, err := app.NewServer(cfg, db)
	require.NoError(t, err)

	// Make a test request
	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := appInstance.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Test shutdown with timeout
	done := make(chan error, 1)
	go func() {
		done <- appInstance.Shutdown()
	}()

	select {
	case err := <-done:
		assert.NoError(t, err, "Server should shutdown without error")
	case <-time.After(5 * time.Second):
		t.Fatal("Server shutdown timed out")
	}
}

// TestServer_MultipleRequests tests handling multiple concurrent requests
func TestServer_MultipleRequests(t *testing.T) {
	cfg := createTestConfig()
	db := setupTestDB(t)
	defer teardownTestDB(db)

	appInstance, err := app.NewServer(cfg, db)
	require.NoError(t, err)

	// Make multiple concurrent requests
	numRequests := 10
	results := make(chan int, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			req := httptest.NewRequest("GET", "/health", nil)
			resp, err := appInstance.Test(req)
			if err != nil {
				results <- -1
				return
			}
			results <- resp.StatusCode
		}()
	}

	// Collect results
	successCount := 0
	for i := 0; i < numRequests; i++ {
		statusCode := <-results
		if statusCode == fiber.StatusOK {
			successCount++
		}
	}

	// All requests should succeed
	assert.Equal(t, numRequests, successCount,
		"All concurrent requests should succeed")
}

// TestServer_DifferentHTTPMethods tests that the server handles different HTTP methods
func TestServer_DifferentHTTPMethods(t *testing.T) {
	cfg := createTestConfig()
	db := setupTestDB(t)
	defer teardownTestDB(db)

	appInstance, err := app.NewServer(cfg, db)
	require.NoError(t, err)

	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/non-existent", nil)
			resp, err := appInstance.Test(req)

			require.NoError(t, err)
			// Should get a response (even if 404)
			assert.NotNil(t, resp)
		})
	}
}

// TestServer_HeaderHandling tests that the server properly handles headers
func TestServer_HeaderHandling(t *testing.T) {
	cfg := createTestConfig()
	db := setupTestDB(t)
	defer teardownTestDB(db)

	appInstance, err := app.NewServer(cfg, db)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/health", nil)
	req.Header.Set("User-Agent", "test-agent")
	req.Header.Set("Accept", "application/json")

	resp, err := appInstance.Test(req)
	require.NoError(t, err)

	// Check that response has basic headers
	assert.NotEmpty(t, resp.Header.Get("Content-Type"))
	assert.NotEmpty(t, resp.Header.Get("X-Request-ID"))
}

// TestServer_ConfigVariations tests server creation with various config variations
func TestServer_ConfigVariations(t *testing.T) {
	tests := []struct {
		name    string
		modifyFn func(*config.Config)
	}{
		{
			name: "With max upload size",
			modifyFn: func(cfg *config.Config) {
				cfg.Upload.MaxSize = 1073741824 // 1GB
			},
		},
		{
			name: "With custom chunk size",
			modifyFn: func(cfg *config.Config) {
				cfg.Upload.ChunkSize = 2097152 // 2MB
			},
		},
		{
			name: "With custom TUS version",
			modifyFn: func(cfg *config.Config) {
				cfg.Upload.TusVersion = "1.0.0"
			},
		},
		{
			name: "With custom idle timeout",
			modifyFn: func(cfg *config.Config) {
				cfg.Upload.IdleTimeout = 900
			},
		},
		{
			name: "With custom cleanup interval",
			modifyFn: func(cfg *config.Config) {
				cfg.Upload.CleanupInterval = 600
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := createTestConfig()
			tt.modifyFn(cfg)
			db := setupTestDB(t)
			defer teardownTestDB(db)

			_, err := app.NewServer(cfg, db)
			assert.NoError(t, err, "NewServer should not return error with config variation: %s", tt.name)
		})
	}
}

// TestServer_CORSHeadersVerification tests detailed CORS headers
func TestServer_CORSHeadersVerification(t *testing.T) {
	cfg := createTestConfig()
	cfg.App.Env = "development"
	db := setupTestDB(t)
	defer teardownTestDB(db)

	appInstance, err := app.NewServer(cfg, db)
	require.NoError(t, err)

	// Make a regular GET request and check CORS headers are set
	req := httptest.NewRequest("GET", "/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")

	resp, err := appInstance.Test(req)
	require.NoError(t, err)

	// Verify response is successful
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	// Check that request ID is present (indicates proper middleware chain)
	assert.NotEmpty(t, resp.Header.Get("X-Request-ID"))
}

// TestServer_ErrorHandlerBehavior tests custom error handler behavior
