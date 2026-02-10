package http

import (
	"errors"
	"fiber-boiler-plate/internal/controller/base"
	"fiber-boiler-plate/internal/domain"
	"fiber-boiler-plate/internal/helper"
	apperrors "fiber-boiler-plate/internal/errors"
	"fiber-boiler-plate/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

// UserController handles user-related HTTP requests
type UserController struct {
	*base.BaseController
	userUsecase usecase.UserUsecase
}

// NewUserController creates a new UserController instance
func NewUserController(userUsecase usecase.UserUsecase) *UserController {
	return &UserController{
		BaseController: base.NewBaseController(nil, nil),
		userUsecase:    userUsecase,
	}
}

// GetUserList handles GET /api/v1/user - List users with pagination
// @Summary Get list of users
// @Description Get paginated list of users with optional search and filters
// @Tags User Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search keyword"
// @Success 200 {object} dto.SuccessResponse{data=domain.UserListData}
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/user [get]
func (ctrl *UserController) GetUserList(c *fiber.Ctx) error {
	var params domain.UserListQueryParams
	if err := c.QueryParser(&params); err != nil {
		return ctrl.SendBadRequest(c, "Parameter query tidak valid")
	}

	result, err := ctrl.userUsecase.GetUserList(params)
	if err != nil {
		return ctrl.SendInternalError(c)
	}

	return helper.SendSuccessResponse(c, helper.StatusOK, "Daftar user berhasil diambil", result)
}

// UpdateUserRole handles PUT /api/v1/user/{id}/role - Update user role
// @Summary Update user role
// @Description Update the role of a specific user
// @Tags User Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param request body domain.UpdateUserRoleRequest true "Role update request"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/user/{id}/role [put]
func (ctrl *UserController) UpdateUserRole(c *fiber.Ctx) error {
	id, err := ctrl.ParsePathID(c)
	if err != nil {
		return err // error response already sent
	}

	var req domain.UpdateUserRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return ctrl.SendBadRequest(c, "Format request tidak valid")
	}

	if !ctrl.ValidateStruct(c, req) {
		return nil // validation error response already sent
	}

	err = ctrl.userUsecase.UpdateUserRole(id, req.Role)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return helper.SendAppError(c, appErr)
		}
		return ctrl.SendInternalError(c)
	}

	return ctrl.SendSuccess(c, nil, "Role user berhasil diperbarui")
}

// DeleteUser handles DELETE /api/v1/user/{id} - Delete a user
// @Summary Delete user
// @Description Delete a specific user by ID
// @Tags User Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/user/{id} [delete]
func (ctrl *UserController) DeleteUser(c *fiber.Ctx) error {
	id, err := ctrl.ParsePathID(c)
	if err != nil {
		return err // error response already sent
	}

	err = ctrl.userUsecase.DeleteUser(id)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return helper.SendAppError(c, appErr)
		}
		return ctrl.SendInternalError(c)
	}

	return ctrl.SendSuccess(c, nil, "User berhasil dihapus")
}

// GetUserFiles handles GET /api/v1/user/{id}/files - Get user files
// @Summary Get user files
// @Description Get list of files owned by a specific user
// @Tags User Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} dto.SuccessResponse{data=domain.UserFilesData}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/user/{id}/files [get]
func (ctrl *UserController) GetUserFiles(c *fiber.Ctx) error {
	id, err := ctrl.ParsePathID(c)
	if err != nil {
		return err // error response already sent
	}

	var params domain.UserFilesQueryParams
	if err := c.QueryParser(&params); err != nil {
		return ctrl.SendBadRequest(c, "Parameter query tidak valid")
	}

	result, err := ctrl.userUsecase.GetUserFiles(id, params)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return helper.SendAppError(c, appErr)
		}
		return ctrl.SendInternalError(c)
	}

	return helper.SendSuccessResponse(c, helper.StatusOK, "Daftar file user berhasil diambil", result)
}

// GetProfile handles GET /api/v1/profile - Get current user profile
// @Summary Get user profile
// @Description Get the profile of the authenticated user
// @Tags User Profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.SuccessResponse{data=domain.ProfileData}
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/profile [get]
func (ctrl *UserController) GetProfile(c *fiber.Ctx) error {
	userID := ctrl.GetAuthenticatedUserID(c)
	if userID == 0 {
		return nil // unauthorized response already sent
	}

	result, err := ctrl.userUsecase.GetProfile(userID)
	if err != nil {
		return ctrl.SendInternalError(c)
	}

	return ctrl.SendSuccess(c, result, "Profil user berhasil diambil")
}

// UpdateProfile handles PUT /api/v1/profile - Update current user profile
// @Summary Update user profile
// @Description Update the profile of the authenticated user
// @Tags User Profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.UpdateProfileRequest true "Profile update request"
// @Param foto_profil formData file false "Profile photo (optional)"
// @Success 200 {object} dto.SuccessResponse{data=domain.ProfileData}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/profile [put]
func (ctrl *UserController) UpdateProfile(c *fiber.Ctx) error {
	userID := ctrl.GetAuthenticatedUserID(c)
	if userID == 0 {
		return nil // unauthorized response already sent
	}

	var req domain.UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return ctrl.SendBadRequest(c, "Format request tidak valid")
	}

	if !ctrl.ValidateStruct(c, req) {
		return nil // validation error response already sent
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
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return helper.SendAppError(c, appErr)
		}
		return ctrl.SendInternalError(c)
	}

	return ctrl.SendSuccess(c, result, "Profil berhasil diperbarui")
}

// GetUserPermissions handles GET /api/v1/user/permissions - Get current user permissions
// @Summary Get user permissions
// @Description Get the permissions of the authenticated user
// @Tags User Profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.SuccessResponse{data=[]domain.UserPermissionItem}
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/user/permissions [get]
func (ctrl *UserController) GetUserPermissions(c *fiber.Ctx) error {
	userID := ctrl.GetAuthenticatedUserID(c)
	if userID == 0 {
		return nil // unauthorized response already sent
	}

	result, err := ctrl.userUsecase.GetUserPermissions(userID)
	if err != nil {
		return ctrl.SendInternalError(c)
	}

	return ctrl.SendSuccess(c, result, "Permissions user berhasil diambil")
}

// DownloadUserFiles handles POST /api/v1/user/{id}/download - Download user files
// @Summary Download user files
// @Description Download files owned by a specific user
// @Tags User Management
// @Accept json
// @Produce application/octet-stream
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param request body domain.DownloadUserFilesRequest true "Download request"
// @Success 200 {file} binary
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/user/{id}/download [post]
func (ctrl *UserController) DownloadUserFiles(c *fiber.Ctx) error {
	ownerUserID, err := ctrl.ParsePathID(c)
	if err != nil {
		return err // error response already sent
	}

	var req domain.DownloadUserFilesRequest
	if err := c.BodyParser(&req); err != nil {
		return ctrl.SendBadRequest(c, "Format request tidak valid")
	}

	if len(req.ProjectIDs) == 0 && len(req.ModulIDs) == 0 {
		return ctrl.SendBadRequest(c, "Project IDs atau Modul IDs harus diisi minimal salah satu")
	}

	filePath, err := ctrl.userUsecase.DownloadUserFiles(ownerUserID, req.ProjectIDs, req.ModulIDs)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return helper.SendAppError(c, appErr)
		}
		return ctrl.SendInternalError(c)
	}

	return c.Download(filePath)
}
