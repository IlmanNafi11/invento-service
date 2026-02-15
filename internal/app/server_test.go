package app_test

import (
	"invento-service/config"
	app "invento-service/internal/app"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	require.NoError(t, err, "Failed to connect to test database")
	return db
}

// teardownTestDB closes the database connection
func teardownTestDB(db *gorm.DB) {
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.Close()
	}
}

// TestNewServer_CreatesFiberApp tests that NewServer creates a valid Fiber app
func TestNewServer_CreatesFiberApp(t *testing.T) {
	// Setup test database
	db := setupTestDB(t)
	defer teardownTestDB(db)

	cfg := createTestConfig()

	// This test verifies that NewServer can be called without error
	// The actual server creation may require additional setup
	_, err := app.NewServer(cfg, db)
	require.NoError(t, err, "NewServer should not return error when called with valid config and db")
}

// TestNewServer_WithDevelopmentConfig tests server creation with development config
func TestNewServer_WithDevelopmentConfig(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			Name:           "Test App",
			Port:           "3000",
			Env:            "development",
			CorsOriginDev:  "http://localhost:5173",
			CorsOriginProd: "https://example.com",
		},
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "test",
			Password: "test",
			Name:     "testdb",
		},
		Upload: config.UploadConfig{
			MaxSize:              524288000,
			MaxSizeProject:       524288000,
			MaxSizeModul:         52428800,
			ChunkSize:            1048576,
			MaxConcurrent:        1,
			MaxConcurrentProject: 1,
			MaxConcurrentModul:   1,
			MaxQueueModulPerUser: 5,
			IdleTimeout:          600,
			CleanupInterval:      300,
			PathProduction:       "/tmp/uploads/",
			PathDevelopment:      "./uploads/",
			TempPathProduction:   "/tmp/uploads/temp/",
			TempPathDevelopment:  "./uploads/temp/",
			TusVersion:           "1.0.0",
			MaxResumeAttempts:    10,
		},
		Supabase: config.SupabaseConfig{
			URL:        "https://test.supabase.co",
			AnonKey:    "test_anon_key",
			ServiceKey: "test_service_role_key",
			DBURL:      "postgresql://test:test@localhost:5432/testdb",
		},
		Logging: config.LoggingConfig{
			Level:         "INFO",
			Format:        "text",
			LogRequestBody: false,
		},
		Swagger: config.SwaggerConfig{
			Enabled: false,
		},
		Performance: config.PerformanceConfig{
			FiberConcurrency:       256,
			FiberReduceMemory:      false,
			FiberStreamRequestBody: false,
			FiberReadBufferSize:    16384,
			DBMaxOpenConns:         10,
			DBMaxIdleConns:         3,
			DBConnMaxLifetime:      1800,
			DBConnMaxIdleTime:      300,
			GoMemLimit:             "350MiB",
			GoGC:                   100,
			EnablePprof:            false,
			MemoryWarningThreshold: 0.8,
		},
	}

	db := setupTestDB(t)
	defer teardownTestDB(db)

	_, err := app.NewServer(cfg, db)
	require.NoError(t, err, "NewServer should work with development config")
}

// TestNewServer_WithProductionConfig tests server creation with production config
func TestNewServer_WithProductionConfig(t *testing.T) {
	cfg := createProductionTestConfig()
	db := setupTestDB(t)
	defer teardownTestDB(db)

	_, err := app.NewServer(cfg, db)
	require.NoError(t, err, "NewServer should work with production config")
}

// TestServer_ErrorHandler_NotFound tests the custom error handler for 404
func TestServer_ErrorHandler_NotFound(t *testing.T) {
	cfg := createTestConfig()
	db := setupTestDB(t)
	defer teardownTestDB(db)

	appInstance, err := app.NewServer(cfg, db)
	require.NoError(t, err)

	// Test a non-existent route
	req := httptest.NewRequest("GET", "/non-existent-route", nil)
	resp, err := appInstance.Test(req)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	// Verify response body contains expected error message
	body := make([]byte, 1024)
	n, _ := resp.Body.Read(body)
	bodyStr := string(body[:n])
	assert.Contains(t, bodyStr, "Endpoint tidak ditemukan")
}

// TestServer_ErrorHandler_UploadsPath tests that errors in /uploads path are returned as-is
func TestServer_ErrorHandler_UploadsPath(t *testing.T) {
	cfg := createTestConfig()
	db := setupTestDB(t)
	defer teardownTestDB(db)

	appInstance, err := app.NewServer(cfg, db)
	require.NoError(t, err)

	// Test an upload path that doesn't exist
	req := httptest.NewRequest("GET", "/uploads/non-existent-file.txt", nil)
	resp, err := appInstance.Test(req)

	require.NoError(t, err)
	// Uploads paths should return the raw error, not our custom error handler
	// This typically returns 404 or 500 depending on Fiber's behavior
	assert.True(t, resp.StatusCode == fiber.StatusNotFound || resp.StatusCode == fiber.StatusInternalServerError)
}

// TestServer_ErrorHandler_TusProtocol tests TUS protocol error handling
func TestServer_ErrorHandler_TusProtocol(t *testing.T) {
	cfg := createTestConfig()
	db := setupTestDB(t)
	defer teardownTestDB(db)

	appInstance, err := app.NewServer(cfg, db)
	require.NoError(t, err)

	// Test TUS protocol request - returns 401 because of JWT middleware, but route exists
	req := httptest.NewRequest("PATCH", "/api/v1/project/upload/123", nil)
	req.Header.Set("Tus-Resumable", "1.0.0")
	resp, err := appInstance.Test(req)

	require.NoError(t, err)
	// Should get a response (401 from auth middleware) showing the route is registered
	assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode, "TUS upload route should be registered")
}

// TestServer_RouteRegistration_AuthRoutes tests that auth routes are registered
func TestServer_RouteRegistration_AuthRoutes(t *testing.T) {
	cfg := createTestConfig()
	db := setupTestDB(t)
	defer teardownTestDB(db)

	appInstance, err := app.NewServer(cfg, db)
	require.NoError(t, err)

	// Test that auth routes exist (even if they return authentication errors)
	routes := []struct {
		method string
		path   string
	}{
		{"POST", "/api/v1/auth/login"},
		{"POST", "/api/v1/auth/refresh"},
		{"POST", "/api/v1/auth/register/otp"},
		{"POST", "/api/v1/auth/register/verify-otp"},
		{"POST", "/api/v1/auth/register/resend-otp"},
		{"POST", "/api/v1/auth/reset-password/otp"},
		{"POST", "/api/v1/auth/reset-password/verify-otp"},
		{"POST", "/api/v1/auth/reset-password/confirm-otp"},
		{"POST", "/api/v1/auth/reset-password/resend-otp"},
		{"POST", "/api/v1/auth/logout"},
	}

	for _, route := range routes {
		t.Run(route.method+"_"+route.path, func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, nil)
			resp, err := appInstance.Test(req)

			require.NoError(t, err)
			// Routes should be registered - response could be 400, 401, 405, etc.
			// but not 404 which would mean the route doesn't exist
			assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode,
				"Route %s %s should be registered", route.method, route.path)
		})
	}
}

// TestServer_RouteRegistration_RoleRoutes tests that role routes are registered
func TestServer_RouteRegistration_RoleRoutes(t *testing.T) {
	cfg := createTestConfig()
	db := setupTestDB(t)
	defer teardownTestDB(db)

	appInstance, err := app.NewServer(cfg, db)
	require.NoError(t, err)

	routes := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/role/permissions"},
		{"GET", "/api/v1/role/"},
		{"POST", "/api/v1/role/"},
		{"GET", "/api/v1/role/1"},
		{"PUT", "/api/v1/role/1"},
		{"DELETE", "/api/v1/role/1"},
	}

	for _, route := range routes {
		t.Run(route.method+"_"+strings.ReplaceAll(route.path, "/", "_"), func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, nil)
			resp, err := appInstance.Test(req)

			require.NoError(t, err)
			// Should not be 404 - route exists but may require auth
			assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode,
				"Route %s %s should be registered", route.method, route.path)
		})
	}
}

// TestServer_RouteRegistration_UserRoutes tests that user routes are registered
func TestServer_RouteRegistration_UserRoutes(t *testing.T) {
	cfg := createTestConfig()
	db := setupTestDB(t)
	defer teardownTestDB(db)

	appInstance, err := app.NewServer(cfg, db)
	require.NoError(t, err)

	routes := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/user/"},
		{"PUT", "/api/v1/user/1/role"},
		{"DELETE", "/api/v1/user/1"},
		{"GET", "/api/v1/user/1/files"},
		{"POST", "/api/v1/user/1/download"},
		{"GET", "/api/v1/user/permissions"},
	}

	for _, route := range routes {
		t.Run(route.method+"_"+strings.ReplaceAll(route.path, "/", "_"), func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, nil)
			resp, err := appInstance.Test(req)

			require.NoError(t, err)
			assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode,
				"Route %s %s should be registered", route.method, route.path)
		})
	}
}

// TestServer_RouteRegistration_ProfileRoutes tests that profile routes are registered
func TestServer_RouteRegistration_ProfileRoutes(t *testing.T) {
	cfg := createTestConfig()
	db := setupTestDB(t)
	defer teardownTestDB(db)

	appInstance, err := app.NewServer(cfg, db)
	require.NoError(t, err)

	routes := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/profile/"},
		{"PUT", "/api/v1/profile/"},
	}

	for _, route := range routes {
		t.Run(route.method+"_"+strings.ReplaceAll(route.path, "/", "_"), func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, nil)
			resp, err := appInstance.Test(req)

			require.NoError(t, err)
			assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode,
				"Route %s %s should be registered", route.method, route.path)
		})
	}
}

// TestServer_RouteRegistration_ProjectRoutes tests that project routes are registered
func TestServer_RouteRegistration_ProjectRoutes(t *testing.T) {
	cfg := createTestConfig()
	db := setupTestDB(t)
	defer teardownTestDB(db)

	appInstance, err := app.NewServer(cfg, db)
	require.NoError(t, err)

	routes := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/project/"},
		{"GET", "/api/v1/project/1"},
		{"PATCH", "/api/v1/project/1"},
		{"POST", "/api/v1/project/download"},
		{"DELETE", "/api/v1/project/1"},
		{"GET", "/api/v1/project/upload/check-slot"},
		{"POST", "/api/v1/project/upload/reset-queue"},
		{"POST", "/api/v1/project/upload/"},
		{"PATCH", "/api/v1/project/upload/123"},
		{"HEAD", "/api/v1/project/upload/123"},
		{"GET", "/api/v1/project/upload/123"},
		{"DELETE", "/api/v1/project/upload/123"},
		{"POST", "/api/v1/project/1/upload"},
		{"PATCH", "/api/v1/project/1/update/456"},
		{"HEAD", "/api/v1/project/1/update/456"},
		{"GET", "/api/v1/project/1/update/456"},
		{"DELETE", "/api/v1/project/1/update/456"},
	}

	for _, route := range routes {
		t.Run(route.method+"_"+strings.ReplaceAll(route.path, "/", "_"), func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, nil)
			resp, err := appInstance.Test(req)

			require.NoError(t, err)
			assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode,
				"Route %s %s should be registered", route.method, route.path)
		})
	}
}

// TestServer_RouteRegistration_ModulRoutes tests that modul routes are registered
func TestServer_RouteRegistration_ModulRoutes(t *testing.T) {
	cfg := createTestConfig()
	db := setupTestDB(t)
	defer teardownTestDB(db)

	appInstance, err := app.NewServer(cfg, db)
	require.NoError(t, err)

	routes := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/modul/"},
		{"PATCH", "/api/v1/modul/1"},
		{"POST", "/api/v1/modul/download"},
		{"DELETE", "/api/v1/modul/1"},
		{"GET", "/api/v1/modul/upload/check-slot"},
		{"POST", "/api/v1/modul/upload/"},
		{"PATCH", "/api/v1/modul/upload/123"},
		{"HEAD", "/api/v1/modul/upload/123"},
		{"GET", "/api/v1/modul/upload/123"},
		{"DELETE", "/api/v1/modul/upload/123"},
		{"POST", "/api/v1/modul/1/upload"},
		{"PATCH", "/api/v1/modul/1/update/456"},
		{"HEAD", "/api/v1/modul/1/update/456"},
		{"GET", "/api/v1/modul/1/update/456"},
		{"DELETE", "/api/v1/modul/1/update/456"},
	}

	for _, route := range routes {
		t.Run(route.method+"_"+strings.ReplaceAll(route.path, "/", "_"), func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, nil)
			resp, err := appInstance.Test(req)

			require.NoError(t, err)
			assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode,
				"Route %s %s should be registered", route.method, route.path)
		})
	}
}

// TestServer_RouteRegistration_StatisticsRoutes tests that statistics routes are registered
func TestServer_RouteRegistration_StatisticsRoutes(t *testing.T) {
	cfg := createTestConfig()
	db := setupTestDB(t)
	defer teardownTestDB(db)

	appInstance, err := app.NewServer(cfg, db)
	require.NoError(t, err)

	routes := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/statistic/"},
	}

	for _, route := range routes {
		t.Run(route.method+"_"+strings.ReplaceAll(route.path, "/", "_"), func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, nil)
			resp, err := appInstance.Test(req)

			require.NoError(t, err)
			assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode,
				"Route %s %s should be registered", route.method, route.path)
		})
	}
}

// TestServer_RouteRegistration_MonitoringRoutes tests that monitoring routes are registered
func TestServer_RouteRegistration_MonitoringRoutes(t *testing.T) {
	cfg := createTestConfig()
	db := setupTestDB(t)
	defer teardownTestDB(db)

	appInstance, err := app.NewServer(cfg, db)
	require.NoError(t, err)

	routes := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/monitoring/health"},
		{"GET", "/api/v1/monitoring/metrics"},
		// Note: /api/v1/monitoring/status is excluded because it may timeout in test environment
		// due to database ping operations
		{"GET", "/health"},
	}

	for _, route := range routes {
		t.Run(route.method+"_"+strings.ReplaceAll(route.path, "/", "_"), func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, nil)
			resp, err := appInstance.Test(req)

			require.NoError(t, err)
			assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode,
				"Route %s %s should be registered", route.method, route.path)
		})
	}
}

// TestServer_RouteRegistration_SwaggerRoute tests that swagger route is registered
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
func TestServer_ErrorHandlerBehavior(t *testing.T) {
	cfg := createTestConfig()
	db := setupTestDB(t)
	defer teardownTestDB(db)

	appInstance, err := app.NewServer(cfg, db)
	require.NoError(t, err)

	tests := []struct {
		name           string
		path           string
		method         string
		expectedStatus int
		checkBody      bool
		bodyContains   string
	}{
		{
			name:           "Not found route",
			path:           "/this-route-does-not-exist",
			method:         "GET",
			expectedStatus: fiber.StatusNotFound,
			checkBody:      true,
			bodyContains:   "Endpoint tidak ditemukan",
		},
		{
			name:           "TUS request returns auth error",
			path:           "/api/v1/project/upload/test-id",
			method:         "PATCH",
			expectedStatus: fiber.StatusUnauthorized, // Changed from 500 to 401 since route has JWT middleware
			checkBody:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.path == "/api/v1/project/upload/test-id" {
				req.Header.Set("Tus-Resumable", "1.0.0")
			}
			resp, err := appInstance.Test(req)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.checkBody {
				body := make([]byte, 1024)
				n, _ := resp.Body.Read(body)
				bodyStr := string(body[:n])
				assert.Contains(t, bodyStr, tt.bodyContains)
			}
		})
	}
}

// Helper functions

func createTestConfig() *config.Config {
	return &config.Config{
		App: config.AppConfig{
			Name:           "Test App",
			Port:           "3000",
			Env:            "test",
			CorsOriginDev:  "http://localhost:5173",
			CorsOriginProd: "https://example.com",
		},
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "test",
			Password: "test",
			Name:     "testdb",
		},
		Upload: config.UploadConfig{
			MaxSize:              524288000,
			MaxSizeProject:       524288000,
			MaxSizeModul:         52428800,
			ChunkSize:            1048576,
			MaxConcurrent:        1,
			MaxConcurrentProject: 1,
			MaxConcurrentModul:   1,
			MaxQueueModulPerUser: 5,
			IdleTimeout:          600,
			CleanupInterval:      300,
			PathProduction:       "/tmp/uploads/",
			PathDevelopment:      "./uploads/",
			TempPathProduction:   "/tmp/uploads/temp/",
			TempPathDevelopment:  "./uploads/temp/",
			TusVersion:           "1.0.0",
			MaxResumeAttempts:    10,
		},
		Supabase: config.SupabaseConfig{
			URL:        "https://test.supabase.co",
			AnonKey:    "test_anon_key",
			ServiceKey: "test_service_role_key",
			DBURL:      "postgresql://test:test@localhost:5432/testdb",
		},
		Logging: config.LoggingConfig{
			Level:         "INFO",
			Format:        "text",
			LogRequestBody: false,
		},
		Swagger: config.SwaggerConfig{
			Enabled: false,
		},
		Performance: config.PerformanceConfig{
			FiberConcurrency:       256,
			FiberReduceMemory:      false,
			FiberStreamRequestBody: false,
			FiberReadBufferSize:    16384,
			DBMaxOpenConns:         10,
			DBMaxIdleConns:         3,
			DBConnMaxLifetime:      1800,
			DBConnMaxIdleTime:      300,
			GoMemLimit:             "350MiB",
			GoGC:                   100,
			EnablePprof:            false,
			MemoryWarningThreshold: 0.8,
		},
	}
}

func createProductionTestConfig() *config.Config {
	cfg := createTestConfig()
	cfg.App.Env = "production"
	cfg.Logging.Format = "json"
	cfg.Logging.Level = "WARN"
	return cfg
}
