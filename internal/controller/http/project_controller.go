package http

import (
	"fiber-boiler-plate/internal/controller/base"
	"fiber-boiler-plate/internal/domain"
	"fiber-boiler-plate/internal/helper"
	apperrors "fiber-boiler-plate/internal/errors"
	"fiber-boiler-plate/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

// ProjectController handles project-related HTTP requests.
// It uses base controller for common operations like authentication and response handling.
type ProjectController struct {
	*base.BaseController
	projectUsecase usecase.ProjectUsecase
}

// NewProjectController creates a new project controller instance.
func NewProjectController(projectUsecase usecase.ProjectUsecase, jwtManager *helper.JWTManager, casbin *helper.CasbinEnforcer) *ProjectController {
	return &ProjectController{
		BaseController: base.NewBaseController(jwtManager, casbin),
		projectUsecase: projectUsecase,
	}
}

// UpdateMetadata handles PATCH /api/v1/project/:id
//
// @Summary Update project metadata
// @Description Update nama_project, kategori, or semester of an existing project
// @Tags Project
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Project ID"
// @Param request body domain.ProjectUpdateRequest true "Update request"
// @Success 200 {object} domain.SuccessResponse "Metadata updated successfully"
// @Failure 400 {object} domain.ErrorResponse "Invalid request format"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 403 {object} domain.ErrorResponse "Forbidden - no access to this project"
// @Failure 404 {object} domain.ErrorResponse "Project not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /project/{id} [patch]
func (ctrl *ProjectController) UpdateMetadata(c *fiber.Ctx) error {
	// Get authenticated user ID using base controller
	userID := ctrl.GetAuthenticatedUserID(c)
	if userID == 0 {
		return nil
	}

	// Parse project ID from path
	projectID, err := ctrl.ParsePathID(c)
	if err != nil {
		return err
	}

	// Parse request body
	var req domain.ProjectUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return ctrl.SendBadRequest(c, "Format request tidak valid")
	}

	// Validate request
	if !ctrl.ValidateStruct(c, req) {
		return nil
	}

	// Call usecase
	err = ctrl.projectUsecase.UpdateMetadata(projectID, userID, req)
	if err != nil {
		// Handle AppError type
		if appErr, ok := err.(*apperrors.AppError); ok {
			return helper.SendAppError(c, appErr)
		}
		// Handle unexpected errors
		return ctrl.SendInternalError(c)
	}

	return ctrl.SendSuccess(c, nil, "Metadata project berhasil diperbarui")
}

// GetByID handles GET /api/v1/project/:id
//
// @Summary Get project by ID
// @Description Retrieve details of a specific project by ID
// @Tags Project
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Project ID"
// @Success 200 {object} domain.SuccessResponse "Project details retrieved successfully"
// @Failure 400 {object} domain.ErrorResponse "Invalid project ID"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 403 {object} domain.ErrorResponse "Forbidden - no access to this project"
// @Failure 404 {object} domain.ErrorResponse "Project not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /project/{id} [get]
func (ctrl *ProjectController) GetByID(c *fiber.Ctx) error {
	// Get authenticated user ID using base controller
	userID := ctrl.GetAuthenticatedUserID(c)
	if userID == 0 {
		return nil
	}

	// Parse project ID from path
	projectID, err := ctrl.ParsePathID(c)
	if err != nil {
		return err
	}

	// Call usecase
	result, err := ctrl.projectUsecase.GetByID(projectID, userID)
	if err != nil {
		// Handle AppError type
		if appErr, ok := err.(*apperrors.AppError); ok {
			return helper.SendAppError(c, appErr)
		}
		// Handle unexpected errors
		return ctrl.SendInternalError(c)
	}

	return ctrl.SendSuccess(c, result, "Detail project berhasil diambil")
}

// GetList handles GET /api/v1/project
//
// @Summary Get list of projects
// @Description Retrieve paginated list of projects with optional filters
// @Tags Project
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param search query string false "Search by project name"
// @Param filter_semester query int false "Filter by semester (1-8)"
// @Param filter_kategori query string false "Filter by category (website, mobile, iot, machine_learning, deep_learning)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} domain.SuccessResponse "List retrieved successfully"
// @Failure 400 {object} domain.ErrorResponse "Invalid query parameters"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /project [get]
func (ctrl *ProjectController) GetList(c *fiber.Ctx) error {
	// Get authenticated user ID using base controller
	userID := ctrl.GetAuthenticatedUserID(c)
	if userID == 0 {
		return nil
	}

	// Parse query parameters
	var params domain.ProjectListQueryParams
	if err := c.QueryParser(&params); err != nil {
		return ctrl.SendBadRequest(c, "Parameter query tidak valid")
	}

	// Set default pagination
	page, limit, _ := ctrl.ParsePagination(c)
	params.Page = page
	params.Limit = limit

	// Call usecase
	result, err := ctrl.projectUsecase.GetList(userID, params.Search, params.FilterSemester, params.FilterKategori, params.Page, params.Limit)
	if err != nil {
		// Handle AppError type
		if appErr, ok := err.(*apperrors.AppError); ok {
			return helper.SendAppError(c, appErr)
		}
		// Handle unexpected errors
		return ctrl.SendInternalError(c)
	}

	return ctrl.SendSuccess(c, result, "Daftar project berhasil diambil")
}

// Delete handles DELETE /api/v1/project/:id
//
// @Summary Delete a project
// @Description Permanently delete a project by ID
// @Tags Project
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Project ID"
// @Success 200 {object} domain.SuccessResponse "Project deleted successfully"
// @Failure 400 {object} domain.ErrorResponse "Invalid project ID"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 403 {object} domain.ErrorResponse "Forbidden - no access to this project"
// @Failure 404 {object} domain.ErrorResponse "Project not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /project/{id} [delete]
func (ctrl *ProjectController) Delete(c *fiber.Ctx) error {
	// Get authenticated user ID using base controller
	userID := ctrl.GetAuthenticatedUserID(c)
	if userID == 0 {
		return nil
	}

	// Parse project ID from path
	projectID, err := ctrl.ParsePathID(c)
	if err != nil {
		return err
	}

	// Call usecase
	err = ctrl.projectUsecase.Delete(projectID, userID)
	if err != nil {
		// Handle AppError type
		if appErr, ok := err.(*apperrors.AppError); ok {
			return helper.SendAppError(c, appErr)
		}
		// Handle unexpected errors
		return ctrl.SendInternalError(c)
	}

	return ctrl.SendSuccess(c, nil, "Project berhasil dihapus")
}

// Download handles POST /api/v1/project/download
//
// @Summary Download projects as ZIP
// @Description Download one or more projects as a ZIP file
// @Tags Project
// @Accept json
// @Produce application/zip
// @Security BearerAuth
// @Param request body domain.ProjectDownloadRequest true "Download request with project IDs"
// @Success 200 {file} binary "ZIP file containing project files"
// @Failure 400 {object} domain.ErrorResponse "Invalid request format or empty IDs"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 404 {object} domain.ErrorResponse "One or more projects not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /project/download [post]
func (ctrl *ProjectController) Download(c *fiber.Ctx) error {
	// Get authenticated user ID using base controller
	userID := ctrl.GetAuthenticatedUserID(c)
	if userID == 0 {
		return nil
	}

	// Parse request body
	var req domain.ProjectDownloadRequest
	if err := c.BodyParser(&req); err != nil {
		return ctrl.SendBadRequest(c, "Format request tidak valid")
	}

	// Validate request
	if !ctrl.ValidateStruct(c, req) {
		return nil
	}

	// Additional validation: ensure IDs are provided
	if len(req.IDs) == 0 {
		return ctrl.SendBadRequest(c, "ID project tidak boleh kosong")
	}

	// Call usecase
	filePath, err := ctrl.projectUsecase.Download(userID, req.IDs)
	if err != nil {
		// Handle AppError type
		if appErr, ok := err.(*apperrors.AppError); ok {
			return helper.SendAppError(c, appErr)
		}
		// Handle unexpected errors
		return ctrl.SendInternalError(c)
	}

	return c.Download(filePath)
}
