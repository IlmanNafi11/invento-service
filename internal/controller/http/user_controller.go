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
		return helper.SendBadRequestResponse(c, "Parameter query tidak valid")
	}

	result, err := ctrl.userUsecase.GetUserList(params)
	if err != nil {
		return helper.SendInternalServerErrorResponse(c)
	}

	return helper.SendSuccessResponse(c, helper.StatusOK, "Daftar user berhasil diambil", result)
}

func (ctrl *UserController) UpdateUserRole(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return helper.SendBadRequestResponse(c, "ID tidak valid")
	}

	var req domain.UpdateUserRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.SendBadRequestResponse(c, "Format request tidak valid")
	}

	if validationErrors := helper.ValidateStruct(req); len(validationErrors) > 0 {
		return helper.SendValidationErrorResponse(c, validationErrors)
	}

	err = ctrl.userUsecase.UpdateUserRole(uint(id), req.Role)
	if err != nil {
		switch err.Error() {
		case "user tidak ditemukan", "role tidak ditemukan":
			return helper.SendNotFoundResponse(c, err.Error())
		default:
			return helper.SendInternalServerErrorResponse(c)
		}
	}

	return helper.SendSuccessResponse(c, helper.StatusOK, "Role user berhasil diperbarui", nil)
}

func (ctrl *UserController) DeleteUser(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return helper.SendBadRequestResponse(c, "ID tidak valid")
	}

	err = ctrl.userUsecase.DeleteUser(uint(id))
	if err != nil {
		if err.Error() == "user tidak ditemukan" {
			return helper.SendNotFoundResponse(c, err.Error())
		}
		return helper.SendInternalServerErrorResponse(c)
	}

	return helper.SendSuccessResponse(c, helper.StatusOK, "User berhasil dihapus", nil)
}

func (ctrl *UserController) GetUserFiles(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return helper.SendBadRequestResponse(c, "ID tidak valid")
	}

	var params domain.UserFilesQueryParams
	if err := c.QueryParser(&params); err != nil {
		return helper.SendBadRequestResponse(c, "Parameter query tidak valid")
	}

	result, err := ctrl.userUsecase.GetUserFiles(uint(id), params)
	if err != nil {
		if err.Error() == "user tidak ditemukan" {
			return helper.SendNotFoundResponse(c, err.Error())
		}
		return helper.SendInternalServerErrorResponse(c)
	}

	return helper.SendSuccessResponse(c, helper.StatusOK, "Daftar file user berhasil diambil", result)
}

func (ctrl *UserController) GetProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	result, err := ctrl.userUsecase.GetProfile(userID)
	if err != nil {
		return helper.SendInternalServerErrorResponse(c)
	}

	return helper.SendSuccessResponse(c, helper.StatusOK, "Profil user berhasil diambil", result)
}

func (ctrl *UserController) UpdateProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var req domain.UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.SendBadRequestResponse(c, "Format request tidak valid")
	}

	if validationErrors := helper.ValidateStruct(req); len(validationErrors) > 0 {
		return helper.SendValidationErrorResponse(c, validationErrors)
	}

	fotoProfil, errFile := c.FormFile("foto_profil")

	var fotoProfilToPass interface{}
	if errFile == nil && fotoProfil != nil {
		fotoProfilToPass = fotoProfil
	} else {
		fotoProfilToPass = nil
	}

	result, err := ctrl.userUsecase.UpdateProfile(userID, req, fotoProfilToPass)
	if err != nil {
		switch err.Error() {
		case "format foto profil harus png, jpg, atau jpeg", "ukuran foto profil tidak boleh lebih dari 2MB":
			return helper.SendBadRequestResponse(c, err.Error())
		default:
			return helper.SendInternalServerErrorResponse(c)
		}
	}

	return helper.SendSuccessResponse(c, helper.StatusOK, "Profil berhasil diperbarui", result)
}

func (ctrl *UserController) GetUserPermissions(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	result, err := ctrl.userUsecase.GetUserPermissions(userID)
	if err != nil {
		return helper.SendInternalServerErrorResponse(c)
	}

	return helper.SendSuccessResponse(c, helper.StatusOK, "Permissions user berhasil diambil", result)
}

func (ctrl *UserController) DownloadUserFiles(c *fiber.Ctx) error {
	idParam := c.Params("id")
	ownerUserID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return helper.SendBadRequestResponse(c, "ID tidak valid")
	}

	var req domain.DownloadUserFilesRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.SendBadRequestResponse(c, "Format request tidak valid")
	}

	if len(req.ProjectIDs) == 0 && len(req.ModulIDs) == 0 {
		return helper.SendBadRequestResponse(c, "Project IDs atau Modul IDs harus diisi minimal salah satu")
	}

	filePath, err := ctrl.userUsecase.DownloadUserFiles(uint(ownerUserID), req.ProjectIDs, req.ModulIDs)
	if err != nil {
		if err.Error() == "user tidak ditemukan" {
			return helper.SendNotFoundResponse(c, err.Error())
		}
		if err.Error() == "file tidak ditemukan" {
			return helper.SendNotFoundResponse(c, err.Error())
		}
		return helper.SendInternalServerErrorResponse(c)
	}

	return c.Download(filePath)
}
