package http

import (
	"fiber-boiler-plate/internal/helper"
	"fiber-boiler-plate/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

type StatisticController struct {
	statisticUsecase usecase.StatisticUsecase
}

func NewStatisticController(statisticUsecase usecase.StatisticUsecase) *StatisticController {
	return &StatisticController{
		statisticUsecase: statisticUsecase,
	}
}

func (ctrl *StatisticController) GetStatistics(c *fiber.Ctx) error {
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		return helper.SendUnauthorizedResponse(c)
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		return helper.SendUnauthorizedResponse(c)
	}

	userRoleVal := c.Locals("user_role")
	if userRoleVal == nil {
		return helper.SendUnauthorizedResponse(c)
	}
	userRole, ok := userRoleVal.(string)
	if !ok {
		return helper.SendUnauthorizedResponse(c)
	}

	data, err := ctrl.statisticUsecase.GetStatistics(userID, userRole)
	if err != nil {
		return helper.SendInternalServerErrorResponse(c)
	}

	return helper.SendSuccessResponse(c, helper.StatusOK, "Data statistik berhasil diambil", data)
}
