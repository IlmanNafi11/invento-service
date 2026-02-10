package domain

import "time"

type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusDegraded  HealthStatus = "degraded"
)

type ServiceStatus string

const (
	ServiceStatusHealthy      ServiceStatus = "healthy"
	ServiceStatusUnhealthy    ServiceStatus = "unhealthy"
	ServiceStatusConnected    ServiceStatus = "connected"
	ServiceStatusDisconnected ServiceStatus = "disconnected"
	ServiceStatusError        ServiceStatus = "error"
	ServiceStatusLoaded       ServiceStatus = "loaded"
	ServiceStatusRunning      ServiceStatus = "running"
)

type BasicHealthCheck struct {
	Status    HealthStatus `json:"status"`
	App       string       `json:"app"`
	Timestamp time.Time    `json:"timestamp"`
}

type AppInfo struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Environment string    `json:"environment"`
	Uptime      string    `json:"uptime"`
	StartTime   time.Time `json:"start_time,omitempty"`
	Status      string    `json:"status,omitempty"`
}

type DatabaseStatus struct {
	Status          ServiceStatus `json:"status"`
	PingTime        string        `json:"ping_time,omitempty"`
	OpenConnections int           `json:"open_connections,omitempty"`
	IdleConnections int           `json:"idle_connections,omitempty"`
	MaxConnections  int           `json:"max_connections,omitempty"`
	TotalQueries    int64         `json:"total_queries,omitempty"`
	Error           string        `json:"error,omitempty"`
	Name            string        `json:"name,omitempty"`
	Version         string        `json:"version,omitempty"`
}

type SystemInfo struct {
	MemoryUsage string `json:"memory_usage"`
	CPUCores    int    `json:"cpu_cores"`
	Goroutines  int    `json:"goroutines"`
}

type DetailedSystemInfo struct {
	Memory  MemoryInfo  `json:"memory"`
	CPU     CPUInfo     `json:"cpu"`
	Runtime RuntimeInfo `json:"runtime"`
}

type MemoryInfo struct {
	Allocated      string `json:"allocated"`
	TotalAllocated string `json:"total_allocated"`
	System         string `json:"system"`
	GCCount        uint32 `json:"gc_count"`
}

type CPUInfo struct {
	Cores      int `json:"cores"`
	Goroutines int `json:"goroutines"`
}

type RuntimeInfo struct {
	GoVersion string `json:"go_version"`
	Compiler  string `json:"compiler"`
	Arch      string `json:"arch"`
	OS        string `json:"os"`
}

type HttpMetrics struct {
	TotalRequests  int64         `json:"total_requests"`
	ActiveRequests int           `json:"active_requests"`
	ResponseTimes  ResponseTimes `json:"response_times"`
}

type ResponseTimes struct {
	Min string `json:"min"`
	Max string `json:"max"`
	Avg string `json:"avg"`
}

type ServicesStatus struct {
	Database DatabaseService `json:"database"`
	Email    EmailService    `json:"email"`
}

type DatabaseService struct {
	Name     string        `json:"name"`
	Status   ServiceStatus `json:"status"`
	Version  string        `json:"version"`
	PingTime string        `json:"ping_time"`
}

type EmailService struct {
	Name      string        `json:"name"`
	Provider  string        `json:"provider"`
	Status    ServiceStatus `json:"status"`
	APIKeySet bool          `json:"api_key_set"`
}

type Dependency struct {
	Name    string        `json:"name"`
	Version string        `json:"version"`
	Status  ServiceStatus `json:"status"`
}

type ComprehensiveHealthCheck struct {
	Status    HealthStatus   `json:"status"`
	App       AppInfo        `json:"app"`
	Database  DatabaseStatus `json:"database"`
	System    SystemInfo     `json:"system"`
	Timestamp time.Time      `json:"timestamp"`
}

type SystemMetrics struct {
	App      AppInfo            `json:"app"`
	System   DetailedSystemInfo `json:"system"`
	Database DatabaseStatus     `json:"database"`
	Http     HttpMetrics        `json:"http"`
}

type ApplicationStatus struct {
	App          AppInfo        `json:"app"`
	Services     ServicesStatus `json:"services"`
	Dependencies []Dependency   `json:"dependencies"`
}
