package testing

import (
	"time"

	"invento-service/internal/domain"
	"invento-service/internal/dto"
)

// User fixtures

// GetTestUser returns a test user entity
func GetTestUser() domain.User {
	roleID := 1
	return domain.User{
		ID:        "00000000-0000-0000-0000-000000000001",
		Name:      "Test User",
		Email:     "test@example.com",
		RoleID:    &roleID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// GetTestAdminUser returns a test admin user entity
func GetTestAdminUser() domain.User {
	roleID := 2
	return domain.User{
		ID:        "00000000-0000-0000-0000-000000000002",
		Name:      "Admin User",
		Email:     "admin@example.com",
		RoleID:    &roleID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// GetTestUsers returns a slice of test users
func GetTestUsers() []domain.User {
	roleID1 := 1
	return []domain.User{
		GetTestUser(),
		GetTestAdminUser(),
		{
			ID:        "00000000-0000-0000-0000-000000000003",
			Name:      "Regular User",
			Email:     "user@example.com",
			RoleID:    &roleID1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
}

// Role fixtures

// GetTestRole returns a test role entity
func GetTestRole() domain.Role {
	return domain.Role{
		ID:        1,
		NamaRole:  "user",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// GetTestAdminRole returns a test admin role entity
func GetTestAdminRole() domain.Role {
	return domain.Role{
		ID:        2,
		NamaRole:  "admin",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// GetTestRoles returns a slice of test roles
func GetTestRoles() []domain.Role {
	return []domain.Role{
		GetTestRole(),
		GetTestAdminRole(),
	}
}

// Project fixtures

// GetTestProject returns a test project entity
func GetTestProject() domain.Project {
	return domain.Project{
		ID:          1,
		NamaProject: "Test Project",
		UserID:      "00000000-0000-0000-0000-000000000001",
		Kategori:    "website",
		Semester:    1,
		Ukuran:      "small",
		PathFile:    "/test/path",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// GetTestProjects returns a slice of test projects
func GetTestProjects() []domain.Project {
	return []domain.Project{
		GetTestProject(),
		{
			ID:          2,
			NamaProject: "Another Project",
			UserID:      "00000000-0000-0000-0000-000000000001",
			Kategori:    "mobile",
			Semester:    2,
			Ukuran:      "medium",
			PathFile:    "/test/path2",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}
}

// Modul fixtures

// GetTestModul returns a test modul entity
func GetTestModul() domain.Modul {
	return domain.Modul{
		ID:        "550e8400-e29b-41d4-a716-446655440001",
		Judul:     "Test Modul",
		Deskripsi: "Test Deskripsi",
		UserID:    "00000000-0000-0000-0000-000000000001",
		FileName:  "test.pdf",
		FilePath:  "/test/modul",
		FileSize:  1024,
		MimeType:  "application/pdf",
		Status:    "completed",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// GetTestModuls returns a slice of test moduls
func GetTestModuls() []domain.Modul {
	return []domain.Modul{
		GetTestModul(),
		{
			ID:        "550e8400-e29b-41d4-a716-446655440002",
			Judul:     "Another Modul",
			Deskripsi: "Another Deskripsi",
			UserID:    "00000000-0000-0000-0000-000000000001",
			FileName:  "another.mp4",
			FilePath:  "/test/modul2",
			FileSize:  2048,
			MimeType:  "video/mp4",
			Status:    "completed",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
}

// Request DTO fixtures

// GetTestRegisterRequest returns a test register request
func GetTestRegisterRequest() dto.RegisterRequest {
	return dto.RegisterRequest{
		Name:     "New User",
		Email:    "newuser@example.com",
		Password: "password123",
	}
}

// GetTestAuthRequest returns a test login request
func GetTestAuthRequest() dto.AuthRequest {
	return dto.AuthRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
}

// GetTestCreateProjectRequest returns a test create project request
func GetTestCreateProjectRequest() dto.CreateProjectRequest {
	return dto.CreateProjectRequest{
		NamaProject: "New Project",
		Semester:    1,
	}
}

// GetTestUpdateProjectRequest returns a test update project request
func GetTestUpdateProjectRequest() dto.UpdateProjectRequest {
	return dto.UpdateProjectRequest{
		NamaProject: "Updated Project",
		Kategori:    "website",
		Semester:    1,
	}
}

// GetTestCreateModulRequest returns a test create modul request
func GetTestCreateModulRequest() dto.TusModulUploadInitRequest {
	return dto.TusModulUploadInitRequest{
		Judul:     "New Modul",
		Deskripsi: "Test Deskripsi",
	}
}

// GetTestUpdateModulRequest returns a test update modul request
func GetTestUpdateModulRequest() dto.UpdateModulRequest {
	return dto.UpdateModulRequest{
		Judul:     "Updated Modul",
		Deskripsi: "Updated Deskripsi",
	}
}

// Response DTO fixtures

// GetTestAuthResponse returns a test auth response
func GetTestAuthResponse() dto.AuthResponse {
	return dto.AuthResponse{
		User: &dto.AuthUserResponse{
			ID:    "00000000-0000-0000-0000-000000000001",
			Name:  "Test User",
			Email: "test@example.com",
		},
		AccessToken: "test_access_token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
		ExpiresAt:   time.Now().Add(time.Hour).Unix(),
	}
}

// GetTestProjectResponse returns a test project response
func GetTestProjectResponse() dto.ProjectResponse {
	return dto.ProjectResponse{
		ID:          1,
		NamaProject: "Test Project",
		Kategori:    "website",
		Semester:    1,
		Ukuran:      "small",
		PathFile:    "/test/path",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// GetTestModulResponse returns a test modul response
func GetTestModulResponse() dto.ModulResponse {
	return dto.ModulResponse{
		ID:        "550e8400-e29b-41d4-a716-446655440001",
		Judul:     "Test Modul",
		Deskripsi: "Test Deskripsi",
		FileName:  "test.pdf",
		MimeType:  "application/pdf",
		FileSize:  1024,
		Status:    "completed",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// Health check fixtures

// GetTestBasicHealthCheck returns a test basic health check
func GetTestBasicHealthCheck() dto.BasicHealthCheck {
	return dto.BasicHealthCheck{
		Status:    dto.HealthStatusHealthy,
		App:       "test-app",
		Timestamp: time.Now(),
	}
}

// GetTestComprehensiveHealthCheck returns a test comprehensive health check
func GetTestComprehensiveHealthCheck() dto.ComprehensiveHealthCheck {
	return dto.ComprehensiveHealthCheck{
		Status: dto.HealthStatusHealthy,
		App: dto.AppInfo{
			Name:        "test-app",
			Version:     "1.0.0",
			Environment: "test",
			Uptime:      "1h",
		},
		Database: dto.DatabaseStatus{
			Status:   dto.ServiceStatusConnected,
			PingTime: "2ms",
		},
		System: dto.SystemInfo{
			MemoryUsage: "45MB",
			CPUCores:    4,
			Goroutines:  10,
		},
		Timestamp: time.Now(),
	}
}

// Statistics fixtures

// GetTestSystemMetrics returns a test system metrics
func GetTestSystemMetrics() dto.SystemMetrics {
	return dto.SystemMetrics{
		App: dto.AppInfo{
			Name:        "test-app",
			Version:     "1.0.0",
			Environment: "test",
			Uptime:      "1h",
		},
		System: dto.DetailedSystemInfo{
			Memory: dto.MemoryInfo{
				Allocated:      "45MB",
				TotalAllocated: "120MB",
				System:         "256MB",
				GCCount:        15,
			},
			CPU: dto.CPUInfo{
				Cores:      4,
				Goroutines: 10,
			},
			Runtime: dto.RuntimeInfo{
				GoVersion: "go1.21",
				Compiler:  "gc",
				Arch:      "amd64",
				OS:        "linux",
			},
		},
		Database: dto.DatabaseStatus{
			Status:   dto.ServiceStatusConnected,
			PingTime: "2ms",
		},
		Http: dto.HttpMetrics{
			TotalRequests:  5420,
			ActiveRequests: 3,
			ResponseTimes: dto.ResponseTimes{
				Min: "5ms",
				Max: "150ms",
				Avg: "25ms",
			},
		},
	}
}

// GetTestStatistics returns a test statistics response
func GetTestStatistics() dto.StatisticResponse {
	return dto.StatisticResponse{
		Data: dto.StatisticData{
			TotalProject: intPtr(45),
			TotalModul:   intPtr(120),
			TotalUser:    intPtr(150),
			TotalRole:    intPtr(3),
		},
	}
}

// Helper function to create pointer to int
func intPtr(i int) *int {
	return &i
}
