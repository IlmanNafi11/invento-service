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

	return helper.SendSuccessResponse(c, helper.StatusOK, "Daftar resource dan permission berhasil diambil", response)
}

func (ctrl *RoleController) GetRoleList(c *fiber.Ctx) error {
	var params domain.RoleListQueryParams
	if err := c.QueryParser(&params); err != nil {
		return helper.SendBadRequestResponse(c, "Parameter query tidak valid")
	}

	result, err := ctrl.roleUsecase.GetRoleList(params)
	if err != nil {
		return helper.SendInternalServerErrorResponse(c)
	}

	return helper.SendSuccessResponse(c, helper.StatusOK, "Daftar role berhasil diambil", result)
}

func (ctrl *RoleController) CreateRole(c *fiber.Ctx) error {
	var req domain.RoleCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.SendBadRequestResponse(c, "Format request tidak valid")
	}

	if validationErrors := helper.ValidateStruct(req); len(validationErrors) > 0 {
		return helper.SendValidationErrorResponse(c, validationErrors)
	}

	result, err := ctrl.roleUsecase.CreateRole(req)
	if err != nil {
		switch err.Error() {
		case "nama role sudah ada":
			return helper.SendConflictResponse(c, err.Error())
		case "permission tidak boleh kosong":
			return helper.SendBadRequestResponse(c, err.Error())
		default:
			if len(err.Error()) > 20 && err.Error()[:6] == "action" {
				return helper.SendBadRequestResponse(c, err.Error())
			}
			return helper.SendInternalServerErrorResponse(c)
		}
	}

	return helper.SendSuccessResponse(c, helper.StatusCreated, "Role berhasil dibuat", result)
}

func (ctrl *RoleController) GetRoleDetail(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return helper.SendBadRequestResponse(c, "ID tidak valid")
	}

	result, err := ctrl.roleUsecase.GetRoleDetail(uint(id))
	if err != nil {
		if err.Error() == "role tidak ditemukan" {
			return helper.SendNotFoundResponse(c, err.Error())
		}
		return helper.SendInternalServerErrorResponse(c)
	}

	return helper.SendSuccessResponse(c, helper.StatusOK, "Detail role berhasil diambil", result)
}

func (ctrl *RoleController) UpdateRole(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return helper.SendBadRequestResponse(c, "ID tidak valid")
	}

	var req domain.RoleUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.SendBadRequestResponse(c, "Format request tidak valid")
	}

	if validationErrors := helper.ValidateStruct(req); len(validationErrors) > 0 {
		return helper.SendValidationErrorResponse(c, validationErrors)
	}

	result, err := ctrl.roleUsecase.UpdateRole(uint(id), req)
	if err != nil {
		switch err.Error() {
		case "role tidak ditemukan":
			return helper.SendNotFoundResponse(c, err.Error())
		case "nama role sudah ada":
			return helper.SendConflictResponse(c, err.Error())
		case "permission tidak boleh kosong":
			return helper.SendBadRequestResponse(c, err.Error())
		default:
			if len(err.Error()) > 20 && err.Error()[:6] == "action" {
				return helper.SendBadRequestResponse(c, err.Error())
			}
			return helper.SendInternalServerErrorResponse(c)
		}
	}

	return helper.SendSuccessResponse(c, helper.StatusOK, "Role berhasil diperbarui", result)
}

func (ctrl *RoleController) DeleteRole(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return helper.SendBadRequestResponse(c, "ID tidak valid")
	}

	err = ctrl.roleUsecase.DeleteRole(uint(id))
	if err != nil {
		if err.Error() == "role tidak ditemukan" {
			return helper.SendNotFoundResponse(c, err.Error())
		}
		return helper.SendInternalServerErrorResponse(c)
	}

	return helper.SendSuccessResponse(c, helper.StatusOK, "Role berhasil dihapus", nil)
}
