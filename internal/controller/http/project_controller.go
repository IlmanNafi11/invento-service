package http

import (
	"fiber-boiler-plate/internal/domain"
	"fiber-boiler-plate/internal/helper"
	"fiber-boiler-plate/internal/usecase"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type ProjectController struct {
	projectUsecase usecase.ProjectUsecase
}

func NewProjectController(projectUsecase usecase.ProjectUsecase) *ProjectController {
	return &ProjectController{
		projectUsecase: projectUsecase,
	}
}

func (ctrl *ProjectController) UpdateMetadata(c *fiber.Ctx) error {
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		return helper.SendUnauthorizedResponse(c)
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		return helper.SendUnauthorizedResponse(c)
	}

	projectID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "ID project tidak valid", nil)
	}

	var req domain.ProjectUpdateMetadataRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "Format request tidak valid", nil)
	}

	err = ctrl.projectUsecase.UpdateMetadata(uint(projectID), userID, req)
	if err != nil {
		if err.Error() == "project tidak ditemukan" {
			return helper.SendNotFoundResponse(c, err.Error())
		}
		if err.Error() == "tidak memiliki akses ke project ini" {
			return helper.SendForbiddenResponse(c)
		}
		return helper.SendErrorResponse(c, fiber.StatusInternalServerError, "Gagal update project: "+err.Error(), nil)
	}

	return helper.SendSuccessResponse(c, fiber.StatusOK, "Metadata project berhasil diperbarui", nil)
}

func (ctrl *ProjectController) GetByID(c *fiber.Ctx) error {
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		return helper.SendUnauthorizedResponse(c)
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		return helper.SendUnauthorizedResponse(c)
	}

	projectID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "ID project tidak valid", nil)
	}

	result, err := ctrl.projectUsecase.GetByID(uint(projectID), userID)
	if err != nil {
		if err.Error() == "project tidak ditemukan" {
			return helper.SendNotFoundResponse(c, err.Error())
		}
		if err.Error() == "tidak memiliki akses ke project ini" {
			return helper.SendForbiddenResponse(c)
		}
		return helper.SendInternalServerErrorResponse(c)
	}

	return helper.SendSuccessResponse(c, fiber.StatusOK, "Detail project berhasil diambil", result)
}

func (ctrl *ProjectController) GetList(c *fiber.Ctx) error {
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		return helper.SendUnauthorizedResponse(c)
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		return helper.SendUnauthorizedResponse(c)
	}

	var params domain.ProjectListQueryParams
	if err := c.QueryParser(&params); err != nil {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "Parameter query tidak valid", nil)
	}

	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 10
	}

	result, err := ctrl.projectUsecase.GetList(userID, params.Search, params.FilterSemester, params.FilterKategori, params.Page, params.Limit)
	if err != nil {
		return helper.SendErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil daftar project: "+err.Error(), nil)
	}

	return helper.SendSuccessResponse(c, fiber.StatusOK, "Daftar project berhasil diambil", result)
}

func (ctrl *ProjectController) Delete(c *fiber.Ctx) error {
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		return helper.SendUnauthorizedResponse(c)
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		return helper.SendUnauthorizedResponse(c)
	}

	projectID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "ID project tidak valid", nil)
	}

	err = ctrl.projectUsecase.Delete(uint(projectID), userID)
	if err != nil {
		if err.Error() == "project tidak ditemukan" {
			return helper.SendNotFoundResponse(c, err.Error())
		}
		if err.Error() == "tidak memiliki akses ke project ini" {
			return helper.SendForbiddenResponse(c)
		}
		return helper.SendInternalServerErrorResponse(c)
	}

	return helper.SendSuccessResponse(c, fiber.StatusOK, "Project berhasil dihapus", nil)
}

func (ctrl *ProjectController) Download(c *fiber.Ctx) error {
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		return helper.SendUnauthorizedResponse(c)
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		return helper.SendUnauthorizedResponse(c)
	}

	var req domain.ProjectDownloadRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "Format request tidak valid", nil)
	}

	if len(req.IDs) == 0 {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "ID project tidak boleh kosong", nil)
	}

	filePath, err := ctrl.projectUsecase.Download(userID, req.IDs)
	if err != nil {
		if err.Error() == "project tidak ditemukan" {
			return helper.SendNotFoundResponse(c, err.Error())
		}
		return helper.SendInternalServerErrorResponse(c)
	}

	return c.Download(filePath)
}
