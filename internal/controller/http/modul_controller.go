package http

import (
	"errors"
	"invento-service/config"
	"invento-service/internal/controller/base"
	"invento-service/internal/dto"
	apperrors "invento-service/internal/errors"
	"invento-service/internal/httputil"
	"invento-service/internal/usecase"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type ModulController struct {
	*base.BaseController
	modulUsecase usecase.ModulUsecase
	config       *config.Config
}

func NewModulController(
	modulUsecase usecase.ModulUsecase,
	config *config.Config,
	baseCtrl *base.BaseController,
) *ModulController {
	return &ModulController{
		BaseController: baseCtrl,
		modulUsecase:   modulUsecase,
		config:         config,
	}
}

// GetList handles GET /api/v1/modul
//
// @Summary Get list of modules
// @Description Retrieve paginated list of modules with optional filters
// @Tags Modul
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param search query string false "Search by judul or deskripsi"
// @Param filter_type query string false "Filter by mime type"
// @Param filter_status query string false "Filter by status (pending, completed)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} dto.SuccessResponse{data=dto.ModulListData} "List retrieved successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid query parameters"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /modul [get]
func (ctrl *ModulController) GetList(c *fiber.Ctx) error {
	ctx := c.UserContext()
	userID := ctrl.GetAuthenticatedUserID(c)
	if userID == "" {
		return nil
	}

	var params dto.ModulListQueryParams
	if err := c.QueryParser(&params); err != nil {
		return ctrl.SendBadRequest(c, "Parameter query tidak valid")
	}

	result, err := ctrl.modulUsecase.GetList(ctx, userID, params.Search, params.FilterType, params.FilterStatus, params.Page, params.Limit)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return httputil.SendAppError(c, appErr)
		}
		return ctrl.SendInternalError(c)
	}

	return ctrl.SendSuccess(c, result, "Daftar modul berhasil diambil")
}

// Delete handles DELETE /api/v1/modul/:id
//
// @Summary Delete a module
// @Description Permanently delete a module by ID
// @Tags Modul
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Module ID (UUID)"
// @Success 200 {object} dto.SuccessResponse "Module deleted successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid module ID"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 403 {object} dto.ErrorResponse "Forbidden - no access to this module"
// @Failure 404 {object} dto.ErrorResponse "Module not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /modul/{id} [delete]
func (ctrl *ModulController) Delete(c *fiber.Ctx) error {
	ctx := c.UserContext()
	userID := ctrl.GetAuthenticatedUserID(c)
	if userID == "" {
		return nil
	}

	modulID, err := ctrl.ParsePathUUID(c)
	if err != nil {
		return nil
	}

	err = ctrl.modulUsecase.Delete(ctx, modulID, userID)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return httputil.SendAppError(c, appErr)
		}
		return ctrl.SendInternalError(c)
	}

	return ctrl.SendSuccess(c, nil, "Modul berhasil dihapus")
}

// UpdateMetadata handles PATCH /api/v1/modul/:id
//
// @Summary Update module metadata
// @Description Update judul or deskripsi of an existing module
// @Tags Modul
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Module ID (UUID)"
// @Param request body dto.UpdateModulRequest true "Update request"
// @Success 200 {object} dto.SuccessResponse "Metadata updated successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request format"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 403 {object} dto.ErrorResponse "Forbidden - no access to this module"
// @Failure 404 {object} dto.ErrorResponse "Module not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /modul/{id} [patch]
func (ctrl *ModulController) UpdateMetadata(c *fiber.Ctx) error {
	ctx := c.UserContext()
	userID := ctrl.GetAuthenticatedUserID(c)
	if userID == "" {
		return nil
	}

	modulID, err := ctrl.ParsePathUUID(c)
	if err != nil {
		return nil
	}

	var req dto.UpdateModulRequest
	if err := c.BodyParser(&req); err != nil {
		return ctrl.SendBadRequest(c, "Format request tidak valid")
	}

	if !ctrl.ValidateStruct(c, req) {
		return nil
	}

	err = ctrl.modulUsecase.UpdateMetadata(ctx, modulID, userID, req)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return httputil.SendAppError(c, appErr)
		}
		return ctrl.SendInternalError(c)
	}

	return ctrl.SendSuccess(c, nil, "Metadata modul berhasil diperbarui")
}

// Download handles POST /api/v1/modul/download
//
// @Summary Download modules as ZIP
// @Description Download one or more modules as a ZIP file
// @Tags Modul
// @Accept json
// @Produce application/zip
// @Security BearerAuth
// @Param request body dto.ModulDownloadRequest true "Download request with module IDs"
// @Success 200 {file} binary "ZIP file containing module files"
// @Failure 400 {object} dto.ErrorResponse "Invalid request format or empty IDs"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "One or more modules not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /modul/download [post]
func (ctrl *ModulController) Download(c *fiber.Ctx) error {
	ctx := c.UserContext()
	userID := ctrl.GetAuthenticatedUserID(c)
	if userID == "" {
		return nil
	}

	var req dto.ModulDownloadRequest
	if err := c.BodyParser(&req); err != nil {
		return ctrl.SendBadRequest(c, "Format request tidak valid")
	}

	if len(req.IDs) == 0 {
		return ctrl.SendBadRequest(c, "ID modul tidak boleh kosong")
	}

	if !ctrl.ValidateStruct(c, req) {
		return nil
	}

	filePath, err := ctrl.modulUsecase.Download(ctx, userID, req.IDs)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return httputil.SendAppError(c, appErr)
		}
		return ctrl.SendInternalError(c)
	}

	err = c.Download(filePath)
	if err != nil {
		return err
	}

	cleanPath := filepath.ToSlash(filepath.Clean(filePath))
	if strings.Contains(cleanPath, "/uploads/temp/") || strings.HasPrefix(cleanPath, "uploads/temp/") || strings.HasPrefix(cleanPath, "./uploads/temp/") {
		_ = os.Remove(filePath)
	}

	return nil
}
