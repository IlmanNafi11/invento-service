package http

import (
	"errors"

	"invento-service/internal/controller/base"
	"invento-service/internal/dto"
	apperrors "invento-service/internal/errors"
	"invento-service/internal/httputil"
	"invento-service/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

// RoleController handles role-related HTTP requests.
// It embeds BaseController for common functionality like authentication,
// authorization, validation, and response helpers.
type RoleController struct {
	*base.BaseController
	roleUsecase usecase.RoleUsecase
}

// NewRoleController creates a new RoleController instance.
// Parameters:
//   - roleUsecase: Role business logic layer
//   - baseController: Base controller for common functionality (optional)
func NewRoleController(roleUsecase usecase.RoleUsecase, baseController *base.BaseController) *RoleController {
	if baseController == nil {
		baseController = base.NewBaseController("", nil)
	}

	return &RoleController{
		BaseController: baseController,
		roleUsecase:    roleUsecase,
	}
}

// GetAvailablePermissions retrieves all available permissions grouped by resource.
//
// @Summary  Get available permissions
// @Description Retrieves all available permissions grouped by resource
// @Tags   Roles
// @Accept   json
// @Produce  json
// @Success  200 {object} dto.SuccessResponse "Permissions retrieved successfully"
// @Failure  500 {object} dto.ErrorResponse "Internal server error"
// @Router   /role/permissions [get]
// @Security  BearerAuth
func (ctrl *RoleController) GetAvailablePermissions(c *fiber.Ctx) error {
	ctx := c.UserContext()
	permissions, err := ctrl.roleUsecase.GetAvailablePermissions(ctx)
	if err != nil {
		return ctrl.SendInternalError(c)
	}

	response := map[string]interface{}{
		"items": permissions,
	}

	return ctrl.SendSuccess(c, response, "Daftar resource dan permission berhasil diambil")
}

// GetRoleList retrieves a paginated list of roles.
//
// @Summary  Get role list
// @Description Retrieves a paginated list of roles with optional search
// @Tags   Roles
// @Accept   json
// @Produce  json
// @Param   search query  string false "Search keyword"
// @Param   page  query  int  false "Page number"  default(1)
// @Param   limit  query  int  false "Items per page" default(10)
// @Success  200  {object} dto.SuccessResponse "Roles retrieved successfully"
// @Failure  400  {object} dto.ErrorResponse "Invalid query parameters"
// @Failure  500  {object} dto.ErrorResponse "Internal server error"
// @Router   /role [get]
// @Security  BearerAuth
func (ctrl *RoleController) GetRoleList(c *fiber.Ctx) error {
	var params dto.RoleListQueryParams
	if err := c.QueryParser(&params); err != nil {
		return ctrl.SendBadRequest(c, "Parameter query tidak valid")
	}

	ctx := c.UserContext()
	result, err := ctrl.roleUsecase.GetRoleList(ctx, params)
	if err != nil {
		return ctrl.SendInternalError(c)
	}

	return ctrl.SendSuccess(c, result, "Daftar role berhasil diambil")
}

// CreateRole creates a new role with specified permissions.
//
// @Summary  Create role
// @Description Creates a new role with the specified name and permissions
// @Tags   Roles
// @Accept   json
// @Produce  json
// @Param   request body  dto.RoleCreateRequest true "Role creation request"
// @Success  201  {object} dto.SuccessResponse "Role created successfully"
// @Failure  400  {object} dto.ErrorResponse "Invalid request format"
// @Failure  409  {object} dto.ErrorResponse "Role name already exists"
// @Failure  500  {object} dto.ErrorResponse "Internal server error"
// @Router   /role [post]
// @Security  BearerAuth
func (ctrl *RoleController) CreateRole(c *fiber.Ctx) error {
	var req dto.RoleCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return ctrl.SendBadRequest(c, "Format request tidak valid")
	}

	if !ctrl.ValidateStruct(c, req) {
		return nil
	}

	ctx := c.UserContext()
	result, err := ctrl.roleUsecase.CreateRole(ctx, req)
	if err != nil {
		return ctrl.handleRoleError(c, err)
	}

	return ctrl.SendCreated(c, result, "Role berhasil dibuat")
}

// GetRoleDetail retrieves details of a specific role by ID.
//
// @Summary  Get role detail
// @Description Retrieves detailed information about a specific role including its permissions
// @Tags   Roles
// @Accept   json
// @Produce  json
// @Param   id path  int true "Role ID"
// @Success  200 {object} dto.SuccessResponse "Role details retrieved successfully"
// @Failure  400 {object} dto.ErrorResponse "Invalid role ID"
// @Failure  404 {object} dto.ErrorResponse "Role not found"
// @Failure  500 {object} dto.ErrorResponse "Internal server error"
// @Router   /role/{id} [get]
// @Security  BearerAuth
func (ctrl *RoleController) GetRoleDetail(c *fiber.Ctx) error {
	id, err := ctrl.ParsePathID(c)
	if err != nil {
		return err
	}

	ctx := c.UserContext()
	result, err := ctrl.roleUsecase.GetRoleDetail(ctx, id)
	if err != nil {
		return ctrl.handleRoleError(c, err)
	}

	return ctrl.SendSuccess(c, result, "Detail role berhasil diambil")
}

// UpdateRole updates an existing role by ID.
//
// @Summary  Update role
// @Description Updates an existing role's name and permissions
// @Tags   Roles
// @Accept   json
// @Produce  json
// @Param   id  path  int      true "Role ID"
// @Param   request body  dto.RoleUpdateRequest true "Role update request"
// @Success  200  {object} dto.SuccessResponse "Role updated successfully"
// @Failure  400  {object} dto.ErrorResponse "Invalid request format"
// @Failure  404  {object} dto.ErrorResponse "Role not found"
// @Failure  409  {object} dto.ErrorResponse "Role name already exists"
// @Failure  500  {object} dto.ErrorResponse "Internal server error"
// @Router   /role/{id} [put]
// @Security  BearerAuth
func (ctrl *RoleController) UpdateRole(c *fiber.Ctx) error {
	id, err := ctrl.ParsePathID(c)
	if err != nil {
		return err
	}

	var req dto.RoleUpdateRequest
	if err = c.BodyParser(&req); err != nil {
		return ctrl.SendBadRequest(c, "Format request tidak valid")
	}

	if !ctrl.ValidateStruct(c, req) {
		return nil
	}

	ctx := c.UserContext()
	result, err := ctrl.roleUsecase.UpdateRole(ctx, id, req)
	if err != nil {
		return ctrl.handleRoleError(c, err)
	}

	return ctrl.SendSuccess(c, result, "Role berhasil diperbarui")
}

// DeleteRole deletes a role by ID.
//
// @Summary  Delete role
// @Description Deletes a role and removes all associated permissions
// @Tags   Roles
// @Accept   json
// @Produce  json
// @Param   id path int true "Role ID"
// @Success  200 {object} dto.SuccessResponse "Role deleted successfully"
// @Failure  400 {object} dto.ErrorResponse "Invalid role ID"
// @Failure  404 {object} dto.ErrorResponse "Role not found"
// @Failure  500 {object} dto.ErrorResponse "Internal server error"
// @Router   /role/{id} [delete]
// @Security  BearerAuth
func (ctrl *RoleController) DeleteRole(c *fiber.Ctx) error {
	id, err := ctrl.ParsePathID(c)
	if err != nil {
		return err
	}

	ctx := c.UserContext()
	err = ctrl.roleUsecase.DeleteRole(ctx, id)
	if err != nil {
		return ctrl.handleRoleError(c, err)
	}

	return ctrl.SendSuccess(c, nil, "Role berhasil dihapus")
}

// handleRoleError handles role usecase errors and maps them to appropriate HTTP responses.
// Uses type-safe error handling with AppError types.
func (ctrl *RoleController) handleRoleError(c *fiber.Ctx, err error) error {
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		return httputil.SendAppError(c, appErr)
	}
	return ctrl.SendInternalError(c)
}
