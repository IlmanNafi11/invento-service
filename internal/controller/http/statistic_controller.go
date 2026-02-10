package http

import (
	"fiber-boiler-plate/internal/helper"
	"fiber-boiler-plate/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

// StatisticController handles statistics endpoint
type StatisticController struct {
	statisticUsecase usecase.StatisticUsecase
}

// NewStatisticController creates a new statistic controller instance
func NewStatisticController(statisticUsecase usecase.StatisticUsecase) *StatisticController {
	return &StatisticController{
		statisticUsecase: statisticUsecase,
	}
}

// GetStatistics retrieves statistics based on user role and permissions
//
//	@Summary		Get User Statistics
//	@Description	Mengambil data statistik berdasarkan role dan permission user.
//	@Description	-
//	@Description	Data yang dikembalikan bersifat dinamis:
//	@Description	  - **Admin** mendapatkan semua statistik (project, modul, user, role)
//	@Description	  - **User biasa** hanya mendapatkan statistik yang mereka miliki aksesnya
//	@Description	  - Field yang tidak memiliki permission akan bernilai null/omit
//	@Description	-
//	@Description	**Permission Mapping:**
//	@Description	  - `total_project`: memerlukan permission "Project:read"
//	@Description	  - `total_modul`: memerlukan permission "Modul:read"
//	@Description	  - `total_user`: memerlukan permission "User:read"
//	@Description	  - `total_role`: memerlukan permission "Role:read"
//	@Tags			Statistics
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	domain.SuccessResponse{data=domain.StatisticData}	"Data statistik berhasil diambil"
//	@Success		200	{object}	domain.SuccessResponse{data=domain.StatisticData}	"Data statistik berhasil diambil (partial untuk user biasa)"
//	@Failure		401	{object}	domain.ErrorResponse	"Tidak memiliki akses"
//	@Failure		500	{object}	domain.ErrorResponse	"Terjadi kesalahan pada server"
//	@Router			/api/v1/statistic [get]
func (ctrl *StatisticController) GetStatistics(c *fiber.Ctx) error {
	// Extract authenticated user ID
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		return helper.SendUnauthorizedResponse(c)
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		return helper.SendUnauthorizedResponse(c)
	}

	// Extract authenticated user role
	userRoleVal := c.Locals("user_role")
	if userRoleVal == nil {
		return helper.SendUnauthorizedResponse(c)
	}
	userRole, ok := userRoleVal.(string)
	if !ok {
		return helper.SendUnauthorizedResponse(c)
	}

	// Get statistics based on user role and permissions
	data, err := ctrl.statisticUsecase.GetStatistics(userID, userRole)
	if err != nil {
		return helper.SendInternalServerErrorResponse(c)
	}

	// Handle empty statistics gracefully - return success with empty/partial data
	// This allows frontend to display appropriate UI based on available permissions
	return helper.SendSuccessResponse(c, helper.StatusOK, "Data statistik berhasil diambil", data)
}
