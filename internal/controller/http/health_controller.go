package http

import (
	"fiber-boiler-plate/internal/helper"
	"fiber-boiler-plate/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

// HealthController handles health check and monitoring endpoints
type HealthController struct {
	healthUsecase usecase.HealthUsecase
}

// NewHealthController creates a new health controller instance
func NewHealthController(healthUsecase usecase.HealthUsecase) *HealthController {
	return &HealthController{
		healthUsecase: healthUsecase,
	}
}

// BasicHealthCheck provides a simple health check endpoint
//
//	@Summary		Basic Health Check
//	@Description	Melakukan pemeriksaan dasar kesehatan server. Mengembalikan status dasar tanpa pengecekan koneksi database.
//	@Tags			Monitoring
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	domain.SuccessResponse{data=domain.BasicHealthCheck}	"Server berjalan dengan baik"
//	@Router			/health [get]
//	@Router			/api/v1/monitoring/health [get]
func (ctrl *HealthController) BasicHealthCheck(c *fiber.Ctx) error {
	healthData := ctrl.healthUsecase.GetBasicHealth()

	return helper.SendSuccessResponse(c, helper.StatusOK, "Server berjalan dengan baik", healthData)
}

// ComprehensiveHealthCheck provides detailed health check including database and system info
//
//	@Summary		Comprehensive Health Check
//	@Description	Melakukan pemeriksaan kesehatan menyeluruh termasuk koneksi database, penggunaan memori, CPU, dan lainnya.
//	@Description	-
//	@Description	Mengembalikan status HTTP 503 jika sistem tidak sehat (database terputus atau error).
//	@Tags			Monitoring
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	domain.SuccessResponse{data=domain.ComprehensiveHealthCheck}	"Sistem sehat"
//	@Failure		503	{object}	domain.ErrorResponse	"Beberapa komponen sistem mengalami masalah"
//	@Router			/api/v1/monitoring/status [get]
func (ctrl *HealthController) ComprehensiveHealthCheck(c *fiber.Ctx) error {
	healthData := ctrl.healthUsecase.GetComprehensiveHealth()

	if healthData.Status == "unhealthy" {
		return helper.SendErrorResponse(c, fiber.StatusServiceUnavailable, "Beberapa komponen sistem mengalami masalah", healthData)
	}

	return helper.SendSuccessResponse(c, helper.StatusOK, "Pemeriksaan kesehatan sistem berhasil", healthData)
}

// GetSystemMetrics provides detailed system metrics for monitoring
//
//	@Summary		System Metrics
//	@Description	Mengambil metrik sistem detail termasuk penggunaan memori, CPU, statistik database, dan metrik HTTP.
//	@Tags			Monitoring
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	domain.SuccessResponse{data=domain.SystemMetrics}	"Metrics sistem berhasil diambil"
//	@Router			/api/v1/monitoring/metrics [get]
func (ctrl *HealthController) GetSystemMetrics(c *fiber.Ctx) error {
	metricsData := ctrl.healthUsecase.GetSystemMetrics()

	return helper.SendSuccessResponse(c, helper.StatusOK, "Metrics sistem berhasil diambil", metricsData)
}

// GetApplicationStatus provides application status and dependencies
//
//	@Summary		Application Status
//	@Description	Mengambil status aplikasi lengkap termasuk uptime, layanan yang aktif, dan dependency eksternal.
//	@Tags			Monitoring
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	domain.SuccessResponse{data=domain.ApplicationStatus}	"Status aplikasi berhasil diambil"
//	@Router			/api/v1/monitoring/app-status [get]
func (ctrl *HealthController) GetApplicationStatus(c *fiber.Ctx) error {
	statusData := ctrl.healthUsecase.GetApplicationStatus()

	return helper.SendSuccessResponse(c, helper.StatusOK, "Status aplikasi berhasil diambil", statusData)
}
