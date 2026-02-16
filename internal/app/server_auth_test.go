package app_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"invento-service/config"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	app "invento-service/internal/app"
)

func TestNewServer_CreatesFiberApp(t *testing.T) {
	t.Skip("requires network access to Supabase JWKS endpoint")
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
	t.Skip("requires network access to Supabase JWKS endpoint")
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
			Level:          "INFO",
			Format:         "text",
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
	t.Skip("requires network access to Supabase JWKS endpoint")
	cfg := createProductionTestConfig()
	db := setupTestDB(t)
	defer teardownTestDB(db)

	_, err := app.NewServer(cfg, db)
	require.NoError(t, err, "NewServer should work with production config")
}

// TestServer_ErrorHandler_NotFound tests the custom error handler for 404
func TestServer_ErrorHandler_NotFound(t *testing.T) {
	t.Skip("requires network access to Supabase JWKS endpoint")
	cfg := createTestConfig()
	db := setupTestDB(t)
	defer teardownTestDB(db)

	appInstance, err := app.NewServer(cfg, db)
	require.NoError(t, err)

	// Test a non-existent route
	req := httptest.NewRequest("GET", "/non-existent-route", http.NoBody)
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
	t.Skip("requires network access to Supabase JWKS endpoint")
	cfg := createTestConfig()
	db := setupTestDB(t)
	defer teardownTestDB(db)

	appInstance, err := app.NewServer(cfg, db)
	require.NoError(t, err)

	// Test an upload path that doesn't exist
	req := httptest.NewRequest("GET", "/uploads/non-existent-file.txt", http.NoBody)
	resp, err := appInstance.Test(req)

	require.NoError(t, err)
	// Uploads paths should return the raw error, not our custom error handler
	// This typically returns 404 or 500 depending on Fiber's behavior
	assert.True(t, resp.StatusCode == fiber.StatusNotFound || resp.StatusCode == fiber.StatusInternalServerError)
}

// TestServer_ErrorHandler_TusProtocol tests TUS protocol error handling
func TestServer_ErrorHandler_TusProtocol(t *testing.T) {
	t.Skip("requires network access to Supabase JWKS endpoint")
	cfg := createTestConfig()
	db := setupTestDB(t)
	defer teardownTestDB(db)

	appInstance, err := app.NewServer(cfg, db)
	require.NoError(t, err)

	// Test TUS protocol request - returns 401 because of JWT middleware, but route exists
	req := httptest.NewRequest("PATCH", "/api/v1/project/upload/123", http.NoBody)
	req.Header.Set("Tus-Resumable", "1.0.0")
	resp, err := appInstance.Test(req)

	require.NoError(t, err)
	// Should get a response (401 from auth middleware) showing the route is registered
	assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode, "TUS upload route should be registered")
}

// TestServer_RouteRegistration_AuthRoutes tests that auth routes are registered
func TestServer_RouteRegistration_AuthRoutes(t *testing.T) {
	t.Skip("requires network access to Supabase JWKS endpoint")
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
			req := httptest.NewRequest(route.method, route.path, http.NoBody)
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
	t.Skip("requires network access to Supabase JWKS endpoint")
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
			req := httptest.NewRequest(route.method, route.path, http.NoBody)
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
	t.Skip("requires network access to Supabase JWKS endpoint")
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
			req := httptest.NewRequest(route.method, route.path, http.NoBody)
			resp, err := appInstance.Test(req)

			require.NoError(t, err)
			assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode,
				"Route %s %s should be registered", route.method, route.path)
		})
	}
}

// TestServer_RouteRegistration_ProfileRoutes tests that profile routes are registered
func TestServer_RouteRegistration_ProfileRoutes(t *testing.T) {
	t.Skip("requires network access to Supabase JWKS endpoint")
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
			req := httptest.NewRequest(route.method, route.path, http.NoBody)
			resp, err := appInstance.Test(req)

			require.NoError(t, err)
			assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode,
				"Route %s %s should be registered", route.method, route.path)
		})
	}
}

// TestServer_RouteRegistration_ProjectRoutes tests that project routes are registered
func TestServer_RouteRegistration_ProjectRoutes(t *testing.T) {
	t.Skip("requires network access to Supabase JWKS endpoint")
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
			req := httptest.NewRequest(route.method, route.path, http.NoBody)
			resp, err := appInstance.Test(req)

			require.NoError(t, err)
			assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode,
				"Route %s %s should be registered", route.method, route.path)
		})
	}
}

// TestServer_RouteRegistration_ModulRoutes tests that modul routes are registered
func TestServer_RouteRegistration_ModulRoutes(t *testing.T) {
	t.Skip("requires network access to Supabase JWKS endpoint")
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
			req := httptest.NewRequest(route.method, route.path, http.NoBody)
			resp, err := appInstance.Test(req)

			require.NoError(t, err)
			assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode,
				"Route %s %s should be registered", route.method, route.path)
		})
	}
}

// TestServer_RouteRegistration_StatisticsRoutes tests that statistics routes are registered
func TestServer_RouteRegistration_StatisticsRoutes(t *testing.T) {
	t.Skip("requires network access to Supabase JWKS endpoint")
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
			req := httptest.NewRequest(route.method, route.path, http.NoBody)
			resp, err := appInstance.Test(req)

			require.NoError(t, err)
			assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode,
				"Route %s %s should be registered", route.method, route.path)
		})
	}
}

// TestServer_RouteRegistration_MonitoringRoutes tests that monitoring routes are registered
func TestServer_RouteRegistration_MonitoringRoutes(t *testing.T) {
	t.Skip("requires network access to Supabase JWKS endpoint")
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
			req := httptest.NewRequest(route.method, route.path, http.NoBody)
			resp, err := appInstance.Test(req)

			require.NoError(t, err)
			assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode,
				"Route %s %s should be registered", route.method, route.path)
		})
	}
}

// TestServer_RouteRegistration_SwaggerRoute tests that swagger route is registered
