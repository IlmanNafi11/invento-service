package http

import (
	"errors"
	"invento-service/internal/controller/base"
	"invento-service/internal/dto"
	"invento-service/internal/httputil"
	"invento-service/internal/usecase"
	"strconv"

	apperrors "invento-service/internal/errors"

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
// @Success 200 {object} dto.SuccessResponse{data=dto.UserListData}
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /user [get]
func (ctrl *UserController) GetUserList(c *fiber.Ctx) error {
	ctx := c.UserContext()
	var params dto.UserListQueryParams
	if err := c.QueryParser(&params); err != nil {
		return ctrl.SendBadRequest(c, "Parameter query tidak valid")
	}

	result, err := ctrl.userUsecase.GetUserList(ctx, params)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return httputil.SendAppError(c, appErr)
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
// @Param id path string true "User ID (UUID)"
// @Param request body dto.UpdateUserRoleRequest true "Role update request"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /user/{id}/role [put]
func (ctrl *UserController) UpdateUserRole(c *fiber.Ctx) error {
	ctx := c.UserContext()
	userID, err := ctrl.ParsePathUUID(c)
	if err != nil {
		return err // error response already sent
	}

	var req dto.UpdateUserRoleRequest
	if err = c.BodyParser(&req); err != nil {
		return ctrl.SendBadRequest(c, "Format request tidak valid")
	}

	if !ctrl.ValidateStruct(c, req) {
		return nil // validation error response already sent
	}

	err = ctrl.userUsecase.UpdateUserRole(ctx, userID, req.Role)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return httputil.SendAppError(c, appErr)
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
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /user/{id} [delete]
func (ctrl *UserController) DeleteUser(c *fiber.Ctx) error {
	ctx := c.UserContext()
	userID, err := ctrl.ParsePathUUID(c)
	if err != nil {
		return err // error response already sent
	}

	err = ctrl.userUsecase.DeleteUser(ctx, userID)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return httputil.SendAppError(c, appErr)
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
// @Param id path string true "User ID (UUID)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} dto.SuccessResponse{data=dto.UserFilesData}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /user/{id}/files [get]
func (ctrl *UserController) GetUserFiles(c *fiber.Ctx) error {
	ctx := c.UserContext()
	userID, err := ctrl.ParsePathUUID(c)
	if err != nil {
		return err // error response already sent
	}

	var params dto.UserFilesQueryParams
	if err = c.QueryParser(&params); err != nil {
		return ctrl.SendBadRequest(c, "Parameter query tidak valid")
	}

	result, err := ctrl.userUsecase.GetUserFiles(ctx, userID, params)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return httputil.SendAppError(c, appErr)
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
// @Success 200 {object} dto.SuccessResponse{data=dto.ProfileData}
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /profile [get]
func (ctrl *UserController) GetProfile(c *fiber.Ctx) error {
	ctx := c.UserContext()
	userID := ctrl.GetAuthenticatedUserID(c)
	if userID == "" {
		return nil // unauthorized response already sent
	}

	result, err := ctrl.userUsecase.GetProfile(ctx, userID)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return httputil.SendAppError(c, appErr)
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
// @Param request body dto.UpdateProfileRequest true "Profile update request"
// @Param foto_profil formData file false "Profile photo (optional)"
// @Success 200 {object} dto.SuccessResponse{data=dto.ProfileData}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /profile [put]
func (ctrl *UserController) UpdateProfile(c *fiber.Ctx) error {
	ctx := c.UserContext()
	userID := ctrl.GetAuthenticatedUserID(c)
	if userID == "" {
		return nil // unauthorized response already sent
	}

	var req dto.UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return ctrl.SendBadRequest(c, "Format request tidak valid")
	}

	if !ctrl.ValidateStruct(c, req) {
		return nil // validation error response already sent
	}

	fotoProfil, _ := c.FormFile("foto_profil")

	result, err := ctrl.userUsecase.UpdateProfile(ctx, userID, req, fotoProfil)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return httputil.SendAppError(c, appErr)
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
// @Success 200 {object} dto.SuccessResponse{data=[]dto.UserPermissionItem}
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /user/permissions [get]
func (ctrl *UserController) GetUserPermissions(c *fiber.Ctx) error {
	ctx := c.UserContext()
	userID := ctrl.GetAuthenticatedUserID(c)
	if userID == "" {
		return nil // unauthorized response already sent
	}

	result, err := ctrl.userUsecase.GetUserPermissions(ctx, userID)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return httputil.SendAppError(c, appErr)
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
// @Param id path string true "User ID (UUID)"
// @Param request body dto.DownloadUserFilesRequest true "Download request"
// @Success 200 {file} binary
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /user/{id}/download [post]
func (ctrl *UserController) DownloadUserFiles(c *fiber.Ctx) error {
	ctx := c.UserContext()
	ownerUserID, err := ctrl.ParsePathUUID(c)
	if err != nil {
		return err // error response already sent
	}

	var req dto.DownloadUserFilesRequest
	if err = c.BodyParser(&req); err != nil {
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

	filePath, err := ctrl.userUsecase.DownloadUserFiles(ctx, ownerUserID, projectIDsStr, modulIDsStr)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return httputil.SendAppError(c, appErr)
		}
		return ctrl.SendInternalError(c)
	}

	return c.Download(filePath)
}

// GetUsersForRole handles GET /api/v1/role/:id/users - Get users for a specific role
// @Summary Get users for a specific role
// @Description Mendapatkan daftar user yang memiliki role tertentu berdasarkan role ID
// @Tags Role Management
// @Produce json
// @Security BearerAuth
// @Param id path int true "Role ID"
// @Success 200 {object} dto.SuccessResponse "Daftar user untuk role berhasil diambil"
// @Failure 400 {object} dto.ErrorResponse "ID role tidak valid"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Role tidak ditemukan"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /role/{id}/users [get]
func (ctrl *UserController) GetUsersForRole(c *fiber.Ctx) error {
	ctx := c.UserContext()
	id, err := ctrl.ParsePathID(c)
	if err != nil {
		return err
	}

	result, err := ctrl.userUsecase.GetUsersForRole(ctx, id)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return httputil.SendAppError(c, appErr)
		}
		return ctrl.SendInternalError(c)
	}

	return ctrl.SendSuccess(c, result, "Daftar user untuk role berhasil diambil")
}

// BulkAssignRole handles POST /api/v1/role/:id/users/bulk - Assign role to multiple users
// @Summary Bulk assign role to users
// @Description Menetapkan role tertentu ke beberapa user sekaligus berdasarkan daftar user ID
// @Tags Role Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Role ID"
// @Param request body dto.BulkAssignRoleRequest true "Daftar user ID"
// @Success 200 {object} dto.SuccessResponse "Role berhasil ditetapkan ke user"
// @Failure 400 {object} dto.ErrorResponse "Format request tidak valid"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Role tidak ditemukan"
// @Failure 422 {object} dto.ErrorResponse "Validasi gagal"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /role/{id}/users/bulk [post]
func (ctrl *UserController) BulkAssignRole(c *fiber.Ctx) error {
	ctx := c.UserContext()
	id, err := ctrl.ParsePathID(c)
	if err != nil {
		return err
	}

	var req dto.BulkAssignRoleRequest
	if err = c.BodyParser(&req); err != nil {
		return ctrl.SendBadRequest(c, "Format request tidak valid")
	}

	if !ctrl.ValidateStruct(c, req) {
		return nil
	}

	err = ctrl.userUsecase.BulkAssignRole(ctx, req.UserIDs, id)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return httputil.SendAppError(c, appErr)
		}
		return ctrl.SendInternalError(c)
	}

	return ctrl.SendSuccess(c, nil, "Role berhasil ditetapkan ke user")
}
