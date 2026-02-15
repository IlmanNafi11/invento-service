package http_test

import (
	"encoding/json"
	httpcontroller "invento-service/internal/controller/http"
	"invento-service/internal/domain"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockHealthUsecase struct {
	mock.Mock
}

func (m *MockHealthUsecase) GetBasicHealth() *domain.BasicHealthCheck {
	args := m.Called()
	return args.Get(0).(*domain.BasicHealthCheck)
}

func (m *MockHealthUsecase) GetComprehensiveHealth() *domain.ComprehensiveHealthCheck {
	args := m.Called()
	return args.Get(0).(*domain.ComprehensiveHealthCheck)
}

func (m *MockHealthUsecase) GetSystemMetrics() *domain.SystemMetrics {
	args := m.Called()
	return args.Get(0).(*domain.SystemMetrics)
}

func (m *MockHealthUsecase) GetApplicationStatus() *domain.ApplicationStatus {
	args := m.Called()
	return args.Get(0).(*domain.ApplicationStatus)
}

func TestHealthController_BasicHealthCheck_Success(t *testing.T) {
	mockUsecase := new(MockHealthUsecase)
	controller := httpcontroller.NewHealthController(mockUsecase)

	expectedData := &domain.BasicHealthCheck{
		Status:    domain.HealthStatusHealthy,
		App:       "test-app",
		Timestamp: time.Now(),
	}

	mockUsecase.On("GetBasicHealth").Return(expectedData)

	app := fiber.New()
	app.Get("/health", controller.BasicHealthCheck)

	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "Server berjalan dengan baik", response["message"])
	assert.Equal(t, float64(200), response["code"])

	mockUsecase.AssertExpectations(t)
}

func TestHealthController_ComprehensiveHealthCheck_Success(t *testing.T) {
	mockUsecase := new(MockHealthUsecase)
	controller := httpcontroller.NewHealthController(mockUsecase)

	expectedData := &domain.ComprehensiveHealthCheck{
		Status: domain.HealthStatusHealthy,
		App: domain.AppInfo{
			Name:        "test-app",
			Version:     "1.0.0",
			Environment: "test",
			Uptime:      "1h",
		},
		Database: domain.DatabaseStatus{
			Status:   domain.ServiceStatusConnected,
			PingTime: "2ms",
		},
		System: domain.SystemInfo{
			MemoryUsage: "45MB",
			CPUCores:    4,
			Goroutines:  10,
		},
		Timestamp: time.Now(),
	}

	mockUsecase.On("GetComprehensiveHealth").Return(expectedData)

	app := fiber.New()
	app.Get("/health", controller.ComprehensiveHealthCheck)

	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "Pemeriksaan kesehatan sistem berhasil", response["message"])
	assert.Equal(t, float64(200), response["code"])

	mockUsecase.AssertExpectations(t)
}

func TestHealthController_ComprehensiveHealthCheck_Unhealthy(t *testing.T) {
	mockUsecase := new(MockHealthUsecase)
	controller := httpcontroller.NewHealthController(mockUsecase)

	expectedData := &domain.ComprehensiveHealthCheck{
		Status: domain.HealthStatusUnhealthy,
		App: domain.AppInfo{
			Name:        "test-app",
			Version:     "1.0.0",
			Environment: "test",
			Uptime:      "1h",
		},
		Database: domain.DatabaseStatus{
			Status: domain.ServiceStatusDisconnected,
			Error:  "Koneksi database terputus",
		},
		System: domain.SystemInfo{
			MemoryUsage: "45MB",
			CPUCores:    4,
			Goroutines:  10,
		},
		Timestamp: time.Now(),
	}

	mockUsecase.On("GetComprehensiveHealth").Return(expectedData)

	app := fiber.New()
	app.Get("/health", controller.ComprehensiveHealthCheck)

	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "error", response["status"])
	assert.Equal(t, "Beberapa komponen sistem mengalami masalah", response["message"])
	assert.Equal(t, float64(503), response["code"])

	mockUsecase.AssertExpectations(t)
}

func TestHealthController_GetSystemMetrics_Success(t *testing.T) {
	mockUsecase := new(MockHealthUsecase)
	controller := httpcontroller.NewHealthController(mockUsecase)

	expectedData := &domain.SystemMetrics{
		App: domain.AppInfo{
			Name:        "test-app",
			Version:     "1.0.0",
			Environment: "test",
			Uptime:      "1h",
		},
		System: domain.DetailedSystemInfo{
			Memory: domain.MemoryInfo{
				Allocated:      "45MB",
				TotalAllocated: "120MB",
				System:         "256MB",
				GCCount:        15,
			},
			CPU: domain.CPUInfo{
				Cores:      4,
				Goroutines: 10,
			},
			Runtime: domain.RuntimeInfo{
				GoVersion: "go1.21",
				Compiler:  "gc",
				Arch:      "amd64",
				OS:        "linux",
			},
		},
		Database: domain.DatabaseStatus{
			Status:   domain.ServiceStatusConnected,
			PingTime: "2ms",
		},
		Http: domain.HttpMetrics{
			TotalRequests:  5420,
			ActiveRequests: 3,
			ResponseTimes: domain.ResponseTimes{
				Min: "5ms",
				Max: "150ms",
				Avg: "25ms",
			},
		},
	}

	mockUsecase.On("GetSystemMetrics").Return(expectedData)

	app := fiber.New()
	app.Get("/metrics", controller.GetSystemMetrics)

	req := httptest.NewRequest("GET", "/metrics", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "Metrics sistem berhasil diambil", response["message"])
	assert.Equal(t, float64(200), response["code"])

	mockUsecase.AssertExpectations(t)
}

func TestHealthController_GetApplicationStatus_Success(t *testing.T) {
	mockUsecase := new(MockHealthUsecase)
	controller := httpcontroller.NewHealthController(mockUsecase)

	expectedData := &domain.ApplicationStatus{
		App: domain.AppInfo{
			Name:        "test-app",
			Version:     "1.0.0",
			Environment: "test",
			Status:      "running",
			Uptime:      "1h",
		},
		Services: domain.ServicesStatus{
			Database: domain.DatabaseService{
				Name:     "PostgreSQL",
				Status:   domain.ServiceStatusHealthy,
				Version:  "15.3",
				PingTime: "2ms",
			},
		},
		Dependencies: []domain.Dependency{
			{
				Name:    "fiber",
				Version: "v2.50.0",
				Status:  domain.ServiceStatusLoaded,
			},
			{
				Name:    "gorm",
				Version: "v1.25.4",
				Status:  domain.ServiceStatusLoaded,
			},
		},
	}

	mockUsecase.On("GetApplicationStatus").Return(expectedData)

	app := fiber.New()
	app.Get("/status", controller.GetApplicationStatus)

	req := httptest.NewRequest("GET", "/status", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "Status aplikasi berhasil diambil", response["message"])
	assert.Equal(t, float64(200), response["code"])

	mockUsecase.AssertExpectations(t)
}
