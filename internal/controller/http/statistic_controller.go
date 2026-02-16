package http

import (
	"errors"
	"invento-service/internal/controller/base"
	apperrors "invento-service/internal/errors"
	"invento-service/internal/httputil"
	"invento-service/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

// StatisticController handles statistics endpoint.
//
// Singular endpoint (/statistic) rationale:
// - /statistic refers to aggregated statistics as a single conceptual resource
// - Returns a single aggregated data object, not a collection of statistic items
// - Represents one "statistics view" for the current user based on permissions
// - Semantically correct to use singular for aggregated singleton concepts
type StatisticController struct {
	*base.BaseController
	statisticUsecase usecase.StatisticUsecase
}

// NewStatisticController creates a new statistic controller instance
func NewStatisticController(statisticUsecase usecase.StatisticUsecase) *StatisticController {
	return &StatisticController{
		BaseController:   base.NewBaseController("", nil),
		statisticUsecase: statisticUsecase,
	}
}

// GetStatistics retrieves statistics based on user role and permissions
// @Summary Get User Statistics
// @Description Mengambil data statistik berdasarkan role dan permission user.
// @Description -
// @Description Data yang dikembalikan bersifat dinamis:
// @Description   - **Admin** mendapatkan semua statistik (project, modul, user, role)
// @Description   - **User biasa** hanya mendapatkan statistik yang mereka miliki aksesnya
// @Description   - Field yang tidak memiliki permission akan bernilai null/omit
// @Description -
// @Description **Permission Mapping:**
// @Description   - `total_project`: memerlukan permission "Project:read"
// @Description   - `total_modul`: memerlukan permission "Modul:read"
// @Description   - `total_user`: memerlukan permission "User:read"
// @Description   - `total_role`: memerlukan permission "Role:read"
// @Tags Statistics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.SuccessResponse{data=domain.StatisticData} "Data statistik berhasil diambil"
// @Success 200 {object} domain.SuccessResponse{data=domain.StatisticData} "Data statistik berhasil diambil (partial untuk user biasa)"
// @Failure 401 {object} domain.ErrorResponse "Tidak memiliki akses"
// @Failure 500 {object} domain.ErrorResponse "Terjadi kesalahan pada server"
// @Router /statistic [get]
func (ctrl *StatisticController) GetStatistics(c *fiber.Ctx) error {
	// Extract authenticated user ID
	userID := ctrl.GetAuthenticatedUserID(c)
	if userID == "" {
		return nil // unauthorized response already sent
	}

	// Extract authenticated user role
	userRole := ctrl.GetAuthenticatedUserRole(c)
	if userRole == "" {
		return nil // unauthorized response already sent
	}

	// Get statistics based on user role and permissions
	data, err := ctrl.statisticUsecase.GetStatistics(userID, userRole)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return httputil.SendAppError(c, appErr)
		}
		return ctrl.SendInternalError(c)
	}

	// Handle empty statistics gracefully - return success with empty/partial data
	// This allows frontend to display appropriate UI based on available permissions
	return ctrl.SendSuccess(c, data, "Data statistik berhasil diambil")
}
