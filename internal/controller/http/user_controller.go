package http

import (
	"fiber-boiler-plate/internal/domain"
	"fiber-boiler-plate/internal/helper"
	"fiber-boiler-plate/internal/usecase"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type UserController struct {
	userUsecase usecase.UserUsecase
}

func NewUserController(userUsecase usecase.UserUsecase) *UserController {
	return &UserController{
		userUsecase: userUsecase,
	}
}

func (ctrl *UserController) GetUserList(c *fiber.Ctx) error {
	var params domain.UserListQueryParams
	if err := c.QueryParser(&params); err != nil {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "Parameter query tidak valid", nil)
	}

	result, err := ctrl.userUsecase.GetUserList(params)
	if err != nil {
		return helper.SendInternalServerErrorResponse(c)
	}

	return helper.SendSuccessResponse(c, fiber.StatusOK, "Daftar user berhasil diambil", result)
}

func (ctrl *UserController) UpdateUserRole(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "ID tidak valid", nil)
	}

	var req domain.UpdateUserRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "Format request tidak valid", nil)
	}

	if validationErrors := helper.ValidateStruct(req); len(validationErrors) > 0 {
		return helper.SendValidationErrorResponse(c, validationErrors)
	}

	err = ctrl.userUsecase.UpdateUserRole(uint(id), req.Role)
	if err != nil {
		if err.Error() == "user tidak ditemukan" {
			return helper.SendNotFoundResponse(c, err.Error())
		}
		if err.Error() == "role tidak ditemukan" {
			return helper.SendNotFoundResponse(c, err.Error())
		}
		return helper.SendInternalServerErrorResponse(c)
	}

	return helper.SendSuccessResponse(c, fiber.StatusOK, "Role user berhasil diperbarui", nil)
}

func (ctrl *UserController) DeleteUser(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "ID tidak valid", nil)
	}

	err = ctrl.userUsecase.DeleteUser(uint(id))
	if err != nil {
		if err.Error() == "user tidak ditemukan" {
			return helper.SendNotFoundResponse(c, err.Error())
		}
		return helper.SendInternalServerErrorResponse(c)
	}

	return helper.SendSuccessResponse(c, fiber.StatusOK, "User berhasil dihapus", nil)
}

func (ctrl *UserController) GetUserFiles(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "ID tidak valid", nil)
	}

	var params domain.UserFilesQueryParams
	if err := c.QueryParser(&params); err != nil {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "Parameter query tidak valid", nil)
	}

	result, err := ctrl.userUsecase.GetUserFiles(uint(id), params)
	if err != nil {
		if err.Error() == "user tidak ditemukan" {
			return helper.SendNotFoundResponse(c, err.Error())
		}
		return helper.SendInternalServerErrorResponse(c)
	}

	return helper.SendSuccessResponse(c, fiber.StatusOK, "Daftar file user berhasil diambil", result)
}

func (ctrl *UserController) GetProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	result, err := ctrl.userUsecase.GetProfile(userID)
	if err != nil {
		return helper.SendInternalServerErrorResponse(c)
	}

	return helper.SendSuccessResponse(c, fiber.StatusOK, "Profil user berhasil diambil", result)
}

func (ctrl *UserController) GetUserPermissions(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	result, err := ctrl.userUsecase.GetUserPermissions(userID)
	if err != nil {
		return helper.SendInternalServerErrorResponse(c)
	}

	return helper.SendSuccessResponse(c, fiber.StatusOK, "Permissions user berhasil diambil", result)
}
