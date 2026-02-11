package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHealthStatus_Constants(t *testing.T) {
	assert.Equal(t, HealthStatus("healthy"), HealthStatusHealthy)
	assert.Equal(t, HealthStatus("unhealthy"), HealthStatusUnhealthy)
	assert.Equal(t, HealthStatus("degraded"), HealthStatusDegraded)
}

func TestServiceStatus_Constants(t *testing.T) {
	assert.Equal(t, ServiceStatus("healthy"), ServiceStatusHealthy)
	assert.Equal(t, ServiceStatus("unhealthy"), ServiceStatusUnhealthy)
	assert.Equal(t, ServiceStatus("connected"), ServiceStatusConnected)
	assert.Equal(t, ServiceStatus("disconnected"), ServiceStatusDisconnected)
	assert.Equal(t, ServiceStatus("error"), ServiceStatusError)
	assert.Equal(t, ServiceStatus("loaded"), ServiceStatusLoaded)
	assert.Equal(t, ServiceStatus("running"), ServiceStatusRunning)
}

func TestBasicHealthCheck_Structure(t *testing.T) {
	timestamp := time.Now()
	health := BasicHealthCheck{
		Status:    HealthStatusHealthy,
		App:       "test-app",
		Timestamp: timestamp,
	}

	assert.Equal(t, HealthStatusHealthy, health.Status)
	assert.Equal(t, "test-app", health.App)
	assert.Equal(t, timestamp, health.Timestamp)
}

func TestAppInfo_Structure(t *testing.T) {
	startTime := time.Now()
	appInfo := AppInfo{
		Name:        "fiber-boiler-plate",
		Version:     "1.0.0",
		Environment: "test",
		Uptime:      "1h 30m 45s",
		StartTime:   startTime,
		Status:      "running",
	}

	assert.Equal(t, "fiber-boiler-plate", appInfo.Name)
	assert.Equal(t, "1.0.0", appInfo.Version)
	assert.Equal(t, "test", appInfo.Environment)
	assert.Equal(t, "1h 30m 45s", appInfo.Uptime)
	assert.Equal(t, startTime, appInfo.StartTime)
	assert.Equal(t, "running", appInfo.Status)
}

func TestDatabaseStatus_Structure(t *testing.T) {
	dbStatus := DatabaseStatus{
		Status:          ServiceStatusConnected,
		PingTime:        "2ms",
		OpenConnections: 5,
		IdleConnections: 3,
		MaxConnections:  100,
		TotalQueries:    1250,
		Name:            "PostgreSQL",
		Version:         "15.3",
	}

	assert.Equal(t, ServiceStatusConnected, dbStatus.Status)
	assert.Equal(t, "2ms", dbStatus.PingTime)
	assert.Equal(t, 5, dbStatus.OpenConnections)
	assert.Equal(t, 3, dbStatus.IdleConnections)
	assert.Equal(t, 100, dbStatus.MaxConnections)
	assert.Equal(t, int64(1250), dbStatus.TotalQueries)
	assert.Equal(t, "PostgreSQL", dbStatus.Name)
	assert.Equal(t, "15.3", dbStatus.Version)
}

func TestSystemInfo_Structure(t *testing.T) {
	systemInfo := SystemInfo{
		MemoryUsage: "45.2MB",
		CPUCores:    4,
		Goroutines:  12,
	}

	assert.Equal(t, "45.2MB", systemInfo.MemoryUsage)
	assert.Equal(t, 4, systemInfo.CPUCores)
	assert.Equal(t, 12, systemInfo.Goroutines)
}

func TestDetailedSystemInfo_Structure(t *testing.T) {
	systemInfo := DetailedSystemInfo{
		Memory: MemoryInfo{
			Allocated:      "45.2MB",
			TotalAllocated: "120.5MB",
			System:         "256MB",
			GCCount:        15,
		},
		CPU: CPUInfo{
			Cores:      4,
			Goroutines: 12,
		},
		Runtime: RuntimeInfo{
			GoVersion: "go1.21",
			Compiler:  "gc",
			Arch:      "amd64",
			OS:        "linux",
		},
	}

	assert.Equal(t, "45.2MB", systemInfo.Memory.Allocated)
	assert.Equal(t, "120.5MB", systemInfo.Memory.TotalAllocated)
	assert.Equal(t, "256MB", systemInfo.Memory.System)
	assert.Equal(t, uint32(15), systemInfo.Memory.GCCount)
	assert.Equal(t, 4, systemInfo.CPU.Cores)
	assert.Equal(t, 12, systemInfo.CPU.Goroutines)
	assert.Equal(t, "go1.21", systemInfo.Runtime.GoVersion)
	assert.Equal(t, "gc", systemInfo.Runtime.Compiler)
	assert.Equal(t, "amd64", systemInfo.Runtime.Arch)
	assert.Equal(t, "linux", systemInfo.Runtime.OS)
}

func TestHttpMetrics_Structure(t *testing.T) {
	httpMetrics := HttpMetrics{
		TotalRequests:  5420,
		ActiveRequests: 3,
		ResponseTimes: ResponseTimes{
			Min: "5ms",
			Max: "150ms",
			Avg: "25ms",
		},
	}

	assert.Equal(t, int64(5420), httpMetrics.TotalRequests)
	assert.Equal(t, 3, httpMetrics.ActiveRequests)
	assert.Equal(t, "5ms", httpMetrics.ResponseTimes.Min)
	assert.Equal(t, "150ms", httpMetrics.ResponseTimes.Max)
	assert.Equal(t, "25ms", httpMetrics.ResponseTimes.Avg)
}

func TestDependency_Structure(t *testing.T) {
	dependency := Dependency{
		Name:    "fiber",
		Version: "v2.50.0",
		Status:  ServiceStatusLoaded,
	}

	assert.Equal(t, "fiber", dependency.Name)
	assert.Equal(t, "v2.50.0", dependency.Version)
	assert.Equal(t, ServiceStatusLoaded, dependency.Status)
}

func TestComprehensiveHealthCheck_Structure(t *testing.T) {
	timestamp := time.Now()
	healthCheck := ComprehensiveHealthCheck{
		Status: HealthStatusHealthy,
		App: AppInfo{
			Name:        "test-app",
			Version:     "1.0.0",
			Environment: "test",
			Uptime:      "1h",
		},
		Database: DatabaseStatus{
			Status:          ServiceStatusConnected,
			PingTime:        "2ms",
			OpenConnections: 5,
		},
		System: SystemInfo{
			MemoryUsage: "45MB",
			CPUCores:    4,
			Goroutines:  10,
		},
		Timestamp: timestamp,
	}

	assert.Equal(t, HealthStatusHealthy, healthCheck.Status)
	assert.Equal(t, "test-app", healthCheck.App.Name)
	assert.Equal(t, ServiceStatusConnected, healthCheck.Database.Status)
	assert.Equal(t, 4, healthCheck.System.CPUCores)
	assert.Equal(t, timestamp, healthCheck.Timestamp)
}
