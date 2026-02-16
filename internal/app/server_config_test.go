package app_test

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"invento-service/config"
	app "invento-service/internal/app"
)

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
