package http

import (
	"fiber-boiler-plate/internal/domain"
	"fiber-boiler-plate/internal/helper"
	"fiber-boiler-plate/internal/usecase"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type RoleController struct {
	roleUsecase usecase.RoleUsecase
}

func NewRoleController(roleUsecase usecase.RoleUsecase) *RoleController {
	return &RoleController{
		roleUsecase: roleUsecase,
	}
}

func (ctrl *RoleController) GetAvailablePermissions(c *fiber.Ctx) error {
	permissions, err := ctrl.roleUsecase.GetAvailablePermissions()
	if err != nil {
		return helper.SendInternalServerErrorResponse(c)
	}

	response := map[string]interface{}{
		"items": permissions,
	}

	return helper.SendSuccessResponse(c, fiber.StatusOK, "Daftar resource dan permission berhasil diambil", response)
}

func (ctrl *RoleController) GetRoleList(c *fiber.Ctx) error {
	var params domain.RoleListQueryParams
	if err := c.QueryParser(&params); err != nil {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "Parameter query tidak valid", nil)
	}

	result, err := ctrl.roleUsecase.GetRoleList(params)
	if err != nil {
		return helper.SendInternalServerErrorResponse(c)
	}

	return helper.SendSuccessResponse(c, fiber.StatusOK, "Daftar role berhasil diambil", result)
}

func (ctrl *RoleController) CreateRole(c *fiber.Ctx) error {
	var req domain.RoleCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "Format request tidak valid", nil)
	}

	if validationErrors := helper.ValidateStruct(req); len(validationErrors) > 0 {
		return helper.SendValidationErrorResponse(c, validationErrors)
	}

	result, err := ctrl.roleUsecase.CreateRole(req)
	if err != nil {
		if err.Error() == "nama role sudah ada" {
			return helper.SendErrorResponse(c, fiber.StatusConflict, err.Error(), nil)
		}
		return helper.SendInternalServerErrorResponse(c)
	}

	return helper.SendSuccessResponse(c, fiber.StatusCreated, "Role berhasil dibuat", result)
}

func (ctrl *RoleController) GetRoleDetail(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "ID tidak valid", nil)
	}

	result, err := ctrl.roleUsecase.GetRoleDetail(uint(id))
	if err != nil {
		if err.Error() == "role tidak ditemukan" {
			return helper.SendNotFoundResponse(c, err.Error())
		}
		return helper.SendInternalServerErrorResponse(c)
	}

	return helper.SendSuccessResponse(c, fiber.StatusOK, "Detail role berhasil diambil", result)
}

func (ctrl *RoleController) UpdateRole(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "ID tidak valid", nil)
	}

	var req domain.RoleUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "Format request tidak valid", nil)
	}

	if validationErrors := helper.ValidateStruct(req); len(validationErrors) > 0 {
		return helper.SendValidationErrorResponse(c, validationErrors)
	}

	result, err := ctrl.roleUsecase.UpdateRole(uint(id), req)
	if err != nil {
		if err.Error() == "role tidak ditemukan" {
			return helper.SendNotFoundResponse(c, err.Error())
		}
		if err.Error() == "nama role sudah ada" {
			return helper.SendErrorResponse(c, fiber.StatusConflict, err.Error(), nil)
		}
		return helper.SendInternalServerErrorResponse(c)
	}

	return helper.SendSuccessResponse(c, fiber.StatusOK, "Role berhasil diperbarui", result)
}

func (ctrl *RoleController) DeleteRole(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "ID tidak valid", nil)
	}

	err = ctrl.roleUsecase.DeleteRole(uint(id))
	if err != nil {
		if err.Error() == "role tidak ditemukan" {
			return helper.SendNotFoundResponse(c, err.Error())
		}
		return helper.SendInternalServerErrorResponse(c)
	}

	return helper.SendSuccessResponse(c, fiber.StatusOK, "Role berhasil dihapus", nil)
}
