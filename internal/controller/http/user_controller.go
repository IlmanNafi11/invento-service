package http

import (
	"errors"
	"fiber-boiler-plate/internal/controller/base"
	"fiber-boiler-plate/internal/domain"
	apperrors "fiber-boiler-plate/internal/errors"
	"fiber-boiler-plate/internal/helper"
	"fiber-boiler-plate/internal/usecase"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// UserController handles user-related HTTP requests.
//
// Singular endpoint (/profile) rationale:
// - /profile refers to the current authenticated user's single profile (not a collection)
// - Each authenticated user has exactly one profile, no ID parameter needed
// - Semantically correct to use singular for singleton current-user resources
type UserController struct {
	*base.BaseController
	userUsecase usecase.UserUsecase
}

// NewUserController creates a new UserController instance
func NewUserController(userUsecase usecase.UserUsecase) *UserController {
	return &UserController{
		BaseController: base.NewBaseController("", nil),
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
// @Success 200 {object} domain.SuccessResponse{data=domain.UserListData}
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /user [get]
func (ctrl *UserController) GetUserList(c *fiber.Ctx) error {
	var params domain.UserListQueryParams
	if err := c.QueryParser(&params); err != nil {
		return ctrl.SendBadRequest(c, "Parameter query tidak valid")
	}

	result, err := ctrl.userUsecase.GetUserList(params)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return helper.SendAppError(c, appErr)
		}
		return ctrl.SendInternalError(c)
	}

	return ctrl.SendSuccess(c, result, "Daftar user berhasil diambil")
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
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /user/{id}/role [put]
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

	err = ctrl.userUsecase.UpdateUserRole(strconv.FormatUint(uint64(id), 10), req.Role)
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
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /user/{id} [delete]
func (ctrl *UserController) DeleteUser(c *fiber.Ctx) error {
	id, err := ctrl.ParsePathID(c)
	if err != nil {
		return err // error response already sent
	}

	err = ctrl.userUsecase.DeleteUser(strconv.FormatUint(uint64(id), 10))
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
// @Success 200 {object} domain.SuccessResponse{data=domain.UserFilesData}
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /user/{id}/files [get]
func (ctrl *UserController) GetUserFiles(c *fiber.Ctx) error {
	id, err := ctrl.ParsePathID(c)
	if err != nil {
		return err // error response already sent
	}

	var params domain.UserFilesQueryParams
	if err := c.QueryParser(&params); err != nil {
		return ctrl.SendBadRequest(c, "Parameter query tidak valid")
	}

	result, err := ctrl.userUsecase.GetUserFiles(strconv.FormatUint(uint64(id), 10), params)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return helper.SendAppError(c, appErr)
		}
		return ctrl.SendInternalError(c)
	}

	return ctrl.SendSuccess(c, result, "Daftar file user berhasil diambil")
}

// GetProfile handles GET /api/v1/profile - Get current user profile
// @Summary Get user profile
// @Description Get the profile of the authenticated user
// @Tags User Profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.SuccessResponse{data=domain.ProfileData}
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /profile [get]
func (ctrl *UserController) GetProfile(c *fiber.Ctx) error {
	userID := ctrl.GetAuthenticatedUserID(c)
	if userID == "" {
		return nil // unauthorized response already sent
	}

	result, err := ctrl.userUsecase.GetProfile(userID)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return helper.SendAppError(c, appErr)
		}
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
// @Success 200 {object} domain.SuccessResponse{data=domain.ProfileData}
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /profile [put]
func (ctrl *UserController) UpdateProfile(c *fiber.Ctx) error {
	userID := ctrl.GetAuthenticatedUserID(c)
	if userID == "" {
		return nil // unauthorized response already sent
	}

	var req domain.UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return ctrl.SendBadRequest(c, "Format request tidak valid")
	}

	if !ctrl.ValidateStruct(c, req) {
		return nil // validation error response already sent
	}

	fotoProfil, _ := c.FormFile("foto_profil")

	result, err := ctrl.userUsecase.UpdateProfile(userID, req, fotoProfil)
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
// @Success 200 {object} domain.SuccessResponse{data=[]domain.UserPermissionItem}
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /user/permissions [get]
func (ctrl *UserController) GetUserPermissions(c *fiber.Ctx) error {
	userID := ctrl.GetAuthenticatedUserID(c)
	if userID == "" {
		return nil // unauthorized response already sent
	}

	result, err := ctrl.userUsecase.GetUserPermissions(userID)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return helper.SendAppError(c, appErr)
		}
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
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /user/{id}/download [post]
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

	// Convert uint IDs to strings
	projectIDsStr := make([]string, len(req.ProjectIDs))
	for i, id := range req.ProjectIDs {
		projectIDsStr[i] = strconv.FormatUint(uint64(id), 10)
	}

	modulIDsStr := make([]string, len(req.ModulIDs))
	for i, id := range req.ModulIDs {
		modulIDsStr[i] = strconv.FormatUint(uint64(id), 10)
	}

	filePath, err := ctrl.userUsecase.DownloadUserFiles(strconv.FormatUint(uint64(ownerUserID), 10), projectIDsStr, modulIDsStr)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return helper.SendAppError(c, appErr)
		}
		return ctrl.SendInternalError(c)
	}

	return c.Download(filePath)
}

// GetUsersForRole handles GET /api/v1/role/:id/users - Get users for a specific role
func (ctrl *UserController) GetUsersForRole(c *fiber.Ctx) error {
	id, err := ctrl.ParsePathID(c)
	if err != nil {
		return err
	}

	result, err := ctrl.userUsecase.GetUsersForRole(uint(id))
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return helper.SendAppError(c, appErr)
		}
		return ctrl.SendInternalError(c)
	}

	return ctrl.SendSuccess(c, result, "Daftar user untuk role berhasil diambil")
}

// BulkAssignRole handles POST /api/v1/role/:id/users/bulk - Assign role to multiple users
func (ctrl *UserController) BulkAssignRole(c *fiber.Ctx) error {
	id, err := ctrl.ParsePathID(c)
	if err != nil {
		return err
	}

	var req domain.BulkAssignRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return ctrl.SendBadRequest(c, "Format request tidak valid")
	}

	if !ctrl.ValidateStruct(c, req) {
		return nil
	}

	err = ctrl.userUsecase.BulkAssignRole(req.UserIDs, uint(id))
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return helper.SendAppError(c, appErr)
		}
		return ctrl.SendInternalError(c)
	}

	return ctrl.SendSuccess(c, nil, "Role berhasil ditetapkan ke user")
}
